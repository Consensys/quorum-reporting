package memory

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

const jsondata = `
[
	{ "type" : "function", "name" : "balance", "constant" : true },
	{ "type" : "function", "name" : "send", "constant" : false, "inputs" : [ { "name" : "amount", "type" : "uint256" } ] }
]`

var (
	address        = common.HexToAddress("0x0000000000000000000000000000000000000001")
	addr           = types.NewAddress("0x0000000000000000000000000000000000000001")
	uselessAddress = common.HexToAddress("0x0000000000000000000000000000000000000002")

	tx1 = &types.Transaction{
		Hash:            common.BytesToHash([]byte("tx1")),
		BlockNumber:     1,
		From:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
		To:              common.Address{0},
		Value:           666,
		CreatedContract: address,
	}
	tx2 = &types.Transaction{
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
	tx3 = &types.Transaction{
		Hash:        common.BytesToHash([]byte("tx3")),
		BlockNumber: 1,
		From:        common.HexToAddress("0x0000000000000000000000000000000000000010"),
		To:          address,
		Value:       666,
		Events: []*types.Event{
			{}, // dummy event
			{Address: addr},
		},
	}
	block = &types.Block{
		Hash:   types.NewHash("dummy"),
		Number: 1,
		Transactions: []common.Hash{
			common.BytesToHash([]byte("tx1")), common.BytesToHash([]byte("tx2")), common.BytesToHash([]byte("tx3")),
		},
	}
)

func TestMemoryDB_WriteTransactions(t *testing.T) {
	db := NewMemoryDB()

	err := db.WriteTransactions([]*types.Transaction{tx1, tx2, tx3})

	assert.Nil(t, err, "unexpected err")

	retrievedTx1, err := db.ReadTransaction(tx1.Hash)
	assert.Nil(t, err, "unexpected err")
	assert.Equal(t, tx1, retrievedTx1, "unexpected tx from db: %s", retrievedTx1)

	retrievedTx2, err := db.ReadTransaction(tx2.Hash)
	assert.Nil(t, err, "unexpected err")
	assert.Equal(t, tx2, retrievedTx2, "unexpected tx from db: %s", retrievedTx2)

	retrievedTx3, err := db.ReadTransaction(tx3.Hash)
	assert.Nil(t, err, "unexpected err")
	assert.Equal(t, tx3, retrievedTx3, "unexpected tx from db: %s", retrievedTx3)
}

func TestMemoryDB_WriteBlocks(t *testing.T) {
	db := NewMemoryDB()

	err := db.WriteBlocks([]*types.Block{block})

	assert.Nil(t, err, "unexpected err")

	retrievedblock, err := db.ReadBlock(block.Number)
	assert.Nil(t, err, "unexpected err")
	assert.Equal(t, block, retrievedblock, "unexpected block from db: %s", retrievedblock)
}

func TestMemoryDB(t *testing.T) {
	// test data
	db := NewMemoryDB()
	rawStorage := map[common.Address]*types.AccountState{
		address: {
			Storage: map[types.Hash]string{
				types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000000"): "2a",
				types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000001"): "2b",
			},
		},
	}
	testTemplateName := "test template name"
	testTemplateStorage := "test template storage"
	// 1. Add an address and get it.
	testAddAddresses(t, db, []common.Address{address}, false)
	testGetAddresses(t, db, 1)
	// 2. Add template, assign template, get templates
	testAddTemplate(t, db, testTemplateName, jsondata, testTemplateStorage, false)
	testAssignTemplate(t, db, address, testTemplateName, false)
	testGetTemplates(t, db, 1)
	testGetStorageLayout(t, db, address, testTemplateStorage)
	testGetContractABI(t, db, address, jsondata)
	// 3. Write transaction and get it.
	testWriteTransactions(t, db, tx1, tx2, tx3)
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
	testGetAllTransactionsToAddress(t, db, address, common.BytesToHash([]byte("tx3")))
	testGetTransactionsToAddressTotal(t, db, address, 1)
	testGetAllTransactionsInternalToAddress(t, db, address, common.BytesToHash([]byte("tx2")))
	testGetTransactionsInternalToAddressTotal(t, db, address, 1)
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

func testGetContractABI(t *testing.T, db database.Database, address common.Address, expected string) {
	retrieved, err := db.GetContractABI(address)
	assert.Nil(t, err)
	assert.Equal(t, expected, retrieved)
}

func testGetStorageLayout(t *testing.T, db database.Database, address common.Address, expected string) {
	retrieved, err := db.GetStorageLayout(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if retrieved != expected {
		t.Fatalf("expected %v events, but got %v", expected, retrieved)
	}
}

func testAddTemplate(t *testing.T, db database.Database, testTemplateName, testABI, testStorageLayout string, expectedErr bool) {
	err := db.AddTemplate(testTemplateName, testABI, testStorageLayout)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testAssignTemplate(t *testing.T, db database.Database, address common.Address, testTemplateName string, expectedErr bool) {
	err := db.AssignTemplate(address, testTemplateName)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testGetTemplates(t *testing.T, db database.Database, expected int) {
	templates, err := db.GetTemplates()
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(templates) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(templates))
	}
}

func testWriteBlock(t *testing.T, db database.Database, block *types.Block, expectedErr bool) {
	err := db.WriteBlocks([]*types.Block{block})
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testReadBlock(t *testing.T, db database.Database, blockNumber uint64, expected types.Hash) {
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

func testWriteTransactions(t *testing.T, db database.Database, txs ...*types.Transaction) {
	err := db.WriteTransactions(txs)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
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
	err := db.IndexBlocks([]common.Address{address}, []*types.Block{block})
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func testIndexStorage(t *testing.T, db database.Database, blockNumber uint64, rawStorage map[common.Address]*types.AccountState) {
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

func testGetAllTransactionsToAddress(t *testing.T, db database.Database, address common.Address, expected common.Hash) {
	txs, err := db.GetAllTransactionsToAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != 1 && txs[0] != expected {
		t.Fatalf("expected %v, but got %v", expected.Hex(), txs)
	}
}

func testGetTransactionsToAddressTotal(t *testing.T, db database.Database, address common.Address, expected int) {
	total, err := db.GetTransactionsToAddressTotal(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if total != uint64(expected) {
		t.Fatalf("expected %v, but got %v", expected, total)
	}
}

func testGetAllTransactionsInternalToAddress(t *testing.T, db database.Database, address common.Address, expected common.Hash) {
	txs, err := db.GetAllTransactionsInternalToAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != 1 && txs[0] != expected {
		t.Fatalf("expected %v, but got %v", expected.Hex(), txs)
	}
}

func testGetTransactionsInternalToAddressTotal(t *testing.T, db database.Database, address common.Address, expected int) {
	total, err := db.GetTransactionsInternalToAddressTotal(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if total != uint64(expected) {
		t.Fatalf("expected %v, but got %v", expected, total)
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
