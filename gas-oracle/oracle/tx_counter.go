package oracle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

const (
	MinCountInterval    = 600
	DefaultBlockTime    = 2
	DefaultWorkerNumber = 3 // Limit concurrent block fetch workers
)

type TxCounter struct {
	mu sync.RWMutex

	rpcClient *ethclient.Client

	countInterval uint64
	workerNumber  uint64
	result        uint64
	stop          chan struct{}
}

func NewTxCounter(rpcClient *ethclient.Client, countInterval uint64, workerNumber uint64) *TxCounter {
	if countInterval < MinCountInterval {
		countInterval = MinCountInterval
		log.Info("Using default count interval", "value", countInterval)
	}
	log.Debug("Given count interval", "value", countInterval)
	if workerNumber == 0 {
		workerNumber = DefaultWorkerNumber
		log.Info("Using default worker number", "value", workerNumber)
	}
	log.Debug("Given worker number", "value", workerNumber)

	return &TxCounter{
		rpcClient:     rpcClient,
		countInterval: countInterval,
		workerNumber:  workerNumber,
		result:        0,
		stop:          make(chan struct{}),
	}
}

func (t *TxCounter) Start() {
	log.Info("Starting transaction counter", "interval", t.countInterval)
	go t.updateLoop()
}

func (t *TxCounter) Stop() {
	log.Info("Stopping transaction counter")
	close(t.stop)
}

func (t *TxCounter) Get() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.result
}

func (t *TxCounter) GetEstimatedDailyTxCount() uint64 {
	txCount := t.Get()
	return txCount * 86400 / t.countInterval
}

// updateLoop runs in the background and updates the transaction count periodically
func (t *TxCounter) updateLoop() {
	ticker := time.NewTicker(time.Duration(t.countInterval) * time.Second)
	defer ticker.Stop()

	// Perform initial update
	t.updateTxCount()

	for {
		select {
		case <-ticker.C:
			t.updateTxCount()
		case <-t.stop:
			log.Info("Transaction counter update loop stopped")
			return
		}
	}
}

// updateTxCount fetches the transaction count
func (t *TxCounter) updateTxCount() {
	// Get block range
	currentBlock, err := t.rpcClient.BlockNumber(context.Background())
	if err != nil {
		log.Error("Failed to get current block number", "error", err)
		return
	}
	blockNumber := uint64(t.countInterval / DefaultBlockTime)
	var startBlock uint64 = 0
	if currentBlock > blockNumber {
		startBlock = currentBlock - blockNumber + 1
	}
	endBlock := currentBlock

	log.Debug("Fetching transaction count for blocks",
		"start_block", startBlock,
		"end_block", endBlock,
		"max_concurrent_workers", t.workerNumber)

	// Calculate blocks per interval for logging
	blocksPerInterval := endBlock - startBlock + 1

	// Use channels for concurrent block fetching with semaphore
	type blockResult struct {
		blockNum uint64
		txCount  uint64
		err      error
	}

	resultChan := make(chan blockResult, blocksPerInterval)
	var wg sync.WaitGroup

	// Create a job channel to feed block numbers to workers
	jobChan := make(chan uint64, blocksPerInterval)

	// Start worker goroutines
	for i := range t.workerNumber {
		wg.Add(1)
		go func(workerID uint64) {
			defer wg.Done()
			for blockNum := range jobChan {
				txCount, err := getBlockTransactionCountByNumber(t.rpcClient, blockNum)
				if err != nil {
					resultChan <- blockResult{blockNum: blockNum, txCount: 0, err: err}
					continue
				}
				// Remove the first system transaction
				if txCount > 1 {
					txCount = txCount - 1
				} else {
					txCount = 0
				}

				resultChan <- blockResult{blockNum: blockNum, txCount: txCount, err: nil}
			}
		}(i)
	}

	// Feed jobs to workers
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		jobChan <- blockNum
	}
	close(jobChan) // Signal workers to finish

	// Close result channel when all workers finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var totalTxCount uint64
	blockCount := 0
	errorCount := 0

	for result := range resultChan {
		if result.err != nil {
			log.Debug("Failed to fetch block transaction count", "block_number", result.blockNum, "error", result.err)
			errorCount++
			continue
		}
		totalTxCount += result.txCount
		blockCount++
	}

	log.Info("Fetched transaction count",
		"total_tx_count", totalTxCount,
		"blocks_processed", blockCount,
		"blocks_per_interval", blocksPerInterval,
		"errors", errorCount)

	if errorCount > int(blocksPerInterval/10) {
		log.Error("Too many errors fetching block transaction counts, skipping update tx count")
		return
	}

	t.mu.Lock()
	t.result = totalTxCount * uint64(blocksPerInterval) / uint64(blockCount)
	t.mu.Unlock()

	log.Info("Updated transaction count", "value", t.result)
}

func getBlockTransactionCountByNumber(rpcClient *ethclient.Client, blockNumber uint64) (uint64, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		var result hexutil.Uint
		err := rpcClient.Client().CallContext(ctx, &result, "eth_getBlockTransactionCountByNumber", hexutil.Uint64(blockNumber))
		cancel()

		if err == nil {
			return uint64(result), nil
		}

		lastErr = err
		log.Debug("Block transaction count fetch failed, retrying",
			"block_number", blockNumber,
			"attempt", attempt+1,
			"max_retries", maxRetries,
			"error", err)

		// Add small delay between retries to avoid overwhelming the RPC endpoint
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	return 0, fmt.Errorf("failed to fetch block transaction count after %d attempts: %w", maxRetries+1, lastErr)
}
