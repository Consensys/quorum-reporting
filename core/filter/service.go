package filter

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
)

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db                database.Database
	transactionFilter *TransactionFilter
	lastPersisted     uint64
	// storageFilter StorageFilter
	stopChan chan interface{}
}

func NewFilterService(db database.Database, lastPersisted uint64) *FilterService {
	return &FilterService{db, &TransactionFilter{db}, lastPersisted, make(chan interface{})}
}

func (fs *FilterService) Start() error {
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
			case <-fs.stopChan:
				return
			}
		}
	}()

	return nil
}

func (fs *FilterService) Stop() {
	close(fs.stopChan)
	fmt.Println("Filter service stopped.")
}

func (fs *FilterService) index(blockNumber uint64) error {
	fmt.Printf("Start to index block %v.\n", blockNumber)
	block, err := fs.db.ReadBlock(blockNumber)
	if err != nil {
		return err
	}
	fs.transactionFilter.IndexBlock(fs.getAddresses(), block)
	return nil
	// TODO: Index storage
}

func (fs *FilterService) getAddresses() []common.Address {
	return fs.db.GetAddresses()
}
