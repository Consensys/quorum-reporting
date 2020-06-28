package client

import (
	"context"
	"encoding/json"
	"errors"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/machinebox/graphql"
	"time"

	graphqlQuery "quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/log"
)

// QuorumClient provides access to quorum blockchain node.
type QuorumClient struct {
	wsClient      *webSocketClient
	graphqlClient *graphql.Client
}

func NewQuorumClient(rawUrl, qgUrl string) (*QuorumClient, error) {
	quorumClient := &QuorumClient{nil, graphql.NewClient(qgUrl)}

	log.Debug("Connecting to Quorum WebSocket endpoint", "rawUrl", rawUrl)
	wsClient, err := newWebSocketClient(rawUrl)
	if err != nil {
		return nil, errors.New("connect Quorum WebSocket endpoint failed")
	}
	quorumClient.wsClient = wsClient
	log.Debug("Connected to WebSocket endpoint")

	// Test graphql endpoint connection.
	log.Debug("Connecting to GraphQL endpoint", "url", qgUrl)
	var resp map[string]interface{}
	if err := quorumClient.ExecuteGraphQLQuery(&resp, graphqlQuery.CurrentBlockQuery()); err != nil || len(resp) == 0 {
		return nil, errors.New("call graphql endpoint failed")
	}
	log.Debug("Connected to GraphQL endpoint")

	// Start websocket receiver.
	go quorumClient.wsClient.listen()

	return quorumClient, nil
}

// Subscribe to chain head event.
func (qc *QuorumClient) SubscribeChainHead(ch chan<- *ethTypes.Header) error {
	return qc.wsClient.subscribeChainHead(ch)
}

// Execute customized graphql query.
func (qc *QuorumClient) ExecuteGraphQLQuery(result interface{}, query string) error {
	// Build a request from query.
	req := graphql.NewRequest(query)
	// Run it and capture the response.
	return qc.graphqlClient.Run(context.Background(), req, &result)
}

// Execute customized rpc call.
func (qc *QuorumClient) RPCCall(result interface{}, method string, args ...interface{}) error {
	resultChan := make(chan *message, 1)
	err := qc.wsClient.sendRPCMsg(resultChan, method, args...)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	select {
	case response := <-resultChan:
		log.Debug("rpc call response", "response", string(response.Result))
		if response.Error != nil {
			return response.Error
		}
		return json.Unmarshal(response.Result, &result)
	case <-ticker.C:
		return errors.New("rpc call timeout")
	}
}
