package filter

import (
	"sync"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type StorageFilter struct {
	db           FilterServiceDB
	quorumClient client.Client
}

func NewStorageFilter(db FilterServiceDB, quorumClient client.Client) *StorageFilter {
	return &StorageFilter{db, quorumClient}
}

func (sf *StorageFilter) IndexStorage(addresses []types.Address, startBlockNumber, endBlockNumber uint64) error {
	var (
		wg        sync.WaitGroup
		returnErr error
	)
	for i := startBlockNumber; i <= endBlockNumber; i++ {
		wg.Add(1)
		go func(blockNumber uint64) {
			rawStorage := make(map[types.Address]*types.AccountState)
			for _, address := range addresses {
				log.Info("Pulling (indexing) contract storage", "address", address.String(), "block number", blockNumber)
				dumpAccount, err := client.DumpAddress(sf.quorumClient, address, blockNumber)
				rawStorage[address] = dumpAccount
				if err != nil {
					returnErr = err
					wg.Done()
					return
				}
			}
			if err := sf.db.IndexStorage(rawStorage, blockNumber); err != nil {
				returnErr = err
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return returnErr
}
