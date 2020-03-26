package core

import (
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

	for _, f := range [](func() error){
		b.monitor.Start, // monitor service
		b.filter.Start,  // filter service
		b.rpc.Start,     // RPC service
	} {
		if err := f(); err != nil {
			log.Fatalf("start up failed: %v.\n", err)
		}
	}

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Println("Got interrupt, shutting down...")
		go b.Stop()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Println("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
		panic("immediate shutdown")
	}()
}

func (b *Backend) Stop() {
	b.rpc.Stop()
	b.filter.Stop()
	b.monitor.Stop()
}
