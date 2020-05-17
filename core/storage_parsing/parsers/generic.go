package parsers

import (
	"github.com/ethereum/go-ethereum/common"
)

func ExtractFromSingleStorage(offset uint64, numberOfBytes uint64, storageEntry string) ([]byte, error) {
	bytes := common.Hex2Bytes(storageEntry)
	extractedBytes := bytes[32-offset-numberOfBytes : 32-offset]

	return extractedBytes, nil
}
