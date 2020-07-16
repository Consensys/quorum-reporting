package rpc

import (
	"quorumengineering/quorum-report/types"

	"quorumengineering/quorum-report/database"
)

type ContractTemplateManager interface {
	AddStorageLayout(address types.Address, layout string) error
	AddContractABI(address types.Address, abi string) error
}

type DefaultContractTemplateManager struct {
	db database.Database
}

func NewDefaultContractManager(db database.Database) *DefaultContractTemplateManager {
	return &DefaultContractTemplateManager{
		db: db,
	}
}

func (cm *DefaultContractTemplateManager) AddStorageLayout(address types.Address, layout string) error {
	// check contract & template existence before updating
	templateName, err := cm.db.GetContractTemplate(address)
	if err != nil {
		return err
	}

	// create new template named contract.Address.String()
	template, err := cm.db.GetTemplateDetails(templateName)
	if err != nil && err != database.ErrNotFound {
		return err
	}

	if err == nil {
		if err := cm.db.AddTemplate(address.String(), template.ABI, layout); err != nil {
			return err
		}
	} else {
		if err := cm.db.AddTemplate(address.String(), "", layout); err != nil {
			return err
		}
	}

	return cm.db.AssignTemplate(address, address.String())
}

func (cm *DefaultContractTemplateManager) AddContractABI(address types.Address, abi string) error {
	// check contract & template existence before updating
	templateName, err := cm.db.GetContractTemplate(address)
	if err != nil {
		return err
	}

	// create new template named contract.Address.String()
	template, err := cm.db.GetTemplateDetails(templateName)
	if err != nil && err != database.ErrNotFound {
		return err
	}

	if err == nil {
		if err := cm.db.AddTemplate(address.String(), abi, template.StorageLayout); err != nil {
			return err
		}
	} else {
		if err := cm.db.AddTemplate(address.String(), abi, ""); err != nil {
			return err
		}
	}

	return cm.db.AssignTemplate(address, address.String())
}
