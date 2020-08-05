package token

import (
	"encoding/hex"
	"math/big"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

var (
	// erc721TransferTopicHash is the topic hash for an ERC721 Transfer event
	erc721TransferTopicHash = types.NewHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

type ERC721Processor struct {
	db     TokenFilterDatabase
	client client.Client
}

func NewERC721Processor(database TokenFilterDatabase, client client.Client) *ERC721Processor {
	return &ERC721Processor{db: database, client: client}
}

func (p *ERC721Processor) ProcessBlock(lastFiltered []types.Address, block *types.Block) {
	for _, tx := range block.Transactions {
		transaction, _ := p.db.ReadTransaction(tx)
		p.ProcessTransaction(lastFiltered, transaction)
	}
}

func (p *ERC721Processor) ProcessTransaction(lastFiltered []types.Address, tx *types.Transaction) {

	//find all ERC721 transfer events
	addrs := make(map[types.Address]bool)
	for _, addr := range lastFiltered {
		addrs[addr] = true
	}
	erc721TransferEvents := p.filterForErc721Events(addrs, tx.Events)

	for _, erc721Event := range erc721TransferEvents {
		//firstAddressHex := string(erc721Event.Topics[1])[24:64]  //only take the last 40 chars (20 bytes)
		secondAddressHex := string(erc721Event.Topics[2])[24:64] //only take the last 40 chars (20 bytes)

		tokenIdAsBytes, _ := hex.DecodeString(string(erc721Event.Topics[3]))
		tokenId := new(big.Int).SetBytes(tokenIdAsBytes)
		p.db.RecordERC721Token(erc721Event.Address, types.NewAddress(secondAddressHex), tx.BlockNumber, tokenId)
	}
}

func (p *ERC721Processor) filterErc721EventsForAddresses(erc721TransferEvents []*types.Event) map[types.Address]map[types.Address]bool {
	//find all senders and recipients for each token
	addressesWithChangedBalances := make(map[types.Address]map[types.Address]bool)

	//TODO: assuming that it follows the ERC721 spec with indexed args in the event - is this always true?
	for _, event := range erc721TransferEvents {
		firstAddressHex := string(event.Topics[1])[24:64]  //only take the last 40 chars (721 bytes)
		secondAddressHex := string(event.Topics[2])[24:64] //only take the last 40 chars (721 bytes)

		if addressesWithChangedBalances[event.Address] == nil {
			addressesWithChangedBalances[event.Address] = make(map[types.Address]bool)
		}

		addressesWithChangedBalances[event.Address][types.NewAddress(firstAddressHex)] = true
		addressesWithChangedBalances[event.Address][types.NewAddress(secondAddressHex)] = true
	}

	return addressesWithChangedBalances
}

func (p *ERC721Processor) filterForErc721Events(lastFiltered map[types.Address]bool, events []*types.Event) []*types.Event {
	// only keep erc721 events
	erc721TransferEvents := make([]*types.Event, 0, len(events))
	for _, event := range events {
		if len(event.Topics) == 4 && event.Topics[0] == erc721TransferTopicHash {
			erc721TransferEvents = append(erc721TransferEvents, event)
		}
	}

	// only keep events from addresses we are filtering on
	filteredAddressTransferEvents := make([]*types.Event, 0, len(erc721TransferEvents))
	for _, event := range erc721TransferEvents {
		if lastFiltered[event.Address] {
			filteredAddressTransferEvents = append(filteredAddressTransferEvents, event)
		}
	}

	return filteredAddressTransferEvents
}
