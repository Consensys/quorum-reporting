package client

import (
	"context"
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
	ExecuteGraphQLQuery(context.Context, string) (map[string]interface{}, error)
}
