package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/graphql"
)

func main() {
	quorumClient, err := client.NewQuorumClient("ws://localhost:23000", "http://localhost:8547/graphql")
	if err != nil {
		log.Fatal(err)
	}

	// graphql test

	var resp map[string]interface{}
	resp, err = quorumClient.ExecuteGraphQLQuery(context.Background(), graphql.CurrentBlockQuery())

	fmt.Println(resp["block"].(map[string]interface{})["number"])

	resp, err = quorumClient.ExecuteGraphQLQuery(context.Background(), graphql.TransactionDetailQuery(common.HexToHash("0xde70fc6431cbde6f2b6d92d77745fdc5aa1521d2e81a8969a72e3a9997e8ae0d")))

	fmt.Println(resp["transaction"].(map[string]interface{}))

	// TODO: How to parse GraphQL result effectively?
}
