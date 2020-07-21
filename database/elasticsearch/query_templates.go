package elasticsearch

import (
	"fmt"
	"math/big"

	"quorumengineering/quorum-report/types"
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

func QueryTokenBalanceAtBlockRange(options *types.QueryOptions) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "contract": "%s"} },
				{ "match": { "holder": "%s" } },
` + createRangeQuery("blockNumber", options.BeginBlockNumber, options.EndBlockNumber) + `
			]
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
