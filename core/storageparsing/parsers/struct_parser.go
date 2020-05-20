package parsers

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/types"
)

func (p *Parser) ParseStruct(entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) ([]*types.StorageItem, error) {
	currentSlot := common.BigToHash(bigN(entry.Slot))
	newOffset := p.ResolveSlot(currentSlot.Big())

	newTemplate := types.SolidityStorageDocument{
		Storage: namedType.Members,
		Types:   p.template.Types,
	}

	structParser := NewParser(p.storageManager, newTemplate, newOffset)
	return structParser.ParseRawStorage()
}
