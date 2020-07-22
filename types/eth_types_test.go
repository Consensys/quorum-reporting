package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/naoina/toml"
	"github.com/stretchr/testify/assert"
)

func TestAddress_MarshalJSON(t *testing.T) {
	address := NewAddress("1932c48b2bf8102ba33b4a6b545c32236e342f34")

	result, err := json.Marshal(address)

	assert.Nil(t, err)
	assert.Equal(t, `"0x1932c48b2bf8102ba33b4a6b545c32236e342f34"`, string(result))
}

func TestAddress_UnmarshalJSON(t *testing.T) {
	sampleAddress := `"0x1932c48b2bf8102ba33b4a6b545c32236e342f34"`

	var addr Address
	err := json.Unmarshal([]byte(sampleAddress), &addr)

	assert.Nil(t, err)
	assert.EqualValues(t, NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"), addr)
}

func TestAddress_UnmarshalTOML(t *testing.T) {
	sampleAddress := `address = "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"`

	type TestAddress struct {
		Addr Address `toml:"address"`
	}

	var addr TestAddress
	err := toml.NewDecoder(strings.NewReader(sampleAddress)).Decode(&addr)

	assert.Nil(t, err)
	assert.EqualValues(t, NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"), addr.Addr)
}

func TestNewAddressFromHex_NoPrefix(t *testing.T) {
	addressAsString := "1932c48b2bf8102ba33b4a6b545c32236e342f34"
	address := NewAddress(addressAsString)

	assert.EqualValues(t, "1932c48b2bf8102ba33b4a6b545c32236e342f34", address)
}

func TestNewAddressFromHex_WithPrefix(t *testing.T) {
	addressAsString := "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"
	address := NewAddress(addressAsString)

	assert.EqualValues(t, "1932c48b2bf8102ba33b4a6b545c32236e342f34", address)
}

func TestHash_MarshalJSON(t *testing.T) {
	hash := NewHash("e625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")

	result, err := json.Marshal(hash)

	assert.Nil(t, err)
	assert.Equal(t, `"0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"`, string(result))
}

func TestHash_UnmarshalJSON(t *testing.T) {
	sampleHash := `"0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"`

	var hsh Hash
	err := json.Unmarshal([]byte(sampleHash), &hsh)

	assert.Nil(t, err)
	assert.EqualValues(t, NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"), hsh)
}

func TestHash_UnmarshalTOML(t *testing.T) {
	sampleHash := `someHash = "0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"`

	type TestHash struct {
		Hsh Hash `toml:"someHash"`
	}

	var hsh TestHash
	err := toml.NewDecoder(strings.NewReader(sampleHash)).Decode(&hsh)

	assert.Nil(t, err)
	assert.EqualValues(t, NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"), hsh.Hsh)
}

func TestNewHashFromHex_NoPrefix(t *testing.T) {
	hashAsString := "e625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"
	hash := NewHash(hashAsString)

	assert.EqualValues(t, "0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8", hash.String())
}

func TestNewHashFromHex_WithPrefix(t *testing.T) {
	hashAsString := "0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"
	hash := NewHash(hashAsString)

	assert.EqualValues(t, "0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8", hash.String())
}

func TestHexNumber_MarshalJSON(t *testing.T) {
	hexNum := HexNumber(16)

	result, err := json.Marshal(hexNum)

	assert.Nil(t, err)
	assert.Equal(t, `"0x10"`, string(result))
}

func TestHexNumber_UnmarshalJSON(t *testing.T) {
	hexNumber := `"0x10"`

	var num HexNumber
	err := json.Unmarshal([]byte(hexNumber), &num)

	assert.Nil(t, err)
	assert.EqualValues(t, 16, num)
}
