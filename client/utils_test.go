package client

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/stretchr/testify/assert"
)

func TestConsensus_BadResponse(t *testing.T) {
	mockRPC := map[string]interface{}{
		"rpc_method": "hi",
	}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	consensus, err := Consensus(stubClient)
	assert.EqualError(t, err, "not found")
	assert.Equal(t, "", consensus)
}

func TestConsensus_IstanbulExists(t *testing.T) {
	nodeInfo := p2p.NodeInfo{
		Protocols: map[string]interface{}{
			"istanbul": "some value",
		},
	}
	mockRPC := map[string]interface{}{
		"admin_nodeInfo": nodeInfo,
	}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	consensus, err := Consensus(stubClient)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "istanbul", consensus)
}

func TestConsensus_RaftExists(t *testing.T) {
	nodeInfo := p2p.NodeInfo{
		Protocols: map[string]interface{}{
			"eth": map[string]interface{}{
				"consensus": "raft",
			},
		},
	}
	mockRPC := map[string]interface{}{
		"admin_nodeInfo": nodeInfo,
	}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	consensus, err := Consensus(stubClient)
	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "raft", consensus)
}

func TestTraceTransaction_WithError(t *testing.T) {
	mockRPC := map[string]interface{}{}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	trace, err := TraceTransaction(stubClient, common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"))
	assert.EqualError(t, err, "not found")
	assert.Nil(t, trace)
}

func TestTraceTransaction(t *testing.T) {
	res := map[string]interface{}{
		"customField": "value",
	}
	mockRPC := map[string]interface{}{
		"debug_traceTransaction<common.Hash Value><*client.TraceConfig Value>": res,
	}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	trace, err := TraceTransaction(stubClient, common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"))
	assert.Nil(t, err)
	assert.Equal(t, res, trace)
}

func TestDumpAddress_WithError(t *testing.T) {
	mockRPC := map[string]interface{}{}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	dump, err := DumpAddress(stubClient, common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"), 1)
	assert.EqualError(t, err, "not found")
	assert.Nil(t, dump)
}

func TestDumpAddress(t *testing.T) {
	res := &state.DumpAccount{
		Balance:  "10",
		Nonce:    5,
		Root:     "some root",
		CodeHash: "some code hash",
		Code:     "some code",
	}
	mockRPC := map[string]interface{}{
		"debug_dumpAddress<common.Address Value>0x1": res,
	}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	dump, err := DumpAddress(stubClient, common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"), 1)
	assert.Nil(t, err)
	assert.Equal(t, res, dump)
}

func TestGetCode(t *testing.T) {
	sampleCode, _ := hexutil.Decode("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36")
	mockRPC := map[string]interface{}{
		"eth_getCode<common.Address Value>0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8": sampleCode,
	}
	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	blockHash := common.HexToHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")
	address := common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	code, err := GetCode(stubClient, address, blockHash)
	assert.Nil(t, err)
	assert.Equal(t, "0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36", code.String())
}

func TestGetCode_WithError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil, nil)

	blockHash := common.HexToHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")
	address := common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	code, err := GetCode(stubClient, address, blockHash)
	assert.EqualError(t, err, "not found")
	assert.Nil(t, code)
}

func TestEIP165(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<ethereum.CallMsg Value><*big.Int Value>": common.LeftPadBytes([]byte{1}, 32),
	}

	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	address := common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234"), big.NewInt(2))
	assert.Nil(t, err)
	assert.True(t, exists)
}

func TestEIP165_WithWrongInterfaceLengthError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil, nil)

	address := common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234567890"), new(big.Int))
	assert.EqualError(t, err, "interfaceId wrong size")
	assert.False(t, exists)
}

func TestEIP165_WithClientError(t *testing.T) {
	stubClient := NewStubQuorumClient(nil, nil, nil)

	address := common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234"), new(big.Int))
	assert.EqualError(t, err, "not found")
	assert.False(t, exists)
}

func TestEIP165_WithWrongSizeResult(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<ethereum.CallMsg Value><*big.Int Value>": []byte{},
	}

	stubClient := NewStubQuorumClient(nil, nil, mockRPC)

	address := common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	exists, err := CallEIP165(stubClient, address, []byte("1234"), big.NewInt(1))
	assert.Nil(t, err)
	assert.False(t, exists)
}
