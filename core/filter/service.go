package filter

import (
	"log"
	"time"

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
	// Filter tick every second to index transactions/ storage
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				current, err := fs.db.GetLastPersistedBlockNumber()
				if err != nil {
					log.Panicf("get last persisted block number failed: %v", err)
				}
				lastFiltered, err := fs.getLastFiltered(current)
				if err != nil {
					log.Panicf("get last filtered failed: %v", err)
				}
				for current > lastFiltered {
					err := fs.index(lastFiltered + 1)
					if err != nil {
						log.Printf("index block %v failed: %v", lastFiltered, err)

						//end the loop early, forgetting about any iterations left
						//they will be picked up in the next round
						break
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
func (fs *FilterService) getLastFiltered(lastFiltered uint64) (uint64, error) {
	addresses, err := fs.db.GetAddresses()
	if err != nil {
		return 0, err
	}
	for _, address := range addresses {
		curLastFiltered, err := fs.db.GetLastFiltered(address)
		if err != nil {
			return 0, err
		}
		if curLastFiltered < lastFiltered {
			lastFiltered = curLastFiltered
		}
	}
	return lastFiltered, nil
}

func (fs *FilterService) index(blockNumber uint64) error {
	block, err := fs.db.ReadBlock(blockNumber)
	if err != nil {
		return err
	}
	addresses, err := fs.db.GetAddresses()
	if err != nil {
		return err
	}
	err = fs.storageFilter.IndexStorage(blockNumber, addresses)
	if err != nil {
		return err
	}
	// if IndexStorage has an error, IndexBlock is never called, last filtered will not be updated
	return fs.db.IndexBlock(addresses, block)
}
