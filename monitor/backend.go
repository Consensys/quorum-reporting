package monitor

import (
	"quorumengineering/quorum-report/client"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool
type Backend struct {
	client *client.QuorumClient
	monitor *MonitorService
}
