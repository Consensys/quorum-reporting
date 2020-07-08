package monitor

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type TokenRule struct {
	scope        string
	deployer     common.Address
	templateName string
	eip165       string
	abi          *types.ContractABI
}

type AddressWithMeta struct {
	address  common.Address
	scope    string
	deployer common.Address
}

type TokenMonitor interface {
	InspectTransaction(tx *types.Transaction) (map[common.Address]string, error)
}

type DefaultTokenMonitor struct {
	quorumClient client.Client
	rules        []TokenRule
}

func NewDefaultTokenMonitor(quorumClient client.Client, rules []TokenRule) *DefaultTokenMonitor {
	return &DefaultTokenMonitor{
		quorumClient: quorumClient,
		rules:        rules,
	}
}

func (tm *DefaultTokenMonitor) InspectTransaction(tx *types.Transaction) (map[common.Address]string, error) {
	var addresses []AddressWithMeta
	if (tx.CreatedContract != common.Address{}) {
		addresses = append(addresses, AddressWithMeta{
			address:  tx.CreatedContract,
			scope:    types.ExternalScope,
			deployer: tx.From,
		})
	}
	for _, ic := range tx.InternalCalls {
		if ic.Type == "CREATE" || ic.Type == "CREATE2" {
			addresses = append(addresses, AddressWithMeta{
				address:  ic.To,
				scope:    types.InternalScope,
				deployer: ic.From,
			})
		}
	}

	tokenContracts := make(map[common.Address]string)

	for _, addressWithMeta := range addresses {
		for _, rule := range tm.rules {
			if !tm.checkRuleMeta(rule, addressWithMeta) {
				continue
			}
			// EIP165
			contractType, err := tm.checkEIP165(rule, addressWithMeta.address, tx.BlockNumber)
			if err != nil {
				return nil, err
			}
			if contractType != "" {
				log.Info("Contract implemented interface via ERC165", "interface", contractType, "address", addressWithMeta.address.String())
				tokenContracts[addressWithMeta.address] = contractType
				break
			}

			// Check contract bytecode directly for all 4bytes presented in abi
			contractBytecode, err := client.GetCode(tm.quorumClient, addressWithMeta.address, tx.BlockHash)
			if err != nil {
				return nil, err
			}
			contractType = tm.checkBytecodeForTokens(rule, contractBytecode)
			if contractType != "" {
				log.Info("Transaction deploys potential token", "type", contractType, "tx", tx.Hash.Hex(), "address", addressWithMeta.address.Hex())
				tokenContracts[addressWithMeta.address] = contractType
				break
			}
		}
	}

	return tokenContracts, nil
}

func (tm *DefaultTokenMonitor) checkRuleMeta(rule TokenRule, meta AddressWithMeta) bool {
	// check scope & deployer
	if rule.scope != types.AllScope {
		if rule.scope != meta.scope {
			return false
		}
		if rule.deployer != (common.Address{0}) && rule.deployer != meta.deployer {
			return false
		}
	}
	return true
}

func (tm *DefaultTokenMonitor) checkEIP165(rule TokenRule, address common.Address, blockNum uint64) (string, error) {
	if rule.eip165 != "" {
		//check if the contract implements EIP165
		eip165Call, err := client.CallEIP165(tm.quorumClient, address, common.Hex2Bytes("01ffc9a70"), new(big.Int).SetUint64(blockNum))
		if err != nil {
			return "", err
		}
		if !eip165Call {
			return "", nil
		}

		eip165CallCheck, err := client.CallEIP165(tm.quorumClient, address, common.Hex2Bytes("ffffffff"), new(big.Int).SetUint64(blockNum))
		if err != nil {
			return "", err
		}
		if eip165CallCheck {
			return "", nil
		}

		//now we know it implements EIP165, so lets check the interfaces
		detected, err := client.CallEIP165(tm.quorumClient, address, common.Hex2Bytes(rule.eip165), new(big.Int).SetUint64(blockNum))
		if err != nil {
			return "", err
		}
		if detected {
			return rule.templateName, nil
		}
	}
	return "", nil
}

func (tm *DefaultTokenMonitor) checkBytecodeForTokens(rule TokenRule, data hexutil.Bytes) string {
	if tm.checkAbiMatch(rule.abi, data) {
		return rule.templateName
	}
	return ""
}

func (tm *DefaultTokenMonitor) checkAbiMatch(abiToCheck *types.ContractABI, data hexutil.Bytes) bool {
	for _, b := range abiToCheck.Functions {
		if !strings.Contains(data.String(), b.Signature()) {
			return false
		}
	}
	for _, event := range abiToCheck.Events {
		if !strings.Contains(data.String(), event.Signature()[2:]) {
			return false
		}
	}
	return true
}
