package token

import (
	"math/big"
	"sort"

	"quorumengineering/quorum-report/types"
)

const erc721AbiString = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"_owner","type":"address"},{"indexed":true,"internalType":"address","name":"_approved","type":"address"},{"indexed":true,"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"_owner","type":"address"},{"indexed":true,"internalType":"address","name":"_operator","type":"address"},{"indexed":false,"internalType":"bool","name":"_approved","type":"bool"}],"name":"ApprovalForAll","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"_from","type":"address"},{"indexed":true,"internalType":"address","name":"_to","type":"address"},{"indexed":true,"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"_approved","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"approve","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"getApproved","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"},{"internalType":"address","name":"_operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"},{"internalType":"bytes","name":"_data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_operator","type":"address"},{"internalType":"bool","name":"_approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

var (
	// erc721TransferTopicHash is the topic hash for an ERC721 Transfer event
	erc721TransferTopicHash = types.NewHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	erc721Abi, _            = types.NewABIStructureFromJSON(erc721AbiString)
)

type ERC721Processor struct {
	db TokenFilterDatabase
}

func NewERC721Processor(database TokenFilterDatabase) *ERC721Processor {
	return &ERC721Processor{db: database}
}

func (p *ERC721Processor) ProcessBlock(lastFilteredWithAbi map[types.Address]string, block *types.Block) error {
	erc721Contracts := p.filterForErc721Contracts(lastFilteredWithAbi)

	events := make([]*types.Event, 0)
	for _, tx := range block.Transactions {
		transaction, err := p.db.ReadTransaction(tx)
		if err != nil {
			return err
		}
		events = append(events, transaction.Events...)
	}
	erc721Events := p.filterForErc721Events(erc721Contracts, events)
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

func (p *ERC721Processor) filterForErc721Contracts(contractsWithAbi map[types.Address]string) map[types.Address]bool {
	erc721Contracts := make(map[types.Address]bool)

	for address, abi := range contractsWithAbi {
		contractAbi, _ := types.NewABIStructureFromJSON(abi)
		isErc721 := isErc721(contractAbi)

		if isErc721 {
			erc721Contracts[address] = true
		}
	}

	return erc721Contracts
}

func isErc721(contractAbi types.ABIStructure) bool {
	for _, erc721Event := range erc721Abi.ToInternalABI().Events {
		found := false
		for _, contractEvent := range contractAbi.ToInternalABI().Events {
			if erc721Event.Signature() == contractEvent.Signature() {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	for _, erc721Method := range erc721Abi.ToInternalABI().Functions {
		found := false
		for _, contractMethod := range contractAbi.ToInternalABI().Functions {
			if erc721Method.Signature() == contractMethod.Signature() {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}
