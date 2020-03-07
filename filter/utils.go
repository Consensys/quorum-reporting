package filter

import (
	ethType "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/types"
)

func createBlock(header *ethType.Header) *types.Block {
	return &types.Block{
		Hash:        header.Hash(),
		ParentHash:  header.ParentHash,
		StateRoot:   header.Root,
		TxRoot:      header.TxHash,
		ReceiptRoot: header.ReceiptHash,
		Number:      header.Number,
		GasLimit:    header.GasLimit,
		GasUsed:     header.GasUsed,
		Timestamp:   header.Time,
		ExtraData:   header.Extra,
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
