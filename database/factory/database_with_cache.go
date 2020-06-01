package factory

import (
	"github.com/ethereum/go-ethereum/core/state"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type DatabaseWithCache struct {
	db                    database.Database
	addressCache          map[common.Address]bool
	blockCache            *lru.Cache
	transactionCache      *lru.Cache
	storageCache          *lru.Cache
	contractCreationCache *lru.Cache
	// mutex lock
	blockMux   sync.RWMutex
	addressMux sync.RWMutex
}

func NewDatabaseWithCache(db database.Database, cacheSize int) (database.Database, error) {
	if cacheSize == 0 {
		return db, nil
	}
	blockCache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}
	transactionCache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}
	storageCache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}
	contractCreationCache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}

	existingAddresses, err := db.GetAddresses()
	if err != nil {
		return nil, err
	}
	addressCache := make(map[common.Address]bool)
	for _, address := range existingAddresses {
		addressCache[address] = true
	}
	return &DatabaseWithCache{
		db:                    db,
		addressCache:          addressCache,
		blockCache:            blockCache,
		transactionCache:      transactionCache,
		storageCache:          storageCache,
		contractCreationCache: contractCreationCache,
	}, nil
}

func (cachingDB *DatabaseWithCache) AddAddresses(addresses []common.Address) error {
	cachingDB.addressMux.Lock()
	defer cachingDB.addressMux.Unlock()
	newAddresses := []common.Address{}
	for _, address := range addresses {
		if !cachingDB.addressCache[address] {
			newAddresses = append(newAddresses, address)
		}
	}
	if len(newAddresses) > 0 {
		err := cachingDB.db.AddAddresses(newAddresses)
		if err != nil {
			return err
		}
		for _, newAddress := range newAddresses {
			cachingDB.addressCache[newAddress] = true
		}
	}
	return nil
}

func (cachingDB *DatabaseWithCache) DeleteAddress(address common.Address) error {
	cachingDB.addressMux.Lock()
	defer cachingDB.addressMux.Unlock()
	if !cachingDB.addressCache[address] {
		return nil
	}
	err := cachingDB.db.DeleteAddress(address)
	if err != nil {
		return err
	}
	delete(cachingDB.addressCache, address)
	return nil
}

func (cachingDB *DatabaseWithCache) GetAddresses() ([]common.Address, error) {
	cachingDB.addressMux.RLock()
	defer cachingDB.addressMux.RUnlock()
	addresses := []common.Address{}
	for address, _ := range cachingDB.addressCache {
		addresses = append(addresses, address)
	}
	return addresses, nil
}

func (cachingDB *DatabaseWithCache) AddContractABI(address common.Address, abi string) error {
	return cachingDB.db.AddContractABI(address, abi)
}

func (cachingDB *DatabaseWithCache) GetContractABI(address common.Address) (string, error) {
	return cachingDB.db.GetContractABI(address)
}

func (cachingDB *DatabaseWithCache) AddStorageABI(address common.Address, abi string) error {
	return cachingDB.db.AddStorageABI(address, abi)
}

func (cachingDB *DatabaseWithCache) GetStorageABI(address common.Address) (string, error) {
	return cachingDB.db.GetStorageABI(address)
}

func (cachingDB *DatabaseWithCache) WriteBlock(block *types.Block) error {
	// make sure write block is mutual exclusive so that last persisted is updated correctly
	cachingDB.blockMux.Lock()
	defer cachingDB.blockMux.Unlock()
	err := cachingDB.db.WriteBlock(block)
	if err != nil {
		return err
	}
	cachingDB.blockCache.Add(block.Number, block)
	return nil
}

func (cachingDB *DatabaseWithCache) WriteBlocks(blocks []*types.Block) error {
	// make sure write block is mutual exclusive so that last persisted is updated correctly
	err := cachingDB.db.WriteBlocks(blocks)
	if err != nil {
		return err
	}
	for _, block := range blocks {
		cachingDB.blockCache.Add(block.Number, block)
	}
	return nil
}

func (cachingDB *DatabaseWithCache) ReadBlock(blockNumber uint64) (*types.Block, error) {
	if cachedBlock, ok := cachingDB.blockCache.Get(blockNumber); ok {
		return cachedBlock.(*types.Block), nil
	}
	block, err := cachingDB.db.ReadBlock(blockNumber)
	if err != nil {
		return nil, err
	}
	cachingDB.blockCache.Add(block.Number, block)
	return block, nil
}

func (cachingDB *DatabaseWithCache) GetLastPersistedBlockNumber() (uint64, error) {
	cachingDB.blockMux.RLock()
	defer cachingDB.blockMux.RUnlock()
	return cachingDB.db.GetLastPersistedBlockNumber()
}

func (cachingDB *DatabaseWithCache) WriteTransaction(tx *types.Transaction) error {
	err := cachingDB.db.WriteTransaction(tx)
	if err != nil {
		return err
	}
	cachingDB.transactionCache.Add(tx.Hash, tx)
	return nil
}

func (cachingDB *DatabaseWithCache) WriteTransactions(txns []*types.Transaction) error {
	err := cachingDB.db.WriteTransactions(txns)
	if err != nil {
		return err
	}
	for _, tx := range txns {
		cachingDB.transactionCache.Add(tx.Hash, tx)
	}
	return nil
}

func (cachingDB *DatabaseWithCache) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	if cachedTx, ok := cachingDB.transactionCache.Get(hash); ok {
		return cachedTx.(*types.Transaction), nil
	}
	tx, err := cachingDB.db.ReadTransaction(hash)
	if err != nil {
		return nil, err
	}
	cachingDB.transactionCache.Add(tx.Hash, tx)
	return tx, nil
}

func (cachingDB *DatabaseWithCache) IndexBlock(addresses []common.Address, block *types.Block) error {
	return cachingDB.db.IndexBlock(addresses, block)
}

func (cachingDB *DatabaseWithCache) IndexStorage(rawStorage map[common.Address]*state.DumpAccount, blockNumber uint64) error {
	return cachingDB.db.IndexStorage(rawStorage, blockNumber)
}

func (cachingDB *DatabaseWithCache) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	if cachedHash, ok := cachingDB.contractCreationCache.Get(address); ok {
		return cachedHash.(common.Hash), nil
	}
	hash, err := cachingDB.db.GetContractCreationTransaction(address)
	if err != nil {
		return common.Hash{}, err
	}
	if hash != common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000") {
		cachingDB.contractCreationCache.Add(address, hash)
	}
	return hash, nil
}

func (cachingDB *DatabaseWithCache) GetAllTransactionsToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	return cachingDB.db.GetAllTransactionsToAddress(address, options)
}

func (cachingDB *DatabaseWithCache) GetAllTransactionsInternalToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	return cachingDB.db.GetAllTransactionsInternalToAddress(address, options)
}

func (cachingDB *DatabaseWithCache) GetAllEventsFromAddress(address common.Address, options *types.QueryOptions) ([]*types.Event, error) {
	return cachingDB.db.GetAllEventsFromAddress(address, options)
}

func (cachingDB *DatabaseWithCache) GetTransactionsToAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	return cachingDB.db.GetTransactionsToAddressTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetTransactionsInternalToAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	return cachingDB.db.GetTransactionsInternalToAddressTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetEventsFromAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	return cachingDB.db.GetEventsFromAddressTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	return cachingDB.db.GetStorage(address, blockNumber)
}

func (cachingDB *DatabaseWithCache) GetLastFiltered(address common.Address) (uint64, error) {
	return cachingDB.db.GetLastFiltered(address)
}

func (cachingDB *DatabaseWithCache) Stop() {
	cachingDB.db.Stop()
}
