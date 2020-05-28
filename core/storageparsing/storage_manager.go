package storageparsing

import "github.com/ethereum/go-ethereum/common"

type StorageManager interface {
	//Get retrieves the value for a given storage slot, padding as needed
	Get(hash common.Hash) []byte
}

type DefaultStorageManager struct {
	rawStorage map[common.Hash]string
}

func NewDefaultStorageHandler(rawStorage map[common.Hash]string) StorageManager {
	return &DefaultStorageManager{rawStorage: rawStorage}
}

func (dms *DefaultStorageManager) Get(hash common.Hash) []byte {
	return common.LeftPadBytes(common.Hex2Bytes(dms.rawStorage[hash]), 32)
}
