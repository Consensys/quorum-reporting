package elasticsearch

import "errors"

// indices
const (
	MetaIndex        = "meta"
	ContractIndex    = "contract"
	BlockIndex       = "block"
	StorageIndex     = "storage"
	TransactionIndex = "transaction"
	EventIndex       = "event"
)

// errors
var (
	ErrAddressNotFound     = errors.New("address not found")
	ErrTooManyResults      = errors.New("too many results")
	ErrCouldNotResolveResp = errors.New("could not resolve response body")
	ErrIndexNotFound       = errors.New("index not found")
)
