package elasticsearch

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/consensys/quorum-go-utils/types"
)

// constant query template strings for ES
const QueryAllAddressesTemplate = `
{
	"_source": ["address"],
	"query": {
		"match_all": {}
	}
}
`

const QueryAllTemplateNamesTemplate = `
{
	"_source": ["templateName"],
	"query": {
		"match_all": {}
	}
}
`

func QueryByToAddressWithOptionsTemplate(options *types.QueryOptions) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "to": "%s" } },
` + createRangeQuery("blockNumber", options.BeginBlockNumber, options.EndBlockNumber) + `,
` + createRangeQuery("timestamp", options.BeginTimestamp, options.EndTimestamp) + `
			]
		}
	}
}
`
}

func QueryByAddressWithOptionsTemplate(options *types.QueryOptions) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "address": "%s" } },
` + createRangeQuery("blockNumber", options.BeginBlockNumber, options.EndBlockNumber) + `,
` + createRangeQuery("timestamp", options.BeginTimestamp, options.EndTimestamp) + `
			]
		}
	}
}
`
}

func QueryInternalTransactionsWithOptionsTemplate(options *types.QueryOptions) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "nested": {
						"path": "internalCalls",
						"query": {
							"bool": {
								"must": [
									{ "match": { "internalCalls.to": "%s" } }
								]
							}
						}
					}
				},
` + createRangeQuery("blockNumber", options.BeginBlockNumber, options.EndBlockNumber) + `,
` + createRangeQuery("timestamp", options.BeginTimestamp, options.EndTimestamp) + `
			]
		}
	}
}
`
}

// This query will get all the balances between a certain block range, as well as the
// last balance before the starting block IF there was no balance update ON the starting block
func QueryTokenBalanceAtBlockRange(options *types.TokenQueryOptions) string {
	rangeQuery := `
      "filter": [
        {
          "bool": {
            "should": [
              ` + createRangeQuery("blockNumber", options.BeginBlockNumber, options.EndBlockNumber) + `,
              {
                "bool": {
                  "must": [{"range": {"blockNumber": {"lt": %d}}}],
                  "filter": [
                    {
                      "bool": {
                        "should": [
                          {"range": {"heldUntil": {"gte": %d}}},
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
`
	rangeQuery = fmt.Sprintf(rangeQuery, options.BeginBlockNumber.Uint64(), options.BeginBlockNumber.Uint64())

	return `
{
  "query": {
    "bool": {
` + rangeQuery + `
      "must": [
        {"match": {"contract": "%s"}},
        {"match": {"holder": "%s"}}
      ]
    }
  }
}
`
}

func QueryERC20TokenBalanceAtBlock() string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "match": { "holder": "%s" } },
				{ "range": { "blockNumber": { "lte": %d } } }
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
}

func QueryERC20TokenHoldersAtBlock() string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "range": { "blockNumber": { "lte": %d } } }
			],
			"filter": [{
                "bool": {
                    "should": [
						{ "range": { "heldUntil": { "gte": %d } } }, 
						{ "bool": { "must_not": { "exists": { "field": "heldUntil" } } } }
					]
                }
            }]
		}
	},
	"size": 0,
	"aggs" : {
		"result_buckets": {
			"composite" : {
				"size": %d,
				%s
				"sources" : [
					{ "holder": { "terms" : { "field": "holder.keyword" } } }
				]
		  	}
		}
	}
}
`
}

func createRangeQuery(name string, start *big.Int, end *big.Int) string {
	if end.Cmp(big.NewInt(-1)) == 0 {
		return fmt.Sprintf(`{ "range": { "%s": { "gte": %s } } }`, name, start.String())
	}
	return fmt.Sprintf(`{ "range": { "%s": { "gte": %s, "lte": %s } } }`, name, start.String(), end.String())
}

func QueryERC721TokenAtBlock() string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "match": { "token": "%s"} },
				{ "range": { "heldFrom": { "lte": %d } } }
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
}

func QueryERC721HolderAtBlock(start *big.Int) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "match": { "holder": "%s"} },
				{ "range": { "heldFrom": { "lte": %d } } },
` + createTokenRangeQuery(start) + `
			],
			"filter": [{
                "bool": {
                    "should": [
						{ "range": { "heldUntil": { "gte": %d } } }, 
						{ "bool": { "must_not": { "exists": { "field": "heldUntil" } } } }
					]
                }
            }]
		}
	}
}
`
}

func QueryERC721AllTokensAtBlock(start *big.Int) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "range": { "heldFrom": { "lte": %d } } },
` + createTokenRangeQuery(start) + `
			],
			"filter": [{
                "bool": {
                    "should": [
						{ "range": { "heldUntil": { "gte": %d } } }, 
						{ "bool": { "must_not": { "exists": { "field": "heldUntil" } } } }
					]
                }
            }]
		}
	}
}
`
}

func QueryERC721AllHoldersAtBlock() string {
	return `
{
	"_source": ["token", "holder"],
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "range": { "heldFrom": { "lte": %d } } }
			],
			"filter": [{
                "bool": {
                    "should": [
						{ "range": { "heldUntil": { "gte": %d } } }, 
						{ "bool": { "must_not": { "exists": { "field": "heldUntil" } } } }
					]
                }
            }]
		}
	},
	"size": 0,
	"aggs" : {
		"result_buckets": {
			"composite" : {
				"size": %d,
				%s
				"sources" : [
					{ "holder": { "terms" : { "field": "holder.keyword" } } }
				]
		  	}
		}
	}
}
`
}

func createTokenRangeQuery(start *big.Int) string {
	next := new(big.Int).Add(start, big.NewInt(1))

	paddedStartTokenId := fmt.Sprintf("%085d", next)
	startFirst, _ := strconv.ParseUint(paddedStartTokenId[0:17], 10, 64)
	startSecond, _ := strconv.ParseUint(paddedStartTokenId[17:34], 10, 64)
	startThird, _ := strconv.ParseUint(paddedStartTokenId[34:51], 10, 64)
	startFourth, _ := strconv.ParseUint(paddedStartTokenId[51:68], 10, 64)
	startFifth, _ := strconv.ParseUint(paddedStartTokenId[68:85], 10, 64)

	return fmt.Sprintf(
		"%s, %s, %s, %s, %s",
		fmt.Sprintf(`{ "range": { "%s": { "gte": %d } } }`, "first", startFirst),
		fmt.Sprintf(`{ "range": { "%s": { "gte": %d } } }`, "second", startSecond),
		fmt.Sprintf(`{ "range": { "%s": { "gte": %d } } }`, "third", startThird),
		fmt.Sprintf(`{ "range": { "%s": { "gte": %d } } }`, "fourth", startFourth),
		fmt.Sprintf(`{ "range": { "%s": { "gte": %d } } }`, "fifth", startFifth),
	)
}
