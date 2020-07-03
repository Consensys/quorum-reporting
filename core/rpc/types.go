package rpc

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

//Inputs

type NullArgs struct{}

type AddressWithOptions struct {
	Address *common.Address
	Options *types.QueryOptions
}

type AddressWithData struct {
	Address *common.Address
	Data    string
}

type TemplateArgs struct {
	Name          string
	Abi           string
	StorageLayout string
}

type AddressWithBlock struct {
	Address     *common.Address
	BlockNumber *uint64
}

type AddressWithOptionalBlock struct {
	Address     common.Address
	BlockNumber *uint64
}

type AddressWithBlockRange struct {
	Address          *common.Address
	StartBlockNumber uint64
	EndBlockNumber   uint64
}

//Outputs

type TransactionsResp struct {
	Transactions []common.Hash       `json:"transactions"`
	Total        uint64              `json:"total"`
	Options      *types.QueryOptions `json:"options"`
}

type EventsResp struct {
	Events  []*types.ParsedEvent `json:"events"`
	Total   uint64               `json:"total"`
	Options *types.QueryOptions  `json:"options"`
}
