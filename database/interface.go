package database

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

type Database interface {
	BlockDB
	TransactionDB
	//StorageDB
	IndexerDB
}

// BlockDB stores the block details for all blocks.
type BlockDB interface {
	WriteBlock(*types.Block) error
	ReadBlock(uint64) (*types.Block, error)
	GetLastPersistedBlockNumber() uint64
}

// TransactionDB stores all transactions change a contract's state.
type TransactionDB interface {
	WriteTransaction(*types.Transaction) error
	ReadTransaction(common.Hash) (*types.Transaction, error)
}

// TODO: StorageDB stores the storage trie key value pairs at all block for a contract.
type StorageDB interface {
	WriteStorage()
	ReadStorage()
}

// TODO: IndexerDB stores the location to find all transactions/ events for a contract.
type IndexerDB interface {
	IndexTransaction([]common.Address, *types.Transaction) error
	GetAllTransactionsByAddress(common.Address) []common.Hash
	GetAllEventsByAddress(common.Address) []*types.Event
}
