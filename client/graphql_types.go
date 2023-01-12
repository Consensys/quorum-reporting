package client

import (
	"quorumengineering/quorum-report/types"
)

type CurrentBlockResult struct {
	Block Block
}

type TransactionResult struct {
	Transaction Transaction
}

type Block struct {
	Number types.HexNumber
}

type Transaction struct {
	Hash              types.Hash
	Status            types.HexNumber
	Index             uint64
	Nonce             types.HexNumber
	From              Address
	To                Address
	Value             types.HexNumber
	GasPrice          types.HexNumber
	Gas               types.HexNumber
	GasUsed           types.HexNumber
	CumulativeGasUsed types.HexNumber
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
