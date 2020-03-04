# Quorum Reporting

## Requirements
# | Area | Requirement 
:---: | :---: | :--- 
1 | Admin | Ability for an admin to register a contract address for monitoring and reporting
2 | Admin | Registration should allow flexibility to monitor <ul><li>Contract events - It should be possible to monitor all or a subset of contract events</li><li>State changes - It should be possible to monitor state change of all or subset of contract attributes </li><li>Internal transactions created as a part of contract execution</li></ul>
3 | Data Fetch | Once a contract is registered for monitoring/reporting, reporting tools should fetch the following historical data from geth <ul><li>Events</li><li>Transactions </li><li>State changes</li></ul>
4 | Data Fetch | For all registered contracts, reporting tool to poll continuously on Quprum geth node for new blocks and if any block contains transactions on the registered contracts, it should fetch all applicable data for reporting
5 | Data Fetch | Data fetch should cater for scenarios when the reporting service was restarted and there is a gap in data captured and current geth data
6 | Data storage | Reporting tool to have its own reporting data base with a well defined data schema for easy querying and reporting. The data fetched from Quourm geth node should be stored here
7 | Dashboard and UI | UI for the following activities <ul><li>Registration of contracts for monitoring with ability to select subset of contract events and storage attribites</li><li>UI displaying all contract transactions, related event logs, internal transactions and state changes</ul>

## Approach
* Recommend building this on top of 1.9.7 version as 1.9.7 supports `graphql` and provides a flexible querying mechanism
* Registration
    * The solidity contract code of the contract to be monitored to be given as input and parsed for attributes, functions and events
    * The events, attributes to be displayed on UI for user to select subset for monitoring
    * User selection is stored in reporting DB
* Data fetch
    * `go` routine which will subscribe to `newChainHead` event of Quorum geth node on a websocket connection
    * For every block check if there are any transactions on registered contracts, if yes then fetch block and transaction data using
        * `graphql` queries
        * `eth` apis
        * use dump account state to fetch the contract state at a given block height
    

