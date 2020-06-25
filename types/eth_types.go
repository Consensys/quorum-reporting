package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type CallArgs struct {
	From     *common.Address `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Data     *hexutil.Bytes  `json:"data"`
}

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
	return Address(fmt.Sprintf("%40v", hexString))
}

func (addr Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr *Address) UnmarshalJSON(input []byte) error {
	var unwrapped string
	if err := json.Unmarshal(input, &unwrapped); err != nil {
		return err
	}
	if _, err := fromHex(unwrapped); err != nil {
		return err
	}
	*addr = NewAddress(unwrapped)
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
	return Hash(fmt.Sprintf("%64v", hexString))
}

func (hsh Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hsh.String())
}

func (hsh *Hash) UnmarshalJSON(input []byte) error {
	var unwrapped string
	if err := json.Unmarshal(input, &unwrapped); err != nil {
		return err
	}
	if _, err := fromHex(unwrapped); err != nil {
		return err
	}
	*hsh = NewHash(unwrapped)
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
