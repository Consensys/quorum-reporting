# Storage reqs

## What is needed

The reporting tool is currently based around reporting data for given __addresses__.

For a particular given address, it allows query of the following information:

- ABI (optional)
- Storage Layout (optional)
- Txn that created it
- Txns sent to it
- Txns that involve it in some way ("internal transactions")
- Events emitted by it
- The storage data at a given block height

## What we have extra

So far, we have pulled in more than the minimal required data. We keep all blocks and transactions data no matter it is 
related to any registered address or not in the database. This allows filtering service to adapt to more complicated 
reporting requirements in the future.

## Data structure

Considering that we are using ElasticSearch, it makes sense to structure our data as a JSON document.

#### Contract Index
```
Contract {
	Address
	ABI
    StorageABI (storage layout)
	ContractCreationTransaction
	LastFiltered
}
```

#### State Index & Storage Index
Storage has been split into "Storage" and "State". This is so that we only need to store the storage mapping with same 
root once in "Storage", which can then be reference by the "State" multiple times in an effort to keep as little 
duplicate data as possible.
```
State {
    Address
    BlockNumber
    StorageRoot
}
```
```
Storage {
    StorageRoot
    Storage : {
        Key: Value
    }
}
```

#### Event Index
```
Event {
    Address
    BlockHash
    BlockNumber
    Data
    LogIndex
    Topics
    TransactionHash
    TransactionIndex
    Timestamp
}
```

#### Transaction Index
```
Transaction {
	Hash
	Status
	BlockNumber
	BlockHash
	Index
	Nonce
	Sender
	Recipient
	Value
	Gas
	GasPrice
	GasUsed
	CumulativeGasUsed
	CreatedContract
	Data
	PrivateData
	IsPrivate
	Events
	InternalCalls
	Timestamp
}
```

#### Block Index
```
Block {
    Hash
    ParentHash
    StateRoot
    TxRoot
    ReceiptRoot
    Number
    GasLimit
    GasUsed
    Timestamp
    ExtraData
    Transactions
}
```
