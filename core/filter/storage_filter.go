package filter

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
)

type StorageFilter struct {
	db           database.Database
	quorumClient client.Client
}

func NewStorageFilter(db database.Database, quorumClient client.Client) *StorageFilter {
	return &StorageFilter{db, quorumClient}
}

func (sf *StorageFilter) IndexStorage(blockNumber uint64, addresses []common.Address) error {
	rawStorage := make(map[common.Address]*state.DumpAccount)
	for _, address := range addresses {
		log.Printf("Pull registered contract %v storage at block %v.\n", address.Hex(), blockNumber)
		dumpAccount := &state.DumpAccount{}
		err := sf.quorumClient.RPCCall(context.Background(), &dumpAccount, "debug_dumpAddress", address, hexutil.EncodeUint64(blockNumber))
		rawStorage[address] = dumpAccount
		if err != nil {
			return err
		}
	}
	return sf.db.IndexStorage(blockNumber, rawStorage)
}
