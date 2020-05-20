package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"quorumengineering/quorum-report/types"
	"strconv"
)

func (p *Parser) ParseStringStorage(storageEntry string, entry types.SolidityStorageEntry) string {
	//determine if this is long or short
	arrResult := p.parseBytes(storageEntry, entry)

	return string(arrResult)
}

func (p *Parser) ParseBytesStorage(storageEntry string, entry types.SolidityStorageEntry) []string {
	//determine if this is long or short
	arrResult := p.parseBytes(storageEntry, entry)

	resultBytes := make([]string, 0, len(arrResult))
	for _, resultByte := range arrResult {
		strVersion := strconv.FormatUint(uint64(resultByte), 16)
		resultBytes = append(resultBytes, strVersion)
	}
	return resultBytes
}

func (p *Parser) parseBytes(storageEntry string, entry types.SolidityStorageEntry) []byte {
	bytes := ExtractFromSingleStorage(0, 1, storageEntry)

	//If the LSB is 0, then the whole array fits into a single storage slot
	isShort := (bytes[0] % 2) == 0

	if isShort {
		return p.handleShortByteArray(storageEntry, bytes[0])
	}
	return p.handleLargeByteArray(storageEntry, entry)
}

func (p *Parser) handleShortByteArray(storageEntry string, numberOfElements byte) []byte {
	trueNumberOfElements := uint64(numberOfElements / 2)

	// To handle a short bytes_storage entry, entries start from the left (offset 32), and take 1 byte per entry
	offset := 32 - trueNumberOfElements //skip the right-most byte, as that stores the length

	return ExtractFromSingleStorage(offset, trueNumberOfElements, storageEntry)
}

func (p *Parser) handleLargeByteArray(storageEntry string, entry types.SolidityStorageEntry) []byte {
	sm := p.storageManager

	bytes := ExtractFromSingleStorage(0, 32, storageEntry)

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
			currentResults := p.handleShortByteArray(sm.Get(currentStorageSlot), 64)

			allResults = append(allResults, currentResults...)

			asBig := currentStorageSlot.Big()
			asBig.Add(asBig, types.BigOne)
			currentStorageSlot = common.BigToHash(asBig)
		} else {
			currentResults := p.handleShortByteArray(sm.Get(currentStorageSlot), byte(resultsLeft.Uint64())*2)
			allResults = append(allResults, currentResults...)
		}
	}

	return allResults
}
