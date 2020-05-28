package storageparsing

import (
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

var maxElementsInRow = BigThirtyTwo

func (p *Parser) ParseStringStorage(storageEntry []byte, entry types.SolidityStorageEntry) string {
	//determine if this is long or short
	arrResult := p.parseBytes(storageEntry, entry)

	return string(arrResult)
}

func (p *Parser) ParseBytesStorage(storageEntry []byte, entry types.SolidityStorageEntry) []string {
	//determine if this is long or short
	arrResult := p.parseBytes(storageEntry, entry)

	resultBytes := make([]string, 0, len(arrResult))
	for _, resultByte := range arrResult {
		strVersion := strconv.FormatUint(uint64(resultByte), 16)
		resultBytes = append(resultBytes, strVersion)
	}
	return resultBytes
}

func (p *Parser) parseBytes(storageEntry []byte, entry types.SolidityStorageEntry) []byte {
	bytes := ExtractFromSingleStorage(0, 1, storageEntry)

	//If the LSB is 0, then the whole array fits into a single storage slot
	isShort := (bytes[0] % 2) == 0

	if isShort {
		return p.handleShortByteArray(storageEntry, bytes[0]/2)
	}
	return p.handleLargeByteArray(storageEntry, entry)
}

func (p *Parser) handleShortByteArray(storageEntry []byte, numberOfElements byte) []byte {
	// To handle a short bytes_storage entry, entries start from the left (offset 32), and take 1 byte per entry
	offset := 32 - numberOfElements //skip the right-most byte, as that stores the length

	return ExtractFromSingleStorage(uint64(offset), uint64(numberOfElements), storageEntry)
}

func (p *Parser) handleLargeByteArray(storageEntry []byte, entry types.SolidityStorageEntry) []byte {
	bytes := ExtractFromSingleStorage(0, 32, storageEntry)

	numberOfElements := p.ParseUint(bytes)
	numberOfElements.Sub(numberOfElements, BigOne).Div(numberOfElements, BigTwo)

	currentStorageSlot := hash(p.ResolveSlot(bigN(entry.Slot)))

	allResults := make([]byte, 0)
	for i := bigN(0); i.Cmp(numberOfElements) < 0; i.Add(i, maxElementsInRow) {
		//read row
		resultsLeft := new(big.Int).Sub(numberOfElements, i)
		isFullRow := resultsLeft.Cmp(maxElementsInRow) > 0

		if isFullRow {
			currentResults := p.handleShortByteArray(p.storageManager.Get(currentStorageSlot), 32)

			allResults = append(allResults, currentResults...)

			asBig := currentStorageSlot.Big()
			asBig.Add(asBig, BigOne)
			currentStorageSlot = common.BigToHash(asBig)
		} else {
			currentResults := p.handleShortByteArray(p.storageManager.Get(currentStorageSlot), byte(resultsLeft.Uint64()))
			allResults = append(allResults, currentResults...)
		}
	}

	return allResults
}
