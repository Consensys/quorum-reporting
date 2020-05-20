package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"quorumengineering/quorum-report/types"
	"sort"
	"strings"
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

type Parser struct {
	storageManager StorageManager
	template       types.SolidityStorageDocument

	slotOffset common.Hash
}

func NewParser(sm StorageManager, template types.SolidityStorageDocument, slotOffset common.Hash) *Parser {
	sort.Sort(template.Storage)

	parser := &Parser{
		storageManager: sm,
		template:       template,
		slotOffset:     slotOffset,
	}

	return parser
}

func (p *Parser) ParseRawStorage() ([]*types.StorageItem, error) {
	parsedStorage := []*types.StorageItem{}

	for _, storageItem := range p.template.Storage {

		namedType := p.template.Types[storageItem.Type]

		result, err := p.parseSingle(storageItem)
		if err != nil {
			return nil, err
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

func (p *Parser) parseSingle(storageItem types.SolidityStorageEntry) (interface{}, error) {

	namedType := p.template.Types[storageItem.Type]
	startingSlot := p.ResolveSlot(bigN(storageItem.Slot))
	directStorageSlot := p.storageManager.Get(startingSlot) //the storage this variable uses by its "Slot"

	var result interface{}

	switch {
	case strings.HasPrefix(storageItem.Type, intPrefix):
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = p.ParseInt(bytes).String()

	case strings.HasPrefix(storageItem.Type, uintPrefix):
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = p.ParseUint(bytes).String()

	case strings.HasPrefix(storageItem.Type, boolPrefix):
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = bytes[0] == 1

	case strings.HasPrefix(storageItem.Type, addressPrefix):
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = common.BytesToAddress(bytes).String()

	case strings.HasPrefix(storageItem.Type, contractPrefix): //TODO: recurse down contracts?
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = common.BytesToAddress(bytes).String()

	case strings.HasPrefix(storageItem.Type, bytesPrefix) && !strings.HasPrefix(storageItem.Type, bytesStoragePrefix):
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = "0x" + common.Bytes2Hex(bytes)

	case strings.HasPrefix(storageItem.Type, enumPrefix):
		bytes := ExtractFromSingleStorage(storageItem.Offset, namedType.NumberOfBytes, directStorageSlot)
		result = uint64(bytes[0])

	case strings.HasPrefix(storageItem.Type, bytesStoragePrefix):
		bytes := p.ParseBytesStorage(directStorageSlot, storageItem)
		result = bytes

	case strings.HasPrefix(storageItem.Type, stringPrefix):
		str := p.ParseStringStorage(directStorageSlot, storageItem)
		result = str

	case strings.HasPrefix(storageItem.Type, arrayPrefix):
		res, err := p.ParseArray(storageItem, namedType)
		if err != nil {
			return nil, err
		}
		result = res

	case strings.HasPrefix(storageItem.Type, structPrefix):
		res, err := p.ParseStruct(storageItem, namedType)
		if err != nil {
			return nil, err
		}
		result = res
	}

	return result, nil
}
