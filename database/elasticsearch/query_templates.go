package elasticsearch

import "quorumengineering/quorum-report/types"

// constant query template strings for ES
const (
	QueryAllAddressesTemplate = `
{
	"_source": ["address"],
	"query": {
		"match_all": {}
	}
}
`
	QueryByToAddressTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "to": "%s" } }
			]
		}
	}
}
`
	QueryByAddressTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "address": "%s" } }
			]
		}
	}
}
`
	QueryInternalTransactionsTemplate = `
{
  "query": {
    "nested": {
      "path": "internalCalls",
      "query": {
        "bool": {
          "must": [
            {
              "match": { "internalCalls.to": "%s" }
            }
          ]
        }
      }
    }
  }
}
`
)

func QueryByToAddressWithOptionsTemplate(options *types.QueryOptions) string {
	return `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "to": "%s" } },
				{ "range": { "blockNumber": { "gte": ` + options.StartBlock.String() + `, "lte": ` + options.EndBlock.String() + `} } },
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
				{ "range": { "blockNumber": { "gte": ` + options.StartBlock.String() + `, "lte": ` + options.EndBlock.String() + `} } },
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
		"nested": {
			"path": "internalCalls",
			"query": {
				"bool": {
					"must": [
						{ "match": { "internalCalls.to": "%s" } },
						{ "range": { "blockNumber": { "gte": ` + options.StartBlock.String() + `, "lte": ` + options.EndBlock.String() + `} } },
					]
				}
			}
		}
	}
}
`
}
