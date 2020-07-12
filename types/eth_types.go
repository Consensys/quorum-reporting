package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

var (
	EmptyHash    = NewHash("")
	EmptyAddress = NewAddress("")
)

type Address string

// NewAddressFromHex creates a new address from a given hex string
// It will left-pad to 20 bytes if the string is shorter than that,
// or truncate to 20 bytes if larger
func NewAddress(hexString string) Address {
	if strings.HasPrefix(hexString, "0x") {
		hexString = hexString[2:]
	}
	if len(hexString) > 40 {
		return Address(hexString[:40])
	}
	return Address(fmt.Sprintf("%040v", hexString))
}

func (addr Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr *Address) UnmarshalJSON(input []byte) error {
	var unwrapped string
	if err := json.Unmarshal(input, &unwrapped); err != nil {
		return err
	}
	bytes, err := fromHex(unwrapped)
	if err != nil {
		return err
	}
	*addr = NewAddress(hex.EncodeToString(bytes))
	return nil
}

func (addr *Address) UnmarshalTOML(input []byte) error {
	return addr.UnmarshalJSON(input)
}

func (addr *Address) String() string {
	return addr.Hex()
}

func (addr *Address) Hex() string {
	return "0x" + string(*addr)
}

type Hash string

// NewHashFromHex creates a new hash from a given hex string
// It will left-pad to 32 bytes if the string is shorter than that,
// or truncate to 32 bytes if larger
func NewHash(hexString string) Hash {
	if strings.HasPrefix(hexString, "0x") {
		hexString = hexString[2:]
	}
	if len(hexString) > 64 {
		return Hash(hexString[:64])
	}
	return Hash(fmt.Sprintf("%064v", hexString))
}

func (hsh Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hsh.String())
}

func (hsh *Hash) UnmarshalJSON(input []byte) error {
	var unwrapped string
	if err := json.Unmarshal(input, &unwrapped); err != nil {
		return err
	}
	bytes, err := fromHex(unwrapped)
	if err != nil {
		return err
	}
	*hsh = NewHash(hex.EncodeToString(bytes))
	return nil
}

func (hsh *Hash) UnmarshalTOML(input []byte) error {
	return hsh.UnmarshalJSON(input)
}

func (hsh *Hash) String() string {
	return hsh.Hex()
}

func (hsh *Hash) Hex() string {
	return "0x" + string(*hsh)
}

func fromHex(hexString string) ([]byte, error) {
	if len(hexString) >= 2 && hexString[:2] == "0x" {
		hexString = hexString[2:]
	}
	return hex.DecodeString(hexString)
}

type AccountState struct {
	Root    Hash            `json:"root"`
	Storage map[Hash]string `json:"storage,omitempty"`
}

type HexData string

func (data HexData) MarshalJSON() ([]byte, error) {
	return json.Marshal(data.String())
}

func (data *HexData) UnmarshalJSON(input []byte) error {
	var unwrapped string
	if err := json.Unmarshal(input, &unwrapped); err != nil {
		return err
	}
	hexBytes, err := fromHex(unwrapped)
	if err != nil {
		return err
	}
	*data = HexData(hex.EncodeToString(hexBytes)) //Removes the leading "0x" if there
	return nil
}

func (data *HexData) String() string {
	return "0x" + string(*data)
}

func (data *HexData) AsBytes() []byte {
	converted, _ := hex.DecodeString(string(*data))
	return converted
}

// Call args for checking a contract for EIP165 interfaces
type EIP165Call struct {
	To   Address `json:"to"`
	Data HexData `json:"data"`
}
