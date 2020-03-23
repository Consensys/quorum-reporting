package filter

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type TransactionFilter struct {
	db database.Database
}

func (tf *TransactionFilter) IndexBlock(addresses []common.Address, block *types.Block) error {
	for _, address := range addresses {
		err := tf.db.IndexBlock(address, block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tf *TransactionFilter) IndexHistory(addresses []common.Address) error {
	return tf.db.IndexHistory(addresses)
}
