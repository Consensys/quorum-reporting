package elasticsearch

import (
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type DefaultBlockIndexer struct {
	addresses map[types.Address]bool
	blocks    []*types.Block
	// function pointers currently originated from ES database implementation only
	// TODO: May convert all functions into an interface. DefaultBlockIndexer can then accept all database implementation and move to a util package.
	updateContract  func(types.Address, string, string) error
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
		updateContract:  db.updateContract,
		createEvents:    db.createEvents,
		readTransaction: db.ReadTransaction,
	}
}

func (indexer *DefaultBlockIndexer) Index() error {
	allTransactions, err := indexer.fetchTransactions()
	if err != nil {
		return err
	}

	for _, transaction := range allTransactions {
		if err := indexer.indexTransaction(transaction); err != nil {
			return err
		}
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

func (indexer *DefaultBlockIndexer) indexTransaction(tx *types.Transaction) error {
	// Compare the address with tx.CreatedContract to check if the transaction is related
	if indexer.addresses[tx.CreatedContract] {
		if err := indexer.updateContract(tx.CreatedContract, "creationTx", tx.Hash.String()); err != nil {
			return err
		}
		log.Info("Indexed contract creation tx of registered address", "tx", tx.Hash.Hex(), "address", tx.CreatedContract.Hex())
	}

	// Check all the internal calls for contract creations as well
	for _, internalCall := range tx.InternalCalls {
		if (internalCall.Type == "CREATE" || internalCall.Type == "CREATE2") && indexer.addresses[internalCall.To] {
			if err := indexer.updateContract(internalCall.To, "creationTx", tx.Hash.String()); err != nil {
				return err
			}
			log.Info("Indexed contract creation tx of registered address", "tx", tx.Hash.Hex(), "address", internalCall.To.Hex())
		}
	}
	return nil
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
