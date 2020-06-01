package rpc

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

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
