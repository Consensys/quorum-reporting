package core

import (
	"fmt"
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
		if config.Connection.MaxReconnectTries == 0 {
			return nil, err
		}

		for i := 0; i < config.Connection.MaxReconnectTries && err != nil; i++ {
			log.Printf("Connection error: %v. Trying to reconnect in %d second...\n", err, config.Connection.ReconnectInterval)
			time.Sleep(time.Duration(config.Connection.ReconnectInterval) * time.Second)
			quorumClient, err = client.NewQuorumClient(config.Connection.WSUrl, config.Connection.GraphQLUrl)
		}

		//max retries reached but still erroring, abort
		if err != nil {
			return nil, err
		}
	}

	consensus, err := client.Consensus(quorumClient)
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
		monitor: monitor.NewMonitorService(db, quorumClient, consensus, config.Tuning),
		filter:  filter.NewFilterService(db, quorumClient),
		rpc:     rpc.NewRPCService(db, config.Server.RPCAddr, config.Server.RPCVHosts, config.Server.RPCCorsList),
		db:      db,
	}, nil
}

func (b *Backend) Start() error {
	for _, f := range []func() error{
		b.monitor.Start, // monitor service
		b.filter.Start,  // filter service
		b.rpc.Start,     // RPC service
	} {
		if err := f(); err != nil {
			return fmt.Errorf("start up failed: %v", err)
		}
	}
	return nil
}

func (b *Backend) Stop() {
	b.rpc.Stop()
	b.filter.Stop()
	b.monitor.Stop()
	b.db.Stop()
}
