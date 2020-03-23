package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/filter"
	"quorumengineering/quorum-report/core/monitor"
	"quorumengineering/quorum-report/core/rpc"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool.
type Backend struct {
	monitor *monitor.MonitorService
	filter  *filter.FilterService
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
		filter:  filter.NewFilterService(db, flags.Addresses),
		rpc:     rpc.NewRPCService(db, flags.RPCAddress, flags.RPCVHOSTS, flags.RPCCORS),
	}, nil
}

func (b *Backend) Start() {
	// Start monitor service.
	go b.monitor.Start()
	// Start filter service.
	go b.filter.Start()
	// Start local RPC service.
	go b.rpc.Start()

	// cleaning...
	defer func() {
		b.rpc.Stop()
		b.filter.Stop()
	}()

	// Keep process alive before killed.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("Process stopped by SIGINT or SIGTERM.")
}
