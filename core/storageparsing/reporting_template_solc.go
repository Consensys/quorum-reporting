package storageparsing

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/core/storageparsing/parsers"
	"quorumengineering/quorum-report/core/storageparsing/types"
)

func ParseRawStorage(rawStorage map[common.Hash]string, template types.SolidityStorageDocument) ([]*types.StorageItem, error) {
	initialStorageManager := types.NewDefaultStorageHandler(rawStorage)
	parser := parsers.NewParser(initialStorageManager, template, common.Hash{})
	return parser.ParseRawStorage()
}
