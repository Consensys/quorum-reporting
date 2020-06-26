package monitor

import (
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

// MonitorService starts all monitors. It pulls data from Quorum node and update the database.
type MonitorService struct {
	db database.Database

	// monitors
	blockMonitor       BlockMonitor
	transactionMonitor TransactionMonitor
	tokenMonitor       TokenMonitor

	// concurrent block processing
	newBlockChan   chan *types.Block
	batchWriteChan chan *BlockAndTransactions
	batchWriter    *BatchWriter
	totalWorkers   int

	// To check we have actually shut down before returning
	shutdownChan chan struct{}
	shutdownWg   sync.WaitGroup
}

func NewMonitorService(db database.Database, quorumClient client.Client, consensus string, config types.ReportingConfig) *MonitorService {
	// rules are only parsed once during monitor service initialization
	var rules []TokenRule
	for _, rule := range config.Rules {
		template, _ := db.GetTemplateDetails(rule.TemplateName)
		if template != nil {
			abi, _ := abi.JSON(strings.NewReader(template.ABI))
			rules = append(rules, TokenRule{
				scope:        rule.Scope,
				deployer:     rule.Deployer,
				templateName: rule.TemplateName,
				eip165:       rule.EIP165,
				abi:          abi,
			})
		}
	}
	newBlockChan := make(chan *types.Block)
	batchWriteChan := make(chan *BlockAndTransactions, config.Tuning.BlockProcessingQueueSize)
	return &MonitorService{
		db:                 db,
		blockMonitor:       NewDefaultBlockMonitor(quorumClient, newBlockChan, consensus),
		transactionMonitor: NewDefaultTransactionMonitor(quorumClient),
		tokenMonitor:       NewDefaultTokenMonitor(quorumClient, rules),
		newBlockChan:       newBlockChan,
		batchWriteChan:     batchWriteChan,
		batchWriter:        NewBatchWriter(db, batchWriteChan, config.Tuning.BlockProcessingFlushPeriod),
		totalWorkers:       3 * runtime.NumCPU(),
		shutdownChan:       make(chan struct{}),
	}
}

func (m *MonitorService) Start() error {
	log.Info("Start monitor service")

	// Start batch writer and workers
	m.startBatchWriter()
	m.startWorkers()

	go m.run()

	return nil
}

func (m *MonitorService) Stop() {
	close(m.shutdownChan)
	m.shutdownWg.Wait()
	log.Info("Monitor service stopped")
}

func (m *MonitorService) startBatchWriter() {
	log.Info("Starting batch writer")
	go func() {
		m.shutdownWg.Add(1)
		m.batchWriter.Run(m.shutdownChan)
		m.shutdownWg.Done()
	}()
}

func (m *MonitorService) startWorkers() {
	log.Info("Starting block processor workers")
	for i := 0; i < m.totalWorkers; i++ {
		go func() {
			m.shutdownWg.Add(1)
			m.startWorker(m.shutdownChan)
			m.shutdownWg.Done()
		}()
	}
}

func (m *MonitorService) startWorker(stopChan <-chan struct{}) {
	for {
		select {
		case block := <-m.newBlockChan:
			// Listen to new block channel and process if new block comes.
			err := m.processBlock(block)
			for err != nil {
				log.Warn("Error processing block", "block number", block.Number, "err", err)
				time.Sleep(time.Second)
				err = m.processBlock(block)
			}
		case <-stopChan:
			log.Debug("Stop message received", "location", "core/monitor/service::startWorker")
			return
		}
	}
}

func (m *MonitorService) run() {
	/*
		We want to sync historical blocks as well as listen to the chain head simultaneously,
		whilst also being able to abort if we are shutting down and retry later if the connection
		to Quorum is lost.

		1. This loop will kick off the processing of historical/chain head syncing.
			a) if an error occurs setting up the chain head subscription, wait for a timeout period and try again
			b) if an error occurs setting up the historical block sync, cancel the chain head sub, wait and try again
		2. If we receive a shutdown message, cancel the chain head listener, wait for the historical block sync to finish and return
		3. If the chain head sub has an error, close the "cancelChan" which will stop the historical sync

		Note: 	errors in the historical sync *after* it is set up will not propagate up to here, but instead be
				handled internally. If the historical sync is cancelled, it returns without giving an error, allowing
				the function to break its internal loop.

				we need to set up the historical sync every time since we may have missed some blocks whilst the
				chain head subscription was down
	*/

	log.Info("Start to sync blocks...")
	m.shutdownWg.Add(1)

	for {
		chStopChan := make(chan bool)
		cancelChan := make(chan bool)
		var wg sync.WaitGroup
		wg.Add(1)

		// listen to chain head
		if err := m.blockMonitor.ListenToChainHead(cancelChan, chStopChan); err != nil {
			log.Error("Subscribe to chain head event error, retrying in 1 second", "err", err)
			time.Sleep(time.Second)
			continue
		}

		// get last persisted block number
		lastPersisted, err := m.db.GetLastPersistedBlockNumber()
		if err != nil {
			log.Error("Get last persisted block number error, retrying in 1 second", "err", err)
			close(chStopChan)
			time.Sleep(time.Second)
			continue
		}

		log.Info("Queried last persisted block", "block number", lastPersisted)
		// sync historic blocks
		if err := m.blockMonitor.SyncHistoricBlocks(lastPersisted, cancelChan, &wg); err != nil {
			log.Error("Sync historic blocks error, retrying in 1 second", "err", err)
			close(chStopChan)
			time.Sleep(time.Second)
			continue
		}

		select {
		case <-m.shutdownChan:
			close(chStopChan)
			<-cancelChan
			wg.Wait()
			m.shutdownWg.Done()
			return
		case <-cancelChan:
			wg.Wait()
			log.Info("Retry in 1 second...")
			time.Sleep(time.Second)
		}
	}
}

func (m *MonitorService) processBlock(block *types.Block) error {
	// Transaction monitor pulls all transactions for the given block.
	fetchedTxns, err := m.transactionMonitor.PullTransactions(block)
	if err != nil {
		return err
	}

	// Token monitor checks if transaction deploys a contract matching auto registration rules.
	for _, tx := range fetchedTxns {
		tokenContracts, err := m.tokenMonitor.InspectTransaction(tx)
		if err != nil {
			return err
		}
		for addr, contractType := range tokenContracts {
			// TODO: error handling?
			m.db.AddAddresses([]common.Address{addr})
			m.db.AssignTemplate(addr, contractType)
		}
	}

	// batch write txs and blocks
	workUnit := &BlockAndTransactions{
		block: block,
		txs:   fetchedTxns,
	}
	m.batchWriteChan <- workUnit

	return nil
}
