package database

import (
	"errors"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

// MemoryDB is a sample memory database for dev only.
type MemoryDB struct {
	// registered contract data
	addressDB []common.Address
	abiDB     map[common.Address]*abi.ABI
	// blockchain data
	blockDB                  map[uint64]*types.Block
	txDB                     map[common.Hash]*types.Transaction
	lastPersistedBlockNumber uint64
	// index data
	txIndexDB      map[common.Address][]common.Hash
	eventIndexDB   map[common.Address][]*types.Event
	storageIndexDB map[common.Address]map[uint64]map[common.Hash]string
	lastFiltered   map[common.Address]uint64
	// mutex lock
	mux sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		addressDB:                []common.Address{},
		abiDB:                    make(map[common.Address]*abi.ABI),
		blockDB:                  make(map[uint64]*types.Block),
		txDB:                     make(map[common.Hash]*types.Transaction),
		txIndexDB:                make(map[common.Address][]common.Hash),
		eventIndexDB:             make(map[common.Address][]*types.Event),
		storageIndexDB:           make(map[common.Address]map[uint64]map[common.Hash]string),
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
			for _, exist := range db.addressDB {
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
		db.addressDB = append(db.addressDB, newAddresses...)
	}
	return nil
}

func (db *MemoryDB) DeleteAddress(address common.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	index := -1
	for i, a := range db.addressDB {
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
		db.addressDB = append(db.addressDB[:index], db.addressDB[index+1:]...)
		return nil
	}
	return errors.New("address does not exist")
}

func (db *MemoryDB) GetAddresses() []common.Address {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.addressDB
}

func (db *MemoryDB) AddContractABI(address common.Address, abi *abi.ABI) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.abiDB[address] = abi
	return nil
}

func (db *MemoryDB) GetContractABI(address common.Address) *abi.ABI {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.abiDB[address]
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
	// index transactions and events
	for _, txHash := range block.Transactions {
		db.indexTransaction(address, db.txDB[txHash])
	}
	// index storage
	if block.PublicState != nil {
		for address, account := range block.PublicState.Accounts {
			if len(db.storageIndexDB[address]) == 0 {
				db.storageIndexDB[address] = make(map[uint64]map[common.Hash]string)
			}
			db.storageIndexDB[address][block.Number] = account.Storage
		}
	}
	if block.PrivateState != nil {
		for address, account := range block.PrivateState.Accounts {
			if len(db.storageIndexDB[address]) == 0 {
				db.storageIndexDB[address] = make(map[uint64]map[common.Hash]string)
			}
			db.storageIndexDB[address][block.Number] = account.Storage
		}
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
	return db.eventIndexDB[address], nil
}

func (db *MemoryDB) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.storageIndexDB[address][blockNumber], nil
}

func (db *MemoryDB) GetLastFiltered(address common.Address) uint64 {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastFiltered[address]
}

// internal functions

func (db *MemoryDB) addressIsRegistered(address common.Address) bool {
	for _, a := range db.addressDB {
		if address == a {
			return true
		}
	}
	return false
}

func (db *MemoryDB) indexHistory(addresses []common.Address) {
	// index all historic transactions and events
	for _, tx := range db.txDB {
		for _, address := range addresses {
			db.indexTransaction(address, tx)
		}
	}
	// index all historic storage
	for _, block := range db.blockDB {
		if block.PublicState != nil {
			for address, account := range block.PublicState.Accounts {
				if len(db.storageIndexDB[address]) == 0 {
					db.storageIndexDB[address] = make(map[uint64]map[common.Hash]string)
				}
				db.storageIndexDB[address][block.Number] = account.Storage
			}
		}
		if block.PrivateState != nil {
			for address, account := range block.PrivateState.Accounts {
				if len(db.storageIndexDB[address]) == 0 {
					db.storageIndexDB[address] = make(map[uint64]map[common.Hash]string)
				}
				db.storageIndexDB[address][block.Number] = account.Storage
			}
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
	// Index events emitted by the given address
	for _, event := range tx.Events {
		if event.Address == address {
			// initialize list if nil
			if len(db.eventIndexDB[address]) == 0 {
				db.eventIndexDB[address] = []*types.Event{}
			}
			db.eventIndexDB[address] = append(db.eventIndexDB[address], event)
			log.Printf("append event emitted in transaction %v to registered address %v.\n", event.TransactionHash.Hex(), address.Hex())
		}
	}
}

func (db *MemoryDB) removeAllIndices(address common.Address) error {
	delete(db.txIndexDB, address)
	delete(db.eventIndexDB, address)
	delete(db.storageIndexDB, address)
	db.lastFiltered[address] = 0
	return nil
}
