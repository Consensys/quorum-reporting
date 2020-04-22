package monitor

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type StorageMonitor struct {
	db database.Database
	// db is currently not in use.
	// Potentially useful if we want to store storage out of block data and only keep public/ private state root in block.
	quorumClient client.Client
}

func NewStorageMonitor(db database.Database, quorumClient client.Client) *StorageMonitor {
	return &StorageMonitor{db, quorumClient}
}

func (sm *StorageMonitor) PullStorage(block *types.Block) error {
	log.Printf("Pull all accounts storage at block %v.\n", block.Number)

	// 1. Get public state dump
	err := sm.quorumClient.RPCCall(context.Background(), &block.PublicState, "debug_dumpBlock", hexutil.EncodeUint64(block.Number))
	if err != nil {
		return err
	}
	// 2. Get private state dump
	err = sm.quorumClient.RPCCall(context.Background(), &block.PrivateState, "debug_dumpBlock", hexutil.EncodeUint64(block.Number), "private")
	if err != nil {
		return err
	}

	return nil
}
