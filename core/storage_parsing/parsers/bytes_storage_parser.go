package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"quorumengineering/quorum-report/core/storage_parsing/types"
	"strconv"
)

func ParseStringStorage(storageEntry string, sm types.StorageManager, entry types.SolidityStorageEntry) (string, error) {
	//determine if this is long or short
	arrResult, err := parseBytes(storageEntry, sm, entry)
	if err != nil {
		return "", err
	}

	return string(arrResult), nil
}

func ParseBytesStorage(storageEntry string, sm types.StorageManager, entry types.SolidityStorageEntry) ([]string, error) {
	//determine if this is long or short
	arrResult, err := parseBytes(storageEntry, sm, entry)
	if err != nil {
		return nil, err
	}

	resultBytes := make([]string, 0)
	for _, resultByte := range arrResult {
		strVersion := strconv.FormatUint(uint64(resultByte), 16)
		resultBytes = append(resultBytes, strVersion)
	}
	return resultBytes, nil
}

func parseBytes(storageEntry string, sm types.StorageManager, entry types.SolidityStorageEntry) ([]byte, error) {
	//determine if this is long or short
	bytes, err := ExtractFromSingleStorage(0, "1", storageEntry)
	if err != nil {
		return nil, err
	}
	isShort := (bytes[0] % 2) == 0

	var arrResult []byte
	if isShort {
		arrResult, err = handleShortByteArray(storageEntry, bytes[0])
		if err != nil {
			return nil, err
		}
	} else {
		arrResult, err = handleLargeByteArray(storageEntry, sm, entry)
		if err != nil {
			return nil, err
		}
	}

	return arrResult, nil
}

func handleShortByteArray(storageEntry string, numberOfElements byte) ([]byte, error) {
	trueNumberOfElements := uint64(numberOfElements / 2)

	// To handle a short bytes_storage entry, entries start from the left (offset 32), and take 1 byte per entry
	offset := 32 - trueNumberOfElements //skip the right-most byte, as that stores the length

	bytes, err := ExtractFromSingleStorage(offset, strconv.FormatUint(trueNumberOfElements, 10), storageEntry)
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func handleLargeByteArray(storageEntry string, sm types.StorageManager, entry types.SolidityStorageEntry) ([]byte, error) {
	bytes, err := ExtractFromSingleStorage(0, "1", storageEntry)
	if err != nil {
		return nil, err
	}
	numberOfElements := ParseUint(bytes)
	numberOfElements.Sub(numberOfElements, new(big.Int).SetUint64(1))
	numberOfElements.Div(numberOfElements, new(big.Int).SetUint64(2))

	currentStorageSlot := hash(entry.Slot)

	allResults := make([]byte, 0)
	maxElementsInRow := new(big.Int).SetUint64(32)
	for i := new(big.Int); i.Cmp(numberOfElements) < 0; i.Add(i, maxElementsInRow) {
		//read row
		resultsLeft := new(big.Int).Sub(numberOfElements, i)
		isFullRow := resultsLeft.Cmp(maxElementsInRow) > 0

		if isFullRow {
			currentResults, err := handleShortByteArray(sm.Get(currentStorageSlot), 64)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, currentResults...)

			asBig := currentStorageSlot.Big()
			asBig.Add(asBig, new(big.Int).SetUint64(1))
			currentStorageSlot = common.BigToHash(asBig)
		} else {
			currentResults, err := handleShortByteArray(sm.Get(currentStorageSlot), byte(resultsLeft.Uint64())*2)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, currentResults...)
		}
	}

	return allResults, nil
}
