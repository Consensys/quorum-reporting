package monitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/filter"
)

// TODO: MonitorService start all filters and listens to them, it pulls data from Quorum node and update the database.
type MonitorService struct {
	db           database.BlockDB // db will change to database.Database after all interfaces are implemented.
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

func (m *MonitorService) StartBlockSync() error {
	fmt.Println("Start to sync blocks...")
	m.blockFilter.Start()
	return nil
}
