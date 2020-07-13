package filter

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"

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

func (sf *StorageFilter) IndexStorage(addresses []common.Address, startBlockNumber, endBlockNumber uint64) error {
	var (
		wg        sync.WaitGroup
		returnErr error
	)
	for i := startBlockNumber; i <= endBlockNumber; i++ {
		wg.Add(1)
		go func(blockNumber uint64) {
			rawStorage := make(map[common.Address]*types.AccountState)
			for _, address := range addresses {
				internalAddress := types.NewAddress(address.Hex())
				log.Info("Pulling (indexing) contract storage", "address", address.String(), "block number", blockNumber)
				dumpAccount, err := client.DumpAddress(sf.quorumClient, internalAddress, blockNumber)
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
