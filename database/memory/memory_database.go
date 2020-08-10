package memory

import (
	"errors"
	"math/big"
	"sync"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

// MemoryDB is a sample memory database for dev only.
type MemoryDB struct {
	// registered contract data
	addressDB       []types.Address
	templateDB      map[types.Address]string
	abiDB           map[string]string
	storageLayoutDB map[string]string
	// blockchain data
	blockDB                  map[uint64]*types.Block
	txDB                     map[types.Hash]*types.Transaction
	lastPersistedBlockNumber uint64
	// index data
	txIndexDB      map[types.Address]*TxIndexer
	eventIndexDB   map[types.Address][]*types.Event
	storageIndexDB map[types.Address]*StorageIndexer
	lastFiltered   map[types.Address]uint64
	// mutex lock
	mux sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		addressDB:                []types.Address{},
		templateDB:               make(map[types.Address]string),
		abiDB:                    make(map[string]string),
		storageLayoutDB:          make(map[string]string),
		blockDB:                  make(map[uint64]*types.Block),
		txDB:                     make(map[types.Hash]*types.Transaction),
		txIndexDB:                make(map[types.Address]*TxIndexer),
		eventIndexDB:             make(map[types.Address][]*types.Event),
		storageIndexDB:           make(map[types.Address]*StorageIndexer),
		lastPersistedBlockNumber: 0,
		lastFiltered:             make(map[types.Address]uint64),
	}
}

type TxIndexer struct {
	contractCreationTx types.Hash
	txsTo              []types.Hash
	txsInternalTo      []types.Hash
}

func NewTxIndexer() *TxIndexer {
	return &TxIndexer{
		contractCreationTx: "",
		txsTo:              []types.Hash{},
		txsInternalTo:      []types.Hash{},
	}
}

type StorageIndexer struct {
	root    map[uint64]string
	storage map[string]map[types.Hash]string
}

func NewStorageIndexer() *StorageIndexer {
	return &StorageIndexer{
		root:    make(map[uint64]string),
		storage: make(map[string]map[types.Hash]string),
	}
}

func (db *MemoryDB) AddAddresses(addresses []types.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	if len(addresses) > 0 {
		newAddresses := []types.Address{}
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

func (db *MemoryDB) AddAddressFrom(address types.Address, from uint64) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	isExist := false
	for _, exist := range db.addressDB {
		if address == exist {
			isExist = true
			break
		}
	}
	if !isExist {
		db.txIndexDB[address] = NewTxIndexer()
		db.eventIndexDB[address] = []*types.Event{}
		db.storageIndexDB[address] = NewStorageIndexer()
		db.addressDB = append(db.addressDB, address)
		db.lastFiltered[address] = from - 1
	}
	return nil
}

func (db *MemoryDB) DeleteAddress(address types.Address) error {
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

func (db *MemoryDB) GetAddresses() ([]types.Address, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.addressDB, nil
}

func (db *MemoryDB) GetContractTemplate(address types.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.templateDB[address], nil
}

func (db *MemoryDB) GetContractABI(address types.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.abiDB[db.templateDB[address]], nil
}

func (db *MemoryDB) GetStorageLayout(address types.Address) (string, error) {
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

func (db *MemoryDB) AssignTemplate(address types.Address, name string) error {
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
	for template := range db.abiDB {
		templateNames[template] = true
	}
	for template := range db.storageLayoutDB {
		templateNames[template] = true
	}
	res := make([]string, 0)
	for template := range templateNames {
		res = append(res, template)
	}
	return res, nil
}

func (db *MemoryDB) GetTemplateDetails(templateName string) (*types.Template, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	if (db.abiDB[templateName] == "") && (db.storageLayoutDB[templateName] == "") {
		return nil, database.ErrNotFound
	}

	return &types.Template{
		TemplateName:  templateName,
		ABI:           db.abiDB[templateName],
		StorageLayout: db.storageLayoutDB[templateName],
	}, nil
}

func (db *MemoryDB) WriteBlocks(blocks []*types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	for _, block := range blocks {
		if block == nil {
			return errors.New("block is nil")
		}
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
		log.Debug("Block stored", "number", block.Number, "hash", block.Hash.String())
		log.Debug("Last persisted block", "number", db.lastPersistedBlockNumber)
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

func (db *MemoryDB) WriteTransactions(transactions []*types.Transaction) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	for _, tx := range transactions {
		if tx == nil {
			return errors.New("transaction is nil")
		}
		db.txDB[tx.Hash] = tx
		log.Debug("Transaction stored", "hash", tx.Hash.Hex())
	}
	return nil
}

func (db *MemoryDB) ReadTransaction(hash types.Hash) (*types.Transaction, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if tx, ok := db.txDB[hash]; ok {
		return tx, nil
	}
	return nil, errors.New("transaction does not exist")
}

func (db *MemoryDB) IndexStorage(rawStorage map[types.Address]*types.AccountState, blockNumber uint64) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	for address, dumpAccount := range rawStorage {
		db.storageIndexDB[address].root[blockNumber] = dumpAccount.Root.String()
		if _, ok := db.storageIndexDB[address].storage[dumpAccount.Root.String()]; !ok {
			db.storageIndexDB[address].storage[dumpAccount.Root.String()] = dumpAccount.Storage
		}
	}
	return nil
}

func (db *MemoryDB) IndexBlocks(addresses []types.Address, blocks []*types.Block) error {
	for _, block := range blocks {
		db.indexBlock(addresses, block)
	}
	return nil
}

func (db *MemoryDB) GetContractCreationTransaction(address types.Address) (types.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return "", errors.New("address is not registered")
	}
	return db.txIndexDB[address].contractCreationTx, nil
}

func (db *MemoryDB) GetAllTransactionsToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsTo, nil
}

func (db *MemoryDB) GetTransactionsToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.txIndexDB[address].txsTo)), nil
}

func (db *MemoryDB) GetAllTransactionsInternalToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsInternalTo, nil
}

func (db *MemoryDB) GetTransactionsInternalToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.txIndexDB[address].txsInternalTo)), nil
}

func (db *MemoryDB) GetAllEventsFromAddress(address types.Address, options *types.QueryOptions) ([]*types.Event, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.eventIndexDB[address], nil
}

func (db *MemoryDB) GetEventsFromAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	// TODO: MemoryDB doesn't implement query options
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.eventIndexDB[address])), nil
}

func (db *MemoryDB) GetStorage(address types.Address, blockNumber uint64) (map[types.Hash]string, error) {
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

func (db *MemoryDB) GetLastFiltered(address types.Address) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastFiltered[address], nil
}

func (db *MemoryDB) Stop() {}

// internal functions

func (db *MemoryDB) addressIsRegistered(address types.Address) bool {
	for _, a := range db.addressDB {
		if address == a {
			return true
		}
	}
	return false
}

func (db *MemoryDB) indexBlock(addresses []types.Address, block *types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	// filter out registered and unfiltered address only
	filteredAddresses := map[types.Address]bool{}
	for _, address := range addresses {
		if db.addressIsRegistered(address) && db.lastFiltered[address] < block.Number {
			filteredAddresses[address] = true
			log.Info("Index registered address ", "address", address.Hex(), "blocknumber", block.Number)
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

func (db *MemoryDB) indexTransaction(filteredAddresses map[types.Address]bool, tx *types.Transaction) {
	// Compare the address with tx.To and tx.CreatedContract to check if the transaction is related.
	if filteredAddresses[tx.CreatedContract] {
		db.txIndexDB[tx.CreatedContract].contractCreationTx = tx.Hash
		log.Debug("Indexed address of contract creation", "tx", tx.Hash.Hex(), "contract", tx.CreatedContract.Hex())
	} else if filteredAddresses[tx.To] {
		db.txIndexDB[tx.To].txsTo = append(db.txIndexDB[tx.To].txsTo, tx.Hash)
		log.Debug("Indexed tx recipient", "tx", tx.Hash.Hex(), "recipient", tx.To.Hex())
	}

	for _, internalCall := range tx.InternalCalls {
		if filteredAddresses[internalCall.To] {
			db.txIndexDB[internalCall.To].txsInternalTo = append(db.txIndexDB[internalCall.To].txsInternalTo, tx.Hash)
			log.Debug("Indexed transactions internal calls", "tx", tx.Hash.Hex(), "internal-recipient", internalCall.To.Hex())
		}
	}
	// Index events emitted by the given address
	for _, event := range tx.Events {
		addr := event.Address
		if filteredAddresses[addr] {
			db.eventIndexDB[addr] = append(db.eventIndexDB[addr], event)
			log.Debug("Indexed emitted event", "tx", event.TransactionHash.Hex(), "address", event.Address.Hex())
		}
	}
}

func (db *MemoryDB) removeAllIndices(address types.Address) error {
	delete(db.txIndexDB, address)
	delete(db.eventIndexDB, address)
	delete(db.storageIndexDB, address)
	db.lastFiltered[address] = 0
	return nil
}

func (db *MemoryDB) RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
	return nil
}

func (db *MemoryDB) GetERC20Balance(contract types.Address, holder types.Address, options *types.TokenQueryOptions) (map[uint64]*big.Int, error) {
	return nil, nil
}

func (db *MemoryDB) GetAllTokenHolders(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	return nil, nil
}

func (db *MemoryDB) RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error {
	return errors.New("")
}

func (db *MemoryDB) ERC721TokenByTokenID(contract types.Address, block uint64, tokenId *big.Int) (types.ERC721Token, error) {
	return types.ERC721Token{}, errors.New("")
}

func (db *MemoryDB) ERC721TokensForAccountAtBlock(contract types.Address, holder types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	return nil, errors.New("")
}

func (db *MemoryDB) AllERC721TokensAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	return nil, errors.New("")
}

func (db *MemoryDB) AllHoldersAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	return nil, errors.New("")
}
