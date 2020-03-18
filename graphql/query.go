package graphql

import "github.com/ethereum/go-ethereum/common"

// Templates for GraphQL queries

func CurrentBlockQuery() string {
	return `
		query {
			block {
				number
			}
		}
	`
}

func TransactionDetailQuery(hash common.Hash) string {
	return `query { transaction(hash:"` + hash.Hex() + `") {
        hash
        nonce
        index
        from { address }
        to { address }
        value
        gasPrice
        gas
        inputData
        block { number }
        status
        gasUsed
        cumulativeGasUsed
        createdContract { address }
        logs { topics }
		isPrivate
		privateInputData
    } }`
}
