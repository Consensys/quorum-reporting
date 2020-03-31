package types

import (
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
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
	PublicState  *state.Dump   `json:"publicState"`
	PrivateState *state.Dump   `json:"privateState"`
}

type Transaction struct {
	Hash              common.Hash            `json:"hash"`
	Status            bool                   `json:"status"`
	BlockNumber       uint64                 `json:"blockNumber"`
	Index             uint64                 `json:"index"`
	Nonce             uint64                 `json:"nonce"`
	From              common.Address         `json:"from"`
	To                common.Address         `json:"to"`
	Value             uint64                 `json:"value"`
	Gas               uint64                 `json:"gas"`
	GasUsed           uint64                 `json:"gasUsed"`
	CumulativeGasUsed uint64                 `json:"cumulativeGasUsed"`
	CreatedContract   common.Address         `json:"createdContract"`
	Data              hexutil.Bytes          `json:"data"`
	PrivateData       hexutil.Bytes          `json:"privateData"`
	IsPrivate         bool                   `json:"isPrivate"`
	Events            []*Event               `json:"events"`
	Sig               string                 `json:"txSig"`
	Func4Bytes        hexutil.Bytes          `json:"func4Bytes"`
	ParsedData        map[string]interface{} `json:"parsedData"`
}

func (tx *Transaction) ParseTransaction(abi *abi.ABI) {
	log.Printf("Parse transaction %v.\n", tx.Hash.Hex())
	// set defaults
	var data []byte
	if len(tx.PrivateData) > 0 {
		data = tx.PrivateData
	} else {
		data = tx.Data
	}
	tx.ParsedData = map[string]interface{}{}
	// parse transaction data
	if tx.To != (common.Address{0}) {
		tx.Func4Bytes = data[:4]
		// check against all abi methods
		for _, method := range abi.Methods {
			if string(method.ID()) == string(tx.Func4Bytes) {
				tx.Sig = method.Sig()
				method.Inputs.UnpackIntoMap(tx.ParsedData, data[4:])
				break
			}
		}
	} else {
		// contract deployment transaction
		tx.Sig = "constructor" + abi.Constructor.Sig()
		if len(data) > 32*abi.Constructor.Inputs.LengthNonIndexed() {
			abi.Constructor.Inputs.UnpackIntoMap(tx.ParsedData, data[(len(data)-32*abi.Constructor.Inputs.LengthNonIndexed()):])
			// TODO: parsing inputs for complex data type in constructor is not supported
		}
	}
	// parse events
	for _, e := range tx.Events {
		e.ParseEvent(abi)
	}
}

type Event struct {
	Index           uint64
	Address         common.Address         `json:"address"`
	Topics          []common.Hash          `json:"topics"`
	Data            hexutil.Bytes          `json:"data"`
	BlockNumber     uint64                 `json:"blockNumber"`
	TransactionHash common.Hash            `json:"transactionHash"`
	Sig             string                 `json:"eventSig"`
	ParsedData      map[string]interface{} `json:"parsedData"`
}

func (e *Event) ParseEvent(abi *abi.ABI) {
	log.Printf("Parse event %v.\n", e.Topics[0].Hex())
	eventABI, err := abi.EventByID(e.Topics[0])
	if err == nil {
		e.Sig = eventABI.String()
		e.ParsedData = map[string]interface{}{}
		eventABI.Inputs.UnpackIntoMap(e.ParsedData, e.Data)
	}
}
