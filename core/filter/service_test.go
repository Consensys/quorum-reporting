package filter

import (
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

func TestIndexBlock(t *testing.T) {
	// setup
	mockRPC := map[string]interface{}{
		"debug_dumpAddress0x00000000000000000000000000000000000000010x4": &types.RawAccountState{},
		"debug_dumpAddress0x00000000000000000000000000000000000000010x5": &types.RawAccountState{},
		"debug_dumpAddress0x00000000000000000000000000000000000000010x6": &types.RawAccountState{},
		"debug_dumpAddress0x00000000000000000000000000000000000000020x6": &types.RawAccountState{},
	}
	db := &FakeDB{
		[]types.Address{types.NewAddress("1"), types.NewAddress("2")},
		map[types.Address]uint64{types.NewAddress("1"): 3, types.NewAddress("2"): 5},
	}
	fs := NewFilterService(db, client.NewStubQuorumClient(nil, mockRPC))

	// test fs.getLastFiltered
	lastFilteredAll, lastFiltered, err := fs.getLastFiltered(6)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, lastFiltered)
	assert.EqualValues(t, 3, lastFilteredAll[types.NewAddress("1")])
	assert.EqualValues(t, 5, lastFilteredAll[types.NewAddress("2")])

	// test fs.index
	err = fs.index(lastFilteredAll, 4, 4)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, db.lastFiltered[types.NewAddress("1")])
	assert.EqualValues(t, 5, db.lastFiltered[types.NewAddress("2")])

	// index multiple blocks
	err = fs.index(lastFilteredAll, 5, 6)
	assert.Nil(t, err)
	assert.EqualValues(t, 6, db.lastFiltered[types.NewAddress("1")])
	assert.EqualValues(t, 6, db.lastFiltered[types.NewAddress("2")])
}

type FakeDB struct {
	addresses    []types.Address
	lastFiltered map[types.Address]uint64
}

func (f *FakeDB) GetAddresses() ([]types.Address, error) {
	return f.addresses, nil
}

func (f *FakeDB) ReadBlock(blockNumber uint64) (*types.Block, error) {
	return &types.Block{Number: blockNumber}, nil
}

func (f *FakeDB) GetLastPersistedBlockNumber() (uint64, error) {
	return 0, errors.New("not implemented")
}

func (f *FakeDB) IndexStorage(map[types.Address]*types.AccountState, uint64) error {
	return nil
}

func (f *FakeDB) IndexBlock(addresses []types.Address, block *types.Block) error {
	for _, address := range addresses {
		if f.lastFiltered[address] < block.Number {
			f.lastFiltered[address] = block.Number
		}
	}
	return nil
}

func (f *FakeDB) IndexBlocks(addresses []types.Address, blocks []*types.Block) error {
	for _, block := range blocks {
		f.IndexBlock(addresses, block)
	}
	return nil
}

func (f *FakeDB) GetLastFiltered(address types.Address) (uint64, error) {
	return f.lastFiltered[address], nil
}

func (f *FakeDB) ReadTransaction(txHash types.Hash) (*types.Transaction, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeDB) RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
	return errors.New("not implemented")
}

func (f *FakeDB) RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error {
	return errors.New("not implemented")
}

func (f *FakeDB) GetContractABI(types.Address) (string, error) {
	return "{}", nil
}

func (f *FakeDB) SetContractCreationTransaction(creationTxns map[types.Hash][]types.Address) error {
	return nil
}
