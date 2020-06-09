package elasticsearch

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"quorumengineering/quorum-report/types"
	"testing"
)

var (
	testIndexBlock = &types.Block{
		Hash:   common.HexToHash("0x4b603921305ebaa48d863b9f577059a63c653cd8e952372622923708fb657806"),
		Number: 10,
		Transactions: []common.Hash{
			common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			common.HexToHash("0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"),
			common.HexToHash("0x5c83fa5955aff33c61813105851777bcd2adc85deb9af6286ba42c05cd768de0"),
		},
	}

	indexTransactionMap = map[string]*types.Transaction{
		"0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59": {
			Hash:            common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			CreatedContract: common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"),
		},
		"0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a": {
			Hash: common.HexToHash("0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"),
			InternalCalls: []*types.InternalCall{
				{
					Type: "CREATE",
					To:   common.HexToAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"),
				},
				{
					Type: "CREATE",
					To:   common.HexToAddress("0x1234567890123456789012345678901234567890"),
				},
				{
					Type: "CREATE2",
					To:   common.HexToAddress("0x123456789fe1721beef56b55d303b09e021e27ab"),
				},
			},
		},
		"0x5c83fa5955aff33c61813105851777bcd2adc85deb9af6286ba42c05cd768de0": {
			Hash: common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			Events: []*types.Event{
				{
					Address: common.HexToAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"),
				},
				{
					Address: common.HexToAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"),
				},
				{
					Address: common.HexToAddress("0x9d13c6d3afe1721beef56b55d303b09123456789"),
				},
			},
		},
	}
)

func TestDefaultBlockIndexer_Index_UnableToReadTransaction(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: nil,
		blocks:    []*types.Block{testIndexBlock},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
			return nil, errors.New("test error: readTransaction")
		},
	}

	err := blockIndexer.Index()

	assert.EqualError(t, err, "test error: readTransaction")
}

func TestDefaultBlockIndexer_IndexTransaction_ContractCreated_WithError(t *testing.T) {
	blockIndexer := &DefaultBlockIndexer{
		addresses: map[common.Address]bool{common.HexToAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address common.Address, prop string, val string) error {
			return errors.New("test error: updateContract")
		},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
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
		addresses: map[common.Address]bool{},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address common.Address, prop string, val string) error {
			return errors.New("test error: updateContract")
		},
		createEvents: func(events []*types.Event) error {
			return nil
		},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
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
		addresses: map[common.Address]bool{common.HexToAddress("0x1349F3e1B8D71eFfb47B840594Ff27dA7E603d17"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address common.Address, prop string, val string) error {
			updates[address.String()] = struct {
				prop string
				val  string
			}{prop: prop, val: val}
			return nil
		},
		createEvents: func(events []*types.Event) error {
			return nil
		},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(updates))
	assert.Equal(t, "creationTx", updates["0x1349F3e1B8D71eFfb47B840594Ff27dA7E603d17"].prop)
	assert.Equal(t, "0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59", updates["0x1349F3e1B8D71eFfb47B840594Ff27dA7E603d17"].val)
}

func TestDefaultBlockIndexer_IndexTransaction_CreateInternalTxUpdatesContract(t *testing.T) {
	updates := map[string]struct {
		prop string
		val  string
	}{}

	blockIndexer := &DefaultBlockIndexer{
		addresses: map[common.Address]bool{
			common.HexToAddress("0x9d13C6D3aFE1721BEef56B55D303B09E021E27ab"): true,
			common.HexToAddress("0x123456789FE1721bEEF56B55D303B09e021E27ab"): true,
		},
		blocks: []*types.Block{testIndexBlock},
		updateContract: func(address common.Address, prop string, val string) error {
			updates[address.String()] = struct {
				prop string
				val  string
			}{prop: prop, val: val}
			return nil
		},
		createEvents: func(events []*types.Event) error {
			return nil
		},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(updates))
	assert.Equal(t, "creationTx", updates["0x9d13C6D3aFE1721BEef56B55D303B09E021E27ab"].prop)
	assert.Equal(t, "0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a", updates["0x9d13C6D3aFE1721BEef56B55D303B09E021E27ab"].val)
	assert.Equal(t, "creationTx", updates["0x123456789FE1721bEEF56B55D303B09e021E27ab"].prop)
	assert.Equal(t, "0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a", updates["0x123456789FE1721bEEF56B55D303B09e021E27ab"].val)
}

func TestDefaultBlockIndexer_IndexTransaction_AllRelevantEventsIndexed(t *testing.T) {
	var indexedEvents []*types.Event

	blockIndexer := &DefaultBlockIndexer{
		addresses: map[common.Address]bool{common.HexToAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address common.Address, prop string, val string) error {
			return nil
		},
		createEvents: func(events []*types.Event) error {
			indexedEvents = events
			return nil
		},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
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
		addresses: map[common.Address]bool{common.HexToAddress("0x9d13c6d3afe1721beef56b55d303b09e021e27ab"): true},
		blocks:    []*types.Block{testIndexBlock},
		updateContract: func(address common.Address, prop string, val string) error {
			return nil
		},
		createEvents: func(events []*types.Event) error {
			return errors.New("test error: createEvents")
		},
		readTransaction: func(hash common.Hash) (*types.Transaction, error) {
			if tx, ok := indexTransactionMap[hash.String()]; ok {
				return tx, nil
			}
			return nil, errors.New("test error: not found")
		},
	}

	err := blockIndexer.Index()

	assert.EqualError(t, err, "test error: createEvents")
}
