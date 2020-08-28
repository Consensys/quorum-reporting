package rpc

import (
	"errors"
	"math/big"

	"github.com/consensys/quorum-go-utils/types"
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

type ERC20TokenQuery struct {
	Contract *types.Address
	Holder   *types.Address
	Block    uint64
	Options  *types.TokenQueryOptions
}

type ERC721TokenQuery struct {
	Contract *types.Address
	Holder   *types.Address
	TokenId  *big.Int
	Block    uint64
	Options  *types.TokenQueryOptions
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
