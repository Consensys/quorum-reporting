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

func (r *RPCService) Start() error {
	modules := []string{}
	for _, apis := range r.apis {
		modules = append(modules, apis.Namespace)
	}
	listener, _, err := rpc.StartHTTPEndpoint(r.httpEndpoint, r.apis, modules, r.cors, r.vhosts, defaultHTTPTimeouts)
	if err != nil {
		return err
	}
	r.listener = listener
	fmt.Printf("HTTP endpoint opened: http://%s.\n", r.httpEndpoint)
	return nil
}

func (r *RPCService) Stop() {
	r.listener.Close()
	fmt.Printf("HTTP endpoint closed: http://%s.\n", r.httpEndpoint)
}
