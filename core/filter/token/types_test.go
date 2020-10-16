package token

import (
	"math/big"
	"quorumengineering/quorum-report/types"
)

func NewFakeTestTokenDatabase(testErr error) *FakeTestTokenDatabase {
	return &FakeTestTokenDatabase{
		testErr: testErr,
	}
}

type FakeTestTokenDatabase struct {
	testErr error

	RecordedContract []types.Address
	RecordedHolder   []types.Address
	RecordedBlock    uint64
	RecordedToken    []*big.Int
}

func (db *FakeTestTokenDatabase) RecordNewERC20Balance(contract types.Address, holder types.Address, block uint64, amount *big.Int) error {
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
