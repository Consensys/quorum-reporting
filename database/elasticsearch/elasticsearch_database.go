package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/types"
)

type ElasticsearchDB struct {
	apiClient APIClient
}

func New(client APIClient) *ElasticsearchDB {
	db := &ElasticsearchDB{
		apiClient: client,
	}

	db.setupMappings()

	return db
}

func (es *ElasticsearchDB) setupMappings() error {
	mapping := `{"mappings":{"properties": {"internalCalls": {"type": "nested" }}}}`
	createRequest := esapi.IndicesCreateRequest{
		Index: "transaction",
		Body:  strings.NewReader(mapping),
	}

	//TODO: check error scenarios
	es.apiClient.DoRequest(createRequest)

	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: "contract"})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: "storage"})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: "event"})
	return nil
}

//AddressDB
func (es *ElasticsearchDB) AddAddresses(addresses []common.Address) error {
	//TODO: use bulk indexing
	for _, address := range addresses {
		contract := Contract{
			Address:             address,
			ABI:                 "",
			CreationTransaction: common.Hash{},
			LastFiltered:        0,
		}

		req := esapi.IndexRequest{
			Index:      "contract",
			DocumentID: address.String(),
			Body:       esutil.NewJSONReader(contract),
			Refresh:    "true",
			OpType:     "create", //This will only create if the contract does not exist
		}

		//TODO: bubble up this error
		es.apiClient.IndexRequest(req)
	}

	return nil
}

func (es *ElasticsearchDB) DeleteAddress(address common.Address) error {
	deleteRequest := esapi.DeleteRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Refresh:    "true",
	}

	_, err := es.apiClient.DoRequest(deleteRequest)
	if err != nil {
		return errors.New("error deleting address: " + err.Error())
	}

	//TODO: delete data from other indices (event + storage)
	return nil
}

func (es *ElasticsearchDB) GetAddresses() ([]common.Address, error) {
	results, err := es.apiClient.ScrollAllResults("contract", QueryAllAddressesTemplate)
	if err != nil {
		return nil, errors.New("error fetching addresses: " + err.Error())
	}
	converted := make([]common.Address, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})
		addr := data["address"].(string)
		converted[i] = common.HexToAddress(addr)
	}

	return converted, nil
}

//ABIDB
func (es *ElasticsearchDB) AddContractABI(address common.Address, abi string) error {
	//TODO: guard against unknown address?
	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"abi": abi,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	//TODO: check if error returned
	es.apiClient.DoRequest(updateRequest)
	return nil
}

func (es *ElasticsearchDB) GetContractABI(address common.Address) (string, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return "", err
	}
	return contract.ABI, nil
}

// BlockDB
func (es *ElasticsearchDB) WriteBlock(block *types.Block) error {
	req := esapi.IndexRequest{
		Index:      "block",
		DocumentID: block.Hash.String(),
		Body:       esutil.NewJSONReader(block),
		Refresh:    "true",
	}

	//TODO: check if response needs reading
	es.apiClient.DoRequest(req)
	return nil
}

func (es *ElasticsearchDB) ReadBlock(number uint64) (*types.Block, error) {
	//TODO: make more readable
	query := fmt.Sprintf(QueryByNumberTemplate, number)
	searchRequest := esapi.SearchRequest{
		Index: []string{"block"},
		Body:  strings.NewReader(query),
	}

	body, err := es.apiClient.DoRequest(searchRequest)
	if err != nil {
		return nil, err
	}

	//TODO: handle error
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if numberOfResults == 0 {
		return nil, fmt.Errorf("block number %d not found", number)
	}
	if numberOfResults > 1 {
		return nil, fmt.Errorf("too many block number %d's found", number)
	}

	result := es.getSingleDocumentData(response)
	marshalled, _ := json.Marshal(result)
	var block types.Block
	json.Unmarshal(marshalled, &block)
	return &block, nil
}

func (es *ElasticsearchDB) GetLastPersistedBlockNumber() (uint64, error) {
	// TODO: We need a separate db storage for last persisted
	query := `{"_source":["number"],"from":0,"size":1,"query":{"match_all":{}},"sort":[{"number":"desc"}]}`
	searchRequest := esapi.SearchRequest{
		Index: []string{"block"},
		Body:  strings.NewReader(query),
	}

	body, err := es.apiClient.DoRequest(searchRequest)
	if err != nil {
		//TODO: return error
		return 0, nil
	}

	var response map[string]interface{}
	json.Unmarshal(body, &response)

	//numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	//if numberOfResults == 0 {
	//	return 0, errors.New("last persisted block not found")
	//}
	//if numberOfResults > 1 {
	//	return 0, errors.New("too many blocks found for last persisted")
	//}

	/////////

	result := es.getSingleDocumentData(response)
	return uint64(result["number"].(float64)), nil
}

// TransactionDB
func (es *ElasticsearchDB) WriteTransaction(transaction *types.Transaction) error {
	//TODO: convert to internal transaction type

	if transaction.InternalCalls == nil {
		transaction.InternalCalls = make([]*types.InternalCall, 0)
	}

	req := esapi.IndexRequest{
		Index:      "transaction",
		DocumentID: transaction.Hash.String(),
		Body:       esutil.NewJSONReader(transaction),
		Refresh:    "true",
	}

	//TODO: check if response needs reading
	es.apiClient.DoRequest(req)
	return nil
}

func (es *ElasticsearchDB) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	return es.getTransactionByHash(hash)
}

// IndexDB
func (es *ElasticsearchDB) IndexBlock(addresses []common.Address, block *types.Block) error {
	// filter out registered and unfiltered address only
	filteredAddresses := map[common.Address]bool{}
	for _, address := range addresses {
		lastFiltered, _ := es.GetLastFiltered(address)
		if es.addressIsRegistered(address) && lastFiltered < block.Number {
			filteredAddresses[address] = true
			log.Printf("Index registered address %v at block %v.\n", address.Hex(), block.Number)
		}
	}

	// index transactions and events
	for _, txHash := range block.Transactions {
		transaction, _ := es.ReadTransaction(txHash)
		es.indexTransaction(filteredAddresses, transaction)
	}

	// index public storage
	es.indexStorage(filteredAddresses, block.Number, block.PublicState)
	//// index private storage
	es.indexStorage(filteredAddresses, block.Number, block.PrivateState)

	for addr := range filteredAddresses {
		es.updateLastFiltered(addr, block.Number)
	}
	return nil
}

func (es *ElasticsearchDB) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return common.Hash{}, err
	}
	return contract.CreationTransaction, nil
}

func (es *ElasticsearchDB) GetAllTransactionsToAddress(address common.Address) ([]common.Hash, error) {
	queryString := fmt.Sprintf(QueryByToAddressTemplate, address.String())
	results, _ := es.apiClient.ScrollAllResults("transaction", queryString)

	converted := make([]common.Hash, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})
		addr := data["hash"].(string)
		converted[i] = common.HexToHash(addr)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetAllTransactionsInternalToAddress(address common.Address) ([]common.Hash, error) {
	queryString := fmt.Sprintf(QueryInternalTransactions, address.String())
	results, _ := es.apiClient.ScrollAllResults("transaction", queryString)

	converted := make([]common.Hash, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})
		addr := data["hash"].(string)
		converted[i] = common.HexToHash(addr)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetAllEventsByAddress(address common.Address) ([]*types.Event, error) {
	query := fmt.Sprintf(QueryByAddressTemplate, address.String())
	results, _ := es.apiClient.ScrollAllResults("event", query)

	convertedList := make([]*types.Event, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})

		marshalled, _ := json.Marshal(data)
		var event Event
		json.Unmarshal(marshalled, &event)

		convertedList[i] = &types.Event{
			Index:           event.LogIndex,
			Address:         event.Address,
			Topics:          event.Topics,
			Data:            event.Data,
			BlockNumber:     event.BlockNumber,
			TransactionHash: event.TransactionHash,
		}
	}

	return convertedList, nil
}

func (es *ElasticsearchDB) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	query := fmt.Sprintf(QueryByAddressAndBlockNumberTemplate, address.String(), blockNumber)

	searchRequest := esapi.SearchRequest{
		Index: []string{"storage"},
		Body:  strings.NewReader(query),
	}

	body, _ := es.apiClient.DoRequest(searchRequest)

	//TODO: handle error
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if numberOfResults == 0 {
		return nil, fmt.Errorf("address %s not found", address.String())
	}
	if numberOfResults > 1 {
		return nil, fmt.Errorf("too many addresses %s's found", address.String())
	}

	/////////

	result := es.getSingleDocumentData(response)
	retrievedStorage := result["storageMap"].(map[string]interface{})
	storage := make(map[common.Hash]string)
	for hsh, val := range retrievedStorage {
		storage[common.HexToHash(hsh)] = val.(string)
	}
	return storage, nil
}

func (es *ElasticsearchDB) GetLastFiltered(address common.Address) (uint64, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return 0, err
	}
	return contract.LastFiltered, nil
}

// Internal functions

func (es *ElasticsearchDB) getContractByAddress(address common.Address) (*Contract, error) {
	//TODO: make more readable
	query := fmt.Sprintf(QueryByAddressTemplate, address.String())

	searchRequest := esapi.SearchRequest{
		Index: []string{"contract"},
		Body:  strings.NewReader(query),
	}

	body, err := es.apiClient.DoRequest(searchRequest)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	json.Unmarshal(body, &response)

	numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if numberOfResults == 0 {
		return nil, ErrAddressNotFound
	}
	if numberOfResults > 1 {
		return nil, ErrTooManyResults
	}

	/////////

	result := es.getSingleDocumentData(response)
	marshalled, _ := json.Marshal(result)
	var contract Contract
	json.Unmarshal(marshalled, &contract)
	return &contract, nil
}

func (es *ElasticsearchDB) getTransactionByHash(hash common.Hash) (*types.Transaction, error) {
	//TODO: make more readable
	query := fmt.Sprintf(QueryByHashTemplate, hash.String())

	searchRequest := esapi.SearchRequest{
		Index: []string{"transaction"},
		Body:  strings.NewReader(query),
	}

	body, _ := es.apiClient.DoRequest(searchRequest)

	//TODO: handle error
	var response map[string]interface{}
	json.Unmarshal(body, &response)

	numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if numberOfResults == 0 {
		return nil, fmt.Errorf("transaction %s not found", hash.String())
	}
	if numberOfResults > 1 {
		return nil, fmt.Errorf("too many transactions with hash %s found", hash.String())
	}

	result := es.getSingleDocumentData(response)
	marshalled, _ := json.Marshal(result)
	var transaction types.Transaction
	json.Unmarshal(marshalled, &transaction)
	return &transaction, nil
}

func (es *ElasticsearchDB) getSingleDocumentData(esResponse map[string]interface{}) map[string]interface{} {
	return esResponse["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})
}

func (es *ElasticsearchDB) addressIsRegistered(address common.Address) bool {
	allAddresses, _ := es.GetAddresses()
	for _, registeredAddress := range allAddresses {
		if registeredAddress == address {
			return true
		}
	}
	return false
}

func (es *ElasticsearchDB) indexTransaction(filteredAddresses map[common.Address]bool, tx *types.Transaction) {
	// Compare the address with tx.To and tx.CreatedContract to check if the transaction is related.
	if filteredAddresses[tx.CreatedContract] {
		es.updateCreatedTx(tx.CreatedContract, tx.Hash)
		log.Printf("Index contract creation tx %v of registered address %v.\n", tx.Hash.Hex(), tx.CreatedContract.Hex())
	}

	// Index events emitted by the given address
	for _, event := range tx.Events {
		if filteredAddresses[event.Address] {
			es.createEvent(event)
			log.Printf("Append event emitted in transaction %v to registered address %v.\n", event.TransactionHash.Hex(), event.Address.Hex())
		}
	}
}

func (es *ElasticsearchDB) updateCreatedTx(address common.Address, creationTxHash common.Hash) error {
	return es.updateContract(address, "creationTx", creationTxHash.String())
}

func (es *ElasticsearchDB) updateLastFiltered(address common.Address, lastFiltered uint64) error {
	return es.updateContract(address, "lastFiltered", lastFiltered)
}

func (es *ElasticsearchDB) updateContract(address common.Address, property string, value interface{}) error {
	//TODO: guard against unknown address?
	query := map[string]interface{}{
		"doc": map[string]interface{}{
			property: value,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	//TODO: check if error returned
	es.apiClient.DoRequest(updateRequest)
	return nil
}

func (es *ElasticsearchDB) createEvent(event *types.Event) error {
	converted := Event{
		Address:         event.Address,
		BlockNumber:     event.BlockNumber,
		Data:            event.Data,
		LogIndex:        event.Index,
		Topics:          event.Topics,
		TransactionHash: event.TransactionHash,
	}

	req := esapi.IndexRequest{
		Index:      "event",
		DocumentID: strconv.FormatUint(event.BlockNumber, 10) + "-" + strconv.FormatUint(event.Index, 10),
		Body:       esutil.NewJSONReader(converted),
		Refresh:    "true",
	}

	//TODO: check response
	es.apiClient.DoRequest(req)
	return nil
}

func (es *ElasticsearchDB) indexStorage(filteredAddresses map[common.Address]bool, blockNumber uint64, stateDump *state.Dump) error {
	if stateDump == nil {
		return nil
	}

	for address, account := range stateDump.Accounts {
		if filteredAddresses[address] {
			stateObj := State{
				Address:     address,
				BlockNumber: blockNumber,
				StorageRoot: common.HexToHash(account.Root),
				StorageMap:  account.Storage,
			}

			req := esapi.IndexRequest{
				Index:      "storage",
				DocumentID: address.String() + "-" + strconv.FormatUint(blockNumber, 10),
				Body:       esutil.NewJSONReader(stateObj),
				Refresh:    "true",
			}

			//TODO: check response
			es.apiClient.DoRequest(req)
		}
	}

	return nil
}
