package token

import (
	"math/big"

	"github.com/consensys/quorum-go-utils/client"
	"github.com/consensys/quorum-go-utils/types"
)

var (
	// erc20TransferTopicHash is the topic hash for an ERC20 Transfer event
	erc20TransferTopicHash = types.NewHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

type ERC20Processor struct {
	db     TokenFilterDatabase
	client client.Client
}

func NewERC20Processor(database TokenFilterDatabase, client client.Client) *ERC20Processor {
	return &ERC20Processor{db: database, client: client}
}

func (p *ERC20Processor) ProcessBlock(lastFiltered []types.Address, block *types.Block) error {
	addressesWithChangedBalances := make(map[types.Address]map[types.Address]bool)

	for _, tx := range block.Transactions {
		transaction, err := p.db.ReadTransaction(tx)
		if err != nil {
			return err
		}

		thisTxTokenChanges := p.ChangedTokenHolders(lastFiltered, transaction)
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

func (p *ERC20Processor) ProcessTransaction(lastFiltered []types.Address, tx *types.Transaction) error {
	//find all senders and recipients for each token
	addressesWithChangedBalances := p.ChangedTokenHolders(lastFiltered, tx)

	return p.UpdateBalances(addressesWithChangedBalances, tx.BlockNumber)
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
func (p *ERC20Processor) ChangedTokenHolders(lastFiltered []types.Address, tx *types.Transaction) map[types.Address]map[types.Address]bool {
	//find all ERC20 transfer events
	addrs := make(map[types.Address]bool)
	for _, addr := range lastFiltered {
		addrs[addr] = true
	}
	erc20TransferEvents := p.filterForErc20Events(addrs, tx.Events)

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
