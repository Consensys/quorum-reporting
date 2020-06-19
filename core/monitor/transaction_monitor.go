package monitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type TransactionMonitor struct {
	db           database.Database
	quorumClient client.Client
}

func NewTransactionMonitor(db database.Database, quorumClient client.Client) *TransactionMonitor {
	return &TransactionMonitor{db, quorumClient}
}

func (tm *TransactionMonitor) PullTransactions(block *types.Block) ([]*types.Transaction, error) {
	log.Info("Fetching transactions", "block", block.Hash.String())

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

func (tm *TransactionMonitor) createTransaction(block *types.Block, hash common.Hash) (*types.Transaction, error) {
	log.Debug("Processing transaction", "hash", hash.String())

	var txResult graphql.TransactionResult
	if err := tm.quorumClient.ExecuteGraphQLQuery(&txResult, graphql.TransactionDetailQuery(hash)); err != nil {
		return nil, err
	}

	txOrigin := txResult.Transaction

	// Create reporting transaction struct fields.
	nonce, err := hexutil.DecodeUint64(txOrigin.Nonce)
	if err != nil {
		return nil, err
	}
	value, err := hexutil.DecodeUint64(txOrigin.Value)
	if err != nil {
		return nil, err
	}
	gas, err := hexutil.DecodeUint64(txOrigin.Gas)
	if err != nil {
		return nil, err
	}
	gasUsed, err := hexutil.DecodeUint64(txOrigin.GasUsed)
	if err != nil {
		return nil, err
	}
	cumulativeGasUsed, err := hexutil.DecodeUint64(txOrigin.CumulativeGasUsed)
	if err != nil {
		return nil, err
	}
	gasPrice, err := hexutil.DecodeUint64(txOrigin.GasPrice)
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
		From:              common.HexToAddress(txOrigin.From.Address),
		To:                common.HexToAddress(txOrigin.To.Address),
		Value:             value,
		Gas:               gas,
		GasUsed:           gasUsed,
		GasPrice:          gasPrice,
		CumulativeGasUsed: cumulativeGasUsed,
		CreatedContract:   common.HexToAddress(txOrigin.CreatedContract.Address),
		Data:              hexutil.MustDecode(txOrigin.InputData),
		PrivateData:       hexutil.MustDecode(txOrigin.PrivateInputData),
		IsPrivate:         txOrigin.IsPrivate,
		Timestamp:         block.Timestamp,
	}
	events := []*types.Event{}
	for _, l := range txOrigin.Logs {
		topics := []common.Hash{}
		for _, t := range l.Topics {
			topics = append(topics, common.HexToHash(t))
		}
		e := &types.Event{
			Index:            l.Index,
			Address:          common.HexToAddress(l.Account.Address),
			Topics:           topics,
			Data:             hexutil.MustDecode(l.Data),
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
	traceResp, err = client.TraceTransaction(tm.quorumClient, tx.Hash)
	if err != nil {
		return nil, err
	}
	if traceResp["calls"] != nil {
		respCalls := traceResp["calls"].([]interface{})
		tx.InternalCalls = make([]*types.InternalCall, len(respCalls))
		for i, respCall := range respCalls {
			respCallMap := respCall.(map[string]interface{})
			gas, err := hexutil.DecodeUint64(respCallMap["gas"].(string))
			if err != nil {
				return nil, err
			}
			gasUsed, err := hexutil.DecodeUint64(respCallMap["gasUsed"].(string))
			if err != nil {
				return nil, err
			}
			value = uint64(0)
			if val, ok := respCallMap["value"].(string); ok {
				value, err = hexutil.DecodeUint64(val)
				if err != nil {
					return nil, err
				}
			}
			tx.InternalCalls[i] = &types.InternalCall{
				From:    common.HexToAddress(respCallMap["from"].(string)),
				To:      common.HexToAddress(respCallMap["to"].(string)),
				Gas:     gas,
				GasUsed: gasUsed,
				Value:   value,
				Input:   hexutil.MustDecode(respCallMap["input"].(string)),
				Output:  hexutil.MustDecode(respCallMap["output"].(string)),
				Type:    respCallMap["type"].(string),
			}
		}
	}
	return tx, nil
}
