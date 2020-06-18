package core

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/core/filter"
	"quorumengineering/quorum-report/core/monitor"
	"quorumengineering/quorum-report/core/rpc"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/database/factory"
	"quorumengineering/quorum-report/log"
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
			log.Error("Failed to connect to Quorum RPC", "err", err)
			log.Error("Trying to reconnect", "wait-time", config.Connection.ReconnectInterval)
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
	log.Info("Consensus found", "algorithm", consensus)

	dbFactory := factory.NewFactory()
	db, err := dbFactory.Database(config.Database)
	if err != nil {
		return nil, err
	}

	// store all templates
	log.Info("Adding templates from configuration file to database")
	for _, template := range config.Templates {
		if err := db.AddTemplate(template.TemplateName, template.ABI, template.StorageLayout); err != nil {
			return nil, err
		}
	}
	// store all addresses
	initialAddresses := []common.Address{}
	for _, address := range config.Addresses {
		initialAddresses = append(initialAddresses, address.Address)
	}
	log.Info("Adding addresses from configuration file to database")
	if err := db.AddAddresses(initialAddresses); err != nil {
		return nil, err
	}
	// assign all addresses
	for _, address := range config.Addresses {
		if err := db.AssignTemplate(address.Address, address.TemplateName); err != nil {
			return nil, err
		}
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
