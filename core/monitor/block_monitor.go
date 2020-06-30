package monitor

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/graphql"
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
	headers := make(chan *ethTypes.Header, 10)
	err := bm.quorumClient.SubscribeChainHead(headers)
	if err != nil {
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
	currentBlockNumber, err := bm.currentBlockNumber()
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

func (bm *DefaultBlockMonitor) processChainHead(header *ethTypes.Header) {
	log.Info("Processing chain head", "block hash", header.Hash().String(), "block number", header.Number.String())
	var blockOrigin types.RawBlock
	err := bm.quorumClient.RPCCall(&blockOrigin, "eth_getBlockByNumber", hexutil.EncodeBig(header.Number), false)
	for err != nil {
		log.Warn("Error fetching block from Quorum", "block hash", header.Hash(), "block number", header.Number.String(), "err", err)
		time.Sleep(1 * time.Second) //TODO: return err and let caller handle?
		err = bm.quorumClient.RPCCall(&blockOrigin, "eth_getBlockByNumber", hexutil.EncodeBig(header.Number), false)
	}
	bm.newBlockChan <- bm.createBlock(&blockOrigin)
}

func (bm *DefaultBlockMonitor) createBlock(block *types.RawBlock) *types.Block {
	txs := []common.Hash{}
	for _, tx := range block.Transactions {
		txs = append(txs, common.HexToHash(tx))
	}

	timestamp := hexutil.MustDecodeUint64(block.Timestamp)
	if bm.consensus == "raft" {
		timestamp = timestamp / 1_000_000_000
	}

	return &types.Block{
		Hash:         common.HexToHash(block.Hash),
		ParentHash:   common.HexToHash(block.ParentHash),
		StateRoot:    common.HexToHash(block.StateRoot),
		TxRoot:       common.HexToHash(block.TxRoot),
		ReceiptRoot:  common.HexToHash(block.ReceiptRoot),
		Number:       hexutil.MustDecodeUint64(block.Number),
		GasLimit:     hexutil.MustDecodeUint64(block.GasLimit),
		GasUsed:      hexutil.MustDecodeUint64(block.GasUsed),
		Timestamp:    timestamp,
		ExtraData:    block.ExtraData,
		Transactions: txs,
	}
}

func (bm *DefaultBlockMonitor) currentBlockNumber() (uint64, error) {
	log.Debug("Fetching current block number")

	var currentBlockResult graphql.CurrentBlockResult
	if err := bm.quorumClient.ExecuteGraphQLQuery(&currentBlockResult, graphql.CurrentBlockQuery()); err != nil {
		return 0, err
	}

	log.Debug("Current block number found", "number", currentBlockResult.Block.Number)
	return hexutil.DecodeUint64(currentBlockResult.Block.Number)
}

func (bm *DefaultBlockMonitor) syncBlocks(start, end uint64, stopChan chan bool) *types.SyncError {
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

		var blockOrigin types.RawBlock
		err := bm.quorumClient.RPCCall(&blockOrigin, "eth_getBlockByNumber", hexutil.EncodeUint64(i), false)
		if err != nil {
			return types.NewSyncError(err.Error(), i)
		}

		select {
		case <-stopChan:
			return nil
		case bm.newBlockChan <- bm.createBlock(&blockOrigin):
		}
	}

	log.Info("Complete historical sync finished")
	return nil
}
