package client

import (
	"quorumengineering/quorum-report/types"
)

type Client interface {
	// SubscribeChainHead subscribes to new chain header
	SubscribeChainHead(chan<- types.RawHeader) error
	// ExecuteGraphQLQuery performs a fully constructed query against the Geth
	// GraphQL server
	ExecuteGraphQLQuery(interface{}, string) error
	// RPCCall makes a JSON RPC call to the Geth RPC server
	RPCCall(interface{}, string, ...interface{}) error
	// Stop quorum client connection
	Stop()
}
