package monitor

import (
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/filter"
)

// TODO: MonitorService start all filters and listens to them, it pulls data from Quorum node and update the database.
type MonitorService struct {
	db           database.BlockDB // TODO: db will change to database.Database after all interfaces are implemented.
	quorumClient *client.QuorumClient
	address      []common.Address
	blockFilter  *filter.BlockFilter
}

func NewMonitorService(db database.BlockDB, quorumClient *client.QuorumClient, addresses []common.Address) *MonitorService {
	return &MonitorService{
		db,
		quorumClient,
		addresses,
		filter.NewBlockFilter(db, quorumClient),
	}
}

func (m *MonitorService) Start() {
	// Start block syncing.
	// Pulling historical blocks since the last persisted while continuously listening to ChainHeadEvent.
	// For every block received, pull transactions/ events related to the registered contracts.
	m.blockFilter.Start()
}
