package token

import (
	"math/big"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

const erc20AbiString = `[{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"supply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_from","type":"address"},{"indexed":true,"name":"_to","type":"address"},{"indexed":false,"name":"_value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_owner","type":"address"},{"indexed":true,"name":"_spender","type":"address"},{"indexed":false,"name":"_value","type":"uint256"}],"name":"Approval","type":"event"}]`

var (
	// erc20TransferTopicHash is the topic hash for an ERC20 Transfer event
	erc20TransferTopicHash = types.NewHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	erc20Abi, _            = types.NewABIStructureFromJSON(erc20AbiString)
)

type ERC20Processor struct {
	db     TokenFilterDatabase
	client client.Client
}

func NewERC20Processor(database TokenFilterDatabase, client client.Client) *ERC20Processor {
	return &ERC20Processor{db: database, client: client}
}

func (p *ERC20Processor) ProcessBlock(lastFilteredWithAbi map[types.Address]string, block *types.BlockWithTransactions) error {
	addressesWithChangedBalances := make(map[types.Address]map[types.Address]bool)
	erc20Contracts := p.filterForErc20Contracts(lastFilteredWithAbi)

	for _, tx := range block.Transactions {
		thisTxTokenChanges := p.ChangedTokenHolders(erc20Contracts, tx)
		for contract, holders := range thisTxTokenChanges {
			if addressesWithChangedBalances[contract] == nil {
				addressesWithChangedBalances[contract] = holders
				continue
			}
			for holder := range holders {
				addressesWithChangedBalances[contract][holder] = true
			}
		}
	}

	return p.UpdateBalances(addressesWithChangedBalances, block.Number)
}

func (p *ERC20Processor) filterForErc20Contracts(contractsWithAbi map[types.Address]string) map[types.Address]bool {
	erc20Contracts := make(map[types.Address]bool)

	for address, abi := range contractsWithAbi {
		contractAbi, _ := types.NewABIStructureFromJSON(abi)
		isErc20 := isErc20(contractAbi)

		if isErc20 {
			erc20Contracts[address] = true
		}
	}

	return erc20Contracts
}

func (p *ERC20Processor) UpdateBalances(addressesWithChangedBalances map[types.Address]map[types.Address]bool, blockNum uint64) error {
	for contract, tokenHolders := range addressesWithChangedBalances {
		for tokenHolder := range tokenHolders {
			bal, err := client.CallBalanceOfERC20(p.client, contract, tokenHolder, blockNum)
			if err != nil {
				return err
			}

			balance := new(big.Int).SetBytes(bal.AsBytes())
			if err := p.db.RecordNewERC20Balance(contract, tokenHolder, blockNum, balance); err != nil {
				return err
			}
		}
	}
	return nil
}

// ChangedTokenHolders filters through all events in the transaction and
// returns a list of all the token holders who have had a balance change
func (p *ERC20Processor) ChangedTokenHolders(lastFilteredWithAbi map[types.Address]bool, tx *types.Transaction) map[types.Address]map[types.Address]bool {
	//find all ERC20 transfer events
	erc20TransferEvents := p.filterForErc20Events(lastFilteredWithAbi, tx.Events)

	//find all senders and recipients for each token
	return p.filterErc20EventsForAddresses(erc20TransferEvents)
}

func (p *ERC20Processor) filterErc20EventsForAddresses(erc20TransferEvents []*types.Event) map[types.Address]map[types.Address]bool {
	//find all senders and recipients for each token
	addressesWithChangedBalances := make(map[types.Address]map[types.Address]bool)

	for _, event := range erc20TransferEvents {
		firstAddressHex := string(event.Topics[1])[24:64]  //only take the last 40 chars (20 bytes)
		secondAddressHex := string(event.Topics[2])[24:64] //only take the last 40 chars (20 bytes)

		if addressesWithChangedBalances[event.Address] == nil {
			addressesWithChangedBalances[event.Address] = make(map[types.Address]bool)
		}

		addressesWithChangedBalances[event.Address][types.NewAddress(firstAddressHex)] = true
		addressesWithChangedBalances[event.Address][types.NewAddress(secondAddressHex)] = true
	}

	return addressesWithChangedBalances
}

// filterForErc20Events filters out all non-ERC20 transfer events, returning
// on the events we are interested in processing further
func (p *ERC20Processor) filterForErc20Events(lastFiltered map[types.Address]bool, events []*types.Event) []*types.Event {
	erc20TransferEvents := make([]*types.Event, 0, len(events))
	for _, event := range events {
		isErc20Transfer := (len(event.Topics) == 3) && (event.Topics[0] == erc20TransferTopicHash)
		if lastFiltered[event.Address] && isErc20Transfer {
			erc20TransferEvents = append(erc20TransferEvents, event)
		}
	}
	return erc20TransferEvents
}

func isErc20(contractAbi types.ABIStructure) bool {
	for _, erc20Event := range erc20Abi.ToInternalABI().Events {
		found := false
		for _, contractEvent := range contractAbi.ToInternalABI().Events {
			if erc20Event.Signature() == contractEvent.Signature() {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	for _, erc20Method := range erc20Abi.ToInternalABI().Functions {
		found := false
		for _, contractMethod := range contractAbi.ToInternalABI().Functions {
			if erc20Method.Signature() == contractMethod.Signature() {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}
