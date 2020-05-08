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
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	defer ws.Close()
	_, err = NewQuorumClient("ws://invalid", "http://invalid")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	_, err = NewQuorumClient(rpcurl, "http://invalid")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	_, err = NewQuorumClient(rpcurl, graphqlServer.URL)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
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
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if block.Hash() != common.HexToHash("0x7e9de74f52b93e8175fa5be8badb83102236ca56d5716a9ffad04192ad23ba27") {
		t.Fatalf("expected hash %v, but got %v", common.HexToHash("0x7e9de74f52b93e8175fa5be8badb83102236ca56d5716a9ffad04192ad23ba27").Hex(), block.Hash().Hex())
	}
	block, err = c.BlockByNumber(context.Background(), big.NewInt(3))
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected error %v, but got %v", "not found", err)
	}
	// test mock GraphQL
	var resp map[string]interface{}
	err = c.ExecuteGraphQLQuery(context.Background(), &resp, "query")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if resp["hello"] != "world" {
		t.Fatalf("expected resp hello world, but got %v", resp["hello"])
	}
	err = c.ExecuteGraphQLQuery(context.Background(), &resp, "random")
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected error %v, but got %v", "not found", err)
	}
	// test mock RPC
	var res string
	err = c.RPCCall(context.Background(), &res, "rpc_method")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if res != "hi" {
		t.Fatalf("expected res hi, but got %v", res)
	}
	err = c.RPCCall(context.Background(), &res, "rpc_nil")
	if err == nil || err.Error() != "not found" {
		t.Fatalf("expected error %v, but got %v", "not found", err)
	}
}
