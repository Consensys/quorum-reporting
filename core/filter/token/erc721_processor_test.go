package token

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/types"
)

func TestERC721Processor_ProcessTransaction_NoEventsDoesNothing(t *testing.T) {
	tokenAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash:   types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				Events: []*types.Event{},
			},
		},
	}

	db := NewFakeTestTokenDatabase(nil)
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{tokenAddress: erc721AbiString}, testBlock)

	assert.Nil(t, err)
	assert.Len(t, db.RecordedContract, 0)
}

func TestERC721Processor_ProcessTransaction_NoErc721Events(t *testing.T) {
	tokenAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash: types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				Events: []*types.Event{
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"),
						Topics:  []types.Hash{types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36")},
					},
				},
			},
		},
	}

	db := NewFakeTestTokenDatabase(nil)
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{tokenAddress: erc721AbiString}, testBlock)

	assert.Nil(t, err)
	assert.Len(t, db.RecordedContract, 0)
}

func TestERC721Processor_ProcessTransaction_Erc721EventForNonTrackedAddress(t *testing.T) {
	tokenAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash: types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				Events: []*types.Event{
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17"),
						Topics: []types.Hash{
							"ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"000000000000000000000000ed9d02e382b34818e88b88a309c7fe71e65f419d",
							"0000000000000000000000001349f3e1b8d71effb47b840594ff27da7e603d17",
							"0000000000000000000000000000000000000000000000000000000000000001",
						},
					},
				},
			},
		},
	}

	db := NewFakeTestTokenDatabase(nil)
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{tokenAddress: erc721AbiString}, testBlock)

	assert.Nil(t, err)
	assert.Len(t, db.RecordedContract, 0)
}

func TestERC721Processor_ProcessTransaction_SingleErc721Event(t *testing.T) {
	tokenAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				BlockNumber: 1,
				Events: []*types.Event{
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"),
						Topics: []types.Hash{
							"ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"000000000000000000000000ed9d02e382b34818e88b88a309c7fe71e65f419d",
							"0000000000000000000000001349f3e1b8d71effb47b840594ff27da7e603d17",
							"0000000000000000000000000000000000000000000000000000000000000001",
						},
					},
				},
			},
		},
	}

	db := NewFakeTestTokenDatabase(nil)
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{tokenAddress: erc721AbiString}, testBlock)

	assert.Nil(t, err)
	assert.Contains(t, db.RecordedContract, types.NewAddress("1932c48b2bf8102ba33b4a6b545c32236e342f34"))
	assert.Len(t, db.RecordedHolder, 1)
	assert.Contains(t, db.RecordedHolder, types.NewAddress("1349f3e1b8d71effb47b840594ff27da7e603d17"))
	assert.EqualValues(t, 1, db.RecordedBlock)
	assert.Len(t, db.RecordedToken, 1)
	assert.EqualValues(t, db.RecordedToken[0], big.NewInt(1))
}

func TestERC721Processor_ProcessTransaction_SingleErc721EventForNonErc721Contract(t *testing.T) {
	tokenAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				BlockNumber: 1,
				Events: []*types.Event{
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"),
						Topics: []types.Hash{
							"ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"000000000000000000000000ed9d02e382b34818e88b88a309c7fe71e65f419d",
							"0000000000000000000000001349f3e1b8d71effb47b840594ff27da7e603d17",
							"0000000000000000000000000000000000000000000000000000000000000001",
						},
					},
				},
			},
		},
	}

	db := NewFakeTestTokenDatabase(nil)
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{tokenAddress: erc20AbiString}, testBlock)

	assert.Nil(t, err)
	assert.Len(t, db.RecordedContract, 0)
	assert.Len(t, db.RecordedHolder, 0)
	assert.EqualValues(t, 0, db.RecordedBlock)
	assert.Len(t, db.RecordedToken, 0)
}

func TestERC721Processor_ProcessTransaction_SingleErc721Event_WithDatabaseError(t *testing.T) {
	tokenAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				BlockNumber: 1,
				Events: []*types.Event{
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"),
						Topics: []types.Hash{
							"ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"000000000000000000000000ed9d02e382b34818e88b88a309c7fe71e65f419d",
							"0000000000000000000000001349f3e1b8d71effb47b840594ff27da7e603d17",
							"0000000000000000000000000000000000000000000000000000000000000001",
						},
					},
				},
			},
		},
	}

	db := NewFakeTestTokenDatabase(errors.New("test error - database"))
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{tokenAddress: erc721AbiString}, testBlock)

	assert.EqualError(t, err, "test error - database")
	assert.Len(t, db.RecordedContract, 0)
}

func TestERC721Processor_ProcessTransaction_MultipleErc721Events(t *testing.T) {
	testBlock := &types.BlockWithTransactions{
		Number: 1,
		Hash:   types.NewHash("0xe625ba9f14eed0671508966080fb01374d0a3a16b9cee545a324179b75f30aa8"),
		Transactions: []*types.Transaction{
			{
				Hash: types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
				Events: []*types.Event{
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"),
						Topics: []types.Hash{
							"ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"000000000000000000000000ed9d02e382b34818e88b88a309c7fe71e65f419d",
							"0000000000000000000000001349f3e1b8d71effb47b840594ff27da7e603d17",
							"0000000000000000000000000000000000000000000000000000000000000001",
						},
					},
					{
						Data:    types.NewHexData("0x00000000000000000000000000000000000000000000000000000000000003e8"),
						Address: types.NewAddress("0x02826f2bce5596f49ef29f11de3dce29d6653f8c"),
						Topics: []types.Hash{
							"ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
							"0000000000000000000000009d13c6d3afe1721beef56b55d303b09e021e27ab",
							"000000000000000000000000e625ba9f14eed0671508966080fb01374d0a3a18",
							"0000000000000000000000000000000000000000000000000000000000000002",
						},
					},
				},
			},
		},
	}

	db := NewFakeTestTokenDatabase(nil)
	processor := NewERC721Processor(db)

	err := processor.ProcessBlock(map[types.Address]string{
		types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"): erc721AbiString,
		types.NewAddress("0x02826f2bce5596f49ef29f11de3dce29d6653f8c"): erc721AbiString,
	}, testBlock)

	assert.Nil(t, err)
	assert.Contains(t, db.RecordedContract, types.NewAddress("1932c48b2bf8102ba33b4a6b545c32236e342f34"))
	assert.Contains(t, db.RecordedContract, types.NewAddress("02826f2bce5596f49ef29f11de3dce29d6653f8c"))
	assert.Len(t, db.RecordedHolder, 2)
	assert.Contains(t, db.RecordedHolder, types.NewAddress("1349f3e1b8d71effb47b840594ff27da7e603d17"))
	assert.Contains(t, db.RecordedHolder, types.NewAddress("e625ba9f14eed0671508966080fb01374d0a3a18"))
	assert.EqualValues(t, 1, db.RecordedBlock)
	assert.Len(t, db.RecordedToken, 2)
	assert.EqualValues(t, big.NewInt(1), db.RecordedToken[0])
	assert.EqualValues(t, big.NewInt(2), db.RecordedToken[1])
}
