package token

import (
	"math/big"

	"github.com/consensys/quorum-go-utils/types"
)

type TokenFilterDatabase interface {
	RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error
	RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error

	ReadTransaction(types.Hash) (*types.Transaction, error)
}
