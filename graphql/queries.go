package graphql

import "github.com/ethereum/go-ethereum/common"

// TODO: Templates for GraphQL queries
func GetTransactionDetailQuery(hash common.Hash) string {
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