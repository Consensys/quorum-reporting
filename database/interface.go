package database

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

type Database interface {
	AddressDB
	ABIDB
	BlockDB
	TransactionDB
	IndexDB
}

// AddressDB stores registered addresses
type AddressDB interface {
	AddAddresses([]common.Address) error
	DeleteAddress(common.Address) error
	GetAddresses() ([]common.Address, error)
}

// ABIDB stores contract ABI of registered address
type ABIDB interface {
	AddContractABI(common.Address, *abi.ABI) error
	GetContractABI(common.Address) (*abi.ABI, error)
}

// BlockDB stores the block details for all blocks.
type BlockDB interface {
	WriteBlock(*types.Block) error
	ReadBlock(uint64) (*types.Block, error)
	GetLastPersistedBlockNumber() (uint64, error)
}

// TransactionDB stores all transactions change a contract's state.
type TransactionDB interface {
	WriteTransaction(*types.Transaction) error
	ReadTransaction(common.Hash) (*types.Transaction, error)
}

// IndexDB stores the location to find all transactions/ events/ storage for a contract.
type IndexDB interface {
	IndexBlock(common.Address, *types.Block) error
	GetContractCreationTransaction(common.Address) (common.Hash, error)
	GetAllTransactionsToAddress(common.Address) ([]common.Hash, error)
	GetAllTransactionsInternalToAddress(common.Address) ([]common.Hash, error)
	GetAllEventsByAddress(common.Address) ([]*types.Event, error)
	GetStorage(common.Address, uint64) (map[common.Hash]string, error)
	GetLastFiltered(common.Address) (uint64, error)
}
