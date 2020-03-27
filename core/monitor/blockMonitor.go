package monitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/mitchellh/mapstructure"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/types"
)

type BlockMonitor struct {
	db                 database.Database
	quorumClient       client.Client
	syncStart          chan uint64
	transactionMonitor *TransactionMonitor
	stopFeed           event.Feed
	syncStarted        bool
	syncStartHead      uint64
	startWaitGroup     *sync.WaitGroup
}

func NewBlockMonitor(db database.Database, quorumClient client.Client) *BlockMonitor {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &BlockMonitor{
		db:                 db,
		quorumClient:       quorumClient,
		syncStart:          make(chan uint64),
		transactionMonitor: NewTransactionMonitor(db, quorumClient),
		syncStarted:        false,
		syncStartHead:      0,
		startWaitGroup:     wg,
	}
}

// to signal all watches when service is stopped
type stopEvent struct {
}

func (bm *BlockMonitor) subscribeStopEvent() (chan stopEvent, event.Subscription) {
	c := make(chan stopEvent)
	s := bm.stopFeed.Subscribe(c)
	return c, s
}

func (bm *BlockMonitor) Start() error {
	// Pulling historical blocks since the last persisted while continuously listening to ChainHeadEvent.
	// For every block received, pull transactions/ events related to the registered contracts.

	fmt.Println("Start to sync blocks...")

	// 1. Fetch the current block height.
	currentBlockNumber, err := bm.currentBlockNumber()
	if err != nil {
		return err
	}
	fmt.Printf("Current block head is: %v\n", currentBlockNumber)

	// 2. Sync from last persisted to current block height.
	go bm.syncBlocks(bm.db.GetLastPersistedBlockNumber(), currentBlockNumber)

	// 3. Listen to ChainHeadEvent and sync.
	err = bm.listenToChainHead()
	log.Println("git error in Start monitor service")
	if err != nil {
		log.Println("git error in Start monitor service 1")
		return err
	}

	return nil
}

func (bm *BlockMonitor) Stop() {
	bm.stopFeed.Send(stopEvent{})
	fmt.Println("monitor service stopped.")
}

func (bm *BlockMonitor) currentBlockNumber() (uint64, error) {
	var (
		resp         map[string]interface{}
		currentBlock graphql.CurrentBlock
	)
	resp, err := bm.quorumClient.ExecuteGraphQLQuery(context.Background(), graphql.CurrentBlockQuery())
	if err != nil {
		return 0, err
	}
	err = mapstructure.Decode(resp["block"].(map[string]interface{}), &currentBlock)
	if err != nil {
		return 0, err
	}
	return hexutil.DecodeUint64(currentBlock.Number)
}

func (bm *BlockMonitor) listenToChainHead() error {
	stopChan, stopSubscription := bm.subscribeStopEvent()
	defer stopSubscription.Unsubscribe()

	headers := make(chan *ethTypes.Header)
	sub, err := bm.quorumClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Println("error 3")
		return err
	}
	go func() {
		for {
			select {
			case err := <-sub.Err():
				// TODO: should gracefully handle error (if quorum node is down, reconnect?)
				log.Fatalf("chain head event subscription error: %v.\n", err)
			case header := <-headers:
				if !bm.syncStarted {
					bm.syncStarted = true
					bm.syncStartHead = header.Number.Uint64()
				}
				// TODO: do we want to change to FIFO queue and push headers to queue instead of direct processing
				blockOrigin, err := bm.quorumClient.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					// TODO: should gracefully handle error (if quorum node is down, reconnect?)
					log.Fatalf("get block %v error: %v.\n", header.Hash(), err)
				}
				err = bm.process(createBlock(blockOrigin))
				if err != nil {
					// TODO: should gracefully handle error (if quorum node is down, reconnect?)
					log.Fatalf("process block %v error: %v.\n", header.Hash(), err)
				}
			case <-stopChan:
				return
			}
		}
	}()
	return nil

}

func (bm *BlockMonitor) syncBlocks(start, end uint64) error {
	execSync := func(start, end uint64) error {
		for i := start + 1; i <= end; i++ {
			blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
			if err != nil {
				return err
			}
			err = bm.process(createBlock(blockOrigin))
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := execSync(start, end)
	if err != nil {
		return errors.New(fmt.Sprintf("sync failed %v\n", err))
	}
	bm.startWaitGroup.Add(1)
	go func(_wg *sync.WaitGroup) {
		stopChan, stopSubscription := bm.subscribeStopEvent()
		pollingTicker := time.NewTicker(10 * time.Millisecond)
		defer func(start time.Time) {
			stopSubscription.Unsubscribe()
			pollingTicker.Stop()
			_wg.Done()
		}(time.Now())
		for {
			select {
			case <-pollingTicker.C:
				if bm.syncStarted {
					lastPersistedBlock := bm.db.GetLastPersistedBlockNumber()
					if lastPersistedBlock < bm.syncStartHead {
						_ = execSync(bm.db.GetLastPersistedBlockNumber(), bm.syncStartHead)
					}
					return
				}
			case <-stopChan:
				return
			}
		}
	}(bm.startWaitGroup)

	return nil
}

func (bm *BlockMonitor) process(block *types.Block) error {
	// Transaction monitor pulls all transactions for the given block.
	err := bm.transactionMonitor.PullTransactions(block)
	if err != nil {
		return err
	}

	// Write block to DB.
	err = bm.db.WriteBlock(block)
	if err != nil {
		return err
	}
	return nil
}

func createBlock(block *ethTypes.Block) *types.Block {
	txs := []common.Hash{}
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash())
	}
	return &types.Block{
		Hash:         block.Hash(),
		ParentHash:   block.ParentHash(),
		StateRoot:    block.Root(),
		TxRoot:       block.TxHash(),
		ReceiptRoot:  block.ReceiptHash(),
		Number:       block.NumberU64(),
		GasLimit:     block.GasLimit(),
		GasUsed:      block.GasUsed(),
		Timestamp:    block.Time(),
		ExtraData:    block.Extra(),
		Transactions: txs,
	}
}
