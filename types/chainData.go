package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Block struct {
	Hash        common.Hash `json:"hash"`
	ParentHash  common.Hash `json:"parentHash"`
	StateRoot   common.Hash `json:"stateRoot"`
	TxRoot      common.Hash `json:"txRoot"`
	ReceiptRoot common.Hash `json:"receiptRoot"`
	Number      *big.Int    `json:"number"`
	GasLimit    uint64      `json:"gasLimit"`
	GasUsed     uint64      `json:"gasUsed"`
	Timestamp   uint64      `json:"timestamp"`
	ExtraData   []byte      `json:"extraData"`
}

// TODO: define event type, transaction type, storage type
