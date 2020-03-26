package examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
)

func chainHeadExample() {
	quorumClient, err := client.NewQuorumClient("ws://localhost:23000", "http://localhost:8547/graphql")
	if err != nil {
		log.Fatal(err)
	}

	headers := make(chan *types.Header)
	sub, err := quorumClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			// Print block head.
			fmt.Println(header.Number)
		}
	}

	// Keep process alive before killed.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")
}
