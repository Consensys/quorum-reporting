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

(Deprecated) `reporting_getAllTransactionsToAddress` currently returns all transactions sending to the address. It can 
be a huge list. Switching to a set of new APIs with query condition.

#### reporting_getAllTransactionsToAddressByNumber

(Todo)

#### reporting_getAllTransactionsToAddressByBlock

(Todo)

#### reporting_getAllTransactionsToAddressByTimestamp

(Todo)

#### reporting_getAllTransactionsInternalToAddress

(Deprecated) `reporting_getAllTransactionsInternalToAddress` currently returns all transactions internally calling to 
the address. Switching to a set of new APIs with query condition.

#### reporting_getAllTransactionsInternalToAddressByNumber

(Todo)

#### reporting_getAllTransactionsInternalToAddressByBlock

(Todo)

#### reporting_getAllTransactionsInternalToAddressByTimestamp

(Todo)

### Event

#### reporting_getAllEventsByAddress

(Deprecated) `reporting_getAllEventsByAddress` currently returns all events emitted from the address. Switching to a 
set of new APIs with query condition.

#### reporting_getAllEventsFromAddressByNumber

(Todo)

#### reporting_getAllEventsFromAddressByBlock

(Todo)

#### reporting_getAllEventsFromAddressByTimestamp

(Todo)
