package main

import (
	"context"
	"fmt"
	"log"

	"quorumengineering/quorum-report/client"
)

func main() {
	quorumClient, err := client.NewQuorumClient("ws://localhost:23000", "http://localhost:8547/graphql")
	if err != nil {
		log.Fatal(err)
	}

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
}
