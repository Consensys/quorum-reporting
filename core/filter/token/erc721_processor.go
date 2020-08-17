package token

import (
	"math/big"
	"sort"

	"quorumengineering/quorum-report/types"
)

var (
	// erc721TransferTopicHash is the topic hash for an ERC721 Transfer event
	erc721TransferTopicHash = types.NewHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

type ERC721Processor struct {
	db TokenFilterDatabase
}

func NewERC721Processor(database TokenFilterDatabase) *ERC721Processor {
	return &ERC721Processor{db: database}
}

func (p *ERC721Processor) ProcessBlock(lastFiltered []types.Address, block *types.Block) error {
	events := make([]*types.Event, 0)
	for _, tx := range block.Transactions {
		transaction, err := p.db.ReadTransaction(tx)
		if err != nil {
			return err
		}
		events = append(events, transaction.Events...)
	}
	erc721Events := p.filterForErc721Events(lastFiltered, events)
	mappedTokens := p.MapEventsToHolders(erc721Events)
	return p.SaveTokenTransfers(mappedTokens, block.Number)
}

func (p *ERC721Processor) SaveTokenTransfers(tokenTransfers map[types.Address]map[string]types.Address, blockNum uint64) error {
	for contract, holderMap := range tokenTransfers {
		for token, holder := range holderMap {
			convertedToken := types.NewHexData(token)
			tokenId := new(big.Int).SetBytes(convertedToken.AsBytes())

			if err := p.db.RecordERC721Token(contract, holder, blockNum, tokenId); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *ERC721Processor) MapEventsToHolders(erc721TransferEvents []*types.Event) map[types.Address]map[string]types.Address {
	sortFunc := func(i, j int) bool { return erc721TransferEvents[i].Index < erc721TransferEvents[j].Index }
	sort.Slice(erc721TransferEvents, sortFunc)

	mappedTransfers := make(map[types.Address]map[string]types.Address)

	for _, erc721Event := range erc721TransferEvents {
		if mappedTransfers[erc721Event.Address] == nil {
			mappedTransfers[erc721Event.Address] = make(map[string]types.Address)
		}

		recipientAddressHex := string(erc721Event.Topics[2])[24:64] //only take the last 40 chars (20 bytes)
		recipientAddress := types.NewAddress(recipientAddressHex)

		tokenId := erc721Event.Topics[3].String()

		//this will overwrite the previous token holder, if there was another receiver
		//of this token in this block
		//this means the resolution of owning tokens is at the block level
		mappedTransfers[erc721Event.Address][tokenId] = recipientAddress
	}
	return mappedTransfers
}

// filterForErc721Events filters out all non-ERC721 transfer events, returning
// on the events we are interested in processing further
func (p *ERC721Processor) filterForErc721Events(lastFiltered []types.Address, events []*types.Event) []*types.Event {
	addrs := make(map[types.Address]bool)
	for _, addr := range lastFiltered {
		addrs[addr] = true
	}

	erc721TransferEvents := make([]*types.Event, 0, len(events))
	for _, event := range events {
		isErc721Transfer := (len(event.Topics) == 4) && (event.Topics[0] == erc721TransferTopicHash)
		if addrs[event.Address] && isErc721Transfer {
			erc721TransferEvents = append(erc721TransferEvents, event)
		}
	}
	return erc721TransferEvents
}
