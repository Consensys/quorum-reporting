package parsers

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

func (p *Parser) ParseArray(entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) ([]interface{}, error) {
	isDynamic := namedType.Encoding == "dynamic_array"
	sizeOfArray, err := p.determineSize(entry)
	if err != nil {
		return nil, err
	}

	sizeOfElement := p.template.Types[namedType.Base].NumberOfBytes

	startSlotNumber := entry.Slot
	storageSlot := ResolveSlot(p.slotOffset.Big(), bigN(startSlotNumber))
	if isDynamic {
		storageSlot = hash(common.LeftPadBytes(storageSlot.Big().Bytes(), 32))
	}

	//build up array of fake storage elements the array has
	storageElements := make([]types.SolidityStorageEntry, 0)

	currentSlot := uint64(0)
	currentOffset := uint64(0)

	for i := uint64(0); i < sizeOfArray; i++ {
		nextEntry := types.SolidityStorageEntry{
			Label:  strconv.FormatUint(i, 10),
			Offset: currentOffset,
			Slot:   currentSlot,
			Type:   namedType.Base,
		}

		currentOffset += sizeOfElement
		if currentOffset >= 32 {
			currentSlot += roundUpTo32(currentOffset) / 32
			currentOffset = 0
		}

		storageElements = append(storageElements, nextEntry)
	}

	newTemplate := types.SolidityStorageDocument{
		Storage: storageElements,
		Types:   p.template.Types,
	}

	arrayParser := NewParser(p.storageManager, newTemplate, storageSlot)
	out, err := arrayParser.ParseRawStorage()
	if err != nil {
		return nil, err
	}
	extractedResults := make([]interface{}, 0, len(out))
	for _, result := range out {
		extractedResults = append(extractedResults, result.Value)
	}
	return extractedResults, nil
}

func (p *Parser) determineSize(storageItem types.SolidityStorageEntry) (uint64, error) {
	name := storageItem.Type

	// determine the position the size starts from
	startOfAmount := strings.LastIndex(name, ")")
	endOfAmount := strings.LastIndex(name, "_")

	size := name[startOfAmount+1 : endOfAmount]

	if size == "dyn" {
		//fetch it from the storage slot
		startSlotNumber := storageItem.Slot
		storageSlot := ResolveSlot(p.slotOffset.Big(), bigN(startSlotNumber))
		numberOfElementsHex := p.storageManager.Get(storageSlot)
		numberAsBytes := common.Hex2Bytes(numberOfElementsHex)
		numberOfElements := new(big.Int).SetBytes(numberAsBytes).Uint64()
		return numberOfElements, nil
	}
	return strconv.ParseUint(size, 10, 0)
}
