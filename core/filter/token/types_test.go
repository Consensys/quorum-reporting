package token

import (
	"errors"
	"math/big"
	"quorumengineering/quorum-report/types"
)

func NewFakeTestTokenDatabase(testErr error, txns []*types.Transaction) *FakeTestTokenDatabase {
	txnMap := make(map[types.Hash]*types.Transaction)
	for _, txn := range txns {
		txnMap[txn.Hash] = txn
	}
	return &FakeTestTokenDatabase{
		testErr: testErr,
		txns:    txnMap,
	}
}

type FakeTestTokenDatabase struct {
	testErr error

	txns map[types.Hash]*types.Transaction

	RecordedContract []types.Address
	RecordedHolder   []types.Address
	RecordedBlock    uint64
	RecordedToken    []*big.Int
}

func (db *FakeTestTokenDatabase) RecordNewBalance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
	if db.testErr != nil {
		return db.testErr
	}
	db.RecordedContract = append(db.RecordedContract, contract)
	db.RecordedHolder = append(db.RecordedHolder, holder)
	db.RecordedBlock = block
	db.RecordedToken = append(db.RecordedToken, amount)
	return nil
}

func (db *FakeTestTokenDatabase) RecordERC721Token(contract types.Address, holder types.Address, block uint64, tokenId *big.Int) error {
	if db.testErr != nil {
		return db.testErr
	}
	db.RecordedContract = append(db.RecordedContract, contract)
	db.RecordedHolder = append(db.RecordedHolder, holder)
	db.RecordedBlock = block
	db.RecordedToken = append(db.RecordedToken, tokenId)
	return nil
}

func (db *FakeTestTokenDatabase) ReadTransaction(hash types.Hash) (*types.Transaction, error) {
	if db.testErr != nil {
		return nil, db.testErr
	}
	if txn, ok := db.txns[hash]; ok {
		return txn, nil
	}
	return nil, errors.New("not found")
}
