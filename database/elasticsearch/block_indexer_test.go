package elasticsearch

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/types"
)

var testIndexBlock = &types.BlockWithTransactions{
	Hash:   types.NewHash("0x4b603921305ebaa48d863b9f577059a63c653cd8e952372622923708fb657806"),
	Number: 10,
	Transactions: []*types.Transaction{
		{
			Hash:            types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			CreatedContract: types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"),
		},
		{
			Hash: types.NewHash("0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"),
			InternalCalls: []*types.InternalCall{
				{
					Type: "CREATE",
					To:   types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"),
				},
				{
					Type: "CREATE",
					To:   types.NewAddress("0x1234567890123456789012345678901234567890"),
				},
				{
					Type: "CREATE2",
					To:   types.NewAddress("0x123456789fe1721beef56b55d303b09e021e27ab"),
				},
			},
		},
		{
			Hash: types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			Events: []*types.Event{
				{
					Address: types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"),
				},
				{
					Address: types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"),
				},
				{
					Address: types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09123456789"),
				},
			},
		},
	},
}

func TestDefaultBlockIndexer_IndexTransaction_AllRelevantEventsIndexed(t *testing.T) {
	var indexedEvents []*types.Event

	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true},
		blocks:    []*types.BlockWithTransactions{testIndexBlock},
		createEvents: func(events []*types.Event) error {
			indexedEvents = events
			return nil
		},
	}

	err := blockIndexer.Index()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(indexedEvents))
}

func TestDefaultBlockIndexer_IndexTransaction_IndexEventsError(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true},
		blocks:    []*types.BlockWithTransactions{testIndexBlock},
		createEvents: func(events []*types.Event) error {
			return errors.New("test error: createEvents")
		},
	}

	err := blockIndexer.Index()
	assert.EqualError(t, err, "test error: createEvents")
}
