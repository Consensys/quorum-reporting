package storageparsing

import (
	"testing"

	"github.com/consensys/quorum-go-utils/types"
	"github.com/stretchr/testify/assert"
)

func TestParser_determineSize_fixedSize(t *testing.T) {
	sampleType := "t_array(t_bytes1)10_storage"
	storageItem := types.SolidityStorageEntry{
		Label:  "",
		Offset: 0,
		Slot:   0,
		Type:   sampleType,
	}

	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	size, err := parser.determineSize(storageItem, false)

	assert.Nil(t, err)
	assert.EqualValues(t, 10, size)
}

func TestParser_determineSize_dynamicSize(t *testing.T) {
	sampleType := "t_array(t_bytes1)dyn_storage"
	storageItem := types.SolidityStorageEntry{
		Label:  "",
		Offset: 0,
		Slot:   0,
		Type:   sampleType,
	}

	storageMap := make(map[types.Hash]string)
	storageMap[types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000000")] = "41a2"
	sm := NewDefaultStorageHandler(storageMap)
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	size, err := parser.determineSize(storageItem, true)

	assert.Nil(t, err)
	assert.Equal(t, uint64(0x41a2), size)
}

func TestParser_createArrayStorageDocument_smallElements(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   make(map[string]types.SolidityTypeEntry),
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	out := parser.createArrayStorageDocument(5, 1, "customType")

	expected := []types.SolidityStorageEntry{
		{Label: "", Offset: 0, Slot: 0, Type: "customType"},
		{Label: "", Offset: 1, Slot: 0, Type: "customType"},
		{Label: "", Offset: 2, Slot: 0, Type: "customType"},
		{Label: "", Offset: 3, Slot: 0, Type: "customType"},
		{Label: "", Offset: 4, Slot: 0, Type: "customType"},
	}

	assert.Equal(t, out.Types, doc.Types)
	assert.EqualValues(t, expected, out.Storage)
}

func TestParser_createArrayStorageDocument_ElementsOverflowSlot(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   make(map[string]types.SolidityTypeEntry),
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	out := parser.createArrayStorageDocument(8, 10, "customType")

	expected := []types.SolidityStorageEntry{
		{Label: "", Offset: 0, Slot: 0, Type: "customType"},
		{Label: "", Offset: 10, Slot: 0, Type: "customType"},
		{Label: "", Offset: 20, Slot: 0, Type: "customType"},
		{Label: "", Offset: 0, Slot: 1, Type: "customType"},
		{Label: "", Offset: 10, Slot: 1, Type: "customType"},
		{Label: "", Offset: 20, Slot: 1, Type: "customType"},
		{Label: "", Offset: 0, Slot: 2, Type: "customType"},
		{Label: "", Offset: 10, Slot: 2, Type: "customType"},
	}

	assert.Equal(t, out.Types, doc.Types)
	assert.EqualValues(t, expected, out.Storage)
}

func TestParser_createArrayStorageDocument_ElementsTakeMultipleSlots(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   make(map[string]types.SolidityTypeEntry),
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	out := parser.createArrayStorageDocument(8, 80, "customType")

	expected := []types.SolidityStorageEntry{
		{Label: "", Offset: 0, Slot: 0, Type: "customType"},
		{Label: "", Offset: 0, Slot: 3, Type: "customType"},
		{Label: "", Offset: 0, Slot: 6, Type: "customType"},
		{Label: "", Offset: 0, Slot: 9, Type: "customType"},
		{Label: "", Offset: 0, Slot: 12, Type: "customType"},
		{Label: "", Offset: 0, Slot: 15, Type: "customType"},
		{Label: "", Offset: 0, Slot: 18, Type: "customType"},
		{Label: "", Offset: 0, Slot: 21, Type: "customType"},
	}

	assert.Equal(t, out.Types, doc.Types)
	assert.EqualValues(t, expected, out.Storage)
}
