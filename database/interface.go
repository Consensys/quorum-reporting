package database

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

type Database interface {
	AddressDB
	TemplateDB
	BlockDB
	TransactionDB
	IndexDB
	Stop()
}

// AddressDB stores registered addresses
type AddressDB interface {
	AddAddresses([]common.Address) error
	AddAddressFrom(common.Address, uint64) error
	DeleteAddress(common.Address) error
	GetAddresses() ([]common.Address, error)
	GetContractTemplate(common.Address) (string, error)
}

// TemplateDB stores contract ABI/ Storage Layout of registered address
type TemplateDB interface {
	AddTemplate(string, string, string) error
	AssignTemplate(common.Address, string) error
	GetContractABI(common.Address) (string, error)
	GetStorageLayout(common.Address) (string, error)
	GetTemplates() ([]string, error)
	GetTemplateDetails(string) (*types.Template, error)
}

// BlockDB stores the block details for all blocks.
type BlockDB interface {
	WriteBlocks([]*types.Block) error
	ReadBlock(uint64) (*types.Block, error)
	GetLastPersistedBlockNumber() (uint64, error)
}

// TransactionDB stores all transactions change a contract's state.
type TransactionDB interface {
	WriteTransactions([]*types.Transaction) error
	ReadTransaction(types.Hash) (*types.Transaction, error)
}

// IndexDB stores the location to find all transactions/ events/ storage for a contract.
type IndexDB interface {
	IndexBlocks([]common.Address, []*types.Block) error
	IndexStorage(map[common.Address]*types.AccountState, uint64) error
	GetContractCreationTransaction(common.Address) (types.Hash, error)
	GetAllTransactionsToAddress(common.Address, *types.QueryOptions) ([]types.Hash, error)
	GetTransactionsToAddressTotal(common.Address, *types.QueryOptions) (uint64, error)
	GetAllTransactionsInternalToAddress(common.Address, *types.QueryOptions) ([]types.Hash, error)
	GetTransactionsInternalToAddressTotal(common.Address, *types.QueryOptions) (uint64, error)
	GetAllEventsFromAddress(common.Address, *types.QueryOptions) ([]*types.Event, error)
	GetEventsFromAddressTotal(common.Address, *types.QueryOptions) (uint64, error)
	GetStorage(common.Address, uint64) (map[types.Hash]string, error)
	GetLastFiltered(common.Address) (uint64, error)
}
