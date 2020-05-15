package rpc

import (
	"math/big"
	"quorumengineering/quorum-report/core/storage_parsing/types"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ReportingRequestTemplate []*StorageItemTemplate

type ReportingResponseTemplate struct {
	Address       common.Address `json:"address"`
	HistoricState []*ParsedState `json:"historicState"`
}

type StorageItemTemplate struct {
	VarName  string `json:"name"`
	VarIndex uint64 `json:"index"`
	VarType  string `json:"type"`
	// for map only
	KeyType string   `json:"keyType,omitempty"`
	Keys    []string `json:"keys,omitempty"`
}

type ParsedState struct {
	BlockNumber     uint64               `json:"blockNumber"`
	HistoricStorage []*types.StorageItem `json:"historicStorage"`
}

func parseRawStorage(rawStorage map[common.Hash]string, template []*StorageItemTemplate) ([]*types.StorageItem, error) {
	parsedStorage := []*types.StorageItem{}
	for _, storageItemTemplate := range template {
		//fmt.Println(storageItemTemplate)
		parsedStorageItem := &types.StorageItem{
			VarName:  storageItemTemplate.VarName,
			VarIndex: storageItemTemplate.VarIndex,
			VarType:  storageItemTemplate.VarType,
			Values:   nil,
		}
		hexKey := hexutil.EncodeUint64(storageItemTemplate.VarIndex)
		paddedHexKey := hexKey[2:]
		if len(paddedHexKey)%2 == 1 {
			paddedHexKey = "0" + paddedHexKey
		}
		key := common.BytesToHash(common.Hex2BytesFixed(paddedHexKey, 32))
		//fmt.Println(key)
		switch storageItemTemplate.VarType {
		case "uint256":
			parsedStorageItem.Value = parseUint256(rawStorage[key])
		case "bool":
			parsedStorageItem.Value = parseBool(rawStorage[key])
		case "string":
			parsedStorageItem.Value = parseString(rawStorage[key])
		case "address":
			parsedStorageItem.Value = parseAddress(rawStorage[key])
			// TODO: implement more types
		default:
			parsedStorageItem.Value = rawStorage[key]
		}
		parsedStorage = append(parsedStorage, parsedStorageItem)
	}
	return parsedStorage, nil
}

func parseUint256(raw string) *big.Int {
	if raw != "" {
		trimedHex := strings.TrimLeft(raw, "0")
		//fmt.Println(trimedHex)
		return hexutil.MustDecodeBig("0x" + trimedHex)
	}
	return big.NewInt(0)
}

func parseBool(raw string) bool {
	if raw == "01" {
		return true
	}
	return false
}

func parseString(raw string) string {
	if raw != "" {
		//fmt.Println(raw[32:])
		length := parseUint256(raw[32:]).Uint64()
		if length <= 32 {
			//fmt.Println(raw[:length])
			//fmt.Println(hexutil.MustDecode("0x" + raw[:length]))
			return string(hexutil.MustDecode("0x" + raw[:length]))
		}
		// TODO: handle string longer than 32 bytes
	}
	return ""
}

func parseAddress(raw string) common.Address {
	return common.HexToAddress(raw)
}
