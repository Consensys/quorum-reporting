package rpc

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type RPCAPIs struct {
	db database.Database
}

func NewRPCAPIs(db database.Database) *RPCAPIs {
	return &RPCAPIs{
		db,
	}
}

func (r *RPCAPIs) GetLastPersistedBlockNumber() uint64 {
	return r.db.GetLastPersistedBlockNumber()
}

func (r *RPCAPIs) GetBlock(blockNumber uint64) (*types.Block, error) {
	return r.db.ReadBlock(blockNumber)
}

func (r *RPCAPIs) GetTransaction(hash common.Hash) (*types.Transaction, error) {
	tx, err := r.db.ReadTransaction(hash)
	if err != nil {
		return nil, err
	}
	address := tx.To
	if address == (common.Address{0}) {
		address = tx.CreatedContract
	}
	contractABI := r.db.GetContractABI(address)
	if contractABI != nil {
		tx.ParseTransaction(contractABI)
	}
	return tx, nil
}

func (r *RPCAPIs) GetAllTransactionsByAddress(address common.Address) ([]common.Hash, error) {
	return r.db.GetAllTransactionsByAddress(address)
}

func (r *RPCAPIs) GetAllEventsByAddress(address common.Address) ([]*types.Event, error) {
	events, err := r.db.GetAllEventsByAddress(address)
	if err != nil {
		return nil, err
	}
	contractABI := r.db.GetContractABI(address)
	if contractABI != nil {
		for _, e := range events {
			e.ParseEvent(contractABI)
		}
	}
	return events, nil
}

func (r *RPCAPIs) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	return r.db.GetStorage(address, blockNumber)
}

func (r *RPCAPIs) AddAddress(address common.Address) error {
	if address == (common.Address{0}) {
		return errors.New("invalid input")
	}
	return r.db.AddAddresses([]common.Address{address})
}

func (r *RPCAPIs) DeleteAddress(address common.Address) error {
	return r.db.DeleteAddress(address)
}

func (r *RPCAPIs) GetAddresses() []common.Address {
	return r.db.GetAddresses()
}

func (r *RPCAPIs) AddContractABI(address common.Address, data string) error {
	contractABI, err := abi.JSON(strings.NewReader(data))
	if err != nil {
		return err
	}
	return r.db.AddContractABI(address, &contractABI)
}

func (r *RPCAPIs) GetContractABI(address common.Address) *abi.ABI {
	return r.db.GetContractABI(address)
}
