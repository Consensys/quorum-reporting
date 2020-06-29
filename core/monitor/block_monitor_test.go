package monitor

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/types"
)

func TestCreateBlock(t *testing.T) {
	cases := []struct {
		originalBlock *types.RawBlock
		expectedBlock *types.Block
		consensus     string
	}{
		{
			&types.RawBlock{
				Number:    "0x2A",
				Timestamp: "0x3B9ACA00",
				GasLimit:  "0x2fa023db",
				GasUsed:   "0x6b8a",
			},
			&types.Block{
				Number:       uint64(42),
				Timestamp:    1_000_000_000,
				Transactions: []common.Hash{},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"istanbul",
		},
		{
			&types.RawBlock{
				Number:       "0x2A",
				Timestamp:    "0x3B9ACA00",
				Transactions: []string{"0x0000000000000000000000000000000000000000000000000000000000000000"},
				GasLimit:     "0x2fa023db",
				GasUsed:      "0x6b8a",
			},
			&types.Block{
				Number:       uint64(42),
				Timestamp:    1_000_000_000,
				Transactions: []common.Hash{common.BigToHash(big.NewInt(0))},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"istanbul",
		},
		{
			&types.RawBlock{
				Number:    "0x2A",
				Timestamp: "0x3B9ACA00",
				GasLimit:  "0x2fa023db",
				GasUsed:   "0x6b8a",
			},
			&types.Block{
				Number:       uint64(42),
				Timestamp:    1,
				Transactions: []common.Hash{},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"raft",
		},
		{
			&types.RawBlock{
				Number:       "0x2A",
				Timestamp:    "0x3B9ACA00",
				Transactions: []string{"0x0000000000000000000000000000000000000000000000000000000000000000"},
				GasLimit:     "0x2fa023db",
				GasUsed:      "0x6b8a",
			},
			&types.Block{
				Number:       uint64(42),
				Timestamp:    1,
				Transactions: []common.Hash{common.BigToHash(big.NewInt(0))},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"raft",
		},
	}

	for _, tc := range cases {
		bm := NewDefaultBlockMonitor(client.NewStubQuorumClient(nil, nil), nil, tc.consensus)
		actual := bm.createBlock(tc.originalBlock)
		if actual.Number != tc.expectedBlock.Number {
			t.Fatalf("expected block number %v, but got %v", tc.expectedBlock.Number, actual.Number)
		}
		if tc.consensus == "raft" && actual.Timestamp != tc.expectedBlock.Timestamp {
			t.Fatalf("expected timestamp %d for raft, but got %v", tc.expectedBlock.Timestamp, actual.Timestamp)
		} else if actual.Timestamp != tc.expectedBlock.Timestamp {
			t.Fatalf("expected timestamp %d for %s, but got %v", tc.expectedBlock.Timestamp, tc.consensus, actual.Timestamp)
		}
		if len(actual.Transactions) != len(tc.expectedBlock.Transactions) {
			t.Fatalf("expected %v transactions, but got %v", len(tc.expectedBlock.Transactions), len(actual.Transactions))
		}
		if actual.GasLimit != tc.expectedBlock.GasLimit {
			t.Fatalf("expected gas limit %v, but got %v", tc.expectedBlock.GasLimit, actual.GasLimit)
		}
		if actual.GasUsed != tc.expectedBlock.GasUsed {
			t.Fatalf("expected gas used %v, but got %v", tc.expectedBlock.GasUsed, actual.GasUsed)
		}
	}
}

func TestCurrentBlock(t *testing.T) {
	mockGraphQL := map[string]map[string]interface{}{
		graphql.CurrentBlockQuery(): {"block": interface{}(map[string]interface{}{"number": "0x10"})},
	}
	bm := NewDefaultBlockMonitor(client.NewStubQuorumClient(mockGraphQL, nil), nil, "raft")
	currentBlockNumber, err := bm.currentBlockNumber()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if currentBlockNumber != 16 {
		t.Fatalf("expected %v, but got %v", 16, currentBlockNumber)
	}
}
