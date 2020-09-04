package monitor

import (
	"sync"
	"time"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type BlockMonitor interface {
	ListenToChainHead(cancelChan chan bool, stopChan chan bool) error
	SyncHistoricBlocks(lastPersisted uint64, cancelChan chan bool, wg *sync.WaitGroup) error
}

type DefaultBlockMonitor struct {
	quorumClient client.Client
	newBlockChan chan *types.Block
	consensus    string
}

func NewDefaultBlockMonitor(quorumClient client.Client, newBlockChan chan *types.Block, consensus string) *DefaultBlockMonitor {
	return &DefaultBlockMonitor{
		quorumClient: quorumClient,
		newBlockChan: newBlockChan,
		consensus:    consensus,
	}
}

func (bm *DefaultBlockMonitor) ListenToChainHead(cancelChan chan bool, stopChan chan bool) error {
	// make headers channel buffered so that it doesn't block websocket listener
	headers := make(chan types.RawHeader, 10)
	if err := bm.quorumClient.SubscribeChainHead(headers); err != nil {
		return err
	}

	go func() {
		defer close(cancelChan)
		log.Info("Starting chain head listener.")
		for {
			select {
			case header := <-headers:
				bm.processChainHead(header)
			case <-stopChan:
				log.Info("Stopping chain head listener.")
				return
			}
		}
	}()

	return nil
}

func (bm *DefaultBlockMonitor) SyncHistoricBlocks(lastPersisted uint64, cancelChan chan bool, wg *sync.WaitGroup) error {
	currentBlockNumber, err := client.CurrentBlock(bm.quorumClient)
	if err != nil {
		return err
	}
	log.Info("Queried current block head from Quorum", "block number", currentBlockNumber)

	// Sync is called in a go routine so that it doesn't block main process.
	go func() {
		defer log.Info("Returning from historical block processing.")
		defer wg.Done()
		err := bm.syncBlocks(lastPersisted+1, currentBlockNumber, cancelChan)
		for err != nil {
			log.Info("Sync historic blocks failed", "end-block", currentBlockNumber, "err", err)
			time.Sleep(time.Second)
			err = bm.syncBlocks(err.EndBlockNumber(), currentBlockNumber, cancelChan)
		}
	}()

	return nil
}

func (bm *DefaultBlockMonitor) processChainHead(header types.RawHeader) {
	log.Info("Processing chain head", "block hash", header.Hash.String(), "block number", header.Number)
	blockOrigin, err := bm.tryFetchingBlock(header.Number.ToUint64(), 10)
	if err != nil {
		log.Error("Error - fetching block from Quorum failed", "block hash", header.Hash, "block number", header.Number, "err", err)
		return
	}
	bm.newBlockChan <- bm.createBlock(blockOrigin)
}

func (bm *DefaultBlockMonitor) createBlock(block *types.RawBlock) *types.Block {
	timestamp := block.Timestamp.ToUint64()
	if bm.consensus == "raft" {
		timestamp = timestamp / 1_000_000_000
	}

	return &types.Block{
		Hash:         block.Hash,
		ParentHash:   block.ParentHash,
		StateRoot:    block.StateRoot,
		TxRoot:       block.TxRoot,
		ReceiptRoot:  block.ReceiptRoot,
		Number:       block.Number.ToUint64(),
		GasLimit:     block.GasLimit.ToUint64(),
		GasUsed:      block.GasUsed.ToUint64(),
		Timestamp:    timestamp,
		ExtraData:    block.ExtraData,
		Transactions: block.Transactions,
	}
}

func (bm *DefaultBlockMonitor) syncBlocks(start, end uint64, stopChan chan bool) *SyncError {
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

		blockOrigin, err := bm.tryFetchingBlock(i, 10)
		if err != nil {
			return NewSyncError(err.Error(), i)
		}

		select {
		case <-stopChan:
			return nil
		case bm.newBlockChan <- bm.createBlock(blockOrigin):
		}
	}

	log.Info("Complete historical sync finished", "start", start, "end", end)
	return nil
}

func (bm *DefaultBlockMonitor) tryFetchingBlock(number uint64, tryCount int) (*types.RawBlock, error) {
	var err error
	var block types.RawBlock
	for tryCount > 0 {
		block, err = client.BlockByNumber(bm.quorumClient, number)
		if err == nil {
			log.Info("fetched block", "block number", number)
			break
		}

		if err != nil {
			log.Warn("fetching block from Quorum failed, retry", "tryCount", tryCount, "block number", number, "err", err)
		}

		tryCount--
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, err
	}
	return &block, err
}
