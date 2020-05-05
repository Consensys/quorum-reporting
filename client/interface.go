package client

import (
	"context"
	"github.com/ethereum/go-ethereum/core/state"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type Client interface {
	SubscribeNewHead(context.Context, chan<- *ethTypes.Header) (ethereum.Subscription, error)
	BlockByHash(context.Context, common.Hash) (*ethTypes.Block, error)
	BlockByNumber(context.Context, *big.Int) (*ethTypes.Block, error)
	// graphql
	ExecuteGraphQLQuery(context.Context, interface{}, string) error
	// rpc
	RPCCall(context.Context, interface{}, string, ...interface{}) error

	DumpAddress(address common.Address, blockNumber uint64) (*state.DumpAccount, error)
	TraceTransaction(txHash common.Hash) (map[string]interface{}, error)
	Consensus() (string, error)
}
