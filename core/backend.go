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

	backendErrorChan chan error
}

func New(config types.ReportingConfig) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(config.Connection.WSUrl, config.Connection.GraphQLUrl)
	if err != nil {
		log.Error("Failed to initialize Quorum Client", "err", err)
		// auto reconnect
		if config.Connection.MaxReconnectTries == 0 {
			return nil, err
		}
		for i := 0; i < config.Connection.MaxReconnectTries && err != nil; i++ {
			log.Error("Trying to reconnect", "wait-time", config.Connection.ReconnectInterval)
			time.Sleep(time.Duration(config.Connection.ReconnectInterval) * time.Second)
			quorumClient, err = client.NewQuorumClient(config.Connection.WSUrl, config.Connection.GraphQLUrl)
		}
		// max retries reached but still erroring, abort
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
	log.Info("Adding addresses from configuration file to database")
	initialAddresses := []common.Address{}
	for _, address := range config.Addresses {
		if address.From > 0 {
			// register address from a given block number
			if err := db.AddAddressFrom(address.Address, address.From); err != nil {
				return nil, err
			}
		} else {
			initialAddresses = append(initialAddresses, address.Address)
		}
	}
	// bulk update initial addresses without from
	if err := db.AddAddresses(initialAddresses); err != nil {
		return nil, err
	}
	log.Info("Assigning address templates from configuration file to database")
	// assign all addresses
	for _, address := range config.Addresses {
		if address.TemplateName != "" {
			if err := db.AssignTemplate(address.Address, address.TemplateName); err != nil {
				return nil, err
			}
			log.Info("Assign template to initial registered contract", "template", address.TemplateName, "address", address.Address.Hex())
		}
	}

	backendErrorChan := make(chan error)
	return &Backend{
		monitor:          monitor.NewMonitorService(db, quorumClient, consensus, config),
		filter:           filter.NewFilterService(db, quorumClient),
		rpc:              rpc.NewRPCService(db, config, backendErrorChan),
		db:               db,
		backendErrorChan: backendErrorChan,
	}, nil
}

func (b *Backend) GetBackendErrorChannel() chan error {
	return b.backendErrorChan
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
