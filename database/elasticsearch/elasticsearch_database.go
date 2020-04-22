package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"io"
	"log"
	"quorumengineering/quorum-report/types"
	"strconv"
	"strings"
	"time"
)

type ElasticsearchDB struct {
	client *elasticsearch7.Client
}

func New() *ElasticsearchDB {
	c, _ := elasticsearch7.NewDefaultClient()
	return &ElasticsearchDB{
		client: c,
	}
}

//AddressDB
func (es *ElasticsearchDB) AddAddresses(addresses []common.Address) error {
	toSave := make([]Contract, len(addresses))

	//TODO: use bulk indexing
	//TODO: check exists
	for i, address := range addresses {
		toSave[i] = Contract{
			Address:             address,
			ABI:                 "",
			CreationTransaction: common.Hash{},
			LastFiltered:        0,
		}

		req := esapi.IndexRequest{
			Index:      "contract",
			DocumentID: address.String(),
			Body:       esutil.NewJSONReader(toSave[i]),
			Refresh:    "true",
		}

		//TODO: check if response needs reading
		res, err := req.Do(context.TODO(), es.client)
		if err != nil {
			return fmt.Errorf("error getting response: %s", err.Error())
		}
		res.Body.Close()
	}

	return nil
}

func (es *ElasticsearchDB) DeleteAddress(address common.Address) error {
	deleteRequest := esapi.DeleteRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Refresh:    "true",
	}

	//TODO: check if response needs reading
	res, err := deleteRequest.Do(context.TODO(), es.client)
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	defer res.Body.Close()

	//TODO: delete data from other indices
	return nil
}

func (es *ElasticsearchDB) GetAddresses() ([]common.Address, error) {
	results := es.scrollAllResults("contract", strings.NewReader(QueryAllAddressesTemplate))
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
	updateRequest := esapi.UpdateRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Body:       strings.NewReader(fmt.Sprintf(UpdateContractABITemplate, abi)),
		Refresh:    "true",
	}

	//TODO: check if error returned
	res, err := updateRequest.Do(context.TODO(), es.client)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	return nil
}

func (es *ElasticsearchDB) GetContractABI(address common.Address) (string, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return "", err
	}
	abi, _ := json.Marshal(contract.ABI)
	return string(abi), nil
}

// BlockDB
func (es *ElasticsearchDB) WriteBlock(block *types.Block) error {
	//blockToSave := Block{
	//	Hash:         block.Hash,
	//	ParentHash:   block.ParentHash,
	//	StateRoot:    block.StateRoot,
	//	TxRoot:       block.TxRoot,
	//	ReceiptRoot:  block.ReceiptRoot,
	//	Number:       block.Number,
	//	GasLimit:     block.GasLimit,
	//	GasUsed:      block.GasUsed,
	//	Timestamp:    block.Timestamp,
	//	ExtraData:    block.ExtraData,
	//	Transactions: block.Transactions,
	//}

	req := esapi.IndexRequest{
		Index:      "block",
		DocumentID: block.Hash.String(),
		Body:       esutil.NewJSONReader(block),
		Refresh:    "true",
	}

	//TODO: check if response needs reading
	res, err := req.Do(context.TODO(), es.client)
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	res.Body.Close()
	return nil
}

func (es *ElasticsearchDB) ReadBlock(number uint64) (*types.Block, error) {
	//TODO: make more readable
	query := fmt.Sprintf(QueryByNumberTemplate, number)
	searchRequest := esapi.SearchRequest{
		Index: []string{"block"},
		Body:  strings.NewReader(query),
	}

	res, err := searchRequest.Do(context.TODO(), es.client)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err.Error())
	}
	defer res.Body.Close()

	//TODO: handle error
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)

	if response["error"] != nil {
		return nil, errors.New("error talking to db")
	}

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

	res, err := searchRequest.Do(context.TODO(), es.client)
	if err != nil {
		return 0, fmt.Errorf("error getting response: %s", err.Error())
	}
	defer res.Body.Close()

	//TODO: handle error
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)

	if response["error"] != nil {
		return 0, nil
	}

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

	req := esapi.IndexRequest{
		Index:      "transaction",
		DocumentID: transaction.Hash.String(),
		Body:       esutil.NewJSONReader(transaction),
		Refresh:    "true",
	}

	//TODO: check if response needs reading
	res, err := req.Do(context.TODO(), es.client)
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	res.Body.Close()
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
	results := es.scrollAllResults("transaction", strings.NewReader(queryString))

	converted := make([]common.Hash, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})
		addr := data["hash"].(string)
		converted[i] = common.HexToHash(addr)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetAllTransactionsInternalToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented 9")
}

func (es *ElasticsearchDB) GetAllEventsByAddress(address common.Address) ([]*types.Event, error) {
	query := fmt.Sprintf(QueryByAddressTemplate, address.String())
	results := es.scrollAllResults("events", strings.NewReader(query))

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

	res, err := searchRequest.Do(context.TODO(), es.client)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err.Error())
	}
	defer res.Body.Close()

	//TODO: handle error
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)

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

	res, err := searchRequest.Do(context.TODO(), es.client)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err.Error())
	}
	defer res.Body.Close()

	//TODO: handle error
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)

	numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if numberOfResults == 0 {
		return nil, fmt.Errorf("address %s not found", address.String())
	}
	if numberOfResults > 1 {
		return nil, fmt.Errorf("too many addresses %s's found", address.String())
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

	res, err := searchRequest.Do(context.TODO(), es.client)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err.Error())
	}
	defer res.Body.Close()

	//TODO: handle error
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)

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

	//for _, internalCall := range tx.InternalCalls {
	//	if filteredAddresses[internalCall.To] {
	//		db.txIndexDB[internalCall.To].txsInternalTo = append(db.txIndexDB[internalCall.To].txsInternalTo, tx.Hash)
	//		log.Printf("Index tx %v internal calling registered address %v.\n", tx.Hash.Hex(), internalCall.To.Hex())
	//	}
	//}

	// Index events emitted by the given address
	for _, event := range tx.Events {
		if filteredAddresses[event.Address] {
			es.createEvent(event)
			log.Printf("Append event emitted in transaction %v to registered address %v.\n", event.TransactionHash.Hex(), event.Address.Hex())
		}
	}
}

func (es *ElasticsearchDB) updateCreatedTx(address common.Address, creationTxHash common.Hash) error {
	//TODO: guard against unknown address?
	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"creationTx": creationTxHash.String(),
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	//TODO: check if error returned
	res, err := updateRequest.Do(context.TODO(), es.client)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	return nil
}

func (es *ElasticsearchDB) updateLastFiltered(address common.Address, lastFiltered uint64) error {
	//TODO: guard against unknown address?
	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"lastFiltered": lastFiltered,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      "contract",
		DocumentID: address.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	//TODO: check if error returned
	res, err := updateRequest.Do(context.TODO(), es.client)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	return nil
}

func (es *ElasticsearchDB) createEvent(event *types.Event) error {
	converted := Event{
		ID:              strconv.FormatUint(event.BlockNumber, 10) + "-" + strconv.FormatUint(event.Index, 10),
		Address:         event.Address,
		BlockNumber:     event.BlockNumber,
		Data:            event.Data,
		LogIndex:        event.Index,
		Topics:          event.Topics,
		TransactionHash: event.TransactionHash,
	}

	req := esapi.IndexRequest{
		Index:      "events",
		DocumentID: strconv.FormatUint(event.BlockNumber, 10) + "-" + strconv.FormatUint(event.Index, 10),
		Body:       esutil.NewJSONReader(converted),
		Refresh:    "true",
	}

	//TODO: check if response needs reading
	res, err := req.Do(context.TODO(), es.client)
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	res.Body.Close()
	return nil
}

func (es *ElasticsearchDB) scrollAllResults(index string, query io.Reader) []interface{} {
	var (
		batchNum int
		scrollID string
		results  []interface{}
	)

	res, _ := es.client.Search(
		es.client.Search.WithIndex(index),
		es.client.Search.WithSort("_doc"),
		es.client.Search.WithSize(10),
		es.client.Search.WithScroll(time.Minute),
		es.client.Search.WithBody(query),
	)

	// Handle the first batch of data and extract the scrollID
	//
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)
	res.Body.Close()

	scrollID = response["_scroll_id"].(string)
	hits := response["hits"].(map[string]interface{})["hits"].([]interface{})
	results = append(results, hits...)

	// Perform the scroll requests in sequence
	//
	for {
		batchNum++

		// Perform the scroll request and pass the scrollID and scroll duration
		//
		res, err := es.client.Scroll(es.client.Scroll.WithScrollID(scrollID), es.client.Scroll.WithScroll(time.Minute))
		if err != nil {
			//log.Fatalf("Error: %s", err)
		}
		if res.IsError() {
			//log.Fatalf("Error response: %s", res)
		}

		var scrollResponse map[string]interface{}
		_ = json.NewDecoder(res.Body).Decode(&scrollResponse)
		res.Body.Close()

		// Extract the scrollID from response
		//
		scrollID = scrollResponse["_scroll_id"].(string)

		// Extract the search results
		//
		hits := scrollResponse["hits"].(map[string]interface{})["hits"].([]interface{})

		// Break out of the loop when there are no results
		//
		if len(hits) < 1 {
			//log.Println("Finished scrolling")
			break
		}

		results = append(results, hits...)
	}

	return results
}

func (es *ElasticsearchDB) indexStorage(filteredAddresses map[common.Address]bool, blockNumber uint64, stateDump *state.Dump) error {
	if stateDump != nil {
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

				//TODO: check if response needs reading
				res, err := req.Do(context.TODO(), es.client)
				if err != nil {
					return fmt.Errorf("error getting response: %s", err.Error())
				}
				res.Body.Close()
			}
		}
	}

	return nil
}
