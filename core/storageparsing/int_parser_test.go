package storageparsing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/consensys/quorum-go-utils/types"
)

func TestParser_ParseUint_0(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	i := make([]byte, 32)

	res := parser.ParseUint(i)

	assert.EqualValues(t, 0, res.Uint64())
}

func TestParser_ParseUint_1(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	i := make([]byte, 32)
	i[31] = byte(1)

	res := parser.ParseUint(i)

	assert.EqualValues(t, 1, res.Uint64())
}

func TestParser_ParseUint_Large(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	unpaddedBytes := bigN(234567).Bytes()
	i := make([]byte, 32)
	copy(i[32-len(unpaddedBytes):32], unpaddedBytes) //left-pad to 32 bytes

	res := parser.ParseUint(i)

	assert.EqualValues(t, 234567, res.Uint64())
}

func TestParser_ParseInt_0(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	i := make([]byte, 32)

	res := parser.ParseInt(i)

	assert.EqualValues(t, 0, res.Int64())
}

func TestParser_ParseInt_1(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	i := make([]byte, 32)
	i[31] = byte(1)

	res := parser.ParseInt(i)

	assert.EqualValues(t, 1, res.Int64())
}

func TestParser_ParseInt_LargePositive(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	unpaddedBytes := bigN(234567).Bytes()
	i := make([]byte, 32)
	copy(i[32-len(unpaddedBytes):32], unpaddedBytes) //left-pad to 32 bytes

	res := parser.ParseInt(i)

	assert.EqualValues(t, 234567, res.Uint64())
}

func TestParser_ParseInt_Minus1(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	i := []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

	res := parser.ParseInt(i)

	assert.EqualValues(t, -1, res.Int64())
}

func TestParser_ParseInt_LargeNegative(t *testing.T) {
	sm := NewDefaultStorageHandler(make(map[types.Hash]string))
	doc := types.SolidityStorageDocument{
		Storage: make([]types.SolidityStorageEntry, 0),
		Types:   nil,
	}
	parser := NewParser(sm, doc, types.NewHash(""))

	i := []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 0}

	res := parser.ParseInt(i)

	assert.EqualValues(t, "-4294967296", res.String())
}
