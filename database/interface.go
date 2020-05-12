package database

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/types"
)

var (
	ErrNotFound = errors.New("not found")
)

type Database interface {
	AddressDB
	ABIDB
	BlockDB
	TransactionDB
	IndexDB
	Stop()
}

// AddressDB stores registered addresses
type AddressDB interface {
	AddAddresses([]common.Address) error
	DeleteAddress(common.Address) error
	GetAddresses() ([]common.Address, error)
}

// ABIDB stores contract ABI of registered address
type ABIDB interface {
	AddContractABI(common.Address, string) error
	GetContractABI(common.Address) (string, error)
}

// BlockDB stores the block details for all blocks.
type BlockDB interface {
	// Deprecated: Always use WriteBlocks
	WriteBlock(*types.Block) error
	WriteBlocks([]*types.Block) error
	ReadBlock(uint64) (*types.Block, error)
	GetLastPersistedBlockNumber() (uint64, error)
}

// TransactionDB stores all transactions change a contract's state.
type TransactionDB interface {
	// Deprecated: Always use WriteTransactions
	WriteTransaction(*types.Transaction) error
	WriteTransactions([]*types.Transaction) error
	ReadTransaction(common.Hash) (*types.Transaction, error)
}

// IndexDB stores the location to find all transactions/ events/ storage for a contract.
type IndexDB interface {
	IndexBlock([]common.Address, *types.Block) error
	IndexStorage(map[common.Address]*state.DumpAccount, uint64) error
	GetContractCreationTransaction(common.Address) (common.Hash, error)
	GetAllTransactionsToAddress(common.Address, *types.QueryOptions) ([]common.Hash, error)
	GetAllTransactionsInternalToAddress(common.Address, *types.QueryOptions) ([]common.Hash, error)
	GetAllEventsFromAddress(common.Address, *types.QueryOptions) ([]*types.Event, error)
	GetStorage(common.Address, uint64) (map[common.Hash]string, error)
	GetLastFiltered(common.Address) (uint64, error)
}
