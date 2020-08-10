# Storage reqs

## What is needed

The reporting tool is based around reporting data for given __addresses__.

For a particular given address, it allows querying of the following information:

- ABI (optional)
- Storage Layout (optional)
- Transaction that created it
- Transactions sent to it
- Transactions that call the contract from another contract ("internal transactions")
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
	TemplateName
	ContractCreationTransaction
	LastFiltered
}
```

#### Contract Template
```
Template {
	TemplateName
	ABI
	StorageABI (storage layout)
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

#### ERC20 Tokens Index

The layout for ERC20 tokens make its straight-forward to be updated and searched to.
For each ERC20 contract, all holders of a balance are listed along with the block number that balance is valid from.
New entries are made when a transfer has taken place, although the balance change doesn't need to occur.

Note: `heldUntil` is an updatable field, and is updated with the block number the balance is valid till. This makes
it easier to query balances without needing to refer to other entries to find out when it was valid until.

```
ERC20TokenHolder {
    Contract
    Holder
    BlockNumber
    Amount
    HeldUntil
}
```

#### ERC721 Tokens Index

ERC721 tokens have a more complex layout. The challenge is to have a structure that can scale both with
the number of holders and with the number of tokens, since there is no combining of tokens like there is for ERC20.

```
    Contract  String
    Holder    String
    Token     String
    HeldFrom  Long
    HeldUntil Long

	First  Long
	Second Long
	Third  Long
	Fourth Long
	Fifth  Long
```

The first half of the structure are the main pieces of information that are used in the application and given 
back to the user. Importantly, the `HeldUntil` field is empty until the token has been transferred to someone else.
This allows us to place a bound on the queries made without needed to make a new entry every block to reaffirm the
status of the token and increasing the database size unnecessarily.

The second half of the structure is ElasticSearch specific. At the worst case, the only field that can differentiate 
one record from the next is the `Token` field (it may be 2 tokens were gained and lost at the same point, for the same
holder). To enable pagination of results, we need a consistent ordering of results, so sorting on results is a must.
Sorting on `string` fields in ElasticSearch increases memory use as the number of records increases, and may become 
prohibitive over time. A `long` in ElasticSearch can have a maximum value of `2^63-1`, but a token ID can be up to 
`2^256-1`. Thus the extra fields are the token ID split into multiple smaller chunks, each fitting inside `long`. The
following holds: `string(tokenId) === string(first) + string(second) + string(third) + string(fourth) + string(fifth)`.
This allows sorting within an acceptable resource limit. Note: each field stores 17 digits.