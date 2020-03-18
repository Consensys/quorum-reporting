package filter

import (
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/types"
)

func createBlock(block *ethTypes.Block) *types.Block {
	return &types.Block{
		block,
	}
}

func isClosed(ch <-chan uint64) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
