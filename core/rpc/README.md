# RPC API Specs

## Contract

Contract APIs register/ deregister contracts to be reported. Complex queries can be run for the registered contract list.

#### reporting_addAddress

Adds a new address to start indexing and can be querying for various reports. Optionally takes a block number from 
which to start indexing.

Input:
```json
{
	"address": "<address>",
	"blockNumber": <integer>
}
```

Output:
None

#### reporting_deleteAddress

Deletes an address from being indexed or queried.

Input:
```json
"<address>"
```

Output:
None

#### reporting_getAddresses

Returns a list of all the addresses the reporting engine is indexing.

Input:
None

Output:
```json
["<address>", ...]
```

#### reporting_getContractTemplate

Returns the name of the template that is currently assigned to the given contract

Input:
```json
"<address>"
```

Output:
```json
"<template name>"
```

#### reporting_addABI

(Deprecated, use `reporting_addTemplate` and `reporting_assignTemplate`)

Assigns a contract ABI to a contract, allowing parsing of function call and event parameters.

Input:
```json
{
    "address": "<address>",
    "data": "<escaped contract ABI json>"
}
```

Output:
None

#### reporting_getABI

Returns the attached contract ABI for the given contract

Input:
```json
"<address>"
```

Output:
```json
"<Contract ABI as escaped JSON>"
```

#### reporting_addStorageABI

(Deprecated. Use `reporting_addTemplate` and `reporting_assignTemplate`)

Assigns a Storage Layout to a contract, allowing parsing of contract storage into variables.

Input:
```json
{
    "address": "<address>",
    "data": "<escaped storage layout json>"
}
```

Output:
None

#### reporting_getStorageABI

Returns the attached Storage Layout for the given contract

Input:
```json
"<address>"
```

Output:
```json
"<Storage Layout as escaped JSON>"
```

#### reporting_addTemplate

Adds a new template that can be assigned to contracts

Input:
```json
{
    "name": "<template identifier>",
    "abi": "<escaped contract ABI JSON>",
    "storageLayout": "<escaped Storage Layout JSON>"
}
```

Output:
None

#### reporting_assignTemplate

Assigns a previously added template to the given contract, replacing any existing assignment that contract had.

Input:
```json
{
    "address": "<address>",
    "data": "<template name>"
}
```

Output:
None

#### reporting_getTemplates

Returns a list of all template names that have been added to the reporting engine

Input:
None

Output:
```json
[
    "<template name>",
    ...
]
```

#### reporting_getTemplateDetails

Returns the details of a given template, which includes the template Contract ABI and the Storage Layout.

Input:
```json
"<template name>"
```

Output:
```json
{
    "name": "<template identifier>",
    "abi": "<escaped contract ABI JSON>",
    "storageLayout": "<escaped Storage Layout JSON>"
}
```

#### reporting_getLastFiltered

(Implemented) `reporting_getLastFiltered` gets the last block number before which storage & txs & events of a contract 
is filtered and stored.

## Block

Block APIs returns basic block information.

#### reporting_getBlock

Fetches the full block data

Input:
```json
100
```

Output:
```json
{
	"hash": "<0x-prefixed hash>",
	"parentHash": "<0x-prefixed hash>",
	"stateRoot": "<0x-prefixed hash>",
	"txRoot": "<0x-prefixed hash>",
	"receiptRoot": "<0x-prefixed hash>",
	"number": <integer>,
	"gasLimit": <integer>,
	"gasUsed": <integer>,
	"timestamp": <integer>,
	"extraData": "<0x-prefixed string",
	"transactions": ["<0x-prefixed hash>"]
}
```

#### reporting_getLastPersistedBlockNumber

Fetches the last block number before which all blocks/transactions are available.

Input:
None

Output:
```json
100
```

## Storage

Storage APIs can query account storage for a given contract at any block

#### reporting_getStorage

Retrieves the full *raw* storage for a contract at a particular block height. This means there is no parsing of the 
data. If no block is given, then the latest block the contract has been indexed at is used. The values of each storage 
slot are truncated to remove any leading 0's, providing there remain an even number of characters (making it valid hex).

Input:
```json
{
    "address": "<address>",
    "block": <integer>
}
```

Output:
```json
{
  "<storage slot 0 hash>": "<storage slot 0 value>",
  "<storage slot 1 hash>": "<storage slot 1 value>",
  ...
}
```

e.g.
```json
{
    "0x00000000000000000000000000000000": "10",
    "0x00000000000000000000000000000001": "12345678901234567890123456789012"
}
```

#### reporting_getStorageHistory

Parses the storage of a contract according to its attached storage layout. It will return a map of variables and their 
values that exist in the contract, except for mappings. This is intended to see how the storage changes over time, 
and so takes a start and end block range. These can be kept the same if a single block is required.

Input:
```json
{
	"address": "<address>",
	"startBlockNumber": <integer>,
	"endBlockNumber": <integer>
}
```

Output:
```json
{
	"address": "<address>",
	"historicState": [
        {
            "blockNumber": <integer>,
            "historicStorage": [
                {
                    "name": "<string>",
                    "type": "<string, solidity variable type>",
                    "value": <variable based on variable type>
                },
                ...
            ]
        },
        ...
    ]
}
```

#### reporting.GetStorageHistoryCount

Fetches the number of storage entries for the given block range and account. It will subdivide the total entries
into block ranges with 1000 results in, so that effective pagination can be used.

Input:
```json
{
	"address": "<address>",
    "options": {
      	"startBlockNumber": <integer>,
      	"endBlockNumber": <integer>
    }
}
```
Note: `startBlockNumber` must be greater than or equal to 0. `endBlockNumber` can be `-1` to indicate the latest 
indexed block for the given address.

Output:
```json
{
    "ranges": [
        {
            "start": 1205,
            "end": 2205,
            "resultCount": 1000
        },
        {
            "start": 205,
            "end": 1204,
            "resultCount": 1000
        },
        {
            "start": 0,
            "end": 204,
            "resultCount": 199
        }
    ]
}
```
Note: the output works backwards, giving the most recent blocks first.

## Transaction

Transaction APIs query 

#### reporting_getTransaction

Fetches transaction data, including events and internal calls & parsed event/function call data

Input:
```json
"<0x-prefixed hash>"
```

Output:
```json
{
	"txSig": "<parsed function name and parameters>",
	"func4Bytes": "<0x-prefixed string", //function 4bytes signature
	"parsedData": {
	  "function parameter 1 name": "function parameter 1 value",
	  "function parameter 2 name": "function parameter 2 value",
      ...
	},
	"parsedEvents": {
	  	"eventSig": "<0x-prefixed hash",
      	"parsedData": {
          "event parameter 1 name": "function parameter 1 value",
          "event parameter 2 name": "function parameter 2 value",
          ...
        }
      	"rawEvent": {
      	    "index": <integer>,
        	"address": "<0x-prefixed address>",
        	"topics": ["<0x-prefixed hash>", ...],
        	"data": "<0x-prefixed string>",
        	"blockNumber": <integer>,
        	"blockHash": "<0x-prefixed hash>",
        	"transactionHash": "<0x-prefixed hash>",
        	"transactionIndex": <integer>,
        	"timestamp": <integer>
      	}
	},
	"rawTransaction": {
	    "hash": "<0x-prefixed hash>",
      	"status": <bool>,
      	"blockNumber": <integer>,
      	"blockHash": "<0x-prefixed hash>",
      	"index": <integer>,
      	"nonce": <integer>,
      	"from": "<0x-prefixed address>",
      	"to": "<0x-prefixed address>",
      	"value": <integer>,
      	"gas": <integer>
      	"gasPrice": <integer>,
      	"gasUsed": <integer>,
      	"cumulativeGasUsed": <integer>,
      	"createdContract": "<0x-prefixed address>",
      	"data": "<0x-prefixed string>",
      	"privateData": "<0x-prefixed string>",
      	"isPrivate": <bool>,
      	"timestamp": <integer>,
      	"events": [
            {
                "index": <integer>,
                "address": "<0x-prefixed address>",
                "topics": ["<0x-prefixed hash>", ...],
                "data": "<0x-prefixed string>",
                "blockNumber": <integer>,
                "blockHash": "<0x-prefixed hash>",
                "transactionHash": "<0x-prefixed hash>",
                "transactionIndex": <integer>,
                "timestamp": <integer>
            },
            ...
        ],
      	"internalCalls": [
            {
                "from": "<0x-prefixed address>",
                "to": "<0x-prefixed address>",
                "value": <integer>,
                "gas": <integer>
                "gasUsed": <integer>,
              	"input": "<0x-prefixed string>",
              	"output": "<0x-prefixed string>",
              	"type": "<opcode name>"
            }, 
            ...
        ]
	}
```

#### reporting_getContractCreationTransaction

Fetches the hash of the transaction that this requested transaction was deployed at.
This can include external deployment, internal deployments from other contracts, or 
via contract extension.

Input:
```json
"<0x-prefixed address>"
```

Output:
```json
"<0x-prefixed hash>"
```

#### reporting_getAllTransactionsToAddress

Returns a list of transaction hashes and total number matching the search options provided.

Input:
```json
{
    "address": "<address>",
    "options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,
        "beginTimestamp": <integer>,
        "endTimestamp": <integer>,
        "pageSize": <integer>,
        "pageNumber": <integer>
    }
}
```

Output:
```$json
{
    "transactions": ["<hash>", ...],
    "total": <integer>,
    "options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,
        "beginTimestamp": <integer>,
        "endTimestamp": <integer>,
        "pageSize": <integer>,
        "pageNumber": <integer>
    }
}
```

#### reporting_getAllTransactionsInternalToAddress

Returns a list of transaction hashes where the contract was called by another contract, 
along with the total number matching records with the search options provided.

Input:
```json
{
    "address": "<address>",
    "options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,
        "beginTimestamp": <integer>,
        "endTimestamp": <integer>,
        "pageSize": <integer>,
        "pageNumber": <integer>
    }
}
```

Output:
```$json
{
    "transactions": ["<hash>", ...],
    "total": <integer>,
    "options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,
        "beginTimestamp": <integer>,
        "endTimestamp": <integer>,
        "pageSize": <integer>,
        "pageNumber": <integer>
    }
}
```

## Event

#### reporting_getAllEventsFromAddress

Returns a list of events for a given contract, along with the total number of events matching the search options 
provided. The events are also parsed for their parameter values if an appropriate ABI is attached to the contract.

Input:
```json
{
    "address": "<address>",
    "options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,
        "beginTimestamp": <integer>,
        "endTimestamp": <integer>,
        "pageSize": <integer>,
        "pageNumber": <integer>
    }
}
```

Output:
```$json
{
    "events": [
        {
            "eventSig": "<0x-prefixed hash",
            "parsedData": {
              "event parameter 1 name": "function parameter 1 value",
              "event parameter 2 name": "function parameter 2 value",
              ...
            }
            "rawEvent": {
                "index": <integer>,
                "address": "<0x-prefixed address>",
                "topics": ["<0x-prefixed hash>", ...],
                "data": "<0x-prefixed string>",
                "blockNumber": <integer>,
                "blockHash": "<0x-prefixed hash>",
                "transactionHash": "<0x-prefixed hash>",
                "transactionIndex": <integer>,
                "timestamp": <integer>
            }
        },
        ...
    ],
    "total": <integer>,
    "options": {
        "beginBlockNumber": <integer>,
        "endBlockNumber": <integer>,
        "beginTimestamp": <integer>,
        "endTimestamp": <integer>,
        "pageSize": <integer>,
        "pageNumber": <integer>
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
