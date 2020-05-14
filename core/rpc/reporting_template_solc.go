package rpc

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

var (
	TemplateString = "{\"storage\":[{\"astId\":3,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"a\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint256\"},{\"astId\":5,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"b\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint8\"},{\"astId\":7,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"c\",\"offset\":1,\"slot\":\"1\",\"type\":\"t_uint8\"},{\"astId\":9,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_int256\"},{\"astId\":11,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d2\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_int256\"},{\"astId\":13,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d3\",\"offset\":0,\"slot\":\"4\",\"type\":\"t_int8\"},{\"astId\":15,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d4\",\"offset\":1,\"slot\":\"4\",\"type\":\"t_int24\"},{\"astId\":17,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"e\",\"offset\":4,\"slot\":\"4\",\"type\":\"t_bool\"},{\"astId\":19,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"f\",\"offset\":5,\"slot\":\"4\",\"type\":\"t_address\"},{\"astId\":21,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"g\",\"offset\":0,\"slot\":\"5\",\"type\":\"t_contract(SimpleStorage)275\"},{\"astId\":23,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h1\",\"offset\":20,\"slot\":\"5\",\"type\":\"t_bytes1\"},{\"astId\":25,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h2\",\"offset\":21,\"slot\":\"5\",\"type\":\"t_bytes1\"},{\"astId\":27,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h3\",\"offset\":22,\"slot\":\"5\",\"type\":\"t_bytes2\"},{\"astId\":29,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h4\",\"offset\":0,\"slot\":\"6\",\"type\":\"t_bytes31\"},{\"astId\":31,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h5\",\"offset\":0,\"slot\":\"7\",\"type\":\"t_bytes32\"},{\"astId\":38,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"choice\",\"offset\":0,\"slot\":\"8\",\"type\":\"t_enum(ActionChoices)36\"},{\"astId\":40,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i1\",\"offset\":0,\"slot\":\"9\",\"type\":\"t_bytes_storage\"},{\"astId\":42,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i2\",\"offset\":0,\"slot\":\"10\",\"type\":\"t_string_storage\"},{\"astId\":44,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i5\",\"offset\":0,\"slot\":\"11\",\"type\":\"t_string_storage\"},{\"astId\":47,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h6\",\"offset\":0,\"slot\":\"12\",\"type\":\"t_array(t_bytes1)dyn_storage\"},{\"astId\":51,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h7\",\"offset\":0,\"slot\":\"13\",\"type\":\"t_array(t_bytes1)10_storage\"},{\"astId\":54,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i3\",\"offset\":0,\"slot\":\"14\",\"type\":\"t_array(t_address)dyn_storage\"},{\"astId\":57,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i4\",\"offset\":0,\"slot\":\"15\",\"type\":\"t_array(t_contract(SimpleStorage)275)dyn_storage\"},{\"astId\":64,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"funder1\",\"offset\":0,\"slot\":\"16\",\"type\":\"t_struct(Funder)62_storage\"},{\"astId\":68,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"fundersFixed\",\"offset\":0,\"slot\":\"18\",\"type\":\"t_array(t_struct(Funder)62_storage)2_storage\"},{\"astId\":71,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"fundersDyn\",\"offset\":0,\"slot\":\"22\",\"type\":\"t_array(t_struct(Funder)62_storage)dyn_storage\"},{\"astId\":75,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"map\",\"offset\":0,\"slot\":\"23\",\"type\":\"t_mapping(t_uint256,t_uint256)\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_array(t_address)dyn_storage\":{\"base\":\"t_address\",\"encoding\":\"dynamic_array\",\"label\":\"address[]\",\"numberOfBytes\":\"32\"},\"t_array(t_bytes1)10_storage\":{\"base\":\"t_bytes1\",\"encoding\":\"inplace\",\"label\":\"bytes1[10]\",\"numberOfBytes\":\"32\"},\"t_array(t_bytes1)dyn_storage\":{\"base\":\"t_bytes1\",\"encoding\":\"dynamic_array\",\"label\":\"bytes1[]\",\"numberOfBytes\":\"32\"},\"t_array(t_contract(SimpleStorage)275)dyn_storage\":{\"base\":\"t_contract(SimpleStorage)275\",\"encoding\":\"dynamic_array\",\"label\":\"contract SimpleStorage[]\",\"numberOfBytes\":\"32\"},\"t_array(t_struct(Funder)62_storage)2_storage\":{\"base\":\"t_struct(Funder)62_storage\",\"encoding\":\"inplace\",\"label\":\"struct SimpleStorage.Funder[2]\",\"numberOfBytes\":\"128\"},\"t_array(t_struct(Funder)62_storage)dyn_storage\":{\"base\":\"t_struct(Funder)62_storage\",\"encoding\":\"dynamic_array\",\"label\":\"struct SimpleStorage.Funder[]\",\"numberOfBytes\":\"32\"},\"t_bool\":{\"encoding\":\"inplace\",\"label\":\"bool\",\"numberOfBytes\":\"1\"},\"t_bytes1\":{\"encoding\":\"inplace\",\"label\":\"bytes1\",\"numberOfBytes\":\"1\"},\"t_bytes2\":{\"encoding\":\"inplace\",\"label\":\"bytes2\",\"numberOfBytes\":\"2\"},\"t_bytes31\":{\"encoding\":\"inplace\",\"label\":\"bytes31\",\"numberOfBytes\":\"31\"},\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_bytes_storage\":{\"encoding\":\"bytes\",\"label\":\"bytes\",\"numberOfBytes\":\"32\"},\"t_contract(SimpleStorage)275\":{\"encoding\":\"inplace\",\"label\":\"contract SimpleStorage\",\"numberOfBytes\":\"20\"},\"t_enum(ActionChoices)36\":{\"encoding\":\"inplace\",\"label\":\"enum SimpleStorage.ActionChoices\",\"numberOfBytes\":\"1\"},\"t_int24\":{\"encoding\":\"inplace\",\"label\":\"int24\",\"numberOfBytes\":\"3\"},\"t_int256\":{\"encoding\":\"inplace\",\"label\":\"int256\",\"numberOfBytes\":\"32\"},\"t_int8\":{\"encoding\":\"inplace\",\"label\":\"int8\",\"numberOfBytes\":\"1\"},\"t_mapping(t_uint256,t_uint256)\":{\"encoding\":\"mapping\",\"key\":\"t_uint256\",\"label\":\"mapping(uint256 => uint256)\",\"numberOfBytes\":\"32\",\"value\":\"t_uint256\"},\"t_string_storage\":{\"encoding\":\"bytes\",\"label\":\"string\",\"numberOfBytes\":\"32\"},\"t_struct(Funder)62_storage\":{\"encoding\":\"inplace\",\"label\":\"struct SimpleStorage.Funder\",\"members\":[{\"astId\":59,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"addr\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_string_storage\"},{\"astId\":61,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"amount\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint256\"}],\"numberOfBytes\":\"64\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"},\"t_uint8\":{\"encoding\":\"inplace\",\"label\":\"uint8\",\"numberOfBytes\":\"1\"}}}"

	Template SolidityStorageDocument
	_        = json.Unmarshal([]byte(TemplateString), &Template)
)

var (
	intPrefix          = "t_int"
	uintPrefix         = "t_uint"
	boolPrefix         = "t_bool"
	addressPrefix      = "t_address"
	contractPrefix     = "t_contract"
	bytesPrefix        = "t_bytes"
	bytesStoragePrefix = "t_bytes_storage"
	enumPrefix         = "t_enum"
)

type SolidityStorageEntries []SolidityStorageEntry

type SolidityStorageDocument struct {
	Storage SolidityStorageEntries       `json:"storage"`
	Types   map[string]SolidityTypeEntry `json:"types"`
}

type SolidityStorageEntry struct {
	Label  string `json:"label"`
	Offset uint64 `json:"offset"`
	Slot   string `json:"slot"`
	Type   string `json:"type"`
}

type SolidityTypeEntry struct {
	Encoding      string                 `json:"encoding"`
	Key           string                 `json:"key"`
	Label         string                 `json:"label"`
	NumberOfBytes string                 `json:"numberOfBytes"`
	Value         string                 `json:"value"`
	Base          string                 `json:"base"`
	Members       SolidityStorageEntries `json:"members"`
}

func (sse SolidityStorageEntries) Len() int {
	return len(sse)
}

func (sse SolidityStorageEntries) Less(i, j int) bool {
	return (sse[i].Slot < sse[j].Slot) || (sse[i].Offset < sse[j].Offset)
}

func (sse SolidityStorageEntries) Swap(i, j int) {
	sse[i], sse[j] = sse[j], sse[i]
}

func parseRawStorageTwo(rawStorage map[common.Hash]string) ([]*StorageItem, error) {
	parsedStorage := []*StorageItem{}

	sort.Sort(Template.Storage)

	//pad all the storage
	for hsh, storage := range rawStorage {
		rawStorage[hsh] = PadLeft(storage, "0", 64)
	}

	//sort the storage based on
	for _, storageItemTemplate := range Template.Storage {
		namedType := Template.Types[storageItemTemplate.Type]
		startingSlot := DecimalStringToHash(storageItemTemplate.Slot)

		var result interface{}

		switch {
		case strings.HasPrefix(storageItemTemplate.Type, intPrefix):
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = parseInt(bytes).String()
		case strings.HasPrefix(storageItemTemplate.Type, uintPrefix):
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = parseUint(bytes).String()
		case strings.HasPrefix(storageItemTemplate.Type, boolPrefix):
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = bytes[0] == 1
		case strings.HasPrefix(storageItemTemplate.Type, addressPrefix):
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = common.BytesToAddress(bytes).String()
		case strings.HasPrefix(storageItemTemplate.Type, contractPrefix): //TODO: recurse down contracts?
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = common.BytesToAddress(bytes).String()
		case strings.HasPrefix(storageItemTemplate.Type, bytesPrefix) && !strings.HasPrefix(storageItemTemplate.Type, bytesStoragePrefix):
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = "0x" + common.Bytes2Hex(bytes)
		case strings.HasPrefix(storageItemTemplate.Type, enumPrefix):
			bytes, err := ExtractFromSingleStorage(storageItemTemplate, namedType, rawStorage[startingSlot])
			if err != nil {
				return nil, err
			}
			result = uint64(bytes[0])
		}

		if result != nil {
			parsedStorageItem := &StorageItem{
				VarName:  storageItemTemplate.Label,
				VarIndex: 0,
				VarType:  namedType.Label,
				Value:    result,
			}
			parsedStorage = append(parsedStorage, parsedStorageItem)
		}

		if namedType.Members != nil {
			// Struct type
		}

		if namedType.Base != "" {
			// Array type
		}

		if namedType.Key != "" {
			// Mapping
		}
	}

	return parsedStorage, nil
}

// Helper functions

//TODO: check how this errors on invalid input
func DecimalStringToHash(decimal string) common.Hash {
	i := new(big.Int)
	i.SetString(decimal, 10)
	return common.BigToHash(i)
}

func PadLeft(str, pad string, length int) string {
	for {
		str = pad + str
		if len(str) > length {
			return str[1:]
		}
	}
}

func ExtractFromSingleStorage(storageItemTemplate SolidityStorageEntry, namedType SolidityTypeEntry, storageEntry string) ([]byte, error) {
	if storageEntry == "" {
		storageEntry = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	bytes := common.Hex2Bytes(storageEntry)
	offset := storageItemTemplate.Offset

	numBytes, err := strconv.ParseUint(namedType.NumberOfBytes, 10, 0)
	if err != nil {
		return nil, err
	}

	extractedBytes := bytes[32-offset-numBytes : 32-offset]
	return extractedBytes, nil
}

// Parser functions

func parseInt(bytes []byte) *big.Int {
	isPositive := bytes[0] < 128

	if isPositive {
		i := new(big.Int)
		i.SetBytes(bytes)
		return i
	}

	//
	for i := 0; i < len(bytes); i++ {
		bytes[i] = ^bytes[i]
	}

	i := new(big.Int)
	i.SetBytes(bytes)
	i.Add(i, new(big.Int).SetUint64(1))
	i.Neg(i)

	return i
}

func parseUint(bytes []byte) *big.Int {
	return new(big.Int).SetBytes(bytes)
}
