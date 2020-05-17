package parsers

import (
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

func ExtractFromSingleStorage(offset uint64, numberOfBytes string, storageEntry string) ([]byte, error) {
	if storageEntry == "" {
		storageEntry = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	bytes := common.Hex2Bytes(storageEntry)

	numBytes, err := strconv.ParseUint(numberOfBytes, 10, 0)
	if err != nil {
		return nil, err
	}

	extractedBytes := bytes[32-offset-numBytes : 32-offset]
	return extractedBytes, nil
}

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
