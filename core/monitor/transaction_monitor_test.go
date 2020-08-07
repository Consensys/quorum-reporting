package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

var (
	graphqlResp = map[string]interface{}{
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
)

func TestCreateTransaction(t *testing.T) {
	testBlock := &types.Block{
		Number:    2,
		Timestamp: uint64(0x1000),
	}
	mockGraphQL := map[string]map[string]interface{}{
		client.TransactionDetailQuery(types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")): {
			"transaction": interface{}(graphqlResp),
		},
	}
	mockRPC := map[string]interface{}{
		"debug_traceTransaction0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8<*client.TraceConfig Value>": map[string]interface{}{
			"calls": []interface{}{
				map[string]interface{}{
					"from":    "0x9d13c6d3afe1721beef56b55d303b09e021e27ab",
					"gas":     "0x279e",
					"gasUsed": "0x18aa",
					"input":   "0x60fe47b10000000000000000000000000000000000000000000000000000000000000042",
					"output":  "0x",
					"to":      "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
					"type":    "CALL",
					"value":   "0x0",
				},
			},
		},
	}
	tm := NewDefaultTransactionMonitor(client.NewStubQuorumClient(mockGraphQL, mockRPC))
	tx, err := tm.createTransaction(testBlock, types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"))
	assert.Nil(t, err)
	assert.EqualValues(t, types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"), tx.Hash)
	assert.True(t, tx.Status)
	assert.EqualValues(t, 2, tx.BlockNumber)
	assert.EqualValues(t, 0, tx.Index)
	assert.EqualValues(t, types.NewAddress("0xed9d02e382b34818e88b88a309c7fe71e65f419d"), tx.From)
	assert.EqualValues(t, 4700000, tx.Gas)
	assert.EqualValues(t, "0x608060405234801561001057600080fd5b506040516020806101a18339810180604052602081101561003057600080fd5b81019080805190602001909291905050508060008190555050610149806100586000396000f3fe608060405234801561001057600080fd5b506004361061005e576000357c0100000000000000000000000000000000000000000000000000000000900480632a1afcd91461006357806360fe47b1146100815780636d4ce63c146100af575b600080fd5b61006b6100cd565b6040518082815260200191505060405180910390f35b6100ad6004803603602081101561009757600080fd5b81019080803590602001909291905050506100d3565b005b6100b7610114565b6040518082815260200191505060405180910390f35b60005481565b806000819055507fefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36816040518082815260200191505060405180910390a150565b6000805490509056fea165627a7a7230582061f6956b053dbf99873b363ab3ba7bca70853ba5efbaff898cd840d71c54fc1d0029000000000000000000000000000000000000000000000000000000000000002a", tx.Data.String())
	assert.EqualValues(t, "0x", tx.PrivateData.String())
	assert.False(t, tx.IsPrivate)

	assert.Len(t, tx.Events, 1)
	assert.EqualValues(t, types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"), tx.Events[0].Topics[0])
	assert.Len(t, tx.InternalCalls, 1)
}

func TestTransactionMonitor_PullTransactions(t *testing.T) {
	mockGraphQL := map[string]map[string]interface{}{
		client.TransactionDetailQuery(types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8")): {
			"transaction": interface{}(graphqlResp),
		},
	}
	mockRPC := map[string]interface{}{
		"debug_traceTransaction0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8<*client.TraceConfig Value>": map[string]interface{}{
			"calls": []interface{}{
				map[string]interface{}{
					"from":    "0x9d13c6d3afe1721beef56b55d303b09e021e27ab",
					"gas":     "0x279e",
					"gasUsed": "0x18aa",
					"input":   "0x60fe47b10000000000000000000000000000000000000000000000000000000000000042",
					"output":  "0x",
					"to":      "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
					"type":    "CALL",
					"value":   "0x0",
				},
			},
		},
	}
	block := &types.Block{
		Hash:   types.NewHash("0xd3b57e8a791a134ddf47772f12fdddbf67480377e633bf55f411166d3be7d66f"),
		Number: 2,
		Transactions: []types.Hash{
			types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		},
	}

	tm := NewDefaultTransactionMonitor(client.NewStubQuorumClient(mockGraphQL, mockRPC))

	txs, err := tm.PullTransactions(block)
	assert.Nil(t, err, "unexpected error")
	assert.Len(t, txs, 1)

	tx := txs[0]
	assert.EqualValues(t, types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"), tx.Hash)
	assert.True(t, tx.Status)
	assert.EqualValues(t, 2, tx.BlockNumber)
	assert.EqualValues(t, 0, tx.Index)
	assert.EqualValues(t, types.NewAddress("0xed9d02e382b34818e88b88a309c7fe71e65f419d"), tx.From)
	assert.EqualValues(t, 4700000, tx.Gas)
	assert.EqualValues(t, "0x608060405234801561001057600080fd5b506040516020806101a18339810180604052602081101561003057600080fd5b81019080805190602001909291905050508060008190555050610149806100586000396000f3fe608060405234801561001057600080fd5b506004361061005e576000357c0100000000000000000000000000000000000000000000000000000000900480632a1afcd91461006357806360fe47b1146100815780636d4ce63c146100af575b600080fd5b61006b6100cd565b6040518082815260200191505060405180910390f35b6100ad6004803603602081101561009757600080fd5b81019080803590602001909291905050506100d3565b005b6100b7610114565b6040518082815260200191505060405180910390f35b60005481565b806000819055507fefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36816040518082815260200191505060405180910390a150565b6000805490509056fea165627a7a7230582061f6956b053dbf99873b363ab3ba7bca70853ba5efbaff898cd840d71c54fc1d0029000000000000000000000000000000000000000000000000000000000000002a", tx.Data.String())
	assert.EqualValues(t, "0x", tx.PrivateData.String())
	assert.False(t, tx.IsPrivate)

	assert.Len(t, tx.Events, 1)
	assert.EqualValues(t, types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"), tx.Events[0].Topics[0])
	assert.Len(t, tx.InternalCalls, 1)
}
