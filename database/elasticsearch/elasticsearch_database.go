package elasticsearch

import (
	"context"
	"errors"
	"quorumengineering/quorum-report/types"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Database struct {
	client *elasticsearch7.Client
}

func New() *Database {
	return &Database{}
}

//AddressDB
func (es *Database) AddAddresses(addresses []common.Address) error {
	toSave := make([]Contract, len(addresses))

	for i, address := range addresses {
		toSave[i] = Contract{
			Address:             address,
			ABI:                 "",
			CreationTransaction: common.Hash{},
			LastFiltered:        0,
		}
	}

	es.client.Bulk(
		es.client.Bulk.WithContext(context.Background()),
	)

	//client.Search.WithContext(context.Background()),
	//client.Search.WithIndex("bank"),
	//client.Search.WithBody(&buf),
	//client.Search.WithTrackTotalHits(true),
	//client.Search.WithPretty(),

	return errors.New("not implemented")
}

func (es *Database) DeleteAddress(common.Address) error {
	return errors.New("not implemented")
}

func (es *Database) GetAddresses() ([]common.Address, error) {
	return nil, errors.New("not implemented")
}

//ABIDB
func (es *Database) AddContractABI(common.Address, *abi.ABI) error {
	return errors.New("not implemented")
}

func (es *Database) GetContractABI(common.Address) (*abi.ABI, error) {
	return nil, errors.New("not implemented")
}

// BlockDB
func (es *Database) WriteBlock(*types.Block) error {
	return errors.New("not implemented")
}

func (es *Database) ReadBlock(uint64) (*types.Block, error) {
	return nil, errors.New("not implemented")
}

func (es *Database) GetLastPersistedBlockNumber() (uint64, error) {
	return 0, errors.New("not implemented")
}

// TransactionDB
func (es *Database) WriteTransaction(*types.Transaction) error {
	return errors.New("not implemented")
}

func (es *Database) ReadTransaction(common.Hash) (*types.Transaction, error) {
	return nil, errors.New("not implemented")
}

// IndexDB
func (es *Database) IndexBlock([]common.Address, *types.Block) error {
	return errors.New("not implemented")
}

func (es *Database) GetContractCreationTransaction(common.Address) (common.Hash, error) {
	return common.Hash{}, errors.New("not implemented")
}

func (es *Database) GetAllTransactionsToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (es *Database) GetAllTransactionsInternalToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented")
}

func (es *Database) GetAllEventsByAddress(common.Address) ([]*types.Event, error) {
	return nil, errors.New("not implemented")
}

func (es *Database) GetStorage(common.Address, uint64) (map[common.Hash]string, error) {
	return nil, errors.New("not implemented")
}

func (es *Database) GetLastFiltered(common.Address) (uint64, error) {
	return 0, errors.New("not implemented")
}
