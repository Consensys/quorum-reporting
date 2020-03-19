package filter

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mitchellh/mapstructure"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
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
	fmt.Printf("Filter block %v\n", block.Number)

	for _, txHash := range block.Transactions {
		// 1. Query transaction details by graphql
		tx, err := tf.createTransaction(txHash)
		if err != nil {
			// TODO: should gracefully handle error (if quorum node is down, reconnect?)
			log.Fatalf("get transaction details error: %v.\n", err)
		}
		fmt.Println(tx.Hash.Hex())
		// 2. Write transactions to DB
		err = tf.db.WriteTransaction(tx)
		if err != nil {
			// TODO: should gracefully handle error (if quorum node is down, reconnect?)
			log.Fatalf("write transaction error: %v.\n", err)
		}
		// TODO: 3. Index transactions related to registered contract addresses
		// TODO: 4. Index events related to registered contract addresses
	}
}

func (tf *TransactionFilter) createTransaction(hash common.Hash) (*types.Transaction, error) {
	var (
		resp     map[string]interface{}
		txOrigin graphql.Transaction
	)
	resp, err := tf.quorumClient.ExecuteGraphQLQuery(context.Background(), graphql.TransactionDetailQuery(hash))
	if err != nil {
		return nil, err
	}
	err = mapstructure.Decode(resp["transaction"].(map[string]interface{}), &txOrigin)
	if err != nil {
		return nil, err
	}

	// create reporting transaction struct fields
	blockNumber, err := hexutil.DecodeUint64(txOrigin.Block.Number)
	if err != nil {
		return nil, err
	}
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

	tx := &types.Transaction{
		Hash:              common.HexToHash(txOrigin.Hash),
		Status:            txOrigin.Status == "0x1",
		BlockNumber:       blockNumber,
		Index:             txOrigin.Index,
		Nonce:             nonce,
		From:              common.HexToAddress(txOrigin.From.Address),
		To:                common.HexToAddress(txOrigin.To.Address),
		Value:             value,
		Gas:               gas,
		GasUsed:           gasUsed,
		CumulativeGasUsed: cumulativeGasUsed,
		Data:              []byte(txOrigin.InputData),
		PrivateData:       []byte(txOrigin.PrivateInputData),
		IsPrivate:         txOrigin.IsPrivate,
	}
	events := []*types.Event{}
	for _, l := range txOrigin.Logs {
		topics := []common.Hash{}
		for _, t := range l.Topics {
			topics = append(topics, common.HexToHash(t))
		}
		e := &types.Event{
			Index:           l.Index,
			Address:         common.HexToAddress(l.Account.Address),
			Topics:          topics,
			Data:            []byte(l.Data),
			BlockNumber:     tx.BlockNumber,
			TransactionHash: tx.Hash,
		}
		events = append(events, e)
	}
	tx.Events = events
	return tx, nil
}
