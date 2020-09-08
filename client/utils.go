package client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

const (
	ethCall          = "eth_call"
	adminInfo        = "admin_nodeInfo"
	dumpAddress      = "debug_dumpAddress"
	traceTransaction = "debug_traceTransaction"
	getCode          = "eth_getCode"
	getBlockByNumber = "eth_getBlockByNumber"
	protocolKey      = "protocols"
	istanbulKey      = "istanbul"
	consensusKey     = "consensus"
	ethKey           = "eth"
)

func DumpAddress(c Client, address types.Address, blockNumber uint64) (*types.AccountState, error) {
	log.Debug("Fetching account dump", "account", address.String(), "blocknumber", blockNumber)
	dumpAccount := &types.RawAccountState{}
	err := c.RPCCall(&dumpAccount, dumpAddress, address.String(), fmtBlockNum(blockNumber))
	if err != nil {
		return nil, err
	}

	converted := make(map[types.Hash]string)
	for k, v := range dumpAccount.Storage {
		converted[types.NewHash(k)] = v
	}
	return &types.AccountState{Root: dumpAccount.Root, Storage: converted}, nil
}

func fmtBlockNum(blockNumber uint64) string {
	return fmt.Sprintf("0x%x", blockNumber)
}

func TraceTransaction(c Client, txHash types.Hash) (map[string]interface{}, error) {
	log.Debug("Tracing transaction", "tx", txHash.String())

	// Trace internal calls of the transaction
	// Reference: https://github.com/ethereum/go-ethereum/issues/3128
	var resp map[string]interface{}
	type TraceConfig struct {
		Tracer string
	}
	err := c.RPCCall(&resp, traceTransaction, txHash.String(), &TraceConfig{Tracer: "callTracer"})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func GetCode(c Client, address types.Address, blockNumber uint64) (types.HexData, error) {
	log.Debug("Querying account code", "account", address.String(), "block number", blockNumber)
	var res types.HexData
	if err := c.RPCCall(&res, getCode, address.String(), fmtBlockNum(blockNumber)); err != nil {
		log.Debug("Error querying account code", "account", address.String(), "block number", blockNumber, "err", err)
		return "", err
	}
	log.Debug("Queried account code", "account", address.String(), "block number", blockNumber, "code", res.String())
	return res, nil
}

func Consensus(c Client) (string, error) {
	log.Debug("Fetching consensus info")

	var resp map[string]interface{}
	err := c.RPCCall(&resp, adminInfo)
	if err != nil {
		return "", err
	}
	if resp[protocolKey] == nil {
		return "", errors.New("no consensus info found")
	}
	protocols, ok := resp[protocolKey].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid consensus info found")
	}
	if protocols[istanbulKey] != nil {
		return "istanbul", nil
	}
	protocol := protocols[ethKey].(map[string]interface{})
	return protocol[consensusKey].(string), nil
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
	err := c.RPCCall(&res, ethCall, msg, fmtBlockNum(blockNum))
	if err != nil {
		return false, err
	}

	asBytes := res.AsBytes()
	if len(asBytes) != 32 {
		return false, nil
	}
	return asBytes[len(asBytes)-1] == 0x1, nil
}

func BlockByNumber(c Client, blockNum uint64) (types.RawBlock, error) {
	var blockOrigin types.RawBlock
	err := c.RPCCall(&blockOrigin, getBlockByNumber, fmtBlockNum(blockNum), false)

	return blockOrigin, err
}

func CurrentBlock(c Client) (uint64, error) {
	log.Debug("Fetching current block number")

	var currentBlockResult CurrentBlockResult
	if err := c.ExecuteGraphQLQuery(&currentBlockResult, CurrentBlockQuery()); err != nil {
		return 0, err
	}

	log.Debug("Current block number found", "number", currentBlockResult.Block.Number)
	return currentBlockResult.Block.Number.ToUint64(), nil
}

func TransactionWithReceipt(c Client, transactionHash types.Hash) (Transaction, error) {
	var txResult TransactionResult
	if err := c.ExecuteGraphQLQuery(&txResult, TransactionDetailQuery(transactionHash)); err != nil {
		return Transaction{}, err
	}
	return txResult.Transaction, nil
}

func CallBalanceOfERC20(c Client, contract types.Address, holder types.Address, blockNum uint64) (types.HexData, error) {
	// 70a08231 is the 4byte function sig for `balanceOf(address)`
	// "000000000000000000000000" + string(holder) is the token holders address, padded to 32 bytes

	blockAsHex := fmtBlockNum(blockNum)
	msg := types.EIP165Call{
		To:   contract,
		Data: types.NewHexData("0x70a08231" + "000000000000000000000000" + string(holder)),
	}

	var res types.HexData
	err := c.RPCCall(&res, ethCall, msg, blockAsHex)
	return res, err
}

func StorageRoot(c Client, account types.Address, blockNum uint64) (types.Hash, error) {
	var res types.Hash
	err := c.RPCCall(&res, "eth_storageRoot", account.String(), fmt.Sprintf("0x%x", blockNum))
	if err != nil && err.Error() == "can't find state object" {
		return types.NewHash(""), nil
	}
	return res, err
}
