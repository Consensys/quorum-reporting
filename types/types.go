package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Template struct {
	TemplateName  string `json:"templateName"`
	ABI           string `json:"abi"`
	StorageLayout string `json:"storageLayout"`
}

// received from eth_getBlockByNumber
type RawBlock struct {
	Hash         string   `json:"hash"`
	ParentHash   string   `json:"parentHash"`
	StateRoot    string   `json:"stateRoot"`
	TxRoot       string   `json:"txRoot"`
	ReceiptRoot  string   `json:"receiptRoot"`
	Number       string   `json:"number"`
	GasLimit     string   `json:"gasLimit"`
	GasUsed      string   `json:"gasUsed"`
	Timestamp    string   `json:"timestamp"`
	ExtraData    string   `json:"extraData"`
	Transactions []string `json:"transactions"`
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
	ExtraData    string        `json:"extraData"`
	Transactions []common.Hash `json:"transactions"`
}

type Transaction struct {
	Hash              common.Hash     `json:"hash"`
	Status            bool            `json:"status"`
	BlockNumber       uint64          `json:"blockNumber"`
	BlockHash         common.Hash     `json:"blockHash"`
	Index             uint64          `json:"index"`
	Nonce             uint64          `json:"nonce"`
	From              common.Address  `json:"from"`
	To                common.Address  `json:"to"`
	Value             uint64          `json:"value"`
	Gas               uint64          `json:"gas"`
	GasPrice          uint64          `json:"gasPrice"`
	GasUsed           uint64          `json:"gasUsed"`
	CumulativeGasUsed uint64          `json:"cumulativeGasUsed"`
	CreatedContract   common.Address  `json:"createdContract"`
	Data              hexutil.Bytes   `json:"data"`
	PrivateData       hexutil.Bytes   `json:"privateData"`
	IsPrivate         bool            `json:"isPrivate"`
	Timestamp         uint64          `json:"timestamp"`
	Events            []*Event        `json:"events"`
	InternalCalls     []*InternalCall `json:"internalCalls"`
}

type InternalCall struct {
	From    common.Address `json:"from"`
	To      common.Address `json:"to"`
	Gas     uint64         `json:"gas"`
	GasUsed uint64         `json:"gasUsed"`
	Value   uint64         `json:"value"`
	Input   hexutil.Bytes  `json:"input"`
	Output  hexutil.Bytes  `json:"output"`
	Type    string         `json:"type"`
}

type Event struct {
	Index            uint64         `json:"index"`
	Address          common.Address `json:"address"`
	Topics           []common.Hash  `json:"topics"`
	Data             hexutil.Bytes  `json:"data"`
	BlockNumber      uint64         `json:"blockNumber"`
	BlockHash        common.Hash    `json:"blockHash"`
	TransactionHash  common.Hash    `json:"transactionHash"`
	TransactionIndex uint64         `json:"transactionIndex"`
	Timestamp        uint64         `json:"timestamp"`
}
