package filter

import (
	"encoding/hex"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

var ContractExtensionTopic = types.NewHash("0x67a92539f3cbd7c5a9b36c23c0e2beceb27d2e1b3cd8eda02c623689267ae71e")

type ContractCreationFilter struct {
	db           FilterServiceDB
	quorumClient client.Client
}

func NewContractCreationFilter(db FilterServiceDB, quorumClient client.Client) *ContractCreationFilter {
	return &ContractCreationFilter{
		db:           db,
		quorumClient: quorumClient,
	}
}

func (ccFilter *ContractCreationFilter) ProcessBlocks(indexedAddresses []types.Address, blocks []*types.BlockWithTransactions) error {
	log.Debug("Filtering for contract creations")
	defer func() { log.Debug("Finished filtering for contract creations") }()

	addrMap := make(map[types.Address]bool)
	for _, addr := range indexedAddresses {
		addrMap[addr] = true
	}

	allDeployedContacts := make(map[types.Hash][]types.Address)
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			deployedContracts, err := ccFilter.findDeployedContracts(tx)
			if err != nil {
				return err
			}

			filteredDeployed := filterDeployedContracts(addrMap, deployedContracts)
			if len(filteredDeployed) != 0 {
				allDeployedContacts[tx.Hash] = filteredDeployed
			}
		}
	}

	//save all the deployed contract updates
	return ccFilter.db.SetContractCreationTransaction(allDeployedContacts)
}

func (ccFilter *ContractCreationFilter) findDeployedContracts(tx *types.Transaction) ([]types.Address, error) {
	deployedContracts := make([]types.Address, 0)

	// Check for external deployment
	if !tx.IsPrivate || !tx.PrivateData.IsEmpty() {
		deployedContracts = append(deployedContracts, tx.CreatedContract)
	}

	// Check all internal deployments
	for _, internalCall := range tx.InternalCalls {
		if (internalCall.Type == "CREATE") || (internalCall.Type == "CREATE2") {
			deployedContracts = append(deployedContracts, internalCall.To)
		}
	}

	// Check for contract extension deployment
	for _, event := range tx.Events {
		if len(event.Topics) == 1 && event.Topics[0] == ContractExtensionTopic {
			//this is an extension tx
			//first 64 chars (32 bytes) of data are the address
			addressBytes := event.Data.AsBytes()[12:32]
			address := types.NewAddress(hex.EncodeToString(addressBytes))

			//check if the code exists to tell if the extension succeeded
			code, err := client.GetCode(ccFilter.quorumClient, address, tx.BlockNumber-1)
			if err != nil {
				return nil, err
			}
			if code == types.NewHexData("") {
				//not been extended before since the code doesn't exist prior
				deployedContracts = append(deployedContracts, address)
			}
			break
		}
	}

	return deployedContracts, nil
}

func filterDeployedContracts(indexedAddresses map[types.Address]bool, deployed []types.Address) []types.Address {
	filtered := make([]types.Address, 0)
	for _, address := range deployed {
		if indexedAddresses[address] {
			filtered = append(filtered, address)
		}
	}
	return filtered
}
