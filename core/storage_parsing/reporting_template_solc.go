package storage_parsing

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"quorumengineering/quorum-report/core/storage_parsing/parsers"
	"quorumengineering/quorum-report/core/storage_parsing/types"
	"sort"
	"strings"
)

var (
	TemplateString = "{\"storage\":[{\"astId\":3,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"a\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint256\"},{\"astId\":5,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"b\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint8\"},{\"astId\":7,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"c\",\"offset\":1,\"slot\":\"1\",\"type\":\"t_uint8\"},{\"astId\":9,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_int256\"},{\"astId\":11,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d2\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_int256\"},{\"astId\":13,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d3\",\"offset\":0,\"slot\":\"4\",\"type\":\"t_int8\"},{\"astId\":15,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"d4\",\"offset\":1,\"slot\":\"4\",\"type\":\"t_int24\"},{\"astId\":17,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"e\",\"offset\":4,\"slot\":\"4\",\"type\":\"t_bool\"},{\"astId\":19,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"f\",\"offset\":5,\"slot\":\"4\",\"type\":\"t_address\"},{\"astId\":21,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"g\",\"offset\":0,\"slot\":\"5\",\"type\":\"t_contract(SimpleStorage)458\"},{\"astId\":23,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h1\",\"offset\":20,\"slot\":\"5\",\"type\":\"t_bytes1\"},{\"astId\":25,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h2\",\"offset\":21,\"slot\":\"5\",\"type\":\"t_bytes1\"},{\"astId\":27,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h3\",\"offset\":22,\"slot\":\"5\",\"type\":\"t_bytes2\"},{\"astId\":29,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h4\",\"offset\":0,\"slot\":\"6\",\"type\":\"t_bytes31\"},{\"astId\":31,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h5\",\"offset\":0,\"slot\":\"7\",\"type\":\"t_bytes32\"},{\"astId\":38,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"choice\",\"offset\":0,\"slot\":\"8\",\"type\":\"t_enum(ActionChoices)36\"},{\"astId\":40,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"lessThan31\",\"offset\":0,\"slot\":\"9\",\"type\":\"t_bytes_storage\"},{\"astId\":42,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"exactly31\",\"offset\":0,\"slot\":\"10\",\"type\":\"t_bytes_storage\"},{\"astId\":44,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"exactly32\",\"offset\":0,\"slot\":\"11\",\"type\":\"t_bytes_storage\"},{\"astId\":46,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"moreThan31\",\"offset\":0,\"slot\":\"12\",\"type\":\"t_bytes_storage\"},{\"astId\":48,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i2\",\"offset\":0,\"slot\":\"13\",\"type\":\"t_string_storage\"},{\"astId\":50,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i5\",\"offset\":0,\"slot\":\"14\",\"type\":\"t_string_storage\"},{\"astId\":53,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h6\",\"offset\":0,\"slot\":\"15\",\"type\":\"t_array(t_bytes1)dyn_storage\"},{\"astId\":56,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h6long\",\"offset\":0,\"slot\":\"16\",\"type\":\"t_array(t_bytes1)dyn_storage\"},{\"astId\":60,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h7\",\"offset\":0,\"slot\":\"17\",\"type\":\"t_array(t_bytes1)10_storage\"},{\"astId\":64,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"h7long\",\"offset\":0,\"slot\":\"18\",\"type\":\"t_array(t_bytes1)60_storage\"},{\"astId\":67,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i3\",\"offset\":0,\"slot\":\"20\",\"type\":\"t_array(t_address)dyn_storage\"},{\"astId\":70,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"i4\",\"offset\":0,\"slot\":\"21\",\"type\":\"t_array(t_contract(SimpleStorage)458)dyn_storage\"},{\"astId\":74,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"doubleArray\",\"offset\":0,\"slot\":\"22\",\"type\":\"t_array(t_array(t_int256)dyn_storage)dyn_storage\"},{\"astId\":81,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"funder1\",\"offset\":0,\"slot\":\"23\",\"type\":\"t_struct(Funder)79_storage\"},{\"astId\":85,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"fundersFixed\",\"offset\":0,\"slot\":\"25\",\"type\":\"t_array(t_struct(Funder)79_storage)2_storage\"},{\"astId\":88,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"fundersDyn\",\"offset\":0,\"slot\":\"29\",\"type\":\"t_array(t_struct(Funder)79_storage)dyn_storage\"},{\"astId\":101,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"longstruct\",\"offset\":0,\"slot\":\"30\",\"type\":\"t_struct(LongerStruct)99_storage\"},{\"astId\":105,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"map\",\"offset\":0,\"slot\":\"34\",\"type\":\"t_mapping(t_uint256,t_uint256)\"}],\"types\":{\"t_address\":{\"encoding\":\"inplace\",\"label\":\"address\",\"numberOfBytes\":\"20\"},\"t_array(t_address)dyn_storage\":{\"base\":\"t_address\",\"encoding\":\"dynamic_array\",\"label\":\"address[]\",\"numberOfBytes\":\"32\"},\"t_array(t_array(t_int256)dyn_storage)dyn_storage\":{\"base\":\"t_array(t_int256)dyn_storage\",\"encoding\":\"dynamic_array\",\"label\":\"int256[][]\",\"numberOfBytes\":\"32\"},\"t_array(t_bytes1)10_storage\":{\"base\":\"t_bytes1\",\"encoding\":\"inplace\",\"label\":\"bytes1[10]\",\"numberOfBytes\":\"32\"},\"t_array(t_bytes1)60_storage\":{\"base\":\"t_bytes1\",\"encoding\":\"inplace\",\"label\":\"bytes1[60]\",\"numberOfBytes\":\"64\"},\"t_array(t_bytes1)dyn_storage\":{\"base\":\"t_bytes1\",\"encoding\":\"dynamic_array\",\"label\":\"bytes1[]\",\"numberOfBytes\":\"32\"},\"t_array(t_contract(SimpleStorage)458)dyn_storage\":{\"base\":\"t_contract(SimpleStorage)458\",\"encoding\":\"dynamic_array\",\"label\":\"contract SimpleStorage[]\",\"numberOfBytes\":\"32\"},\"t_array(t_int256)dyn_storage\":{\"base\":\"t_int256\",\"encoding\":\"dynamic_array\",\"label\":\"int256[]\",\"numberOfBytes\":\"32\"},\"t_array(t_struct(Funder)79_storage)2_storage\":{\"base\":\"t_struct(Funder)79_storage\",\"encoding\":\"inplace\",\"label\":\"struct SimpleStorage.Funder[2]\",\"numberOfBytes\":\"128\"},\"t_array(t_struct(Funder)79_storage)dyn_storage\":{\"base\":\"t_struct(Funder)79_storage\",\"encoding\":\"dynamic_array\",\"label\":\"struct SimpleStorage.Funder[]\",\"numberOfBytes\":\"32\"},\"t_bool\":{\"encoding\":\"inplace\",\"label\":\"bool\",\"numberOfBytes\":\"1\"},\"t_bytes1\":{\"encoding\":\"inplace\",\"label\":\"bytes1\",\"numberOfBytes\":\"1\"},\"t_bytes2\":{\"encoding\":\"inplace\",\"label\":\"bytes2\",\"numberOfBytes\":\"2\"},\"t_bytes31\":{\"encoding\":\"inplace\",\"label\":\"bytes31\",\"numberOfBytes\":\"31\"},\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_bytes_storage\":{\"encoding\":\"bytes\",\"label\":\"bytes\",\"numberOfBytes\":\"32\"},\"t_contract(SimpleStorage)458\":{\"encoding\":\"inplace\",\"label\":\"contract SimpleStorage\",\"numberOfBytes\":\"20\"},\"t_enum(ActionChoices)36\":{\"encoding\":\"inplace\",\"label\":\"enum SimpleStorage.ActionChoices\",\"numberOfBytes\":\"1\"},\"t_int24\":{\"encoding\":\"inplace\",\"label\":\"int24\",\"numberOfBytes\":\"3\"},\"t_int256\":{\"encoding\":\"inplace\",\"label\":\"int256\",\"numberOfBytes\":\"32\"},\"t_int8\":{\"encoding\":\"inplace\",\"label\":\"int8\",\"numberOfBytes\":\"1\"},\"t_mapping(t_uint256,t_uint256)\":{\"encoding\":\"mapping\",\"key\":\"t_uint256\",\"label\":\"mapping(uint256 => uint256)\",\"numberOfBytes\":\"32\",\"value\":\"t_uint256\"},\"t_string_storage\":{\"encoding\":\"bytes\",\"label\":\"string\",\"numberOfBytes\":\"32\"},\"t_struct(Funder)79_storage\":{\"encoding\":\"inplace\",\"label\":\"struct SimpleStorage.Funder\",\"members\":[{\"astId\":76,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"addr\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_string_storage\"},{\"astId\":78,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"amount\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint256\"}],\"numberOfBytes\":\"64\"},\"t_struct(LongerStruct)99_storage\":{\"encoding\":\"inplace\",\"label\":\"struct SimpleStorage.LongerStruct\",\"members\":[{\"astId\":90,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"addr\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_string_storage\"},{\"astId\":92,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"amount\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_uint256\"},{\"astId\":94,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"val\",\"offset\":0,\"slot\":\"2\",\"type\":\"t_int8\"},{\"astId\":96,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"otherval\",\"offset\":1,\"slot\":\"2\",\"type\":\"t_uint8\"},{\"astId\":98,\"contract\":\"/Users/peter/IdeaProjects/quorum-examples/examples/7nodes/simplestorage.sol:SimpleStorage\",\"label\":\"custommessage\",\"offset\":0,\"slot\":\"3\",\"type\":\"t_string_storage\"}],\"numberOfBytes\":\"128\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"},\"t_uint8\":{\"encoding\":\"inplace\",\"label\":\"uint8\",\"numberOfBytes\":\"1\"}}}"

	Template types.SolidityStorageDocument
	_        = json.Unmarshal([]byte(TemplateString), &Template)
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

func ParseRawStorage(rawStorage map[common.Hash]string) ([]*types.StorageItem, error) {
	parsedStorage := []*types.StorageItem{}

	sort.Sort(Template.Storage)

	storageManager := types.NewDefaultStorageHandler(rawStorage)

	//sort the storage based on
	for _, storageItem := range Template.Storage {
		namedType := Template.Types[storageItem.Type]
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
