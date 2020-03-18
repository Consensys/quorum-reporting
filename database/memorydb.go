package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

// A sample in memory database
type MemoryDB struct {
	blockDB                  map[uint64]*types.Block
	transactionDB	map[common.Address]*types.Transaction
	eventDB map[common.Address]*types.Event
	lastPersistedBlockNumber uint64
	sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		blockDB:                  make(map[uint64]*types.Block),
		lastPersistedBlockNumber: 0,
	}
}

func (db *MemoryDB) WriteBlock(block *types.Block) error {
	db.Lock()
	defer db.Unlock()
	if block != nil {
		blockNumber := block.Number
		db.blockDB[blockNumber] = block
		// update last persisted
		if blockNumber == db.lastPersistedBlockNumber+1 {
			for {
				if _, ok := db.blockDB[blockNumber+1]; ok {
					blockNumber++
				} else {
					break
				}
			}
			db.lastPersistedBlockNumber = blockNumber
		}
		// Debug printing
		fmt.Printf("Block stored: number = %v, hash = %v.\n", block.Number, block.Hash.Hex())
		fmt.Printf("Last persisted block: %v.\n", db.lastPersistedBlockNumber)
		return nil
	}
	return errors.New("block is nil")
}

func (db *MemoryDB) ReadBlock(blockNumber uint64) (*types.Block, error) {
	db.RLock()
	defer db.RUnlock()
	if block, ok := db.blockDB[blockNumber]; ok {
		return block, nil
	}
	return nil, errors.New("block does not exist")
}

func (db *MemoryDB) GetLastPersistedBlockNumber() uint64 {
	db.RLock()
	defer db.RUnlock()
	return db.lastPersistedBlockNumber
}
