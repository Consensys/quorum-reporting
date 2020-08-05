package elasticsearch

import (
	"quorumengineering/quorum-report/types"
)

type Contract struct {
	Address             types.Address `json:"address"`
	TemplateName        string        `json:"templateName"`
	CreationTransaction types.Hash    `json:"creationTx"`
	LastFiltered        uint64        `json:"lastFiltered"`
}

type Template struct {
	TemplateName string `json:"templateName"`
	ABI          string `json:"abi"`
	StorageABI   string `json:"storageAbi"`
}

type State struct {
	Address     types.Address `json:"address"`
	BlockNumber uint64        `json:"blockNumber"`
	StorageRoot types.Hash    `json:"storageRoot"`
}

type Storage struct {
	StorageRoot types.Hash     `json:"storageRoot"`
	StorageMap  []StorageEntry `json:"storageMap"`
}

type StorageEntry struct {
	Key   types.Hash
	Value string
}

type TokenHolder struct {
	Contract    types.Address `json:"contract"`
	Holder      types.Address `json:"holder"`
	BlockNumber uint64        `json:"blockNumber"`
	Amount      string        `json:"amount"`
}

type SortableERC721Token struct {
	types.ERC721Token

	//Allows the token to be sortable by splitting it into component parts
	First  uint64 `json:"first"`
	Second uint64 `json:"second"`
	Third  uint64 `json:"third"`
	Fourth uint64 `json:"fourth"`
	Fifth  uint64 `json:"fifth"`
}

//

type ContractQueryResult struct {
	Source Contract `json:"_source"`
}

type TemplateQueryResult struct {
	Source Template `json:"_source"`
}

type TransactionQueryResult struct {
	Source *types.Transaction `json:"_source"`
}

type BlockQueryResult struct {
	Source *types.Block `json:"_source"`
}

type TokenHolderQueryResult struct {
	Source TokenHolder `json:"_source"`
}

type StateQueryResult struct {
	Source State `json:"_source"`
}

type StorageQueryResult struct {
	Source Storage `json:"_source"`
}

type LastPersistedResult struct {
	Source struct {
		LastPersisted uint64 `json:"lastPersisted"`
	} `json:"_source"`
}

type SearchQueryResult struct {
	Hits struct {
		Hits []IndividualResult `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		Results map[string]interface{} `json:"result_buckets"`
	} `json:"aggregations"`
}

type CountQueryResult struct {
	Count uint64 `json:"count"`
}

type IndividualResult struct {
	Id     string                 `json:"_id"`
	Source map[string]interface{} `json:"_source"`
}

type ERC721HolderAggregateResult struct {
	AfterKey struct {
		Holder string
	} `mapstructure:"after_key"`
	Buckets []struct {
		Key struct {
			Holder string
		}
	}
}
