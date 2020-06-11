package elasticsearch

import (
	"log"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

type DefaultBlockIndexer struct {
	addresses map[common.Address]bool
	blocks    []*types.Block

	updateContract  func(common.Address, string, string) error
	createEvents    func([]*types.Event) error
	readTransaction func(common.Hash) (*types.Transaction, error)
}

func NewBlockIndexer(addresses []common.Address, blocks []*types.Block, db *ElasticsearchDB) *DefaultBlockIndexer {
	addressMap := map[common.Address]bool{}
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
		if err := indexer.IndexTransaction(transaction); err != nil {
			return err
		}
	}

	return indexer.IndexEvents(allTransactions)
}

func (indexer *DefaultBlockIndexer) IndexEvents(transactions []*types.Transaction) error {
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
		log.Printf("Index contract creation tx %v of registered address %v.\n", tx.Hash.Hex(), tx.CreatedContract.Hex())
	}

	// Check all the internal calls for contract creations as well
	for _, internalCall := range tx.InternalCalls {
		if (internalCall.Type == "CREATE" || internalCall.Type == "CREATE2") && indexer.addresses[internalCall.To] {
			if err := indexer.updateContract(internalCall.To, "creationTx", tx.Hash.String()); err != nil {
				return err
			}
			log.Printf("Index contract creation tx %v of registered address %v.\n", tx.Hash.Hex(), internalCall.To.Hex())
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
