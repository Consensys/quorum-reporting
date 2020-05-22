package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"quorumengineering/quorum-report/types"
	"testing"
)

func TestParser_ParseBytesStorage_ShortByteArray(t *testing.T) {
	sampleStorageItem := []byte("sample")
	paddedItem := common.RightPadBytes(sampleStorageItem, 32)
	paddedItem[31] = 2 * 6

	sm := NewDefaultStorageHandler(make(map[common.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, common.Hash{})

	out := parser.ParseBytesStorage(paddedItem, types.SolidityStorageEntry{})

	expectedOut := []string{"73", "61", "6d", "70", "6c", "65"}
	assert.Equal(t, expectedOut, out)
}

func TestParser_ParseBytesStorage_LargeByteArrayDoubleSlot(t *testing.T) {
	sampleStorageItem := common.RightPadBytes([]byte("large sample that exceeds the 32 bytes of a single slot"), 64)
	paddedItem := make([]byte, 32)
	paddedItem[31] = 111

	storageMap := make(map[common.Hash]string)
	storageMap[common.HexToHash("0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563")] = common.Bytes2Hex(sampleStorageItem[:32])
	storageMap[common.HexToHash("0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e564")] = common.Bytes2Hex(sampleStorageItem[32:])
	sm := NewDefaultStorageHandler(storageMap)
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, common.Hash{})

	//empty storage entry same as first storage entry with no offset
	out := parser.ParseBytesStorage(paddedItem, types.SolidityStorageEntry{})

	expectedOut := []string{"6c", "61", "72", "67", "65", "20", "73", "61", "6d", "70", "6c", "65", "20", "74", "68",
		"61", "74", "20", "65", "78", "63", "65", "65", "64", "73", "20", "74", "68", "65", "20", "33", "32", "20",
		"62", "79", "74", "65", "73", "20", "6f", "66", "20", "61", "20", "73", "69", "6e", "67", "6c", "65", "20",
		"73", "6c", "6f", "74"}
	assert.Equal(t, expectedOut, out)
}

func TestParser_ParseStringStorage_ShortByteArray(t *testing.T) {
	sampleStorageItem := []byte("sample")
	paddedItem := common.RightPadBytes(sampleStorageItem, 32)
	paddedItem[31] = 2 * 6

	sm := NewDefaultStorageHandler(make(map[common.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, common.Hash{})

	out := parser.ParseStringStorage(paddedItem, types.SolidityStorageEntry{})

	assert.Equal(t, string(sampleStorageItem), out)
}

func TestParser_ParseStringStorage_LargeByteArrayDoubleSlot(t *testing.T) {
	message := "large sample that exceeds the 32 bytes of a single slot"
	sampleStorageItem := common.RightPadBytes([]byte(message), 64)
	paddedItem := make([]byte, 32)
	paddedItem[31] = 111

	storageMap := make(map[common.Hash]string)
	storageMap[common.HexToHash("0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563")] = common.Bytes2Hex(sampleStorageItem[:32])
	storageMap[common.HexToHash("0x290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e564")] = common.Bytes2Hex(sampleStorageItem[32:])
	sm := NewDefaultStorageHandler(storageMap)
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, common.Hash{})

	//empty storage entry same as first storage entry with no offset
	out := parser.ParseStringStorage(paddedItem, types.SolidityStorageEntry{})

	assert.Equal(t, message, out)
}
