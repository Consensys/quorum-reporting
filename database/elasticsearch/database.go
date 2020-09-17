package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type ElasticsearchDB struct {
	apiClient APIClient
	deleter   DeletionCoordinator

	deleteMux   sync.Mutex
	deleteQueue map[types.Address]*sync.WaitGroup
}

func New(client APIClient) (*ElasticsearchDB, error) {
	return NewWithDeps(client, NewDefaultDeletionCoordinator(client))
}

func NewWithDeps(client APIClient, dataDeleter DeletionCoordinator) (*ElasticsearchDB, error) {
	db := &ElasticsearchDB{
		apiClient:   client,
		deleter:     dataDeleter,
		deleteQueue: make(map[types.Address]*sync.WaitGroup),
	}

	initialized, err := db.checkIsInitialized()
	if err != nil {
		log.Error("Error communicating with ElasticSearch", "err", err)
		log.Error("Please check all ElasticSearch settings, including certificates, URL and username/password.")
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
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: TemplateIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: StorageIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: EventIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: MetaIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: ERC20TokenIndex})
	es.apiClient.DoRequest(esapi.IndicesCreateRequest{Index: ERC721TokenIndex})

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
func (es *ElasticsearchDB) AddAddresses(addresses []types.Address) error {
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
				TemplateName:        address.String(),
				CreationTransaction: "",
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
		TemplateName:        addresses[0].String(),
		CreationTransaction: "",
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

func (es *ElasticsearchDB) AddAddressFrom(address types.Address, from uint64) error {
	contract := Contract{
		Address:             address,
		TemplateName:        address.String(),
		CreationTransaction: "",
		LastFiltered:        from - 1,
	}

	req := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: address.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
		OpType:     "create", //This will only create if the contract does not exist
	}
	_, err := es.apiClient.DoRequest(req)
	return err
}

func (es *ElasticsearchDB) DeleteAddress(address types.Address) error {
	var wg sync.WaitGroup
	wg.Add(1)
	es.deleteMux.Lock()
	es.deleteQueue[address] = &wg
	es.deleteMux.Unlock()
	wg.Wait()
	return nil
}

func (es *ElasticsearchDB) GetAddresses() ([]types.Address, error) {
	results, err := es.apiClient.ScrollAllResults(ContractIndex, QueryAllAddressesTemplate)
	if err != nil {
		return nil, errors.New("error fetching addresses: " + err.Error())
	}
	converted := make([]types.Address, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})
		addr := data["address"].(string)
		converted[i] = types.NewAddress(addr)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetContractTemplate(address types.Address) (string, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return "", err
	}
	return contract.TemplateName, nil
}

//TemplateDB
func (es *ElasticsearchDB) GetContractABI(address types.Address) (string, error) {

	contract, err := es.getContractByAddress(address)
	if err != nil && err != database.ErrNotFound {
		return "", err
	}

	if contract != nil {
		template, err := es.getTemplateByName(contract.TemplateName)
		if err != nil && err != database.ErrNotFound {
			return "", err
		}
		if template != nil {
			return template.ABI, nil
		}
	}
	return "", nil
}

func (es *ElasticsearchDB) GetStorageLayout(address types.Address) (string, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil && err != database.ErrNotFound {
		return "", err
	}
	if contract != nil {
		template, err := es.getTemplateByName(contract.TemplateName)
		if err != nil && err != database.ErrNotFound {
			return "", err
		}
		if template != nil {
			return template.StorageABI, nil
		}
	}
	return "", nil
}

func (es *ElasticsearchDB) AddTemplate(name string, abi string, layout string) error {
	template := Template{
		TemplateName: name,
		ABI:          abi,
		StorageABI:   layout,
	}

	req := esapi.IndexRequest{
		Index:      TemplateIndex,
		DocumentID: name,
		Body:       esutil.NewJSONReader(template),
		Refresh:    "true",
	}
	_, err := es.apiClient.DoRequest(req)
	return err
}

func (es *ElasticsearchDB) AssignTemplate(address types.Address, name string) error {
	return es.updateContract(address, "templateName", name)
}

func (es *ElasticsearchDB) GetTemplates() ([]string, error) {
	results, err := es.apiClient.ScrollAllResults(TemplateIndex, QueryAllTemplateNamesTemplate)
	if err != nil {
		return nil, errors.New("error fetching templates: " + err.Error())
	}
	converted := make([]string, len(results))
	for i, result := range results {
		data := result.(map[string]interface{})["_source"].(map[string]interface{})
		converted[i] = data["templateName"].(string)
	}
	return converted, nil
}

func (es *ElasticsearchDB) GetTemplateDetails(templateName string) (*types.Template, error) {
	template, err := es.getTemplateByName(templateName)
	if err != nil {
		return nil, err
	}
	return &types.Template{
		TemplateName:  templateName,
		ABI:           template.ABI,
		StorageLayout: template.StorageABI,
	}, nil
}

// BlockDB
func (es *ElasticsearchDB) WriteBlock(block *types.Block) error {
	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: strconv.FormatUint(block.Number, 10),
		Body:       esutil.NewJSONReader(block),
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
		_ = bi.Add(context.Background(), esutil.BulkIndexerItem{
			Action:     "create",
			DocumentID: strconv.FormatUint(block.Number, 10),
			Body:       esutil.NewJSONReader(block),
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem) {
				wg.Done()
			},
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem, err error) {
				returnErr = err
				wg.Done()
			},
		})
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
	if err = json.Unmarshal(body, &blockResult); err != nil {
		return nil, err
	}
	return blockResult.Source, nil
}

func (es *ElasticsearchDB) GetLastPersistedBlockNumber() (uint64, error) {
	// At this point, we know no data insertions are happening so we can safely
	// delete data
	es.deleteMux.Lock()
	for address, wg := range es.deleteQueue {
		if err := es.deleter.Delete(address); err != nil {
			log.Warn("Error when servicing deletion request", "address", address.String(), "err", err)
			return 0, errors.New("error performing deletion of contract " + address.String())
		}
		delete(es.deleteQueue, address)
		wg.Done()
	}
	es.deleteMux.Unlock()

	fetchReq := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return 0, err
	}

	var lastPersisted LastPersistedResult
	if err = json.Unmarshal(body, &lastPersisted); err != nil {
		return 0, err
	}
	return lastPersisted.Source.LastPersisted, nil
}

// TransactionDB
func (es *ElasticsearchDB) WriteTransaction(transaction *types.Transaction) error {
	req := esapi.IndexRequest{
		Index:      TransactionIndex,
		DocumentID: transaction.Hash.String(),
		Body:       esutil.NewJSONReader(transaction),
		Refresh:    "true",
	}

	_, err := es.apiClient.DoRequest(req)
	return err
}

func (es *ElasticsearchDB) WriteTransactions(transactions []*types.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	if len(transactions) == 1 {
		return es.WriteTransaction(transactions[0])
	}

	bi := es.apiClient.GetBulkHandler(TransactionIndex)

	var (
		wg        sync.WaitGroup
		returnErr error
	)
	wg.Add(len(transactions))
	for _, transaction := range transactions {
		_ = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: transaction.Hash.String(),
				Body:       esutil.NewJSONReader(transaction),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem) {
					wg.Done()
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, item2 esutil.BulkIndexerResponseItem, err error) {
					returnErr = err
					wg.Done()
				},
			})
	}

	wg.Wait()
	return returnErr
}

func (es *ElasticsearchDB) ReadTransaction(hash types.Hash) (*types.Transaction, error) {
	fetchReq := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: hash.String(),
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return nil, err
	}

	var transactionResult TransactionQueryResult
	if err = json.Unmarshal(body, &transactionResult); err != nil {
		return nil, err
	}
	return transactionResult.Source, nil
}

// IndexDB

func (es *ElasticsearchDB) IndexBlocks(addresses []types.Address, blocks []*types.BlockWithTransactions) error {
	indexer := NewBlockIndexer(addresses, blocks, es)
	if err := indexer.Index(); err != nil {
		return err
	}
	return es.updateAllLastFiltered(addresses, blocks[len(blocks)-1].Number)
}

func (es *ElasticsearchDB) IndexStorage(rawStorage map[types.Address]*types.AccountState, blockNumber uint64) error {
	biStorage := es.apiClient.GetBulkHandler(StorageIndex)

	var (
		wg        sync.WaitGroup
		returnErr error
	)
	for address, dumpAccount := range rawStorage {
		wg.Add(1)
		converted := make([]StorageEntry, 0, len(dumpAccount.Storage))
		for slot, val := range dumpAccount.Storage {
			converted = append(converted, StorageEntry{slot, val})
		}
		storageMap := Storage{
			Contract:    address,
			BlockNumber: blockNumber,
			StorageRoot: dumpAccount.Root,
			StorageMap:  converted,
		}

		_ = biStorage.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: address.String() + "-" + strconv.FormatUint(blockNumber, 10),
				Body:       esutil.NewJSONReader(storageMap),
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

func (es *ElasticsearchDB) SetContractCreationTransaction(creationTxns map[types.Hash][]types.Address) error {
	for txHash, addresses := range creationTxns {
		for _, createdAddress := range addresses {
			if err := es.updateContract(createdAddress, "creationTx", txHash.String()); err != nil {
				log.Error("Failed to index contract creation tx", "tx", txHash, "contract", createdAddress, "err", err)
				return err
			}
			log.Info("Indexed contract creation tx for address", "tx", txHash, "contract", createdAddress)
		}
	}
	return nil
}

func (es *ElasticsearchDB) GetContractCreationTransaction(address types.Address) (types.Hash, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return "", err
	}
	return contract.CreationTransaction, nil
}

func (es *ElasticsearchDB) GetAllTransactionsToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
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

	converted := make([]types.Hash, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		hsh := result.Source["hash"].(string)
		converted[i] = types.NewHash(hsh)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetTransactionsToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
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

func (es *ElasticsearchDB) GetAllTransactionsInternalToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
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

	converted := make([]types.Hash, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		hsh := result.Source["hash"].(string)
		converted[i] = types.NewHash(hsh)
	}

	return converted, nil
}

func (es *ElasticsearchDB) GetTransactionsInternalToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
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

func (es *ElasticsearchDB) GetAllEventsFromAddress(address types.Address, options *types.QueryOptions) ([]*types.Event, error) {
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
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}
	results, err := es.doSearchRequest(req)
	if err != nil {
		return nil, err
	}

	convertedList := make([]*types.Event, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		marshalled, _ := json.Marshal(result.Source)
		var event types.Event
		if err = json.Unmarshal(marshalled, &event); err != nil {
			return nil, err
		}
		convertedList[i] = &event
	}

	return convertedList, nil
}

func (es *ElasticsearchDB) GetStorageWithOptions(address types.Address, options *types.PageOptions) ([]*types.StorageResult, error) {
	return es.getStorageWithOptionsAndDirection(address, options, false)
}

func (es *ElasticsearchDB) GetEventsFromAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
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

func (es *ElasticsearchDB) GetStorageTotal(address types.Address, options *types.PageOptions) (uint64, error) {
	queryString := fmt.Sprintf(QueryByAddressWithBlockRangeOptionsTemplate(options), address.String())

	req := esapi.CountRequest{
		Index: []string{StorageIndex},
		Body:  strings.NewReader(queryString),
	}
	results, err := es.doCountRequest(req)
	if err != nil {
		return 0, err
	}
	return results.Count, nil
}

func (es *ElasticsearchDB) GetStorage(address types.Address, blockNumber uint64) (*types.StorageResult, error) {
	size := 1
	searchReq := esapi.SearchRequest{
		Index: []string{StorageIndex},
		Body:  strings.NewReader(fmt.Sprintf(QueryMatchContract, address.String(), blockNumber)),
		Size:  &size,
	}
	result, err := es.doSearchRequest(searchReq)
	if err != nil {
		return nil, err
	}

	// There are no results, probably trying to query before the contract exists
	if len(result.Hits.Hits) == 0 {
		return &types.StorageResult{
			Storage:     make(map[types.Hash]string),
			StorageRoot: types.NewHash(""),
			BlockNumber: blockNumber,
		}, nil
	}

	// There is a single result, map it to the proper output object
	marshalled, _ := json.Marshal(result.Hits.Hits[0])
	var storageResult StorageQueryResult
	if err := json.Unmarshal(marshalled, &storageResult); err != nil {
		return nil, err
	}
	converted := make(map[types.Hash]string)
	for _, storageEntry := range storageResult.Source.StorageMap {
		converted[storageEntry.Key] = storageEntry.Value
	}
	return &types.StorageResult{
		Storage:     converted,
		StorageRoot: storageResult.Source.StorageRoot,
		BlockNumber: blockNumber,
	}, nil
}

func (es *ElasticsearchDB) GetStorageRanges(contract types.Address, options *types.PageOptions) ([]types.RangeResult, error) {
	end := options.EndBlockNumber
	if big.NewInt(-1).Cmp(end) == 0 {
		endUint64, err := es.GetLastFiltered(contract)
		if err != nil {
			return nil, err
		}
		end = new(big.Int).SetUint64(endUint64)
	}

	startUint64 := options.BeginBlockNumber.Uint64()
	endUint64 := end.Uint64()

	var results []types.RangeResult
	for endUint64 >= startUint64 {
		options := &types.PageOptions{
			BeginBlockNumber: options.BeginBlockNumber,
			EndBlockNumber:   new(big.Int).SetUint64(endUint64),
			PageSize:         1000,
		}
		options.SetDefaults()
		res, err := es.getStorageWithOptionsAndDirection(contract, options, false)
		if err != nil {
			return nil, err
		}

		foundEndBlockNumber := startUint64
		if len(res) != 0 {
			foundEndBlockNumber = res[len(res)-1].BlockNumber
		}
		rangeRes := types.RangeResult{
			Start:       foundEndBlockNumber,
			End:         endUint64,
			ResultCount: len(res),
		}
		results = append(results, rangeRes)

		if foundEndBlockNumber == startUint64 {
			break
		}
		endUint64 = foundEndBlockNumber - 1
	}

	//TODO: remove this limitation
	//This is caused by the fact that indexing doesn't happen for block 0, but queries can start at block 0
	if options.BeginBlockNumber.Uint64() == 0 && len(results) > 1 {
		//more than 1 result && the penultimate result wasn't full, merge the last 2
		results = results[:len(results)-1]
		results[len(results)-1].Start = 0
	}

	return results, nil
}

func (es *ElasticsearchDB) GetLastFiltered(address types.Address) (uint64, error) {
	contract, err := es.getContractByAddress(address)
	if err != nil {
		return 0, err
	}
	return contract.LastFiltered, nil
}

// Internal functions

func (es *ElasticsearchDB) checkIsInitialized() (bool, error) {
	fetchReq := esapi.CatIndicesRequest{
		Index: []string{MetaIndex, ContractIndex, BlockIndex, StorageIndex, TransactionIndex, EventIndex, ERC20TokenIndex, ERC721TokenIndex},
	}

	if _, err := es.apiClient.DoRequest(fetchReq); err != nil {
		if err == ErrIndexNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (es *ElasticsearchDB) getContractByAddress(address types.Address) (*Contract, error) {
	fetchReq := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: address.String(),
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return nil, err
	}

	var contract ContractQueryResult
	if err = json.Unmarshal(body, &contract); err != nil {
		return nil, err
	}
	return &contract.Source, nil
}

func (es *ElasticsearchDB) getTemplateByName(name string) (*Template, error) {
	fetchReq := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: name,
	}

	body, err := es.apiClient.DoRequest(fetchReq)
	if err != nil {
		return nil, err
	}

	var template TemplateQueryResult
	if err = json.Unmarshal(body, &template); err != nil {
		return nil, err
	}
	return &template.Source, nil
}

func (es *ElasticsearchDB) updateAllLastFiltered(addresses []types.Address, lastFiltered uint64) error {
	bi := es.apiClient.GetBulkHandler(ContractIndex)

	for _, address := range addresses {
		_ = bi.Add(
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

func (es *ElasticsearchDB) updateContract(address types.Address, property string, value interface{}) error {
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

	var (
		wg        sync.WaitGroup
		returnErr error
	)
	for _, event := range events {
		wg.Add(1)
		_ = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "create",
				DocumentID: strconv.FormatUint(event.BlockNumber, 10) + "-" + strconv.FormatUint(event.Index, 10),
				Body:       esutil.NewJSONReader(event),
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

func (es *ElasticsearchDB) Stop() {
	es.apiClient.CloseIndexers()
	log.Info("Elasticsearch indexers closed")
}

func (es *ElasticsearchDB) doSearchRequest(req esapi.SearchRequest) (*SearchQueryResult, error) {
	body, err := es.apiClient.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var ret SearchQueryResult
	if err = json.Unmarshal(body, &ret); err != nil {
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
	if err = json.Unmarshal(body, &ret); err != nil {
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

func (es *ElasticsearchDB) getStorageWithOptionsAndDirection(address types.Address, options *types.PageOptions, ascending bool) ([]*types.StorageResult, error) {
	queryString := fmt.Sprintf(QueryByAddressWithBlockRangeOptionsTemplate(options), address.String())
	from := options.PageSize * options.PageNumber

	direction := "desc"
	if ascending {
		direction = "asc"
	}

	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}
	req := esapi.SearchRequest{
		Index: []string{StorageIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"blockNumber:" + direction},
	}

	results, err := es.doSearchRequest(req)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	convertedList := make([]*types.StorageResult, len(results.Hits.Hits))
	for i, result := range results.Hits.Hits {
		marshalled, err := json.Marshal(result)
		var storageResult StorageQueryResult
		if err = json.Unmarshal(marshalled, &storageResult); err != nil {
			return nil, err
		}
		converted := make(map[types.Hash]string)
		for _, storageEntry := range storageResult.Source.StorageMap {
			converted[storageEntry.Key] = storageEntry.Value
		}
		convertedList[i] = &types.StorageResult{
			Storage:     converted,
			StorageRoot: storageResult.Source.StorageRoot,
			BlockNumber: storageResult.Source.BlockNumber}
	}

	return convertedList, nil
}
