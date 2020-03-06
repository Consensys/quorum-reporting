package main

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

func main() {
	quorumClient, err := client.NewQuorumClient("ws://localhost:23000", "http://localhost:8547/graphql")
	if err != nil {
		log.Fatal(err)
	}
	// test chain head event
	go listenToChainHead(quorumClient)

	// test with graphql
	query := `
		query {
			block {
				number
			}
		}
	`

	var resp map[string]interface{}
	resp, err = quorumClient.ExecuteGraphQLQuery(context.Background(), query)

	fmt.Println(resp["block"].(map[string]interface{})["number"])

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")
}

func listenToChainHead(quorumClient *client.QuorumClient) {
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
			fmt.Println(header.Hash().Hex()) // print block head
		}
	}
}