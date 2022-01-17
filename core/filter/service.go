package filter

import (
	"math/big"
	"sync"
	"time"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/filter/token"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

//TODO: clean this type up, find a better way to pass specific methods to needed pieces
type FilterServiceDB interface {
	RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error
	RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error

	ReadTransaction(types.Hash) (*types.Transaction, error)
	ReadBlock(uint64) (*types.Block, error)
	GetLastPersistedBlockNumber() (uint64, error)
	GetLastFiltered(types.Address) (uint64, error)

	GetAddresses() ([]types.Address, error)
	GetContractABI(types.Address) (string, error)

	IndexBlocks([]types.Address, []*types.BlockWithTransactions) error
	IndexStorage(map[types.Address]*types.AccountState, uint64) error
	SetContractCreationTransaction(map[types.Hash][]types.Address) error
}

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db FilterServiceDB

	storageFilter          *StorageFilter
	contractCreationFilter *ContractCreationFilter
	erc20processor         *token.ERC20Processor
	erc721processor        *token.ERC721Processor

	// To check we have actually shut down before returning
	shutdownChan chan struct{}
	shutdownWg   sync.WaitGroup
}

func NewFilterService(db FilterServiceDB, client client.Client) *FilterService {
	return &FilterService{
		db:                     db,
		storageFilter:          NewStorageFilter(db, client),
		contractCreationFilter: NewContractCreationFilter(db, client),
		shutdownChan:           make(chan struct{}),
		erc20processor:         token.NewERC20Processor(db, client),
		erc721processor:        token.NewERC721Processor(db),
	}
}

func (fs *FilterService) Start() error {
	log.Info("Starting filter service")

	fs.shutdownWg.Add(1)

	go func() {
		// Filter tick every 2 seconds to index transactions/ storage
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		defer fs.shutdownWg.Done()
		for {
			select {
			case <-ticker.C:
				current, err := fs.db.GetLastPersistedBlockNumber()
				if err != nil {
					log.Warn("Fetching last persisted block number failed", "err", err)
					continue
				}
				log.Debug("FilterService - Last persisted block number found", "block number", current)
				lastFilteredAll, lastFiltered, err := fs.getLastFiltered(current)
				if err != nil {
					log.Warn("Fetching last filtered failed", "err", err)
					continue
				}
				log.Debug("FilterService - check Indexing for registered addresses", "current", current, "lastFiltered", lastFiltered)
				for current > lastFiltered {
					//check if we are shutting down before next round
					select {
					case <-fs.shutdownChan:
						return
					default:
					}
					//index 1000 blocks at a time
					//TODO: make configurable
					endBlock := lastFiltered + 1000
					if endBlock > current {
						endBlock = current
					}
					err := fs.index(lastFilteredAll, lastFiltered+1, endBlock)
					if err != nil {
						log.Warn("Index block failed", "lastFiltered", lastFiltered, "err", err)
						break
					}
					lastFiltered = endBlock
				}
			case <-fs.shutdownChan:
				return
			}
		}
	}()
	return nil
}

func (fs *FilterService) Stop() {
	close(fs.shutdownChan)
	fs.shutdownWg.Wait()
	fs.storageFilter.Stop()
	log.Info("Filter service stopped")
}

// getLastFiltered finds the minimum value of "lastFiltered" across all addresses
func (fs *FilterService) getLastFiltered(current uint64) (map[types.Address]uint64, uint64, error) {
	addresses, err := fs.db.GetAddresses()
	if err != nil {
		return nil, current, err
	}

	lastFiltered := make(map[types.Address]uint64)
	for _, address := range addresses {
		curLastFiltered, err := fs.db.GetLastFiltered(address)
		if err != nil {
			return nil, current, err
		}
		if curLastFiltered < current {
			current = curLastFiltered
		}
		lastFiltered[address] = curLastFiltered
	}

	return lastFiltered, current, nil
}

type IndexBatch struct {
	addresses []types.Address
	blocks    []*types.BlockWithTransactions
}

func (fs *FilterService) index(lastFiltered map[types.Address]uint64, blockNumber uint64, endBlockNumber uint64) error {
	log.Debug("Index registered address", "start-block", blockNumber, "end-block", endBlockNumber)
	indexBatches := make([]IndexBatch, 0)
	curBatch := IndexBatch{
		addresses: make([]types.Address, 0),
		blocks:    make([]*types.BlockWithTransactions, 0),
	}
	addressInBatch := make(map[types.Address]bool)
	for blockNumber <= endBlockNumber {
		// check if a new batch should be created
		oldBatch := curBatch
		for address, curLastFiltered := range lastFiltered {
			if curLastFiltered < blockNumber {
				if !addressInBatch[address] {
					addrList := curBatch.addresses
					curBatch = IndexBatch{
						addresses: []types.Address{address},
						blocks:    make([]*types.BlockWithTransactions, 0),
					}
					curBatch.addresses = append(curBatch.addresses, addrList...)
					addressInBatch[address] = true
				}
				log.Info("Indexing registered address", "address", address.Hex(), "blocknumber", blockNumber)
			}
		}
		// if new batch is created, append old batch to indexBatches
		if len(oldBatch.addresses) > 0 && len(curBatch.addresses) > len(oldBatch.addresses) {
			indexBatches = append(indexBatches, oldBatch)
		}
		// appending block to current batch
		block, err := fs.db.ReadBlock(blockNumber)
		if err != nil {
			return err
		}
		blockWithTxns, err := fs.makeBlockWithTransactions(block)
		if err != nil {
			return err
		}
		curBatch.blocks = append(curBatch.blocks, blockWithTxns)
		blockNumber++
	}
	if len(curBatch.addresses) > 0 {
		indexBatches = append(indexBatches, curBatch)
	}

	// index storage and blocks for all batches
	for _, batch := range indexBatches {
		if err := fs.processBatch(batch); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FilterService) processBatch(batch IndexBatch) error {
	log.Info("Processing batch", "start", batch.blocks[0].Number, "end", batch.blocks[len(batch.blocks)-1].Number)
	if err := fs.storageFilter.IndexStorage(batch.addresses, batch.blocks[0].Number, batch.blocks[len(batch.blocks)-1].Number); err != nil {
		return err
	}

	// if IndexStorage has an error, IndexBlocks is never called, last filtered will not be updated
	if err := fs.db.IndexBlocks(batch.addresses, batch.blocks); err != nil {
		return err
	}

	if err := fs.contractCreationFilter.ProcessBlocks(batch.addresses, batch.blocks); err != nil {
		return err
	}

	addressesWithAbi := make(map[types.Address]string)
	for _, address := range batch.addresses {
		abi, err := fs.db.GetContractABI(address)
		if err != nil {
			return err
		}
		addressesWithAbi[address] = abi
	}
	for _, b := range batch.blocks {
		if err := fs.erc20processor.ProcessBlock(addressesWithAbi, b); err != nil {
			return err
		}
		if err := fs.erc721processor.ProcessBlock(addressesWithAbi, b); err != nil {
			return err
		}
	}

	log.Info("Processed batch", "start", batch.blocks[0].Number, "end", batch.blocks[len(batch.blocks)-1].Number)
	return nil
}

func (fs *FilterService) makeBlockWithTransactions(block *types.Block) (*types.BlockWithTransactions, error) {
	allTxns := make([]*types.Transaction, 0, len(block.Transactions))
	for _, txHash := range block.Transactions {
		tx, err := fs.db.ReadTransaction(txHash)
		if err != nil {
			return nil, err
		}
		allTxns = append(allTxns, tx)
	}
	return &types.BlockWithTransactions{
		Hash:         block.Hash,
		ParentHash:   block.ParentHash,
		StateRoot:    block.StateRoot,
		TxRoot:       block.TxRoot,
		ReceiptRoot:  block.ReceiptRoot,
		Number:       block.Number,
		GasLimit:     block.GasLimit,
		GasUsed:      block.GasUsed,
		Timestamp:    block.Timestamp,
		ExtraData:    block.ExtraData,
		Transactions: allTxns,
	}, nil
}
