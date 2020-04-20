package filter

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

func TestIndexBlock(t *testing.T) {
	// setup
	db := &FakeDB{[]common.Address{{1}, {2}}, map[common.Address]uint64{{1}: 3, {2}: 5}}
	fs := NewFilterService(db)
	lastFiltered, err := fs.getLastFiltered(6)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if lastFiltered != 3 {
		t.Fatalf("expected last filtered %v, but got %v", 3, lastFiltered)
	}
	err = fs.index(4)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if db.lastFiltered[common.Address{1}] != 4 {
		t.Fatalf("expected common.Address{1} last filtered %v, but got %v", 4, db.lastFiltered[common.Address{1}])
	}
	if db.lastFiltered[common.Address{2}] != 5 {
		t.Fatalf("expected common.Address{2} last filtered %v, but got %v", 5, db.lastFiltered[common.Address{2}])
	}
	err = fs.index(6)
	if db.lastFiltered[common.Address{1}] != 6 {
		t.Fatalf("expected common.Address{1} last filtered %v, but got %v", 6, db.lastFiltered[common.Address{1}])
	}
	if db.lastFiltered[common.Address{2}] != 6 {
		t.Fatalf("expected common.Address{2} last filtered %v, but got %v", 6, db.lastFiltered[common.Address{2}])
	}
}

type FakeDB struct {
	addresses    []common.Address
	lastFiltered map[common.Address]uint64
}

func (f *FakeDB) AddAddresses([]common.Address) error {
	return errors.New("not implemented")
}

func (f *FakeDB) DeleteAddress(common.Address) error {
	return errors.New("not implemented")
}

func (f *FakeDB) GetAddresses() ([]common.Address, error) {
	return f.addresses, nil
}

func (f *FakeDB) AddContractABI(common.Address, string) error {
	return errors.New("not implemented")
}

func (f *FakeDB) GetContractABI(common.Address) (string, error) {
	return "", errors.New("not implemented")
}

func (f *FakeDB) WriteBlock(*types.Block) error {
	return errors.New("not implemented")
}

func (f *FakeDB) ReadBlock(blockNumber uint64) (*types.Block, error) {
	return &types.Block{Number: blockNumber}, nil
}

func (f *FakeDB) GetLastPersistedBlockNumber() (uint64, error) {
	return 0, errors.New("not implemented")
}

func (f *FakeDB) WriteTransaction(*types.Transaction) error {
	return errors.New("not implemented")
}

func (f *FakeDB) ReadTransaction(common.Hash) (*types.Transaction, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeDB) IndexBlock(addresses []common.Address, block *types.Block) error {
	for _, address := range addresses {
		if f.lastFiltered[address] < block.Number {
			f.lastFiltered[address] = block.Number
		}
	}
	return nil
}

func (f *FakeDB) GetContractCreationTransaction(common.Address) (common.Hash, error) {
	return common.Hash{}, errors.New("not implemented")
}

func (f *FakeDB) GetAllTransactionsToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeDB) GetAllTransactionsInternalToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeDB) GetAllEventsByAddress(common.Address) ([]*types.Event, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeDB) GetStorage(common.Address, uint64) (map[common.Hash]string, error) {
	return nil, errors.New("not implemented")
}

func (f *FakeDB) GetLastFiltered(address common.Address) (uint64, error) {
	return f.lastFiltered[address], nil
}
