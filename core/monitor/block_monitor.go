package monitor

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/database"
	"quorumengineering/quorum-report/graphql"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

type BlockMonitor struct {
	db                 database.Database
	quorumClient       client.Client
	transactionMonitor *TransactionMonitor
	newBlockChan       chan *types.Block          // concurrent block processing
	batchWriteChan     chan *BlockAndTransactions // concurrent block processing
	consensus          string
	tokenMonitor       TokenMonitor
}

func NewBlockMonitor(db database.Database, quorumClient client.Client, consensus string, batchWriteChan chan *BlockAndTransactions) *BlockMonitor {
	return &BlockMonitor{
		db:                 db,
		quorumClient:       quorumClient,
		transactionMonitor: NewTransactionMonitor(db, quorumClient),
		newBlockChan:       make(chan *types.Block),
		batchWriteChan:     batchWriteChan,
		consensus:          consensus,
		tokenMonitor:       NewDefaultTokenMonitor(quorumClient),
	}
}

func (bm *BlockMonitor) startWorker(stopChan <-chan types.StopEvent) {
	for {
		select {
		case block := <-bm.newBlockChan:
			// Listen to new block channel and process if new block comes.
			err := bm.process(block)
			for err != nil {
				log.Warn("Error processing block", "block number", block.Number, "err", err)
				time.Sleep(time.Second)
				err = bm.process(block)
			}
		case <-stopChan:
			log.Debug("Stop message received", "location", "BlockMonitor::startWorker")
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
		if (tx.CreatedContract != common.Address{}) {
			addrs = append(addrs, tx.CreatedContract)
		}
		for _, ic := range tx.InternalCalls {
			if ic.Type == "CREATE" || ic.Type == "CREATE2" {
				addrs = append(addrs, ic.To)
			}
		}

		tokens, err := bm.tokenMonitor.InspectAddresses(addrs, tx, block.Number)
		if err != nil {
			return err
		}

		for addr, contractType := range tokens {
			bm.db.AddAddresses([]common.Address{addr})
			bm.db.AssignTemplate(addr, contractType)
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
	log.Debug("Fetching current block number")

	var currentBlockResult graphql.CurrentBlockResult
	if err := bm.quorumClient.ExecuteGraphQLQuery(&currentBlockResult, graphql.CurrentBlockQuery()); err != nil {
		return 0, err
	}

	log.Debug("Current block number found", "number", currentBlockResult.Block.Number)
	return hexutil.DecodeUint64(currentBlockResult.Block.Number)
}

func (bm *BlockMonitor) syncBlocks(start, end uint64, stopChan chan bool) *types.SyncError {
	if start > end {
		return nil
	}

	log.Info("Syncing historic blocks", "start", start, "end", end)
	for i := start; i <= end; i++ {
		select {
		case <-stopChan:
			return nil
		default:
		}

		blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			return types.NewSyncError(err.Error(), i)
		}

		select {
		case <-stopChan:
			return nil
		case bm.newBlockChan <- bm.createBlock(blockOrigin):
		}
	}
	log.Info("Complete historical sync finished")
	return nil
}

func (bm *BlockMonitor) processChainHead(header *ethTypes.Header) {
	log.Info("Processing chain head", "block hash", header.Hash().String(), "block number", header.Number.String())
	blockOrigin, err := bm.quorumClient.BlockByNumber(context.Background(), header.Number)
	for err != nil {
		log.Warn("Error fetching block from Quorum", "block hash", header.Hash(), "block number", header.Number.String(), "err", err)
		time.Sleep(1 * time.Second) //TODO: return err and let caller handle?
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

func (bm *BlockMonitor) checkEIP165(address common.Address, blockNum uint64) (string, error) {
	//check if the contract implements EIP165

	eip165Call, err := client.CallEIP165(bm.quorumClient, address, common.Hex2Bytes("01ffc9a70"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}

	if !eip165Call {
		return "", nil
	}

	eip165CallCheck, err := client.CallEIP165(bm.quorumClient, address, common.Hex2Bytes("ffffffff"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if eip165CallCheck {
		return "", nil
	}

	//now we know it implements EIP165, so lets check the interfaces

	erc20check, err := client.CallEIP165(bm.quorumClient, address, common.Hex2Bytes("36372b07"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if erc20check {
		return types.ERC20, nil
	}

	erc721check, err := client.CallEIP165(bm.quorumClient, address, common.Hex2Bytes("80ac58cd"), new(big.Int).SetUint64(blockNum))
	if err != nil {
		return "", err
	}
	if erc721check {
		return types.ERC721, nil
	}

	return "", nil
}
