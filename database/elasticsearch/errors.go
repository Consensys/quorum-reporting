package elasticsearch

import "errors"

var (
	ErrAddressNotFound = errors.New("address not found")
	ErrTooManyResults  = errors.New("too many results")
)
