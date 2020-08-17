package elasticsearch

import "errors"

// indices
const (
	MetaIndex        = "meta"
	ContractIndex    = "contract"
	TemplateIndex    = "template"
	BlockIndex       = "block"
	StorageIndex     = "storage"
	StateIndex       = "state"
	TransactionIndex = "transaction"
	EventIndex       = "event"
	ERC20TokenIndex  = "erc20token"
	ERC721TokenIndex = "erc721token"
)

var (
	AllIndexes = []string{MetaIndex, ContractIndex, TemplateIndex, BlockIndex, StorageIndex, StateIndex, TransactionIndex, EventIndex, ERC20TokenIndex, ERC721TokenIndex}
	// errors
	ErrCouldNotResolveResp     = errors.New("could not resolve response body")
	ErrIndexNotFound           = errors.New("index not found")
	ErrPaginationLimitExceeded = errors.New("pagination limit exceeded")
)
