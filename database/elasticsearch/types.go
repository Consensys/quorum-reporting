package elasticsearch

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"quorumengineering/quorum-report/types"
)

type Contract struct {
	Address             common.Address `json:"address"`
	ABI                 string         `json:"abi"`
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
	Address         common.Address `json:"address"`
	BlockNumber     uint64         `json:"blockNumber"`
	Data            hexutil.Bytes  `json:"data"`
	LogIndex        uint64         `json:"logIndex"`
	Topics          []common.Hash  `json:"topics"`
	TransactionHash common.Hash    `json:"transactionHash"`
}

func (e Event) From(event *types.Event) {

}

func (e Event) To() *types.Event {
	return &types.Event{
		Index:           e.LogIndex,
		Address:         e.Address,
		Topics:          e.Topics,
		Data:            e.Data,
		BlockNumber:     e.BlockNumber,
		TransactionHash: e.TransactionHash,
	}
}

type Transaction struct {
	Hash        common.Hash    `json:"hash"`
	BlockHash   common.Hash    `json:"blockHash"`
	BlockNumber uint64         `json:"blockNumber"`
	From        common.Address `json:"from"`
	Gas         uint64         `json:"gas"`
	GasPrice    uint64         `json:"gasPrice"`
	Data        hexutil.Bytes  `json:"data"`
	Nonce       uint64         `json:"nonce"`
	To          common.Address `json:"to"`
	Index       uint64         `json:"index"`
	Value       uint64         `json:"value"`
	IsPrivate   bool           `json:"isPrivate"`
	PrivateData hexutil.Bytes  `json:"privateData"`

	Receipt Receipt `json:"receipt"`

	InternalCalls []*InternalCall `json:"internalCalls"`
}

type Receipt struct {
	ContractAddress   common.Address `json:"contractAddress"`
	CumulativeGasUsed uint64         `json:"cumulativeGasUsed"`
	GasUsed           uint64         `json:"gasUsed"`
	Events            []*Event       `json:"events"`
	LogsBloom         []byte         `json:"logsBloom"`
	Status            bool           `json:"status"`
	Root              common.Hash    `json:"root"`
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

//

type ContractQueryResult struct {
	Source Contract `json:"_source"`
}

type TransactionQueryResult struct {
	Source types.Transaction `json:"_source"`
}

type BlockQueryResult struct {
	Source types.Block `json:"_source"`
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
