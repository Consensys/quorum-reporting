package types

import (
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi"
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
	Parsed            string         `json:"parsedTx"`
}

func (tx *Transaction) ParseTransaction(abi *abi.ABI) {
	log.Printf("Parse transaction %v.\n", tx.Hash.Hex())
	// parse transaction data
	if tx.To != (common.Address{0}) {
		var data []byte
		if len(tx.PrivateData) > 0 {
			data = tx.PrivateData[:4]
		} else {
			data = tx.Data[:4]
		}
		for _, method := range abi.Methods {
			if string(method.ID()) == string(data) {
				tx.Parsed = method.Sig()
				break
			}
		}
	} else {
		tx.Parsed = "contract deployment transaction"
	}
	// parse events
	for _, e := range tx.Events {
		e.ParseEvent(abi)
	}
}

type Event struct {
	Index           uint64
	Address         common.Address `json:"address"`
	Topics          []common.Hash  `json:"topics"`
	Data            hexutil.Bytes  `json:"data"`
	BlockNumber     uint64         `json:"blockNumber"`
	TransactionHash common.Hash    `json:"transactionHash"`
	Parsed          string         `json:"parsedEvent"`
}

func (e *Event) ParseEvent(abi *abi.ABI) {
	log.Printf("Parse event %v.\n", e.Topics[0].Hex())
	eventABI, err := abi.EventByID(e.Topics[0])
	if err == nil {
		e.Parsed = eventABI.String()
	}
}
