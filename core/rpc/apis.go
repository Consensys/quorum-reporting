package rpc

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (r *RPCAPIs) GetLastPersistedBlockNumber(req *http.Request, args *NullArgs, reply *uint64) error {
	val, err := r.db.GetLastPersistedBlockNumber()
	if err != nil {
		return err
	}
	*reply = val
	return nil
}

func (r *RPCAPIs) GetLastFiltered(req *http.Request, args *types.Address, reply *uint64) error {
	val, err := r.db.GetLastFiltered(*args)
	if err != nil {
		return err
	}
	*reply = val
	return nil
}

func (r *RPCAPIs) GetBlock(req *http.Request, blockNumber *uint64, reply *types.Block) error {
	block, err := r.db.ReadBlock(*blockNumber)
	if err != nil {
		return err
	}
	*reply = *block
	return nil
}

func (r *RPCAPIs) GetTransaction(req *http.Request, hash *types.Hash, reply *types.ParsedTransaction) error {
	if hash.IsEmpty() {
		return errors.New("no transaction hash given")
	}
	tx, err := r.db.ReadTransaction(*hash)
	if err != nil {
		return err
	}
	address := tx.To
	if address.IsEmpty() {
		address = tx.CreatedContract
	}
	contractABI, err := r.db.GetContractABI(address)
	if err != nil {
		return err
	}
	parsedTx := &types.ParsedTransaction{
		RawTransaction: tx,
	}
	if contractABI != "" {
		if err = parsedTx.ParseTransaction(contractABI); err != nil {
			return err
		}
	}
	parsedTx.ParsedEvents = make([]*types.ParsedEvent, len(parsedTx.RawTransaction.Events))
	for i, e := range parsedTx.RawTransaction.Events {
		parsedTx.ParsedEvents[i] = &types.ParsedEvent{
			RawEvent: e,
		}
		contractABI, err := r.db.GetContractABI(e.Address)
		if err != nil {
			return err
		}
		if contractABI != "" {
			if err := parsedTx.ParsedEvents[i].ParseEvent(contractABI); err != nil {
				return err
			}
		}
	}
	*reply = *parsedTx
	return nil
}

func (r *RPCAPIs) GetContractCreationTransaction(req *http.Request, address *types.Address, reply *types.Hash) error {
	txHash, err := r.db.GetContractCreationTransaction(*address)
	if err != nil {
		return err
	}
	if txHash.IsEmpty() {
		return errors.New("contract creation tx not found")
	}
	*reply = txHash
	return nil
}

func (r *RPCAPIs) GetAllTransactionsToAddress(req *http.Request, args *AddressWithOptions, reply *TransactionsResp) error {
	if args.Address == nil {
		return ErrNoAddress
	}
	if args.Options == nil {
		args.Options = &types.QueryOptions{}
	}
	args.Options.SetDefaults()

	total, err := r.db.GetTransactionsToAddressTotal(*args.Address, args.Options)
	if err != nil {
		return err
	}
	txs, err := r.db.GetAllTransactionsToAddress(*args.Address, args.Options)
	if err != nil {
		return err
	}

	*reply = TransactionsResp{
		Transactions: txs,
		Total:        total,
		Options:      args.Options,
	}
	return nil
}

func (r *RPCAPIs) GetAllTransactionsInternalToAddress(req *http.Request, args *AddressWithOptions, reply *TransactionsResp) error {
	if args.Address == nil {
		return ErrNoAddress
	}
	if args.Options == nil {
		args.Options = &types.QueryOptions{}
	}
	args.Options.SetDefaults()

	total, err := r.db.GetTransactionsInternalToAddressTotal(*args.Address, args.Options)
	if err != nil {
		return err
	}
	txs, err := r.db.GetAllTransactionsInternalToAddress(*args.Address, args.Options)
	if err != nil {
		return err
	}

	*reply = TransactionsResp{
		Transactions: txs,
		Total:        total,
		Options:      args.Options,
	}
	return nil
}

func (r *RPCAPIs) GetAllEventsFromAddress(req *http.Request, args *AddressWithOptions, reply *EventsResp) error {
	if args.Address == nil {
		return ErrNoAddress
	}
	if args.Options == nil {
		args.Options = &types.QueryOptions{}
	}
	args.Options.SetDefaults()

	total, err := r.db.GetEventsFromAddressTotal(*args.Address, args.Options)
	if err != nil {
		return err
	}
	events, err := r.db.GetAllEventsFromAddress(*args.Address, args.Options)
	if err != nil {
		return err
	}
	contractABI, err := r.db.GetContractABI(*args.Address)
	if err != nil {
		return err
	}
	parsedEvents := make([]*types.ParsedEvent, len(events))
	for i, e := range events {
		parsedEvents[i] = &types.ParsedEvent{
			RawEvent: e,
		}
		if contractABI != "" {
			if err = parsedEvents[i].ParseEvent(contractABI); err != nil {
				return err
			}
		}
	}

	*reply = EventsResp{
		Events:  parsedEvents,
		Total:   total,
		Options: args.Options,
	}
	return nil
}

func (r *RPCAPIs) GetStorage(req *http.Request, args *AddressWithOptionalBlock, reply *map[types.Hash]string) error {
	if args.Address == nil {
		return ErrNoAddress
	}
	if args.BlockNumber == nil {
		lastFiltered, err := r.db.GetLastFiltered(*args.Address)
		if err != nil {
			if err == database.ErrNotFound {
				return errors.New("address is not indexed")
			}
			return err
		}
		args.BlockNumber = &lastFiltered
	}
	result, err := r.db.GetStorage(*args.Address, *args.BlockNumber)
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

func (r *RPCAPIs) GetStorageHistory(req *http.Request, args *AddressWithBlockRange, reply *types.ReportingResponseTemplate) error {
	if args.Address == nil {
		return ErrNoAddress
	}

	rawAbi, err := r.db.GetStorageLayout(*args.Address)
	if err != nil {
		return err
	}
	if rawAbi == "" {
		return errors.New("no Storage Layout present to parse with")
	}
	var parsedAbi types.SolidityStorageDocument
	if err = json.Unmarshal([]byte(rawAbi), &parsedAbi); err != nil {
		return errors.New("unable to decode Storage Layout: " + err.Error())
	}

	// TODO: implement GetStorageRoot to reduce the response list
	historicStates := []*types.ParsedState{}
	for i := args.StartBlockNumber; i <= args.EndBlockNumber; i++ {
		rawStorage, err := r.db.GetStorage(*args.Address, i)
		if err != nil {
			return err
		}
		if rawStorage == nil {
			continue
		}
		historicStorage, err := storageparsing.ParseRawStorage(rawStorage, parsedAbi)
		if err != nil {
			return err
		}
		historicStates = append(historicStates, &types.ParsedState{
			BlockNumber:     i,
			HistoricStorage: historicStorage,
		})
	}
	*reply = types.ReportingResponseTemplate{
		Address:       *args.Address,
		HistoricState: historicStates,
	}
	return nil
}

func (r *RPCAPIs) AddAddress(req *http.Request, args *AddressWithOptionalBlock, reply *NullArgs) error {
	if args.Address == nil {
		return ErrNoAddress
	}

	if args.BlockNumber != nil && *args.BlockNumber > 0 {
		// add address from
		return r.db.AddAddressFrom(*args.Address, *args.BlockNumber)
	}
	return r.db.AddAddresses([]types.Address{*args.Address})
}

func (r *RPCAPIs) DeleteAddress(req *http.Request, address *types.Address, reply *NullArgs) error {
	return r.db.DeleteAddress(*address)
}

func (r *RPCAPIs) GetAddresses(req *http.Request, args *NullArgs, reply *[]types.Address) error {
	result, err := r.db.GetAddresses()
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

func (r *RPCAPIs) GetContractTemplate(req *http.Request, address *types.Address, reply *string) error {
	result, err := r.db.GetContractTemplate(*address)
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

func (r *RPCAPIs) AddABI(req *http.Request, args *AddressWithData, reply *NullArgs) error {
	if args.Address == nil {
		return ErrNoAddress
	}

	// check ABI is valid
	if _, err := types.NewABIStructureFromJSON(args.Data); err != nil {
		return err
	}
	return r.contractTemplateManager.AddContractABI(*args.Address, args.Data)
}

func (r *RPCAPIs) GetABI(req *http.Request, address *types.Address, reply *string) error {
	result, err := r.db.GetContractABI(*address)
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

func (r *RPCAPIs) AddStorageABI(req *http.Request, args *AddressWithData, reply *NullArgs) error {
	if args.Address == nil {
		return ErrNoAddress
	}

	var storageAbi types.SolidityStorageDocument
	if err := json.Unmarshal([]byte(args.Data), &storageAbi); err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}
	return r.contractTemplateManager.AddStorageLayout(*args.Address, args.Data)
}

func (r *RPCAPIs) GetStorageABI(req *http.Request, address *types.Address, reply *string) error {
	result, err := r.db.GetStorageLayout(*address)
	if err != nil {
		return err
	}
	*reply = result
	return nil
}

func (r *RPCAPIs) AddTemplate(req *http.Request, args *TemplateArgs, reply *NullArgs) error {
	// check ABI is valid
	if _, err := types.NewABIStructureFromJSON(args.Abi); err != nil {
		return err
	}
	// check storage layout is valid
	var storageAbi types.SolidityStorageDocument
	if err := json.Unmarshal([]byte(args.StorageLayout), &storageAbi); err != nil {
		return errors.New("invalid JSON: " + err.Error())
	}
	return r.db.AddTemplate(args.Name, args.Abi, args.StorageLayout)
}

func (r *RPCAPIs) AssignTemplate(req *http.Request, args *AddressWithData, reply *NullArgs) error {
	if args.Address == nil {
		return ErrNoAddress
	}
	return r.db.AssignTemplate(*args.Address, args.Data)
}

func (r *RPCAPIs) GetTemplates(req *http.Request, args *NullArgs, result *[]string) error {
	templates, err := r.db.GetTemplates()
	if err != nil {
		return err
	}
	*result = templates
	return nil
}

func (r *RPCAPIs) GetTemplateDetails(req *http.Request, templateName *string, reply *types.Template) error {
	template, err := r.db.GetTemplateDetails(*templateName)
	if err != nil {
		return err
	}
	*reply = *template
	return nil
}
