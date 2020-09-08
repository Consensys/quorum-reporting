package elasticsearch

import (
	"math/big"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearchmocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"quorumengineering/quorum-report/types"
)

func TestElasticsearchDB_RecordNewERC20Balance_NoPrevious(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holderAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	blockNumber := uint64(10)
	balance := big.NewInt(1989)

	token := ERC20TokenHolder{
		Contract:    tokenContractAddress,
		Holder:      holderAddress,
		BlockNumber: blockNumber,
		Amount:      balance.String(),
	}
	ex := esapi.IndexRequest{
		Index:      ERC20TokenIndex,
		DocumentID: "0x1932c48b2bf8102ba33b4a6b545c32236e342f34-0x1349f3e1b8d71effb47b840594ff27da7e603d17-10",
		Body:       esutil.NewJSONReader(token),
	}

	searchQuery := `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"} },
				{ "match": { "holder": "0x1349f3e1b8d71effb47b840594ff27da7e603d17" } },
				{ "range": { "blockNumber": { "lte": 9 } } }
			]
		}
	},
	"sort": [
			{
				"blockNumber": {
					"order": "desc",
					"unmapped_type": "long"
				}
			}
	]
}
`
	size := 1
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(searchQuery),
		Size:  &size,
	}
	searchResult := `{"hits": {"hits": []}}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(searchResult), nil)
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(ex)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	})

	db, _ := New(mockedClient)
	err := db.RecordNewERC20Balance(tokenContractAddress, holderAddress, blockNumber, balance)
	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_RecordNewERC20Balance_WithPrevious(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holderAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	blockNumber := uint64(10)
	balance := big.NewInt(1989)

	token := ERC20TokenHolder{
		Contract:    tokenContractAddress,
		Holder:      holderAddress,
		BlockNumber: blockNumber,
		Amount:      balance.String(),
	}
	ex := esapi.IndexRequest{
		Index:      ERC20TokenIndex,
		DocumentID: "0x1932c48b2bf8102ba33b4a6b545c32236e342f34-0x1349f3e1b8d71effb47b840594ff27da7e603d17-10",
		Body:       esutil.NewJSONReader(token),
	}

	searchQuery := `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"} },
				{ "match": { "holder": "0x1349f3e1b8d71effb47b840594ff27da7e603d17" } },
				{ "range": { "blockNumber": { "lte": 9 } } }
			]
		}
	},
	"sort": [
			{
				"blockNumber": {
					"order": "desc",
					"unmapped_type": "long"
				}
			}
	]
}
`
	size := 1
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(searchQuery),
		Size:  &size,
	}
	searchResult := `{"hits": {"hits": [
{"_source": {
		"contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
		"holder": "0x1349f3e1b8d71effb47b840594ff27da7e603d17",
		"token": "500",
		"blockNumber": 7
	}
}
]}}`

	oldTokenUpdateReq := esapi.UpdateRequest{
		Index:      ERC20TokenIndex,
		DocumentID: "0x1932c48b2bf8102ba33b4a6b545c32236e342f34-0x1349f3e1b8d71effb47b840594ff27da7e603d17-7",
		Body: strings.NewReader(`{"doc":{"heldUntil":9}}
`),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(searchResult), nil)
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(ex)).Do(func(input esapi.IndexRequest) {
		assert.Equal(t, "create", input.OpType)
	})
	mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(oldTokenUpdateReq)).Return(nil, nil)

	db, _ := New(mockedClient)
	err := db.RecordNewERC20Balance(tokenContractAddress, holderAddress, blockNumber, balance)
	assert.Nil(t, err, "expected error to be nil")
}

func TestElasticsearchDB_GetERC20Balance_PaginationTooLarge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)
	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holderAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	options := &types.TokenQueryOptions{
		PageSize:   100,
		PageNumber: 11,
	}
	options.SetDefaults()

	db, _ := New(mockedClient)
	results, err := db.GetERC20Balance(tokenContractAddress, holderAddress, options)

	assert.Nil(t, results)
	assert.EqualError(t, err, "pagination limit exceeded")
}

func TestElasticsearchDB_GetERC20Balance_NoResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holderAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	options := &types.TokenQueryOptions{}
	options.SetDefaults()

	expectedQuery := `
{
  "query": {
    "bool": {

      "filter": [
        {
          "bool": {
            "should": [
              { "range": { "blockNumber": { "gte": 0 } } },
              {
                "bool": {
                  "must": [{"range": {"blockNumber": {"lt": 0}}}],
                  "filter": [
                    {
                      "bool": {
                        "should": [
                          {"range": {"heldUntil": {"gte": 0}}},
                          {"bool": {"must_not": {"exists": {"field": "heldUntil"}}}}
                        ]
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      ],

      "must": [
        {"match": {"contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"}},
        {"match": {"holder": "0x1349f3e1b8d71effb47b840594ff27da7e603d17"}}
      ]
    }
  }
}
`

	from := 0
	size := 10
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(expectedQuery),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc"},
	}

	result := `{"hits": {"hits": []}}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(result), nil)

	db, _ := New(mockedClient)
	results, err := db.GetERC20Balance(tokenContractAddress, holderAddress, options)

	assert.Nil(t, err)
	assert.Len(t, results, 0)
}

func TestElasticsearchDB_GetERC20Balance_MultipleResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holderAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	options := &types.TokenQueryOptions{}
	options.SetDefaults()

	expectedQuery := `
{
  "query": {
    "bool": {

      "filter": [
        {
          "bool": {
            "should": [
              { "range": { "blockNumber": { "gte": 0 } } },
              {
                "bool": {
                  "must": [{"range": {"blockNumber": {"lt": 0}}}],
                  "filter": [
                    {
                      "bool": {
                        "should": [
                          {"range": {"heldUntil": {"gte": 0}}},
                          {"bool": {"must_not": {"exists": {"field": "heldUntil"}}}}
                        ]
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      ],

      "must": [
        {"match": {"contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"}},
        {"match": {"holder": "0x1349f3e1b8d71effb47b840594ff27da7e603d17"}}
      ]
    }
  }
}
`

	from := 0
	size := 10
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(expectedQuery),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc"},
	}

	result := `{"hits": {"hits": [
  {"_source": {"blockNumber": 1, "amount": "500"}},
  {"_source": {"blockNumber": 2, "amount": "2000"}}
]}}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(result), nil)

	db, _ := New(mockedClient)
	results, err := db.GetERC20Balance(tokenContractAddress, holderAddress, options)

	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.EqualValues(t, 500, results[1].Int64())
	assert.EqualValues(t, 2000, results[2].Int64())
}

func TestElasticsearchDB_GetERC20Balance_ResultBeforeBeginBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holderAddress := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	options := &types.TokenQueryOptions{}
	options.BeginBlockNumber = big.NewInt(1)
	options.SetDefaults()

	expectedQuery := `
{
  "query": {
    "bool": {

      "filter": [
        {
          "bool": {
            "should": [
              { "range": { "blockNumber": { "gte": 1 } } },
              {
                "bool": {
                  "must": [{"range": {"blockNumber": {"lt": 1}}}],
                  "filter": [
                    {
                      "bool": {
                        "should": [
                          {"range": {"heldUntil": {"gte": 1}}},
                          {"bool": {"must_not": {"exists": {"field": "heldUntil"}}}}
                        ]
                      }
                    }
                  ]
                }
              }
            ]
          }
        }
      ],

      "must": [
        {"match": {"contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"}},
        {"match": {"holder": "0x1349f3e1b8d71effb47b840594ff27da7e603d17"}}
      ]
    }
  }
}
`

	from := 0
	size := 10
	req := esapi.SearchRequest{
		Index: []string{ERC20TokenIndex},
		Body:  strings.NewReader(expectedQuery),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc"},
	}

	result := `{"hits": {"hits": [
  {"_source": {"blockNumber": 0, "amount": "500"}}
]}}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(result), nil)

	db, _ := New(mockedClient)
	results, err := db.GetERC20Balance(tokenContractAddress, holderAddress, options)

	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.EqualValues(t, 500, results[0].Int64())
	assert.EqualValues(t, 500, results[1].Int64())
}

func TestElasticsearchDB_ERC721TokenByTokenID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	tokenId := big.NewInt(2000)

	expectedQuery := `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"} },
				{ "match": { "token": "2000"} },
				{ "range": { "heldFrom": { "lte": 12 } } }
			]
		}
	},
	"sort": [
		{
			"heldFrom": {
				"order": "desc",
				"unmapped_type": "long"
			}
		}
	]
}
`
	size := 1
	req := esapi.SearchRequest{
		Index: []string{ERC721TokenIndex},
		Body:  strings.NewReader(expectedQuery),
		Size:  &size,
	}

	resultJson := `{"hits": {"hits": []}}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(resultJson), nil)

	db, _ := New(mockedClient)
	result, err := db.ERC721TokenByTokenID(tokenContractAddress, 12, tokenId)

	assert.EqualError(t, err, "not found")
	assert.EqualValues(t, types.ERC721Token{}, result)
}

func TestElasticsearchDB_ERC721TokenByTokenID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	tokenContractAddress := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	tokenId := big.NewInt(2000)

	expectedQuery := `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"} },
				{ "match": { "token": "2000"} },
				{ "range": { "heldFrom": { "lte": 12 } } }
			]
		}
	},
	"sort": [
		{
			"heldFrom": {
				"order": "desc",
				"unmapped_type": "long"
			}
		}
	]
}
`
	size := 1
	req := esapi.SearchRequest{
		Index: []string{ERC721TokenIndex},
		Body:  strings.NewReader(expectedQuery),
		Size:  &size,
	}

	resultJson := `{"hits": {"hits": [{"_source": {
"contract": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
"holder": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
"token": "500",
"heldFrom": 1
}}]}}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(resultJson), nil)

	db, _ := New(mockedClient)
	result, err := db.ERC721TokenByTokenID(tokenContractAddress, 12, tokenId)

	expected := types.ERC721Token{
		Contract: "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
		Holder:   "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
		Token:    "500",
		HeldFrom: 1,
	}
	assert.Nil(t, err)
	assert.EqualValues(t, expected, result)
}
