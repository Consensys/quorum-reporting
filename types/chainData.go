package types

import (
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type Block struct {
	*ethTypes.Block
}

type Transaction struct {
	*ethTypes.Transaction
}

type Event struct {
	*ethTypes.Log
}
