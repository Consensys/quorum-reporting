package core

import (
	"log"
	"time"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/filter"
	"quorumengineering/quorum-report/core/monitor"
	"quorumengineering/quorum-report/core/rpc"
	"quorumengineering/quorum-report/database/memory"
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
		if config.Reporting.MaxReconnectTries > 0 {
			i := 0
			for err != nil {
				i++
				if i == config.Reporting.MaxReconnectTries {
					return nil, err
				}
				log.Printf("Connection error: %v. Trying to reconnect in 3 second...\n", err)
				time.Sleep(time.Duration(config.Reporting.ReconnectInterval) * time.Second)
				quorumClient, err = client.NewQuorumClient(config.Reporting.WSUrl, config.Reporting.GraphQLUrl)
			}
		} else {
			return nil, err
		}
	}
	db := memory.NewMemoryDB()

	// add addresses from config file as initial registered addresses
	err = db.AddAddresses(config.Reporting.Addresses)
	if err != nil {
		return nil, err
	}

	return &Backend{
		monitor: monitor.NewMonitorService(db, quorumClient),
		filter:  filter.NewFilterService(db),
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
			log.Panicf("start up failed: %v.\n", err)
		}
	}
}

func (b *Backend) Stop() {
	b.rpc.Stop()
	b.filter.Stop()
	b.monitor.Stop()
}
