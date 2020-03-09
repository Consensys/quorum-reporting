package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/monitor"
)

func main() {
	// define flags with default value
	quorumWSURL := flag.String("quorumWSURL", "ws://localhost:23000", "websocket url to quorum node")
	quorumGraphQLURL := flag.String("quorumGraphQLURL", "http://localhost:8547/graphql", "graphql url to quorum node")
	addresses := flag.String("addresses", "", "common separated hex contract addresses")
	// once all flags are declared, call flag.Parse() to execute the command-line parsing.
	flag.Parse()

	addressList := []common.Address{}
	for _, a := range strings.Split(*addresses, ",") {
		addressList = append(addressList, common.HexToAddress(a))
	}

	var err error
	monitorBackend, err := monitor.New(*quorumWSURL, *quorumGraphQLURL, addressList)
	if err != nil {
		fmt.Printf("exiting... err: %v.\n", err)
		return
	}

	monitorBackend.Start()

	// keep process alive before killed
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")
}
