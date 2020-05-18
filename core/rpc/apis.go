package rpc

import (
	"encoding/json"
	"errors"
	"quorumengineering/quorum-report/core/storageparsing"
	types2 "quorumengineering/quorum-report/core/storageparsing/types"
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
		parsedTx.ParsedEvents = make([]*types.ParsedEvent, len(parsedTx.RawTransaction.Events))
		for i, e := range parsedTx.RawTransaction.Events {
			contractABI, err := r.db.GetContractABI(e.Address)
			if err != nil {
				return nil, err
			}
			if contractABI != "" {
				parsedTx.ParsedEvents[i] = &types.ParsedEvent{
					RawEvent: e,
				}
				if err := parsedTx.ParsedEvents[i].ParseEvent(contractABI); err != nil {
					return nil, err
				}
			}
		}
	}
	return parsedTx, nil
}

func (r *RPCAPIs) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	return r.db.GetContractCreationTransaction(address)
}

func (r *RPCAPIs) GetAllTransactionsToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	if options == nil {
		options = &types.QueryOptions{}
	}
	options.SetDefaults()

	return r.db.GetAllTransactionsToAddress(address, options)
}

func (r *RPCAPIs) GetAllTransactionsInternalToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	if options == nil {
		options = &types.QueryOptions{}
	}
	options.SetDefaults()

	return r.db.GetAllTransactionsInternalToAddress(address, options)
}

func (r *RPCAPIs) GetAllEventsFromAddress(address common.Address, options *types.QueryOptions) ([]*types.ParsedEvent, error) {
	if options == nil {
		options = &types.QueryOptions{}
	}
	options.SetDefaults()

	events, err := r.db.GetAllEventsFromAddress(address, options)
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

func (r *RPCAPIs) GetStorageHistory(address common.Address) (*types2.ReportingResponseTemplate, error) {
	rawAbi, err := r.db.GetStorageABI(address)
	if err != nil {
		return nil, err
	}
	if rawAbi == "" {
		return nil, errors.New("no storage ABI present to parse with")
	}
	var parsedAbi types2.SolidityStorageDocument
	err = json.Unmarshal([]byte(rawAbi), &parsedAbi)
	if err != nil {
		return nil, errors.New("unable to decode storage ABI: " + err.Error())
	}

	// TODO: implement GetStorageRoot to reduce the response list
	historicStates := []*types2.ParsedState{}
	for i := 1; i <= 1; i++ {
		rawStorage, err := r.db.GetStorage(address, uint64(i))
		if err != nil {
			return nil, err
		}
		if rawStorage == nil {
			continue
		}
		historicStorage, err := storageparsing.ParseRawStorage(rawStorage, parsedAbi)
		if err != nil {
			return nil, err
		}
		historicStates = append(historicStates, &types2.ParsedState{
			BlockNumber:     uint64(i),
			HistoricStorage: historicStorage,
		})
	}
	return &types2.ReportingResponseTemplate{
		Address:       address,
		HistoricState: historicStates,
	}, nil
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

func (r *RPCAPIs) AddABI(address common.Address, data string) error {
	//check ABI is valid
	_, err := abi.JSON(strings.NewReader(data))
	if err != nil {
		return err
	}
	return r.db.AddContractABI(address, data)
}

func (r *RPCAPIs) GetABI(address common.Address) (string, error) {
	return r.db.GetContractABI(address)
}

func (r *RPCAPIs) AddStorageABI(address common.Address, data string) error {
	var storageAbi types2.SolidityStorageDocument
	err := json.Unmarshal([]byte(data), &storageAbi)
	if err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}
	return r.db.AddStorageABI(address, data)
}

func (r *RPCAPIs) GetStorageABI(address common.Address) (string, error) {
	return r.db.GetStorageABI(address)
}
