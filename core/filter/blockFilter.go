package filter

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type BlockFilter struct {
	db database.IndexDB
}

func (bf *BlockFilter) IndexBlock(addresses []common.Address, block *types.Block) error {
	for _, address := range addresses {
		err := bf.db.IndexBlock(address, block)
		if err != nil {
			return err
		}
	}
	return nil
}
