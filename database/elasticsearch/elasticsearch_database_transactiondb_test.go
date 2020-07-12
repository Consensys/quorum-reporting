package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearch_mocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"quorumengineering/quorum-report/types"
)

//Tests

var testTransaction = types.Transaction{
	Hash:              common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
	Status:            true,
	BlockNumber:       1,
	BlockHash:         types.NewHash(""),
	Index:             0,
	Nonce:             4,
	From:              types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad"),
	To:                common.HexToAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"),
	Value:             10,
	Gas:               30,
	GasUsed:           20,
	CumulativeGasUsed: 40,
	CreatedContract:   common.HexToAddress("0x67bb49f7bd40b6a1226d77dc07fb38f03680c94f"),
	Data:              common.Hex2Bytes("0x4ae157f8a703379222a96b5c01ec83b11b0a0a579b4abc68a10a4c0e7d"),
	PrivateData:       common.Hex2Bytes("0x6060604052341561000f57600080fd5b60405160208061014983398101"),
	IsPrivate:         true,
	Timestamp:         1000,
	Events:            nil,
	InternalCalls:     nil,
}

func TestElasticsearchDB_WriteSingleTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
		Body:       esutil.NewJSONReader(&testTransaction),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)

	db, _ := New(mockedClient)
	err := db.WriteTransactions([]*types.Transaction{&testTransaction})
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_WriteTransactions_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)
	mockedBulkIndexer := elasticsearch_mocks.NewMockBulkIndexer(ctrl)

	req := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: testTransaction.Hash.String(),
		Body:       esutil.NewJSONReader(&testTransaction),
	}
	reqMatcher := NewBulkIndexerItemMatcher(req)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().GetBulkHandler(TransactionIndex).Return(mockedBulkIndexer)
	mockedBulkIndexer.EXPECT().Add(gomock.Any(), reqMatcher).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnFailure(context.Background(), req, esutil.BulkIndexerResponseItem{}, errors.New("test error"))
		})
	mockedBulkIndexer.EXPECT().Add(gomock.Any(), reqMatcher).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req, esutil.BulkIndexerResponseItem{})
		})

	db, _ := New(mockedClient)
	err := db.WriteTransactions([]*types.Transaction{&testTransaction, &testTransaction})
	assert.EqualError(t, err, "test error")
}

func TestElasticsearchDB_WriteTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)
	mockedBulkIndexer := elasticsearch_mocks.NewMockBulkIndexer(ctrl)

	req := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: testTransaction.Hash.String(),
		Body:       esutil.NewJSONReader(&testTransaction),
	}
	reqMatcher := NewBulkIndexerItemMatcher(req)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().GetBulkHandler(TransactionIndex).Return(mockedBulkIndexer)
	mockedBulkIndexer.EXPECT().Add(gomock.Any(), reqMatcher).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req, esutil.BulkIndexerResponseItem{})
		})
	mockedBulkIndexer.EXPECT().Add(gomock.Any(), reqMatcher).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req, esutil.BulkIndexerResponseItem{})
		})

	db, _ := New(mockedClient)
	err := db.WriteTransactions([]*types.Transaction{&testTransaction, &testTransaction})
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_ReadTransaction_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(req)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	tx, err := db.ReadTransaction(testTransaction.Hash)

	assert.Nil(t, tx, "unexpected transaction return value")
	assert.EqualError(t, err, "test error")
}

func TestElasticsearchDB_ReadTransaction_WithErrorUnmarshalling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(req)).Return([]byte("{invalid"), nil)

	db, _ := New(mockedClient)

	tx, err := db.ReadTransaction(testTransaction.Hash)

	assert.Nil(t, tx, "unexpected transaction return value")
	assert.EqualError(t, err, "invalid character 'i' looking for beginning of object key string")
}

func TestElasticsearchDB_ReadTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	asJson, _ := json.Marshal(testTransaction)

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(req)).
		Return([]byte(fmt.Sprintf(`{"_source": %s}`, asJson)), nil)

	db, _ := New(mockedClient)

	tx, err := db.ReadTransaction(testTransaction.Hash)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, tx, &testTransaction, "unexpected transaction returned")
}
