package monitor

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type TokenRule struct {
	scope        string
	deployer     types.Address
	templateName string
	eip165       string
	abi          *types.ContractABI
}

type AddressWithMeta struct {
	address  types.Address
	scope    string
	deployer types.Address
}

type TokenMonitor interface {
	InspectTransaction(tx *types.Transaction) (map[types.Address]string, error)
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

func (tm *DefaultTokenMonitor) InspectTransaction(tx *types.Transaction) (map[types.Address]string, error) {
	var addresses []AddressWithMeta
	if !tx.CreatedContract.IsEmpty() {
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

	tokenContracts := make(map[types.Address]string)

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
			contractBytecode, err := client.GetCode(tm.quorumClient, types.NewAddress(addressWithMeta.address.Hex()), tx.BlockHash) //TODO: remove
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
		if !rule.deployer.IsEmpty() && rule.deployer != meta.deployer {
			return false
		}
	}
	return true
}

func (tm *DefaultTokenMonitor) checkEIP165(rule TokenRule, addr types.Address, blockNum uint64) (string, error) {
	address := types.NewAddress(addr.Hex()) //TODO: remove
	if rule.eip165 != "" {
		//check if the contract implements EIP165
		eip165Call, err := client.CallEIP165(tm.quorumClient, address, common.Hex2Bytes("01ffc9a70"), blockNum)
		if err != nil {
			return "", err
		}
		if !eip165Call {
			return "", nil
		}

		eip165CallCheck, err := client.CallEIP165(tm.quorumClient, address, common.Hex2Bytes("ffffffff"), blockNum)
		if err != nil {
			return "", err
		}
		if eip165CallCheck {
			return "", nil
		}

		//now we know it implements EIP165, so lets check the interfaces
		detected, err := client.CallEIP165(tm.quorumClient, address, common.Hex2Bytes(rule.eip165), blockNum)
		if err != nil {
			return "", err
		}
		if detected {
			return rule.templateName, nil
		}
	}
	return "", nil
}

func (tm *DefaultTokenMonitor) checkBytecodeForTokens(rule TokenRule, data types.HexData) string {
	if tm.checkAbiMatch(rule.abi, data) {
		return rule.templateName
	}
	return ""
}

func (tm *DefaultTokenMonitor) checkAbiMatch(abiToCheck *types.ContractABI, data types.HexData) bool {
	for _, b := range abiToCheck.Functions {
		if !strings.Contains(data.String(), b.Signature()) {
			return false
		}
	}
	for _, event := range abiToCheck.Events {
		if !strings.Contains(data.String(), event.Signature()) {
			return false
		}
	}
	return true
}
