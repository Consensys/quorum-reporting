package types

import (
	"math/big"
)

var defaultQueryOptions = &QueryOptions{
	BeginBlockNumber: big.NewInt(0),
	EndBlockNumber:   big.NewInt(-1),
	BeginTimestamp:   big.NewInt(0),
	EndTimestamp:     big.NewInt(-1),
	PageSize:         10,
	ExtraOpts:        []string{},
}

type QueryOptions struct {
	BeginBlockNumber *big.Int `json:"beginBlockNumber"`
	EndBlockNumber   *big.Int `json:"endBlockNumber"`

	BeginTimestamp *big.Int `json:"beginTimestamp"`
	EndTimestamp   *big.Int `json:"endTimestamp"`

	PageSize  int      `json:"pageSize"`
	ExtraOpts []string `json:"lastResult"`
}

func (opts *QueryOptions) SetDefaults() {
	if opts.BeginBlockNumber == nil {
		opts.BeginBlockNumber = defaultQueryOptions.BeginBlockNumber
	}
	if opts.EndBlockNumber == nil {
		opts.EndBlockNumber = defaultQueryOptions.EndBlockNumber
	}
	if opts.BeginTimestamp == nil {
		opts.BeginTimestamp = defaultQueryOptions.BeginTimestamp
	}
	if opts.EndTimestamp == nil {
		opts.EndTimestamp = defaultQueryOptions.EndTimestamp
	}
	if opts.PageSize == 0 {
		opts.PageSize = defaultQueryOptions.PageSize
	}
	if opts.ExtraOpts == nil {
		opts.ExtraOpts = defaultQueryOptions.ExtraOpts
	}
}
