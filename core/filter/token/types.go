package token

import (
	"math/big"

	"quorumengineering/quorum-report/types"
)

type TokenFilterDatabase interface {
	RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error
	RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error
}
