package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/types"
)

var testIndexBlock = &types.BlockWithTransactions{
	Hash:   types.NewHash("0x4b603921305ebaa48d863b9f577059a63c653cd8e952372622923708fb657806"),
	Number: 10,
	Transactions: []*types.Transaction{
		{
			Hash:            types.NewHash("0x86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb"),
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
			Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			BlockNumber: 10,
			Events: []*types.Event{
				{
					Topics: []types.Hash{ContractExtensionTopic},
					Data:   types.NewHexData("0x0000000000000000000000001932c48b2bf8102ba33b4a6b545c32236e342f34"),
				},
				{
					Topics: []types.Hash{ContractExtensionTopic},
					Data:   types.NewHexData("0x0000000000000000000000008a5e2a6343108babed07899510fb42297938d41f"),
				},
			},
		},
	},
}

func TestContractCreationFilter_ProcessBlocks_UnableToReadTransaction(t *testing.T) {
	db := memory.NewMemoryDB()
	ccFilter := NewContractCreationFilter(db, client.NewStubQuorumClient(nil, nil))
	sampleAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	err := ccFilter.ProcessBlocks([]types.Address{sampleAddress}, []*types.BlockWithTransactions{testIndexBlock})

	assert.EqualError(t, err, "transaction does not exist")
}

func TestContractCreationFilter_ProcessBlocks(t *testing.T) {
	testAddresses := []types.Address{
		"1349f3e1b8d71effb47b840594ff27da7e603d17", "9d13c6d3afe1721beef56b55d303b09e021e27ab",
		"1234567890123456789012345678901234567890", "123456789fe1721beef56b55d303b09e021e27ab",
		"1932c48b2bf8102ba33b4a6b545c32236e342f34", "8a5e2a6343108babed07899510fb42297938d41f",
	}
	db := memory.NewMemoryDB()
	_ = db.AddAddresses(testAddresses)
	ccFilter := NewContractCreationFilter(db, client.NewStubQuorumClient(nil, map[string]interface{}{
		"eth_getCode0x1932c48b2bf8102ba33b4a6b545c32236e342f340x9": types.NewHexData("0x"),
		"eth_getCode0x8a5e2a6343108babed07899510fb42297938d41f0x9": types.NewHexData("0x1234"),
	}))

	err := ccFilter.ProcessBlocks(testAddresses, []*types.BlockWithTransactions{testIndexBlock})
	assert.Nil(t, err)

	testCases := []struct {
		address types.Address
		hash    types.Hash
	}{
		{"1349f3e1b8d71effb47b840594ff27da7e603d17", "86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb"},
		{"9d13c6d3afe1721beef56b55d303b09e021e27ab", "693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"},
		{"1234567890123456789012345678901234567890", "693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"},
		{"123456789fe1721beef56b55d303b09e021e27ab", "693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"},
		{"1932c48b2bf8102ba33b4a6b545c32236e342f34", "f4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"},
		{"8a5e2a6343108babed07899510fb42297938d41f", ""},
	}

	for _, testCase := range testCases {
		creationTx, _ := db.GetContractCreationTransaction(testCase.address)
		assert.Equal(t, testCase.hash, creationTx)
	}
}

func TestContractCreationFilter_ProcessBlocks_NonIndexedContracts(t *testing.T) {
	testAddresses := []types.Address{
		"1349f3e1b8d71effb47b840594ff27da7e603d17", "9d13c6d3afe1721beef56b55d303b09e021e27ab",
		"1234567890123456789012345678901234567890", "123456789fe1721beef56b55d303b09e021e27ab",
		"1932c48b2bf8102ba33b4a6b545c32236e342f34", "8a5e2a6343108babed07899510fb42297938d41f",
	}
	db := memory.NewMemoryDB()
	_ = db.AddAddresses(testAddresses)
	ccFilter := NewContractCreationFilter(db, client.NewStubQuorumClient(nil, map[string]interface{}{
		"eth_getCode0x1932c48b2bf8102ba33b4a6b545c32236e342f340x9": types.NewHexData("0x"),
		"eth_getCode0x8a5e2a6343108babed07899510fb42297938d41f0x9": types.NewHexData("0x1234"),
	}))

	err := ccFilter.ProcessBlocks([]types.Address{}, []*types.BlockWithTransactions{testIndexBlock})
	assert.Nil(t, err)

	for _, address := range testAddresses {
		creationTx, _ := db.GetContractCreationTransaction(address)
		assert.Equal(t, types.Hash(""), creationTx)
	}
}
