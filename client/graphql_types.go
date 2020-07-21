package client

type CurrentBlockResult struct {
	Block Block
}

type TransactionResult struct {
	Transaction Transaction
}

type Block struct {
	Number    string
	Hash      string
	Timestamp string
}

type Transaction struct {
	Hash              string
	Status            string
	Index             uint64
	Nonce             string
	From              Address
	To                Address
	Value             string
	GasPrice          string
	Gas               string
	GasUsed           string
	CumulativeGasUsed string
	CreatedContract   Address
	InputData         string
	PrivateInputData  string
	IsPrivate         bool
	Logs              []Event
}

type Event struct {
	Index   uint64
	Account Address
	Topics  []string
	Data    string
}

type Address struct {
	Address string
}
