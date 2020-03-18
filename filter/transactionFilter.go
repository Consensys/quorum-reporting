package filter

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

type TransactionFilter struct {
	db           database.Database
	quorumClient *client.QuorumClient
	addresses    []common.Address
}

func NewTransactionFilter(db database.Database, quorumClient *client.QuorumClient, addresses []common.Address) *TransactionFilter {
	return &TransactionFilter{
		db,
		quorumClient,
		addresses,
	}
}

func (tf *TransactionFilter) FilterBlock(block *types.Block) {
	fmt.Printf("Filter block %v\n", block.NumberU64())
	for _, tx := range block.Transactions() {
		fmt.Printf("Get TX %v\n", tx.Hash().Hex())
	}
	// Use GraphQL to get isPrivate & privateInputData
	// Write transactions to DB
	// Index transactions related to registered contract addresses
	// Index events related to registered contract addresses
}
