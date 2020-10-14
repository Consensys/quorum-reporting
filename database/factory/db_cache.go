package factory

import (
	"math/big"
	"sync"

	"github.com/bluele/gcache"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type DatabaseWithCache struct {
	db                    database.Database
	addressCache          map[types.Address]bool
	blockCache            gcache.Cache
	transactionCache      gcache.Cache
	storageCache          gcache.Cache
	contractCreationCache gcache.Cache
	// mutex lock
	blockMux   sync.RWMutex
	addressMux sync.RWMutex
}

func NewDatabaseWithCache(db database.Database, cacheSize int) (database.Database, error) {
	if cacheSize == 0 {
		return db, nil
	}

	existingAddresses, err := db.GetAddresses()
	if err != nil {
		return nil, err
	}
	addressCache := make(map[types.Address]bool)
	for _, address := range existingAddresses {
		addressCache[address] = true
	}
	return &DatabaseWithCache{
		db:                    db,
		addressCache:          addressCache,
		blockCache:            gcache.New(cacheSize).LRU().Build(),
		transactionCache:      gcache.New(cacheSize).LRU().Build(),
		storageCache:          gcache.New(cacheSize).LRU().Build(),
		contractCreationCache: gcache.New(cacheSize).LRU().Build(),
	}, nil
}

func (cachingDB *DatabaseWithCache) AddAddresses(addresses []types.Address) error {
	cachingDB.addressMux.Lock()
	defer cachingDB.addressMux.Unlock()
	newAddresses := []types.Address{}
	for _, address := range addresses {
		if !cachingDB.addressCache[address] {
			newAddresses = append(newAddresses, address)
		}
	}
	if len(newAddresses) > 0 {
		if err := cachingDB.db.AddAddresses(newAddresses); err != nil {
			return err
		}
		for _, newAddress := range newAddresses {
			cachingDB.addressCache[newAddress] = true
		}
	}
	return nil
}

func (cachingDB *DatabaseWithCache) AddAddressFrom(address types.Address, from uint64) error {
	cachingDB.addressMux.Lock()
	defer cachingDB.addressMux.Unlock()
	if err := cachingDB.db.AddAddressFrom(address, from); err != nil {
		return err
	}
	if !cachingDB.addressCache[address] {
		cachingDB.addressCache[address] = true
	}
	return nil
}

func (cachingDB *DatabaseWithCache) DeleteAddress(address types.Address) error {
	cachingDB.addressMux.Lock()
	defer cachingDB.addressMux.Unlock()
	if !cachingDB.addressCache[address] {
		return nil
	}
	if err := cachingDB.db.DeleteAddress(address); err != nil {
		return err
	}
	delete(cachingDB.addressCache, address)
	return nil
}

func (cachingDB *DatabaseWithCache) GetAddresses() ([]types.Address, error) {
	cachingDB.addressMux.RLock()
	defer cachingDB.addressMux.RUnlock()
	addresses := []types.Address{}
	for address, _ := range cachingDB.addressCache {
		addresses = append(addresses, address)
	}
	return addresses, nil
}

func (cachingDB *DatabaseWithCache) GetContractTemplate(address types.Address) (string, error) {
	return cachingDB.db.GetContractTemplate(address)
}

func (cachingDB *DatabaseWithCache) GetContractABI(address types.Address) (string, error) {
	return cachingDB.db.GetContractABI(address)
}

func (cachingDB *DatabaseWithCache) GetStorageLayout(address types.Address) (string, error) {
	return cachingDB.db.GetStorageLayout(address)
}

func (cachingDB *DatabaseWithCache) AddTemplate(name string, abi string, layout string) error {
	return cachingDB.db.AddTemplate(name, abi, layout)
}

func (cachingDB *DatabaseWithCache) AssignTemplate(address types.Address, name string) error {
	return cachingDB.db.AssignTemplate(address, name)
}

func (cachingDB *DatabaseWithCache) GetTemplates() ([]string, error) {
	return cachingDB.db.GetTemplates()
}

func (cachingDB *DatabaseWithCache) GetTemplateDetails(templateName string) (*types.Template, error) {
	return cachingDB.db.GetTemplateDetails(templateName)
}

func (cachingDB *DatabaseWithCache) WriteBlocks(blocks []*types.Block) error {
	cachingDB.blockMux.Lock()
	defer cachingDB.blockMux.Unlock()
	if err := cachingDB.db.WriteBlocks(blocks); err != nil {
		return err
	}
	for _, block := range blocks {
		cachingDB.blockCache.Set(block.Number, block)
	}
	return nil
}

func (cachingDB *DatabaseWithCache) ReadBlock(blockNumber uint64) (*types.Block, error) {
	if cachedBlock, err := cachingDB.blockCache.Get(blockNumber); err == nil {
		return cachedBlock.(*types.Block), nil
	}
	block, err := cachingDB.db.ReadBlock(blockNumber)
	if err != nil {
		return nil, err
	}
	cachingDB.blockCache.Set(block.Number, block)
	return block, nil
}

func (cachingDB *DatabaseWithCache) GetLastPersistedBlockNumber() (uint64, error) {
	cachingDB.blockMux.RLock()
	defer cachingDB.blockMux.RUnlock()
	return cachingDB.db.GetLastPersistedBlockNumber()
}

func (cachingDB *DatabaseWithCache) WriteTransactions(txns []*types.Transaction) error {
	err := cachingDB.db.WriteTransactions(txns)
	if err != nil {
		return err
	}
	for _, tx := range txns {
		cachingDB.transactionCache.Set(tx.Hash.String(), tx)
	}
	return nil
}

func (cachingDB *DatabaseWithCache) ReadTransaction(hash types.Hash) (*types.Transaction, error) {
	if cachedTx, err := cachingDB.transactionCache.Get(hash.String()); err == nil {
		return cachedTx.(*types.Transaction), nil
	}
	tx, err := cachingDB.db.ReadTransaction(hash)
	if err != nil {
		return nil, err
	}
	cachingDB.transactionCache.Set(tx.Hash.String(), tx)
	return tx, nil
}

func (cachingDB *DatabaseWithCache) IndexBlocks(addresses []types.Address, blocks []*types.BlockWithTransactions) error {
	return cachingDB.db.IndexBlocks(addresses, blocks)
}

func (cachingDB *DatabaseWithCache) IndexStorage(rawStorage map[types.Address]*types.AccountState, blockNumber uint64) error {
	return cachingDB.db.IndexStorage(rawStorage, blockNumber)
}

func (cachingDB *DatabaseWithCache) SetContractCreationTransaction(creationTxns map[types.Hash][]types.Address) error {
	return cachingDB.db.SetContractCreationTransaction(creationTxns)
}

func (cachingDB *DatabaseWithCache) GetContractCreationTransaction(address types.Address) (types.Hash, error) {
	if cachedHash, err := cachingDB.contractCreationCache.Get(address); err == nil {
		return cachedHash.(types.Hash), nil
	}
	hash, err := cachingDB.db.GetContractCreationTransaction(address)
	if err != nil {
		return "", err
	}
	if hash.IsEmpty() {
		cachingDB.contractCreationCache.Set(address, hash)
	}
	return hash, nil
}

func (cachingDB *DatabaseWithCache) GetAllTransactionsToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
	return cachingDB.db.GetAllTransactionsToAddress(address, options)
}

func (cachingDB *DatabaseWithCache) GetAllTransactionsInternalToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
	return cachingDB.db.GetAllTransactionsInternalToAddress(address, options)
}

func (cachingDB *DatabaseWithCache) GetAllEventsFromAddress(address types.Address, options *types.QueryOptions) ([]*types.Event, error) {
	return cachingDB.db.GetAllEventsFromAddress(address, options)
}

func (cachingDB *DatabaseWithCache) GetTransactionsToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	return cachingDB.db.GetTransactionsToAddressTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetTransactionsInternalToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	return cachingDB.db.GetTransactionsInternalToAddressTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetEventsFromAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	return cachingDB.db.GetEventsFromAddressTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetStorage(address types.Address, blockNumber uint64) (*types.StorageResult, error) {
	return cachingDB.db.GetStorage(address, blockNumber)
}

func (cachingDB *DatabaseWithCache) GetStorageWithOptions(address types.Address, options *types.PageOptions) ([]*types.StorageResult, error) {
	return cachingDB.db.GetStorageWithOptions(address, options)
}

func (cachingDB *DatabaseWithCache) GetStorageTotal(address types.Address, options *types.PageOptions) (uint64, error) {
	return cachingDB.db.GetStorageTotal(address, options)
}

func (cachingDB *DatabaseWithCache) GetStorageRanges(contract types.Address, options *types.PageOptions) ([]types.RangeResult, error) {
	return cachingDB.db.GetStorageRanges(contract, options)
}

func (cachingDB *DatabaseWithCache) GetLastFiltered(address types.Address) (uint64, error) {
	return cachingDB.db.GetLastFiltered(address)
}

func (cachingDB *DatabaseWithCache) RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
	return cachingDB.db.RecordNewERC20Balance(contract, holder, block, amount)
}

func (cachingDB *DatabaseWithCache) GetERC20Balance(contract types.Address, holder types.Address, options *types.TokenQueryOptions) (map[uint64]*big.Int, error) {
	return cachingDB.db.GetERC20Balance(contract, holder, options)
}

func (cachingDB *DatabaseWithCache) GetAllTokenHolders(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	return cachingDB.db.GetAllTokenHolders(contract, block, options)
}

func (cachingDB *DatabaseWithCache) RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error {
	return cachingDB.db.RecordERC721Token(contract, holder, block, tokenId)
}

func (cachingDB *DatabaseWithCache) ERC721TokenByTokenID(contract types.Address, block uint64, tokenId *big.Int) (*types.ERC721Token, error) {
	return cachingDB.db.ERC721TokenByTokenID(contract, block, tokenId)
}

func (cachingDB *DatabaseWithCache) ERC721TokensForAccountAtBlock(contract types.Address, holder types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	return cachingDB.db.ERC721TokensForAccountAtBlock(contract, holder, block, options)
}

func (cachingDB *DatabaseWithCache) AllERC721TokensAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	return cachingDB.db.AllERC721TokensAtBlock(contract, block, options)
}

func (cachingDB *DatabaseWithCache) AllHoldersAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	return cachingDB.db.AllHoldersAtBlock(contract, block, options)
}

func (cachingDB *DatabaseWithCache) Stop() {
	cachingDB.db.Stop()
}
