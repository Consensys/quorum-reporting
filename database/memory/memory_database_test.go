package memory

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

const jsondata = `
[
	{ "type" : "function", "name" : "balance", "constant" : true },
	{ "type" : "function", "name" : "send", "constant" : false, "inputs" : [ { "name" : "amount", "type" : "uint256" } ] }
]`

func TestMemoryDB(t *testing.T) {
	// test data
	db := NewMemoryDB()
	address := common.HexToAddress("0x0000000000000000000000000000000000000001")
	uselessAddress := common.HexToAddress("0x0000000000000000000000000000000000000002")
	testABI, _ := abi.JSON(strings.NewReader(jsondata))
	block := &types.Block{
		Hash:   common.BytesToHash([]byte("dummy")),
		Number: 1,
		Transactions: []common.Hash{
			common.BytesToHash([]byte("tx1")), common.BytesToHash([]byte("tx2")), common.BytesToHash([]byte("tx3")),
		},
	}
	rawStorage := map[common.Address]*state.DumpAccount{
		address: {
			Storage: map[common.Hash]string{
				common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"): "2a",
				common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"): "2b",
			},
		},
	}
	tx1 := &types.Transaction{
		Hash:            common.BytesToHash([]byte("tx1")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              common.Address{0},
		Value:           666,
		CreatedContract: address,
	}
	tx2 := &types.Transaction{
		Hash:        common.BytesToHash([]byte("tx2")),
		BlockNumber: 1,
		From:        common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:          uselessAddress,
		Value:       666,
		InternalCalls: []*types.InternalCall{
			{
				To: address,
			},
		},
	}
	tx3 := &types.Transaction{
		Hash:        common.BytesToHash([]byte("tx3")),
		BlockNumber: 1,
		From:        common.HexToAddress("0x0000000000000000000000000000000000000010"),
		To:          address,
		Value:       666,
		Events: []*types.Event{
			{}, // dummy event
			{Address: address},
		},
	}
	// 1. Add an address and get it.
	testAddAddresses(t, db, []common.Address{address}, false)
	testGetAddresses(t, db, 1)
	// 2. Add Contract ABI and get it.
	testAddContractABI(t, db, address, jsondata, false)
	testGetContractABI(t, db, address, &testABI)
	// 3. Write transaction and get it.
	testWriteTransaction(t, db, tx1, false)
	testWriteTransaction(t, db, tx2, false)
	testWriteTransaction(t, db, tx3, false)
	testReadTransaction(t, db, tx1.Hash, tx1)
	// 4. Write block and get it. Check last persisted block number.
	testGetLastPersistedBlockNumeber(t, db, 0)
	testWriteBlock(t, db, block, false)
	testReadBlock(t, db, 1, block.Hash)
	testGetLastPersistedBlockNumeber(t, db, 1)
	// 5. Index block and check last filtered. Retrieve all transactions/ events.
	testGetLastFiltered(t, db, address, 0)
	testIndexStorage(t, db, 1, rawStorage)
	testIndexBlock(t, db, address, block)
	testGetLastFiltered(t, db, address, 1)
	testGetContractCreationTransaction(t, db, address, common.BytesToHash([]byte("tx1")))
	testGetAllTransactionsToAddress(t, db, address, 1)
	testGetAllTransactionsInternalToAddress(t, db, address, 1)
	testGetAllEventsByAddress(t, db, address, 1)
	testGetStorage(t, db, address, 1, 2)
	// 6. Delete address and check last filtered
	testDeleteAddress(t, db, address, false)
	testGetLastFiltered(t, db, address, 0)
}

func testAddAddresses(t *testing.T, db database.Database, addresses []common.Address, expectedErr bool) {
	err := db.AddAddresses(addresses)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testDeleteAddress(t *testing.T, db database.Database, address common.Address, expectedErr bool) {
	err := db.DeleteAddress(address)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testGetAddresses(t *testing.T, db database.Database, expected int) {
	actual, err := db.GetAddresses()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(actual) != expected {
		t.Fatalf("expected %v addresses, but got %v", expected, len(actual))
	}
}

func testAddContractABI(t *testing.T, db database.Database, address common.Address, contractABI string, expectedErr bool) {
	err := db.AddContractABI(address, contractABI)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testGetContractABI(t *testing.T, db database.Database, address common.Address, expected *abi.ABI) {
	retrieved, err := db.GetContractABI(address)
	parsed, _ := abi.JSON(strings.NewReader(retrieved))
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(parsed.Events) != len(expected.Events) {
		t.Fatalf("expected %v events, but got %v", len(expected.Events), len(parsed.Events))
	}
	if len(parsed.Methods) != len(expected.Methods) {
		t.Fatalf("expected %v methods, but got %v", len(expected.Methods), len(parsed.Methods))
	}
}

func testWriteBlock(t *testing.T, db database.Database, block *types.Block, expectedErr bool) {
	err := db.WriteBlock(block)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testReadBlock(t *testing.T, db database.Database, blockNumber uint64, expected common.Hash) {
	block, err := db.ReadBlock(blockNumber)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if block.Hash != expected {
		t.Fatalf("expected %v, but got %v", expected, block.Hash)
	}
}

func testGetLastPersistedBlockNumeber(t *testing.T, db database.Database, expected uint64) {
	actual, err := db.GetLastPersistedBlockNumber()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %v, but got %v", expected, actual)
	}
}

func testWriteTransaction(t *testing.T, db database.Database, tx *types.Transaction, expectedErr bool) {
	err := db.WriteTransaction(tx)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testReadTransaction(t *testing.T, db database.Database, hash common.Hash, expected *types.Transaction) {
	tx, err := db.ReadTransaction(hash)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if tx.From != expected.From {
		t.Fatalf("expected from %v, but got %v", expected.From, tx.From)
	}
	if tx.To != expected.To {
		t.Fatalf("expected from %v, but got %v", expected.To, tx.To)
	}
	if tx.Value != expected.Value {
		t.Fatalf("expected from %v, but got %v", expected.Value, tx.Value)
	}
}

func testIndexBlock(t *testing.T, db database.Database, address common.Address, block *types.Block) {
	err := db.IndexBlock([]common.Address{address}, block)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func testIndexStorage(t *testing.T, db database.Database, blockNumber uint64, rawStorage map[common.Address]*state.DumpAccount) {
	err := db.IndexStorage(rawStorage, blockNumber)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func testGetLastFiltered(t *testing.T, db database.Database, address common.Address, expected uint64) {
	actual, err := db.GetLastFiltered(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %v, but got %v", expected, actual)
	}
}

func testGetContractCreationTransaction(t *testing.T, db database.Database, address common.Address, expected common.Hash) {
	actual, err := db.GetContractCreationTransaction(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %v, but got %v", expected, actual)
	}
}

func testGetAllTransactionsToAddress(t *testing.T, db database.Database, address common.Address, expected int) {
	txs, err := db.GetAllTransactionsToAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(txs))
	}
}

func testGetAllTransactionsInternalToAddress(t *testing.T, db database.Database, address common.Address, expected int) {
	txs, err := db.GetAllTransactionsInternalToAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(txs))
	}
}

func testGetAllEventsByAddress(t *testing.T, db database.Database, address common.Address, expected int) {
	events, err := db.GetAllEventsFromAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(events) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(events))
	}
}

func testGetStorage(t *testing.T, db database.Database, address common.Address, blockNumber uint64, expected int) {
	storage, err := db.GetStorage(address, blockNumber)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(storage) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(storage))
	}
}
