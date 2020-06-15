package monitor

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/mitchellh/mapstructure"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
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
	log.Printf("Pull all transactions for block %v.\n", block.Number)

	fetchedTransactions := make([]*types.Transaction, 0, len(block.Transactions))
	for _, txHash := range block.Transactions {
		// 1. Query transaction details by graphql.
		tx, err := tm.createTransaction(block, txHash)
		if err != nil {
			return nil, err
		}
		var addrs []common.Address
		addrs = append(addrs, tx.CreatedContract)
		for _, ic := range tx.InternalCalls {
			if ic.Type == "CREATE" || ic.Type == "CREATE2" {
				addrs = append(addrs, ic.To)
			}
		}

		for _, addr := range addrs {
			res, err := client.GetCode(tm.quorumClient, addr, tx.BlockHash)
			if err != nil {
				return nil, err
			}

			// 2. Check if transaction deploys a public ERC20 contract
			if checkAbiMatch(types.ERC20ABI, res) {
				log.Printf("tx %v deploys %v which is a potential ERC20 contract.\n", tx.Hash.Hex(), addr.Hex())
				// add contract address
				tm.db.AddAddresses([]common.Address{tx.CreatedContract})
				// assign ERC20 template
				tm.db.AssignTemplate(tx.CreatedContract, types.ERC20)
			}

			if checkAbiMatch(types.ERC721ABI, res) {
				log.Printf("tx %v deploys %v which is a potential ERC721 contract.\n", tx.Hash.Hex(), addr.Hex())
				// add contract address
				tm.db.AddAddresses([]common.Address{tx.CreatedContract})
				// assign ERC20 template
				tm.db.AssignTemplate(tx.CreatedContract, types.ERC20)
			}
		}
		fetchedTransactions = append(fetchedTransactions, tx)
	}
	return fetchedTransactions, nil
}

func checkAbiMatch(abiToCheck abi.ABI, data hexutil.Bytes) bool {
	for _, b := range abiToCheck.Methods {
		if !strings.Contains(data.String(), common.Bytes2Hex(b.ID())) {
			return false
		}
	}
	for _, event := range abiToCheck.Events {
		if !strings.Contains(data.String(), event.ID().Hex()[2:]) {
			return false
		}
	}
	return true
}

func (tm *TransactionMonitor) createTransaction(block *types.Block, hash common.Hash) (*types.Transaction, error) {
	var (
		resp     map[string]interface{}
		txOrigin graphql.Transaction
	)
	err := tm.quorumClient.ExecuteGraphQLQuery(context.Background(), &resp, graphql.TransactionDetailQuery(hash))
	if err != nil {
		// TODO: if quorum node is down, reconnect?
		return nil, err
	}

	if err = mapstructure.Decode(resp["transaction"].(map[string]interface{}), &txOrigin); err != nil {
		return nil, err
	}

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

	resp, err = client.TraceTransaction(tm.quorumClient, tx.Hash)
	if err != nil {
		return nil, err
	}
	if resp["calls"] != nil {
		respCalls := resp["calls"].([]interface{})
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
