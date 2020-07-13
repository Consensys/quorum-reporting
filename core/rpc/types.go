package rpc

import (
	"errors"

	"quorumengineering/quorum-report/types"
)

var ErrNoAddress = errors.New("address not provided")

//Inputs

type NullArgs struct{}

type AddressWithOptions struct {
	Address *types.Address
	Options *types.QueryOptions
}

type AddressWithData struct {
	Address *types.Address
	Data    string
}

type TemplateArgs struct {
	Name          string
	Abi           string
	StorageLayout string
}

type AddressWithOptionalBlock struct {
	Address     *types.Address
	BlockNumber *uint64
}

type AddressWithBlockRange struct {
	Address          *types.Address
	StartBlockNumber uint64
	EndBlockNumber   uint64
}

//Outputs

type TransactionsResp struct {
	Transactions []types.Hash        `json:"transactions"`
	Total        uint64              `json:"total"`
	Options      *types.QueryOptions `json:"options"`
}

type EventsResp struct {
	Events  []*types.ParsedEvent `json:"events"`
	Total   uint64               `json:"total"`
	Options *types.QueryOptions  `json:"options"`
}
