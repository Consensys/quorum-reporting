package monitor

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/rpc"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// Backend wraps MonitorService and QuorumClient, controls the start/stop of the reporting tool.
type Backend struct {
	monitor *MonitorService
	rpc     *RPCService
}

func New(flags *types.Flags) (*Backend, error) {
	quorumClient, err := client.NewQuorumClient(flags.QuorumWSURL, flags.QuorumGraphQLURL)
	if err != nil {
		return nil, err
	}
	db := database.NewMemoryDB()
	apis := []rpc.API{
		{
			"reporting",
			"1.0",
			db,
			true,
		},
	}

	return &Backend{
		monitor: NewMonitorService(db, quorumClient, flags.Addresses),
		rpc:     NewRPCService(flags.RPCAddress, flags.RPCVHOSTS, flags.RPCCORS, apis), // Crudely expose all database API endpoints for now...
	}, nil
}

func (b *Backend) Start() {
	// Start monitor service.
	go b.monitor.Start()
	// Start local RPC service.
	listener, err := b.rpc.Start()
	if err != nil {
		log.Fatalf("rpc service failed to start: %v", err)
		return
	}

	// keep process alive before killed
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("exiting")

	defer func() {
		listener.Close()
		fmt.Printf("HTTP endpoint closed: http://%s", b.rpc.httpEndpoint)
	}()
}
