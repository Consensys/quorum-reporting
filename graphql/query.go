package graphql

import "github.com/ethereum/go-ethereum/common"

// templates for GraphQL queries

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
        status
		block { number, hash }
		index
        nonce
        from { address }
        to { address }
        value
        gasPrice
        gas
        gasUsed
        cumulativeGasUsed
        createdContract { address }
		inputData
		privateInputData
		isPrivate
		logs {
			index
			account { address }
			topics
			data
		}
    } }`
}
