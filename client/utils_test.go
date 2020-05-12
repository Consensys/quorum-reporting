package client

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
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
