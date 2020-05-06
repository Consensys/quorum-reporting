package monitor

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/types"
)

func TestCreateBlock(t *testing.T) {
	cases := []struct {
		originalBlock *ethTypes.Block
		expectedBlock *types.Block
		consensus     string
	}{
		{ethTypes.NewBlock(&ethTypes.Header{Number: big.NewInt(42), Time: 1_000_000_000}, nil, nil, nil),
			&types.Block{
				Hash:         common.HexToHash("0x1e492b9b3fceea83d5a94abf3486f2ee03b609ab7a8500af28d90f02ddbce7b9"),
				Number:       uint64(42),
				Transactions: []common.Hash{},
				Timestamp:    1_000_000_000,
			},
			"istanbul",
		},
		{ethTypes.NewBlock(&ethTypes.Header{Number: big.NewInt(42), Time: 1_000_000_000}, []*ethTypes.Transaction{
			ethTypes.NewTransaction(0, common.Address{0}, nil, 0, nil, nil),
		}, nil, nil),
			&types.Block{
				Hash:         common.HexToHash("0x58e58dae7e4bbcb4459b0ee01c2e87d2840b12bfdd3f26dd2f9b3b5b0f4f23cd"),
				Number:       uint64(42),
				Transactions: []common.Hash{common.BigToHash(big.NewInt(0))},
				Timestamp:    1_000_000_000,
			},
			"istanbul",
		},
		{ethTypes.NewBlock(&ethTypes.Header{Number: big.NewInt(42), Time: 1_000_000_000}, nil, nil, nil),
			&types.Block{
				Hash:         common.HexToHash("0x1e492b9b3fceea83d5a94abf3486f2ee03b609ab7a8500af28d90f02ddbce7b9"),
				Number:       uint64(42),
				Transactions: []common.Hash{},
				Timestamp:    1,
			},
			"raft",
		},
		{ethTypes.NewBlock(&ethTypes.Header{Number: big.NewInt(42), Time: 1_000_000_000}, []*ethTypes.Transaction{
			ethTypes.NewTransaction(0, common.Address{0}, nil, 0, nil, nil),
		}, nil, nil),
			&types.Block{
				Hash:         common.HexToHash("0x58e58dae7e4bbcb4459b0ee01c2e87d2840b12bfdd3f26dd2f9b3b5b0f4f23cd"),
				Number:       uint64(42),
				Transactions: []common.Hash{common.BigToHash(big.NewInt(0))},
				Timestamp:    1,
			},
			"raft",
		},
	}

	for _, tc := range cases {
		bm := NewBlockMonitor(nil, client.NewStubQuorumClient(nil, nil, nil), tc.consensus)
		actual := bm.createBlock(tc.originalBlock)
		if actual.Hash != tc.expectedBlock.Hash {
			t.Fatalf("expected hash %v, but got %v", tc.expectedBlock.Hash.Hex(), actual.Hash.Hex())
		}
		if actual.Number != tc.expectedBlock.Number {
			t.Fatalf("expected block number %v, but got %v", tc.expectedBlock.Number, actual.Number)
		}
		if len(actual.Transactions) != len(tc.expectedBlock.Transactions) {
			t.Fatalf("expected %v transactions, but got %v", len(tc.expectedBlock.Transactions), len(actual.Transactions))
		}
		if tc.consensus == "raft" && actual.Timestamp != tc.expectedBlock.Timestamp {
			t.Fatalf("expected timestamp %d for raft, but got %v", tc.expectedBlock.Timestamp, actual.Timestamp)
		} else if actual.Timestamp != tc.expectedBlock.Timestamp {
			t.Fatalf("expected timestamp %d for %s, but got %v", tc.expectedBlock.Timestamp, tc.consensus, actual.Timestamp)
		}
	}
}

func TestCurrentBlock(t *testing.T) {
	mockGraphQL := map[string]map[string]interface{}{
		graphql.CurrentBlockQuery(): {"block": interface{}(map[string]interface{}{"number": "0x10"})},
	}
	bm := NewBlockMonitor(nil, client.NewStubQuorumClient(nil, mockGraphQL, nil), "raft")
	currentBlockNumber, err := bm.currentBlockNumber()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if currentBlockNumber != 16 {
		t.Fatalf("expected %v, but got %v", 16, currentBlockNumber)
	}
}
