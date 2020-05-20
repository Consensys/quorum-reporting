package parsers

import "github.com/ethereum/go-ethereum/common"

type StorageManager interface {
	//Get retrieves the value for a given storage slot, padding as needed
	Get(hash common.Hash) string
}

type DefaultStorageManager struct {
	rawStorage map[common.Hash]string
}

func NewDefaultStorageHandler(rawStorage map[common.Hash]string) StorageManager {
	return &DefaultStorageManager{rawStorage: rawStorage}
}

func (dms *DefaultStorageManager) Get(hash common.Hash) string {
	return PadLeft(dms.rawStorage[hash], "0", 64)
}

func PadLeft(str, pad string, length int) string {
	for {
		str = pad + str
		if len(str) > length {
			return str[1:]
		}
	}
}
