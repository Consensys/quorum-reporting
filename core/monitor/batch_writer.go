package monitor

import (
	"time"

	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

//TODO: Arbitrary for now, allow for updating based on seen blocks?
const maxTransactionMultiplier = 10

type BlockAndTransactions struct {
	block *types.Block
	txs   []*types.Transaction
}

type BatchWriter struct {
	maxBlocks       int
	maxTransactions int
	flushPeriod     int

	currentWorkUnits        []*BlockAndTransactions
	currentTransactionCount int

	BatchWorkChan chan *BlockAndTransactions
	db            database.Database

	stopFeed event.Feed
}

func NewBatchWriter(db database.Database, batchWorkChan chan *BlockAndTransactions, flushPeriod int) *BatchWriter {
	return &BatchWriter{
		maxBlocks:               cap(batchWorkChan),
		maxTransactions:         maxTransactionMultiplier * cap(batchWorkChan),
		flushPeriod:             flushPeriod,
		currentWorkUnits:        make([]*BlockAndTransactions, 0, cap(batchWorkChan)),
		currentTransactionCount: 0,
		BatchWorkChan:           batchWorkChan,
		db:                      db,
	}
}

func (bw *BatchWriter) Run(stopChan <-chan types.StopEvent) {
	log.Info("Starting batch block processor", "timeout period", time.Duration(bw.flushPeriod)*time.Second, "max blocks", bw.maxBlocks, "max txns", bw.maxTransactions)

	ticker := time.NewTicker(time.Duration(bw.flushPeriod) * time.Second)
	defer ticker.Stop()
	for {
		// Listen to new block channel and process if new block comes.
		select {
		case newWorkUnit := <-bw.BatchWorkChan:
			log.Debug("Next block found for batch processing", "block", newWorkUnit.block.Hash.String(), "tx count", len(newWorkUnit.txs))
			bw.currentWorkUnits = append(bw.currentWorkUnits, newWorkUnit)
			bw.currentTransactionCount += len(newWorkUnit.txs)

			if len(bw.currentWorkUnits) >= bw.maxBlocks || bw.currentTransactionCount >= bw.maxTransactions {
				log.Info("Max batch write limit reached")
				//if the write fails, keep trying until it succeeds, waiting
				//the defined timeout period between attempts
				for err := bw.BatchWrite(); err != nil; err = bw.BatchWrite() {
					log.Warn("Batch write failed", "err", err)
					<-ticker.C
				}
			}
		case <-ticker.C:
			log.Debug("Batch writing blocks/transactions from ticker")
			//if this fails, it will try again on the next run
			if err := bw.BatchWrite(); err != nil {
				log.Warn("Batch write failed", "err", err)
			}
		case <-stopChan:
			return
		}
	}
}

func (bw *BatchWriter) BatchWrite() error {
	if len(bw.currentWorkUnits) == 0 {
		log.Debug("No blocks/transaction to write")
		return nil
	}

	allTxns := make([]*types.Transaction, 0, bw.currentTransactionCount)
	allBlocks := make([]*types.Block, 0, len(bw.currentWorkUnits))
	for _, workUnit := range bw.currentWorkUnits {
		allTxns = append(allTxns, workUnit.txs...)
		allBlocks = append(allBlocks, workUnit.block)
	}

	log.Info("Batch writing blocks and transactions", "block count", len(allBlocks), "tx count", len(allTxns))

	if err := bw.db.WriteTransactions(allTxns); err != nil {
		return err
	}
	if err := bw.db.WriteBlocks(allBlocks); err != nil {
		return err
	}

	// reset
	bw.currentTransactionCount = 0
	bw.currentWorkUnits = make([]*BlockAndTransactions, 0, bw.maxBlocks)
	return nil
}
