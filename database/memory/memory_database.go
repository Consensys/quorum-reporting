package memory

import (
	"errors"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/types"
)

// MemoryDB is a sample memory database for dev only.
type MemoryDB struct {
	// registered contract data
	addressDB []common.Address
	abiDB     map[common.Address]string
	// blockchain data
	blockDB                  map[uint64]*types.Block
	txDB                     map[common.Hash]*types.Transaction
	lastPersistedBlockNumber uint64
	// index data
	txIndexDB      map[common.Address]*TxIndexer
	eventIndexDB   map[common.Address][]*types.Event
	storageIndexDB map[common.Address]*StorageIndexer
	lastFiltered   map[common.Address]uint64
	// mutex lock
	mux sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		addressDB:                []common.Address{},
		abiDB:                    make(map[common.Address]string),
		blockDB:                  make(map[uint64]*types.Block),
		txDB:                     make(map[common.Hash]*types.Transaction),
		txIndexDB:                make(map[common.Address]*TxIndexer),
		eventIndexDB:             make(map[common.Address][]*types.Event),
		storageIndexDB:           make(map[common.Address]*StorageIndexer),
		lastPersistedBlockNumber: 0,
		lastFiltered:             make(map[common.Address]uint64),
	}
}

type TxIndexer struct {
	contractCreationTx common.Hash
	txsTo              []common.Hash
	txsInternalTo      []common.Hash
}

func NewTxIndexer() *TxIndexer {
	return &TxIndexer{
		contractCreationTx: common.Hash{},
		txsTo:              []common.Hash{},
		txsInternalTo:      []common.Hash{},
	}
}

type StorageIndexer struct {
	root    map[uint64]string
	storage map[string]map[common.Hash]string
}

func NewStorageIndexer() *StorageIndexer {
	return &StorageIndexer{
		root:    make(map[uint64]string),
		storage: make(map[string]map[common.Hash]string),
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
				db.txIndexDB[a] = NewTxIndexer()
				db.eventIndexDB[a] = []*types.Event{}
				db.storageIndexDB[a] = NewStorageIndexer()
				newAddresses = append(newAddresses, a)
			}
		}
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

func (db *MemoryDB) GetAddresses() ([]common.Address, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.addressDB, nil
}

func (db *MemoryDB) AddContractABI(address common.Address, abi string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.abiDB[address] = abi
	return nil
}

func (db *MemoryDB) GetContractABI(address common.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.abiDB[address], nil
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

func (db *MemoryDB) GetLastPersistedBlockNumber() (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastPersistedBlockNumber, nil
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

func (db *MemoryDB) IndexBlock(addresses []common.Address, block *types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	// filter out registered and unfiltered address only
	filteredAddresses := map[common.Address]bool{}
	for _, address := range addresses {
		if db.addressIsRegistered(address) && db.lastFiltered[address] < block.Number {
			filteredAddresses[address] = true
			log.Printf("Index registered address %v at block %v.\n", address.Hex(), block.Number)
		}
	}

	// index transactions and events
	for _, txHash := range block.Transactions {
		db.indexTransaction(filteredAddresses, db.txDB[txHash])
	}

	// index public storage
	db.indexStorage(filteredAddresses, block.Number, block.PublicState)
	// index private storage
	db.indexStorage(filteredAddresses, block.Number, block.PrivateState)

	for address := range filteredAddresses {
		db.lastFiltered[address] = block.Number
	}
	return nil
}

func (db *MemoryDB) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return common.Hash{}, errors.New("address is not registered")
	}
	return db.txIndexDB[address].contractCreationTx, nil
}

func (db *MemoryDB) GetAllTransactionsToAddress(address common.Address) ([]common.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsTo, nil
}

func (db *MemoryDB) GetAllTransactionsInternalToAddress(address common.Address) ([]common.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsInternalTo, nil
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
	storageRoot, ok := db.storageIndexDB[address].root[blockNumber]
	if db.storageIndexDB[address] == nil || !ok {
		return nil, errors.New("no record found")
	}
	return db.storageIndexDB[address].storage[storageRoot], nil
}

func (db *MemoryDB) GetLastFiltered(address common.Address) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastFiltered[address], nil
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

func (db *MemoryDB) indexTransaction(filteredAddresses map[common.Address]bool, tx *types.Transaction) {
	// Compare the address with tx.To and tx.CreatedContract to check if the transaction is related.
	if filteredAddresses[tx.CreatedContract] {
		db.txIndexDB[tx.CreatedContract].contractCreationTx = tx.Hash
		log.Printf("Index contract creation tx %v of registered address %v.\n", tx.Hash.Hex(), tx.CreatedContract.Hex())
	} else if filteredAddresses[tx.To] {
		db.txIndexDB[tx.To].txsTo = append(db.txIndexDB[tx.To].txsTo, tx.Hash)
		log.Printf("Index tx %v to registered address %v.\n", tx.Hash.Hex(), tx.To.Hex())
	} else {
		for _, internalCall := range tx.InternalCalls {
			if filteredAddresses[internalCall.To] {
				db.txIndexDB[internalCall.To].txsInternalTo = append(db.txIndexDB[internalCall.To].txsInternalTo, tx.Hash)
				log.Printf("Index tx %v internal calling registered address %v.\n", tx.Hash.Hex(), internalCall.To.Hex())
			}
		}
	}
	// Index events emitted by the given address
	for _, event := range tx.Events {
		if filteredAddresses[event.Address] {
			db.eventIndexDB[event.Address] = append(db.eventIndexDB[event.Address], event)
			log.Printf("Append event emitted in transaction %v to registered address %v.\n", event.TransactionHash.Hex(), event.Address.Hex())
		}
	}
}

func (db *MemoryDB) indexStorage(filteredAddresses map[common.Address]bool, blockNumber uint64, stateDump *state.Dump) {
	if stateDump != nil {
		for address, account := range stateDump.Accounts {
			if filteredAddresses[address] {
				db.storageIndexDB[address].root[blockNumber] = account.Root
				if _, ok := db.storageIndexDB[address].storage[account.Root]; !ok {
					db.storageIndexDB[address].storage[account.Root] = account.Storage
				}
			}
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
