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
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)

	// ExecuteGraphQLQuery performs a fully constructed query against the Geth
	// GraphQL server
	ExecuteGraphQLQuery(interface{}, string) error

	// RPCCall makes an arbitrary call to the Geth RPC server
	RPCCall(context.Context, interface{}, string, ...interface{}) error
}
