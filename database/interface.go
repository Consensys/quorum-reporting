package database

import (
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
	AddAddresses([]types.Address) error
	AddAddressFrom(types.Address, uint64) error
	DeleteAddress(types.Address) error
	GetAddresses() ([]types.Address, error)
	GetContractTemplate(types.Address) (string, error)
}

// TemplateDB stores contract ABI/ Storage Layout of registered address
type TemplateDB interface {
	AddTemplate(string, string, string) error
	AssignTemplate(types.Address, string) error
	GetContractABI(types.Address) (string, error)
	GetStorageLayout(types.Address) (string, error)
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
	IndexBlocks([]types.Address, []*types.Block) error
	IndexStorage(map[types.Address]*types.AccountState, uint64) error
	GetContractCreationTransaction(types.Address) (types.Hash, error)
	GetAllTransactionsToAddress(types.Address, *types.QueryOptions) ([]types.Hash, error)
	GetTransactionsToAddressTotal(types.Address, *types.QueryOptions) (uint64, error)
	GetAllTransactionsInternalToAddress(types.Address, *types.QueryOptions) ([]types.Hash, error)
	GetTransactionsInternalToAddressTotal(types.Address, *types.QueryOptions) (uint64, error)
	GetAllEventsFromAddress(types.Address, *types.QueryOptions) ([]*types.Event, error)
	GetEventsFromAddressTotal(types.Address, *types.QueryOptions) (uint64, error)
	GetStorage(types.Address, uint64) (map[types.Hash]string, error)
	GetLastFiltered(types.Address) (uint64, error)
}
