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

Note: the storage here can still get extremely large. This may be problematic, see https://www.elastic.co/guide/en/elasticsearch/reference/current/general-recommendations.html#maximum-document-size
```
States {
    Address
    BlockNumber
    StorageRoot
    Storage : {
        Key: Value
    }
}
```

```
Events {
    ID (BlockNum + LogIndex)
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
    Receipt: {
      ContractAddress
      CumulativeGasUsed
      GasUsed
      Logs: [{
          Address
          Data,
          LogIndex
          Topics
      }],
      LogsBloom
      Status
      Root
    }
    
    ContractsCalled
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
