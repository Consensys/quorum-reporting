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
	// storageFilter StorageFilter
	stopChan chan interface{}
}

func NewFilterService(db database.Database) *FilterService {
	return &FilterService{
		db,
		&TransactionFilter{
			db,
		},
		make(chan interface{}),
	}
}

func (fs *FilterService) Start(lastPersisted uint64) {
	// Filter tick every second to index transactions/ storage
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			current := fs.db.GetLastPersistedBlockNumber()
			for current > lastPersisted {
				err := fs.index(lastPersisted + 1)
				if err != nil {
					// TODO: should gracefully handle error
					//log.Fatalf("index block %v failed: %v.\n", lastPersisted+1, err)
				}
				lastPersisted++
			}
		case <-fs.stopChan:
			return
		}
	}
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
	return fs.transactionFilter.IndexBlock(fs.getAddresses(), block)
	// TODO: Index storage
}

func (fs *FilterService) getAddresses() []common.Address {
	return fs.db.GetAddresses()
}
