package rpc

import (
	"errors"
	"math/big"
	"net/http"
	"quorumengineering/quorum-report/database"

	"github.com/consensys/quorum-go-utils/types"
)

type TokenRPCAPIs struct {
	db database.TokenDB
}

func NewTokenRPCAPIs(db database.TokenDB) *TokenRPCAPIs {
	return &TokenRPCAPIs{db}
}

func (r *TokenRPCAPIs) GetERC20TokenBalance(req *http.Request, query *ERC20TokenQuery, reply *map[uint64]*big.Int) error {
	if query.Contract == nil {
		return errors.New("no token contract provided")
	}
	if query.Holder == nil {
		return errors.New("no token holder provided")
	}
	if query.Options == nil {
		query.Options = &types.TokenQueryOptions{}
	}
	query.Options.SetDefaults()

	bal, err := r.db.GetERC20Balance(*query.Contract, *query.Holder, query.Options)
	if err != nil {
		return err
	}

	*reply = bal
	return nil
}

func (r *TokenRPCAPIs) GetERC20TokenHoldersAtBlock(req *http.Request, query *ERC20TokenQuery, reply *[]types.Address) error {
	if query.Contract == nil {
		return errors.New("no token contract provided")
	}
	if query.Block == 0 {
		return errors.New("block must be provided and not 0")
	}
	if query.Options == nil {
		query.Options = &types.TokenQueryOptions{}
	}
	query.Options.SetDefaults()

	bal, err := r.db.GetAllTokenHolders(*query.Contract, query.Block, query.Options)
	if err != nil {
		return err
	}

	*reply = bal
	return nil
}

func (r *TokenRPCAPIs) GetHolderForERC721TokenAtBlock(req *http.Request, query *ERC721TokenQuery, reply *types.Address) error {
	if query.Contract == nil {
		return errors.New("no token contract provided")
	}
	if query.TokenId == nil {
		return errors.New("no token ID provided")
	}
	if query.Block == 0 {
		return errors.New("no block given")
	}

	result, err := r.db.ERC721TokenByTokenID(*query.Contract, query.Block, query.TokenId)
	if err != nil {
		return err
	}

	*reply = result.Holder
	return nil
}

func (r *TokenRPCAPIs) ERC721TokensForAccountAtBlock(req *http.Request, query *ERC721TokenQuery, reply *[]types.ERC721Token) error {
	if query.Contract == nil {
		return errors.New("no token contract provided")
	}
	if query.Holder == nil {
		return errors.New("no token holder provided")
	}
	if query.Block == 0 {
		return errors.New("no block given")
	}
	if query.Options == nil {
		query.Options = &types.TokenQueryOptions{}
	}
	query.Options.SetDefaults()

	results, err := r.db.ERC721TokensForAccountAtBlock(*query.Contract, *query.Holder, query.Block, query.Options)
	if err != nil {
		return err
	}

	*reply = results
	return nil
}

func (r *TokenRPCAPIs) AllERC721TokensAtBlock(req *http.Request, query *ERC721TokenQuery, reply *[]types.ERC721Token) error {
	if query.Contract == nil {
		return errors.New("no token contract provided")
	}
	if query.Block == 0 {
		return errors.New("no block given")
	}
	if query.Options == nil {
		query.Options = &types.TokenQueryOptions{}
	}
	query.Options.SetDefaults()

	results, err := r.db.AllERC721TokensAtBlock(*query.Contract, query.Block, query.Options)
	if err != nil {
		return err
	}

	*reply = results
	return nil
}

func (r *TokenRPCAPIs) AllERC721HoldersAtBlock(req *http.Request, query *ERC721TokenQuery, reply *[]types.Address) error {
	if query.Contract == nil {
		return errors.New("no token contract provided")
	}
	if query.Block == 0 {
		return errors.New("no block given")
	}
	if query.Options == nil {
		query.Options = &types.TokenQueryOptions{}
	}
	query.Options.SetDefaults()

	results, err := r.db.AllHoldersAtBlock(*query.Contract, query.Block, query.Options)
	if err != nil {
		return err
	}

	*reply = results
	return nil
}
