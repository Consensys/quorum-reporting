package storageparsing

import (
	"encoding/hex"
	"fmt"

	"quorumengineering/quorum-report/types"
)

type StorageManager interface {
	//Get retrieves the value for a given storage slot, padding as needed
	Get(hash types.Hash) []byte
}

type DefaultStorageManager struct {
	rawStorage map[types.Hash]string
}

func NewDefaultStorageHandler(rawStorage map[types.Hash]string) StorageManager {
	return &DefaultStorageManager{rawStorage: rawStorage}
}

func (dms *DefaultStorageManager) Get(hash types.Hash) []byte {
	paddedString := fmt.Sprintf("%064v", dms.rawStorage[hash]) //pad to 64 hex chars, 32 bytes
	decoded, _ := hex.DecodeString(paddedString)
	return decoded
}
