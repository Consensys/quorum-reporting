# RPC API Specs

### Contract

Contract APIs register/ deregister contracts to be reported. Complex queries can be run for the registered contract list.

#### reporting_addAddress

(Implemented)

#### reporting_deleteAddress

(Implemented)

#### reporting_getAddresses

(Implemented)

#### reporting_addABI

(Implemented)

#### reporting_getABI

(Implemented)

#### reporting_getLastFiltered

(Implemented) `reporting_getLastFiltered` gets the last block number before which storage & txs & events of a contract 
is filtered and stored.

### Block

Block APIs returns basic block information.

#### reporting_getBlock

(Implemented)

#### reporting_getLastPersistedBlockNumber

(Implemented) `reporting_getLastPersistedBlockNumber` gets the last block number before which all blocks are available 
and properly indexed.

### Storage

Storage APIs can query account storage for a given contract at any block

#### reporting_getStorage

(Implemented)

### Transaction

Transaction APIs query 

#### reporting_getTransaction

(Implemented)

#### reporting_getContractCreationTransaction

(Implemented)

#### reporting_getAllTransactionsToAddress

(Todo) `reporting_getAllTransactionsToAddress` returns a list of tx hash and total number matching the search criteria 
provided.

Sample Response:
```$json
{
    transactions: [common.Hash...],
    total: uint64,
    options: {
        beginBlockNumber: 0,
        endBlockNumber: "latest",
        beginTimestamp: 0,
        endTimestamp: 9999999999,
        pageSize: 10,
        pageNumber: 0,
    }
}
```

Default Criteria:
```$json
{
    beginBlockNumber: 0,
    endBlockNumber: "latest",
    beginTimestamp: 0,
    endTimestamp: 9999999999,
    pageSize: 10,
    pageNumber: 0,
}
```

#### reporting_getAllTransactionsInternalToAddress

(Todo) `reporting_getAllTransactionsInternalToAddress` returns a list of tx hash and total number matching the search 
criteria provided.

Sample Response:
```$json
{
    transactions: [common.Hash...],
    total: uint64,
    options: {
        beginBlockNumber: 0,
        endBlockNumber: "latest",
        beginTimestamp: 0,
        endTimestamp: 9999999999,
        pageSize: 10,
        pageNumber: 0,
    }
}
```

Default Criteria:
```$json
{
    beginBlockNumber: 0,
    endBlockNumber: "latest",
    beginTimestamp: 0,
    endTimestamp: 9999999999,
    pageSize: 10,
    pageNumber: 0,
}
```

### Event

#### reporting_getAllEventsFromAddress

(Todo) `reporting_getAllEventsFromAddress` returns a list of event objs and total number of events matching the search 
criteria provided.

Sample Response:
```$json
{
    events: [eventObj...],
    total: uint64,
    options: {
        beginBlockNumber: 0,
        endBlockNumber: "latest",
        beginTimestamp: 0,
        endTimestamp: 9999999999,
        pageSize: 10,
        pageNumber: 0,
    }
}
```

Default Criteria:
```$json
{
    beginBlockNumber: 0,
    endBlockNumber: "latest",
    beginTimestamp: 0,
    endTimestamp: 9999999999,
    pageSize: 10,
    pageNumber: 0,
}
```