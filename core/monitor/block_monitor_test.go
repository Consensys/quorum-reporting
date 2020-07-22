package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/client"
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
				Number:    42,
				Timestamp: 1000000000,
				GasLimit:  799024091,
				GasUsed:   27530,
			},
			&types.Block{
				Number:       42,
				Timestamp:    1_000_000_000,
				Transactions: []types.Hash{},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"istanbul",
		},
		{
			&types.RawBlock{
				Number:       42,
				Timestamp:    1000000000,
				Transactions: []types.Hash{types.NewHash("")},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			&types.Block{
				Number:       42,
				Timestamp:    1_000_000_000,
				Transactions: []types.Hash{types.NewHash("")},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"istanbul",
		},
		{
			&types.RawBlock{
				Number:    42,
				Timestamp: 1000000000,
				GasLimit:  799024091,
				GasUsed:   27530,
			},
			&types.Block{
				Number:       42,
				Timestamp:    1,
				Transactions: []types.Hash{},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"raft",
		},
		{
			&types.RawBlock{
				Number:       42,
				Timestamp:    1000000000,
				Transactions: []types.Hash{types.NewHash("")},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			&types.Block{
				Number:       42,
				Timestamp:    1,
				Transactions: []types.Hash{types.NewHash("")},
				GasLimit:     799024091,
				GasUsed:      27530,
			},
			"raft",
		},
	}

	for _, tc := range cases {
		bm := NewDefaultBlockMonitor(client.NewStubQuorumClient(nil, nil), nil, tc.consensus)

		actual := bm.createBlock(tc.originalBlock)

		assert.EqualValues(t, tc.expectedBlock.Number, actual.Number)
		assert.EqualValues(t, tc.expectedBlock.Timestamp, actual.Timestamp)
		assert.EqualValues(t, tc.expectedBlock.GasLimit, actual.GasLimit)
		assert.EqualValues(t, tc.expectedBlock.GasUsed, actual.GasUsed)
		assert.EqualValues(t, len(tc.expectedBlock.Transactions), len(actual.Transactions))
	}
}
