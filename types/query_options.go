package types

import "math/big"

var DefaultQueryOptions = &QueryOptions{
	StartBlock:     big.NewInt(0),
	EndBlock:       big.NewInt(-1),
	BeginTimestamp: big.NewInt(0),
	EndTimestamp:   big.NewInt(-1),
	PageSize:       big.NewInt(10),
	PageNumber:     big.NewInt(0),
}

type QueryOptions struct {
	StartBlock *big.Int `json:"startBlock"`
	EndBlock   *big.Int `json:"endBlock"`

	BeginTimestamp *big.Int `json:"beginTimestamp"`
	EndTimestamp   *big.Int `json:"endTimestamp"`

	PageSize   *big.Int `json:"pageSize"`
	PageNumber *big.Int `json:"pageNumber"`
}

func (opts *QueryOptions) SetDefaults() {
	if opts.StartBlock == nil {
		opts.StartBlock = DefaultQueryOptions.StartBlock
	}
	if opts.EndBlock == nil {
		opts.EndBlock = DefaultQueryOptions.EndBlock
	}
	if opts.BeginTimestamp == nil {
		opts.BeginTimestamp = DefaultQueryOptions.BeginTimestamp
	}
	if opts.EndTimestamp == nil {
		opts.EndTimestamp = DefaultQueryOptions.EndTimestamp
	}
	if opts.PageSize == nil {
		opts.PageSize = DefaultQueryOptions.PageSize
	}
	if opts.PageNumber == nil {
		opts.PageNumber = DefaultQueryOptions.PageNumber
	}
}
