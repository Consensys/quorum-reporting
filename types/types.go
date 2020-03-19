package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

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

type Transaction struct {
	Hash              common.Hash    `json:"hash"`
	Status            bool           `json:"status"`
	BlockNumber       uint64         `json:"blockNumber"`
	Index             uint64         `json:"index"`
	Nonce             uint64         `json:"nonce"`
	From              common.Address `json:"from"`
	To                common.Address `json:"to"`
	Value             uint64         `json:"value"`
	Gas               uint64         `json:"gas"`
	GasUsed           uint64         `json:"gasUsed"`
	CumulativeGasUsed uint64         `json:"cumulativeGasUsed"`
	CreatedContract   common.Address `json:"createdContract"`
	Data              hexutil.Bytes  `json:"data"`
	PrivateData       hexutil.Bytes  `json:"privateData"`
	IsPrivate         bool           `json:"isPrivate"`
	Events            []*Event       `json:"events"`
}

type Event struct {
	Index           uint64
	Address         common.Address `json:"address"`
	Topics          []common.Hash  `json:"topics"`
	Data            hexutil.Bytes  `json:"data"`
	BlockNumber     uint64         `json:"blockNumber"`
	TransactionHash common.Hash    `json:"transactionHash"`
}
