package monitor

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

//TODO: Arbitrary for now, allow for updating based on seen blocks?
const maxTransactionMultiplier = 10

const timeoutPeriod = 2 * time.Second

type BlockAndTransactions struct {
	block *types.Block
	txs   []*types.Transaction
}

type BatchWriter struct {
	maxBlocks       int
	maxTransactions int

	currentWorkUnits        []*BlockAndTransactions
	currentTransactionCount int

	BatchWorkChan chan *BlockAndTransactions
	db            database.Database

	stopFeed event.Feed
}

func NewBatchWriter(batchWorkChan chan *BlockAndTransactions, db database.Database) *BatchWriter {
	bp := &BatchWriter{
		maxBlocks:               cap(batchWorkChan),
		maxTransactions:         maxTransactionMultiplier * cap(batchWorkChan),
		currentWorkUnits:        make([]*BlockAndTransactions, 0, cap(batchWorkChan)),
		currentTransactionCount: 0,
		BatchWorkChan:           batchWorkChan,
		db:                      db,
	}
	return bp
}

func (bw *BatchWriter) Run(stopChan <-chan types.StopEvent) {
	ticker := time.NewTicker(timeoutPeriod)
	for {
		select {
		case newWorkUnit := <-bw.BatchWorkChan:
			// Listen to new block channel and process if new block comes.
			bw.currentWorkUnits = append(bw.currentWorkUnits, newWorkUnit)
			bw.currentTransactionCount += len(newWorkUnit.txs)
			if len(bw.currentWorkUnits) >= bw.maxBlocks || bw.currentTransactionCount >= bw.maxTransactions {
				//if the write fails, keep trying until it succeeds, waiting
				//the defined timeout period between attempts
				for err := bw.BatchWrite(); err != nil; err = bw.BatchWrite() {
					log.Printf("batch write failed: %v", err)
					<-ticker.C
				}
			}
			ticker.Stop()
			ticker = time.NewTicker(timeoutPeriod)
		case <-ticker.C:
			//if this fails, it will try again on the next run
			if err := bw.BatchWrite(); err != nil {
				log.Printf("batch write failed: %v", err)
			}
		case <-stopChan:
			ticker.Stop()
			return
		}
	}
}

func (bw *BatchWriter) BatchWrite() error {
	if len(bw.currentWorkUnits) == 0 {
		return nil
	}

	allTxns := make([]*types.Transaction, 0, bw.currentTransactionCount)
	allBlocks := make([]*types.Block, 0, len(bw.currentWorkUnits))
	for _, workUnit := range bw.currentWorkUnits {
		allTxns = append(allTxns, workUnit.txs...)
		allBlocks = append(allBlocks, workUnit.block)
	}

	err := bw.db.WriteTransactions(allTxns)
	if err != nil {
		return err
	}
	err = bw.db.WriteBlocks(allBlocks)
	if err != nil {
		return err
	}

	// reset
	bw.currentTransactionCount = 0
	bw.currentWorkUnits = make([]*BlockAndTransactions, 0, bw.maxBlocks)
	return nil
}
