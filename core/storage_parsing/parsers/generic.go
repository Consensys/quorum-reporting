package parsers

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
)

func ExtractFromSingleStorage(offset uint64, numberOfBytes string, storageEntry string) ([]byte, error) {
	bytes := common.Hex2Bytes(storageEntry)

	numBytes, err := strconv.ParseUint(numberOfBytes, 10, 0)
	if err != nil {
		return nil, err
	}

	extractedBytes := bytes[32-offset-numBytes : 32-offset]
	return extractedBytes, nil
}
