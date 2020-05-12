package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

type Client interface {
	// from ethclient.Client
	SubscribeNewHead(context.Context, chan<- *ethTypes.Header) (ethereum.Subscription, error)
	BlockByNumber(context.Context, *big.Int) (*ethTypes.Block, error)
	// graphql
	ExecuteGraphQLQuery(context.Context, interface{}, string) error
	// rpc
	RPCCall(context.Context, interface{}, string, ...interface{}) error
}
