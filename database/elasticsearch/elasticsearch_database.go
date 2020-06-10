package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/types"
)

type ElasticsearchDB struct {
	apiClient APIClient
}

func New(client APIClient) (*ElasticsearchDB, error) {
	db := &ElasticsearchDB{
		apiClient: client,
	}

	initialized, err := db.checkIsInitialized()
	if err != nil {
		log.Printf("Error communicating with ElasticSearch: %v.\n", err)
		log.Println("Please check all ElasticSearch settings, including certificates, URL and username/password.")
		return nil, err
	}
	if !initialized {
		if err := db.init(); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func (es *ElasticsearchDB) init() error {
	mapping := `{"mappings":{"properties": {"internalCalls": {"type": "nested" }}}}`
	createRequest := esapi.IndicesCreateRequest{
		Index: TransactionIndex,
		Body:  strings.NewReader(mapping),
	}

	//TODO: check error scenarios
	es.apiClient.DoRequest(createRequest)

	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: ContractIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: StorageIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: EventIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: MetaIndex})

	req := esapi.IndexRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
		Body:       strings.NewReader(`{"lastPersisted": 0}`),
		Refresh:    "true",
		OpType:     "create",
	}
	es.apiClient.DoRequest(req)

	return nil
}

//AddressDB
func (es *ElasticsearchDB) AddAddresses(addresses []common.Address) error {
	if len(addresses) == 0 {
		return nil
	}
	// Only use bulk update if more than one address is given
	if len(addresses) > 1 {
		bi := es.apiClient.GetBulkHandler(ContractIndex)

		var (
			wg        sync.WaitGroup
			returnErr error
		)
		for _, address := range addresses {
			contract := Contract{
				Address:             address,
				ABI:                 "",
				StorageABI:          "",
				CreationTransaction: common.Hash{},
				LastFiltered:        0,
			}
			wg.Add(1)
			bi.Add(
				context.Background(),
				esutil.BulkIndexerItem{
					Action:     "create",
					DocumentID: address.String(),
					Body:       esutil.NewJSONReader(contract),
					OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem) {
						wg.Done()
					},
					OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem, err error) {
						returnErr = err
						wg.Done()
					},
				},
			)
		}

		wg.Wait()
		return returnErr
	}
	// add single address
	contract := Contract{
		Address:             addresses[0],
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}

	req := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addresses[0].String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
		OpType:     "create", //This will only create if the contract does not exist
	}
	_, err := es.apiClient.DoRequest(req)
	return err
}

func (es *ElasticsearchDB) DeleteAddress(address common.Address) error {
	deleteRequest := esapi.DeleteRequest{
		Index:      ContractIndex,
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
	results, err := es.apiClient.ScrollAllResults(ContractIndex, QueryAllAddressesTemplate)
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
	return es.updateContract(address, "abi", abi)
}

func (es *ElasticsearchDB) GetContractABI(address common.Address) (string, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil && err != ErrNotFound {
		return "", err
	}
	if contract != nil {
		return contract.ABI, nil
	}
	return "", nil
}

func (es *ElasticsearchDB) AddStorageABI(address common.Address, abi string) error {
	return es.updateContract(address, "storageAbi", abi)
}

func (es *ElasticsearchDB) GetStorageABI(address common.Address) (string, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil && err != ErrNotFound {
		return "", err
	}
	if contract != nil {
		return contract.StorageABI, nil
	}
	return "", nil
}

// BlockDB
func (es *ElasticsearchDB) WriteBlock(block *types.Block) error {
	var internalBlock Block
	internalBlock.From(block)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: strconv.FormatUint(block.Number, 10),
		Body:       esutil.NewJSONReader(internalBlock),
		Refresh:    "true",
	}

	if _, err := es.apiClient.DoRequest(req); err != nil {
		return err
	}

	return es.updateLastPersisted(block.Number)
}

func (es *ElasticsearchDB) WriteBlocks(blocks []*types.Block) error {
	if len(blocks) == 0 {
		return nil
	}
	if len(blocks) == 1 {
		return es.WriteBlock(blocks[0])
	}

	bi := es.apiClient.GetBulkHandler(BlockIndex)
	var (
		wg        sync.WaitGroup
		returnErr error
	)
	wg.Add(len(blocks))
	for _, block := range blocks {
		var dbBlock Block
		dbBlock.From(block)
		err := bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: strconv.FormatUint(block.Number, 10),
				Body:       esutil.NewJSONReader(dbBlock),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem) {
					wg.Done()
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem, err error) {
					returnErr = err
					wg.Done()
				},
			})
		if err != nil {
			return err
		}
	}
	wg.Wait()
	if returnErr != nil {
		return returnErr
	}

	//find lowest block number
	lowest := blocks[0].Number
	for _, block := range blocks {
		if block.Number < lowest {
			lowest = block.Number
		}
	}
	return es.updateLastPersisted(lowest)
}

func (es *ElasticsearchDB) ReadBlock(number uint64) (*types.Block, error) {
	fetchReq := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: strconv.FormatUint(number, 10),
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return nil, err
	}

	var blockResult BlockQueryResult
	err = json.Unmarshal(body, &blockResult)
	if err != nil {
		return nil, err
	}
	return blockResult.Source.To(), nil
}

func (es *ElasticsearchDB) GetLastPersistedBlockNumber() (uint64, error) {
	fetchReq := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return 0, err
	}

	var lastPersisted LastPersistedResult
	json.Unmarshal(body, &lastPersisted)
	return lastPersisted.Source.LastPersisted, nil
}

// TransactionDB
func (es *ElasticsearchDB) WriteTransaction(transaction *types.Transaction) error {
	var tx Transaction
	tx.From(transaction)

	req := esapi.IndexRequest{
		Index:      TransactionIndex,
		DocumentID: transaction.Hash.String(),
		Body:       esutil.NewJSONReader(tx),
		Refresh:    "true",
	}

	_, err := es.apiClient.DoRequest(req)
	return err
}

func (es *ElasticsearchDB) WriteTransactions(transactions []*types.Transaction) error {
	bi := es.apiClient.GetBulkHandler(TransactionIndex)

	var (
		wg        sync.WaitGroup
		returnErr error
	)
	wg.Add(len(transactions))
	for _, transaction := range transactions {
		var tx Transaction
		tx.From(transaction)
		err := bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: transaction.Hash.String(),
				Body:       esutil.NewJSONReader(tx),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem) {
					wg.Done()
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem, err error) {
					returnErr = err
					wg.Done()
				},
			})
		if err != nil {
			return err
		}
	}

	wg.Wait()
	return returnErr
}

func (es *ElasticsearchDB) ReadTransaction(hash common.Hash) (*types.Transaction, error) {
	fetchReq := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: hash.String(),
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return nil, err
	}

	var transactionResult TransactionQueryResult
	err = json.Unmarshal(body, &transactionResult)
	if err != nil {
		return nil, err
	}
	return transactionResult.Source.To(), nil
}

// IndexDB
func (es *ElasticsearchDB) IndexBlock(addresses []common.Address, block *types.Block) error {
	// filter out registered and unfiltered address only
	filteredAddresses := map[common.Address]bool{}
	for _, address := range addresses {
		lastFiltered, _ := es.GetLastFiltered(address)
		if lastFiltered < block.Number {
			filteredAddresses[address] = true
			//log.Printf("Index registered address %v at block %v.\n", address.Hex(), block.Number)
		}
	}

	// index transactions and events
	for _, txHash := range block.Transactions {
		transaction, _ := es.ReadTransaction(txHash)
		err := es.indexTransaction(filteredAddresses, transaction)
		if err != nil {
			return err
		}
	}

	return es.updateAllLastFiltered(filteredAddresses, block.Number)
}

func (es *ElasticsearchDB) IndexStorage(rawStorage map[common.Address]*state.DumpAccount, blockNumber uint64) error {
	biState := es.apiClient.GetBulkHandler(StateIndex)
	biStorage := es.apiClient.GetBulkHandler(StorageIndex)

	for address, dumpAccount := range rawStorage {
		stateObj := State{
			Address:     address,
			BlockNumber: blockNumber,
			StorageRoot: common.HexToHash(dumpAccount.Root),
		}
		storageMap := Storage{
			StorageRoot: common.HexToHash(dumpAccount.Root),
			StorageMap:  dumpAccount.Storage,
		}

		biState.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: address.String() + "-" + strconv.FormatUint(blockNumber, 10),
				Body:       esutil.NewJSONReader(stateObj),
			},
		)
		biStorage.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: "0x" + dumpAccount.Root,
				Body:       esutil.NewJSONReader(storageMap),
			},
		)
	}
	// TODO: must make sure bulk update is successful and also not blocking to slow down...
	return nil
}

func (es *ElasticsearchDB) GetContractCreationTransaction(address common.Address) (common.Hash, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return common.Hash{}, err
	}
	return contract.CreationTransaction, nil
}

func (es *ElasticsearchDB) GetAllTransactionsToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	queryString := fmt.Sprintf(QueryByToAddressWithOptionsTemplate(options), address.String())

	from := options.PageSize * options.PageNumber
	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}
	req := esapi.SearchRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}
	results, err := es.doSearchRequest(req)
	if err != nil {
		return nil, err
	}

	converted := make([]common.Hash, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		addr := result.Source["hash"].(string)
		converted[i] = common.HexToHash(addr)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetTransactionsToAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	queryString := fmt.Sprintf(QueryByToAddressWithOptionsTemplate(options), address.String())

	req := esapi.CountRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(queryString),
	}
	results, err := es.doCountRequest(req)
	if err != nil {
		return 0, err
	}
	return results.Count, nil
}

func (es *ElasticsearchDB) GetAllTransactionsInternalToAddress(address common.Address, options *types.QueryOptions) ([]common.Hash, error) {
	queryString := fmt.Sprintf(QueryInternalTransactionsWithOptionsTemplate(options), address.String())

	from := options.PageSize * options.PageNumber
	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}
	req := esapi.SearchRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}
	results, err := es.doSearchRequest(req)
	if err != nil {
		return nil, err
	}

	converted := make([]common.Hash, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		addr := result.Source["hash"].(string)
		converted[i] = common.HexToHash(addr)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetTransactionsInternalToAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	queryString := fmt.Sprintf(QueryInternalTransactionsWithOptionsTemplate(options), address.String())

	req := esapi.CountRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(queryString),
	}
	results, err := es.doCountRequest(req)
	if err != nil {
		return 0, err
	}
	return results.Count, nil
}

func (es *ElasticsearchDB) GetAllEventsFromAddress(address common.Address, options *types.QueryOptions) ([]*types.Event, error) {
	queryString := fmt.Sprintf(QueryByAddressWithOptionsTemplate(options), address.String())

	from := options.PageSize * options.PageNumber
	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}
	req := esapi.SearchRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"blockNumber:desc", "logIndex:asc"},
	}
	results, err := es.doSearchRequest(req)
	if err != nil {
		return nil, err
	}

	convertedList := make([]*types.Event, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		marshalled, _ := json.Marshal(result.Source)
		var event Event
		json.Unmarshal(marshalled, &event)

		convertedList[i] = event.To()
	}

	return convertedList, nil
}

func (es *ElasticsearchDB) GetEventsFromAddressTotal(address common.Address, options *types.QueryOptions) (uint64, error) {
	queryString := fmt.Sprintf(QueryByAddressWithOptionsTemplate(options), address.String())

	req := esapi.CountRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(queryString),
	}
	results, err := es.doCountRequest(req)
	if err != nil {
		return 0, err
	}
	return results.Count, nil
}

func (es *ElasticsearchDB) GetStorage(address common.Address, blockNumber uint64) (map[common.Hash]string, error) {
	fetchReq := esapi.GetRequest{
		Index:      StateIndex,
		DocumentID: address.String() + "-" + strconv.FormatUint(blockNumber, 10),
	}
	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	if err == ErrNotFound {
		return nil, nil
	}
	var stateResult StateQueryResult
	json.Unmarshal(body, &stateResult)

	storageFetchReq := esapi.GetRequest{
		Index:      StorageIndex,
		DocumentID: stateResult.Source.StorageRoot.String(),
	}
	body, err = es.apiClient.DoRequest(storageFetchReq)
	if err != nil && err != ErrNotFound {
		return nil, err
	}
	if err == ErrNotFound {
		return nil, nil
	}
	var storageResult StorageQueryResult
	json.Unmarshal(body, &storageResult)

	return storageResult.Source.StorageMap, nil
}

func (es *ElasticsearchDB) GetLastFiltered(address common.Address) (uint64, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return 0, err
	}
	return contract.LastFiltered, nil
}

// Internal functions

func (es *ElasticsearchDB) checkIsInitialized() (bool, error) {
	fetchReq := esapi.CatIndicesRequest{
		Index: []string{MetaIndex, ContractIndex, BlockIndex, StorageIndex, TransactionIndex, EventIndex},
	}

	if _, err := es.apiClient.DoRequest(fetchReq); err != nil {
		if err == ErrIndexNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (es *ElasticsearchDB) getContractByAddress(address common.Address) (*Contract, error) {
	fetchReq := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: address.String(),
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return nil, err
	}

	var contract ContractQueryResult
	json.Unmarshal(body, &contract)
	return &contract.Source, nil
}

func (es *ElasticsearchDB) indexTransaction(filteredAddresses map[common.Address]bool, tx *types.Transaction) error {
	// Compare the address with tx.To and tx.CreatedContract to check if the transaction is related.
	if filteredAddresses[tx.CreatedContract] {
		if err := es.updateCreatedTx(tx.CreatedContract, tx.Hash); err != nil {
			return err
		}
		log.Printf("Index contract creation tx %v of registered address %v.\n", tx.Hash.Hex(), tx.CreatedContract.Hex())
	}
	for _, internalCall := range tx.InternalCalls {
		if internalCall.Type == "CREATE" || internalCall.Type == "CREATE2" {
			if err := es.updateCreatedTx(internalCall.To, tx.Hash); err != nil {
				return err
			}
			log.Printf("Index contract creation tx %v of registered address %v.\n", tx.Hash.Hex(), internalCall.To.Hex())
		}
	}

	// Index events emitted by the given address
	pendingIndexEvents := []*types.Event{}
	for _, event := range tx.Events {
		if filteredAddresses[event.Address] {
			pendingIndexEvents = append(pendingIndexEvents, event)
		}
	}
	return es.createEvents(pendingIndexEvents)
}

func (es *ElasticsearchDB) updateCreatedTx(address common.Address, creationTxHash common.Hash) error {
	return es.updateContract(address, "creationTx", creationTxHash.String())
}

func (es *ElasticsearchDB) updateAllLastFiltered(addresses map[common.Address]bool, lastFiltered uint64) error {
	bi := es.apiClient.GetBulkHandler(ContractIndex)

	for address := range addresses {
		bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "update",
				DocumentID: address.String(),
				Body:       strings.NewReader(fmt.Sprintf(`{"doc":{"lastFiltered":%d}}`, lastFiltered)),
			},
		)
	}
	return nil
}

func (es *ElasticsearchDB) updateContract(address common.Address, property string, value interface{}) error {
	//check contract exists before updating
	_, err := es.getContractByAddress(address)
	if err != nil {
		return err
	}

	query := map[string]interface{}{
		"doc": map[string]interface{}{
			property: value,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: address.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	_, err = es.apiClient.DoRequest(updateRequest)
	return err
}

func (es *ElasticsearchDB) createEvents(events []*types.Event) error {
	bi := es.apiClient.GetBulkHandler(EventIndex)

	for _, event := range events {
		var e Event
		e.From(event)
		bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: strconv.FormatUint(event.BlockNumber, 10) + "-" + strconv.FormatUint(event.Index, 10),
				Body:       esutil.NewJSONReader(e),
			},
		)
	}
	// TODO: must make sure bulk update is successful and also not blocking to slow down...
	return nil
}

func (es *ElasticsearchDB) Stop() {
	es.apiClient.CloseIndexers()
}

func (es *ElasticsearchDB) doSearchRequest(req esapi.SearchRequest) (*SearchQueryResult, error) {
	body, err := es.apiClient.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var ret SearchQueryResult
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (es *ElasticsearchDB) doCountRequest(req esapi.CountRequest) (*CountQueryResult, error) {
	body, err := es.apiClient.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var ret CountQueryResult
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (es *ElasticsearchDB) updateLastPersisted(startingBlockNumber uint64) error {
	last, err := es.GetLastPersistedBlockNumber()
	if err != nil {
		return err
	}

	blockNumber := startingBlockNumber
	if blockNumber == last+1 {
		for {
			if block, _ := es.ReadBlock(blockNumber + 1); block != nil {
				blockNumber++
			} else {
				break
			}
		}
		req := esapi.IndexRequest{
			Index:      MetaIndex,
			DocumentID: "lastPersisted",
			Body:       strings.NewReader(fmt.Sprintf(`{"lastPersisted": %d}`, blockNumber)),
			Refresh:    "true",
		}
		_, err := es.apiClient.DoRequest(req)
		return err
	}
	return nil
}
