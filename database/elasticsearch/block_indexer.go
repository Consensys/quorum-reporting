package elasticsearch

import (
	"quorumengineering/quorum-report/types"
)

type DefaultBlockIndexer struct {
	addresses map[types.Address]bool
	blocks    []*types.Block
	// function pointers currently originated from ES database implementation only
	// TODO: May convert all functions into an interface. DefaultBlockIndexer can then accept all database implementation and move to a util package.
	createEvents    func([]*types.Event) error
	readTransaction func(types.Hash) (*types.Transaction, error)
}

func NewBlockIndexer(addresses []types.Address, blocks []*types.Block, db *ElasticsearchDB) *DefaultBlockIndexer {
	addressMap := map[types.Address]bool{}
	for _, address := range addresses {
		addressMap[address] = true
	}

	return &DefaultBlockIndexer{
		addresses:       addressMap,
		blocks:          blocks,
		createEvents:    db.createEvents,
		readTransaction: db.ReadTransaction,
	}
}

func (indexer *DefaultBlockIndexer) Index() error {
	allTransactions, err := indexer.fetchTransactions()
	if err != nil {
		return err
	}

	return indexer.indexEvents(allTransactions)
}

func (indexer *DefaultBlockIndexer) indexEvents(transactions []*types.Transaction) error {
	var pendingIndexEvents []*types.Event
	for _, transaction := range transactions {
		for _, event := range transaction.Events {
			if indexer.addresses[event.Address] {
				pendingIndexEvents = append(pendingIndexEvents, event)
			}
		}
	}

	return indexer.createEvents(pendingIndexEvents)
}

func (indexer *DefaultBlockIndexer) fetchTransactions() ([]*types.Transaction, error) {
	transactions := make([]*types.Transaction, 0)
	for _, block := range indexer.blocks {
		for _, txHash := range block.Transactions {
			transaction, err := indexer.readTransaction(txHash)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, transaction)
		}
	}
	return transactions, nil
}
