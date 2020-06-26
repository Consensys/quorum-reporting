package types

type SyncError struct {
	endBlockNumber uint64
	errorMessage   string
}

func NewSyncError(message string, endBlockNumber uint64) *SyncError {
	return &SyncError{
		endBlockNumber: endBlockNumber,
		errorMessage:   message,
	}
}

func (se *SyncError) EndBlockNumber() uint64 {
	return se.endBlockNumber
}

func (se *SyncError) Error() string {
	return se.errorMessage
}
