package client

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/p2p"

	"quorumengineering/quorum-report/log"
)

func DumpAddress(c Client, address common.Address, blockNumber uint64) (*state.DumpAccount, error) {
	log.Debug("fetching account dump", "account", address.String(), "blocknumber", blockNumber)
	dumpAccount := &state.DumpAccount{}
	err := c.RPCCall(context.Background(), &dumpAccount, "debug_dumpAddress", address, hexutil.EncodeUint64(blockNumber))
	if err != nil {
		return nil, err
	}
	return dumpAccount, nil
}

func TraceTransaction(c Client, txHash common.Hash) (map[string]interface{}, error) {
	log.Debug("tracing transaction", "tx", txHash.String())

	// Trace internal calls of the transaction
	// Reference: https://github.com/ethereum/go-ethereum/issues/3128
	var resp map[string]interface{}
	type TraceConfig struct {
		Tracer string
	}
	err := c.RPCCall(context.Background(), &resp, "debug_traceTransaction", txHash, &TraceConfig{Tracer: "callTracer"})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func Consensus(c Client) (string, error) {
	log.Debug("fetching consensus info")

	var resp p2p.NodeInfo
	err := c.RPCCall(context.Background(), &resp, "admin_nodeInfo")
	if err != nil {
		return "", err
	}
	if resp.Protocols["istanbul"] != nil {
		return "istanbul", nil
	}
	protocol := resp.Protocols["eth"].(map[string]interface{})
	return protocol["consensus"].(string), nil
}
