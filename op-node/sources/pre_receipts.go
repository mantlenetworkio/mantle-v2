package sources

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
	"github.com/ethereum/go-ethereum/log"
	gosync "sync"
	"time"
)

const maxConcurent = 16

type PreFetcher struct {
	l1      *L1Client
	resetL1 chan uint64

	l1BlockCache *caching.OrderCache[eth.L1BlockRef]

	log    log.Logger
	wg     gosync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func NewPreFetcher(
	l1 *L1Client,
	log log.Logger,
	metrics caching.Metrics,
) *PreFetcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &PreFetcher{
		l1:           l1,
		resetL1:      make(chan uint64, 1),
		l1BlockCache: caching.NewOrderCache[eth.L1BlockRef](metrics, "receipts", 12),
		log:          log,
		ctx:          ctx,
		cancel:       cancel,
		done:         make(chan struct{}),
	}
}

func (p *PreFetcher) Start() error {
	log.Info("Starting receipts pre fetcher")
	p.wg.Add(1)
	go p.evenLoop()
	return nil
}

func (p *PreFetcher) Close() error {
	p.cancel()
	p.done <- struct{}{}
	p.wg.Wait()
	return nil
}

func (p *PreFetcher) Reset(ctx context.Context, start uint64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.resetL1 <- start:
		return nil
	}
}

func (p *PreFetcher) evenLoop() {
	defer p.wg.Done()
	p.log.Info("receipts pre fetcher started")
	defer p.log.Info("receipts pre fetcher returned")
	defer p.cancel()
	var lastUnsafeL1 uint64
	for {
		select {
		case number := <-p.resetL1:
			lastUnsafeL1 = number
			p.l1BlockCache.RemoveAll()
			p.l1.GetRecProvider().GetReceiptsCache().RemoveAll()
		case <-p.ctx.Done():
			return
		case <-p.done:
			return
		default:
			if lastUnsafeL1 == 0 {
				continue
			}
			if p.l1.GetRecProvider().GetReceiptsCache().IsFull() {
				continue
			}
			blockRef, err := p.l1.BlockRefByLabel(p.ctx, eth.Unsafe)
			if err != nil {
				p.log.Debug("failed to fetch the unsafe latest block", "err", err)
				time.Sleep(2 * time.Second)
				continue
			}
			if lastUnsafeL1 > blockRef.Number {
				p.log.Error("last unsafe l1 bigger than latest block, something wrong about l1 provider", "cache block", lastUnsafeL1, "latest block", blockRef.Number)
				time.Sleep(2 * time.Second)
				continue
			}
			//check l1 reorg
			v, b := p.checkL1Reorg(&lastUnsafeL1)
			if !b {
				time.Sleep(2 * time.Second)
				continue
			}
			if lastUnsafeL1 != v {
				lastUnsafeL1 = v
			}

			if err := p.processBatch(p.ctx, &lastUnsafeL1, blockRef.Number); err != nil {
				p.log.Debug("failed to process batch", "err", err)
				time.Sleep(2 * time.Second)
			}
		}
	}
}

type BlockFetchTask struct {
	blockNumber uint64
	ctx         context.Context
	result      chan<- eth.L1BlockRef
}

type BlockFetchResult struct {
	blockInfo eth.L1BlockRef
	err       error
}

type WorkerPool struct {
	workers  int
	taskChan chan BlockFetchTask
	done     chan struct{}
	log      log.Logger
	l1       *L1Client
}

func NewWorkerPool(workers int, log log.Logger, l1 *L1Client) *WorkerPool {
	return &WorkerPool{
		workers:  workers,
		taskChan: make(chan BlockFetchTask),
		done:     make(chan struct{}),
		log:      log,
		l1:       l1,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		go wp.worker()
	}
}

func (wp *WorkerPool) Stop() {
	close(wp.done)
}

func (wp *WorkerPool) worker() {
	for {
		select {
		case <-wp.done:
			return
		case task := <-wp.taskChan:
			wp.processTask(task)
		}
	}
}

func (wp *WorkerPool) processTask(task BlockFetchTask) {
	ctx, cancel := context.WithTimeout(task.ctx, 30*time.Second)
	defer cancel()

	// Create a backoff strategy
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 25 * time.Second
	b.Multiplier = 2
	b.RandomizationFactor = 0.1

	// Use backoff.WithContext to support context cancellation
	backoffCtx := backoff.WithContext(b, ctx)

	operation := func() error {
		if result := wp.tryFetchBlock(ctx, task.blockNumber); result != nil {
			task.result <- *result
			return nil
		}
		return fmt.Errorf("failed to fetch block")
	}

	// Use Retry to perform retries
	err := backoff.Retry(operation, backoffCtx)
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			wp.log.Debug("task context cancelled", "blockNumber", task.blockNumber)
		} else {
			wp.log.Warn("max retries exceeded",
				"blockNumber", task.blockNumber,
				"err", err)
		}
	}
}

func (wp *WorkerPool) tryFetchBlock(ctx context.Context, blockNumber uint64) *eth.L1BlockRef {

	blockInfo, err := wp.l1.L1BlockRefByNumber(ctx, blockNumber)
	if err != nil {
		wp.log.Debug("failed to fetch block ref", "err", err, "blockNumber", blockNumber)
		return nil
	}

	if pair, ok := wp.l1.GetRecProvider().GetReceiptsCache().Get(blockNumber, false); ok {
		if err == nil && pair.BlockHash == blockInfo.Hash {
			return &blockInfo
		}
	}

	isSuccess, err := wp.l1.PreFetchReceipts(ctx, blockInfo.Hash)
	if err != nil {
		wp.log.Warn("failed to pre-fetch receipts", "err", err)
		return nil
	}
	if !isSuccess {
		wp.log.Debug("receipts cache full",
			"blockHash", blockInfo.Hash,
			"blockNumber", blockNumber)
		return nil
	}

	wp.log.Debug("pre-fetching receipts done",
		"block", blockInfo.Number,
		"hash", blockInfo.Hash)
	return &blockInfo
}

func (p *PreFetcher) processBatch(ctx context.Context, currentL1Block *uint64, l1UnsafeBlockNumber uint64) error {

	// calculate task
	taskCount := p.calculateTaskCount(*currentL1Block, l1UnsafeBlockNumber)
	if taskCount == 0 {
		return nil
	}

	// create result chan
	results := make(chan eth.L1BlockRef, taskCount)

	// create fetch receipts pool
	pool := NewWorkerPool(maxConcurent, p.log, p.l1)
	pool.Start()
	defer pool.Stop()

	// submit a task
	startBlock := *currentL1Block
	for i := uint64(1); i <= uint64(taskCount); i++ {
		task := BlockFetchTask{
			blockNumber: startBlock + i,
			ctx:         ctx,
			result:      results,
		}
		pool.taskChan <- task
	}

	// Collect results
	for i := 1; i <= taskCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-results:
			p.l1BlockCache.AddAndRemove(result.Number, result)
		}
	}

	// update current l1 block
	*currentL1Block = startBlock + uint64(taskCount)

	return nil
}

func (p *PreFetcher) calculateTaskCount(current, latest uint64) int {
	if current > latest {
		return 0
	}

	remaining := latest - current

	if remaining >= maxConcurent {
		return maxConcurent
	}
	return int(remaining)
}

func (p *PreFetcher) checkL1Reorg(lastUnsafeL1 *uint64) (uint64, bool) {
	//check L1 reorg
	minCacheBlock := p.l1BlockCache.GetMin()

	maxNumber := *lastUnsafeL1
	for i := maxNumber; i >= minCacheBlock.Number; i-- {
		latestL1Block, err := p.l1.BlockRefByNumber(p.ctx, i)
		if err != nil {
			p.log.Warn("failed to fetch the cache latest block", "err", err)
			return i, false
		}

		if cacheL1, ok := p.l1BlockCache.Get(i, true); ok {
			if latestL1Block.Hash == cacheL1.Hash {
				return i, true
			}
		} else {
			p.log.Warn("check l1 reorg, can not get cache from l1BlockCache, there maybe something wrong, check!", "lastUnsafeL1", i)
		}
	}
	p.log.Error("minCacheBlock is different, reorg deeper than 12 also occurred on L1", "min cache block num", minCacheBlock.Number, "hash", minCacheBlock.Hash)
	return minCacheBlock.Number, false

}
