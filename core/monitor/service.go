package monitor

import (
	"context"
	"log"
	"runtime"
	"time"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/types"
)

// MonitorService starts all monitors. It pulls data from Quorum node and update the database.
type MonitorService struct {
	db           database.Database
	quorumClient client.Client
	blockMonitor *BlockMonitor
	batchWriter  *BatchWriter
	stopFeed     event.Feed
	totalWorkers uint64
}

func NewMonitorService(db database.Database, quorumClient client.Client, consensus string, tuningConfig types.TuningConfig) *MonitorService {
	batchWriteChan := make(chan *BlockAndTransactions, tuningConfig.BlockProcessingQueueSize)
	return &MonitorService{
		db:           db,
		quorumClient: quorumClient,
		blockMonitor: NewBlockMonitor(db, quorumClient, consensus, batchWriteChan),
		batchWriter:  NewBatchWriter(batchWriteChan, db),
		totalWorkers: 3 * uint64(runtime.NumCPU()),
	}
}

func (m *MonitorService) Start() error {
	log.Println("Start monitor service...")

	// Pulling historical blocks since the last persisted while continuously listening to ChainHeadEvent.
	// For every block received, pull transactions/ events related to the registered contracts.

	log.Println("Start to sync blocks...")

	// 1. Start batch writer and workers
	m.startBatchWriter()
	m.startWorkers()

	// 2. Listen to ChainHeadEvent and sync.
	if err := m.listenToChainHead(); err != nil {
		return err
	}

	// 3. Sync from last persisted to current block height.
	if err := m.syncHistoricBlocks(); err != nil {
		return err
	}

	return nil
}

func (m *MonitorService) Stop() {
	m.stopFeed.Send(types.StopEvent{})
	log.Println("Monitor service stopped.")
}

func (m *MonitorService) subscribeStopEvent() (chan types.StopEvent, event.Subscription) {
	c := make(chan types.StopEvent)
	s := m.stopFeed.Subscribe(c)
	return c, s
}

func (m *MonitorService) startBatchWriter() {
	go func() {
		stopChan, stopSubscription := m.subscribeStopEvent()
		defer stopSubscription.Unsubscribe()
		m.batchWriter.Run(stopChan)
	}()
}

func (m *MonitorService) startWorkers() {
	for i := uint64(0); i < m.totalWorkers; i++ {
		go func() {
			stopChan, stopSubscription := m.subscribeStopEvent()
			defer stopSubscription.Unsubscribe()
			m.blockMonitor.startWorker(stopChan)
		}()
	}
}

func (m *MonitorService) syncHistoricBlocks() error {
	currentBlockNumber, err := m.blockMonitor.currentBlockNumber()
	if err != nil {
		return err
	}
	log.Printf("Current block head is: %v.\n", currentBlockNumber)
	lastPersisted, err := m.db.GetLastPersistedBlockNumber()
	if err != nil {
		return err
	}

	// Sync is called in a go routine so that it doesn't block main process.
	go func() {
		stopChan, stopSubscription := m.subscribeStopEvent()
		defer stopSubscription.Unsubscribe()
		err := m.blockMonitor.syncBlocks(lastPersisted+1, currentBlockNumber, stopChan)
		for err != nil {
			log.Printf("sync historic blocks from %v to %v failed: %v\n", lastPersisted, currentBlockNumber, err)
			<-time.NewTicker(time.Second).C
			err = m.blockMonitor.syncBlocks(err.EndBlockNumber(), currentBlockNumber, stopChan)
		}
	}()

	return nil
}

func (m *MonitorService) listenToChainHead() error {
	headers := make(chan *ethTypes.Header)
	sub, err := m.quorumClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		return err
	}

	go func() {
		stopChan, stopSubscription := m.subscribeStopEvent()
		defer stopSubscription.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				log.Panicf("chain head event subscription error: %v", err)
			case header := <-headers:
				m.blockMonitor.processChainHead(header)
			case <-stopChan:
				return
			}
		}
	}()

	return nil
}
