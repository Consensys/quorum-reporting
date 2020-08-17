# RPC API Specs

## Contract

Contract APIs register/ deregister contracts to be reported. Complex queries can be run for the registered contract list.

#### reporting_addAddress

(Implemented)

#### reporting_deleteAddress

(Implemented)

#### reporting_getAddresses

(Implemented)

#### reporting_getContractTemplate

(Implemented)

#### reporting_addABI

(Deprecated. Use `reporting_addTemplate` and `reporting_assignTemplate`)

#### reporting_getABI

(Implemented)

#### reporting_addStorageABI

(Deprecated. Use `reporting_addTemplate` and `reporting_assignTemplate`)

#### reporting_getStorageABI

(Implemented)

#### reporting_addTemplate

(Implemented)

#### reporting_assignTemplate

(Implemented)

#### reporting_getTemplates

(Implemented)

#### reporting_getTemplateDetails

(Implemented)

#### reporting_getLastFiltered

(Implemented) `reporting_getLastFiltered` gets the last block number before which storage & txs & events of a contract 
is filtered and stored.

## Block

Block APIs returns basic block information.

#### reporting_getBlock

(Implemented)

#### reporting_getLastPersistedBlockNumber

(Implemented) `reporting_getLastPersistedBlockNumber` gets the last block number before which all blocks are available 
and properly indexed.

## Storage

Storage APIs can query account storage for a given contract at any block

#### reporting_getStorage

(Implemented)

#### reporting_getStorageHistory

(Todo) `reporting_getStorageHistory` provides extended feature on top of simply getting raw storage. It can search by 
block range, and provides a list of historical state formatted by the given template.

## Transaction

Transaction APIs query 

#### reporting_getTransaction

(Implemented)

#### reporting_getContractCreationTransaction

(Implemented)

#### reporting_getAllTransactionsToAddress

(Implemented) `reporting_getAllTransactionsToAddress` returns a list of tx hash and total number matching the search options 
provided.

Sample Response:
```$json
{
    transactions: [types.Hash...],
    total: uint64,
    options: {
        beginBlockNumber, endBlockNumber,
        beginTimestamp, endTimestamp,
        pageSize, pageNumber,
    }
}
```

#### reporting_getAllTransactionsInternalToAddress

(Implemented) `reporting_getAllTransactionsInternalToAddress` returns a list of tx hash and total number matching the search 
options provided.

Sample Response:
```$json
{
    transactions: [types.Hash...],
    total: uint64,
    options: {
        beginBlockNumber, endBlockNumber,
        beginTimestamp, endTimestamp,
        pageSize, pageNumber,
    }
}
```

## Event

#### reporting_getAllEventsFromAddress

(Implemented) `reporting_getAllEventsFromAddress` returns a list of event objs and total number of events matching the search 
options provided.

Sample Response:
```$json
{
    events: [eventObj...],
    total: uint64,
    options: {
        beginBlockNumber, endBlockNumber,
        beginTimestamp, endTimestamp,
        pageSize, pageNumber,
    }
}
```

## Default Query Options
```$json
{
    beginBlockNumber: 0,
    endBlockNumber: -1("latest"),
    beginTimestamp: 0,
    endTimestamp: -1("latest"),
    pageSize: 10,
    pageNumber: 0,
}
```

## Token APIs

#### token.getERC20TokenBalance

Fetches the balances for a particular ERC20 holder for the given block range.
It will only list blocks where a balance change has taken place, so keys may not be consecutive.
It will also list a balance prior to the starting block, if the balance did not change at the starting block;
this value is replicated for the starting block as well.

Input:
```$json
{
	"contract": "0x<address>"
	"holder": "0x<address>"
	"options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,

        "pageSize": <integer>,
        "pageNumber": <integer>
    }
```

Output:
```$json
{
	"5": 100,
    "6": 200,
    "10": 1000,
    ...
}
```

#### token.getERC20TokenHoldersAtBlock

Returns all the holders of a token at a particular block.
The maximum amount of results that can be returned is 1000 per request.
To continue retrieving accounts, specify the last account retrieved as 
the `after` parameter in the `options` object; continue until all accounts have been retrieved.

Input:
```$json
{
	"contract": "0x<address>"
	"block": <integer>,
	"options": {
        "after": "0x<address>"
        "pageSize": <integer>
    }
```

Output:
```$json
[
    "0x<address>",
    "0x<address>",
    "0x<address>"
]
```

#### token.getHolderForERC721TokenAtBlock

Fetches the address of the given token holder at a given block height.

Input:
```$json
{
	"contract": "0x<address>"
	"tokenId": <integer>,
    "block": <integer>
```

Output:
```$json
"0x<address>"
```

#### token.eRC721TokensForAccountAtBlock

Fetches all ERC721 tokens for an account at a given block. Since the total number of held tokens may exceed 
the maximum request size (using `pageNumber` and `pageSize`), a start token ID may be specified using `after` 
(exclusive).

A list of all tokens are returned, detailing their ID number, when they were first held from and
(optionally) when they were held until.

Input:
```$json
{
	"contract": "0x<address>"
	"holder": "0x<address>"
	"block": <integer>,
	"options": {
        "after": "<integer>",
        "pageNumber": <integer>,
        "pageSize": <integer>
    }
```

Output:
```$json
[
    {
        	"contract": "0x<address>",
        	"holder": "0x<address>",
        	"token": "<integer>"
        	"heldFrom": <integer>,
        	"heldUntil": <integer>
    },
    ...
]
```

#### token.allERC721TokensAtBlock

Fetches all ERC721 tokens at a given block. Since the total number of held tokens may exceed 
the maximum request size (using `pageNumber` and `pageSize`), a start token ID may be specified using
`after` (exclusive).

A list of all tokens are returned, detailing their ID number, who holds the token, when they were first held from and
(optionally) when they were held until.

Input:
```$json
{
	"contract": "0x<address>"
	"holder": "0x<address>"
	"block": <integer>,
	"options": {
        "after": "<integer>",
        "pageNumber": <integer>,
        "pageSize": <integer>
    }
```

Output:
```$json
[
    {
        	"contract": "0x<address>",
        	"holder": "0x<address>",
        	"token": "<integer>"
        	"heldFrom": <integer>,
        	"heldUntil": <integer>
    },
    ...
]
```



#### token.allERC721HoldersAtBlock

Returns all the holders of a token at a particular block.
The maximum amount of results that can be returned is 1000 per request.
To continue retrieving accounts, specify the last account retrieved as 
the `after` parameter in the `options` object; continue until all accounts have been retrieved.

Input:
```$json
{
	"contract": "0x<address>"
	"block": <integer>,
	"options": {
        "after": "0x<address>"
        "pageSize": <integer>
    }
```

Output:
```$json
[
    "0x<address>",
    "0x<address>",
    "0x<address>"
]
```
