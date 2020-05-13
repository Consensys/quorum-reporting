package rpc

import (
	"github.com/ethereum/go-ethereum/common"
)

type SolidityStorageDocument struct {
	Storage []SolidityStorageEntry `json:"storage"`
	Types   []SolidityTypeEntry    `json:"types"`
}

type SolidityStorageEntry struct {
	Label  string `json:"label"`
	Offset uint64 `json:"offset"`
	Slot   string `json:"slot"`
	Type   string `json:"type"`
}

type SolidityTypeEntry struct {
	Encoding      string `json:"encoding"`
	Key           uint64 `json:"key"`
	Label         string `json:"label"`
	NumberOfBytes string `json:"numberOfBytes"`
	Value         string `json:"value"`
}

func parseRawStorageTwo(rawStorage map[common.Hash]string, template SolidityStorageDocument) ([]*StorageItem, error) {
	parsedStorage := []*StorageItem{}
	//for _, storageItemTemplate := range template {
	//	parsedStorageItem := &StorageItem{
	//		VarName:  storageItemTemplate.VarName,
	//		VarIndex: storageItemTemplate.VarIndex,
	//		VarType:  storageItemTemplate.VarType,
	//		Values:   nil,
	//	}
	//	hexKey := hexutil.EncodeUint64(storageItemTemplate.VarIndex)
	//	paddedHexKey := hexKey[2:]
	//	if len(paddedHexKey)%2 == 1 {
	//		paddedHexKey = "0" + paddedHexKey
	//	}
	//	key := common.BytesToHash(common.Hex2BytesFixed(paddedHexKey, 32))
	//	//fmt.Println(key)
	//	switch storageItemTemplate.VarType {
	//	case "uint256":
	//		parsedStorageItem.Value = parseUint256(rawStorage[key])
	//	case "bool":
	//		parsedStorageItem.Value = parseBool(rawStorage[key])
	//	case "string":
	//		parsedStorageItem.Value = parseString(rawStorage[key])
	//	case "address":
	//		parsedStorageItem.Value = parseAddress(rawStorage[key])
	//		// TODO: implement more types
	//	default:
	//		parsedStorageItem.Value = rawStorage[key]
	//	}
	//	parsedStorage = append(parsedStorage, parsedStorageItem)
	//}
	return parsedStorage, nil
}
