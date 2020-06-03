package monitor

import (
	"context"
	"log"
	"runtime"

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
	syncStart    chan uint64
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
		syncStart:    make(chan uint64, 1), // make channel buffered so that it does not block chain head listener
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

	// 2. Sync from last persisted to current block height.
	if err := m.syncHistoricBlocks(); err != nil {
		return err
	}

	// 3. Listen to ChainHeadEvent and sync.
	if err := m.listenToChainHead(); err != nil {
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
		err := m.blockMonitor.syncBlocks(lastPersisted+1, currentBlockNumber)
		for err != nil {
			log.Printf("sync historic blocks from %v to %v failed: %v\n", lastPersisted, currentBlockNumber, err)
			err = m.blockMonitor.syncBlocks(err.EndBlockNumber(), currentBlockNumber)
		}

		// Sync from currentBlockNumber + 1 to the first ChainHeadEvent if there is any gap.
		stopChan, stopSubscription := m.subscribeStopEvent()
		defer stopSubscription.Unsubscribe()

		select {
		case latestChainHead := <-m.syncStart:
			close(m.syncStart)
			err := m.blockMonitor.syncBlocks(currentBlockNumber+1, latestChainHead-1)
			for err != nil {
				log.Printf("sync historic blocks from %v to %v failed: %v\n", currentBlockNumber, latestChainHead-1, err)
				err = m.blockMonitor.syncBlocks(err.EndBlockNumber(), latestChainHead-1)
			}
		case <-stopChan:
			return
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
		syncStarted := false
		for {
			select {
			case err := <-sub.Err():
				log.Panicf("chain head event subscription error: %v", err)
			case header := <-headers:
				if !syncStarted {
					m.syncStart <- header.Number.Uint64()
					syncStarted = true
				}
				m.blockMonitor.processChainHead(header)
			case <-stopChan:
				return
			}
		}
	}()

	return nil
}
