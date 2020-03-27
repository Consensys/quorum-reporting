package filter

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/database"
)

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db                database.Database
	transactionFilter *TransactionFilter
	lastPersisted     uint64
	// storageFilter StorageFilter
	stopFeed event.Feed
}

// to signal all watches when service is stopped
type stopEvent struct {
}

func (fs *FilterService) subscribeStopEvent() (chan stopEvent, event.Subscription) {
	c := make(chan stopEvent)
	s := fs.stopFeed.Subscribe(c)
	return c, s
}

func NewFilterService(db database.Database, lastPersisted uint64) *FilterService {
	return &FilterService{
		db:                db,
		transactionFilter: &TransactionFilter{db},
		lastPersisted:     lastPersisted}
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
						// TODO: should gracefully handle error
						//log.Fatalf("index block %v failed: %v.\n", lastPersisted+1, err)
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
	fs.stopFeed.Send(stopEvent{})
	log.Println("Filter service stopped.")
}

func (fs *FilterService) index(blockNumber uint64) error {
	log.Printf("Start to index block %v.\n", blockNumber)
	block, err := fs.db.ReadBlock(blockNumber)
	if err != nil {
		return err
	}
	// TODO: Unhandled error
	fs.transactionFilter.IndexBlock(fs.getAddresses(), block)
	return nil
	// TODO: Index storage
}

func (fs *FilterService) getAddresses() []common.Address {
	return fs.db.GetAddresses()
}
