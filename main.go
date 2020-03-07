package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"os/signal"
	"quorumengineering/quorum-report/monitor"
	"syscall"
)

// hardcoded parameters should be converted into flags after integrating with github.com/urfave/cli/v2
var (
	quorumWSURL      = "ws://localhost:23000"
	quorumGraphQLURL = "http://localhost:8547/graphql"
	addresses        = []common.Address{
		common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34"),
	}
)

func main() {
	var err error
	backend, err := monitor.New(quorumWSURL, quorumGraphQLURL, addresses)
	if err != nil {
		fmt.Printf("exiting... err: %v.\n", err)
		return
	}

	backend.Start()

	// keep process alive before killed
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")
}
