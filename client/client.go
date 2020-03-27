package client

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/machinebox/graphql"

	graphqlQuery "quorumengineering/quorum-report/graphql"
)

// QuorumClient is a wrapper to Ethereum client and provides Quorum specific access to blockchain.
type QuorumClient struct {
	*ethclient.Client
	graphqlClient *graphql.Client
}

func NewQuorumClient(rawurl, qgurl string) (quorumClient *QuorumClient, err error) {
	rawClient, err := ethclient.Dial(rawurl)
	if err != nil {
		return
	}
	// Create a client. (safe to share across requests)
	graphqlClient := graphql.NewClient(qgurl)
	quorumClient = &QuorumClient{rawClient, graphqlClient}
	// Test graphql endpoint connection.
	resp, err := quorumClient.ExecuteGraphQLQuery(context.Background(), graphqlQuery.CurrentBlockQuery())
	if err != nil {
		return nil, err
	}
	if len(resp) == 0 {
		return nil, errors.New("call graphql endpoint failed")
	}
	return quorumClient, err
}

// Execute customized graphql query.
func (qc *QuorumClient) ExecuteGraphQLQuery(ctx context.Context, query string) (respData map[string]interface{}, err error) {
	// Build a request from query.
	req := graphql.NewRequest(query)
	// Run it and capture the response.
	err = qc.graphqlClient.Run(ctx, req, &respData)
	return
}
