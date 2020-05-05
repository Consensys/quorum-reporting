package client

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/p2p"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/machinebox/graphql"
	"github.com/mitchellh/mapstructure"

	graphqlQuery "quorumengineering/quorum-report/graphql"
)

// QuorumClient is a wrapper to Ethereum client and provides Quorum specific access to blockchain.
type QuorumClient struct {
	*ethclient.Client
	rpcClient     *ethRPC.Client
	graphqlClient *graphql.Client
}

func NewQuorumClient(rawurl, qgurl string) (quorumClient *QuorumClient, err error) {
	rpcClient, err := ethRPC.Dial(rawurl)
	if err != nil {
		return
	}
	rawClient := ethclient.NewClient(rpcClient)
	// Create a client. (safe to share across requests)
	graphqlClient := graphql.NewClient(qgurl)
	quorumClient = &QuorumClient{rawClient, rpcClient, graphqlClient}
	// Test graphql endpoint connection.
	var resp map[string]interface{}
	err = quorumClient.ExecuteGraphQLQuery(context.Background(), &resp, graphqlQuery.CurrentBlockQuery())
	if err != nil || len(resp) == 0 {
		return nil, errors.New("call graphql endpoint failed")
	}
	return quorumClient, err
}

// Execute customized graphql query.
func (qc *QuorumClient) ExecuteGraphQLQuery(ctx context.Context, result interface{}, query string) error {
	// Build a request from query.
	req := graphql.NewRequest(query)
	// Run it and capture the response.
	return qc.graphqlClient.Run(ctx, req, &result)
}

// Execute customized rpc call.
func (qc *QuorumClient) RPCCall(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return qc.rpcClient.CallContext(ctx, result, method, args...)
}

func (qc *QuorumClient) DumpAddress(address common.Address, blockNumber uint64) (*state.DumpAccount, error) {
	dumpAccount := &state.DumpAccount{}
	err := qc.RPCCall(context.Background(), &dumpAccount, "debug_dumpAddress", address, hexutil.EncodeUint64(blockNumber))
	if err != nil {
		return nil, err
	}
	return dumpAccount, nil
}

func (qc *QuorumClient) TraceTransaction(txHash common.Hash) (map[string]interface{}, error) {
	// Trace internal calls of the transaction
	// Reference: https://github.com/ethereum/go-ethereum/issues/3128
	var resp map[string]interface{}
	type TraceConfig struct {
		Tracer string
	}
	err := qc.RPCCall(context.Background(), &resp, "debug_traceTransaction", txHash, &TraceConfig{Tracer: "callTracer"})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (qc *QuorumClient) Consensus() (string, error) {
	var resp p2p.NodeInfo
	err := qc.RPCCall(context.Background(), &resp, "admin_nodeInfo")
	if err != nil {
		return "", err
	}
	if resp.Protocols["istanbul"] != nil {
		return "istanbul", nil
	}
	protocol := resp.Protocols["eth"].(map[string]interface{})
	return protocol["consensus"].(string), nil
}

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
	for _, arg := range args {
		method += reflect.ValueOf(arg).String()
	}
	if resp, ok := qc.mockRPC[method]; ok {
		reflect.ValueOf(result).Elem().Set(reflect.ValueOf(resp))
		return nil
	}
	return errors.New("not found")
}

func (qc *StubQuorumClient) DumpAddress(address common.Address, blockNumber uint64) (*state.DumpAccount, error) {
	return nil, errors.New("not implemented")
}

func (qc *StubQuorumClient) TraceTransaction(txHash common.Hash) (map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (qc *StubQuorumClient) Consensus() (string, error) {
	return "", errors.New("not implemented")
}
