package main

import (
	"flag"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/core"
	"quorumengineering/quorum-report/types"
)

func main() {
	// Define flags with default values.
	quorumWSURL := flag.String("quorumWSURL", "ws://localhost:23000", "WebSocket URL to quorum node")
	quorumGraphQLURL := flag.String("quorumGraphQLURL", "http://localhost:8547/graphql", "GraphQL URL to quorum node")
	addresses := flag.String("addresses", "", "Common separated hex contract addresses")
	rpcAddr := flag.String("rpcaddr", "localhost:6666", "HTTP-RPC server listening interface")
	rpccors := flag.String("rpccors", "localhost", "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.")
	rpcvhosts := flag.String("rpcvhosts", "localhost", "Common separated hex contract addresses")
	// Once all flags are declared, call flag.Parse() to execute the command-line parsing.
	flag.Parse()

	// Process all flags to make them into a single `flags` struct.
	addressList := []common.Address{}
	for _, a := range strings.Split(*addresses, ",") {
		addressList = append(addressList, common.HexToAddress(a))
	}
	rpccorsList := []string{}
	for _, h := range strings.Split(*rpccors, ",") {
		rpccorsList = append(rpccorsList, h)
	}
	rpcvhostsList := []string{}
	for _, h := range strings.Split(*rpcvhosts, ",") {
		rpcvhostsList = append(rpcvhostsList, h)
	}
	flags := &types.Flags{
		QuorumWSURL:      *quorumWSURL,
		QuorumGraphQLURL: *quorumGraphQLURL,
		RPCAddress:       *rpcAddr,
		RPCCORS:          rpccorsList,
		RPCVHOSTS:        rpcvhostsList,
		Addresses:        addressList,
	}

	backend, err := core.New(flags)
	if err != nil {
		log.Fatalf("exiting... err: %v.\n", err)
		return
	}

	backend.Start()
}
