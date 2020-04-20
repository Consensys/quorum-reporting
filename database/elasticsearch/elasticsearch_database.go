package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"quorumengineering/quorum-report/types"
	"strings"
)

type Database struct {
	client *elasticsearch7.Client
}

func New() *Database {
	c, _ := elasticsearch7.NewDefaultClient()
	return &Database{
		client: c,
	}
}

//AddressDB
func (es *Database) AddAddresses(addresses []common.Address) error {
	toSave := make([]Contract, len(addresses))

	//TODO: use bulk indexing
	//TODO: check exists
	for i, address := range addresses {
		toSave[i] = Contract{
			Address:             address,
			ABI:                 nil,
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

func (es *Database) DeleteAddress(address common.Address) error {
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

//TODO: fix
func (es *Database) GetAddresses() ([]common.Address, error) {
	//queryString := `{ "_source": ["address"], "query": { "match_all": {} } }`
	//searchRequest := esapi.SearchRequest{
	//	Index: []string{"contract"},
	//	Body: strings.NewReader(queryString),
	//}
	//
	//res, err := searchRequest.Do(context.TODO(), es.client)
	//if err != nil {
	//	return nil, fmt.Errorf("error getting response: %s", err.Error())
	//}

	return []common.Address{common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")}, nil
}

//ABIDB
func (es *Database) AddContractABI(address common.Address, abi *abi.ABI) error {
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

	//TODO: check if error returned with unknown address
	res, err := updateRequest.Do(context.TODO(), es.client)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("error getting response: %s", err.Error())
	}
	return nil
}

func (es *Database) GetContractABI(address common.Address) (*abi.ABI, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return nil, err
	}
	return contract.ABI, nil
}

// BlockDB
func (es *Database) WriteBlock(block *types.Block) error {
	blockToSave := Block{
		Hash:         block.Hash,
		ParentHash:   block.ParentHash,
		StateRoot:    block.StateRoot,
		TxRoot:       block.TxRoot,
		ReceiptRoot:  block.ReceiptRoot,
		Number:       block.Number,
		GasLimit:     block.GasLimit,
		GasUsed:      block.GasUsed,
		Timestamp:    block.Timestamp,
		ExtraData:    block.ExtraData,
		Transactions: block.Transactions,
	}

	req := esapi.IndexRequest{
		Index:      "block",
		DocumentID: block.Hash.String(),
		Body:       esutil.NewJSONReader(blockToSave),
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

func (es *Database) ReadBlock(number uint64) (*types.Block, error) {
	//TODO: make more readable
	queryTemplate := "{\"query\": {\"bool\": {\"must\": [{ \"match\": { \"number\": \"%d\" }}]}}}"
	query := fmt.Sprintf(queryTemplate, number)

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

	numberOfResults := int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if numberOfResults == 0 {
		return nil, fmt.Errorf("block number %d not found", number)
	}
	if numberOfResults > 1 {
		return nil, fmt.Errorf("too many block number %d's found", number)
	}

	result := response["hits"].(map[string]interface{})["hits"].(map[string]interface{})["_source"].(map[string]interface{})
	marshalled, _ := json.Marshal(result)
	var block types.Block
	json.Unmarshal(marshalled, &block)
	return &block, nil
}

//TODO: fix
func (es *Database) GetLastPersistedBlockNumber() (uint64, error) {
	return 0, nil
}

// TransactionDB
func (es *Database) WriteTransaction(transaction *types.Transaction) error {
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

func (es *Database) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	return es.getTransactionByHash(hash)
}

// IndexDB
func (es *Database) IndexBlock([]common.Address, *types.Block) error {
	return errors.New("not implemented 6")
}

func (es *Database) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return common.Hash{}, err
	}
	return contract.CreationTransaction, nil
}

func (es *Database) GetAllTransactionsToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented 8")
}

func (es *Database) GetAllTransactionsInternalToAddress(common.Address) ([]common.Hash, error) {
	return nil, errors.New("not implemented 9")
}

func (es *Database) GetAllEventsByAddress(common.Address) ([]*types.Event, error) {
	return nil, errors.New("not implemented 10")
}

func (es *Database) GetStorage(common.Address, uint64) (map[common.Hash]string, error) {
	return nil, errors.New("not implemented 11")
}

//TODO: fix
func (es *Database) GetLastFiltered(common.Address) (uint64, error) {
	return 0, nil
}

// Internal functions

func (es *Database) getContractByAddress(address common.Address) (*Contract, error) {
	//TODO: make more readable
	queryTemplate := "{\"query\": {\"bool\": {\"must\": [{ \"match\": { \"address\": \"%s\" }}]}}}"
	query := fmt.Sprintf(queryTemplate, address.String())

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

	result := response["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})
	marshalled, _ := json.Marshal(result)
	fmt.Println(string(marshalled))
	var contract Contract
	json.Unmarshal(marshalled, &contract)
	return &contract, nil
}

func (es *Database) getTransactionByHash(hash common.Hash) (*types.Transaction, error) {
	//TODO: make more readable
	queryTemplate := "{\"query\": {\"bool\": {\"must\": [{ \"match\": { \"hash\": \"%s\" }}]}}}"
	query := fmt.Sprintf(queryTemplate, hash.String())

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

func (es *Database) getSingleDocumentData(esResponse map[string]interface{}) map[string]interface{} {
	return esResponse["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})
}
