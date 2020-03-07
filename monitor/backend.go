package monitor

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool.
type Backend struct {
	monitor *MonitorService
}

func New(quorumWSURL string, quorumGraphQLURL string, addresses []common.Address) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(quorumWSURL, quorumGraphQLURL)
	if err != nil {
		return nil, err
	}
	db := database.NewMemoryDB()

	return &Backend{
		monitor: NewMonitorService(db, quorumClient, addresses),
	}, nil
}

func (b *Backend) Start() error {
	// The first to start is block syncing.
	// Then pulling new blocks since the last persisted while continuously listening to ChainHeadEvent.
	go b.monitor.StartBlockSync()
	return nil
}
