package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"quorumengineering/quorum-report/core/storage_parsing/types"
	"strconv"
	"strings"
)

func ParseArray(rawStorage map[common.Hash]string, allTypes map[string]types.SolidityTypeEntry,
	entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) error {

	sizeOfArray, err := determineSize(entry)
	if err != nil {
		return err
	}

	sizeOfType := allTypes[namedType.Base].NumberOfBytes
	size, _ := strconv.ParseUint(sizeOfType, 10, 0)

	// fixed array size
	if sizeOfArray != 0 {
		handleFixedArray(sizeOfArray, size, rawStorage, entry, namedType)
	} else {
		handleDynamicArray(size, rawStorage, entry, namedType)
	}

	return nil
}

func handleFixedArray(numberOfElements uint64, sizeOfType uint64, rawStorage map[common.Hash]string, entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) {
	totalBytesToRead := roundUp(numberOfElements * sizeOfType)

	startSlot, _ := strconv.ParseUint(entry.Slot, 10, 0)

	relevantSlots := make([]string, 0)

	allItems := ""
	for totalBytesToRead > 0 {
		currentSlot := common.BigToHash(new(big.Int).SetUint64(startSlot))
		relevantSlots = append(relevantSlots, rawStorage[currentSlot])
		if rawStorage[currentSlot] == "" {
			rawStorage[currentSlot] = "0000000000000000000000000000000000000000000000000000000000000000"
		}
		allItems = rawStorage[currentSlot] + allItems

		totalBytesToRead -= 32
		startSlot++
	}

	splitItems := make([]string, 0)
	for allItems != "" {
		nextItem := allItems[uint64(len(allItems))-(sizeOfType*2):]
		splitItems = append(splitItems, nextItem)
		allItems = allItems[:uint64(len(allItems))-(sizeOfType*2)]
	}

	splitItems = splitItems[:numberOfElements]

	//subhandler for type
}

func handleDynamicArray(sizeOfType uint64, rawStorage map[common.Hash]string, entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) {
	startSlotNumber, _ := strconv.ParseUint(entry.Slot, 10, 0)
	numberOfElementsHex := rawStorage[common.BigToHash(new(big.Int).SetUint64(startSlotNumber))]

	numberAsBytes := common.Hex2Bytes(numberOfElementsHex)
	numberOfElements := new(big.Int).SetBytes(numberAsBytes).Uint64()

	startSlotAsBig := hash(startSlotNumber).Big()

	totalBytesToRead := roundUp(numberOfElements * sizeOfType)
	relevantSlots := make([]string, 0)

	allItems := ""
	for totalBytesToRead > 0 {
		currentSlot := common.BigToHash(startSlotAsBig)
		relevantSlots = append(relevantSlots, rawStorage[currentSlot])
		if rawStorage[currentSlot] == "" {
			rawStorage[currentSlot] = "0000000000000000000000000000000000000000000000000000000000000000"
		}
		allItems = rawStorage[currentSlot] + allItems

		totalBytesToRead -= 32
		startSlotAsBig.Add(startSlotAsBig, new(big.Int).SetUint64(1))
	}

	splitItems := make([]string, 0)
	for allItems != "" {
		nextItem := allItems[uint64(len(allItems))-(sizeOfType*2):]
		splitItems = append(splitItems, nextItem)
		allItems = allItems[:uint64(len(allItems))-(sizeOfType*2)]
	}

	splitItems = splitItems[:numberOfElements]

	//handle items
}

func determineSize(storageItem types.SolidityStorageEntry) (uint64, error) {
	name := storageItem.Type

	// determine the position the size starts from
	startOfAmount := strings.LastIndex(name, ")")
	endOfAmount := strings.LastIndex(name, "_")

	size := name[startOfAmount+1 : endOfAmount]

	if size == "dyn" {
		return 0, nil
	}
	return strconv.ParseUint(size, 10, 0)
}

//rounds up to nearest multiple of 32
func roundUp(n uint64) uint64 {
	return ((n + 31) / 32) * 32
}
