package monitor

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/filter"
)

// MonitorService starts all filters and listens to them, it pulls data from Quorum node and update the database.
type MonitorService struct {
	db           database.Database // TODO: `db` will change to database.Database after all interfaces are implemented.
	quorumClient *client.QuorumClient
	blockFilter  *filter.BlockFilter
}

func NewMonitorService(db database.Database, quorumClient *client.QuorumClient, addresses []common.Address) *MonitorService {
	return &MonitorService{
		db,
		quorumClient,
		filter.NewBlockFilter(db, quorumClient, addresses),
	}
}

func (m *MonitorService) Start() {
	// BlockFilter is the master filter, it will sync all new blocks and historical blocks.
	// It will process TransactionFilter, StorageFilter internally.
	// It will index Transaction/ Events for registered contracts.
	go m.blockFilter.Start()
}
