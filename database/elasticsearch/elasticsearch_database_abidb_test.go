package elasticsearch

import (
	"errors"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestElasticsearchDB_AddContractABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test ABI string"

	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"abi": abi,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(updateRequest))

	db, _ := New(mockedClient)

	err := db.AddContractABI(addr, abi)

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_AddContractABI_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test ABI string"

	query := map[string]interface{}{
		"doc": map[string]interface{}{
			"abi": abi,
		},
	}

	updateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(query),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(updateRequest)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.AddContractABI(addr, abi)

	assert.EqualError(t, err, "test error", "wrong error message")
}

func TestElasticsearchDB_AddContractABI_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test ABI string"

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return(nil, errors.New("not found"))

	db, _ := New(mockedClient)

	err := db.AddContractABI(addr, abi)

	assert.EqualError(t, err, "not found", "wrong error message")
}

func TestElasticsearchDB_GetContractABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	contractSearchReturnValue := `{
        "_source": {
          "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
          "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
          "lastFiltered" : 20,
          "abi": "test abi"
        }
}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return([]byte(contractSearchReturnValue), nil)

	db, _ := New(mockedClient)

	abi, err := db.GetContractABI(addr)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "test abi", abi)
}

func TestElasticsearchDB_GetContractABI_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return(nil, errors.New("not found"))

	db, _ := New(mockedClient)

	abi, err := db.GetContractABI(addr)

	assert.Equal(t, "", abi, "unexpected error")
	assert.EqualError(t, err, "not found", "unexpected error message")
}
