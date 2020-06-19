package client

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/mitchellh/mapstructure"
)

// StubQuorumClient is used for unit test.
type StubQuorumClient struct {
	blocks      []*ethTypes.Block
	mockGraphQL map[string]map[string]interface{}
	mockRPC     map[string]interface{}
}

func NewStubQuorumClient(blocks []*ethTypes.Block, mockGraphQL map[string]map[string]interface{}, mockRPC map[string]interface{}) Client {
	if mockGraphQL == nil {
		mockGraphQL = map[string]map[string]interface{}{}
	}
	if mockRPC == nil {
		mockRPC = map[string]interface{}{}
	}
	return &StubQuorumClient{blocks, mockGraphQL, mockRPC}
}

func (qc *StubQuorumClient) SubscribeNewHead(context.Context, chan<- *ethTypes.Header) (ethereum.Subscription, error) {
	return nil, errors.New("not implemented")
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
	fmt.Println(method)
	for _, arg := range args {
		method += reflect.ValueOf(arg).String()
	}
	if resp, ok := qc.mockRPC[method]; ok {
		reflect.ValueOf(result).Elem().Set(reflect.ValueOf(resp))
		return nil
	}
	return errors.New("not found")
}
