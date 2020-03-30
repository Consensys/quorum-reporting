# Quorum Reporting

## Requirements
Number | Area | Requirement 
:---: | :---: | :--- 
1 | Admin | Ability for an admin to register a contract address for monitoring and reporting
2 | Admin | Registration should allow flexibility to monitor <ul><li>Contract events - It should be possible to monitor all or a subset of contract events</li><li>State changes - It should be possible to monitor state change of all or subset of contract attributes </li><li>Internal transactions created as a part of contract execution</li></ul>
3 | Data Fetch | Once a contract is registered for monitoring/reporting, reporting tools should fetch the following historical data from geth <ul><li>Events</li><li>Transactions </li><li>State changes</li></ul>
4 | Data Fetch | For all registered contracts, reporting tool to poll continuously on Quorum geth node for new blocks and if the state of the registered contracts have changed in the new block, it should fetch all applicable data for reporting
5 | Data Fetch | Data fetch should cater for scenarios when the reporting service was restarted and thus the reporting db data is behind the Quorum geth node data.
6 | Data storage | Reporting tool to have its own reporting database with a well defined data schema for easy querying and reporting. The data fetched from Quorum geth node to be stored here
7 | Dashboard and UI | UI for the following activities <ul><li>Registration of contracts for monitoring with ability to select subset of contract events and storage attributes</li><li>UI displaying all contract transactions, related event logs, internal transactions and state changes with drill down capability</ul>

## Approach
* To build the tool on top of 1.9.7 version as 1.9.7 supports `graphql` and provides a flexible querying mechanism
* Registration
    * The solidity contract code of the contract to be monitored to be given as input and parsed for attributes, functions and events
    * The events, attributes to be displayed on UI for user to select subset for monitoring
    * User selection is stored in reporting DB
* Data fetch
    * `go` routine which will subscribe to `newChainHead` event of 
    geth node on a websocket connection
    * The routine checks if the state of the registered contracts has changed for each new block and if yes fetch all applicable data
    * Data fetch to use the following:
        * `graphql` queries
        * RPC APIs
* Data storage
    * *yet to finalize on storage schema and database*
* User interface 
    * Dashboard and configuration options to be added current Cakeshop UI for the first version

## Up and Running

* After clone the repo, build `quorum-report` tool
```bash
go build
```
* See usage of `quorum-report`
```bash
./quorum-report --help
```
* Start `quorum-report` tool with default params
```
./quorum-report --config config.toml
```

## Architecture & Design


```
Quorum Reporting -----> [ Backend ] ----------> [ RPC Service ]
                           |   |                       |
                           |   |                       |
                           |   +--> [ Filter Service ] |
                           |             |      |      |
                           |             |      |      |
                           V             |      |      |
                  [ Monitor Service ]    |      |      |
                           |             |      |      |
                           |             |      |      |
   +-----------------------|-------------+      |      |
   |                       |                    |      |
   |                       |                    |      |
   V                       V                    V      V
Quorum <--------- [ Block Monitor ] ----------> Database <---------- Visualization (e.g. Cakeshop)
   ^                       |                       ^
   |                       |                       |
   |                       |                       | 
   |                       |                       | 
   |                       |                       |
   |                       V                       |
   +---------- [ Transaction Monitor ] ------------+
```

#### Items Required in Persistent Database

- All blocks
- All transactions
- Storage at each block for registered contract addresses
- **optional:** Indices (transactions/ events linked to registered contract addresses). While this may be implicitly 
achieved by database, we may still store the indices result for easier query of transactions/ events.

## Roadmap

#### Phase 0 (done)

- Complete the base code architecture
- Sync blocks & Store blocks/ transactions in a memory database
- Filter transactions by registered addresses
- Filter events by registered addresses
- Dynamically change registered addresses, clean up and refilter
- Expose basic RPC endpoints to serve queries
- Unit tests & CI/CD

#### Phase 1 (in progress)

- Integrate persistent database
- Handle restart & recover from fail-stop scenarios
- Filter contract detailed storage by registered addresses (with dumpAccount available on geth side)
- Resolve internal calls (incoming/ outgoing) for transactions of registered addresses

#### Phase 3 (todo)

- Docker file & make file support
- Integrate UI for visualization
- Fully functional reporting tool
