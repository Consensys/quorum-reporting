# Quorum Reporting Feature set

## RPC APIs for fetching data

All the data in the Reporting Engine can be viewed through calls to the RPC API.
A full run down on the APIs can be viewed [here](core/rpc/README.md).

## Block & transaction fetching/filtering

All blocks and transactions are imported into the Reporting Engine, which includes a trace of all the internal calls
made.
This used to allow search filtering on transactions made to particular contracts, as well as view all internal message 
calls made to contracts as well.

## User-defined contract filtering for state, events, creation transaction

Contracts can be added to fetch their state at each block, events that are relevant to them, as well as find
the transaction hash in which the contract was created.
Note: events can be seen for all transactions *when searching by transaction*, but can only be searched for by contract 
if that contract has been added to the filter list.

To add contracts to the filter list, see below

## Rules-based contract monitoring

Rules can put in place that will monitor all newly created contracts and add them automatically to the contract filter 
list. This includes checking via whether an ABI matches the contracts bytecode, or using an EIP165 identifier to call 
the contract explicitly.

## ERC20 & ERC721 token tracking

Support for filtering on ERC20 and ERC721 contracts and recording balance changes that occur, and being able to query
on absolute balances at any given block height.

## Event/contract storage/contract call variable parsing (requires ABI & storage map)

With an attached ABI & Solidity storage mapping, event, function & storage variable names and values can be parsed 
and presented back to the user.

# Walkthroughs

## Adding a new contract to filter on

Adding contracts to the filtering list means that you are interested in seeing more details about that contract.
This includes being able to parse its events/storage/function calls, searching for events and transactions that are
from/for that address, and seeing any ERC token transfers if relevant.

There are two ways to add a new address to filter on:

1. Add it to the configuration. This will make sure the contract is in the filter list **at every startup**, as well as
being able to assign it a template (more on that later).
Here is a sample: 
```toml
addresses = [
    { address = "0x8a5e2a6343108babed07899510fb42297938d41f", templateName = "SimpleStorage" }
]
```

2. Using the RPC API:
Addresses can also be added at runtime via the RPC API. There is also some more granular control that can be achieved
this way:
```bash
curl -H 'Content-Type: application/json' -X POST http://localhost:4000 --data '{"jsonrpc":"2.0", "method":"reporting.AddAddress", "params":[{"address": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34", "blockNumber": 500}], "id":67}'
```

This example adds the address `0x1932c48b2bf8102ba33b4a6b545c32236e342f34` to the filter list, and will start indexing 
data from block number `500`, if you are not interested in data before that point. The block number can be omitted and
will default to `0`.

## Templates

Templates are a way of reusing an ABI or storage mapping across several contracts. The template can be created/updated 
in the startup configuration, and also assigned to contracts there too. Alternatively, they can be created/updated via 
the RPC API, as well as being assigned to contracts that way.

Here is a sample empty template:
```toml
templates = [
    { templateName = "SimpleStorage", abi = '[]', storageLayout = '{}' }
]
```

Each template has a name, which is how it is referred to when assigning it to contracts.

The `abi` is the standard Ethereum JSON ABI, which details all functions (including constructor) and events.
The `storageLayout` describes the layout of a contracts variables in storage. This only works for Solidity contracts,
and currently does not support mappings, due to how Solidity determines where to store them - although they can be 
included in the `storageLayout`, they will be ignored.
The storage layout is one of the outputs of compiling the contract using `solc`, from version 0.6.7 - although you may
be able to use v0.6.7 to compile the storage layout and apply it to a contract compiled against an earlier version, as 
storage has not changed dramatically.

The command to run to get the storage layout is `solc <path to sol file> --combined-json storage-layout --pretty-json`


## Rules-based monitoring

One can define rules that will allow contracts to be automatically added to the filter list, meaning all contracts of a 
particular type can be monitored without user intervention. If all of the sub-rules (listed below) match, then the rule 
is considered matched, and the contract is added to the filter list from block 0.

Here are some sample rules:
```toml
rules = [
    { scope = "external", templateName = "ERC20", eip165 = "36372b07"},
    { scope = "all", templateName = "ERC721", eip165 = "80ac58cd", deployer = "0x8a5e2a6343108babed07899510fb42297938d41f"}
]
```

The `scope` of a rule determines whether the deployment was directly from a transaction (i.e. the `to` field was empty),
or if it was deployed from another contract internally.

The `template` name states what template should be assigned on a successful match.

The `eip165` field contains a 4-byte interface identifier, according to 
[EIP165](https://eips.ethereum.org/EIPS/eip-165). 
If the match is not successful, then it will fallback to checking the
contracts bytecode, to see if it contains the signatures of all the methods and events of the template in order to 
produce a match. Note that this can flag false positives, if the contract is not of the template type, but deploys a 
contract that is of that type, since if will need the sub-contracts bytecode embedded within its own.
This field is optional, and if omitted, will default to template-based matching.

The `deployer` field states which address must have done the deployment. This is useful, for example, if you are only 
interested in your deployed contracts. This is an optional field.

## ERC20 & ERC721 token tracking

Contracts that are filtered on, and have an ABI that matches the ERC20 or ERC721 are also queried for account balances 
when transfer events happen. From this, the RPC API can be queried for a range of information, including specific 
account balances, seeing which accounts have a balance and more.

Please note the only extra limitation that is required by the contract (on top of making sure the token spec is 
followed) is to make sure if any balance is assigned during an ERC721 constructor, then a transfer event still 
takes place - this is required by default for ERC20 tokens.

## Event, storage and function parsing

If the assigned template contains an ABI, then the contracts events and function calls can be parsed to show their 
variable names, as well as what values they were assigned.

If the template has a Storage Layout attached to it, then the storage history RPC APIs will parse the storage back into 
the variables in the contract; this only works for Solidity compiled contracts. It can handle primitive types, as well 
as static/dynamic arrays and structs, but it currently does not handle mappings. This is because Solidity does not store 
the keys of a map to be used later, rather preferring to work with a key at runtime as it is needed, to save on gas 
costs.
