package elasticsearch

// constant query template strings for ES
const (
	UpdateContractABITemplate = `
{
	"doc": {
		"abi": "%s"
	}
}`
	QueryAllAddressesTemplate = `
{
	"_source": ["address"],
	"query": {
		"match_all": {}
	}
}`
	QueryByNumberTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "number": "%d" } }
			]
		}
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
	QueryByHashTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "hash": "%s" } }
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
	QueryByAddressAndBlockNumberTemplate = `
{
	"query": {
		"bool": {
			"must": [
				{ "match": { "address": "%s" } },
				{ "match": { "blockNumber": "%d" } }
			]
		}
	}
}`
)
