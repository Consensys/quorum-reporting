package client

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/machinebox/graphql"

	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

// QuorumClient provides access to quorum blockchain node.
type QuorumClient struct {
	wsClient      *webSocketClient
	graphqlClient *graphql.Client

	// To check we have actually shut down before returning
	shutdownChan chan struct{}
	shutdownWg   sync.WaitGroup
}

func NewQuorumClient(rawUrl, qgUrl string) (*QuorumClient, error) {
	quorumClient := &QuorumClient{
		graphqlClient: graphql.NewClient(qgUrl),
		shutdownChan:  make(chan struct{}),
	}
	var err error
	log.Debug("Connecting to Quorum WebSocket endpoint", "rawUrl", rawUrl)
	quorumClient.wsClient, err = newWebSocketClient(rawUrl)
	if err != nil {
		return nil, errors.New("connect Quorum WebSocket endpoint failed")
	}
	log.Debug("Connected to WebSocket endpoint")

	// Test graphql endpoint connection.
	log.Debug("Connecting to GraphQL endpoint", "url", qgUrl)
	var resp map[string]interface{}
	if err := quorumClient.ExecuteGraphQLQuery(&resp, CurrentBlockQuery()); err != nil || len(resp) == 0 {
		return nil, errors.New("call graphql endpoint failed")
	}
	log.Debug("Connected to GraphQL endpoint")

	// Start websocket receiver.
	go func() {
		quorumClient.shutdownWg.Add(1)
		quorumClient.wsClient.listen(quorumClient.shutdownChan)
		quorumClient.shutdownWg.Done()
	}()

	return quorumClient, nil
}

// Subscribe to chain head event.
func (qc *QuorumClient) SubscribeChainHead(ch chan<- types.RawHeader) error {
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

	rpcCallTimeout := time.NewTicker(time.Second * 1)
	defer rpcCallTimeout.Stop()
	select {
	case response := <-resultChan:
		if response == nil {
			return errors.New("nil rpc response")
		}
		log.Debug("rpc call response", "response", string(response.Result))
		if response.Error != nil {
			return response.Error
		}
		if err := json.Unmarshal(response.Result, &result); err != nil {
			// if response.Result is not a JSON, assign to result directly
			reflect.ValueOf(result).Elem().Set(reflect.ValueOf(response.Result))
		}
		return nil
	case <-rpcCallTimeout.C:
		return errors.New("rpc call timeout")
	}
}

func (qc *QuorumClient) Stop() {
	close(qc.shutdownChan)
	if qc.wsClient.conn != nil {
		qc.wsClient.conn.Close()
	}
	qc.shutdownWg.Wait()
	log.Info("Quorum client stopped")
}
