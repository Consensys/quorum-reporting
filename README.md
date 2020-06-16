# Quorum Reporting

Quorum Reporting is a tool that indexes the blockchain and generates reports to 
give users insights into what is happening to their contracts.

It generates reports of the contracts state, allow you to see how the contract changes and evolves
over its lifetime. It will also interpret and parse events the contract has emitted in a human
readable manner, meaning they can be viewed in a dashboard or other application with little extra
effort.

## Usage 

### Pre-requisites
- Running Quorum
    - Quorum needs to be run with GraphQL and websockets open, with `eth`, `admin` and `debug` endpoints available.
    - Quorum Reporting fetches a lot of data that is usually cleaned up by Quorum, so it is recommended to run Quorum in archive mode
    - e.g. `geth --graphql --graphql.port 11000 --graphql.vhosts=* --ws --wsport 10000 --wsapi admin,eth,debug --wsorigins=* --gcmode=archive ...`

- ElasticSearch v7
    - Quorum Reporting uses ElasticSearch as its data store, and can be set up in many configurations.
        [Click here](https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html) to get started with ElasticSearch

### Running

#### Using native binary

- Running with default configuration file path of `config.toml`:
`./quorum-report`

- Running with custom configuration path:
`./quorum-report -config <path to config file>`

- Running with help
`./quorum-report -help`

#### Using Docker

A configuration must be supplied to the Docker container (there is no default)
docker run -p 6666:6666 --mount type=bind,source=<path to config>,target=/config.toml quorum-reporting:latest

### Configuration
A [sample configuration](./config.sample.toml) file has been provided with details about each of the options [here]

## Using the application

The application has a set of RPC API's that are used to interact with the application
- add new templates
- add new addresses to index
- fetch contract storage
- fetch transactions

See [here](core/rpc/README.md) for all the RPC APIs.

## Development

### Pre-Requisites
- golang 1.13

### Development Environment
- Clone the Git repo with `git clone https://github.com/QuorumEngineering/quorum-reporting.git`
- Fetch dependencies using gomod: `go get ./...`

### Building

#### Build on native OS
Building the project uses standard go tooling: `go build` or `go build -o quorum-reporting`

#### Build with Docker
```bash
docker build . -t quorum-reporting
```
