package filter

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type TransactionFilter struct {
	db database.Database
}

func (tf *TransactionFilter) IndexBlock(addresses []common.Address, block *types.Block) {
	for _, address := range addresses {
		tf.db.IndexBlock(address, block)
	}
}
