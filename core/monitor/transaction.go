package monitor

import (
	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type TransactionMonitor interface {
	PullTransactions(block *types.Block) ([]*types.Transaction, error)
}

type DefaultTransactionMonitor struct {
	quorumClient client.Client
}

func NewDefaultTransactionMonitor(quorumClient client.Client) *DefaultTransactionMonitor {
	return &DefaultTransactionMonitor{
		quorumClient: quorumClient,
	}
}

func (tm *DefaultTransactionMonitor) PullTransactions(block *types.Block) ([]*types.Transaction, error) {
	log.Info("Fetching transactions", "block", block.Hash.String(), "blockNumber", block.Number)

	fetchedTransactions := make([]*types.Transaction, 0, len(block.Transactions))
	for _, txHash := range block.Transactions {
		// Query transaction details by graphql.
		tx, err := tm.fetchTransaction(block, txHash)
		if err != nil {
			return nil, err
		}
		fetchedTransactions = append(fetchedTransactions, tx)
	}
	return fetchedTransactions, nil
}

func (tm *DefaultTransactionMonitor) fetchTransaction(block *types.Block, hash types.Hash) (*types.Transaction, error) {
	log.Debug("Processing transaction", "hash", hash.String())

	txOrigin, err := client.TransactionWithReceipt(tm.quorumClient, hash)
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		Hash:              hash,
		Status:            txOrigin.Status == 1,
		BlockNumber:       block.Number,
		BlockHash:         block.Hash,
		Index:             txOrigin.Index,
		Nonce:             txOrigin.Nonce.ToUint64(),
		From:              txOrigin.From.Address,
		To:                txOrigin.To.Address,
		Value:             txOrigin.Value.ToUint64(),
		Gas:               txOrigin.Gas.ToUint64(),
		GasUsed:           txOrigin.GasUsed.ToUint64(),
		GasPrice:          txOrigin.GasPrice.ToUint64(),
		CumulativeGasUsed: txOrigin.CumulativeGasUsed.ToUint64(),
		CreatedContract:   txOrigin.CreatedContract.Address,
		Data:              txOrigin.InputData,
		PrivateData:       txOrigin.PrivateInputData,
		IsPrivate:         txOrigin.IsPrivate,
		Timestamp:         block.Timestamp,
	}

	tx.Events = make([]*types.Event, len(txOrigin.Logs))
	for i, l := range txOrigin.Logs {
		tx.Events[i] = &types.Event{
			Index:            l.Index,
			Address:          l.Account.Address,
			Topics:           l.Topics,
			Data:             l.Data,
			BlockNumber:      block.Number,
			BlockHash:        block.Hash,
			TransactionHash:  tx.Hash,
			TransactionIndex: txOrigin.Index,
			Timestamp:        block.Timestamp,
		}
	}

	traceResp, err := client.TraceTransaction(tm.quorumClient, tx.Hash)
	if err != nil {
		return nil, err
	}

	calls := flattenCalls(traceResp.Calls)
	tx.InternalCalls = make([]*types.InternalCall, len(calls))
	for i, respCall := range calls {
		tx.InternalCalls[i] = &types.InternalCall{
			From:    respCall.From,
			To:      respCall.To,
			Gas:     respCall.Gas.ToUint64(),
			GasUsed: respCall.GasUsed.ToUint64(),
			Value:   respCall.Value.ToUint64(),
			Input:   respCall.Input,
			Output:  respCall.Output,
			Type:    respCall.Type,
		}
	}
	return tx, nil
}

//flattens the list of internal calls to a single list
//e.g [1 [2 3 [4 5] 6 [7]]] -> [1 2 3 4 5 6 7]
func flattenCalls(calls []types.RawInnerCall) []types.RawInnerCall {
	if len(calls) == 0 {
		return []types.RawInnerCall{}
	}

	var results []types.RawInnerCall
	for _, c := range calls {
		results = append(results, c)
		results = append(results, flattenCalls(c.Calls)...)
	}
	return results
}
