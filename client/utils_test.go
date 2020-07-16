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
	assert.Nil(t, trace)
}

func TestTraceTransaction(t *testing.T) {
	res := map[string]interface{}{
		"customField": "value",
	}
	mockRPC := map[string]interface{}{
		"debug_traceTransaction0x0000000000000000000000000000000000000000000000000000000000000000<*client.TraceConfig Value>": res,
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	trace, err := TraceTransaction(stubClient, types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000000"))
	assert.Nil(t, err)
	assert.Equal(t, res, trace)
}

func TestDumpAddress_WithError(t *testing.T) {
	mockRPC := map[string]interface{}{}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	dump, err := DumpAddress(stubClient, types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"), 1)
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

	dump, err := DumpAddress(stubClient, types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"), 1)
	assert.Nil(t, err)
	assert.EqualValues(t, &types.AccountState{
		Root:    types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		Storage: make(map[types.Hash]string),
	}, dump)
}

func TestGetCode(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_getCode0x1349f3e1b8d71effb47b840594ff27da7e603d170xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8": types.HexData("efe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
	}
	stubClient := NewStubQuorumClient(nil, mockRPC)

	blockHash := types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	code, err := GetCode(stubClient, address, blockHash)
	assert.Nil(t, err)
	assert.Equal(t, "0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36", code.String())
}

func TestGetCode_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil)

	blockHash := types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	code, err := GetCode(stubClient, address, blockHash)
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
