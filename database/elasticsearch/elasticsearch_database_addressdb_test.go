package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

//Tests

func TestAddSingleAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contract := Contract{
		Address:             addr,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}

	ex := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(ex)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	})

	db, err := New(mockedClient)

	err = db.AddAddresses([]common.Address{addr})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddMultipleAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr1 := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	addr2 := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f35")

	contract1 := Contract{
		Address:             addr1,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}
	req1 := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addr1.String(),
		Body:       esutil.NewJSONReader(contract1),
		Refresh:    "true",
	}
	contract2 := Contract{
		Address:             addr2,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}
	req2 := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addr2.String(),
		Body:       esutil.NewJSONReader(contract2),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(req1)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	})
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(req2)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	})

	db, _ := New(mockedClient)

	err := db.AddAddresses([]common.Address{addr1, addr2})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddNoAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test

	db, _ := New(mockedClient)

	err := db.AddAddresses([]common.Address{})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddSingleAddressWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contract := Contract{
		Address:             addr,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}

	ex := esapi.IndexRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(ex)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	}).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.AddAddresses([]common.Address{addr})

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_DeleteAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	req := esapi.DeleteRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewDeleteRequestMatcher(req)).Return(nil, nil)

	db, _ := New(mockedClient)

	err := db.DeleteAddress(addr)

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_DeleteAddress_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	req := esapi.DeleteRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewDeleteRequestMatcher(req)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.DeleteAddress(addr)

	assert.EqualError(t, err, "error deleting address: test error", "wrong error message")
}

func TestElasticsearchDB_GetAddresses_NoAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

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

	sampleAddress := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	createReturnValue := func(addr common.Address) interface{} {
		sampleReturnValue := `{"_source" : { "address": "%s"}}`
		withAddress := fmt.Sprintf(sampleReturnValue, addr.String())
		var asInterface map[string]interface{}
		_ = json.Unmarshal([]byte(withAddress), &asInterface)
		return asInterface
	}

	mockedClient := NewMockAPIClient(ctrl)

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

	sampleAddress1 := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	sampleAddress2 := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f35")
	createReturnValue := func(addr common.Address) interface{} {
		sampleReturnValue := `{"_source" : { "address": "%s"}}`
		withAddress := fmt.Sprintf(sampleReturnValue, addr.String())
		var asInterface map[string]interface{}
		_ = json.Unmarshal([]byte(withAddress), &asInterface)
		return asInterface
	}

	mockedClient := NewMockAPIClient(ctrl)

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

	mockedClient := NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(ContractIndex, QueryAllAddressesTemplate).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)
	allAddresses, err := db.GetAddresses()

	assert.Nil(t, allAddresses, "error was not nil")
	assert.EqualError(t, err, "error fetching addresses: test error", "wrong error message")
}
