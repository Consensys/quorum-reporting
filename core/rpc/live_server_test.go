package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/types"
)

type rpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

var (
	apiDatabase  = memory.NewMemoryDB()
	testHttpAddr = "http://localhost:30000"
)

func TestMain(m *testing.M) {
	_ = apiDatabase.AddAddresses([]common.Address{address, common.HexToAddress("0x0000000000000000000000000000000000000009")})
	_ = apiDatabase.WriteBlocks([]*types.Block{block})
	_ = apiDatabase.WriteTransactions([]*types.Transaction{tx1, tx2, tx3})
	_ = apiDatabase.IndexBlocks([]common.Address{address}, []*types.Block{block})

	rpcServer := SetupRpcServer(apiDatabase)
	if err := rpcServer.Start(); err != nil {
		fmt.Printf("Failed to start test RPC server: %s\n", err.Error())
		os.Exit(1)
	}
	time.Sleep(2 * time.Second)

	exitCode := m.Run()

	rpcServer.Stop()
	os.Exit(exitCode)
}

func SetupRpcServer(db database.Database) *RPCService {
	errorChan := make(chan error)
	serverConfig := struct {
		RPCAddr     string   `toml:"rpcAddr"`
		RPCCorsList []string `toml:"rpcCorsList,omitempty"`
		RPCVHosts   []string `toml:"rpcvHosts,omitempty"`
		UIPort      int      `toml:"uiPort,omitempty"`
	}{
		RPCAddr:     "localhost:30000",
		RPCCorsList: []string{"*"},
		RPCVHosts:   nil,
		UIPort:      0,
	}
	config := types.ReportingConfig{Server: serverConfig}

	return NewRPCService(db, config, errorChan)
}

//TODO: error case
func TestRPCAPIs_GetLastPersistedBlockNumber(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetLastPersistedBlockNumber",
		Params:  json.RawMessage("[]"),
	}

	rpcResponse, err := doRequest(msg)

	assert.Nil(t, err)
	assert.EqualValues(t, "1", rpcResponse.Result)
}

func TestRPCAPIs_GetBlock(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetBlock",
		Params:  json.RawMessage("[1]"),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var retrievedBlock types.Block
	_ = json.Unmarshal(rpcResponse.Result, &retrievedBlock)

	assert.Equal(t, "null", string(rpcResponse.Error))
	assert.Equal(t, block, &retrievedBlock)
}

func TestRPCAPIs_GetBlock_BlockNotExist(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetBlock",
		Params:  json.RawMessage("[2]"),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var errorMessage string
	_ = json.Unmarshal(rpcResponse.Error, &errorMessage)

	assert.Equal(t, "block does not exist", errorMessage)
	assert.Equal(t, "null", string(rpcResponse.Result))
}

func TestRPCAPIs_GetContractCreationTransaction(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetContractCreationTransaction",
		Params:  json.RawMessage(fmt.Sprintf(`["%s"]`, address.Hex())),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var txHash string
	_ = json.Unmarshal(rpcResponse.Result, &txHash)

	assert.Equal(t, "null", string(rpcResponse.Error))
	assert.Equal(t, "0x1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7", txHash)
}

func TestRPCAPIs_GetContractCreationTransaction_CreationTxNotFound(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetContractCreationTransaction",
		Params:  json.RawMessage(fmt.Sprintf(`["0x0000000000000000000000000000000000000009"]`)),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var errorMessage string
	_ = json.Unmarshal(rpcResponse.Error, &errorMessage)

	assert.Equal(t, "contract creation tx not found", errorMessage)
	assert.Equal(t, "null", string(rpcResponse.Result))
}

func TestRPCAPIs_GetContractCreationTransaction_AddressNotIndexed(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetContractCreationTransaction",
		Params:  json.RawMessage(fmt.Sprintf(`["0x0000000000000000000000000000000000000010"]`)),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var errorMessage string
	_ = json.Unmarshal(rpcResponse.Error, &errorMessage)

	assert.Equal(t, "address is not registered", errorMessage)
	assert.Equal(t, "null", string(rpcResponse.Result))
}

//TODO: error cases + given QueryOptions
func TestRPCAPIs_GetAllTransactionsToAddress(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetAllTransactionsToAddress",
		Params:  json.RawMessage(fmt.Sprintf(`[{"address": "0x0000000000000000000000000000000000000001"}]`)),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var result TransactionsResp
	_ = json.Unmarshal(rpcResponse.Result, &result)

	expectedOptions := &types.QueryOptions{}
	expectedOptions.SetDefaults()

	assert.Equal(t, "null", string(rpcResponse.Error))
	assert.EqualValues(t, 2, result.Total)
	assert.Contains(t, result.Transactions, tx2.Hash)
	assert.Contains(t, result.Transactions, tx3.Hash)
	assert.Equal(t, result.Options, expectedOptions)
}

//TODO: error cases + given QueryOptions
func TestRPCAPIs_GetAllTransactionsInternalToAddress(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetAllTransactionsInternalToAddress",
		Params:  json.RawMessage(fmt.Sprintf(`[{"address": "0x0000000000000000000000000000000000000001"}]`)),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var result TransactionsResp
	_ = json.Unmarshal(rpcResponse.Result, &result)

	expectedOptions := &types.QueryOptions{}
	expectedOptions.SetDefaults()

	assert.Equal(t, "null", string(rpcResponse.Error))
	assert.EqualValues(t, 1, result.Total)
	assert.Contains(t, result.Transactions, tx3.Hash)
	assert.Equal(t, result.Options, expectedOptions)
}

func TestNewRPCAPIs_AddAddress_WithEmptyAddress(t *testing.T) {
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.AddAddress",
		Params:  json.RawMessage(fmt.Sprintf(`[{}]`)),
	}

	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)

	var errorMessage string
	_ = json.Unmarshal(rpcResponse.Error, &errorMessage)
	assert.Equal(t, "address not provided", errorMessage)
}

func TestNewRPCAPIs_AddAddress(t *testing.T) {
	//address not in before
	msgBefore := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetAddresses",
		Params:  json.RawMessage("[]"),
	}
	rpcResponseBefore, err := doRequest(msgBefore)
	assert.Nil(t, err)
	var resultBefore []common.Address
	_ = json.Unmarshal(rpcResponseBefore.Result, &resultBefore)
	assert.NotContains(t, resultBefore, common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"))

	//add the address to the DB
	msg := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.AddAddress",
		Params:  json.RawMessage(fmt.Sprintf(`[{"address": "0x1349f3e1b8d71effb47b840594ff27da7e603d17"}]`)),
	}
	rpcResponse, err := doRequest(msg)
	assert.Nil(t, err)
	assert.Equal(t, "null", string(rpcResponse.Error))

	//address is in after
	msgAfter := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetAddresses",
		Params:  json.RawMessage("[]"),
	}
	rpcResponseAfter, err := doRequest(msgAfter)
	assert.Nil(t, err)
	var resultAfter []common.Address
	_ = json.Unmarshal(rpcResponseAfter.Result, &resultAfter)
	assert.Contains(t, resultAfter, common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"))

	//delete the address from the database
	msgDelete := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.DeleteAddress",
		Params:  json.RawMessage(fmt.Sprintf(`["0x1349f3e1b8d71effb47b840594ff27da7e603d17"]`)),
	}
	rpcResponseDelete, err := doRequest(msgDelete)
	assert.Nil(t, err)
	assert.Equal(t, "null", string(rpcResponseDelete.Error))

	//address no longer present
	msgAfterDelete := rpcMessage{
		Version: "2.0",
		ID:      "67",
		Method:  "reporting.GetAddresses",
		Params:  json.RawMessage("[]"),
	}
	rpcResponseAfterDelete, err := doRequest(msgAfterDelete)
	assert.Nil(t, err)
	var resultAfterDelete []common.Address
	_ = json.Unmarshal(rpcResponseAfterDelete.Result, &resultAfterDelete)
	assert.NotContains(t, resultAfterDelete, common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"))
}

func doRequest(request rpcMessage) (rpcMessage, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return rpcMessage{}, err
	}

	resp, err := http.Post(testHttpAddr, "application/json", buf)
	if err != nil {
		return rpcMessage{}, err
	}

	defer resp.Body.Close()

	var rpcResponse rpcMessage
	if err := json.NewDecoder(resp.Body).Decode(&rpcResponse); err != nil {
		return rpcMessage{}, err
	}
	return rpcResponse, nil
}
