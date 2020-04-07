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
	bf := &BlockFilter{db}
	addresses := []common.Address{{1}, {2}}
	bf.IndexBlock(addresses, nil)
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

func (f *FakeIndexDB) GetContractCreationTransaction(common.Address) (common.Hash, error) {
	return common.Hash{}, errors.New("not implemented")
}

func (f *FakeIndexDB) GetAllTransactionsToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetAllTransactionsInternalToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetAllEventsByAddress(common.Address) ([]*types.Event, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetStorage(common.Address, uint64) (map[common.Hash]string, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeIndexDB) GetLastFiltered(common.Address) (uint64, error) {
	return 0, errors.New("not implemented")
}
