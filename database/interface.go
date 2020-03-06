package database

import (
	"github.com/ethereum/go-ethereum/common"
)

// TODO: EventDB stores all event logs for a contract
type EventDB interface {
	WriteEvent()
	ReadEvent()
}

// TODO: TransactionDB stores all transactions change a contract's state
type TransactionDB interface {
	WriteTransaction()
	ReadTransaction()
}

// TODO: StorageDB stores the storage trie key value pairs at all block for a contract
type StorageDB interface {
	WriteStorage()
	ReadStorage()
}

// TODO: BlockDB stores the block details for all blocks
type BlockDB interface {
	WriteBlock()
	ReadBlock()
}

// TODO: PersistentIndexDB stores the last block number a contract has all the required data for reporting
type PersistentIndexDB interface {
	WritePersistentIndex(common.Address, uint64) error
	ReadPersistentIndex(common.Address) (uint64, error)
}