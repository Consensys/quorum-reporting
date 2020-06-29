package rpc

import (
	"fmt"
	"net"
	"time"

	ethRPC "github.com/ethereum/go-ethereum/rpc"

	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
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

func NewRPCService(db database.Database, config types.ReportingConfig) *RPCService {
	go MakeServer(db, config)

	apis := []ethRPC.API{
		{
			"reporting",
			"1.0",
			NewRPCAPIs(db),
			true,
		},
	}
	return &RPCService{
		httpEndpoint: config.Server.RPCAddr,
		vhosts:       config.Server.RPCVHosts,
		cors:         config.Server.RPCCorsList,
		apis:         apis,
	}
}

func (r *RPCService) Start() error {

	log.Info("Starting rpc service")

	//var modules []string
	//for _, apis := range r.apis {
	//	modules = append(modules, apis.Namespace)
	//}
	//listener, _, err := ethRPC.StartHTTPEndpoint(r.httpEndpoint, r.apis, modules, r.cors, r.vhosts, defaultHTTPTimeouts)
	//if err != nil {
	//	return err
	//}
	//r.listener = listener
	log.Info("RPC HTTP endpoint opened", "url", fmt.Sprintf("http://%s", r.httpEndpoint))
	return nil
}

func (r *RPCService) Stop() {
	if r.listener != nil {
		r.listener.Close()
	}
	log.Info("RPC HTTP endpoint closed", "url", fmt.Sprintf("http://%s", r.httpEndpoint))
	log.Info("RPC service stopped")
}
