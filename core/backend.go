package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/monitor"
	"quorumengineering/quorum-report/core/rpc"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool.
type Backend struct {
	monitor *monitor.MonitorService
	rpc     *rpc.RPCService
}

func New(flags *types.Flags) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(flags.QuorumWSURL, flags.QuorumGraphQLURL)
	if err != nil {
		return nil, err
	}
	db := database.NewMemoryDB()
	return &Backend{
		monitor: monitor.NewMonitorService(db, quorumClient),
		//filter: filter(db, quorumClient, flags.Addresses),
		rpc: rpc.NewRPCService(db, flags.RPCAddress, flags.RPCVHOSTS, flags.RPCCORS),
	}, nil
}

func (b *Backend) Start() {
	// Start monitor service.
	go b.monitor.Start()
	// TODO: Start filter service.
	// Start local RPC service.
	go b.rpc.Start()

	// cleaning...
	defer func() {
		b.rpc.Stop()
	}()

	// Keep process alive before killed.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("Process stopped by SIGINT or SIGTERM.")
}
