package elasticsearch

import (
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

func createRangeQuery(name string, start *big.Int, end *big.Int) string {
	if end.Cmp(big.NewInt(-1)) == 0 {
		return "{ \"range\": { \"" + name + "\": { \"gte\": " + start.String() + "} } }"
	}
	return "{ \"range\": { \"" + name + "\": { \"gte\": " + start.String() + ", \"lte\": " + end.String() + "} } }"
}
