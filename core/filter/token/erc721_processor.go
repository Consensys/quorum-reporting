package token

import (
	"encoding/hex"
	"math/big"

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
	for _, tx := range block.Transactions {
		transaction, err := p.db.ReadTransaction(tx)
		if err != nil {
			return err
		}
		if err := p.ProcessTransaction(lastFiltered, transaction); err != nil {
			return err
		}
	}
	return nil
}

func (p *ERC721Processor) ProcessTransaction(lastFiltered []types.Address, tx *types.Transaction) error {
	//find all ERC721 transfer events
	addrs := make(map[types.Address]bool)
	for _, addr := range lastFiltered {
		addrs[addr] = true
	}
	erc721TransferEvents := p.filterForErc721Events(addrs, tx.Events)

	for _, erc721Event := range erc721TransferEvents {
		recipientAddressHex := string(erc721Event.Topics[2])[24:64] //only take the last 40 chars (20 bytes)
		recipientAddress := types.NewAddress(recipientAddressHex)

		tokenIdAsBytes, _ := hex.DecodeString(string(erc721Event.Topics[3]))
		tokenId := new(big.Int).SetBytes(tokenIdAsBytes)
		if err := p.db.RecordERC721Token(erc721Event.Address, recipientAddress, tx.BlockNumber, tokenId); err != nil {
			return err
		}
	}
	return nil
}

// filterForErc721Events filters out all non-ERC721 transfer events, returning
// on the events we are interested in processing further
func (p *ERC721Processor) filterForErc721Events(lastFiltered map[types.Address]bool, events []*types.Event) []*types.Event {
	erc721TransferEvents := make([]*types.Event, 0, len(events))
	for _, event := range events {
		isErc721Transfer := (len(event.Topics) == 4) && (event.Topics[0] == erc721TransferTopicHash)
		if lastFiltered[event.Address] && isErc721Transfer {
			erc721TransferEvents = append(erc721TransferEvents, event)
		}
	}
	return erc721TransferEvents
}
