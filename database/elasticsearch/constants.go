package elasticsearch

import "errors"

// indices
const (
	MetaIndex        = "meta"
	ContractIndex    = "contract"
	BlockIndex       = "block"
	StorageIndex     = "storage"
	StateIndex       = "state"
	TransactionIndex = "transaction"
	EventIndex       = "event"
)

// errors
var (
	ErrCouldNotResolveResp = errors.New("could not resolve response body")
	ErrIndexNotFound       = errors.New("index not found")
)
