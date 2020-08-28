package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	elasticsearchmocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"testing"

	"github.com/consensys/quorum-go-utils/types"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestElasticsearchDB_AddTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

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

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
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

func TestElasticsearchDB_GetContractABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

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

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

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

func TestElasticsearchDB_GetStorageABI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

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

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

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

func TestElasticsearchDB_GetTemplates_NoTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(TemplateIndex, QueryAllTemplateNamesTemplate).
		Return(make([]interface{}, 0, 0), nil)

	db, _ := New(mockedClient)
	allAddresses, err := db.GetTemplates()

	assert.Nil(t, err, "error was not nil")
	assert.Equal(t, 0, len(allAddresses), "templates found when none expected: %s", allAddresses)
}

func TestElasticsearchDB_GetTemplates_SingleTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleTemplate := "sampleTemplate"
	createReturnValue := func(template string) interface{} {
		sampleReturnValue := `{"_source" : { "templateName": "%s"}}`
		var asInterface map[string]interface{}
		_ = json.Unmarshal([]byte(fmt.Sprintf(sampleReturnValue, template)), &asInterface)
		return asInterface
	}

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(TemplateIndex, QueryAllTemplateNamesTemplate).
		Return([]interface{}{createReturnValue(sampleTemplate)}, nil)

	db, _ := New(mockedClient)
	allTemplates, err := db.GetTemplates()

	assert.Nil(t, err, "error was not nil")
	assert.Equal(t, 1, len(allTemplates), "wrong number of templates found: %v", len(allTemplates))
	assert.Equal(t, sampleTemplate, allTemplates[0], "unexpected template found: %s", allTemplates[0])
}

func TestElasticsearchDB_GetTemplates_MultipleTemplates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sampleTemplate1 := "sampleTemplate1"
	sampleTemplate2 := "sampleTemplate2"
	createReturnValue := func(template string) interface{} {
		sampleReturnValue := `{"_source" : { "templateName": "%s"}}`
		var asInterface map[string]interface{}
		_ = json.Unmarshal([]byte(fmt.Sprintf(sampleReturnValue, template)), &asInterface)
		return asInterface
	}

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(TemplateIndex, QueryAllTemplateNamesTemplate).
		Return([]interface{}{createReturnValue(sampleTemplate1), createReturnValue(sampleTemplate2)}, nil)

	db, _ := New(mockedClient)
	allTemplates, err := db.GetTemplates()

	assert.Nil(t, err, "error was not nil")
	assert.Equal(t, 2, len(allTemplates), "wrong number of templates found: %v", len(allTemplates))
	assert.Equal(t, sampleTemplate1, allTemplates[0], "unexpected template found: %s", allTemplates[0])
	assert.Equal(t, sampleTemplate2, allTemplates[1], "unexpected template found: %s", allTemplates[0])
}

func TestElasticsearchDB_GetTemplates_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		ScrollAllResults(TemplateIndex, QueryAllTemplateNamesTemplate).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)
	allTemplates, err := db.GetTemplates()

	assert.Nil(t, allTemplates, "error was not nil")
	assert.EqualError(t, err, "error fetching templates: test error", "wrong error message")
}
