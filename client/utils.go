package client

import (
	"encoding/hex"
	"errors"
	"fmt"

	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

func DumpAddress(c Client, address types.Address, blockNumber uint64) (*types.AccountState, error) {
	log.Debug("Fetching account dump", "account", address.String(), "blocknumber", blockNumber)
	dumpAccount := &types.RawAccountState{}
	err := c.RPCCall(&dumpAccount, "debug_dumpAddress", address, fmt.Sprintf("0x%x", blockNumber))
	if err != nil {
		return nil, err
	}

	converted := make(map[types.Hash]string)
	for k, v := range dumpAccount.Storage {
		converted[types.NewHash(k)] = v
	}
	return &types.AccountState{Root: dumpAccount.Root, Storage: converted}, nil
}

func TraceTransaction(c Client, txHash types.Hash) (map[string]interface{}, error) {
	log.Debug("Tracing transaction", "tx", txHash.String())

	// Trace internal calls of the transaction
	// Reference: https://github.com/ethereum/go-ethereum/issues/3128
	var resp map[string]interface{}
	type TraceConfig struct {
		Tracer string
	}
	err := c.RPCCall(&resp, "debug_traceTransaction", txHash, &TraceConfig{Tracer: "callTracer"})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func GetCode(c Client, address types.Address, blockHash types.Hash) (types.HexData, error) {
	var res types.HexData
	if err := c.RPCCall(&res, "eth_getCode", address, blockHash.String()); err != nil {
		return "", err
	}
	return res, nil
}

func Consensus(c Client) (string, error) {
	log.Debug("Fetching consensus info")

	var resp map[string]interface{}
	err := c.RPCCall(&resp, "admin_nodeInfo")
	if err != nil {
		return "", err
	}
	if resp["protocols"] == nil {
		return "", errors.New("no consensus info found")
	}
	protocols, ok := resp["protocols"].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid consensus info found")
	}
	if protocols["istanbul"] != nil {
		return "istanbul", nil
	}
	protocol := protocols["eth"].(map[string]interface{})
	return protocol["consensus"].(string), nil
}

func CallEIP165(c Client, address types.Address, interfaceId []byte, blockNum uint64) (bool, error) {
	eip165Id, _ := hex.DecodeString("01ffc9a70")

	//interfaceId should be 4 bytes long
	if len(interfaceId) != 4 {
		return false, errors.New("interfaceId wrong size")
	}

	paddedInterface := make([]byte, 32)
	copy(paddedInterface, interfaceId)
	calldata := append(eip165Id, paddedInterface...)

	msg := types.EIP165Call{
		To:   address,
		Data: types.HexData(hex.EncodeToString(calldata)),
	}

	var res types.HexData
	err := c.RPCCall(&res, "eth_call", msg, fmt.Sprintf("0x%x", blockNum))
	if err != nil {
		return false, err
	}

	asBytes := res.AsBytes()
	if len(asBytes) != 32 {
		return false, nil
	}
	return asBytes[len(asBytes)-1] == 0x1, nil
}
