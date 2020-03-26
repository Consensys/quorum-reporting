package monitor

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/types"
)

func TestCreateBlock(t *testing.T) {
	cases := []struct {
		originalBlock *ethTypes.Block
		expectedBlock *types.Block
	}{
		{ethTypes.NewBlock(&ethTypes.Header{Number: big.NewInt(42)}, nil, nil, nil),
			&types.Block{
				Hash:         common.HexToHash("0xc8e9124049353943a45cc95f07bc7cfdffb27e8ea2eb44167393181903d7ef3a"),
				Number:       uint64(42),
				Transactions: []common.Hash{},
			},
		},
		{ethTypes.NewBlock(&ethTypes.Header{Number: big.NewInt(42)}, []*ethTypes.Transaction{
			ethTypes.NewTransaction(0, common.Address{0}, nil, 0, nil, nil),
		}, nil, nil),
			&types.Block{
				Hash:         common.HexToHash("0x6d7b7e0605ca6afef8b8f811ce922019d15eda90230e36d8d2391f5023d67f1f"),
				Number:       uint64(42),
				Transactions: []common.Hash{common.BigToHash(big.NewInt(0))},
			},
		},
	}

	for _, tc := range cases {
		actual := createBlock(tc.originalBlock)
		if actual.Hash != tc.expectedBlock.Hash {
			t.Fatalf("expected hash %v, but got %v", tc.expectedBlock.Hash.Hex(), actual.Hash.Hex())
		}
		if actual.Number != tc.expectedBlock.Number {
			t.Fatalf("expected block number %v, but got %v", tc.expectedBlock.Number, actual.Number)
		}
		if len(actual.Transactions) != len(tc.expectedBlock.Transactions) {
			t.Fatalf("expected %v transactions, but got %v", len(tc.expectedBlock.Transactions), len(actual.Transactions))
		}
	}
}

func TestCurrentBlock(t *testing.T) {
	// TODO: simulate quorum client
}

func TestSyncBlocks(t *testing.T) {
	// TODO: simulate quorum client
}
