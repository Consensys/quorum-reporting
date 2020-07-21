package client

import "quorumengineering/quorum-report/types"

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

func TransactionDetailQuery(hash types.Hash) string {
	return `query { transaction(hash:"` + hash.Hex() + `") {
        hash
        status
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
