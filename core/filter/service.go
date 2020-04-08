package filter

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db database.Database
	// storageFilter StorageFilter
	stopFeed event.Feed
}

func NewFilterService(db database.Database) *FilterService {
	return &FilterService{db: db}
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
	return fs.db.IndexBlock(addresses, block)
}
