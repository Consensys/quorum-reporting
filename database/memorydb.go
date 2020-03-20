package database

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

// MemoryDB is a sample in memory database.
type MemoryDB struct {
	blockDB                  map[uint64]*types.Block
	txDB                     map[common.Hash]*types.Transaction
	txIndexDB                map[common.Address][]common.Hash
	txIndexed                map[common.Hash]bool
	lastPersistedBlockNumber uint64
	sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		blockDB:                  make(map[uint64]*types.Block),
		txDB:                     make(map[common.Hash]*types.Transaction),
		txIndexDB:                make(map[common.Address][]common.Hash),
		txIndexed:                make(map[common.Hash]bool),
		lastPersistedBlockNumber: 0,
	}
}

func (db *MemoryDB) WriteBlock(block *types.Block) error {
	db.Lock()
	defer db.Unlock()
	if block != nil {
		blockNumber := block.Number
		db.blockDB[blockNumber] = block
		// Update last persisted block number.
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
		// debug printing
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

func (db *MemoryDB) WriteTransaction(transaction *types.Transaction) error {
	db.Lock()
	defer db.Unlock()
	if transaction != nil {
		db.txDB[transaction.Hash] = transaction
		// debug printing
		fmt.Printf("Transaction stored: hash = %v.\n", transaction.Hash.Hex())
		return nil
	}
	return errors.New("transaction is nil")
}

func (db *MemoryDB) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	db.RLock()
	defer db.RUnlock()
	if tx, ok := db.txDB[hash]; ok {
		return tx, nil
	}
	return nil, errors.New("transaction does not exist")
}

func (db *MemoryDB) IndexTransaction(addresses []common.Address, tx *types.Transaction) error {
	db.Lock()
	defer db.Unlock()
	if !db.txIndexed[tx.Hash] {
		// Loop through addresses to check if the transaction is related.
		for _, a := range addresses {
			// Compare the address with tx.To and tx.CreatedContract.
			if a == tx.To || a == tx.CreatedContract {
				// initialize list if nil
				if len(db.txIndexDB[a]) == 0 {
					db.txIndexDB[a] = []common.Hash{}
				}
				db.txIndexDB[a] = append(db.txIndexDB[a], tx.Hash)
			}
		}
		db.txIndexed[tx.Hash] = true
		return nil
	}
	return errors.New("transaction is indexed already")
}

func (db *MemoryDB) GetAllTransactionsByAddress(address common.Address) []common.Hash {
	db.RLock()
	defer db.RUnlock()
	return db.txIndexDB[address]
}

func (db *MemoryDB) GetAllEventsByAddress(address common.Address) []*types.Event {
	db.RLock()
	defer db.RUnlock()
	events := []*types.Event{}
	for _, hash := range db.txIndexDB[address] {
		tx := db.txDB[hash]
		events = append(events, tx.Events...)
	}
	return events
}
