package filter

import (
	"runtime"
	"sync"
	"time"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type StorageFilter struct {
	db           FilterServiceDB
	quorumClient client.Client

	outstandingBlocks sync.WaitGroup
	maxEntriesToSave  int

	incomingBlockChan chan AccountStateWithBlock
	pulledStateChan   chan AccountStateWithBlock

	shutdownWg      sync.WaitGroup
	shutdownChannel chan struct{}
}

type AccountStateWithBlock struct {
	BlockNumber  uint64
	AccountState map[types.Address]*types.AccountState
	Addresses    []types.Address
}

func NewStorageFilter(db FilterServiceDB, quorumClient client.Client) *StorageFilter {
	sf := &StorageFilter{
		db:                db,
		quorumClient:      quorumClient,
		maxEntriesToSave:  100,
		incomingBlockChan: make(chan AccountStateWithBlock),
		pulledStateChan:   make(chan AccountStateWithBlock, 1000),

		shutdownChannel: make(chan struct{}),
	}

	log.Info("Starting storage filter state fetch workers", "number", runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		sf.shutdownWg.Add(1)
		sf.StateFetchWorker()
	}
	log.Info("Started storage filter state fetch workers")
	log.Info("Starting storage filter state saving workers", "number", runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		sf.shutdownWg.Add(1)
		sf.StateSavingWorker()
	}
	log.Info("Started storage filter state saving workers")

	return sf
}

func (sf *StorageFilter) IndexStorage(addresses []types.Address, startBlockNumber, endBlockNumber uint64) error {
	log.Info("Indexing storage", "start", startBlockNumber, "end", endBlockNumber)
	for i := startBlockNumber; i <= endBlockNumber; i++ {
		sf.outstandingBlocks.Add(1)
		emptyStorage := AccountStateWithBlock{
			BlockNumber:  i,
			AccountState: make(map[types.Address]*types.AccountState),
			Addresses:    addresses,
		}
		sf.incomingBlockChan <- emptyStorage
	}

	sf.outstandingBlocks.Wait()
	log.Info("Indexing storage complete", "start", startBlockNumber, "end", endBlockNumber)
	return nil
}

func (sf *StorageFilter) StateFetchWorker() {
	go func() {
		defer sf.shutdownWg.Done()
		var lastFilteredMap = make(map[types.Address]bool)
		for {
			select {
			case <-sf.shutdownChannel:
				log.Debug("Shutdown request received", "loc", "storage filter - state fetch worker")
				return
			case blockToPull := <-sf.incomingBlockChan:

				log.Debug("Fetching contract storage", "block number", blockToPull.BlockNumber)
				for _, address := range blockToPull.Addresses {
					found, filtered := lastFilteredMap[address]
					if !found || !filtered {
						if lastFiltered, err := sf.db.GetLastFiltered(address); err != nil {
							log.Error("getLastFiltered failed", "address", address, "err", err)
						} else {
							if lastFiltered > 0 {
								lastFilteredMap[address] = true
							} else {
								lastFilteredMap[address] = false
							}
						}
					}

					changed, err := sf.didStorageRootChange(address, blockToPull.BlockNumber)
					for err != nil {
						changed, err = sf.didStorageRootChange(address, blockToPull.BlockNumber)
					}
					if !changed {
						continue
					}

					log.Debug("Fetching contract storage", "address", address.String(), "block number", blockToPull.BlockNumber)
					dumpAccount, err := client.DumpAddress(sf.quorumClient, address, blockToPull.BlockNumber-1, blockToPull.BlockNumber, lastFilteredMap[address])
					for err != nil {
						log.Error("Unable to fetch contract state", "address", address.String(), "block number", blockToPull.BlockNumber, "err", err)
						time.Sleep(time.Second) //TODO: make adaptive or block until websocket available
						dumpAccount, err = client.DumpAddress(sf.quorumClient, address, blockToPull.BlockNumber-1, blockToPull.BlockNumber, lastFilteredMap[address])
					}
					blockToPull.AccountState[address] = dumpAccount
				}
				sf.pulledStateChan <- blockToPull
			}
		}
	}()
}

func (sf *StorageFilter) StateSavingWorker() {
	go func() {
		defer sf.shutdownWg.Done()
		for {
			storage := make([]AccountStateWithBlock, 0)

			select {
			case st := <-sf.pulledStateChan:
				storage = append(storage, st)
			case <-sf.shutdownChannel:
				log.Debug("Shutdown request received", "loc", "storage filter - state save worker")
				return
			}

			for {
				if len(storage) == sf.maxEntriesToSave {
					break
				}

				isEmpty := false
				select {
				case st := <-sf.pulledStateChan:
					storage = append(storage, st)
				default:
					isEmpty = true
				}
				if isEmpty {
					break
				}
			}

			log.Debug("Saving storage entries", "number of entries", len(storage))

			sf.SaveStorage(storage)
		}
	}()
}

func (sf *StorageFilter) SaveStorage(storage []AccountStateWithBlock) {
	var thisRunWg sync.WaitGroup
	thisRunWg.Add(len(storage))

	saveSingle := func(storageData AccountStateWithBlock) {
		defer thisRunWg.Done()

		log.Debug("Persisting storage", "blockNum", storageData.BlockNumber)
		err := sf.db.IndexStorage(storageData.AccountState, storageData.BlockNumber)
		//TODO: use error channel for returning error instead of looping
		for err != nil {
			err = sf.db.IndexStorage(storageData.AccountState, storageData.BlockNumber)
		}
		sf.outstandingBlocks.Done()
	}

	for _, storageData := range storage {
		//TODO: change storage indexing to accept all data at once to remove goroutine call
		go saveSingle(storageData)
	}

	thisRunWg.Wait()
}

func (sf *StorageFilter) Stop() {
	log.Info("Stopping down storage filter")
	sf.outstandingBlocks.Wait()
	close(sf.shutdownChannel)
	log.Info("Finished stopping storage filter")
}

func (sf *StorageFilter) didStorageRootChange(contract types.Address, blockNum uint64) (bool, error) {
	storageRootThisBlock, err := client.StorageRoot(sf.quorumClient, contract, blockNum)
	if err != nil {
		return false, err
	}

	storageRootPrevBlock, err := client.StorageRoot(sf.quorumClient, contract, blockNum-1)
	if err != nil {
		return false, err
	}

	return storageRootPrevBlock != storageRootThisBlock, nil
}
