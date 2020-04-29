package elasticsearch

// constant query template strings for ES
const (
	QueryAllAddressesTemplate = `
{
	"_source": ["address"],
	"query": {
		"match_all": {}
	}
}`
	QueryByToAddressTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "to": "%s" } }
			]
		}
	}
}`
	QueryByAddressTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "address": "%s" } }
			]
		}
	}
}`
	QueryInternalTransactions = `
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
