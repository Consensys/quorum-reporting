package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"quorumengineering/quorum-report/types"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	elasticsearchmocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
)

func TestDefaultDeletionCoordinator_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	deleter := NewDefaultDeletionCoordinator(mockedClient)

	addressToDelete := types.NewAddress("1")

	ercDelete := esapi.DeleteByQueryRequest{
		Index: []string{ERC20TokenIndex, ERC721TokenIndex},
		Body:  strings.NewReader(`{ "query": { "match": { "contract": "0x0000000000000000000000000000000000000001" } } }`),
	}
	mockedClient.EXPECT().DoRequest(NewDeleteByQueryRequestMatcher(ercDelete)).Return(nil, nil)
	eventDelete := esapi.DeleteByQueryRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(`{ "query": { "match": { "address": "0x0000000000000000000000000000000000000001" } } }`),
	}
	mockedClient.EXPECT().DoRequest(NewDeleteByQueryRequestMatcher(eventDelete)).Return(nil, nil)
	storageDelete := esapi.DeleteByQueryRequest{
		Index: []string{StorageIndex},
		Body:  strings.NewReader(`{ "query": { "match": { "contract": "0x0000000000000000000000000000000000000001" } } }`),
	}
	mockedClient.EXPECT().DoRequest(NewDeleteByQueryRequestMatcher(storageDelete)).Return(nil, nil)
	templateDelete := esapi.DeleteRequest{
		Index:      TemplateIndex,
		DocumentID: addressToDelete.String(),
	}
	mockedClient.EXPECT().DoRequest(NewDeleteRequestMatcher(templateDelete)).Return(nil, nil)
	contractDelete := esapi.DeleteRequest{
		Index:      ContractIndex,
		DocumentID: addressToDelete.String(),
	}
	mockedClient.EXPECT().DoRequest(NewDeleteRequestMatcher(contractDelete)).Return(nil, nil)

	err := deleter.Delete(addressToDelete)
	assert.Nil(t, err)
}
