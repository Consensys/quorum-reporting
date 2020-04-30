# Storage reqs

## What is needed

The reporting tool is currently based around reporting data for given __addresses__.

For a particular given address, it has the following information:

- ABI
- Txns sent to it
- Txn that created it
- Txns that involve it in some way ("internal transactions")
- The state root, at a given block height
- The storage data, at a given block height


## What we have extra

So far, we have pulled in more than we need whilst the requirements were not clear and can be removed:

- Whole state dump for each block, this can end up being tens or even hundreds of GB large and most of it may be unneeded.
- Block data itself, there is nothing in the block itself that relevant to an address (we can pull out txns separately)


## Data structure

Considering that we are using ElasticSearch, it makes sense to structure our data as a JSON document.

```
Contract {
	Address
	ABI
	ContractCreationTransaction
	LastFiltered
}
```

Storage has been split into "storage" and "state".
This is so that we only need to store the storage mapping once, which can 
then be reference by the state multiple times in an effort to keep as little
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

```
Events {
    Address
    BlockHash
    BlockNumber
    Data
    LogIndex
    Topics
    TransactionHash
    TransactionIndex
}
```

```
Transaction {
    BlockHash
    BlockNumber
    From
    Gas
    GasPrice
    Hash
    Input
    Nonce
    To
    TransactionIndex
    Value
    IsPrivate
    ContractAddress
    CumulativeGasUsed
    GasUsed
    Status
    Root
    InternalCalls
}
```

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
