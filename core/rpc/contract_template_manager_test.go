package rpc

import (
	"quorumengineering/quorum-report/database/memory"
	"testing"

	"github.com/consensys/quorum-go-utils/types"
	"github.com/stretchr/testify/assert"
)

func TestDefaultContractManager_AddContractABI_ExistingTemplate(t *testing.T) {
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	db := memory.NewMemoryDB()
	_ = db.AddTemplate("sample template", "sample abi", "sample layout")
	_ = db.AssignTemplate(address, "sample template")

	contractManager := NewDefaultContractManager(db)

	err := contractManager.AddContractABI(address, "new sample abi")
	assert.Nil(t, err)

	allTemplates, _ := db.GetTemplates()
	assert.Equal(t, 2, len(allTemplates))
	assert.Contains(t, allTemplates, address.Hex()) //check a new template was made with the addresses hex form

	template, _ := db.GetTemplateDetails(address.Hex())
	assert.Equal(t, address.Hex(), template.TemplateName)
	assert.Equal(t, "new sample abi", template.ABI)
	assert.Equal(t, "sample layout", template.StorageLayout)

	templateName, _ := db.GetContractTemplate(address)
	assert.Equal(t, address.Hex(), templateName)
}

func TestDefaultContractManager_AddContractABI_NoExistingTemplate(t *testing.T) {
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	db := memory.NewMemoryDB()
	contractManager := NewDefaultContractManager(db)

	err := contractManager.AddContractABI(address, "new sample abi")
	assert.Nil(t, err)

	allTemplates, _ := db.GetTemplates()
	assert.Equal(t, 1, len(allTemplates))
	assert.Contains(t, allTemplates, address.Hex()) //check a new template was made with the addresses hex form

	template, _ := db.GetTemplateDetails(address.Hex())
	assert.Equal(t, address.Hex(), template.TemplateName)
	assert.Equal(t, "new sample abi", template.ABI)
	assert.Equal(t, "", template.StorageLayout)

	templateName, _ := db.GetContractTemplate(address)
	assert.Equal(t, address.Hex(), templateName)
}

func TestDefaultContractManager_AddStorageLayout_ExistingTemplate(t *testing.T) {
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")

	db := memory.NewMemoryDB()
	_ = db.AddTemplate("sample template", "sample abi", "sample layout")
	_ = db.AssignTemplate(address, "sample template")

	contractManager := NewDefaultContractManager(db)

	err := contractManager.AddStorageLayout(address, "new sample layout")
	assert.Nil(t, err)

	allTemplates, _ := db.GetTemplates()
	assert.Equal(t, 2, len(allTemplates))
	assert.Contains(t, allTemplates, address.Hex()) //check a new template was made with the addresses hex form

	template, _ := db.GetTemplateDetails(address.Hex())
	assert.Equal(t, address.Hex(), template.TemplateName)
	assert.Equal(t, "sample abi", template.ABI)
	assert.Equal(t, "new sample layout", template.StorageLayout)

	templateName, _ := db.GetContractTemplate(address)
	assert.Equal(t, address.Hex(), templateName)
}

func TestDefaultContractManager_AddStorageLayout_NoExistingTemplate(t *testing.T) {
	address := types.NewAddress("0x1349f3e1b8d71effb47b840594ff27da7e603d17")
	db := memory.NewMemoryDB()
	contractManager := NewDefaultContractManager(db)

	err := contractManager.AddStorageLayout(address, "new sample layout")
	assert.Nil(t, err)

	allTemplates, _ := db.GetTemplates()
	assert.Equal(t, 1, len(allTemplates))
	assert.Contains(t, allTemplates, address.Hex()) //check a new template was made with the addresses hex form

	template, _ := db.GetTemplateDetails(address.Hex())
	assert.Equal(t, address.Hex(), template.TemplateName)
	assert.Equal(t, "", template.ABI)
	assert.Equal(t, "new sample layout", template.StorageLayout)

	templateName, _ := db.GetContractTemplate(address)
	assert.Equal(t, address.Hex(), templateName)
}
