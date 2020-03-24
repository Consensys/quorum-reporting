package core

import (
	"fmt"
	"log"
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

func New(config types.ReportInputStruct) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(config.Reporting.WSUrl, config.Reporting.GraphQLUrl)
	if err != nil {
		return nil, err
	}
	db := database.NewMemoryDB()
	lastPersisted := db.GetLastPersistedBlockNumber()

	// add the addresses from config file for
	err = db.AddAddresses(config.Reporting.Addresses)
	if err != nil {
		return nil, err
	}

	return &Backend{
		monitor: monitor.NewMonitorService(db, quorumClient),
		filter:  filter.NewFilterService(db, lastPersisted),
		rpc:     rpc.NewRPCService(db, config.Reporting.RPCAddr, config.Reporting.RPCVHosts, config.Reporting.RPCCorsList),
	}, nil
}

func (b *Backend) Start() {

	for _, f := range []func() error{
		b.monitor.Start, // monitor service
		b.filter.Start,  // filter service
		b.rpc.Start,     // RPC service
	} {
		if err := f(); err != nil {
			log.Fatal("start up failed %v", err)
		}
	}
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
