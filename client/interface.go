package client

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/mitchellh/mapstructure"
)

type Client interface {
	SubscribeNewHead(context.Context, chan<- *ethTypes.Header) (ethereum.Subscription, error)
	BlockByHash(context.Context, common.Hash) (*ethTypes.Block, error)
	BlockByNumber(context.Context, *big.Int) (*ethTypes.Block, error)
	// graphql
	ExecuteGraphQLQuery(context.Context, interface{}, string) error
	// rpc
	RPCCall(context.Context, interface{}, string, ...interface{}) error
}

// StubQuorumClient is used for unit test.
type StubQuorumClient struct {
	blocks      []*ethTypes.Block
	mockGraphQL map[string]map[string]interface{}
}

func NewStubQuorumClient(blocks []*ethTypes.Block, mockGraphQL map[string]map[string]interface{}) Client {
	return &StubQuorumClient{blocks, mockGraphQL}
}

func (qc *StubQuorumClient) SubscribeNewHead(context.Context, chan<- *ethTypes.Header) (ethereum.Subscription, error) {
	return nil, errors.New("not implemented")
}
func (qc *StubQuorumClient) BlockByHash(ctx context.Context, hash common.Hash) (*ethTypes.Block, error) {
	for _, b := range qc.blocks {
		if b.Hash() == hash {
			return b, nil
		}
	}
	return nil, errors.New("not found")
}

func (qc *StubQuorumClient) BlockByNumber(ctx context.Context, blockNumber *big.Int) (*ethTypes.Block, error) {
	for _, b := range qc.blocks {
		if b.Number().Cmp(blockNumber) == 0 {
			return b, nil
		}
	}
	return nil, errors.New("not found")
}

func (qc *StubQuorumClient) ExecuteGraphQLQuery(ctx context.Context, result interface{}, query string) error {
	if resp, ok := qc.mockGraphQL[query]; ok {
		err := mapstructure.Decode(resp, &result)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("not found")
}

func (qc *StubQuorumClient) RPCCall(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return errors.New("not implemented")
}
