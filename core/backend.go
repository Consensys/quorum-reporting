package core

import (
	"log"
	"time"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/filter"
	"quorumengineering/quorum-report/core/monitor"
	"quorumengineering/quorum-report/core/rpc"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/factory"
	"quorumengineering/quorum-report/types"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool.
type Backend struct {
	monitor *monitor.MonitorService
	filter  *filter.FilterService
	rpc     *rpc.RPCService
	db      database.Database
}

func New(config types.ReportingConfig) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(config.Connection.WSUrl, config.Connection.GraphQLUrl)
	if err != nil {
		if config.Connection.MaxReconnectTries > 0 {
			i := 0
			for err != nil {
				i++
				if i == config.Connection.MaxReconnectTries {
					return nil, err
				}
				log.Printf("Connection error: %v. Trying to reconnect in 3 second...\n", err)
				time.Sleep(time.Duration(config.Connection.ReconnectInterval) * time.Second)
				quorumClient, err = client.NewQuorumClient(config.Connection.WSUrl, config.Connection.GraphQLUrl)
			}
		} else {
			return nil, err
		}
	}

	consensus, err := quorumClient.Consensus()
	if err != nil {
		return nil, err
	}

	dbFactory := factory.NewFactory()
	db, err := dbFactory.Database(config.Database)
	if err != nil {
		return nil, err
	}

	// add addresses from config file as initial registered addresses
	err = db.AddAddresses(config.Addresses)
	if err != nil {
		return nil, err
	}

	return &Backend{
		monitor: monitor.NewMonitorService(db, quorumClient, consensus),
		filter:  filter.NewFilterService(db, quorumClient),
		rpc:     rpc.NewRPCService(db, config.Server.RPCAddr, config.Server.RPCVHosts, config.Server.RPCCorsList),
		db:      db,
	}, nil
}

func (b *Backend) Start() {
	for _, f := range []func() error{
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
	b.db.Stop()
}
