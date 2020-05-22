package storageparsing

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

func ParseRawStorage(rawStorage map[common.Hash]string, template types.SolidityStorageDocument) ([]*types.StorageItem, error) {
	initialStorageManager := NewDefaultStorageHandler(rawStorage)
	parser := NewParser(initialStorageManager, template, common.Hash{})
	return parser.ParseRawStorage()
}
