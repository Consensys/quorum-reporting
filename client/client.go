package client

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/machinebox/graphql"
)

// QuorumClient is a wrapper to Ethereum client and provides Quorum specific access to blockchain.
type QuorumClient struct {
	*ethclient.Client
	qgurl string
}

func NewQuorumClient(rawurl, qgurl string) (quorumClient *QuorumClient, err error) {
	rawClient, err := ethclient.Dial(rawurl)
	return &QuorumClient{rawClient, qgurl}, err
}

// Execute customized graphql query
func (qc *QuorumClient) ExecuteGraphQLQuery(ctx context.Context, query string) (respData map[string]interface{}, err error) {
	// create a client (safe to share across requests)
	client := graphql.NewClient(qc.qgurl)

	// make a request
	req := graphql.NewRequest(query)

	// run it and capture the response
	err = client.Run(ctx, req, &respData)
	return
}
