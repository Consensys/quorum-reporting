package filter

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

func TestIndexBlock(t *testing.T) {
	// setup
	db := &FakeIndexDB{make(map[common.Address]bool)}
	tf := &BlockFilter{db}
	addresses := []common.Address{{1}, {2}}
	tf.IndexBlock(addresses, nil)
	if !db.Indexed[common.Address{1}] {
		t.Fatalf("expected %v indexed but not", common.Address{1})
	}
	if !db.Indexed[common.Address{2}] {
		t.Fatalf("expected %v indexed but not", common.Address{2})
	}
	if db.Indexed[common.Address{3}] {
		t.Fatalf("expected %v not indexed", common.Address{3})
	}
}

type FakeIndexDB struct {
	Indexed map[common.Address]bool
}

func (f *FakeIndexDB) IndexBlock(address common.Address, block *types.Block) error {
	f.Indexed[address] = true
	return nil
}

func (f *FakeIndexDB) GetAllTransactionsByAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetAllEventsByAddress(common.Address) ([]*types.Event, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetStorage(common.Address, uint64) (map[common.Hash]string, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetLastFiltered(common.Address) uint64 {
	return 0
}
