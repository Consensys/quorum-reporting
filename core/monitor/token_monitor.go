package monitor

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type TokenMonitor interface {
	InspectAddresses(addressesToFindTokens []common.Address, tx *types.Transaction) (map[common.Address]string, error)
}

type DefaultTokenMonitor struct {
	quorumClient client.Client
}

func NewDefaultTokenMonitor(quorumClient client.Client) *DefaultTokenMonitor {
	return &DefaultTokenMonitor{
		quorumClient: quorumClient,
	}
}

func (dtm *DefaultTokenMonitor) InspectAddresses(addresses []common.Address, tx *types.Transaction) (map[common.Address]string, error) {
	tokenContracts := make(map[common.Address]string)

	for _, addr := range addresses {
		contractType, err := dtm.checkEIP165(addr, tx.BlockNumber)
		if err != nil {
			return nil, err
		}
		if contractType != "" {
			log.Info("Contract implemented interface via ERC165", "interface", contractType, "address", addr.String())
			tokenContracts[addr] = contractType
			continue
		}

		//Check if contract has bytecode for contract types
		contractBytecode, err := client.GetCode(dtm.quorumClient, addr, tx.BlockHash)
		if err != nil {
			return nil, err
		}

		contractType = checkBytecodeForTokens(contractBytecode)
		if contractType != "" {
			log.Info("Transaction deploys potential token", "type", contractType, "tx", tx.Hash.Hex(), "address", addr.Hex())
			tokenContracts[addr] = contractType
		}
	}
	return tokenContracts, nil
}

func (dtm *DefaultTokenMonitor) checkEIP165(address common.Address, blockNum uint64) (string, error) {
	//check if the contract implements EIP165

	eip165Call, err := client.CallEIP165(dtm.quorumClient, address, common.Hex2Bytes("01ffc9a70"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if !eip165Call {
		return "", nil
	}

	eip165CallCheck, err := client.CallEIP165(dtm.quorumClient, address, common.Hex2Bytes("ffffffff"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if eip165CallCheck {
		return "", nil
	}

	//now we know it implements EIP165, so lets check the interfaces

	erc20check, err := client.CallEIP165(dtm.quorumClient, address, common.Hex2Bytes("36372b07"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if erc20check {
		return types.ERC20, nil
	}

	erc721check, err := client.CallEIP165(dtm.quorumClient, address, common.Hex2Bytes("80ac58cd"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if erc721check {
		return types.ERC721, nil
	}

	return "", nil
}

func checkBytecodeForTokens(data hexutil.Bytes) string {
	// check ERC20
	if checkAbiMatch(types.ERC20ABI, data) {
		return types.ERC20
	}
	// check ERC721
	if checkAbiMatch(types.ERC721ABI, data) {
		return types.ERC721
	}
	return ""
}

func checkAbiMatch(abiToCheck abi.ABI, data hexutil.Bytes) bool {
	for _, b := range abiToCheck.Methods {
		if !strings.Contains(data.String(), common.Bytes2Hex(b.ID())) {
			return false
		}
	}
	for _, event := range abiToCheck.Events {
		if !strings.Contains(data.String(), event.ID().Hex()[2:]) {
			return false
		}
	}
	return true
}
