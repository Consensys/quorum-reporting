package rpc

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/database/memory"
	"quorumengineering/quorum-report/types"
)

func TestAPIValidation(t *testing.T) {
	var err error
	db := memory.NewMemoryDB()
	apis := NewRPCAPIs(db)
	// Test AddAddress validation
	err = apis.AddAddress(common.Address{0})
	if err == nil || err.Error() != "invalid input" {
		t.Fatalf("expected %v, but got %v", "invalid input", err)
	}
}

const validABI = `
[
	{"constant":true,"inputs":[],"name":"storedData","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
	{"constant":false,"inputs":[{"name":"_x","type":"uint256"}],"name":"set","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},
	{"constant":true,"inputs":[],"name":"get","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
	{"inputs":[{"name":"_initVal","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},
	{"anonymous":false,"inputs":[{"indexed":false,"name":"_value","type":"uint256"}],"name":"valueSet","type":"event"}
]`

var (
	address = common.HexToAddress("0x0000000000000000000000000000000000000001")
	block   = &types.Block{
		Hash:   common.BytesToHash([]byte("dummy")),
		Number: 1,
		Transactions: []common.Hash{
			common.BytesToHash([]byte("tx1")), common.BytesToHash([]byte("tx2")), common.BytesToHash([]byte("tx3")),
		},
	}
	tx1 = &types.Transaction{ // deployment
		Hash:            common.BytesToHash([]byte("tx1")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              common.Address{0},
		Data:            hexutil.MustDecode("0x608060405234801561001057600080fd5b506040516020806101a18339810180604052602081101561003057600080fd5b81019080805190602001909291905050508060008190555050610149806100586000396000f3fe608060405234801561001057600080fd5b506004361061005e576000357c0100000000000000000000000000000000000000000000000000000000900480632a1afcd91461006357806360fe47b1146100815780636d4ce63c146100af575b600080fd5b61006b6100cd565b6040518082815260200191505060405180910390f35b6100ad6004803603602081101561009757600080fd5b81019080803590602001909291905050506100d3565b005b6100b7610114565b6040518082815260200191505060405180910390f35b60005481565b806000819055507fefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36816040518082815260200191505060405180910390a150565b6000805490509056fea165627a7a7230582061f6956b053dbf99873b363ab3ba7bca70853ba5efbaff898cd840d71c54fc1d0029000000000000000000000000000000000000000000000000000000000000002a"),
		CreatedContract: address,
	}
	tx2 = &types.Transaction{ // set
		Hash:            common.BytesToHash([]byte("tx2")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              address,
		Data:            hexutil.MustDecode("0x60fe47b100000000000000000000000000000000000000000000000000000000000003e7"),
		CreatedContract: common.Address{0},
	}
	tx3 = &types.Transaction{ // private
		Hash:            common.BytesToHash([]byte("tx3")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              address,
		PrivateData:     hexutil.MustDecode("0x60fe47b100000000000000000000000000000000000000000000000000000000000003e8"),
		CreatedContract: common.Address{0},
		Events: []*types.Event{
			{
				Data:    hexutil.MustDecode("0x00000000000000000000000000000000000000000000000000000000000003e8"),
				Address: address,
				Topics:  []common.Hash{common.HexToHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36")},
			},
		},
	}
)

func TestAPIParsing(t *testing.T) {
	var err error
	db := memory.NewMemoryDB()
	apis := NewRPCAPIs(db)
	err = apis.AddAddress(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	// Test AddABI string to ABI parsing.
	err = apis.AddABI(address, "hello")
	if err == nil || err.Error() != "invalid character 'h' looking for beginning of value" {
		t.Fatalf("expected %v, but got %v", "invalid input", err)
	}
	err = apis.AddABI(address, validABI)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	// Set up test data.
	err = db.WriteTransaction(tx1)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	err = db.WriteTransaction(tx2)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	err = db.WriteTransaction(tx3)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	err = db.WriteBlock(block)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	// Test GetTransaction parse transaction data.
	parsedTx1, err := apis.GetTransaction(tx1.Hash)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedTx1.Sig != "constructor(uint256)" {
		t.Fatalf("expected %v, but got %v", "constructor(uint256)", parsedTx1.Sig)
	}
	if parsedTx1.ParsedData["_initVal"].(*big.Int).Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("expected %v, but got %v", 42, parsedTx1.ParsedData["_initVal"])
	}
	parsedTx2, err := apis.GetTransaction(tx2.Hash)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedTx2.Sig != "set(uint256)" {
		t.Fatalf("expected %v, but got %v", "set(uint256)", parsedTx2.Sig)
	}
	if parsedTx2.ParsedData["_x"].(*big.Int).Cmp(big.NewInt(999)) != 0 {
		t.Fatalf("expected %v, but got %v", 999, parsedTx2.ParsedData["_x"])
	}
	if parsedTx2.Func4Bytes.String() != "0x60fe47b1" {
		t.Fatalf("expected %v, but got %v", "0x60fe47b1", parsedTx2.Func4Bytes.String())
	}
	parsedTx3, err := apis.GetTransaction(tx3.Hash)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedTx3.ParsedEvents[0].Sig != "event valueSet(uint256 _value)" {
		t.Fatalf("expected %v, but got %v", "event valueSet(uint256 _value)", parsedTx3.ParsedEvents[0].Sig)
	}
	if parsedTx3.ParsedEvents[0].ParsedData["_value"].(*big.Int).Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("expected %v, but got %v", 1000, parsedTx3.ParsedEvents[0].ParsedData["_value"])
	}
	// Test GetAllEventsByAddress parse event.
	err = db.IndexBlock([]common.Address{address}, block)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	parsedEvents, err := apis.GetAllEventsByAddress(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedEvents[0].Sig != "event valueSet(uint256 _value)" {
		t.Fatalf("expected %v, but got %v", "event valueSet(uint256 _value)", parsedEvents[0].Sig)
	}
	if parsedEvents[0].ParsedData["_value"].(*big.Int).Cmp(big.NewInt(1000)) != 0 {
		t.Fatalf("expected %v, but got %v", 1000, parsedEvents[0].ParsedData["_value"])
	}
}
