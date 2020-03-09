package monitor

import (
	"fmt"
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
}

func NewRPCService(httpEndpoint string, vhosts []string, cors []string, apis []rpc.API) *RPCService {
	return &RPCService{
		httpEndpoint,
		vhosts,
		cors,
		apis,
	}
}

func (r *RPCService) Start() (net.Listener, error) {
	modules := []string{}
	for _, apis := range r.apis {
		modules = append(modules, apis.Namespace)
	}
	listener, _, err := rpc.StartHTTPEndpoint(r.httpEndpoint, r.apis, modules, r.cors, r.vhosts, defaultHTTPTimeouts)
	if err != nil {
		return nil, err
	}
	fmt.Printf("HTTP endpoint opened: http://%s", r.httpEndpoint)
	return listener, nil
}
