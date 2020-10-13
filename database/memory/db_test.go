package memory

import (
	"math/big"
	"testing"

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
	addr           = types.NewAddress("0x0000000000000000000000000000000000000001")
	uselessAddress = types.NewAddress("0x0000000000000000000000000000000000000002")

	tx1 = &types.Transaction{
		Hash:            types.NewHash("0x1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7"),
		BlockNumber:     1,
		From:            types.NewAddress("0x0000000000000000000000000000000000000009"),
		To:              "",
		Value:           666,
		CreatedContract: addr,
	}
	tx2 = &types.Transaction{
		Hash:        types.NewHash("0xbc77a72b3409ba3e098cb45bac1b7727b59dae9a05f37a0dbc61007949c8cede"),
		BlockNumber: 1,
		From:        types.NewAddress("0x0000000000000000000000000000000000000009"),
		To:          uselessAddress,
		Value:       666,
		InternalCalls: []*types.InternalCall{
			{
				To: addr,
			},
		},
	}
	tx3 = &types.Transaction{
		Hash:        types.NewHash("0xb2d58900a820afddd1d926845e7655d445885524b9af1cc946b45949be74cc08"),
		BlockNumber: 1,
		From:        types.NewAddress("0x0000000000000000000000000000000000000010"),
		To:          addr,
		Value:       666,
		Events: []*types.Event{
			{}, // dummy event
			{Address: addr},
		},
	}
	block = &types.Block{
		Hash:   types.NewHash("dummy"),
		Number: 1,
		Transactions: []types.Hash{
			types.NewHash("0x1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7"), types.NewHash("0xbc77a72b3409ba3e098cb45bac1b7727b59dae9a05f37a0dbc61007949c8cede"), types.NewHash("0xb2d58900a820afddd1d926845e7655d445885524b9af1cc946b45949be74cc08"),
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
	rawStorage := map[types.Address]*types.AccountState{
		addr: {
			Storage: map[types.Hash]string{
				types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000000"): "2a",
				types.NewHash("0x0000000000000000000000000000000000000000000000000000000000000001"): "2b",
			},
		},
	}

	testTemplateName := "test template name"
	testTemplateStorage := "test template storage"
	// 1. Add an address and get it.
	testAddAddresses(t, db, []types.Address{addr}, false)
	testGetAddresses(t, db, 1)
	// 2. Add template, assign template, get templates
	testAddTemplate(t, db, testTemplateName, jsondata, testTemplateStorage, false)
	testAssignTemplate(t, db, addr, testTemplateName, false)
	testGetTemplates(t, db, 1)
	testGetStorageLayout(t, db, addr, testTemplateStorage)
	testGetContractABI(t, db, addr, jsondata)
	// 3. Write transaction and get it.
	testWriteTransactions(t, db, tx1, tx2, tx3)
	testReadTransaction(t, db, tx1.Hash, tx1)
	// 4. Write block and get it. Check last persisted block number.
	testGetLastPersistedBlockNumeber(t, db, 0)
	testWriteBlock(t, db, block, false)
	testReadBlock(t, db, 1, block.Hash)
	testGetLastPersistedBlockNumeber(t, db, 1)
	// 5. Index block and check last filtered. Retrieve all transactions/ events.
	testGetLastFiltered(t, db, addr, 0)
	testIndexStorage(t, db, 1, rawStorage)
	testIndexBlock(t, db, addr, block)
	testGetLastFiltered(t, db, addr, 1)
	testGetAllTransactionsToAddress(t, db, addr, types.NewHash("0xb2d58900a820afddd1d926845e7655d445885524b9af1cc946b45949be74cc08"))
	testGetTransactionsToAddressTotal(t, db, addr, 1)
	testGetAllTransactionsInternalToAddress(t, db, addr, types.NewHash("0xbc77a72b3409ba3e098cb45bac1b7727b59dae9a05f37a0dbc61007949c8cede"))
	testGetTransactionsInternalToAddressTotal(t, db, addr, 1)
	testGetAllEventsByAddress(t, db, addr, 1)
	testGetStorage(t, db, addr, 1, 2)
	testGetStorageTotal(t, db, addr, &types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(1)}, 1)
	testGetStorageTotal(t, db, addr, &types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(-1)}, 1)
	testGetStorageWithOptions(t, db, addr, &types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(1)}, 1)
	// 6. Delete address and check last filtered
	testDeleteAddress(t, db, addr, false)
	testGetLastFiltered(t, db, addr, 0)
}

func testAddAddresses(t *testing.T, db database.Database, addresses []types.Address, expectedErr bool) {
	err := db.AddAddresses(addresses)
	if err != nil && !expectedErr {
		t.Fatalf("expected no error, but got %v", err)
	}
	if err == nil && expectedErr {
		t.Fatalf("expected error but got nil")
	}
}

func testDeleteAddress(t *testing.T, db database.Database, address types.Address, expectedErr bool) {
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

func testGetContractABI(t *testing.T, db database.Database, address types.Address, expected string) {
	retrieved, err := db.GetContractABI(address)
	assert.Nil(t, err)
	assert.Equal(t, expected, retrieved)
}

func testGetStorageLayout(t *testing.T, db database.Database, address types.Address, expected string) {
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

func testAssignTemplate(t *testing.T, db database.Database, address types.Address, testTemplateName string, expectedErr bool) {
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

func testReadTransaction(t *testing.T, db database.Database, hash types.Hash, expected *types.Transaction) {
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

func testIndexBlock(t *testing.T, db database.Database, address types.Address, block *types.Block) {
	err := db.IndexBlocks([]types.Address{address}, []*types.Block{block})
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func testIndexStorage(t *testing.T, db database.Database, blockNumber uint64, rawStorage map[types.Address]*types.AccountState) {
	err := db.IndexStorage(rawStorage, blockNumber)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func testGetLastFiltered(t *testing.T, db database.Database, address types.Address, expected uint64) {
	actual, err := db.GetLastFiltered(address)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %v, but got %v", expected, actual)
	}
}

func testGetAllTransactionsToAddress(t *testing.T, db database.Database, address types.Address, expected types.Hash) {
	txs, err := db.GetAllTransactionsToAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != 1 && txs[0] != expected {
		t.Fatalf("expected %v, but got %v", expected.Hex(), txs)
	}
}

func testGetTransactionsToAddressTotal(t *testing.T, db database.Database, address types.Address, expected int) {
	total, err := db.GetTransactionsToAddressTotal(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if total != uint64(expected) {
		t.Fatalf("expected %v, but got %v", expected, total)
	}
}

func testGetAllTransactionsInternalToAddress(t *testing.T, db database.Database, address types.Address, expected types.Hash) {
	txs, err := db.GetAllTransactionsInternalToAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(txs) != 1 && txs[0] != expected {
		t.Fatalf("expected %v, but got %v", expected.Hex(), txs)
	}
}

func testGetTransactionsInternalToAddressTotal(t *testing.T, db database.Database, address types.Address, expected int) {
	total, err := db.GetTransactionsInternalToAddressTotal(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if total != uint64(expected) {
		t.Fatalf("expected %v, but got %v", expected, total)
	}
}

func testGetAllEventsByAddress(t *testing.T, db database.Database, address types.Address, expected int) {
	events, err := db.GetAllEventsFromAddress(address, nil)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if len(events) != expected {
		t.Fatalf("expected %v, but got %v", expected, len(events))
	}
}

func testGetStorage(t *testing.T, db database.Database, address types.Address, blockNumber uint64, expected int) {
	storage, err := db.GetStorage(address, blockNumber)
	assert.Nil(t, err)
	assert.Len(t, storage.Storage, expected)

	//test on a block number we don't have storage for
	storageUnknown, err := db.GetStorage(address, blockNumber+1)
	assert.Nil(t, err)
	assert.Len(t, storageUnknown.Storage, 0)
	assert.EqualValues(t, types.NewHash(""), storageUnknown.StorageRoot)
}

func testGetStorageTotal(t *testing.T, db database.Database, address types.Address, options *types.PageOptions, expected uint64) {
	res, err := db.GetStorageTotal(address, options)
	assert.Nil(t, err)
	assert.Equal(t, res, expected)
}

func testGetStorageWithOptions(t *testing.T, db database.Database, address types.Address, options *types.PageOptions, expected int) {
	res, err := db.GetStorageWithOptions(address, options)
	assert.Nil(t, err)
	assert.Equal(t, len(res), expected)
}

func TestMemoryDB_ContractCreationTransactions(t *testing.T) {
	db := NewMemoryDB()
	_ = db.AddAddresses([]types.Address{
		"1932c48b2bf8102ba33b4a6b545c32236e342f34",
		"ed9d02e382b34818e88b88a309c7fe71e65f419d",
		"8a5e2a6343108babed07899510fb42297938d41f",
	})
	creationTxns := map[types.Hash][]types.Address{
		"1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7": {
			"1932c48b2bf8102ba33b4a6b545c32236e342f34",
			"ed9d02e382b34818e88b88a309c7fe71e65f419d",
		},
		"86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb": {
			"8a5e2a6343108babed07899510fb42297938d41f",
		},
	}

	err := db.SetContractCreationTransaction(creationTxns)
	assert.Nil(t, err)

	testCases := []struct {
		contractAddress types.Address
		txHash          types.Hash
	}{
		{
			"1932c48b2bf8102ba33b4a6b545c32236e342f34",
			"1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7",
		}, {
			"ed9d02e382b34818e88b88a309c7fe71e65f419d",
			"1a6f4292bac138df9a7854a07c93fd14ca7de53265e8fe01b6c986f97d6c1ee7",
		}, {
			"8a5e2a6343108babed07899510fb42297938d41f",
			"86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb",
		},
	}

	for _, testCase := range testCases {
		actualTxHash, err := db.GetContractCreationTransaction(testCase.contractAddress)
		assert.Nil(t, err)
		assert.EqualValues(t, testCase.txHash, actualTxHash)
	}
}

func TestMemoryDB_ContractCreationTransactions_DeletedAddress(t *testing.T) {
	db := NewMemoryDB()
	sampleAddress := types.NewAddress("8a5e2a6343108babed07899510fb42297938d41f")
	creationTxns := map[types.Hash][]types.Address{
		"86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb": {
			sampleAddress,
		},
	}

	err := db.SetContractCreationTransaction(creationTxns)
	assert.Nil(t, err)

	actualTxHash, err := db.GetContractCreationTransaction(sampleAddress)
	assert.EqualError(t, err, "address is not registered")
	assert.EqualValues(t, "", actualTxHash)
}

func TestMemoryDB_GetStorageRanges(t *testing.T) {
	db := NewMemoryDB()
	contract := types.NewAddress("0x8a5e2a6343108babed07899510fb42297938d41f")
	db.AddAddressFrom(contract, 0)

	for i := uint64(1); i < 4500; i += 2 {
		storageMap := map[types.Address]*types.AccountState{
			contract: {Root: "0x73607aa4f228bd19dc95575d08adacede9550df70b9ca9253cb3abf7d8115990"},
		}
		db.IndexStorage(storageMap, i)
	}

	//every odd block num has storage

	testCases := []struct {
		options        types.PageOptions
		expectedResult []types.RangeResult
	}{
		{
			options:        types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(0)},
			expectedResult: []types.RangeResult{{Start: 0, End: 0, ResultCount: 0}},
		},
		{
			options:        types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(800)},
			expectedResult: []types.RangeResult{{Start: 0, End: 800, ResultCount: 400}},
		},
		{
			options:        types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(1500)},
			expectedResult: []types.RangeResult{{Start: 0, End: 1500, ResultCount: 750}},
		},
		{
			options: types.PageOptions{BeginBlockNumber: big.NewInt(0), EndBlockNumber: big.NewInt(4499)},
			expectedResult: []types.RangeResult{
				{Start: 2501, End: 4499, ResultCount: 1000},
				{Start: 501, End: 2500, ResultCount: 1000},
				{Start: 0, End: 500, ResultCount: 250},
			},
		},
		{
			options: types.PageOptions{BeginBlockNumber: big.NewInt(1300), EndBlockNumber: big.NewInt(3500)},
			expectedResult: []types.RangeResult{
				{Start: 1501, End: 3500, ResultCount: 1000},
				{Start: 1300, End: 1500, ResultCount: 100},
			},
		},
	}

	for _, test := range testCases {
		res, err := db.GetStorageRanges(contract, &test.options)
		assert.Nil(t, err)
		assert.Equal(t, test.expectedResult, res)
	}
}

func TestMemorydb_erc20Balance(t *testing.T) {
	db := NewMemoryDB()
	contrAddr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	contrAddr1 := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f55")
	holder0 := types.NewAddress("0xed9d02e382b34818e88b88a309c7fe71e65f419d")
	holder1 := types.NewAddress("0xca843569e3427144cead5e4d5999a3d0ccf92b8e")
	holder2 := types.NewAddress("0xca843569e3427144cead5e4d5999a3d0ccf92bac")
	var balances = []ERC20TokenHolder{
		{
			Contract:    contrAddr,
			Holder:      holder0,
			BlockNumber: 1,
			Amount:      "1000",
		},
		{
			Contract:    contrAddr,
			Holder:      holder0,
			BlockNumber: 2,
			Amount:      "900",
		},
		{
			Contract:    contrAddr,
			Holder:      holder1,
			BlockNumber: 2,
			Amount:      "100",
		},
		{
			Contract:    contrAddr,
			Holder:      holder1,
			BlockNumber: 3,
			Amount:      "150",
		},
		{
			Contract:    contrAddr,
			Holder:      holder0,
			BlockNumber: 3,
			Amount:      "850",
		},
		{
			Contract:    contrAddr1,
			Holder:      holder2,
			BlockNumber: 4,
			Amount:      "850",
		},
	}

	for _, b := range balances {
		amount, _ := new(big.Int).SetString(b.Amount, 10)
		err := db.RecordNewERC20Balance(b.Contract, b.Holder, b.BlockNumber, amount)
		assert.Nil(t, err)
	}
	assert.Equal(t, len(db.erc20BalancesDB), len(balances))
	var result, err = db.GetERC20Balance(contrAddr, holder0, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(1)})
	assert.Nil(t, err)
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[1], big.NewInt(1000))

	result, err = db.GetERC20Balance(contrAddr, holder0, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(2)})
	assert.Nil(t, err)
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[1], big.NewInt(1000))
	assert.Equal(t, result[2], big.NewInt(900))

	result, err = db.GetERC20Balance(contrAddr, holder1, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(2), EndBlockNumber: big.NewInt(2)})
	assert.Nil(t, err)
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[2], big.NewInt(100))

	result, err = db.GetERC20Balance(contrAddr, holder1, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(2), EndBlockNumber: big.NewInt(3)})
	assert.Nil(t, err)
	assert.Equal(t, len(result), 2)
	assert.Equal(t, result[2], big.NewInt(100))
	assert.Equal(t, result[3], big.NewInt(150))

	holdrArr, err := db.GetAllTokenHolders(contrAddr, 1, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(1)})
	assert.Nil(t, err)
	assert.Equal(t, len(holdrArr), 1)
	assert.Equal(t, holdrArr[0], holder0)

	holdrArr, err = db.GetAllTokenHolders(contrAddr, 2, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(2)})
	assert.Nil(t, err)
	assert.Equal(t, len(holdrArr), 2)

	holder0Found := false
	holder1Found := false
	for _, h := range holdrArr {
		if h == holder0 {
			holder0Found = true
		} else if h == holder1 {
			holder1Found = true
		}
	}
	assert.Equal(t, holder0Found, true)
	assert.Equal(t, holder1Found, true)

	// balance before begin block
	result, err = db.GetERC20Balance(contrAddr, holder0, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(5), EndBlockNumber: big.NewInt(5)})
	assert.Nil(t, err)
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[5], big.NewInt(850))
}

func TestMemorydb_erc721Balance(t *testing.T) {
	db := NewMemoryDB()
	contrAddr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	holder0 := types.NewAddress("0xed9d02e382b34818e88b88a309c7fe71e65f419d")
	holder1 := types.NewAddress("0xca843569e3427144cead5e4d5999a3d0ccf92b8e")
	holder2 := types.NewAddress("0xee843569e3427144cead5e4d5999a3d0ccf92bdc")
	var balances = []types.ERC721Token{
		{
			Contract: contrAddr,
			Holder:   holder0,
			Token:    "1",
			HeldFrom: 1,
		},
		{
			Contract: contrAddr,
			Holder:   holder1,
			Token:    "2",
			HeldFrom: 3,
		},
		{
			Contract: contrAddr,
			Holder:   holder2,
			Token:    "3",
			HeldFrom: 5,
		},
	}
	for _, b := range balances {
		tokenId, _ := new(big.Int).SetString(b.Token, 10)
		err := db.RecordERC721Token(b.Contract, b.Holder, b.HeldFrom, tokenId)
		assert.Nil(t, err)
	}

	token, err := db.ERC721TokenByTokenID(contrAddr, 1, big.NewInt(1))
	assert.Nil(t, err)
	assert.Equal(t, token.Holder, holder0)

	token, err = db.ERC721TokenByTokenID(contrAddr, 3, big.NewInt(2))
	assert.Nil(t, err)
	assert.Equal(t, token.Holder, holder1)

	tokenArr, err := db.ERC721TokensForAccountAtBlock(contrAddr, holder0, 2, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(1)})
	assert.Nil(t, err)
	assert.Equal(t, len(tokenArr), 1)
	assert.Equal(t, tokenArr[0].Holder, holder0)
	assert.Equal(t, tokenArr[0].Token, "1")
	assert.Equal(t, tokenArr[0].HeldFrom, uint64(1))

	tokenArr, err = db.AllERC721TokensAtBlock(contrAddr, 3, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(1)})
	assert.Nil(t, err)
	assert.Equal(t, len(tokenArr), 2)
	assert.Equal(t, tokenArr[0].Holder, holder0)
	assert.Equal(t, tokenArr[0].Token, "1")
	assert.Equal(t, tokenArr[0].HeldFrom, uint64(1))

	assert.Equal(t, tokenArr[1].Holder, holder1)
	assert.Equal(t, tokenArr[1].Token, "2")
	assert.Equal(t, tokenArr[1].HeldFrom, uint64(3))

	holdrArr, err := db.AllHoldersAtBlock(contrAddr, 3, &types.TokenQueryOptions{BeginBlockNumber: big.NewInt(1), EndBlockNumber: big.NewInt(1)})
	assert.Nil(t, err)
	assert.Equal(t, len(tokenArr), 2)
	assert.Equal(t, holdrArr[0], holder0)
	assert.Equal(t, holdrArr[1], holder1)

}
