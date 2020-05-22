package storageparsing

import (
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
	"math/big"
)

var (
	BigOne       = new(big.Int).SetUint64(1)
	BigTwo       = new(big.Int).SetUint64(2)
	BigThirtyTwo = new(big.Int).SetUint64(32)
)

func hash(input *big.Int) common.Hash {
	out := common.LeftPadBytes(input.Bytes(), 32)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(out)
	return common.BytesToHash(hasher.Sum(nil))
}

func ExtractFromSingleStorage(offset uint64, numberOfBytes uint64, storageEntry []byte) []byte {
	extractedBytes := storageEntry[32-offset-numberOfBytes : 32-offset]

	return extractedBytes
}

//rounds up to nearest multiple of 32
func roundUpTo32(n uint64) uint64 {
	return ((n + 31) / 32) * 32
}

func bigN(n uint64) *big.Int {
	return new(big.Int).SetUint64(n)
}
