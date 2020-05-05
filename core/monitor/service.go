package monitor

import (
	"log"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
)

// MonitorService starts all monitors. It pulls data from Quorum node and update the database.
type MonitorService struct {
	blockMonitor *BlockMonitor
}

func NewMonitorService(db database.Database, quorumClient client.Client, consensus string) *MonitorService {
	return &MonitorService{
		blockMonitor: NewBlockMonitor(db, quorumClient, consensus),
	}
}

func (m *MonitorService) Start() error {
	log.Println("Start monitor service...")

	// BlockMonitor will sync all new blocks and historical blocks.
	// It will invoke TransactionMonitor internally.
	return m.blockMonitor.Start()
}

func (m *MonitorService) Stop() {
	// BlockMonitor will sync all new blocks and historical blocks.
	// It will invoke TransactionMonitor internally.
	m.blockMonitor.Stop()
	log.Println("Monitor service stopped.")
}
