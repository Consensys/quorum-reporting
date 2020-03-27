package database

import (
	"errors"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

// MemoryDB is a sample in memory database.
type MemoryDB struct {
	addresses                []common.Address
	blockDB                  map[uint64]*types.Block
	txDB                     map[common.Hash]*types.Transaction
	txIndexDB                map[common.Address][]common.Hash
	lastPersistedBlockNumber uint64
	lastFiltered             map[common.Address]uint64
	mux                      sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		addresses:                []common.Address{},
		blockDB:                  make(map[uint64]*types.Block),
		txDB:                     make(map[common.Hash]*types.Transaction),
		txIndexDB:                make(map[common.Address][]common.Hash),
		lastPersistedBlockNumber: 0,
		lastFiltered:             make(map[common.Address]uint64),
	}
}

func (db *MemoryDB) AddAddresses(addresses []common.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	if len(addresses) > 0 {
		newAddresses := []common.Address{}
		for _, a := range addresses {
			isExist := false
			for _, exist := range db.addresses {
				if a == exist {
					isExist = true
					break
				}
			}
			if !isExist {
				newAddresses = append(newAddresses, a)
			}
		}
		db.indexHistory(newAddresses)
		db.addresses = append(db.addresses, newAddresses...)
	}
	return nil
}

func (db *MemoryDB) DeleteAddress(address common.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	index := -1
	for i, a := range db.addresses {
		if address == a {
			index = i
			break
		}
	}
	if index != -1 {
		err := db.removeAllIndices(address)
		if err != nil {
			return err
		}
		db.addresses = append(db.addresses[:index], db.addresses[index+1:]...)
		return nil
	}
	return errors.New("address does not exist")
}

func (db *MemoryDB) GetAddresses() []common.Address {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.addresses
}

func (db *MemoryDB) WriteBlock(block *types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()
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
		log.Printf("Block stored: number = %v, hash = %v.\n", block.Number, block.Hash.Hex())
		log.Printf("Last persisted block: %v.\n", db.lastPersistedBlockNumber)
		return nil
	}
	return errors.New("block is nil")
}

func (db *MemoryDB) ReadBlock(blockNumber uint64) (*types.Block, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if block, ok := db.blockDB[blockNumber]; ok {
		return block, nil
	}
	return nil, errors.New("block does not exist")
}

func (db *MemoryDB) GetLastPersistedBlockNumber() uint64 {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastPersistedBlockNumber
}

func (db *MemoryDB) WriteTransaction(transaction *types.Transaction) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	if transaction != nil {
		db.txDB[transaction.Hash] = transaction
		// debug printing
		log.Printf("Transaction stored: hash = %v.\n", transaction.Hash.Hex())
		return nil
	}
	return errors.New("transaction is nil")
}

func (db *MemoryDB) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if tx, ok := db.txDB[hash]; ok {
		return tx, nil
	}
	return nil, errors.New("transaction does not exist")
}

func (db *MemoryDB) IndexBlock(address common.Address, block *types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	if !db.addressIsRegistered(address) {
		return errors.New("address is not registered")
	}
	for _, txHash := range block.Transactions {
		db.indexTransaction(address, db.txDB[txHash])
	}
	db.lastFiltered[address] = db.lastPersistedBlockNumber
	return nil
}

func (db *MemoryDB) GetAllTransactionsByAddress(address common.Address) ([]common.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address], nil
}

func (db *MemoryDB) GetAllEventsByAddress(address common.Address) ([]*types.Event, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	events := []*types.Event{}
	if txs, ok := db.txIndexDB[address]; ok {
		for _, hash := range txs {
			tx := db.txDB[hash]
			events = append(events, tx.Events...)
		}
	}
	return events, nil
}

func (db *MemoryDB) GetLastFiltered(address common.Address) uint64 {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastFiltered[address]
}

// internal functions

func (db *MemoryDB) addressIsRegistered(address common.Address) bool {
	for _, a := range db.addresses {
		if address == a {
			return true
		}
	}
	return false
}

func (db *MemoryDB) indexHistory(addresses []common.Address) {
	last := uint64(0)

	for _, tx := range db.txDB {
		for _, address := range addresses {
			db.indexTransaction(address, tx)
		}
		if last < tx.BlockNumber {
			last = tx.BlockNumber
		}
	}
	for _, address := range addresses {
		db.lastFiltered[address] = db.lastPersistedBlockNumber
	}

}

func (db *MemoryDB) indexTransaction(address common.Address, tx *types.Transaction) {
	// Compare the address with tx.To and tx.CreatedContract to check if the transaction is related.
	if address == tx.To || address == tx.CreatedContract {
		// initialize list if nil
		if len(db.txIndexDB[address]) == 0 {
			db.txIndexDB[address] = []common.Hash{}
		}
		db.txIndexDB[address] = append(db.txIndexDB[address], tx.Hash)
		log.Printf("append tx %v to registered address %v.\n", tx.Hash.Hex(), address.Hex())
	}
}

func (db *MemoryDB) removeAllIndices(address common.Address) error {
	delete(db.txIndexDB, address)
	return nil
}
