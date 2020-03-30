package rpc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

func TestAPIValidation(t *testing.T) {
	var err error
	db := database.NewMemoryDB()
	apis := NewRPCAPIs(db)
	// Test AddAddress validation
	err = apis.AddAddress(common.Address{0})
	if err.Error() != "invalid input" {
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
			{Address: address, Topics: []common.Hash{common.HexToHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36")}},
		},
	}
)

func TestAPIParsing(t *testing.T) {
	var err error
	db := database.NewMemoryDB()
	apis := NewRPCAPIs(db)
	err = apis.AddAddress(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	// Test AddContractABI string to ABI parsing.
	err = apis.AddContractABI(address, "hello")
	if err.Error() != "invalid character 'h' looking for beginning of value" {
		t.Fatalf("expected %v, but got %v", "invalid input", err)
	}
	err = apis.AddContractABI(address, validABI)
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
	if parsedTx1.Parsed != "contract deployment transaction" {
		t.Fatalf("expected %v, but got %v", "contract deployment transaction", err)
	}
	parsedTx2, err := apis.GetTransaction(tx2.Hash)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedTx2.Parsed != "set(uint256)" {
		t.Fatalf("expected %v, but got %v", "set(uint256)", err)
	}
	parsedTx3, err := apis.GetTransaction(tx3.Hash)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedTx3.Parsed != "set(uint256)" {
		t.Fatalf("expected %v, but got %v", "set(uint256)", err)
	}
	// Test GetAllEventsByAddress parse event.
	err = db.IndexBlock(address, block)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	parsedEvents, err := apis.GetAllEventsByAddress(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if parsedEvents[0].Parsed != "event valueSet(uint256 _value)" {
		t.Fatalf("expected %v, but got %v", "set(uint256)", err)
	}
}
