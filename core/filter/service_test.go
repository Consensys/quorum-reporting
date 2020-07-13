package filter

import (
	"errors"
	"testing"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

func TestIndexBlock(t *testing.T) {
	// setup
	mockRPC := map[string]interface{}{
		"debug_dumpAddress00000000000000000000000000000000000000010x4": &types.AccountState{},
		"debug_dumpAddress00000000000000000000000000000000000000010x5": &types.AccountState{},
		"debug_dumpAddress00000000000000000000000000000000000000010x6": &types.AccountState{},
		"debug_dumpAddress00000000000000000000000000000000000000020x6": &types.AccountState{},
	}
	db := &FakeDB{
		[]types.Address{types.NewAddress("1"), types.NewAddress("2")},
		map[types.Address]uint64{types.NewAddress("1"): 3, types.NewAddress("2"): 5},
	}
	fs := NewFilterService(db, client.NewStubQuorumClient(nil, mockRPC))
	// test fs.getLastFiltered
	lastFilteredAll, lastFiltered, err := fs.getLastFiltered(6)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if lastFilteredAll[types.NewAddress("1")] != 3 {
		t.Fatalf("expected last filtered of %v is %v, but got %v", types.NewAddress("1"), 3, lastFiltered)
	}
	if lastFilteredAll[types.NewAddress("2")] != 5 {
		t.Fatalf("expected last filtered of %v is %v, but got %v", types.NewAddress("2"), 5, lastFiltered)
	}
	if lastFiltered != 3 {
		t.Fatalf("expected last filtered %v, but got %v", 3, lastFiltered)
	}
	// test fs.index
	err = fs.index(lastFilteredAll, 4, 4)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if db.lastFiltered[types.NewAddress("1")] != 4 {
		t.Fatalf(`expected types.NewAddress("1") last filtered %v, but got %v`, 4, db.lastFiltered[types.NewAddress("1")])
	}
	if db.lastFiltered[types.NewAddress("2")] != 5 {
		t.Fatalf(`expected types.NewAddress("2") last filtered %v, but got %v`, 5, db.lastFiltered[types.NewAddress("2")])
	}
	// index multiple blocks
	err = fs.index(lastFilteredAll, 5, 6)
	if db.lastFiltered[types.NewAddress("1")] != 6 {
		t.Fatalf(`expected types.NewAddress("1") last filtered %v, but got %v`, 6, db.lastFiltered[types.NewAddress("1")])
	}
	if db.lastFiltered[types.NewAddress("2")] != 6 {
		t.Fatalf(`expected types.NewAddress("2") last filtered %v, but got %v`, 6, db.lastFiltered[types.NewAddress("2")])
	}
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
