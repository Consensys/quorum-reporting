package rpc

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/core/storageparsing"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type RPCAPIs struct {
	db                      database.Database
	contractTemplateManager ContractTemplateManager
}

func NewRPCAPIs(db database.Database, contractTemplateManager ContractTemplateManager) *RPCAPIs {
	return &RPCAPIs{db, contractTemplateManager}
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
	parsedTx.ParsedEvents = make([]*types.ParsedEvent, len(parsedTx.RawTransaction.Events))
	for i, e := range parsedTx.RawTransaction.Events {
		parsedTx.ParsedEvents[i] = &types.ParsedEvent{
			RawEvent: e,
		}
		contractABI, err := r.db.GetContractABI(e.Address)
		if err != nil {
			return nil, err
		}
		if contractABI != "" {
			if err := parsedTx.ParsedEvents[i].ParseEvent(contractABI); err != nil {
				return nil, err
			}
		}
	}
	return parsedTx, nil
}

func (r *RPCAPIs) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	txHash, err := r.db.GetContractCreationTransaction(address)
	if err != nil {
		return common.Hash{0}, err
	}
	if txHash == (common.Hash{0}) {
		return common.Hash{0}, errors.New("contract creation tx not found")
	}
	return txHash, nil
}

func (r *RPCAPIs) GetAllTransactionsToAddress(address common.Address, options *types.QueryOptions) (*TransactionsResp, error) {
	if options == nil {
		options = &types.QueryOptions{}
	}
	options.SetDefaults()

	total, err := r.db.GetTransactionsToAddressTotal(address, options)
	if err != nil {
		return nil, err
	}
	txs, err := r.db.GetAllTransactionsToAddress(address, options)
	if err != nil {
		return nil, err
	}

	return &TransactionsResp{
		Transactions: txs,
		Total:        total,
		Options:      options,
	}, nil
}

func (r *RPCAPIs) GetAllTransactionsInternalToAddress(address common.Address, options *types.QueryOptions) (*TransactionsResp, error) {
	if options == nil {
		options = &types.QueryOptions{}
	}
	options.SetDefaults()

	total, err := r.db.GetTransactionsInternalToAddressTotal(address, options)
	if err != nil {
		return nil, err
	}
	txs, err := r.db.GetAllTransactionsInternalToAddress(address, options)
	if err != nil {
		return nil, err
	}

	return &TransactionsResp{
		Transactions: txs,
		Total:        total,
		Options:      options,
	}, nil
}

func (r *RPCAPIs) GetAllEventsFromAddress(address common.Address, options *types.QueryOptions) (*EventsResp, error) {
	if options == nil {
		options = &types.QueryOptions{}
	}
	options.SetDefaults()

	total, err := r.db.GetEventsFromAddressTotal(address, options)
	if err != nil {
		return nil, err
	}
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

	return &EventsResp{
		Events:  parsedEvents,
		Total:   total,
		Options: options,
	}, nil
}

func (r *RPCAPIs) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	return r.db.GetStorage(address, blockNumber)
}

func (r *RPCAPIs) GetStorageHistory(address common.Address, startBlockNumber, endBlockNumber uint64) (*types.ReportingResponseTemplate, error) {
	rawAbi, err := r.db.GetStorageLayout(address)
	if err != nil {
		return nil, err
	}
	if rawAbi == "" {
		return nil, errors.New("no Storage Layout present to parse with")
	}
	var parsedAbi types.SolidityStorageDocument
	if err = json.Unmarshal([]byte(rawAbi), &parsedAbi); err != nil {
		return nil, errors.New("unable to decode Storage Layout: " + err.Error())
	}

	// TODO: implement GetStorageRoot to reduce the response list
	historicStates := []*types.ParsedState{}
	for i := startBlockNumber; i <= endBlockNumber; i++ {
		rawStorage, err := r.db.GetStorage(address, i)
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
		historicStates = append(historicStates, &types.ParsedState{
			BlockNumber:     i,
			HistoricStorage: historicStorage,
		})
	}
	return &types.ReportingResponseTemplate{
		Address:       address,
		HistoricState: historicStates,
	}, nil
}

func (r *RPCAPIs) AddAddress(address common.Address, from *uint64) error {
	if address == (common.Address{}) {
		return errors.New("invalid input")
	}
	if from != nil && *from > 0 {
		// add address from
		return r.db.AddAddressFrom(address, *from)
	}
	return r.db.AddAddresses([]common.Address{address})
}

func (r *RPCAPIs) DeleteAddress(address common.Address) error {
	return r.db.DeleteAddress(address)
}

func (r *RPCAPIs) GetAddresses() ([]common.Address, error) {
	return r.db.GetAddresses()
}

func (r *RPCAPIs) GetContractTemplate(address common.Address) (string, error) {
	return r.db.GetContractTemplate(address)
}

func (r *RPCAPIs) AddABI(address common.Address, data string) error {
	// check ABI is valid
	if _, err := abi.JSON(strings.NewReader(data)); err != nil {
		return err
	}
	return r.contractTemplateManager.AddContractABI(address, data)
}

func (r *RPCAPIs) GetABI(address common.Address) (string, error) {
	return r.db.GetContractABI(address)
}

func (r *RPCAPIs) AddStorageABI(address common.Address, data string) error {
	var storageAbi types.SolidityStorageDocument
	if err := json.Unmarshal([]byte(data), &storageAbi); err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}
	return r.contractTemplateManager.AddStorageLayout(address, data)
}

func (r *RPCAPIs) GetStorageABI(address common.Address) (string, error) {
	return r.db.GetStorageLayout(address)
}

func (r *RPCAPIs) AddTemplate(name string, abiData string, layout string) error {
	// check ABI is valid
	if _, err := abi.JSON(strings.NewReader(abiData)); err != nil {
		return err
	}
	// check storage layout is valid
	var storageAbi types.SolidityStorageDocument
	if err := json.Unmarshal([]byte(layout), &storageAbi); err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}
	return r.db.AddTemplate(name, abiData, layout)
}

func (r *RPCAPIs) AssignTemplate(address common.Address, name string) error {
	return r.db.AssignTemplate(address, name)
}

func (r *RPCAPIs) GetTemplates() ([]string, error) {
	return r.db.GetTemplates()
}

func (r *RPCAPIs) GetTemplateDetails(templateName string) (*types.Template, error) {
	return r.db.GetTemplateDetails(templateName)
}
