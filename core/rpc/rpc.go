package rpc

import (
	"log"
	"net"
	"time"

	ethRPC "github.com/ethereum/go-ethereum/rpc"

	"quorumengineering/quorum-report/database"
)

var defaultHTTPTimeouts = ethRPC.HTTPTimeouts{
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	IdleTimeout:  120 * time.Second,
}

type RPCService struct {
	httpEndpoint string
	vhosts       []string
	cors         []string
	apis         []ethRPC.API
	listener     net.Listener
}

func NewRPCService(db database.Database, httpEndpoint string, vhosts []string, cors []string) *RPCService {
	apis := []ethRPC.API{
		{
			"reporting",
			"1.0",
			NewRPCAPIs(db),
			true,
		},
	}
	return &RPCService{
		httpEndpoint: httpEndpoint,
		vhosts:       vhosts,
		cors:         cors,
		apis:         apis,
	}
}

func (r *RPCService) Start() error {
	log.Println("Start rpc service...")

	modules := []string{}
	for _, apis := range r.apis {
		modules = append(modules, apis.Namespace)
	}
	listener, _, err := ethRPC.StartHTTPEndpoint(r.httpEndpoint, r.apis, modules, r.cors, r.vhosts, defaultHTTPTimeouts)
	if err != nil {
		return err
	}
	r.listener = listener
	log.Printf("HTTP endpoint opened: http://%s.\n", r.httpEndpoint)
	return nil
}

func (r *RPCService) Stop() {
	r.listener.Close()
	log.Printf("HTTP endpoint closed: http://%s.\n", r.httpEndpoint)
}
