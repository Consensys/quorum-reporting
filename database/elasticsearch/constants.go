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
)

var (
	AllIndexes = []string{MetaIndex, ContractIndex, TemplateIndex, BlockIndex, StorageIndex, StateIndex, TransactionIndex, EventIndex}
	// errors
	ErrCouldNotResolveResp     = errors.New("could not resolve response body")
	ErrIndexNotFound           = errors.New("index not found")
	ErrPaginationLimitExceeded = errors.New("pagination limit exceeded")
)
