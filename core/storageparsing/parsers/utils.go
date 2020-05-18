package parsers

import (
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

func hash(input uint64) common.Hash {
	buf := make([]byte, 32)
	n := binary.PutUvarint(buf, input)
	b := buf[:n]
	a := common.LeftPadBytes(b, 32)

	return hashBytes(a)
}

func hashBytes(input []byte) common.Hash {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(input)
	return common.BytesToHash(hasher.Sum(nil))
}
