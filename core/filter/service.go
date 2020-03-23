package filter

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
)

// FilterService filters transactions and storage based on registered address list.
type FilterService struct {
	db                database.Database
	addresses         []common.Address
	transactionFilter *TransactionFilter
	// storageFilter StorageFilter
	stopChan chan interface{}
}

func NewFilterService(db database.Database, addresses []common.Address) *FilterService {
	return &FilterService{
		db,
		addresses,
		&TransactionFilter{
			db,
		},
		make(chan interface{}),
	}
}

func (fs *FilterService) Start() {
	// Index historical transactions
	lastPersisted := fs.db.GetLastPersistedBlockNumber()

	fmt.Println("Start to index history.")
	err := fs.transactionFilter.IndexHistory(fs.addresses)
	if err != nil {
		// TODO: should gracefully handle error (if quorum node is down, reconnect?)
		log.Fatalf("index history failed: %v.\n", err)
	}
	// TODO: Index storage

	// Filter tick every second to index transactions/ storage
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			current := fs.db.GetLastPersistedBlockNumber()
			for current > lastPersisted {
				err := fs.index(lastPersisted + 1)
				if err != nil {
					// TODO: should gracefully handle error (if quorum node is down, reconnect?)
					log.Fatalf("index block %v failed: %v.\n", lastPersisted+1, err)
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
	return fs.transactionFilter.IndexBlock(fs.addresses, block)
	// TODO: Index storage
}
