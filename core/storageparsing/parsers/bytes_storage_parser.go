package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"quorumengineering/quorum-report/core/storageparsing/types"
	"strconv"
)

func (p *Parser) ParseStringStorage(storageEntry string, entry types.SolidityStorageEntry) (string, error) {
	//determine if this is long or short
	arrResult, err := p.parseBytes(storageEntry, entry)
	if err != nil {
		return "", err
	}

	return string(arrResult), nil
}

func (p *Parser) ParseBytesStorage(storageEntry string, entry types.SolidityStorageEntry) ([]string, error) {
	//determine if this is long or short
	arrResult, err := p.parseBytes(storageEntry, entry)
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

func (p *Parser) parseBytes(storageEntry string, entry types.SolidityStorageEntry) ([]byte, error) {
	//determine if this is long or short
	bytes, err := ExtractFromSingleStorage(0, 1, storageEntry)
	if err != nil {
		return nil, err
	}
	isShort := (bytes[0] % 2) == 0

	var arrResult []byte
	if isShort {
		arrResult, err = p.handleShortByteArray(storageEntry, bytes[0])
		if err != nil {
			return nil, err
		}
	} else {
		arrResult, err = p.handleLargeByteArray(storageEntry, entry)
		if err != nil {
			return nil, err
		}
	}

	return arrResult, nil
}

func (p *Parser) handleShortByteArray(storageEntry string, numberOfElements byte) ([]byte, error) {
	trueNumberOfElements := uint64(numberOfElements / 2)

	// To handle a short bytes_storage entry, entries start from the left (offset 32), and take 1 byte per entry
	offset := 32 - trueNumberOfElements //skip the right-most byte, as that stores the length

	bytes, err := ExtractFromSingleStorage(offset, trueNumberOfElements, storageEntry)
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func (p *Parser) handleLargeByteArray(storageEntry string, entry types.SolidityStorageEntry) ([]byte, error) {
	sm := p.storageManager

	bytes, err := ExtractFromSingleStorage(0, 1, storageEntry)
	if err != nil {
		return nil, err
	}
	numberOfElements := p.ParseUint(bytes)
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
			currentResults, err := p.handleShortByteArray(sm.Get(currentStorageSlot), 64)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, currentResults...)

			asBig := currentStorageSlot.Big()
			asBig.Add(asBig, new(big.Int).SetUint64(1))
			currentStorageSlot = common.BigToHash(asBig)
		} else {
			currentResults, err := p.handleShortByteArray(sm.Get(currentStorageSlot), byte(resultsLeft.Uint64())*2)
			if err != nil {
				return nil, err
			}
			allResults = append(allResults, currentResults...)
		}
	}

	return allResults, nil
}
