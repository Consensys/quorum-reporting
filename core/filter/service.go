package filter

import (
	"math/big"
	"quorumengineering/quorum-report/core/filter/token"
	"sync"
	"time"

	"github.com/consensys/quorum-go-utils/client"
	"github.com/consensys/quorum-go-utils/log"
	"github.com/consensys/quorum-go-utils/types"
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
	IndexBlocks([]types.Address, []*types.Block) error
	IndexStorage(map[types.Address]*types.AccountState, uint64) error
}

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db              FilterServiceDB
	storageFilter   *StorageFilter
	erc20processor  *token.ERC20Processor
	erc721processor *token.ERC721Processor

	// To check we have actually shut down before returning
	shutdownChan chan struct{}
	shutdownWg   sync.WaitGroup
}

func NewFilterService(db FilterServiceDB, client client.Client) *FilterService {
	return &FilterService{
		db:              db,
		storageFilter:   NewStorageFilter(db, client),
		shutdownChan:    make(chan struct{}),
		erc20processor:  token.NewERC20Processor(db, client),
		erc721processor: token.NewERC721Processor(db),
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
				log.Debug("Last persisted block number found", "block number", current)
				lastFilteredAll, lastFiltered, err := fs.getLastFiltered(current)
				if err != nil {
					log.Warn("Fetching last filtered failed", "err", err)
					continue
				}
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
	blocks    []*types.Block
}

func (fs *FilterService) index(lastFiltered map[types.Address]uint64, blockNumber uint64, endBlockNumber uint64) error {
	log.Debug("Index registered address", "start-block", blockNumber, "end-block", endBlockNumber)
	indexBatches := make([]IndexBatch, 0)
	curBatch := IndexBatch{
		addresses: make([]types.Address, 0),
		blocks:    make([]*types.Block, 0),
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
						blocks:    make([]*types.Block, 0),
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
		curBatch.blocks = append(curBatch.blocks, block)
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

	for _, b := range batch.blocks {
		if err := fs.erc20processor.ProcessBlock(batch.addresses, b); err != nil {
			return err
		}
		if err := fs.erc721processor.ProcessBlock(batch.addresses, b); err != nil {
			return err
		}
	}

	log.Info("Processed batch", "start", batch.blocks[0].Number, "end", batch.blocks[len(batch.blocks)-1].Number)
	return nil
}
