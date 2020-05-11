package monitor

import (
	"github.com/ethereum/go-ethereum/event"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
	"time"
)

type BlockAndTransactions struct {
	block *types.Block
	txns  []*types.Transaction
}

type BlockPersister struct {
	maxBlocks       uint64
	maxTransactions uint64

	currentBlockCount       uint64
	currentTransactionCount uint64
	currentWorkUnits        []*BlockAndTransactions

	toPersistChan chan *BlockAndTransactions
	db            database.Database

	stopFeed event.Feed
}

func NewBlockPersister(toPersistChan chan *BlockAndTransactions, db database.Database) *BlockPersister {
	bp := &BlockPersister{
		toPersistChan:   toPersistChan,
		db:              db,
		maxBlocks:       100,
		maxTransactions: 1000,
	}
	bp.Reset()
	return bp
}

func (bp *BlockPersister) Run() {
	stopChan, stopSubscription := bp.subscribeStopEvent()
	defer stopSubscription.Unsubscribe()

	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case workunit := <-bp.toPersistChan:
			// Listen to new block channel and process if new block comes.
			bp.currentWorkUnits = append(bp.currentWorkUnits, workunit)
			if uint64(len(bp.currentWorkUnits)) == bp.maxBlocks {
				for err := bp.Persist(); err != nil; {
					<-ticker.C
				}
				bp.Reset()
			}
			ticker = time.NewTicker(2 * time.Second)
		case <-ticker.C:
			bp.Persist()
			bp.Reset()
		case <-stopChan:
			return
		}
	}
}

func (bp *BlockPersister) Stop() {
	bp.stopFeed.Send(types.StopEvent{})
}

func (bp *BlockPersister) Reset() {
	bp.currentBlockCount = bp.maxBlocks
	bp.currentTransactionCount = bp.maxTransactions
	bp.currentWorkUnits = make([]*BlockAndTransactions, 0, bp.maxBlocks)
}

func (bp *BlockPersister) Persist() error {
	allTxns := make([]*types.Transaction, 0, 1000)
	allBlocks := make([]*types.Block, 0, 1000)
	for _, workunit := range bp.currentWorkUnits {
		allTxns = append(allTxns, workunit.txns...)
		allBlocks = append(allBlocks, workunit.block)
	}

	err := bp.db.WriteTransactions(allTxns)
	if err != nil {
		return err
	}
	return bp.db.WriteBlocks(allBlocks)
}

func (bp *BlockPersister) subscribeStopEvent() (chan types.StopEvent, event.Subscription) {
	c := make(chan types.StopEvent)
	s := bp.stopFeed.Subscribe(c)
	return c, s
}
