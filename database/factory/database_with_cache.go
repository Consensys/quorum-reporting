package factory

import (
	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type DatabaseWithCache struct {
	db                    database.Database
	blockCache            *lru.Cache
	transactionCache      *lru.Cache
	storageCache          *lru.Cache
	contractCreationCache *lru.Cache
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
	return &DatabaseWithCache{
		db:                    db,
		blockCache:            blockCache,
		transactionCache:      transactionCache,
		storageCache:          storageCache,
		contractCreationCache: contractCreationCache,
	}, nil
}

func (cachingDB *DatabaseWithCache) AddAddresses(addresses []common.Address) error {
	return cachingDB.db.AddAddresses(addresses)
}

func (cachingDB *DatabaseWithCache) DeleteAddress(address common.Address) error {
	return cachingDB.db.DeleteAddress(address)
}

func (cachingDB *DatabaseWithCache) GetAddresses() ([]common.Address, error) {
	return cachingDB.db.GetAddresses()
}

func (cachingDB *DatabaseWithCache) AddContractABI(address common.Address, abi string) error {
	return cachingDB.db.AddContractABI(address, abi)
}

func (cachingDB *DatabaseWithCache) GetContractABI(address common.Address) (string, error) {
	return cachingDB.db.GetContractABI(address)
}

func (cachingDB *DatabaseWithCache) WriteBlock(block *types.Block) error {
	err := cachingDB.db.WriteBlock(block)
	if err != nil {
		return err
	}
	cachingDB.blockCache.Add(block.Number, block)
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

func (cachingDB *DatabaseWithCache) GetAllTransactionsToAddress(address common.Address) ([]common.Hash, error) {
	return cachingDB.db.GetAllTransactionsToAddress(address)
}

func (cachingDB *DatabaseWithCache) GetAllTransactionsInternalToAddress(address common.Address) ([]common.Hash, error) {
	return cachingDB.db.GetAllTransactionsInternalToAddress(address)
}

func (cachingDB *DatabaseWithCache) GetAllEventsByAddress(address common.Address) ([]*types.Event, error) {
	return cachingDB.db.GetAllEventsByAddress(address)
}

func (cachingDB *DatabaseWithCache) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	return cachingDB.db.GetStorage(address, blockNumber)
}

func (cachingDB *DatabaseWithCache) GetLastFiltered(address common.Address) (uint64, error) {
	return cachingDB.db.GetLastFiltered(address)
}
