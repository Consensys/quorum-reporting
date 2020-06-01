package elasticsearch

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"quorumengineering/quorum-report/types"
)

type Contract struct {
	Address             common.Address `json:"address"`
	ABI                 string         `json:"abi"`
	StorageABI          string         `json:"storageAbi"`
	CreationTransaction common.Hash    `json:"creationTx"`
	LastFiltered        uint64         `json:"lastFiltered"`
}

type State struct {
	Address     common.Address `json:"address"`
	BlockNumber uint64         `json:"blockNumber"`
	StorageRoot common.Hash    `json:"storageRoot"`
}

type Storage struct {
	StorageRoot common.Hash            `json:"storageRoot"`
	StorageMap  map[common.Hash]string `json:"storageMap"`
}

type Event struct {
	Address          common.Address `json:"address"`
	BlockHash        common.Hash    `json:"blockHash"`
	BlockNumber      uint64         `json:"blockNumber"`
	Data             hexutil.Bytes  `json:"data"`
	LogIndex         uint64         `json:"logIndex"`
	Topics           []common.Hash  `json:"topics"`
	TransactionHash  common.Hash    `json:"transactionHash"`
	TransactionIndex uint64         `json:"transactionIndex"`
	Timestamp        uint64         `json:"timestamp"`
}

func (e *Event) From(event *types.Event) {
	e.Address = event.Address
	e.BlockNumber = event.BlockNumber
	e.Data = event.Data
	e.LogIndex = event.Index
	e.Topics = event.Topics
	e.TransactionHash = event.TransactionHash
	e.BlockHash = event.BlockHash
	e.TransactionIndex = event.TransactionIndex
	e.Timestamp = event.Timestamp
}

func (e *Event) To() *types.Event {
	return &types.Event{
		Index:            e.LogIndex,
		Address:          e.Address,
		Topics:           e.Topics,
		Data:             e.Data,
		BlockNumber:      e.BlockNumber,
		TransactionHash:  e.TransactionHash,
		BlockHash:        e.BlockHash,
		TransactionIndex: e.TransactionIndex,
		Timestamp:        e.Timestamp,
	}
}

type Transaction struct {
	Hash              common.Hash     `json:"hash"`
	Status            bool            `json:"status"`
	BlockNumber       uint64          `json:"blockNumber"`
	BlockHash         common.Hash     `json:"blockHash"`
	Index             uint64          `json:"index"`
	Nonce             uint64          `json:"nonce"`
	Sender            common.Address  `json:"from"`
	Recipient         common.Address  `json:"to"`
	Value             uint64          `json:"value"`
	Gas               uint64          `json:"gas"`
	GasPrice          uint64          `json:"gasPrice"`
	GasUsed           uint64          `json:"gasUsed"`
	CumulativeGasUsed uint64          `json:"cumulativeGasUsed"`
	CreatedContract   common.Address  `json:"createdContract"`
	Data              hexutil.Bytes   `json:"data"`
	PrivateData       hexutil.Bytes   `json:"privateData"`
	IsPrivate         bool            `json:"isPrivate"`
	Events            []*Event        `json:"events"`
	InternalCalls     []*InternalCall `json:"internalCalls"`
	Timestamp         uint64          `json:"timestamp"`
}

func (t *Transaction) To() *types.Transaction {
	var internalCalls []*types.InternalCall
	for _, call := range t.InternalCalls {
		internalCalls = append(internalCalls, call.To())
	}

	var events []*types.Event
	for _, ev := range t.Events {
		events = append(events, ev.To())
	}

	return &types.Transaction{
		Hash:              t.Hash,
		Status:            t.Status,
		BlockNumber:       t.BlockNumber,
		BlockHash:         t.BlockHash,
		Index:             t.Index,
		Nonce:             t.Nonce,
		From:              t.Sender,
		To:                t.Recipient,
		Value:             t.Value,
		Gas:               t.Gas,
		GasPrice:          t.GasPrice,
		GasUsed:           t.GasUsed,
		CumulativeGasUsed: t.CumulativeGasUsed,
		CreatedContract:   t.CreatedContract,
		Data:              t.Data,
		PrivateData:       t.PrivateData,
		IsPrivate:         t.IsPrivate,
		Timestamp:         t.Timestamp,
		Events:            events,
		InternalCalls:     internalCalls,
	}
}

func (t *Transaction) From(tx *types.Transaction) {
	internalCalls := make([]*InternalCall, 0)
	for _, call := range tx.InternalCalls {
		var ic InternalCall
		ic.From(call)
		internalCalls = append(internalCalls, &ic)
	}

	events := make([]*Event, 0)
	for _, ev := range tx.Events {
		var event Event
		event.From(ev)
		events = append(events, &event)
	}

	t.Hash = tx.Hash
	t.Status = tx.Status
	t.BlockNumber = tx.BlockNumber
	t.BlockHash = tx.BlockHash
	t.Index = tx.Index
	t.Nonce = tx.Nonce
	t.Sender = tx.From
	t.Recipient = tx.To
	t.Value = tx.Value
	t.Gas = tx.Gas
	t.GasPrice = tx.GasPrice
	t.GasUsed = tx.GasUsed
	t.CumulativeGasUsed = tx.CumulativeGasUsed
	t.CreatedContract = tx.CreatedContract
	t.Data = tx.Data
	t.PrivateData = tx.PrivateData
	t.IsPrivate = tx.IsPrivate
	t.Timestamp = tx.Timestamp
	t.Events = events
	t.InternalCalls = internalCalls
}

type InternalCall struct {
	Sender    common.Address `json:"from"`
	Recipient common.Address `json:"to"`
	Gas       uint64         `json:"gas"`
	GasUsed   uint64         `json:"gasUsed"`
	Value     uint64         `json:"value"`
	Input     hexutil.Bytes  `json:"input"`
	Output    hexutil.Bytes  `json:"output"`
	Type      string         `json:"type"`
}

func (ic *InternalCall) To() *types.InternalCall {
	return &types.InternalCall{
		From:    ic.Sender,
		To:      ic.Recipient,
		Gas:     ic.Gas,
		GasUsed: ic.GasUsed,
		Value:   ic.Value,
		Input:   ic.Input,
		Output:  ic.Output,
		Type:    ic.Type,
	}
}

func (ic *InternalCall) From(internalCall *types.InternalCall) {
	ic.Sender = internalCall.From
	ic.Recipient = internalCall.To
	ic.Gas = internalCall.Gas
	ic.GasUsed = internalCall.GasUsed
	ic.Value = internalCall.Value
	ic.Input = internalCall.Input
	ic.Output = internalCall.Output
	ic.Type = internalCall.Type
}

type Block struct {
	Hash         common.Hash   `json:"hash"`
	ParentHash   common.Hash   `json:"parentHash"`
	StateRoot    common.Hash   `json:"stateRoot"`
	TxRoot       common.Hash   `json:"txRoot"`
	ReceiptRoot  common.Hash   `json:"receiptRoot"`
	Number       uint64        `json:"number"`
	GasLimit     uint64        `json:"gasLimit"`
	GasUsed      uint64        `json:"gasUsed"`
	Timestamp    uint64        `json:"timestamp"`
	ExtraData    hexutil.Bytes `json:"extraData"`
	Transactions []common.Hash `json:"transactions"`
}

func (b *Block) To() *types.Block {
	return &types.Block{
		Hash:         b.Hash,
		ParentHash:   b.ParentHash,
		StateRoot:    b.StateRoot,
		TxRoot:       b.TxRoot,
		ReceiptRoot:  b.ReceiptRoot,
		Number:       b.Number,
		GasLimit:     b.GasLimit,
		GasUsed:      b.GasUsed,
		Timestamp:    b.Timestamp,
		ExtraData:    b.ExtraData,
		Transactions: b.Transactions,
	}
}

func (b *Block) From(block *types.Block) {
	b.Hash = block.Hash
	b.ParentHash = block.ParentHash
	b.StateRoot = block.StateRoot
	b.TxRoot = block.TxRoot
	b.ReceiptRoot = block.ReceiptRoot
	b.Number = block.Number
	b.GasLimit = block.GasLimit
	b.GasUsed = block.GasUsed
	b.Timestamp = block.Timestamp
	b.ExtraData = block.ExtraData
	b.Transactions = block.Transactions
}

//

type ContractQueryResult struct {
	Source Contract `json:"_source"`
}

type TransactionQueryResult struct {
	Source Transaction `json:"_source"`
}

type BlockQueryResult struct {
	Source Block `json:"_source"`
}

type StateQueryResult struct {
	Source State `json:"_source"`
}

type StorageQueryResult struct {
	Source Storage `json:"_source"`
}

type LastPersistedResult struct {
	Source struct {
		LastPersisted uint64 `json:"lastPersisted"`
	} `json:"_source"`
}

type SearchQueryResult struct {
	Hits struct {
		Hits []IndividualResult `json:"hits"`
	} `json:"hits"`
}

type CountQueryResult struct {
	Count uint64 `json:"count"`
}

type IndividualResult struct {
	Id     string                 `json:"_id"`
	Source map[string]interface{} `json:"_source"`
}
