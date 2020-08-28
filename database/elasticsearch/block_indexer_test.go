package elasticsearch

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/consensys/quorum-go-utils/types"
)

var (
	testIndexBlock = &types.Block{
		Hash:   types.NewHash("0x4b603921305ebaa48d863b9f577059a63c653cd8e952372622923708fb657806"),
		Number: 10,
		Transactions: []types.Hash{
			types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			types.NewHash("0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"),
			types.NewHash("0x5c83fa5955aff33c61813105851777bcd2adc85deb9af6286ba42c05cd768de0"),
		},
	}

	indexTransactionMap = map[string]*types.Transaction{
		"0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59": {
			Hash:            types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			CreatedContract: types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"),
		},
		"0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a": {
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
		"0x5c83fa5955aff33c61813105851777bcd2adc85deb9af6286ba42c05cd768de0": {
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
	}
)

func TestDefaultBlockIndexer_Index_UnableToReadTransaction(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: nil,
		blocks:    []*types.Block{testIndexBlock},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			return nil, errors.New("test error: readTransaction")
		},
	}

	err := blockIndexer.Index()

	assert.EqualError(t, err, "test error: readTransaction")
}

func TestDefaultBlockIndexer_IndexTransaction_ContractCreated_WithError(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address types.Address, prop string, val string) error {
			return errors.New("test error: updateContract")
		},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.EqualError(t, err, "test error: updateContract")
}

func TestDefaultBlockIndexer_IndexTransaction_ContractCreatedNotCalledForUnregisteredContract(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address types.Address, prop string, val string) error {
			return errors.New("test error: updateContract")
		},
		createEvents: func(events []*types.Event) error {
			return nil
		},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.Nil(t, err)
}

func TestDefaultBlockIndexer_IndexTransaction_ContractCreatedUpdatesContract(t *testing.T) {
	updates := map[string]struct {
		prop string
		val  string
	}{}

	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address types.Address, prop string, val string) error {
			updates[address.String()] = struct {
				prop string
				val  string
			}{prop: prop, val: val}
			return nil
		},
		createEvents: func(events []*types.Event) error {
			return nil
		},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(updates))
	assert.Equal(t, "creationTx", updates["0x1349f3e1b8d71effb47b840594ff27da7e603d17"].prop)
	assert.Equal(t, "0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59", updates["0x1349f3e1b8d71effb47b840594ff27da7e603d17"].val)
}

func TestDefaultBlockIndexer_IndexTransaction_CreateInternalTxUpdatesContract(t *testing.T) {
	updates := map[string]struct {
		prop string
		val  string
	}{}

	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{
			types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true,
			types.NewAddress("0x123456789fe1721beef56b55d303b09e021e27ab"): true,
		},
		blocks: []*types.Block{testIndexBlock},
		updateContract: func(address types.Address, prop string, val string) error {
			updates[address.String()] = struct {
				prop string
				val  string
			}{prop: prop, val: val}
			return nil
		},
		createEvents: func(events []*types.Event) error {
			return nil
		},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(updates))
	assert.Equal(t, "creationTx", updates["0x9d13c6d3afe1721beef56b55d303b09e021e27ab"].prop)
	assert.Equal(t, "0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a", updates["0x9d13c6d3afe1721beef56b55d303b09e021e27ab"].val)
	assert.Equal(t, "creationTx", updates["0x123456789fe1721beef56b55d303b09e021e27ab"].prop)
	assert.Equal(t, "0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a", updates["0x123456789fe1721beef56b55d303b09e021e27ab"].val)
}

func TestDefaultBlockIndexer_IndexTransaction_AllRelevantEventsIndexed(t *testing.T) {
	var indexedEvents []*types.Event

	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address types.Address, prop string, val string) error {
			return nil
		},
		createEvents: func(events []*types.Event) error {
			indexedEvents = events
			return nil
		},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(indexedEvents))
}

func TestDefaultBlockIndexer_IndexTransaction_IndexEventsError(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: map[types.Address]bool{types.NewAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address types.Address, prop string, val string) error {
			return nil
		},
		createEvents: func(events []*types.Event) error {
			return errors.New("test error: createEvents")
		},
		readTransaction: func(hash types.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.EqualError(t, err, "test error: createEvents")
}
