package storageparsing

import (
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/core/storageparsing/parsers"
	"quorumengineering/quorum-report/core/storageparsing/types"
	types2 "quorumengineering/quorum-report/core/storageparsing/types"
)

var (
	intPrefix      = "t_int"
	uintPrefix     = "t_uint"
	boolPrefix     = "t_bool"
	addressPrefix  = "t_address"
	contractPrefix = "t_contract"
	bytesPrefix    = "t_bytes"
	enumPrefix     = "t_enum"

	bytesStoragePrefix = "t_bytes_storage"
	stringPrefix       = "t_string_storage"

	arrayPrefix  = "t_array"
	structPrefix = "t_struct"
)

func ParseRawStorage(rawStorage map[common.Hash]string, template types.SolidityStorageDocument) ([]*types.StorageItem, error) {
	parsedStorage := []*types.StorageItem{}

	sort.Sort(template.Storage)

	storageManager := types2.NewDefaultStorageHandler(rawStorage)

	//sort the storage based on
	for _, storageItem := range template.Storage {
		namedType := template.Types[storageItem.Type]
		startingSlot := common.BigToHash(new(big.Int).SetUint64(storageItem.Slot))
		directStorageSlot := storageManager.Get(startingSlot) //the storage this variable uses by its "Slot"

		var result interface{}

		switch {
		case strings.HasPrefix(storageItem.Type, intPrefix):
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = parsers.ParseInt(bytes).String()
		case strings.HasPrefix(storageItem.Type, uintPrefix):
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = parsers.ParseUint(bytes).String()
		case strings.HasPrefix(storageItem.Type, boolPrefix):
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = bytes[0] == 1
		case strings.HasPrefix(storageItem.Type, addressPrefix):
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = common.BytesToAddress(bytes).String()
		case strings.HasPrefix(storageItem.Type, contractPrefix): //TODO: recurse down contracts?
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = common.BytesToAddress(bytes).String()
		case strings.HasPrefix(storageItem.Type, bytesPrefix) && !strings.HasPrefix(storageItem.Type, bytesStoragePrefix):
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = "0x" + common.Bytes2Hex(bytes)
		case strings.HasPrefix(storageItem.Type, enumPrefix):
			bytes, err := parsers.ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
			if err != nil {
				return nil, err
			}
			result = uint64(bytes[0])
		case strings.HasPrefix(storageItem.Type, bytesStoragePrefix):
			bytes, err := parsers.ParseBytesStorage(directStorageSlot, storageManager, storageItem)
			if err != nil {
				return nil, err
			}
			result = bytes
		case strings.HasPrefix(storageItem.Type, stringPrefix):
			str, err := parsers.ParseStringStorage(directStorageSlot, storageManager, storageItem)
			if err != nil {
				return nil, err
			}
			result = str
			//case strings.HasPrefix(storageItem.Type, arrayPrefix):
			//	parsers.ParseArray(storageManager, Template.Types, storageItem, namedType)
		case strings.HasPrefix(storageItem.Type, structPrefix):
			parsers.ParseStruct(storageManager, storageItem, namedType)
		}

		if result != nil {
			parsedStorageItem := &types.StorageItem{
				VarName:  storageItem.Label,
				VarIndex: 0,
				VarType:  namedType.Label,
				Value:    result,
			}
			parsedStorage = append(parsedStorage, parsedStorageItem)
		}
	}

	return parsedStorage, nil
}
