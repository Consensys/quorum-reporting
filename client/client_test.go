package client

import (
	"context"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}

func TestQuorumClient(t *testing.T) {
	// Create test rpc websocket server with the echo handler.
	rpcServer := httptest.NewServer(http.HandlerFunc(echo))
	defer rpcServer.Close()
	// Convert http://127.0.0.1 to ws://127.0.0.1.
	rpcurl := "ws" + strings.TrimPrefix(rpcServer.URL, "http")

	// Create test graphql server.
	graphqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(b), "block") {
			io.WriteString(w, `{
				"data": {
					"block": {
						"number": "0x6"
					}
				}
			}`)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer graphqlServer.Close()

	// Connect to the server.
	ws, _, err := websocket.DefaultDialer.Dial(rpcurl, nil)
	assert.Nil(t, err, "expected no error, but got %v", err)
	_ = ws.Close()

	_, err = NewQuorumClient("ws://invalid", "http://invalid")
	assert.NotNil(t, err, "expected error but got nil")

	_, err = NewQuorumClient(rpcurl, "http://invalid")
	assert.NotNil(t, err, "expected error but got nil")

	_, err = NewQuorumClient(rpcurl, graphqlServer.URL)
	assert.Nil(t, err, "expected no error, but got %v", err)
}

func TestStubQuorumClient(t *testing.T) {
	mockGraphQL := map[string]map[string]interface{}{
		"query": {"hello": "world"},
	}
	mockRPC := map[string]interface{}{
		"rpc_method": "hi",
	}
	blocks := []*ethTypes.Block{
		ethTypes.NewBlockWithHeader(&ethTypes.Header{Number: big.NewInt(1)}),
		ethTypes.NewBlockWithHeader(&ethTypes.Header{Number: big.NewInt(2)}),
	}
	var (
		block *ethTypes.Block
		err   error
	)
	c := NewStubQuorumClient(blocks, mockGraphQL, mockRPC)

	// test BlockByNumber
	block, err = c.BlockByNumber(context.Background(), big.NewInt(2))
	assert.Nil(t, err, "expected no error, but got %v", err)
	assert.Equal(t, common.HexToHash("0x7e9de74f52b93e8175fa5be8badb83102236ca56d5716a9ffad04192ad23ba27"), block.Hash())

	block, err = c.BlockByNumber(context.Background(), big.NewInt(3))
	assert.EqualError(t, err, "not found", "unexpected error message")

	// test mock GraphQL
	var resp map[string]interface{}
	err = c.ExecuteGraphQLQuery(&resp, "query")
	assert.Nil(t, err, "expected no error, but got %v", err)
	assert.Equal(t, "world", resp["hello"], "expected resp hello world, but got %v", resp["hello"])

	err = c.ExecuteGraphQLQuery(&resp, "random")
	assert.EqualError(t, err, "not found", "unexpected error message")

	// test mock RPC
	var res string
	err = c.RPCCall(context.Background(), &res, "rpc_method")
	assert.Nil(t, err, "expected no error, but got %v", err)
	assert.Equal(t, "hi", res, "expected res hi, but got %v", res)

	err = c.RPCCall(context.Background(), &res, "rpc_nil")
	assert.EqualError(t, err, "not found", "unexpected error message")
}
