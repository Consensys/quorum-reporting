package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"quorumengineering/quorum-report/types"
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

	resultBytes := make([]string, 0, len(arrResult))
	for _, resultByte := range arrResult {
		strVersion := strconv.FormatUint(uint64(resultByte), 16)
		resultBytes = append(resultBytes, strVersion)
	}
	return resultBytes, nil
}

func (p *Parser) parseBytes(storageEntry string, entry types.SolidityStorageEntry) ([]byte, error) {
	bytes, err := ExtractFromSingleStorage(0, 1, storageEntry)
	if err != nil {
		return nil, err
	}

	//If the LSB is 0, then the whole array fits into a single storage slot
	isShort := (bytes[0] % 2) == 0

	if isShort {
		return p.handleShortByteArray(storageEntry, bytes[0])
	}
	return p.handleLargeByteArray(storageEntry, entry)
}

func (p *Parser) handleShortByteArray(storageEntry string, numberOfElements byte) ([]byte, error) {
	trueNumberOfElements := uint64(numberOfElements / 2)

	// To handle a short bytes_storage entry, entries start from the left (offset 32), and take 1 byte per entry
	offset := 32 - trueNumberOfElements //skip the right-most byte, as that stores the length

	return ExtractFromSingleStorage(offset, trueNumberOfElements, storageEntry)
}

func (p *Parser) handleLargeByteArray(storageEntry string, entry types.SolidityStorageEntry) ([]byte, error) {
	sm := p.storageManager

	bytes, err := ExtractFromSingleStorage(0, 32, storageEntry)
	if err != nil {
		return nil, err
	}
	numberOfElements := p.ParseUint(bytes)
	numberOfElements.Sub(numberOfElements, types.BigOne)
	numberOfElements.Div(numberOfElements, types.BigTwo)

	currentStorageSlot := hash(bigN(entry.Slot).Bytes())

	allResults := make([]byte, 0)
	maxElementsInRow := types.BigThirtyTwo
	for i := bigN(0); i.Cmp(numberOfElements) < 0; i.Add(i, maxElementsInRow) {
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
			asBig.Add(asBig, types.BigOne)
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
