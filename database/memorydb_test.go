package database

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

func TestMemoryDB(t *testing.T) {
	// setup
	db := NewMemoryDB()
	address := common.HexToAddress("0x0000000000000000000000000000000000000001")
	block := &types.Block{
		Hash:   common.BytesToHash([]byte("dummy")),
		Number: 1,
		Transactions: []common.Hash{
			common.BytesToHash([]byte("tx1")), common.BytesToHash([]byte("tx2")), common.BytesToHash([]byte("tx3")),
		},
	}
	tx1 := &types.Transaction{
		Hash:            common.BytesToHash([]byte("tx1")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              common.Address{0},
		Value:           666,
		CreatedContract: common.HexToAddress("0x0000000000000000000000000000000000000001"),
	}
	tx2 := &types.Transaction{
		Hash:            common.BytesToHash([]byte("tx2")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              common.HexToAddress("0x0000000000000000000000000000000000000009"),
		Value:           666,
		CreatedContract: common.Address{0},
	}
	tx3 := &types.Transaction{
		Hash:            common.BytesToHash([]byte("tx3")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000010"),
		To:              common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Value:           666,
		CreatedContract: common.Address{0},
		Events:          []*types.Event{{}}, // dummy event
	}
	// Add address
	testAddAddresses(t, db, []common.Address{address}, false)
	// Get address
	testGetAddress(t, db, 1)
	// Write transaction
	testWriteTransaction(t, db, tx1, false)
	testWriteTransaction(t, db, tx2, false)
	testWriteTransaction(t, db, tx3, false)
	// Read transaction
	testReadTransaction(t, db, tx1.Hash, tx1)
	// Get last persisted block number
	testGetLastPersistedBlockNumeber(t, db, 0)
	// Write block
	testWriteBlock(t, db, block, false)
	// Read block
	testReadBlock(t, db, 1, block.Hash)
	// Get last persisted block number
	testGetLastPersistedBlockNumeber(t, db, 1)
	// Get last filtered
	testGetLastFiltered(t, db, address, 0)
	// Index block
	testIndexBlock(t, db, address, block, false)
	// Get last filtered
	testGetLastFiltered(t, db, address, 1)
	// Get all transactions by address
	testGetAllTransactionsByAddress(t, db, address, 2)
	// Get all events by address
	testGetAllEventsByAddress(t, db, address, 1)
	// Delete address
	testDeleteAddress(t, db, address, false)
}

func testAddAddresses(t *testing.T, db Database, addresses []common.Address, expectedErr bool) {
	err := db.AddAddresses(addresses)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testDeleteAddress(t *testing.T, db Database, address common.Address, expectedErr bool) {
	err := db.DeleteAddress(address)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testGetAddress(t *testing.T, db Database, expected int) {
	actual := db.GetAddresses()
	if len(actual) != expected {
		t.Fatalf("expected %v addresses, but got %v", expected, len(actual))
	}
}

func testWriteBlock(t *testing.T, db Database, block *types.Block, expectedErr bool) {
	err := db.WriteBlock(block)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testReadBlock(t *testing.T, db Database, blockNumber uint64, expected common.Hash) {
	block, err := db.ReadBlock(blockNumber)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if block.Hash != expected {
		t.Fatalf("expected %v, but got %v", expected, block.Hash)
	}
}

func testGetLastPersistedBlockNumeber(t *testing.T, db Database, expected uint64) {
	actual := db.GetLastPersistedBlockNumber()
	if actual != expected {
		t.Fatalf("expected %v, but got %v", expected, actual)
	}
}

func testWriteTransaction(t *testing.T, db Database, tx *types.Transaction, expectedErr bool) {
	err := db.WriteTransaction(tx)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testReadTransaction(t *testing.T, db Database, hash common.Hash, expected *types.Transaction) {
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

func testIndexBlock(t *testing.T, db Database, address common.Address, block *types.Block, expectedErr bool) {
	err := db.IndexBlock(address, block)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testGetLastFiltered(t *testing.T, db Database, address common.Address, expected uint64) {
	actual := db.GetLastFiltered(address)
	if actual != expected {
		t.Fatalf("expected %v, but got %v", expected, actual)
	}
}

func testGetAllTransactionsByAddress(t *testing.T, db Database, address common.Address, expected int) {
	txs, err := db.GetAllTransactionsByAddress(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(txs))
	}
}

func testGetAllEventsByAddress(t *testing.T, db Database, address common.Address, expected int) {
	events, err := db.GetAllEventsByAddress(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(events) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(events))
	}
}
