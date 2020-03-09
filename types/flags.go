package types

import "github.com/ethereum/go-ethereum/common"

type Flags struct {
	QuorumWSURL      string           `json:"quorumWSURL"`
	QuorumGraphQLURL string           `json:"quorumGraphQLURL"`
	RPCAddress       string           `json:"rpcAddress"`
	RPCCORS          []string         `json:"rpccors"`
	RPCVHOSTS        []string         `json:"rpcvhosts"`
	Addresses        []common.Address `json:"addresses"`
}
