package elasticsearch

import (
	"errors"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearch_mocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
)

func TestElasticsearchDB_AddTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	template := Template{
		TemplateName: "test template",
		ABI:          "test abi",
		StorageABI:   "test storage",
	}

	ex := esapi.IndexRequest{
		Index:      TemplateIndex,
		DocumentID: template.TemplateName,
		Body:       esutil.NewJSONReader(template),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(ex))

	db, err := New(mockedClient)

	err = db.AddTemplate(template.TemplateName, template.ABI, template.StorageABI)

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_AssignTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	templateName := "test template"

	searchContractRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
	       "_source": {
	         "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
	         "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
	         "lastFiltered" : 20,
	         "templateName": "old template"
	       }
	}`
	contractQuery := map[string]interface{}{
		"doc": map[string]interface{}{
			"templateName": templateName,
		},
	}
	contractUpdateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contractQuery),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(contractUpdateRequest))

	db, err := New(mockedClient)

	err = db.AssignTemplate(addr, templateName)

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_AddContractABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test ABI string"

	contractQuery := map[string]interface{}{
		"doc": map[string]interface{}{
			"templateName": addr.String(),
		},
	}
	contractUpdateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contractQuery),
		Refresh:    "true",
	}

	searchContractRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
	       "_source": {
	         "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
	         "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
	         "lastFiltered" : 20,
	         "templateName": "template"
	       }
	}`
	searchTemplateRequest1 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: "template",
	}
	searchTemplateRequest2 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: addr.String(),
	}
	templateSearchResultValue := `{
	       "_source": {
	         "templateName": "template",
	         "abi": "template abi",
	         "storageAbi": "template storage layout",
	       }
	}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest1)).Return([]byte(templateSearchResultValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest2)).Return(nil, ErrNotFound)
	mockedClient.EXPECT().DoRequest(gomock.Any()) // update template
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(contractUpdateRequest))

	db, _ := New(mockedClient)

	err := db.AddContractABI(addr, abi)

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_AddContractABI_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test ABI string"

	contractQuery := map[string]interface{}{
		"doc": map[string]interface{}{
			"templateName": addr.String(),
		},
	}
	contractUpdateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contractQuery),
		Refresh:    "true",
	}

	searchContractRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
	       "_source": {
	         "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
	         "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
	         "lastFiltered" : 20,
	         "templateName": "template"
	       }
	}`
	searchTemplateRequest1 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: "template",
	}
	searchTemplateRequest2 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: addr.String(),
	}
	templateSearchResultValue := `{
	       "_source": {
	         "templateName": "template",
	         "abi": "template abi",
	         "storageAbi": "template storage layout",
	       }
	}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest1)).Return([]byte(templateSearchResultValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest2)).Return(nil, ErrNotFound)
	mockedClient.EXPECT().DoRequest(gomock.Any()) // update template
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(contractUpdateRequest)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.AddContractABI(addr, abi)

	assert.EqualError(t, err, "test error", "wrong error message")
}

func TestElasticsearchDB_AddContractABI_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

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

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contractSearchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
        "_source": {
          "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
          "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
          "lastFiltered" : 20,
          "templateName": "template"
        }
	}`
	templateSearchRequest := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: "template",
	}
	templateSearchReturnValue := `{
        "_source": {
          "templateName": "template",
          "abi": "test abi"
        }
	}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(contractSearchRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(templateSearchRequest)).Return([]byte(templateSearchReturnValue), nil)

	db, _ := New(mockedClient)

	abi, err := db.GetContractABI(addr)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "test abi", abi)
}

func TestElasticsearchDB_GetContractABI_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

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

func TestElasticsearchDB_AddStorageABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test storage ABI string"

	contractQuery := map[string]interface{}{
		"doc": map[string]interface{}{
			"templateName": addr.String(),
		},
	}
	contractUpdateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contractQuery),
		Refresh:    "true",
	}

	searchContractRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
	       "_source": {
	         "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
	         "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
	         "lastFiltered" : 20,
	         "templateName": "template"
	       }
	}`
	searchTemplateRequest1 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: "template",
	}
	searchTemplateRequest2 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: addr.String(),
	}
	templateSearchResultValue := `{
	       "_source": {
	         "templateName": "template",
	         "abi": "template abi",
	         "storageAbi": "template storage layout",
	       }
	}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest1)).Return([]byte(templateSearchResultValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest2)).Return(nil, ErrNotFound)
	mockedClient.EXPECT().DoRequest(gomock.Any()) // update template
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(contractUpdateRequest))

	db, _ := New(mockedClient)

	err := db.AddStorageLayout(addr, abi)

	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_AddStorageABI_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test storage ABI string"

	contractQuery := map[string]interface{}{
		"doc": map[string]interface{}{
			"templateName": addr.String(),
		},
	}
	contractUpdateRequest := esapi.UpdateRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contractQuery),
		Refresh:    "true",
	}

	searchContractRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
	       "_source": {
	         "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
	         "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
	         "lastFiltered" : 20,
	         "templateName": "template"
	       }
	}`
	searchTemplateRequest1 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: "template",
	}
	searchTemplateRequest2 := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: addr.String(),
	}
	templateSearchResultValue := `{
	       "_source": {
	         "templateName": "template",
	         "abi": "template abi",
	         "storageAbi": "template storage layout",
	       }
	}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest1)).Return([]byte(templateSearchResultValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchTemplateRequest2)).Return(nil, ErrNotFound)
	mockedClient.EXPECT().DoRequest(gomock.Any()) // update template
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchContractRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(contractUpdateRequest)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.AddStorageLayout(addr, abi)

	assert.EqualError(t, err, "test error", "wrong error message")
}

func TestElasticsearchDB_AddStorageABI_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	abi := "test storage ABI string"

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return(nil, errors.New("not found"))

	db, _ := New(mockedClient)

	err := db.AddStorageLayout(addr, abi)

	assert.EqualError(t, err, "not found", "wrong error message")
}

func TestElasticsearchDB_GetStorageABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contractSearchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	contractSearchReturnValue := `{
        "_source": {
          "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
          "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
          "lastFiltered" : 20,
          "templateName" : "template"
        }
	}`
	templateSearchRequest := esapi.GetRequest{
		Index:      TemplateIndex,
		DocumentID: "template",
	}
	templateSearchReturnValue := `{
        "_source": {
          "templateName": "template",
          "storageAbi": "some storage ABI"
        }
	}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(contractSearchRequest)).Return([]byte(contractSearchReturnValue), nil)
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(templateSearchRequest)).Return([]byte(templateSearchReturnValue), nil)

	db, _ := New(mockedClient)

	abi, err := db.GetStorageLayout(addr)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, "some storage ABI", abi)
}

func TestElasticsearchDB_GetStorageABI_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return(nil, errors.New("not found"))

	db, _ := New(mockedClient)

	abi, err := db.GetStorageLayout(addr)

	assert.Equal(t, "", abi, "unexpected error")
	assert.EqualError(t, err, "not found", "unexpected error message")
}
