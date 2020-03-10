package types

// TODO: This struct is used to signal block received.
//  It should be captured by transactionFilter to pull necessary transactions and storageFilter to pull necessary
//  storage key-values.
type BlockReceived struct {
	Block
}

// TODO: This struct is used to signal transaction received.
//  It should be captured by eventFilter to pull necessary events.
type TransactionRecieved struct {
}
