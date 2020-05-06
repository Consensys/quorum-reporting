package monitor

import (
	"context"
	"log"
	"math/big"

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
	transactionMonitor *TransactionMonitor
	newBlockChan       chan *types.Block
}

func NewBlockMonitor(db database.Database, quorumClient client.Client) *BlockMonitor {
	return &BlockMonitor{
		db:                 db,
		quorumClient:       quorumClient,
		transactionMonitor: NewTransactionMonitor(db, quorumClient),
		newBlockChan:       make(chan *types.Block),
	}
}

func (bm *BlockMonitor) startWorker(stopChan <-chan types.StopEvent) {
	for {
		select {
		case block := <-bm.newBlockChan:
			// listening to new block channel and process if new block comes
			err := bm.process(block)
			if err != nil {
				log.Panicf("process block %v error: %v", block.Number, err)
			}
		case <-stopChan:
			return
		}
	}
}

func (bm *BlockMonitor) process(block *types.Block) error {
	// Transaction monitor pulls all transactions for the given block.
	err := bm.transactionMonitor.PullTransactions(block)
	if err != nil {
		return err
	}
	// Write block to DB.
	return bm.db.WriteBlock(block)
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

func (bm *BlockMonitor) syncBlocks(start, end uint64) error {
	if start <= end {
		log.Printf("Start to sync historic blocks from %v to %v. \n", start, end)
		for i := start; i <= end; i++ {
			blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
			if err != nil {
				// TODO: if quorum node is down, reconnect?
				return err
			}
			bm.newBlockChan <- createBlock(blockOrigin)
		}
	}
	return nil
}

func (bm *BlockMonitor) processChainHead(header *ethTypes.Header) {
	blockOrigin, err := bm.quorumClient.BlockByHash(context.Background(), header.Hash())
	if err != nil {
		// TODO: if quorum node is down, reconnect?
		log.Panicf("get block with hash %v error: %v", header.Hash(), err)
	}
	bm.newBlockChan <- createBlock(blockOrigin)
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
