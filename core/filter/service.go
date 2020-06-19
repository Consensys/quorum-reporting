package filter

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

type FilterServiceDB interface {
	ReadBlock(uint64) (*types.Block, error)
	GetLastPersistedBlockNumber() (uint64, error)
	GetLastFiltered(common.Address) (uint64, error)
	GetAddresses() ([]common.Address, error)
	IndexBlocks([]common.Address, []*types.Block) error
	IndexStorage(map[common.Address]*state.DumpAccount, uint64) error
}

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db            FilterServiceDB
	storageFilter *StorageFilter
	stopFeed      event.Feed
}

func NewFilterService(db FilterServiceDB, client client.Client) *FilterService {
	return &FilterService{
		db:            db,
		storageFilter: NewStorageFilter(db, client),
	}
}

func (fs *FilterService) Start() error {
	log.Println("Start filter service...")

	stopChan, stopSubscription := fs.subscribeStopEvent()
	defer stopSubscription.Unsubscribe()

	go func() {
		// Filter tick every second to index transactions/ storage
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				current, err := fs.db.GetLastPersistedBlockNumber()
				if err != nil {
					log.Printf("get last persisted block number failed: %v\n", err)
					continue
				}
				//log.Printf("Last persisted block %v.\n", current)
				lastFilteredAll, lastFiltered, err := fs.getLastFiltered(current)
				if err != nil {
					log.Printf("get last filtered failed: %v\n", err)
					continue
				}
				//log.Printf("Last filtered block %v.\n", lastFiltered)
				for current > lastFiltered {
					//index 1000 blocks at a time
					//TODO: make configurable
					endBlock := lastFiltered + 1000
					if endBlock > current {
						endBlock = current
					}
					err := fs.index(lastFilteredAll, lastFiltered+1, endBlock)
					if err != nil {
						log.Printf("index block %v failed: %v", lastFiltered, err)
						break
					}
					lastFiltered = endBlock
				}
			case <-stopChan:
				return
			}
		}
	}()
	return nil
}

func (fs *FilterService) Stop() {
	fs.stopFeed.Send(types.StopEvent{})
	log.Println("Filter service stopped.")
}

func (fs *FilterService) subscribeStopEvent() (chan types.StopEvent, event.Subscription) {
	c := make(chan types.StopEvent)
	s := fs.stopFeed.Subscribe(c)
	return c, s
}

// getLastFiltered finds the minimum value of "lastFiltered" across all addresses
func (fs *FilterService) getLastFiltered(current uint64) (map[common.Address]uint64, uint64, error) {
	lastFiltered := make(map[common.Address]uint64)
	addresses, err := fs.db.GetAddresses()
	if err != nil {
		return lastFiltered, current, err
	}
	for _, address := range addresses {
		curLastFiltered, err := fs.db.GetLastFiltered(address)
		if err != nil {
			return lastFiltered, current, err
		}
		if curLastFiltered < current {
			current = curLastFiltered
		}
		lastFiltered[address] = curLastFiltered
	}
	return lastFiltered, current, nil
}

type IndexBatch struct {
	addresses []common.Address
	blocks    []*types.Block
}

func (fs *FilterService) index(lastFiltered map[common.Address]uint64, blockNumber uint64, endBlockNumber uint64) error {
	log.Printf("Index from block %v to block %v.\n", blockNumber, endBlockNumber)
	indexBatches := make([]IndexBatch, 0)
	curBatch := IndexBatch{
		addresses: make([]common.Address, 0),
		blocks:    make([]*types.Block, 0),
	}
	addressInBatch := make(map[common.Address]bool)
	for blockNumber <= endBlockNumber {
		// check if a new batch should be created
		oldBatch := curBatch
		for address, curLastFiltered := range lastFiltered {
			if curLastFiltered < blockNumber {
				if !addressInBatch[address] {
					addrList := curBatch.addresses
					curBatch = IndexBatch{
						addresses: []common.Address{address},
						blocks:    make([]*types.Block, 0),
					}
					curBatch.addresses = append(curBatch.addresses, addrList...)
					addressInBatch[address] = true
				}
				log.Printf("Index registered addresses %v at block %v.\n", address.Hex(), blockNumber)
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
		if err := fs.storageFilter.IndexStorage(batch.addresses, batch.blocks[0].Number, batch.blocks[len(batch.blocks)-1].Number); err != nil {
			return err
		}
		// if IndexStorage has an error, IndexBlocks is never called, last filtered will not be updated
		if err := fs.db.IndexBlocks(batch.addresses, batch.blocks); err != nil {
			return err
		}
	}
	return nil
}
