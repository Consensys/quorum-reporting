package memory

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

// MemoryDB is a sample memory database for dev only.
type MemoryDB struct {
	// registered contract data
	addressDB       []types.Address
	templateDB      map[types.Address]string
	abiDB           map[string]string
	storageLayoutDB map[string]string
	// blockchain data
	blockDB                  map[uint64]*types.Block
	txDB                     map[types.Hash]*types.Transaction
	lastPersistedBlockNumber uint64
	// index data
	txIndexDB        map[types.Address]*TxIndexer
	eventIndexDB     map[types.Address][]*types.Event
	storageIndexDB   map[types.Address]*StorageIndexer
	lastFiltered     map[types.Address]uint64
	erc20BalancesDB  []ERC20TokenHolder
	erc721BalancesDB []SortableERC721Token
	// mutex lock
	mux sync.RWMutex
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		addressDB:                []types.Address{},
		templateDB:               make(map[types.Address]string),
		abiDB:                    make(map[string]string),
		storageLayoutDB:          make(map[string]string),
		blockDB:                  make(map[uint64]*types.Block),
		txDB:                     make(map[types.Hash]*types.Transaction),
		txIndexDB:                make(map[types.Address]*TxIndexer),
		eventIndexDB:             make(map[types.Address][]*types.Event),
		storageIndexDB:           make(map[types.Address]*StorageIndexer),
		lastPersistedBlockNumber: 0,
		lastFiltered:             make(map[types.Address]uint64),
	}
}

type TxIndexer struct {
	contractCreationTx types.Hash
	txsTo              []types.Hash
	txsInternalTo      []types.Hash
}

type ERC20TokenHolder struct {
	Contract    types.Address
	Holder      types.Address
	BlockNumber uint64
	Amount      string
	HeldUntil   *uint64
}

type SortableERC721Token struct {
	types.ERC721Token

	//Allows the token to be sortable by splitting it into component parts
	First  uint64
	Second uint64
	Third  uint64
	Fourth uint64
	Fifth  uint64
}

func NewTxIndexer() *TxIndexer {
	return &TxIndexer{
		contractCreationTx: "",
		txsTo:              []types.Hash{},
		txsInternalTo:      []types.Hash{},
	}
}

type StorageIndexer struct {
	root    map[uint64]string
	storage map[string]map[types.Hash]string
}

func NewStorageIndexer() *StorageIndexer {
	return &StorageIndexer{
		root:    make(map[uint64]string),
		storage: make(map[string]map[types.Hash]string),
	}
}

func (db *MemoryDB) AddAddresses(addresses []types.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	if len(addresses) > 0 {
		newAddresses := []types.Address{}
		for _, a := range addresses {
			isExist := false
			for _, exist := range db.addressDB {
				if a == exist {
					isExist = true
					break
				}
			}
			if !isExist {
				db.txIndexDB[a] = NewTxIndexer()
				db.eventIndexDB[a] = []*types.Event{}
				db.storageIndexDB[a] = NewStorageIndexer()
				newAddresses = append(newAddresses, a)
			}
		}
		db.addressDB = append(db.addressDB, newAddresses...)
	}
	return nil
}

func (db *MemoryDB) AddAddressFrom(address types.Address, from uint64) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	isExist := false
	for _, exist := range db.addressDB {
		if address == exist {
			isExist = true
			break
		}
	}
	if !isExist {
		db.txIndexDB[address] = NewTxIndexer()
		db.eventIndexDB[address] = []*types.Event{}
		db.storageIndexDB[address] = NewStorageIndexer()
		db.addressDB = append(db.addressDB, address)
		db.lastFiltered[address] = from - 1
	}
	return nil
}

func (db *MemoryDB) DeleteAddress(address types.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	index := -1
	for i, a := range db.addressDB {
		if address == a {
			index = i
			break
		}
	}
	if index != -1 {
		err := db.removeAllIndices(address)
		if err != nil {
			return err
		}
		db.addressDB = append(db.addressDB[:index], db.addressDB[index+1:]...)
		return nil
	}
	return errors.New("address does not exist")
}

func (db *MemoryDB) GetAddresses() ([]types.Address, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.addressDB, nil
}

func (db *MemoryDB) GetContractTemplate(address types.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.templateDB[address], nil
}

func (db *MemoryDB) GetContractABI(address types.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.abiDB[db.templateDB[address]], nil
}

func (db *MemoryDB) GetStorageLayout(address types.Address) (string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.storageLayoutDB[db.templateDB[address]], nil
}

func (db *MemoryDB) AddTemplate(name string, abi string, layout string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.abiDB[name] = abi
	db.storageLayoutDB[name] = layout
	return nil
}

func (db *MemoryDB) AssignTemplate(address types.Address, name string) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	db.templateDB[address] = name
	return nil
}

func (db *MemoryDB) GetTemplates() ([]string, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	// merge abiDB and storageLayoutDB to find the full template name list
	templateNames := make(map[string]bool)
	for template := range db.abiDB {
		templateNames[template] = true
	}
	for template := range db.storageLayoutDB {
		templateNames[template] = true
	}
	res := make([]string, 0)
	for template := range templateNames {
		res = append(res, template)
	}
	return res, nil
}

func (db *MemoryDB) GetTemplateDetails(templateName string) (*types.Template, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	if (db.abiDB[templateName] == "") && (db.storageLayoutDB[templateName] == "") {
		return nil, database.ErrNotFound
	}

	return &types.Template{
		TemplateName:  templateName,
		ABI:           db.abiDB[templateName],
		StorageLayout: db.storageLayoutDB[templateName],
	}, nil
}

func (db *MemoryDB) WriteBlocks(blocks []*types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	for _, block := range blocks {
		if block == nil {
			return errors.New("block is nil")
		}
		blockNumber := block.Number
		db.blockDB[blockNumber] = block
		// Update last persisted block number.
		if blockNumber == db.lastPersistedBlockNumber+1 {
			for {
				if _, ok := db.blockDB[blockNumber+1]; ok {
					blockNumber++
				} else {
					break
				}
			}
			db.lastPersistedBlockNumber = blockNumber
		}
		log.Debug("Block stored", "number", block.Number, "hash", block.Hash.String())
		log.Debug("Last persisted block", "number", db.lastPersistedBlockNumber)
	}
	return nil
}

func (db *MemoryDB) ReadBlock(blockNumber uint64) (*types.Block, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if block, ok := db.blockDB[blockNumber]; ok {
		return block, nil
	}
	return nil, errors.New("block does not exist")
}

func (db *MemoryDB) GetLastPersistedBlockNumber() (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastPersistedBlockNumber, nil
}

func (db *MemoryDB) WriteTransactions(transactions []*types.Transaction) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	for _, tx := range transactions {
		if tx == nil {
			return errors.New("transaction is nil")
		}
		db.txDB[tx.Hash] = tx
		log.Debug("Transaction stored", "hash", tx.Hash.Hex())
	}
	return nil
}

func (db *MemoryDB) ReadTransaction(hash types.Hash) (*types.Transaction, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if tx, ok := db.txDB[hash]; ok {
		return tx, nil
	}
	return nil, errors.New("transaction does not exist")
}

func (db *MemoryDB) IndexStorage(rawStorage map[types.Address]*types.AccountState, blockNumber uint64) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	for address, dumpAccount := range rawStorage {
		db.storageIndexDB[address].root[blockNumber] = dumpAccount.Root.String()
		if _, ok := db.storageIndexDB[address].storage[dumpAccount.Root.String()]; !ok {
			db.storageIndexDB[address].storage[dumpAccount.Root.String()] = dumpAccount.Storage
		}
	}
	return nil
}

func (db *MemoryDB) IndexBlocks(addresses []types.Address, blocks []*types.Block) error {
	for _, block := range blocks {
		db.indexBlock(addresses, block)
	}
	return nil
}

func (db *MemoryDB) SetContractCreationTransaction(creationTxns map[types.Hash][]types.Address) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	for txHash, addresses := range creationTxns {
		for _, createdAddress := range addresses {
			if _, ok := db.txIndexDB[createdAddress]; !ok {
				//tried to index a deleted address, do nothing
				log.Debug("Ignored deleted address contract creation", "tx", txHash.Hex(), "contract", createdAddress)
				return nil
			}
			db.txIndexDB[createdAddress].contractCreationTx = txHash
			log.Debug("Indexed address of contract creation", "tx", txHash.Hex(), "contract", createdAddress)
		}
	}
	return nil
}

func (db *MemoryDB) GetContractCreationTransaction(address types.Address) (types.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return "", errors.New("address is not registered")
	}
	return db.txIndexDB[address].contractCreationTx, nil
}

func (db *MemoryDB) GetAllTransactionsToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsTo, nil
}

func (db *MemoryDB) GetTransactionsToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.txIndexDB[address].txsTo)), nil
}

func (db *MemoryDB) GetAllTransactionsInternalToAddress(address types.Address, options *types.QueryOptions) ([]types.Hash, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.txIndexDB[address].txsInternalTo, nil
}

func (db *MemoryDB) GetTransactionsInternalToAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.txIndexDB[address].txsInternalTo)), nil
}

func (db *MemoryDB) GetAllEventsFromAddress(address types.Address, options *types.QueryOptions) ([]*types.Event, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	return db.eventIndexDB[address], nil
}

func (db *MemoryDB) GetEventsFromAddressTotal(address types.Address, options *types.QueryOptions) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	return uint64(len(db.eventIndexDB[address])), nil
}

func (db *MemoryDB) GetStorageWithOptions(address types.Address, options *types.PageOptions) ([]*types.StorageResult, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	var convertedList []*types.StorageResult

	fromBlockNum := options.BeginBlockNumber.Uint64()
	endBlockNum := options.EndBlockNumber.Int64()
	blockNum := fromBlockNum

	storageIndexer, ok := db.storageIndexDB[address]
	if ok {
		for blkNum, storageRoot := range storageIndexer.root {
			if blkNum >= fromBlockNum && (blkNum <= uint64(endBlockNum) || endBlockNum == -1) {
				convertedList = append(convertedList, &types.StorageResult{
					Storage:     storageIndexer.storage[storageRoot],
					StorageRoot: types.NewHash(storageRoot),
					BlockNumber: blockNum,
				})
			}
		}
	}

	return convertedList, nil
}

func (db *MemoryDB) GetStorageTotal(address types.Address, options *types.PageOptions) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return 0, errors.New("address is not registered")
	}
	fromBlockNum := options.BeginBlockNumber.Uint64()
	endBlockNum := options.EndBlockNumber.Uint64()
	var total uint64
	blockNum := fromBlockNum
	for blockNum <= endBlockNum {
		storageRoot, ok := db.storageIndexDB[address].root[blockNum]
		if ok {
			total += uint64(len(db.storageIndexDB[address].storage[storageRoot]))
		}
		blockNum++
	}

	return total, nil
}

func (db *MemoryDB) GetStorageRanges(contract types.Address, options *types.PageOptions) ([]types.RangeResult, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(contract) {
		return nil, errors.New("address is not registered")
	}

	end := options.EndBlockNumber
	if big.NewInt(-1).Cmp(end) == 0 {
		endUint64, _ := db.GetLastFiltered(contract)
		end = new(big.Int).SetUint64(endUint64)
	}

	startUint64 := options.BeginBlockNumber.Uint64()
	endUint64 := end.Uint64()

	storage, ok := db.storageIndexDB[contract]
	if !ok {
		return nil, errors.New("contract is not storage indexed")
	}

	var results []types.RangeResult

	currentCount := 0
	lastEnd := endUint64
	for endUint64 >= startUint64 {
		_, ok := storage.root[endUint64]
		if ok {
			currentCount++
		}

		if currentCount == 1000 {
			rangeRes := types.RangeResult{
				Start:       endUint64,
				End:         lastEnd,
				ResultCount: 1000,
			}
			results = append(results, rangeRes)
			currentCount = 0
			lastEnd = endUint64 - 1
		}
		if endUint64 == startUint64 {
			break
		}
		endUint64--
	}
	rangeRes := types.RangeResult{
		Start:       endUint64,
		End:         lastEnd,
		ResultCount: currentCount,
	}
	if options.BeginBlockNumber.Uint64() == 0 {
		rangeRes.Start = 0
	}
	results = append(results, rangeRes)

	return results, nil
}

func (db *MemoryDB) GetStorage(address types.Address, blockNumber uint64) (*types.StorageResult, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if !db.addressIsRegistered(address) {
		return nil, errors.New("address is not registered")
	}
	storageRoot, ok := db.storageIndexDB[address].root[blockNumber]
	if !ok {
		return &types.StorageResult{
			Storage:     make(map[types.Hash]string),
			StorageRoot: types.NewHash(""),
			BlockNumber: blockNumber,
		}, nil
	}
	return &types.StorageResult{
		Storage:     db.storageIndexDB[address].storage[storageRoot],
		StorageRoot: types.NewHash(storageRoot),
		BlockNumber: blockNumber,
	}, nil
}

func (db *MemoryDB) GetLastFiltered(address types.Address) (uint64, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	return db.lastFiltered[address], nil
}

func (db *MemoryDB) Stop() {}

// internal functions

func (db *MemoryDB) addressIsRegistered(address types.Address) bool {
	for _, a := range db.addressDB {
		if address == a {
			return true
		}
	}
	return false
}

func (db *MemoryDB) indexBlock(addresses []types.Address, block *types.Block) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	// filter out registered and unfiltered address only
	filteredAddresses := map[types.Address]bool{}
	for _, address := range addresses {
		if db.addressIsRegistered(address) && db.lastFiltered[address] < block.Number {
			filteredAddresses[address] = true
			log.Info("Index registered address ", "address", address.Hex(), "blocknumber", block.Number)
		}
	}

	// index transactions and events
	for _, txHash := range block.Transactions {
		db.indexTransaction(filteredAddresses, db.txDB[txHash])
	}

	for address := range filteredAddresses {
		db.lastFiltered[address] = block.Number
	}
	return nil
}

func (db *MemoryDB) indexTransaction(filteredAddresses map[types.Address]bool, tx *types.Transaction) {
	if filteredAddresses[tx.To] {
		db.txIndexDB[tx.To].txsTo = append(db.txIndexDB[tx.To].txsTo, tx.Hash)
		log.Debug("Indexed tx recipient", "tx", tx.Hash.Hex(), "recipient", tx.To.Hex())
	}

	for _, internalCall := range tx.InternalCalls {
		if filteredAddresses[internalCall.To] {
			db.txIndexDB[internalCall.To].txsInternalTo = append(db.txIndexDB[internalCall.To].txsInternalTo, tx.Hash)
			log.Debug("Indexed transactions internal calls", "tx", tx.Hash.Hex(), "internal-recipient", internalCall.To.Hex())
		}
	}
	// Index events emitted by the given address
	for _, event := range tx.Events {
		addr := event.Address
		if filteredAddresses[addr] {
			db.eventIndexDB[addr] = append(db.eventIndexDB[addr], event)
			log.Debug("Indexed emitted event", "tx", event.TransactionHash.Hex(), "address", event.Address.Hex())
		}
	}
}

func (db *MemoryDB) removeAllIndices(address types.Address) error {
	delete(db.txIndexDB, address)
	delete(db.eventIndexDB, address)
	delete(db.storageIndexDB, address)
	db.lastFiltered[address] = 0
	return nil
}

func (db *MemoryDB) GetERC20EntryAtBlock(contract types.Address, holder types.Address, block uint64) (*ERC20TokenHolder, error) {
	var tmpItem int
	found := false
	for i, item := range db.erc20BalancesDB {
		if item.BlockNumber <= block && item.Contract == contract && item.Holder == holder {
			if !found {
				tmpItem = i
				found = true
			} else {
				if item.BlockNumber > db.erc20BalancesDB[tmpItem].BlockNumber {
					tmpItem = i
				}
			}
		}
	}
	if !found {
		return nil, database.ErrNotFound
	}
	return &db.erc20BalancesDB[tmpItem], nil
}

func (db *MemoryDB) RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
	existingTokenEntry, errExisting := db.GetERC20EntryAtBlock(contract, holder, block-1)
	if errExisting != nil && errExisting != database.ErrNotFound {
		return errExisting
	}

	//add new entry
	tokenInfo := ERC20TokenHolder{
		Contract:    contract,
		Holder:      holder,
		BlockNumber: block,
		Amount:      amount.String(),
	}
	db.erc20BalancesDB = append(db.erc20BalancesDB, tokenInfo)
	/////
	if errExisting == database.ErrNotFound {
		return nil
	}
	blk := block - 1
	existingTokenEntry.HeldUntil = &blk

	return nil
}

func (db *MemoryDB) GetERC20Balance(contract types.Address, holder types.Address, options *types.TokenQueryOptions) (map[uint64]*big.Int, error) {
	balanceMap := make(map[uint64]*big.Int)
	frmBlkNum := options.BeginBlockNumber.Uint64()
	endBlkNum := options.EndBlockNumber.Int64()
	for _, b := range db.erc20BalancesDB {
		if contract == b.Contract && holder == b.Holder && b.BlockNumber >= frmBlkNum && (b.BlockNumber <= uint64(endBlkNum) || endBlkNum == -1) {
			tokAmt, success := new(big.Int).SetString(b.Amount, 10)
			if !success {
				return nil, errors.New("could not parse token value")
			}
			balanceMap[b.BlockNumber] = tokAmt
		}
	}
	return balanceMap, nil
}

func (db *MemoryDB) GetAllTokenHolders(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	var holderMap = make(map[types.Address]bool)
	for _, k := range db.erc20BalancesDB {
		if k.Contract == contract && k.BlockNumber <= block && k.Holder != "0000000000000000000000000000000000000000" {
			holderMap[k.Holder] = true
		}
	}
	var holderArr []types.Address
	for holdr := range holderMap {
		holderArr = append(holderArr, holdr)
	}
	return holderArr, nil
}

func (db *MemoryDB) RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error {
	//find old entry
	existingTokenEntry, errExisting := db.ERC721TokenByTokenID(contract, block-1, tokenId)
	if errExisting != nil && errExisting != database.ErrNotFound {
		return errExisting
	}

	paddedTokenId := fmt.Sprintf("%085d", tokenId)
	first, _ := strconv.ParseUint(paddedTokenId[0:17], 10, 64)
	second, _ := strconv.ParseUint(paddedTokenId[17:34], 10, 64)
	third, _ := strconv.ParseUint(paddedTokenId[34:51], 10, 64)
	fourth, _ := strconv.ParseUint(paddedTokenId[51:68], 10, 64)
	fifth, _ := strconv.ParseUint(paddedTokenId[68:85], 10, 64)

	//add new entry
	tokenHolderInfo := SortableERC721Token{
		types.ERC721Token{
			Contract:  contract,
			Holder:    holder,
			Token:     tokenId.String(),
			HeldFrom:  block,
			HeldUntil: nil,
		},
		first, second, third, fourth, fifth,
	}
	db.erc721BalancesDB = append(db.erc721BalancesDB, tokenHolderInfo)
	/////
	if errExisting == database.ErrNotFound {
		return nil
	}

	blk := block - 1
	existingTokenEntry.HeldUntil = &blk
	return nil
}

func (db *MemoryDB) ERC721TokenByTokenID(contract types.Address, block uint64, tokenId *big.Int) (*types.ERC721Token, error) {
	var tmpItem int
	found := false
	for i, item := range db.erc721BalancesDB {
		if item.Contract == contract && item.HeldFrom <= block && item.Token == tokenId.String() {
			if !found {
				tmpItem = i
				found = true
			} else {
				if item.HeldFrom > db.erc721BalancesDB[tmpItem].HeldFrom {
					tmpItem = i
				}
			}

		}
	}
	if !found {
		return nil, database.ErrNotFound
	}
	return &db.erc721BalancesDB[tmpItem].ERC721Token, nil
}

func (db *MemoryDB) ERC721TokensForAccountAtBlock(contract types.Address, holder types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	return db.erc721TokensAtBlock(contract, &holder, block, options)
}

func (db *MemoryDB) erc721TokensAtBlock(contract types.Address, holder *types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	startTokenId := big.NewInt(-1)
	if options.After != "" {
		parsed, success := new(big.Int).SetString(options.After, 10)
		if !success {
			return nil, errors.New(`could not parse "after" token ID`)
		}
		startTokenId = parsed
	}
	next := new(big.Int).Add(startTokenId, big.NewInt(1))

	paddedStartTokenId := fmt.Sprintf("%085d", next)
	startFirst, _ := strconv.ParseUint(paddedStartTokenId[0:17], 10, 64)
	startSecond, _ := strconv.ParseUint(paddedStartTokenId[17:34], 10, 64)
	startThird, _ := strconv.ParseUint(paddedStartTokenId[34:51], 10, 64)
	startFourth, _ := strconv.ParseUint(paddedStartTokenId[51:68], 10, 64)
	startFifth, _ := strconv.ParseUint(paddedStartTokenId[68:85], 10, 64)
	var tmpItem types.ERC721Token
	found := false
	var result []types.ERC721Token
	for _, k := range db.erc721BalancesDB {
		if k.Contract == contract && (holder == nil || *holder == k.Holder) && k.HeldFrom <= block {
			var first, second, third, fourth, fifth bool
			if k.First >= startFirst {
				first = true
			}
			if k.Second >= startSecond {
				second = true
			}
			if k.Third >= startThird {
				third = true
			}
			if k.Fourth >= startFourth {
				fourth = true
			}
			if k.Fifth >= startFifth {
				fifth = true
			}
			matched := first || second || third || fourth || fifth
			if matched {
				if !found {
					tmpItem = k.ERC721Token
					result = append(result, tmpItem)
					found = true
				} else {
					if k.HeldFrom > tmpItem.HeldFrom {
						tmpItem = k.ERC721Token
						result = append(result, tmpItem)
					}
				}
			}

		}

	}

	return result, nil
}

func (db *MemoryDB) AllERC721TokensAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.ERC721Token, error) {
	return db.erc721TokensAtBlock(contract, nil, block, options)
}

func (db *MemoryDB) AllHoldersAtBlock(contract types.Address, block uint64, options *types.TokenQueryOptions) ([]types.Address, error) {
	res, err := db.erc721TokensAtBlock(contract, nil, block, options)
	if err != nil {
		return nil, err
	}
	var holders []types.Address
	for _, k := range res {
		holders = append(holders, k.Holder)
	}
	return holders, nil
}
