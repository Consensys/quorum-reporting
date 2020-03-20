package monitor

import (
	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
)

// MonitorService starts all monitors. It pulls data from Quorum node and update the database.
type MonitorService struct {
	db           database.Database // TODO: `db` will change to database.Database after all interfaces are implemented.
	quorumClient *client.QuorumClient
	blockMonitor *BlockMonitor
}

func NewMonitorService(db database.Database, quorumClient *client.QuorumClient) *MonitorService {
	return &MonitorService{
		db,
		quorumClient,
		NewBlockMonitor(db, quorumClient),
	}
}

func (m *MonitorService) Start() {
	// BlockMonitor will sync all new blocks and historical blocks.
	// It will invoke TransactionMonitor internally.
	go m.blockMonitor.Start()
}
