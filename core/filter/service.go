package filter

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db            database.Database
	blockFilter   *BlockFilter
	lastPersisted uint64
	// storageFilter StorageFilter
	stopFeed event.Feed
}

func NewFilterService(db database.Database, lastPersisted uint64) *FilterService {
	return &FilterService{
		db:            db,
		blockFilter:   &BlockFilter{db},
		lastPersisted: lastPersisted,
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
				current := fs.db.GetLastPersistedBlockNumber()
				for current > fs.lastPersisted {
					err := fs.index(fs.lastPersisted + 1)
					if err != nil {
						log.Panicf("index block %v failed: %v", fs.lastPersisted+1, err)
					}
					fs.lastPersisted++
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

func (fs *FilterService) index(blockNumber uint64) error {
	log.Printf("Start to index block %v.\n", blockNumber)
	block, err := fs.db.ReadBlock(blockNumber)
	if err != nil {
		return err
	}
	return fs.blockFilter.IndexBlock(fs.getAddresses(), block)
}

func (fs *FilterService) getAddresses() []common.Address {
	return fs.db.GetAddresses()
}
