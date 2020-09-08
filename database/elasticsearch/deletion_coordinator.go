package elasticsearch

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
	"strings"
)

const (
	DeleteQueryContract = `{ "query": { "match": { "contract": "%s" } } }`
	DeleteQueryAddress  = `{ "query": { "match": { "address": "%s" } } }`
)

// Delete requests need a pointer value, so this is used instead of creating a new variable every request
var RequestParameterTrue = true

//go:generate mockgen -destination=./mocks/deletion_coordiantor_mock.go -package elasticsearch_mocks . DeletionCoordinator
type DeletionCoordinator interface {
	Delete(contract types.Address) error
}

type DefaultDeletionCoordinator struct {
	apiClient APIClient
}

func NewDefaultDeletionCoordinator(apiClient APIClient) *DefaultDeletionCoordinator {
	return &DefaultDeletionCoordinator{
		apiClient: apiClient,
	}
}

func (coordinator *DefaultDeletionCoordinator) Delete(contract types.Address) error {
	deleteByAddressQuery := fmt.Sprintf(DeleteQueryAddress, contract.String())
	deleteByContractQuery := fmt.Sprintf(DeleteQueryContract, contract.String())

	// delete ERC20 & ERC721 tokens
	log.Debug("Deleting ERC20/ERC721 token data", "contract", contract.String())
	erc20Req := esapi.DeleteByQueryRequest{
		Index:             []string{ERC20TokenIndex, ERC721TokenIndex},
		Body:              strings.NewReader(deleteByContractQuery),
		Refresh:           &RequestParameterTrue,
		WaitForCompletion: &RequestParameterTrue,
	}
	_, err := coordinator.apiClient.DoRequest(erc20Req)
	if err != nil {
		return err
	}
	log.Debug("Deleted ERC20/ERC721 token data", "contract", contract.String())

	//delete event
	log.Debug("Deleting contract events", "contract", contract.String())
	eventReq := esapi.DeleteByQueryRequest{
		Index:             []string{EventIndex},
		Body:              strings.NewReader(deleteByAddressQuery),
		Refresh:           &RequestParameterTrue,
		WaitForCompletion: &RequestParameterTrue,
	}
	_, err = coordinator.apiClient.DoRequest(eventReq)
	if err != nil {
		return err
	}
	log.Debug("Deleted contract events", "contract", contract.String())

	log.Debug("Deleting contract storage", "contract", contract.String())
	storageDeleteReq := esapi.DeleteByQueryRequest{
		Index:             []string{StorageIndex},
		Body:              strings.NewReader(deleteByContractQuery),
		Refresh:           &RequestParameterTrue,
		WaitForCompletion: &RequestParameterTrue,
	}
	_, err = coordinator.apiClient.DoRequest(storageDeleteReq)
	if err != nil {
		return err
	}
	log.Debug("Deleted contract storage", "contract", contract.String())

	//delete template if specialised
	log.Debug("Deleting contract template", "contract", contract.String())
	deleteRequest := esapi.DeleteRequest{
		Index:      TemplateIndex,
		DocumentID: contract.String(),
		Refresh:    "true",
	}
	_, err = coordinator.apiClient.DoRequest(deleteRequest)
	if err != nil && err != database.ErrNotFound {
		return err
	}
	log.Debug("Deleted contract template", "contract", contract.String())

	//delete contract
	log.Debug("Deleting contract", "contract", contract.String())
	deleteContractRequest := esapi.DeleteRequest{
		Index:      ContractIndex,
		DocumentID: contract.String(),
		Refresh:    "true",
	}
	_, err = coordinator.apiClient.DoRequest(deleteContractRequest)
	log.Debug("Deleted contract", "contract", contract.String())
	return err
}
