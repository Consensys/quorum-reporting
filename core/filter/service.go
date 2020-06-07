package filter

import (
	"log"
	"sync"
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
		ticker := time.NewTicker(time.Second)
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

func (fs *FilterService) index(lastFiltered map[common.Address]uint64, blockNumber uint64, endBlockNumber uint64) error {
	//read the block range that we are indexing
	allBlocks := make([]*types.Block, 0)
	for i := blockNumber; i <= endBlockNumber; i++ {
		block, err := fs.db.ReadBlock(i)
		if err != nil {
			return err
		}
		allBlocks = append(allBlocks, block)
	}

	// find all the addresses we may need to index
	// this may result in some extra indexing if an address has had some of the block range
	// indexed before
	addresses := []common.Address{}
	for address, curLastFiltered := range lastFiltered {
		if curLastFiltered < endBlockNumber {
			addresses = append(addresses, address)
			log.Printf("Index registered addresses %v at block %v.\n", address.Hex(), blockNumber)
		}
	}

	var (
		wg        sync.WaitGroup
		returnErr error
	)

	for _, block := range allBlocks {
		wg.Add(1)
		go func(block *types.Block) {
			if err := fs.storageFilter.IndexStorage(addresses, block.Number); err != nil {
				returnErr = err
			}
			wg.Done()
		}(block)
	}
	wg.Wait()

	if returnErr != nil {
		return returnErr
	}

	// if IndexStorage has an error, IndexBlock is never called, last filtered will not be updated
	return fs.db.IndexBlocks(addresses, allBlocks)
}
