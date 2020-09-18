package database

import (
	"math/big"
	"quorumengineering/quorum-report/types"
)

type Database interface {
	AddressDB
	TemplateDB
	BlockDB
	TransactionDB
	IndexDB
	TokenDB
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

	// SetContractCreationTransaction sets the transaction hash that a contract was created at
	// It accepts multiple entries at once to bulk set the contract creation txs
	SetContractCreationTransaction(creationTxns map[types.Hash][]types.Address) error
	// GetContractCreationTransaction fetches the transaction hash of the transaction that
	// the given contract address was created at
	GetContractCreationTransaction(types.Address) (types.Hash, error)

	GetAllTransactionsToAddress(types.Address, *types.QueryOptions) ([]types.Hash, error)
	GetTransactionsToAddressTotal(types.Address, *types.QueryOptions) (uint64, error)
	GetAllTransactionsInternalToAddress(types.Address, *types.QueryOptions) ([]types.Hash, error)
	GetTransactionsInternalToAddressTotal(types.Address, *types.QueryOptions) (uint64, error)
	GetAllEventsFromAddress(types.Address, *types.QueryOptions) ([]*types.Event, error)
	GetEventsFromAddressTotal(types.Address, *types.QueryOptions) (uint64, error)

	GetStorage(types.Address, uint64) (*types.StorageResult, error)
	GetStorageTotal(types.Address, *types.PageOptions) (uint64, error)
	GetStorageWithOptions(types.Address, *types.PageOptions) ([]*types.StorageResult, error)
	GetStorageRanges(types.Address, *types.PageOptions) ([]types.RangeResult, error)

	GetLastFiltered(types.Address) (uint64, error)
}

type TokenDB interface {
	RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error
	GetERC20Balance(contract types.Address, holder types.Address, options *types.TokenQueryOptions) (map[uint64]*big.Int, error)
	GetAllTokenHolders(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error)

	RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error
	ERC721TokenByTokenID(contract types.Address, block uint64, tokenId *big.Int) (types.ERC721Token, error)
	ERC721TokensForAccountAtBlock(contract types.Address, holder types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error)
	AllERC721TokensAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error)
	AllHoldersAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error)
}
