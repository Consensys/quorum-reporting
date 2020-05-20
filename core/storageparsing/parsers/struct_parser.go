package parsers

import (
	"github.com/ethereum/go-ethereum/common"
	"quorumengineering/quorum-report/core/storageparsing/types"
)

func (p *Parser) ParseStruct(entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) ([]*types.StorageItem, error) {
	existingOffset := p.slotOffset
	currentSlot := common.BigToHash(bigN(entry.Slot))
	newOffset := ResolveSlot(existingOffset.Big(), currentSlot.Big())

	newTemplate := types.SolidityStorageDocument{
		Storage: namedType.Members,
		Types:   p.template.Types,
	}

	structParser := NewParser(p.storageManager, newTemplate, newOffset)
	return structParser.ParseRawStorage()
}
