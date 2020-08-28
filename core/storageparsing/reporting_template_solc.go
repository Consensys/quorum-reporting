package storageparsing

import (
	"github.com/consensys/quorum-go-utils/types"
)

func ParseRawStorage(rawStorage map[types.Hash]string, template types.SolidityStorageDocument) ([]*types.StorageItem, error) {
	initialStorageManager := NewDefaultStorageHandler(rawStorage)
	parser := NewParser(initialStorageManager, template, types.NewHash(""))
	return parser.ParseRawStorage()
}
