package client

import (
	"encoding/json"
	"errors"
	"reflect"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
)

// StubQuorumClient is used for unit test.
type StubQuorumClient struct {
	mockGraphQL map[string]map[string]interface{}
	mockRPC     map[string]interface{}
}

func NewStubQuorumClient(mockGraphQL map[string]map[string]interface{}, mockRPC map[string]interface{}) *StubQuorumClient {
	if mockGraphQL == nil {
		mockGraphQL = map[string]map[string]interface{}{}
	}
	if mockRPC == nil {
		mockRPC = map[string]interface{}{}
	}
	return &StubQuorumClient{mockGraphQL, mockRPC}
}

func (qc *StubQuorumClient) SubscribeChainHead(chan<- *ethTypes.Header) error {
	return errors.New("not implemented")
}

func (qc *StubQuorumClient) ExecuteGraphQLQuery(result interface{}, query string) error {
	if resp, ok := qc.mockGraphQL[query]; ok {
		out, _ := json.Marshal(resp)
		return json.Unmarshal(out, &result)
	}
	return errors.New("not found")
}

func (qc *StubQuorumClient) RPCCall(result interface{}, method string, args ...interface{}) error {
	for _, arg := range args {
		method += reflect.ValueOf(arg).String()
	}
	if resp, ok := qc.mockRPC[method]; ok {
		reflect.ValueOf(result).Elem().Set(reflect.ValueOf(resp))
		return nil
	}
	return errors.New("not found")
}

func (qc *StubQuorumClient) Stop() {}
