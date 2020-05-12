package filter

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db            database.Database
	storageFilter *StorageFilter
	stopFeed      event.Feed
}

func NewFilterService(db database.Database, client client.Client) *FilterService {
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
					log.Panicf("get last persisted block number failed: %v", err)
				}
				//log.Printf("Last persisted block %v.\n", current)
				lastFilteredAll, lastFiltered, err := fs.getLastFiltered(current)
				if err != nil {
					log.Panicf("get last filtered failed: %v", err)
				}
				//log.Printf("Last filtered block %v.\n", lastFiltered)
				for current > lastFiltered {
					err := fs.index(lastFilteredAll, lastFiltered+1)
					if err != nil {
						log.Panicf("index block %v failed: %v", lastFiltered, err)
					}
					lastFiltered++
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

func (fs *FilterService) index(lastFiltered map[common.Address]uint64, blockNumber uint64) error {
	block, err := fs.db.ReadBlock(blockNumber)
	if err != nil {
		return err
	}
	addresses := []common.Address{}
	for address, curLastFiltered := range lastFiltered {
		if curLastFiltered < blockNumber {
			addresses = append(addresses, address)
			log.Printf("Index registered addresses %v at block %v.\n", address.Hex(), blockNumber)
		}
	}

	if err = fs.storageFilter.IndexStorage(addresses, blockNumber); err != nil {
		return err
	}
	// if IndexStorage has an error, IndexBlock is never called, last filtered will not be updated
	return fs.db.IndexBlock(addresses, block)
}
