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
	addressDB       []common.Address
	templateDB      map[common.Address]string
	abiDB           map[string]string
	storageLayoutDB map[string]string
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
		templateDB:               make(map[common.Address]string),
		abiDB:                    make(map[string]string),
		storageLayoutDB:          make(map[string]string),
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

func (db *MemoryDB) GetContractTemplate(address common.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.templateDB[address], nil
}

func (db *MemoryDB) AddContractABI(address common.Address, abi string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.templateDB[address] = address.Hex()
	db.abiDB[address.Hex()] = abi
	return nil
}

func (db *MemoryDB) GetContractABI(address common.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.abiDB[db.templateDB[address]], nil
}

func (db *MemoryDB) AddStorageLayout(address common.Address, layout string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.templateDB[address] = address.Hex()
	db.storageLayoutDB[address.Hex()] = layout
	return nil
}

func (db *MemoryDB) GetStorageLayout(address common.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.storageLayoutDB[db.templateDB[address]], nil
}

func (db *MemoryDB) AddTemplate(name string, abi string, layout string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.abiDB[name] = abi
	db.storageLayoutDB[name] = layout
	return nil
}

func (db *MemoryDB) AssignTemplate(address common.Address, name string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.templateDB[address] = name
	return nil
}

func (db *MemoryDB) GetTemplates() ([]string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	// merge abiDB and storageLayoutDB to find the full template name list
	templateNames := make(map[string]bool)
	for template, _ := range db.abiDB {
		templateNames[template] = true
	}
	for template, _ := range db.storageLayoutDB {
		templateNames[template] = true
	}
	res := make([]string, 0)
	for template, _ := range templateNames {
		res = append(res, template)
	}
	return res, nil
}

func (db *MemoryDB) GetTemplateDetails(templateName string) (*types.Template, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return &types.Template{
		TemplateName:  templateName,
		ABI:           db.abiDB[templateName],
		StorageLayout: db.storageLayoutDB[templateName],
	}, nil
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

func (db *MemoryDB) WriteBlocks(blocks []*types.Block) error {
	for _, block := range blocks {
		err := db.WriteBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
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

func (db *MemoryDB) WriteTransactions(transactions []*types.Transaction) error {
	for _, tx := range transactions {
		err := db.WriteTransaction(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MemoryDB) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if tx, ok := db.txDB[hash]; ok {
		return tx, nil
	}
	return nil, errors.New("transaction does not exist")
}

func (db *MemoryDB) IndexStorage(rawStorage map[common.Address]*state.DumpAccount, blockNumber uint64) error {
	for address, dumpAccount := range rawStorage {
		db.storageIndexDB[address].root[blockNumber] = dumpAccount.Root
		if _, ok := db.storageIndexDB[address].storage[dumpAccount.Root]; !ok {
			db.storageIndexDB[address].storage[dumpAccount.Root] = dumpAccount.Storage
		}
	}
	return nil
}

func (db *MemoryDB) IndexBlocks(addresses []common.Address, blocks []*types.Block) error {
	for _, block := range blocks {
		db.indexBlock(addresses, block)
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

func (db *MemoryDB) GetAllTransactionsToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsTo, nil
}

func (db *MemoryDB) GetTransactionsToAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.txIndexDB[address].txsTo)), nil
}

func (db *MemoryDB) GetAllTransactionsInternalToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsInternalTo, nil
}

func (db *MemoryDB) GetTransactionsInternalToAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.txIndexDB[address].txsInternalTo)), nil
}

func (db *MemoryDB) GetAllEventsFromAddress(address common.Address, options *types.QueryOptions) ([]*types.Event, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.eventIndexDB[address], nil
}

func (db *MemoryDB) GetEventsFromAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.eventIndexDB[address])), nil
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

func (db *MemoryDB) Stop() {}

// internal functions

func (db *MemoryDB) addressIsRegistered(address common.Address) bool {
	for _, a := range db.addressDB {
		if address == a {
			return true
		}
	}
	return false
}

func (db *MemoryDB) indexBlock(addresses []common.Address, block *types.Block) error {
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

	for address := range filteredAddresses {
		db.lastFiltered[address] = block.Number
	}
	return nil
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

func (db *MemoryDB) removeAllIndices(address common.Address) error {
	delete(db.txIndexDB, address)
	delete(db.eventIndexDB, address)
	delete(db.storageIndexDB, address)
	db.lastFiltered[address] = 0
	return nil
}
