package monitor

import (
	"log"

	"quorumengineering/quorum-report/types"
)

type BlockProcessor struct {
	blockMonitor *BlockMonitor

	workChan <-chan *types.Block
}

func NewBlockProcessor(workChan <-chan *types.Block, blockMonitor *BlockMonitor) *BlockProcessor {
	return &BlockProcessor{
		blockMonitor: blockMonitor,
		workChan:     workChan,
	}
}

func (bp *BlockProcessor) Run() {
	go func() {
		stopChan, stopSubscription := bp.blockMonitor.subscribeStopEvent()
		defer stopSubscription.Unsubscribe()

		for {
			select {
			case block := <-bp.workChan:
				err := bp.Process(block)
				for err != nil {
					log.Printf("process block %v error: %v\n", block.Number, err)
					err = bp.Process(block)
				}
			case <-stopChan:
				return
			}
		}
	}()
}

func (bp *BlockProcessor) Process(block *types.Block) error {
	// Transaction monitor pulls all transactions for the given block.
	err := bp.blockMonitor.transactionMonitor.PullTransactions(block)
	if err != nil {
		return err
	}

	// Write block to DB.
	return bp.blockMonitor.db.WriteBlock(block)
}
