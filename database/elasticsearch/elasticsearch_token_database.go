package elasticsearch

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/mitchellh/mapstructure"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// Token DB
func (es *ElasticsearchDB) RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
	//find old entry
	existingTokenEntry, errExisting := es.GetERC20EntryAtBlock(contract, holder, block-1)
	if errExisting != nil && errExisting != database.ErrNotFound {
		return errExisting
	}

	//add new entry
	tokenInfo := ERC20TokenHolder{
		Contract:    contract,
		Holder:      holder,
		BlockNumber: block,
		Amount:      amount.String(),
	}

	req := esapi.IndexRequest{
		Index:      ERC20TokenIndex,
		DocumentID: fmt.Sprintf("%s-%s-%d", contract.String(), holder.String(), block),
		Body:       esutil.NewJSONReader(tokenInfo),
		Refresh:    "true",
		OpType:     "create",
	}

	if _, err := es.apiClient.DoRequest(req); err != nil {
		return err
	}

	/////

	if errExisting == database.ErrNotFound {
		return nil
	}

	//update the older entry
	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"heldUntil": block - 1,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      ERC20TokenIndex,
		DocumentID: fmt.Sprintf("%s-%s-%d", contract.String(), holder.String(), existingTokenEntry.BlockNumber),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	_, err := es.apiClient.DoRequest(updateRequest)
	return err
}

func (es *ElasticsearchDB) GetERC20EntryAtBlock(contract types.Address, holder types.Address, block uint64) (ERC20TokenHolder, error) {
	queryString := fmt.Sprintf(QueryERC20TokenBalanceAtBlock(), contract.String(), holder.String(), block)

	size := 1
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(queryString),
		Size:  &size,
	}
	results, err := es.doSearchRequest(req)
	if err != nil {
		return ERC20TokenHolder{}, err
	}

	if len(results.Hits.Hits) == 0 {
		return ERC20TokenHolder{}, database.ErrNotFound
	}

	var tokenResult ERC20TokenHolder
	err = mapstructure.Decode(results.Hits.Hits[0].Source, &tokenResult)
	return tokenResult, err
}

func (es *ElasticsearchDB) GetERC20Balance(contract types.Address, holder types.Address, options *types.TokenQueryOptions) (map[uint64]*big.Int, error) {
	queryString := fmt.Sprintf(QueryTokenBalanceAtBlockRange(options), contract.String(), holder.String())

	from := options.PageSize * options.PageNumber
	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"blockNumber:desc"},
	}
	results, err := es.doSearchRequest(req)
	if err != nil {
		return nil, err
	}

	balanceMap := make(map[uint64]*big.Int)
	for _, result := range results.Hits.Hits {
		blockNumber := uint64(result.Source["blockNumber"].(float64))
		tokenAmount, success := new(big.Int).SetString(result.Source["amount"].(string), 10)
		if !success {
			return nil, errors.New("could not parse token value")
		}
		balanceMap[blockNumber] = tokenAmount
		if blockNumber < options.BeginBlockNumber.Uint64() {
			balanceMap[options.BeginBlockNumber.Uint64()] = tokenAmount
		}
	}

	return balanceMap, nil
}

func (es *ElasticsearchDB) GetAllTokenHolders(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	if options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}

	afterQuery := ""
	if options.After != "" {
		afterQuery = fmt.Sprintf(`"after": { "holder": "%s"},`, options.After)
	}

	formattedQuery := fmt.Sprintf(QueryERC20TokenHoldersAtBlock(), contract.String(), block, block, options.PageSize, afterQuery)

	searchReq := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(formattedQuery),
	}

	results, err := es.doSearchRequest(searchReq)
	if err != nil {
		return nil, err
	}

	var aggResult ERC721HolderAggregateResult
	rawAggResult := results.Aggregations.Results
	if err := mapstructure.Decode(rawAggResult, &aggResult); err != nil {
		return nil, err
	}

	convertedResults := make([]types.Address, 0, len(aggResult.Buckets))
	for _, result := range aggResult.Buckets {
		convertedResults = append(convertedResults, types.NewAddress(result.Key.Holder))
	}
	return convertedResults, nil
}

func (es *ElasticsearchDB) RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error {
	//find old entry
	existingTokenEntry, errExisting := es.ERC721TokenByTokenID(contract, block-1, tokenId)
	if errExisting != nil && errExisting != database.ErrNotFound {
		return errExisting
	}

	paddedTokenId := fmt.Sprintf("%085d", tokenId)
	first, _ := strconv.ParseUint(paddedTokenId[0:17], 10, 64)
	second, _ := strconv.ParseUint(paddedTokenId[17:34], 10, 64)
	third, _ := strconv.ParseUint(paddedTokenId[34:51], 10, 64)
	fourth, _ := strconv.ParseUint(paddedTokenId[51:68], 10, 64)
	fifth, _ := strconv.ParseUint(paddedTokenId[68:85], 10, 64)

	//add new entry
	tokenHolderInfo := SortableERC721Token{
		types.ERC721Token{
			Contract:  contract,
			Holder:    holder,
			Token:     tokenId.String(),
			HeldFrom:  block,
			HeldUntil: nil,
		},
		first, second, third, fourth, fifth,
	}

	req := esapi.IndexRequest{
		Index:      ERC721TokenIndex,
		DocumentID: fmt.Sprintf("%s-%s-%d", contract.String(), tokenId.String(), block),
		Body:       esutil.NewJSONReader(tokenHolderInfo),
		Refresh:    "true",
		OpType:     "create", //This will only create if the contract does not exist
	}

	if _, err := es.apiClient.DoRequest(req); err != nil {
		return err
	}

	/////

	if errExisting == database.ErrNotFound {
		return nil
	}

	//update the older entry
	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"heldUntil": block - 1,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      ERC721TokenIndex,
		DocumentID: fmt.Sprintf("%s-%s-%d", contract.String(), tokenId.String(), existingTokenEntry.HeldFrom),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	_, err := es.apiClient.DoRequest(updateRequest)
	return err
}

func (es *ElasticsearchDB) ERC721TokenByTokenID(contract types.Address, block uint64, tokenId *big.Int) (types.ERC721Token, error) {
	formattedQuery := fmt.Sprintf(QueryERC721TokenAtBlock(), contract.String(), tokenId.String(), block)

	pageSize := 1
	searchReq := esapi.SearchRequest{
		Index: []string{ERC721TokenIndex},
		Body:  strings.NewReader(formattedQuery),
		Size:  &pageSize,
	}

	results, err := es.doSearchRequest(searchReq)
	if err != nil {
		return types.ERC721Token{}, err
	}

	if len(results.Hits.Hits) == 0 {
		return types.ERC721Token{}, database.ErrNotFound
	}

	var tokenResult types.ERC721Token
	err = mapstructure.Decode(results.Hits.Hits[0].Source, &tokenResult)
	return tokenResult, err
}

func (es *ElasticsearchDB) ERC721TokensForAccountAtBlock(contract types.Address, holder types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	formattedQuery := fmt.Sprintf(QueryERC721HolderAtBlock(options.BeginTokenId, options.EndTokenId), contract.String(), holder.String(), block, block)

	from := options.PageSize * options.PageNumber
	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}

	searchReq := esapi.SearchRequest{
		Index: []string{ERC721TokenIndex},
		Body:  strings.NewReader(formattedQuery),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"first:desc", "second:desc", "third:desc", "fourth:desc", "fifth:desc"},
	}

	results, err := es.doSearchRequest(searchReq)
	if err != nil {
		return nil, err
	}

	convertedResults := make([]types.ERC721Token, 0, len(results.Hits.Hits))
	for _, result := range results.Hits.Hits {
		var tokenResult types.ERC721Token
		if err := mapstructure.Decode(result.Source, &tokenResult); err != nil {
			return nil, err
		}
		convertedResults = append(convertedResults, tokenResult)
	}
	return convertedResults, nil
}

func (es *ElasticsearchDB) AllERC721TokensAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	formattedQuery := fmt.Sprintf(QueryERC721AllTokensAtBlock(options.BeginTokenId, options.EndTokenId), contract.String(), block, block)

	from := options.PageSize * options.PageNumber
	if from+options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}

	searchReq := esapi.SearchRequest{
		Index: []string{ERC721TokenIndex},
		Body:  strings.NewReader(formattedQuery),
		From:  &from,
		Size:  &options.PageSize,
		Sort:  []string{"first:desc", "second:desc", "third:desc", "fourth:desc", "fifth:desc"},
	}

	results, err := es.doSearchRequest(searchReq)
	if err != nil {
		return nil, err
	}

	convertedResults := make([]types.ERC721Token, 0, len(results.Hits.Hits))
	for _, result := range results.Hits.Hits {
		var tokenResult types.ERC721Token
		if err := mapstructure.Decode(result.Source, &tokenResult); err != nil {
			return nil, err
		}
		convertedResults = append(convertedResults, tokenResult)
	}
	return convertedResults, nil
}

func (es *ElasticsearchDB) AllHoldersAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	if options.PageSize > 1000 {
		return nil, ErrPaginationLimitExceeded
	}

	afterQuery := ""
	if options.After != "" {
		afterQuery = fmt.Sprintf(`"after": { "holder": "%s"},`, options.After)
	}

	formattedQuery := fmt.Sprintf(QueryERC721AllHoldersAtBlock(), contract.String(), block, block, options.PageSize, afterQuery)

	searchReq := esapi.SearchRequest{
		Index: []string{ERC721TokenIndex},
		Body:  strings.NewReader(formattedQuery),
	}

	results, err := es.doSearchRequest(searchReq)
	if err != nil {
		return nil, err
	}

	var aggResult ERC721HolderAggregateResult
	rawAggResult := results.Aggregations.Results
	if err := mapstructure.Decode(rawAggResult, &aggResult); err != nil {
		return nil, err
	}

	convertedResults := make([]types.Address, 0, len(aggResult.Buckets))
	for _, result := range aggResult.Buckets {
		convertedResults = append(convertedResults, types.NewAddress(result.Key.Holder))
	}
	return convertedResults, nil
}
