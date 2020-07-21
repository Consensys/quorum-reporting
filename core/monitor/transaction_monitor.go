package monitor

import (
	"strconv"

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
		tx, err := tm.createTransaction(block, txHash)
		if err != nil {
			return nil, err
		}
		fetchedTransactions = append(fetchedTransactions, tx)
	}
	return fetchedTransactions, nil
}

func (tm *DefaultTransactionMonitor) createTransaction(block *types.Block, hash types.Hash) (*types.Transaction, error) {
	log.Debug("Processing transaction", "hash", hash.String())

	txOrigin, err := client.TransactionWithReceipt(tm.quorumClient, hash)
	if err != nil {
		return nil, err
	}

	// Create reporting transaction struct fields.
	nonce, err := strconv.ParseUint(txOrigin.Nonce, 0, 64)
	if err != nil {
		return nil, err
	}
	value, err := strconv.ParseUint(txOrigin.Value, 0, 64)
	if err != nil {
		return nil, err
	}
	gas, err := strconv.ParseUint(txOrigin.Gas, 0, 64)
	if err != nil {
		return nil, err
	}
	gasUsed, err := strconv.ParseUint(txOrigin.GasUsed, 0, 64)
	if err != nil {
		return nil, err
	}
	cumulativeGasUsed, err := strconv.ParseUint(txOrigin.CumulativeGasUsed, 0, 64)
	if err != nil {
		return nil, err
	}
	gasPrice, err := strconv.ParseUint(txOrigin.GasPrice, 0, 64)
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		Hash:              hash,
		Status:            txOrigin.Status == "0x1",
		BlockNumber:       block.Number,
		BlockHash:         block.Hash,
		Index:             txOrigin.Index,
		Nonce:             nonce,
		From:              types.NewAddress(txOrigin.From.Address),
		To:                types.NewAddress(txOrigin.To.Address),
		Value:             value,
		Gas:               gas,
		GasUsed:           gasUsed,
		GasPrice:          gasPrice,
		CumulativeGasUsed: cumulativeGasUsed,
		CreatedContract:   types.NewAddress(txOrigin.CreatedContract.Address),
		Data:              types.NewHexData(txOrigin.InputData),
		PrivateData:       types.NewHexData(txOrigin.PrivateInputData),
		IsPrivate:         txOrigin.IsPrivate,
		Timestamp:         block.Timestamp,
	}
	events := []*types.Event{}
	for _, l := range txOrigin.Logs {
		topics := []types.Hash{}
		for _, t := range l.Topics {
			topics = append(topics, types.NewHash(t))
		}
		e := &types.Event{
			Index:            l.Index,
			Address:          types.NewAddress(l.Account.Address),
			Topics:           topics,
			Data:             types.NewHexData(l.Data),
			BlockNumber:      block.Number,
			BlockHash:        block.Hash,
			TransactionHash:  tx.Hash,
			TransactionIndex: txOrigin.Index,
			Timestamp:        block.Timestamp,
		}
		events = append(events, e)
	}
	tx.Events = events

	var traceResp map[string]interface{}
	traceResp, err = client.TraceTransaction(tm.quorumClient, types.NewHash(tx.Hash.Hex()))
	if err != nil {
		return nil, err
	}
	if traceResp["calls"] != nil {
		respCalls := traceResp["calls"].([]interface{})
		tx.InternalCalls = make([]*types.InternalCall, len(respCalls))
		for i, respCall := range respCalls {
			respCallMap := respCall.(map[string]interface{})
			gas, err := strconv.ParseUint(respCallMap["gas"].(string), 0, 64)
			if err != nil {
				return nil, err
			}
			gasUsed, err := strconv.ParseUint(respCallMap["gasUsed"].(string), 0, 64)
			if err != nil {
				return nil, err
			}
			value = uint64(0)
			if val, ok := respCallMap["value"].(string); ok {
				value, err = strconv.ParseUint(val, 0, 64)
				if err != nil {
					return nil, err
				}
			}
			tx.InternalCalls[i] = &types.InternalCall{
				From:    types.NewAddress(respCallMap["from"].(string)),
				To:      types.NewAddress(respCallMap["to"].(string)),
				Gas:     gas,
				GasUsed: gasUsed,
				Value:   value,
				Input:   types.NewHexData(respCallMap["input"].(string)),
				Output:  types.NewHexData(respCallMap["output"].(string)),
				Type:    respCallMap["type"].(string),
			}
		}
	}
	return tx, nil
}
