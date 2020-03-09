package main

import (
	"flag"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/monitor"
	"quorumengineering/quorum-report/types"
)

func main() {
	// define flags with default value
	quorumWSURL := flag.String("quorumWSURL", "ws://localhost:23000", "WebSocket URL to quorum node")
	quorumGraphQLURL := flag.String("quorumGraphQLURL", "http://localhost:8547/graphql", "GraphQL URL to quorum node")
	addresses := flag.String("addresses", "", "Common separated hex contract addresses")
	rpcAddr := flag.String("rpcaddr", "localhost:6666", "HTTP-RPC server listening interface")
	rpccors := flag.String("rpccors", "localhost", "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.")
	rpcvhosts := flag.String("rpcvhosts", "", "Common separated hex contract addresses")
	// once all flags are declared, call flag.Parse() to execute the command-line parsing.
	flag.Parse()

	// process all flags to make them into a single `flags` struct
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

	monitorBackend, err := monitor.New(flags)
	if err != nil {
		log.Fatalf("exiting... err: %v.\n", err)
		return
	}

	monitorBackend.Start()
}
