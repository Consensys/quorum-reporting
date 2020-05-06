package filter

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type StorageFilter struct {
	db           database.Database
	quorumClient client.Client
}

func NewStorageFilter(db database.Database, quorumClient client.Client) *StorageFilter {
	return &StorageFilter{db, quorumClient}
}

func (sf *StorageFilter) IndexStorage(block *types.Block, addresses []common.Address) error {
	rawStorage := make(map[common.Address]*state.DumpAccount)
	for _, address := range addresses {
		log.Printf("Pull registered contract %v storage at block %v.\n", address.Hex(), block.Number)
		dumpAccount, err := client.DumpAddress(sf.quorumClient, address, block.Number)
		rawStorage[address] = dumpAccount
		if err != nil {
			return err
		}
	}
	return sf.db.IndexStorage(block.Number, block.Timestamp, rawStorage)
}
