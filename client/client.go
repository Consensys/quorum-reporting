package client

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/machinebox/graphql"
)

// NewEthereumClient connects a client to the given URL.
func NewEthereumClient(rawurl string) (rawClient *ethclient.Client, err error) {
	rawClient, err = ethclient.Dial(rawurl)
	return
}

// QuorumClient is a wrapper to Ethereum client and provides Quorum specific access to blockchain.
type QuorumClient struct {
	ec *ethclient.Client
	graphqlurl string
}

func NewQuorumClient(rawurl, graphqlurl string) (quorumClient *QuorumClient, err error) {
	rawClient, err := NewEthereumClient(rawurl)
	return &QuorumClient{rawClient, graphqlurl}, err
}

// Replication of Ethereum APIs.
// QuorumClient and EthereumClient can share a same set of interfaces as go-ethereum/interface.go
func (qc *QuorumClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return qc.ec.SubscribeNewHead(ctx, ch)
}

// Quorum specific APIs
func (qc *QuorumClient) ExecuteGraphQLQuery(ctx context.Context, query string) (respData map[string]interface{}, err error) {
	// create a client (safe to share across requests)
	client := graphql.NewClient(qc.graphqlurl)

	// make a request
	req := graphql.NewRequest(query)

	// run it and capture the response
	if err = client.Run(ctx, req, &respData); err != nil {
		log.Fatal(err)
	}
	return
}