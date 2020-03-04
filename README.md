# Quorum Reporting

## Requirements
Number | Area | Requirement 
:---: | :---: | :--- 
1 | Admin | Ability for an admin to register a contract address for monitoring and reporting
2 | Admin | Registration should allow flexibility to monitor <ul><li>Contract events - It should be possible to monitor all or a subset of contract events</li><li>State changes - It should be possible to monitor state change of all or subset of contract attributes </li><li>Internal transactions created as a part of contract execution</li></ul>
3 | Data Fetch | Once a contract is registered for monitoring/reporting, reporting tools should fetch the following historical data from geth <ul><li>Events</li><li>Transactions </li><li>State changes</li></ul>
4 | Data Fetch | For all registered contracts, reporting tool to poll continuously on Quorum geth node for new blocks and if the state of the registered contracts have changed in the new block, it should fetch all applicable data for reporting
5 | Data Fetch | Data fetch should cater for scenarios when the reporting service was restarted and there is a gap in data captured and current geth data
6 | Data storage | Reporting tool to have its own reporting database with a well defined data schema for easy querying and reporting. The data fetched from Quorum geth node should be stored here
7 | Dashboard and UI | UI for the following activities <ul><li>Registration of contracts for monitoring with ability to select subset of contract events and storage attributes</li><li>UI displaying all contract transactions, related event logs, internal transactions and state changes</ul>

## Approach
* To build the tool on top of 1.9.7 version as 1.9.7 supports `graphql` and provides a flexible querying mechanism
* Registration
    * The solidity contract code of the contract to be monitored to be given as input and parsed for attributes, functions and events
    * The events, attributes to be displayed on UI for user to select subset for monitoring
    * User selection is stored in reporting DB
* Data fetch
    * `go` routine which will subscribe to `newChainHead` event of 
    rum geth node on a websocket connection
    * The routine check if the state of the registered contracts has changed for each new block and if yes fetch all applicable data
    * Data fetch to use the following:
        * `graphql` queries
        * RPC APIs
* Data storage
    * *yet to finalize on storage schema and database*
* User interface 
    * UI to be aligned with Cakeshop UI: Initial version of the tool to add the dashboard and configuration pages to Cakeshop UI
        