package client

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/ethclient"
	ethRPC "github.com/ethereum/go-ethereum/rpc"
	"github.com/machinebox/graphql"

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
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
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
