package database

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

type Database interface {
	BlockDB
	TransactionDB
	IndexDB
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

// TODO: IndexDB stores the location to find all transactions/ events/ storage for a contract.
type IndexDB interface {
	IndexBlock(common.Address, *types.Block) error
	IndexHistory([]common.Address) error
	RemoveAllIndices(common.Address) error
	// TODO: IndexStorage stores the storage trie key value pairs at all block for a contract.
	// IndexStorage(common.Address, uint64, map[bytes32]bytes32) error
	GetAllTransactionsByAddress(common.Address) ([]common.Hash, error)
	GetAllEventsByAddress(common.Address) ([]*types.Event, error)
	GetLastFiltered(common.Address) uint64
}
