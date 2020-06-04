package monitor

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type BlockMonitor struct {
	db                 database.Database
	quorumClient       client.Client
	transactionMonitor *TransactionMonitor
	newBlockChan       chan *types.Block          // concurrent block processing
	batchWriteChan     chan *BlockAndTransactions // concurrent block processing
	consensus          string
}

func NewBlockMonitor(db database.Database, quorumClient client.Client, consensus string, batchWriteChan chan *BlockAndTransactions) *BlockMonitor {
	return &BlockMonitor{
		db:                 db,
		quorumClient:       quorumClient,
		transactionMonitor: NewTransactionMonitor(db, quorumClient),
		newBlockChan:       make(chan *types.Block),
		batchWriteChan:     batchWriteChan,
		consensus:          consensus,
	}
}

func (bm *BlockMonitor) startWorker(stopChan <-chan types.StopEvent) {
	for {
		select {
		case block := <-bm.newBlockChan:
			// Listen to new block channel and process if new block comes.
			err := bm.process(block)
			for err != nil {
				log.Warn("error processing block", "block number", block.Number, "err", err)
				time.Sleep(time.Second)
				err = bm.process(block)
			}
		case <-stopChan:
			log.Debug("stop message received", "location", "BlockMonitor::startWorker")
			return
		}
	}
}

func (bm *BlockMonitor) process(block *types.Block) error {
	// Transaction monitor pulls all transactions for the given block.
	fetchedTxns, err := bm.transactionMonitor.PullTransactions(block)
	if err != nil {
		return err
	}

	workUnit := &BlockAndTransactions{
		block: block,
		txs:   fetchedTxns,
	}

	bm.batchWriteChan <- workUnit
	return nil
}

func (bm *BlockMonitor) currentBlockNumber() (uint64, error) {
	log.Debug("fetching current block number")

	var currentBlockResult graphql.CurrentBlockResult
	if err := bm.quorumClient.ExecuteGraphQLQuery(&currentBlockResult, graphql.CurrentBlockQuery()); err != nil {
		return 0, err
	}

	log.Debug("current block number found", "number", currentBlockResult.Block.Number)
	return hexutil.DecodeUint64(currentBlockResult.Block.Number)
}

func (bm *BlockMonitor) syncBlocks(start, end uint64, stopChan chan bool) *types.SyncError {
	if start > end {
		return nil
	}

	log.Info("Syncing historic blocks", "start", start, "end", end)
	for i := start; i <= end; i++ {
		select {
		case <-stopChan:
			return nil
		default:
		}

		blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			return types.NewSyncError(err.Error(), i)
		}

		select {
		case <-stopChan:
			return nil
		case bm.newBlockChan <- bm.createBlock(blockOrigin):
		}
	}
	return nil
}

func (bm *BlockMonitor) processChainHead(header *ethTypes.Header) {
	log.Info("processing chain head", "block hash", header.Hash().String(), "block number", header.Number.String())
	blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), header.Number)
	for err != nil {
		log.Warn("error fetching block from Quorum", "block hash", header.Hash(), "block number", header.Number.String(), "err", err)
		time.Sleep(1 * time.Second) //TODO: return err and let caller handle?
		blockOrigin, err = bm.quorumClient.BlockByNumber(context.Background(), header.Number)
	}
	bm.newBlockChan <- bm.createBlock(blockOrigin)
}

func (bm *BlockMonitor) createBlock(block *ethTypes.Block) *types.Block {
	txs := []common.Hash{}
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash())
	}

	timestamp := block.Time()
	if bm.consensus == "raft" {
		timestamp = timestamp / 1_000_000_000
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
		Timestamp:    timestamp,
		ExtraData:    block.Extra(),
		Transactions: txs,
	}
}
