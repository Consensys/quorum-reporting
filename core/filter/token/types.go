package token

import (
	"math/big"

	"quorumengineering/quorum-report/types"
)

type TokenFilterDatabase interface {
	RecordNewBalance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error
	RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error

	ReadTransaction(types.Hash) (*types.Transaction, error)
}
