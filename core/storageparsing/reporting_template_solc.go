package storageparsing

import (
	"quorumengineering/quorum-report/types"
)

func ParseRawStorage(rawStorage map[types.Hash]string, template types.SolidityStorageDocument) ([]*types.StorageItem, error) {
	initialStorageManager := NewDefaultStorageHandler(rawStorage)
	parser := NewParser(initialStorageManager, template, types.NewHash(""))
	return parser.ParseRawStorage()
}
