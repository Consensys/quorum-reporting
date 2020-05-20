package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
	"math/big"
)

func hash(input []byte) common.Hash {
	out := common.LeftPadBytes(input, 32)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(out)
	return common.BytesToHash(hasher.Sum(nil))
}

func ExtractFromSingleStorage(offset uint64, numberOfBytes uint64, storageEntry string) ([]byte, error) {
	bytes := common.Hex2Bytes(storageEntry)
	extractedBytes := bytes[32-offset-numberOfBytes : 32-offset]

	return extractedBytes, nil
}

func ResolveSlot(offsetSlot *big.Int, givenSlot *big.Int) common.Hash {
	combined := offsetSlot.Add(offsetSlot, givenSlot)
	return common.BigToHash(combined)
}

//rounds up to nearest multiple of 32
func roundUpTo32(n uint64) uint64 {
	return ((n + 31) / 32) * 32
}

func bigN(n uint64) *big.Int {
	return new(big.Int).SetUint64(n)
}
