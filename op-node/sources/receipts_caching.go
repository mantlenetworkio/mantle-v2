package sources

import (
	"context"
	"sync"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// A CachingReceiptsProvider caches successful receipt fetches from the inner
// ReceiptsProvider. It also avoids duplicate in-flight requests per block hash.
type CachingReceiptsProvider struct {
	inner InnerReceiptsProvider
	cache *caching.OrderCache[*ReceiptsHashPair]

	// lock fetching process for each block hash to avoid duplicate requests
	fetching   map[common.Hash]*sync.Mutex
	fetchingMu sync.Mutex // only protects map
}

func NewCachingReceiptsProvider(inner InnerReceiptsProvider, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return &CachingReceiptsProvider{
		inner:    inner,
		cache:    caching.NewOrderCache[*ReceiptsHashPair](m, "receipts", cacheSize),
		fetching: make(map[common.Hash]*sync.Mutex),
	}
}

func NewCachingRPCReceiptsProvider(client rpcClient, log log.Logger, config RPCReceiptsConfig, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return NewCachingReceiptsProvider(NewRPCReceiptsFetcher(client, log, config), m, cacheSize)
}

func (p *CachingReceiptsProvider) getOrCreateFetchingLock(blockHash common.Hash) *sync.Mutex {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	if mu, ok := p.fetching[blockHash]; ok {
		return mu
	}
	mu := new(sync.Mutex)
	p.fetching[blockHash] = mu
	return mu
}

func (p *CachingReceiptsProvider) deleteFetchingLock(blockHash common.Hash) {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	delete(p.fetching, blockHash)
}

// FetchReceipts fetches receipts for the given block and transaction hashes
// it expects that the inner FetchReceipts implementation handles validation
func (p *CachingReceiptsProvider) FetchReceipts(ctx context.Context, blockInfo eth.BlockInfo, txHashes []common.Hash, isPreFetch bool) (types.Receipts, error, bool) {
	block := eth.ToBlockID(blockInfo)
	var isFull bool

	if v, ok := p.cache.Get(block.Number, !isPreFetch); ok && v.BlockHash == block.Hash {
		return v.Receipts, nil, isFull
	}

	mu := p.getOrCreateFetchingLock(block.Hash)
	mu.Lock()
	defer mu.Unlock()
	// Other routine might have fetched in the meantime
	if v, ok := p.cache.Get(block.Number, !isPreFetch); ok && v.BlockHash == block.Hash {
		// we might have created a new lock above while the old
		// fetching job completed.
		p.deleteFetchingLock(block.Hash)
		return v.Receipts, nil, isFull
	}

	isFull = p.cache.IsFull()
	if isFull && isPreFetch {
		return nil, nil, isFull
	}

	r, err := p.inner.FetchReceipts(ctx, blockInfo, txHashes)
	if err != nil {
		return nil, err, isFull
	}

	p.cache.AddIfNotFull(block.Number, &ReceiptsHashPair{BlockHash: block.Hash, Receipts: r})
	// result now in cache, can delete fetching lock
	p.deleteFetchingLock(block.Hash)
	return r, nil, isFull
}

func (p *CachingReceiptsProvider) isInnerNil() bool {
	return p.inner == nil
}

func (p *CachingReceiptsProvider) GetReceiptsCache() *caching.OrderCache[*ReceiptsHashPair] {
	return p.cache
}
