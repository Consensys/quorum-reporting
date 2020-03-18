package types

import (
	"github.com/ethereum/go-ethereum/common"
)

type Block struct {
	Hash        common.Hash `json:"hash"`
	ParentHash  common.Hash `json:"parentHash"`
	StateRoot   common.Hash `json:"stateRoot"`
	TxRoot      common.Hash `json:"txRoot"`
	ReceiptRoot common.Hash `json:"receiptRoot"`
	Number      uint64      `json:"number"`
	GasLimit    uint64      `json:"gasLimit"`
	GasUsed     uint64      `json:"gasUsed"`
	Timestamp   uint64      `json:"timestamp"`
	ExtraData   []byte      `json:"extraData"`
}

type Transaction struct {
	TxHash  common.Hash    `json:"txHash"`
	From    common.Address `json:"from"`
	To      common.Address `json:"to"`
	Data []byte `json:"data"`
}

type Event struct {
	Address common.Address `json:"address"`
	BlockHash common.Hash `json:"blockHash"`
	BlockNumber uint64 `json:"blockNumber"`
	Data []byte `json:"data"`
	Topics []common.Hash `json:"topics"`
	TransactionHash common.Hash `json:"transactionHash"`
	TransactionIndex uint64 `json:"transactionIndex"`
}
