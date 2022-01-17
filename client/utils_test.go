package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/types"
)

func TestConsensus_BadResponse(t *testing.T) {
	mockRPC := map[string]interface{}{
		"rpc_method": "hi",
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	consensus, err := Consensus(stubClient)
	assert.EqualError(t, err, "not found")
	assert.Equal(t, "", consensus)
}

func TestConsensus_IstanbulExists(t *testing.T) {
	nodeInfo := map[string]interface{}{
		"protocols": map[string]interface{}{
			"istanbul": "some value",
		},
	}
	mockRPC := map[string]interface{}{
		"admin_nodeInfo": nodeInfo,
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	consensus, err := Consensus(stubClient)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "istanbul", consensus)
}

func TestConsensus_RaftExists(t *testing.T) {
	nodeInfo := map[string]interface{}{
		"protocols": map[string]interface{}{
			"eth": map[string]interface{}{
				"consensus": "raft",
			},
		},
	}
	mockRPC := map[string]interface{}{
		"admin_nodeInfo": nodeInfo,
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	consensus, err := Consensus(stubClient)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "raft", consensus)
}

func TestTraceTransaction_WithError(t *testing.T) {
	mockRPC := map[string]interface{}{}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	trace, err := TraceTransaction(stubClient, types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000000"))
	assert.EqualError(t, err, "not found")
	assert.Equal(t, trace, types.RawOuterCall{})
}

func TestTraceTransaction(t *testing.T) {
	mockRPC := map[string]interface{}{
		"debug_traceTransaction0x0000000000000000000000000000000000000000000000000000000000000000<*client.TraceConfig Value>": types.RawOuterCall{
			Calls: []types.RawInnerCall{{}},
		},
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	trace, err := TraceTransaction(stubClient, types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000000"))
	assert.Nil(t, err)
	assert.Len(t, trace.Calls, 1)
}

func TestDumpAddress_WithError(t *testing.T) {
	mockRPC := map[string]interface{}{}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	dump, err := DumpAddress(stubClient, types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"), 0, 1, false)
	assert.EqualError(t, err, "not found")
	assert.Nil(t, dump)
}

func TestDumpAddress(t *testing.T) {
	res := &types.RawAccountState{
		Root: types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
	}
	mockRPC := map[string]interface{}{
		"debug_dumpAddress0x1349f3e1b8d71effb47b840594ff27da7e603d170x1": res,
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	dump, err := DumpAddress(stubClient, types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"), 0, 1, false)
	assert.Nil(t, err)
	assert.EqualValues(t, &types.AccountState{
		Root:    types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		Storage: make(map[types.Hash]string),
	}, dump)
}

func TestGetCode(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_getCode0x1349f3e1b8d71effb47b840594ff27da7e603d170x5": types.HexData("efe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	blockNum := uint64(5)
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	code, err := GetCode(stubClient, address, blockNum)
	assert.Nil(t, err)
	assert.Equal(t, "0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36", code.String())
}

func TestGetCode_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	blockNum := uint64(5)
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	code, err := GetCode(stubClient, address, blockNum)
	assert.EqualError(t, err, "not found")
	assert.Equal(t, types.HexData(""), code)
}

func TestEIP165(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<types.EIP165Call Value>0x2": types.HexData("0000000000000000000000000000000000000000000000000000000000000001"),
	}

	stubClient := NewStubQuorumClient(nil, mockRPC)

	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234"), 2)
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestEIP165_WithWrongInterfaceLengthError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234567890"), 0)
	assert.EqualError(t, err, "interfaceId wrong size")
	assert.False(t, exists)
}

func TestEIP165_WithClientError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234"), 0)
	assert.EqualError(t, err, "not found")
	assert.False(t, exists)
}

func TestEIP165_WithWrongSizeResult(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<types.EIP165Call Value>0x1": types.HexData(""),
	}

	stubClient := NewStubQuorumClient(nil, mockRPC)

	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234"), 1)
	assert.Nil(t, err)
	assert.False(t, exists)
}

func TestCurrentBlock(t *testing.T) {
	mockGraphQL := map[string]map[string]interface{}{
		CurrentBlockQuery(): {"block": interface{}(map[string]interface{}{"number": "0x10"})},
	}
	stubClient := NewStubQuorumClient(mockGraphQL, nil)

	currentBlockNumber, err := CurrentBlock(stubClient)

	assert.Nil(t, err)
	assert.EqualValues(t, 16, currentBlockNumber)
}

func TestCurrentBlock_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	currentBlockNumber, err := CurrentBlock(stubClient)

	assert.EqualError(t, err, "not found")
	assert.EqualValues(t, 0, currentBlockNumber)
}

func TestTransactionWithReceipt(t *testing.T) {
	testTransactionHash := types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")
	fullGraphQLTransaction := map[string]interface{}{
		"hash":              "0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8",
		"status":            "0x1",
		"block":             map[string]interface{}{"number": "0x2", "timestamp": "0x1000"},
		"index":             0,
		"nonce":             "0x1",
		"from":              map[string]interface{}{"address": "0xed9d02e382b34818e88b88a309c7fe71e65f419d"},
		"to":                nil,
		"value":             "0x0",
		"gasPrice":          "0x0",
		"gas":               "0x47b760",
		"gasUsed":           "0x280a7",
		"cumulativeGasUsed": "0x280a7",
		"createdContract":   map[string]interface{}{"address": "0x1349f3e1b8d71effb47b840594ff27da7e603d17"},
		"inputData":         "0x608060405234801561001057600080fd5b506040516020806101a18339810180604052602081101561003057600080fd5b81019080805190602001909291905050508060008190555050610149806100586000396000f3fe608060405234801561001057600080fd5b506004361061005e576000357c0100000000000000000000000000000000000000000000000000000000900480632a1afcd91461006357806360fe47b1146100815780636d4ce63c146100af575b600080fd5b61006b6100cd565b6040518082815260200191505060405180910390f35b6100ad6004803603602081101561009757600080fd5b81019080803590602001909291905050506100d3565b005b6100b7610114565b6040518082815260200191505060405180910390f35b60005481565b806000819055507fefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36816040518082815260200191505060405180910390a150565b6000805490509056fea165627a7a7230582061f6956b053dbf99873b363ab3ba7bca70853ba5efbaff898cd840d71c54fc1d0029000000000000000000000000000000000000000000000000000000000000002a",
		"privateInputData":  "0x",
		"isPrivate":         false,
		"logs": []map[string]interface{}{
			{
				"index":   0,
				"account": map[string]interface{}{"address": "0x1349f3e1b8d71effb47b840594ff27da7e603d17"},
				"topics":  []string{"0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"},
				"data":    "0x0000000000000000000000000000000000000000000000000000000000000042",
			},
		},
	}
	fullGraphQLQuery := `query { transaction(hash:"0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8") {
        hash
        status
		index
        nonce
        from { address }
        to { address }
        value
        gasPrice
        gas
        gasUsed
        cumulativeGasUsed
        createdContract { address }
		inputData
		privateInputData
		isPrivate
		logs {
			index
			account { address }
			topics
			data
		}
    } }`

	mockGraphQL := map[string]map[string]interface{}{fullGraphQLQuery: {"transaction": fullGraphQLTransaction}}
	stubClient := NewStubQuorumClient(mockGraphQL, nil)

	result, err := TransactionWithReceipt(stubClient, testTransactionHash)

	expectedResult := Transaction{
		Hash:              "e625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8",
		Status:            "0x1",
		Index:             0,
		Nonce:             types.HexNumber(1),
		From:              Address{Address: "ed9d02e382b34818e88b88a309c7fe71e65f419d"},
		To:                Address{},
		Value:             types.HexNumber(0),
		GasPrice:          types.HexNumber(0),
		Gas:               types.HexNumber(4700000),
		GasUsed:           types.HexNumber(164007),
		CumulativeGasUsed: types.HexNumber(164007),
		CreatedContract:   Address{Address: "1349f3e1b8d71effb47b840594ff27da7e603d17"},
		InputData:         "608060405234801561001057600080fd5b506040516020806101a18339810180604052602081101561003057600080fd5b81019080805190602001909291905050508060008190555050610149806100586000396000f3fe608060405234801561001057600080fd5b506004361061005e576000357c0100000000000000000000000000000000000000000000000000000000900480632a1afcd91461006357806360fe47b1146100815780636d4ce63c146100af575b600080fd5b61006b6100cd565b6040518082815260200191505060405180910390f35b6100ad6004803603602081101561009757600080fd5b81019080803590602001909291905050506100d3565b005b6100b7610114565b6040518082815260200191505060405180910390f35b60005481565b806000819055507fefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36816040518082815260200191505060405180910390a150565b6000805490509056fea165627a7a7230582061f6956b053dbf99873b363ab3ba7bca70853ba5efbaff898cd840d71c54fc1d0029000000000000000000000000000000000000000000000000000000000000002a",
		PrivateInputData:  "",
		IsPrivate:         false,
		Logs: []Event{
			{
				Index:   0,
				Account: Address{Address: "1349f3e1b8d71effb47b840594ff27da7e603d17"},
				Topics:  []types.Hash{"efe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"},
				Data:    "0000000000000000000000000000000000000000000000000000000000000042",
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestTransactionWithReceipt_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	result, err := TransactionWithReceipt(stubClient, types.NewHash(""))

	assert.EqualError(t, err, "not found")
	assert.Equal(t, Transaction{}, result)
}

func TestCallBalanceOfERC20_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	tokenContract := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	holder := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contractCallResult, err := CallBalanceOfERC20(stubClient, tokenContract, holder, 1)
	assert.EqualError(t, err, "not found")
	assert.Equal(t, types.HexData(""), contractCallResult)
}

func TestCallBalanceOfERC20(t *testing.T) {
	//TODO: check that the values inside the call data are correct
	mockRPC := map[string]interface{}{
		"eth_call<types.EIP165Call Value>0x1": types.NewHexData("0x12345"),
	}

	stubClient := NewStubQuorumClient(nil, mockRPC)

	tokenContract := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	holder := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contractCallResult, err := CallBalanceOfERC20(stubClient, tokenContract, holder, 1)
	assert.Nil(t, err)
	assert.Equal(t, types.HexData("12345"), contractCallResult)
}

func TestStorageRoot_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	result, err := StorageRoot(stubClient, types.NewAddress(""), 1)
	assert.EqualError(t, err, "not found")
	assert.EqualValues(t, "", result)
}

func TestStorageRoot(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_storageRoot0x00000000000000000000000000000000000000000x1": types.NewHash("1"),
	}

	stubClient := NewStubQuorumClient(nil, mockRPC)

	result, err := StorageRoot(stubClient, types.NewAddress(""), 1)

	assert.Nil(t, err)
	assert.EqualValues(t, "0000000000000000000000000000000000000000000000000000000000000001", result)
}
