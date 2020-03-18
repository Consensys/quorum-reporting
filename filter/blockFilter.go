package filter

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
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

	// 1. Fetch the current block height
	currentHead, err := bf.getCurrentHead()
	if err != nil {
		// TODO: should gracefully handle error (if quorum node is down, reconnect?)
		log.Fatalf("get current head error: %v.\n", err)
	}
	fmt.Println("Current block head is: %v", currentHead)

	// 2. Sync from last persisted to current block height
	go bf.syncBlocks(bf.lastPersisted, currentHead)

	// 3. Listen to ChainHeadEvent and sync
	go bf.listenToChainHead()
	latestChainHead := <-bf.syncStart
	close(bf.syncStart)

	// 4. Sync from current block height + 1 to the first ChainHeadEvent if there is any gap
	go bf.syncBlocks(currentHead, latestChainHead)
}

func (bf *BlockFilter) getCurrentHead() (uint64, error) {
	query := `
		query {
			block {
				number
			}
		}
	`
	var resp map[string]interface{}
	resp, err := bf.quorumClient.ExecuteGraphQLQuery(context.Background(), query)
	if err != nil {
		return 0, err
	}
	return hexutil.DecodeUint64(resp["block"].(map[string]interface{})["number"].(string))
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
