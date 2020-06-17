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
    - Quorum Reporting fetches a lot of historic data that is pruned by Quorum under default `full` gcmode. It is recommended to run Quorum in `archive` mode.
    
    e.g. `geth --graphql --graphql.vhosts=* --ws --wsport 23000 --wsapi admin,eth,debug --wsorigins=* --gcmode=archive ...`

- ElasticSearch v7 (For Production)
    - Quorum Reporting uses ElasticSearch as its data store, and can be set up in many configurations.
        [Click here](https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html) to get started with ElasticSearch.

### Up & Running

#### Using Binary

##### Build

```bash
go build [-o quorum-reporting]
```

##### Run

- Running with default configuration file path of `config.toml`
```bash
./quorum-report
```
- Running with custom configuration path
```bash
./quorum-report -config <path to config file>
```
- Running with help
```bash
./quorum-report -help
```

#### Using Docker

##### Build
```bash
docker build . -t quorum-reporting
```

##### Run

- A configuration must be supplied to the Docker container
```bash
docker run -p <port mapping> --mount type=bind,source=<path to config>,target=/config.toml quorum-reporting:latest
```

### Configuration

A [sample configuration](./config.sample.toml) file has been provided with details about each of the options.

### Interact with Quorum Reporting through RPC

The application has a set of RPC API's that are used to interact with the application. See [here](core/rpc/README.md) for all the available RPC APIs.

## Development

### Pre-Requisites

- golang 1.13

### Development Environment

- Clone the Git repo
```bash
git clone https://github.com/QuorumEngineering/quorum-reporting.git
```
- Fetch dependencies using gomod
```bash
go get ./...
```

## Design & Roadmap

Refer to [design document](design.md).