package filter

import (
	"context"
	"fmt"
	"log"
	"math/big"

	ethType "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
)

// TODO: BlockFilter subscribes to new blocks and pull historical blocks.

type BlockFilter struct {
	db            database.BlockDB
	quorumClient  *client.QuorumClient
	lastPersisted uint64
	syncStart     chan uint64
}

func NewBlockFilter(db database.BlockDB, quorumClient *client.QuorumClient) *BlockFilter {
	return &BlockFilter{
		db,
		quorumClient,
		db.GetLastPersistedBlockNumber(),
		make(chan uint64),
	}
}

func (bf *BlockFilter) Start() {
	// Pulling historical blocks since the last persisted while continuously listening to ChainHeadEvent.
	// For every block received, pull transactions/ events related to the registered contracts.

	fmt.Println("Start to sync blocks...")
	// listen to ChainHeadEvent
	go bf.listenToChainHead()

	// sync old blocks
	latestChainHead := <-bf.syncStart
	close(bf.syncStart)

	if latestChainHead > bf.lastPersisted+1 {
		bf.syncBlocks(bf.lastPersisted, latestChainHead)
	}
}

func (bf *BlockFilter) listenToChainHead() {
	headers := make(chan *ethType.Header)
	sub, err := bf.quorumClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		// TODO: should gracefully handle error (if quorum node is down, reconnect?)
		log.Fatalf("subscribe to chain head event failed: %v.\n", err)
	}
	for {
		select {
		case err := <-sub.Err():
			// TODO: should gracefully handle error (if quorum node is down, reconnect?)
			log.Fatalf("chain head event subscription error: %v.\n", err)
		case header := <-headers:
			if !isClosed(bf.syncStart) {
				bf.syncStart <- header.Number.Uint64()
			}
			bf.db.WriteBlock(createBlock(header))
		}
	}
}

func (bf *BlockFilter) syncBlocks(start, end uint64) {
	fmt.Printf("Start to sync historic block from %v to %v. \n", start, end)
	for i := start + 1; i < end; i++ {
		header, err := bf.quorumClient.HeaderByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			// TODO: should gracefully handle error (if quorum node is down, reconnect?)
			log.Fatalf("fetch block %v failed: %v.\n", i, err)
		}
		bf.db.WriteBlock(createBlock(header))
	}
}
