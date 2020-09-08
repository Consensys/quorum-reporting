package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearchmocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"quorumengineering/quorum-report/types"
)

//Tests

func TestAddSingleAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contract := Contract{
		Address:             addr,
		TemplateName:        addr.String(),
		CreationTransaction: "",
		LastFiltered:        0,
	}

	ex := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(ex)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	})

	db, _ := New(mockedClient)
	err := db.AddAddresses([]types.Address{addr})
	assert.Nil(t, err, "expected error to be nil")
}

func TestAddMultipleAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)
	mockedBulkIndexer := elasticsearchmocks.NewMockBulkIndexer(ctrl)

	addr1 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	addr2 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f35")

	contract1 := Contract{
		Address:             addr1,
		TemplateName:        addr1.String(),
		CreationTransaction: "",
		LastFiltered:        0,
	}
	req1 := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: addr1.String(),
		Body:       esutil.NewJSONReader(contract1),
	}
	contract2 := Contract{
		Address:             addr2,
		TemplateName:        addr2.String(),
		CreationTransaction: "",
		LastFiltered:        0,
	}
	req2 := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: addr2.String(),
		Body:       esutil.NewJSONReader(contract2),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().GetBulkHandler(ContractIndex).Return(mockedBulkIndexer)
	mockedBulkIndexer.EXPECT().
		Add(gomock.Any(), NewBulkIndexerItemMatcher(req1)).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req1, esutil.BulkIndexerResponseItem{})
		})
	mockedBulkIndexer.EXPECT().
		Add(gomock.Any(), NewBulkIndexerItemMatcher(req2)).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req2, esutil.BulkIndexerResponseItem{})
		})

	db, _ := New(mockedClient)

	err := db.AddAddresses([]types.Address{addr1, addr2})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddNoAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test

	db, _ := New(mockedClient)

	err := db.AddAddresses([]types.Address{})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddSingleAddressWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contract := Contract{
		Address:             addr,
		TemplateName:        addr.String(),
		CreationTransaction: "",
		LastFiltered:        0,
	}

	ex := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(ex)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	}).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.AddAddresses([]types.Address{addr})

	assert.EqualError(t, err, "test error", "expected test error")
}

func TestAddMultipleAddressWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)
	mockedBulkIndexer := elasticsearchmocks.NewMockBulkIndexer(ctrl)

	addr1 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	addr2 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f35")

	contract1 := Contract{
		Address:             addr1,
		TemplateName:        addr1.String(),
		CreationTransaction: "",
		LastFiltered:        0,
	}
	req1 := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: addr1.String(),
		Body:       esutil.NewJSONReader(contract1),
	}
	contract2 := Contract{
		Address:             addr2,
		TemplateName:        addr2.String(),
		CreationTransaction: "",
		LastFiltered:        0,
	}
	req2 := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: addr2.String(),
		Body:       esutil.NewJSONReader(contract2),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().GetBulkHandler(ContractIndex).Return(mockedBulkIndexer)
	mockedBulkIndexer.EXPECT().
		Add(gomock.Any(), NewBulkIndexerItemMatcher(req1)).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req1, esutil.BulkIndexerResponseItem{})
		})
	mockedBulkIndexer.EXPECT().
		Add(gomock.Any(), NewBulkIndexerItemMatcher(req2)).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnFailure(context.Background(), req2, esutil.BulkIndexerResponseItem{}, errors.New("test error"))
		}).Return(errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.AddAddresses([]types.Address{addr1, addr2})

	assert.EqualError(t, err, "test error", "expected test error")
}

func TestElasticsearchDB_DeleteAddress_Delegates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)
	mockedDeleter := elasticsearchmocks.NewMockDeletionCoordinator(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test

	db, _ := NewWithDeps(mockedClient, mockedDeleter)

	// simulate calling point of actually deleting address
	// in the live app, this is done by GetLastPersistedBlockNumber()
	go func() {
		for len(db.deleteQueue) == 0 {
		}
		db.deleteQueue[addr].Done()
	}()

	err := db.DeleteAddress(addr)
	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_GetAddresses_NoAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(ContractIndex, QueryAllAddressesTemplate).
		Return(make([]interface{}, 0, 0), nil)

	db, _ := New(mockedClient)
	allAddresses, err := db.GetAddresses()

	assert.Nil(t, err, "error was not nil")
	assert.Equal(t, 0, len(allAddresses), "addresses found when none expected: %s", allAddresses)
}

func TestElasticsearchDB_GetAddresses_SingleAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	createReturnValue := func(addr types.Address) interface{} {
		sampleReturnValue := `{"_source" : { "address": "%s"}}`
		withAddress := fmt.Sprintf(sampleReturnValue, addr.String())
		var asInterface map[string]interface{}
		_ = json.Unmarshal([]byte(withAddress), &asInterface)
		return asInterface
	}

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(ContractIndex, QueryAllAddressesTemplate).
		Return([]interface{}{createReturnValue(sampleAddress)}, nil)

	db, _ := New(mockedClient)
	allAddresses, err := db.GetAddresses()

	assert.Nil(t, err, "error was not nil")
	assert.Equal(t, 1, len(allAddresses), "wrong number of addresses found: %s", allAddresses)
	assert.Equal(t, allAddresses[0], sampleAddress, "unexpected address found: %s", allAddresses[0])
}

func TestElasticsearchDB_GetAddresses_MultipleAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleAddress1 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	sampleAddress2 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f35")
	createReturnValue := func(addr types.Address) interface{} {
		sampleReturnValue := `{"_source" : { "address": "%s"}}`
		withAddress := fmt.Sprintf(sampleReturnValue, addr.String())
		var asInterface map[string]interface{}
		_ = json.Unmarshal([]byte(withAddress), &asInterface)
		return asInterface
	}

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(ContractIndex, QueryAllAddressesTemplate).
		Return([]interface{}{createReturnValue(sampleAddress1), createReturnValue(sampleAddress2)}, nil)

	db, _ := New(mockedClient)
	allAddresses, err := db.GetAddresses()

	assert.Nil(t, err, "error was not nil")
	assert.Equal(t, 2, len(allAddresses), "wrong number of addresses found: %s", allAddresses)
	assert.Equal(t, allAddresses[0], sampleAddress1, "unexpected address found: %s", allAddresses[0])
	assert.Equal(t, allAddresses[1], sampleAddress2, "unexpected address found: %s", allAddresses[1])
}

func TestElasticsearchDB_GetAddresses_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(ContractIndex, QueryAllAddressesTemplate).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)
	allAddresses, err := db.GetAddresses()

	assert.Nil(t, allAddresses, "error was not nil")
	assert.EqualError(t, err, "error fetching addresses: test error", "wrong error message")
}
