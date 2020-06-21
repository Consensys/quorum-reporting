package client

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// StubQuorumClient is used for unit test.
type StubQuorumClient struct {
	blocks      []*ethTypes.Block
	mockGraphQL map[string]map[string]interface{}
	mockRPC     map[string]interface{}
}

func NewStubQuorumClient(blocks []*ethTypes.Block, mockGraphQL map[string]map[string]interface{}, mockRPC map[string]interface{}) *StubQuorumClient {
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

func (qc *StubQuorumClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var res []byte
	err := qc.RPCCall(ctx, &res, "eth_call", msg, blockNumber)
	return res, err
}

func (qc *StubQuorumClient) ExecuteGraphQLQuery(result interface{}, query string) error {
	if resp, ok := qc.mockGraphQL[query]; ok {
		out, _ := json.Marshal(resp)
		return json.Unmarshal(out, &result)
	}
	return errors.New("not found")
}

func (qc *StubQuorumClient) RPCCall(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	for _, arg := range args {
		method += reflect.ValueOf(arg).String()
	}
	if resp, ok := qc.mockRPC[method]; ok {
		reflect.ValueOf(result).Elem().Set(reflect.ValueOf(resp))
		return nil
	}
	return errors.New("not found")
}
