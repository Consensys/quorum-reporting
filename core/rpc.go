package monitor

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

var defaultHTTPTimeouts = rpc.HTTPTimeouts{
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	IdleTimeout:  120 * time.Second,
}

type RPCService struct {
	httpEndpoint string
	vhosts       []string
	cors         []string
	apis         []rpc.API
	listener     net.Listener
}

func NewRPCService(httpEndpoint string, vhosts []string, cors []string, apis []rpc.API) *RPCService {
	return &RPCService{
		httpEndpoint: httpEndpoint,
		vhosts:       vhosts,
		cors:         cors,
		apis:         apis,
	}
}

func (r *RPCService) Start() {
	modules := []string{}
	for _, apis := range r.apis {
		modules = append(modules, apis.Namespace)
	}
	listener, _, err := rpc.StartHTTPEndpoint(r.httpEndpoint, r.apis, modules, r.cors, r.vhosts, defaultHTTPTimeouts)
	if err != nil {
		// TODO: should gracefully handle error
		log.Fatalf("rpc service failed to start: %v", err)
	}
	r.listener = listener
	fmt.Printf("HTTP endpoint opened: http://%s.\n", r.httpEndpoint)
}

func (r *RPCService) Stop() {
	r.listener.Close()
	fmt.Printf("HTTP endpoint closed: http://%s.\n", r.httpEndpoint)
}
