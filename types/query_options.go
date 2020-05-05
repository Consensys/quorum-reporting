package types

import "math/big"

var defaultQueryOptions = &QueryOptions{
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
		opts.StartBlock = defaultQueryOptions.StartBlock
	}
	if opts.EndBlock == nil {
		opts.EndBlock = defaultQueryOptions.EndBlock
	}
	if opts.BeginTimestamp == nil {
		opts.BeginTimestamp = defaultQueryOptions.BeginTimestamp
	}
	if opts.EndTimestamp == nil {
		opts.EndTimestamp = defaultQueryOptions.EndTimestamp
	}
	if opts.PageSize == nil {
		opts.PageSize = defaultQueryOptions.PageSize
	}
	if opts.PageNumber == nil {
		opts.PageNumber = defaultQueryOptions.PageNumber
	}
}
