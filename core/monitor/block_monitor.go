package monitor

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
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
	// concurrent block processing
	newBlockChan     chan *types.Block
	availableWorkers uint64
}

func NewBlockMonitor(db database.Database, quorumClient client.Client) *BlockMonitor {
	return &BlockMonitor{
		db:                 db,
		quorumClient:       quorumClient,
		syncStart:          make(chan uint64, 1), // make channel buffered so that it does not block chain head listener
		transactionMonitor: NewTransactionMonitor(db, quorumClient),
		newBlockChan:       make(chan *types.Block),
		availableWorkers:   10,
	}
}

func (bm *BlockMonitor) Start() error {
	// Pulling historical blocks since the last persisted while continuously listening to ChainHeadEvent.
	// For every block received, pull transactions/ events related to the registered contracts.

	log.Println("Start to sync blocks...")

	// 1. Start worker
	for i := uint64(0); i < bm.availableWorkers; i++ {
		NewBlockProcessor(bm.newBlockChan, bm).Run()
	}

	// 2. Fetch the current block height.
	currentBlockNumber, err := bm.currentBlockNumber()
	if err != nil {
		return err
	}
	log.Printf("Current block head is: %v.\n", currentBlockNumber)

	// 3. Sync from last persisted to current block height.
	lastPersisted, err := bm.db.GetLastPersistedBlockNumber()
	if err != nil {
		return err
	}
	go bm.sync(lastPersisted, currentBlockNumber)

	// 4. Listen to ChainHeadEvent and sync.
	err = bm.listenToChainHead()
	if err != nil {
		return err
	}

	return nil
}

func (bm *BlockMonitor) Stop() {
	bm.stopFeed.Send(types.StopEvent{})
}

func (bm *BlockMonitor) subscribeStopEvent() (chan types.StopEvent, event.Subscription) {
	c := make(chan types.StopEvent)
	s := bm.stopFeed.Subscribe(c)
	return c, s
}

func (bm *BlockMonitor) currentBlockNumber() (uint64, error) {
	var (
		resp         map[string]interface{}
		currentBlock graphql.CurrentBlock
	)
	err := bm.quorumClient.ExecuteGraphQLQuery(context.Background(), &resp, graphql.CurrentBlockQuery())
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
	headers := make(chan *ethTypes.Header)
	sub, err := bm.quorumClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		return err
	}
	go func() {
		stopChan, stopSubscription := bm.subscribeStopEvent()
		defer stopSubscription.Unsubscribe()
		syncStarted := false
		for {
			select {
			case err := <-sub.Err():
				log.Panicf("chain head event subscription error: %v", err)
			case header := <-headers:
				if !syncStarted {
					bm.syncStart <- header.Number.Uint64()
					syncStarted = true
				}
				blockOrigin, err := bm.quorumClient.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					// TODO: if quorum node is down, reconnect?
					log.Panicf("get block with hash %v error: %v", header.Hash(), err)
				}
				bm.newBlockChan <- createBlock(blockOrigin)
			case <-stopChan:
				return
			}
		}
	}()
	return nil

}

func (bm *BlockMonitor) syncBlocks(start, end uint64) error {
	log.Printf("Start to sync historic blocks from %v to %v. \n", start, end)
	for i := start + 1; i <= end; i++ {
		blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			// TODO: if quorum node is down, reconnect?
			return err
		}
		bm.newBlockChan <- createBlock(blockOrigin)
	}
	return nil
}

func (bm *BlockMonitor) sync(start, end uint64) {
	err := bm.syncBlocks(start, end)
	if err != nil {
		log.Panicf("sync historic blocks from %v to %v failed: %v", start, end, err)
	}

	// Sync from end + 1 to the first ChainHeadEvent if there is any gap.
	stopChan, stopSubscription := bm.subscribeStopEvent()
	defer stopSubscription.Unsubscribe()

	select {
	case latestChainHead := <-bm.syncStart:
		close(bm.syncStart)
		err := bm.syncBlocks(end, latestChainHead-1)
		if err != nil {
			log.Panicf("sync historic blocks from %v to %v failed: %v", end, latestChainHead-1, err)
		}
	case <-stopChan:
		return
	}
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
