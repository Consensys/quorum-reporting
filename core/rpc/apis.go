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

func (r *RPCAPIs) GetLastPersistedBlockNumber() (uint64, error) {
	return r.db.GetLastPersistedBlockNumber()
}

func (r *RPCAPIs) GetBlock(blockNumber uint64) (*types.Block, error) {
	return r.db.ReadBlock(blockNumber)
}

func (r *RPCAPIs) GetTransaction(hash common.Hash) (*types.ParsedTransaction, error) {
	tx, err := r.db.ReadTransaction(hash)
	if err != nil {
		return nil, err
	}
	address := tx.To
	if address == (common.Address{0}) {
		address = tx.CreatedContract
	}
	contractABI, err := r.db.GetContractABI(address)
	if err != nil {
		return nil, err
	}
	parsedTx := &types.ParsedTransaction{
		RawTransaction: tx,
	}
	if contractABI != "" {
		if err = parsedTx.ParseTransaction(contractABI); err != nil {
			return nil, err
		}
	}
	return parsedTx, nil
}

func (r *RPCAPIs) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	return r.db.GetContractCreationTransaction(address)
}

func (r *RPCAPIs) GetAllTransactionsToAddress(address common.Address) ([]common.Hash, error) {
	return r.db.GetAllTransactionsToAddress(address)
}

func (r *RPCAPIs) GetAllTransactionsInternalToAddress(address common.Address) ([]common.Hash, error) {
	return r.db.GetAllTransactionsInternalToAddress(address)
}

func (r *RPCAPIs) GetAllEventsByAddress(address common.Address) ([]*types.ParsedEvent, error) {
	events, err := r.db.GetAllEventsByAddress(address)
	if err != nil {
		return nil, err
	}
	contractABI, err := r.db.GetContractABI(address)
	if err != nil {
		return nil, err
	}
	parsedEvents := make([]*types.ParsedEvent, len(events))
	for i, e := range events {
		parsedEvents[i] = &types.ParsedEvent{
			RawEvent: e,
		}
		if contractABI != "" {
			if err = parsedEvents[i].ParseEvent(contractABI); err != nil {
				return nil, err
			}
		}
	}
	return parsedEvents, nil
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

func (r *RPCAPIs) GetAddresses() ([]common.Address, error) {
	return r.db.GetAddresses()
}

func (r *RPCAPIs) AddContractABI(address common.Address, data string) error {
	//check ABI is valid
	_, err := abi.JSON(strings.NewReader(data))
	if err != nil {
		return err
	}
	return r.db.AddContractABI(address, data)
}

func (r *RPCAPIs) GetContractABI(address common.Address) (string, error) {
	return r.db.GetContractABI(address)
}
