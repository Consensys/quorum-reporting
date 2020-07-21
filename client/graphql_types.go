package client

import (
	"encoding/json"
	"strconv"

	"quorumengineering/quorum-report/types"
)

type CurrentBlockResult struct {
	Block Block
}

type TransactionResult struct {
	Transaction Transaction
}

type Block struct {
	Number HexNumber
}

type Transaction struct {
	Hash              types.Hash
	Status            string
	Index             uint64
	Nonce             HexNumber
	From              Address
	To                Address
	Value             HexNumber
	GasPrice          HexNumber
	Gas               HexNumber
	GasUsed           HexNumber
	CumulativeGasUsed HexNumber
	CreatedContract   Address
	InputData         types.HexData
	PrivateInputData  types.HexData
	IsPrivate         bool
	Logs              []Event
}

type Event struct {
	Index   uint64
	Account Address
	Topics  []types.Hash
	Data    types.HexData
}

type Address struct {
	Address types.Address
}

type HexNumber uint64

func (num *HexNumber) UnmarshalJSON(input []byte) error {
	var unwrapped string
	if err := json.Unmarshal(input, &unwrapped); err != nil {
		return err
	}
	out, err := strconv.ParseUint(unwrapped, 0, 64)
	if err != nil {
		return err
	}
	*num = HexNumber(out)
	return nil
}

func (num *HexNumber) ToUint64() uint64 {
	return uint64(*num)
}
