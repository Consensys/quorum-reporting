package monitor

import (
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/types"
)

type BlockMonitor struct {
	db                 database.Database
	quorumClient       client.Client
	transactionMonitor *TransactionMonitor
	newBlockChan       chan *types.Block          // concurrent block processing
	batchWriteChan     chan *BlockAndTransactions // concurrent block processing
	consensus          string
}

func NewBlockMonitor(db database.Database, quorumClient client.Client, consensus string, batchWriteChan chan *BlockAndTransactions) *BlockMonitor {
	return &BlockMonitor{
		db:                 db,
		quorumClient:       quorumClient,
		transactionMonitor: NewTransactionMonitor(db, quorumClient),
		newBlockChan:       make(chan *types.Block),
		batchWriteChan:     batchWriteChan,
		consensus:          consensus,
	}
}

func (bm *BlockMonitor) startWorker(stopChan <-chan types.StopEvent) {
	for {
		select {
		case block := <-bm.newBlockChan:
			// Listen to new block channel and process if new block comes.
			err := bm.process(block)
			for err != nil {
				log.Printf("Process block %v error: %v.\n", block.Number, err)
				//log.Println("Waiting a second before processing...") //TODO: this should be at a "debug" level
				time.Sleep(time.Second)
				err = bm.process(block)
			}
		case <-stopChan:
			return
		}
	}
}

func (bm *BlockMonitor) process(block *types.Block) error {
	// Transaction monitor pulls all transactions for the given block.
	fetchedTxns, err := bm.transactionMonitor.PullTransactions(block)
	if err != nil {
		return err
	}

	// Check if transaction deploys a public ERC20/ERC721 contract directly or internally
	for _, tx := range fetchedTxns {
		var addrs []common.Address
		if (tx.CreatedContract != common.Address{0}) {
			addrs = append(addrs, tx.CreatedContract)
		}
		for _, ic := range tx.InternalCalls {
			if ic.Type == "CREATE" || ic.Type == "CREATE2" {
				addrs = append(addrs, ic.To)
			}
		}
		for _, addr := range addrs {
			res, err := client.GetCode(bm.quorumClient, addr, tx.BlockHash)
			if err != nil {
				return err
			}

			// check ERC20
			if checkAbiMatch(types.ERC20ABI, res) {
				log.Printf("tx %v deploys %v which is a potential ERC20 contract.\n", tx.Hash.Hex(), addr.Hex())
				// add contract address
				bm.db.AddAddresses([]common.Address{tx.CreatedContract})
				// assign ERC20 template
				bm.db.AssignTemplate(tx.CreatedContract, types.ERC20)
			}

			// check ERC721
			if checkAbiMatch(types.ERC721ABI, res) {
				log.Printf("tx %v deploys %v which is a potential ERC721 contract.\n", tx.Hash.Hex(), addr.Hex())
				// add contract address
				bm.db.AddAddresses([]common.Address{tx.CreatedContract})
				// assign ERC721 template
				bm.db.AssignTemplate(tx.CreatedContract, types.ERC721)
			}
		}
	}

	// batch write txs and blocks
	workunit := &BlockAndTransactions{
		block: block,
		txs:   fetchedTxns,
	}
	bm.batchWriteChan <- workunit
	return nil
}

func (bm *BlockMonitor) currentBlockNumber() (uint64, error) {
	var currentBlockResult graphql.CurrentBlockResult
	if err := bm.quorumClient.ExecuteGraphQLQuery(&currentBlockResult, graphql.CurrentBlockQuery()); err != nil {
		return 0, err
	}
	return hexutil.DecodeUint64(currentBlockResult.Block.Number)
}

func (bm *BlockMonitor) syncBlocks(start, end uint64, stopChan chan bool) *types.SyncError {
	if start > end {
		return nil
	}

	log.Printf("Start to sync historic blocks from %v to %v. \n", start, end)
	for i := start; i <= end; i++ {
		select {
		case <-stopChan:
			return nil
		default:
		}

		blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			// TODO: if quorum node is down, reconnect?
			return types.NewSyncError(err.Error(), i)
		}

		select {
		case <-stopChan:
			return nil
		case bm.newBlockChan <- bm.createBlock(blockOrigin):
		}
	}
	return nil
}

func (bm *BlockMonitor) processChainHead(header *ethTypes.Header) {
	blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), header.Number)
	for err != nil {
		log.Printf("get block with hash %v error: %v\n", header.Hash(), err)
		blockOrigin, err = bm.quorumClient.BlockByNumber(context.Background(), header.Number)
	}
	bm.newBlockChan <- bm.createBlock(blockOrigin)
}

func (bm *BlockMonitor) createBlock(block *ethTypes.Block) *types.Block {
	txs := []common.Hash{}
	for _, tx := range block.Transactions() {
		txs = append(txs, tx.Hash())
	}

	timestamp := block.Time()
	if bm.consensus == "raft" {
		timestamp = timestamp / 1_000_000_000
	}

	return &types.Block{
		Hash:         block.Hash(),
		ParentHash:   block.ParentHash(),
		StateRoot:    block.Root(),
		TxRoot:       block.TxHash(),
		ReceiptRoot:  block.ReceiptHash(),
		Number:       block.NumberU64(),
		GasLimit:     block.GasLimit(),
		GasUsed:      block.GasUsed(),
		Timestamp:    timestamp,
		ExtraData:    block.Extra(),
		Transactions: txs,
	}
}

func checkAbiMatch(abiToCheck abi.ABI, data hexutil.Bytes) bool {
	for _, b := range abiToCheck.Methods {
		if !strings.Contains(data.String(), common.Bytes2Hex(b.ID())) {
			return false
		}
	}
	for _, event := range abiToCheck.Events {
		if !strings.Contains(data.String(), event.ID().Hex()[2:]) {
			return false
		}
	}
	return true
}
