package parsers

import (
	"quorumengineering/quorum-report/types"
)

func (p *Parser) ParseStruct(entry types.SolidityStorageEntry, namedType types.SolidityTypeEntry) ([]*types.StorageItem, error) {
	newOffset := p.ResolveSlot(bigN(entry.Slot))
	newTemplate := types.SolidityStorageDocument{
		Storage: namedType.Members,
		Types:   p.template.Types,
	}

	structParser := NewParser(p.storageManager, newTemplate, newOffset)
	return structParser.ParseRawStorage()
}
