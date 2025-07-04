package sources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/client"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/sources/caching"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type L2ClientConfig struct {
	EthClientConfig

	L2BlockRefsCacheSize int
	L1ConfigsCacheSize   int

	RollupCfg *rollup.Config
}

func L2ClientDefaultConfig(config *rollup.Config, trustRPC bool) *L2ClientConfig {
	// Cache 3/2 worth of sequencing window of payloads, block references, receipts and txs
	span := int(config.SeqWindowSize) * 3 / 2
	// Estimate number of L2 blocks in this span of L1 blocks
	// (there's always one L2 block per L1 block, L1 is thus the minimum, even if block time is very high)
	if config.BlockTime < 12 && config.BlockTime > 0 {
		span *= 12
		span /= int(config.BlockTime)
	}
	fullSpan := span
	if span > 1000 { // sanity cap. If a large sequencing window is configured, do not make the cache too large
		span = 1000
	}
	return &L2ClientConfig{
		EthClientConfig: EthClientConfig{
			// receipts and transactions are cached per block
			ReceiptsCacheSize:     span,
			TransactionsCacheSize: span,
			HeadersCacheSize:      span,
			PayloadsCacheSize:     span,
			MaxRequestsPerBatch:   20, // TODO: tune batch param
			MaxConcurrentRequests: 10,
			TrustRPC:              trustRPC,
			MustBePostMerge:       true,
			RPCProviderKind:       RPCKindBasic,
			MethodResetDuration:   time.Minute,
		},
		// Not bounded by span, to cover find-sync-start range fully for speedy recovery after errors.
		L2BlockRefsCacheSize: fullSpan,
		L1ConfigsCacheSize:   span,
		RollupCfg:            config,
	}
}

// L2Client extends EthClient with functions to fetch and cache eth.L2BlockRef values.
type L2Client struct {
	*EthClient
	rollupCfg *rollup.Config

	// cache L2BlockRef by hash
	// common.Hash -> eth.L2BlockRef
	l2BlockRefsCache *caching.LRUCache[common.Hash, eth.L2BlockRef]

	// cache SystemConfig by L2 hash
	// common.Hash -> eth.SystemConfig
	systemConfigsCache *caching.LRUCache[common.Hash, eth.SystemConfig]
}

// NewL2Client constructs a new L2Client instance. The L2Client is a thin wrapper around the EthClient with added functions
// for fetching and caching eth.L2BlockRef values. This includes fetching an L2BlockRef by block number, label, or hash.
// See: [L2BlockRefByLabel], [L2BlockRefByNumber], [L2BlockRefByHash]
func NewL2Client(client client.RPC, log log.Logger, metrics caching.Metrics, config *L2ClientConfig) (*L2Client, error) {
	ethClient, err := NewEthClient(client, log, metrics, &config.EthClientConfig)
	if err != nil {
		return nil, err
	}

	return &L2Client{
		EthClient:          ethClient,
		rollupCfg:          config.RollupCfg,
		l2BlockRefsCache:   caching.NewLRUCache[common.Hash, eth.L2BlockRef](metrics, "blockrefs", config.L2BlockRefsCacheSize),
		systemConfigsCache: caching.NewLRUCache[common.Hash, eth.SystemConfig](metrics, "systemconfigs", config.L1ConfigsCacheSize),
	}, nil
}

// L2BlockRefByLabel returns the [eth.L2BlockRef] for the given block label.
func (s *L2Client) L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	envelope, err := s.PayloadByLabel(ctx, label)
	if err != nil {
		// Both geth and erigon like to serve non-standard errors for the safe and finalized heads, correct that.
		// This happens when the chain just started and nothing is marked as safe/finalized yet.
		if strings.Contains(err.Error(), "block not found") || strings.Contains(err.Error(), "Unknown block") {
			err = ethereum.NotFound
		}
		// w%: wrap to preserve ethereum.NotFound case
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine L2BlockRef of %s, could not get payload: %w", label, err)
	}
	ref, err := derive.PayloadToBlockRef(envelope.ExecutionPayload, &s.rollupCfg.Genesis)
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	s.l2BlockRefsCache.Add(ref.Hash, ref)
	return ref, nil
}

// L2BlockRefByNumber returns the [eth.L2BlockRef] for the given block number.
func (s *L2Client) L2BlockRefByNumber(ctx context.Context, num uint64) (eth.L2BlockRef, error) {
	envelope, err := s.PayloadByNumber(ctx, num)
	if err != nil {
		// w%: wrap to preserve ethereum.NotFound case
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine L2BlockRef of height %v, could not get payload: %w", num, err)
	}
	ref, err := derive.PayloadToBlockRef(envelope.ExecutionPayload, &s.rollupCfg.Genesis)
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	s.l2BlockRefsCache.Add(ref.Hash, ref)
	return ref, nil
}

// L2BlockRefByHash returns the [eth.L2BlockRef] for the given block hash.
// The returned BlockRef may not be in the canonical chain.
func (s *L2Client) L2BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L2BlockRef, error) {
	if ref, ok := s.l2BlockRefsCache.Get(hash); ok {
		return ref, nil
	}

	envelope, err := s.PayloadByHash(ctx, hash)
	if err != nil {
		// w%: wrap to preserve ethereum.NotFound case
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of hash %v, could not get payload: %w", hash, err)
	}
	ref, err := derive.PayloadToBlockRef(envelope.ExecutionPayload, &s.rollupCfg.Genesis)
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	s.l2BlockRefsCache.Add(ref.Hash, ref)
	return ref, nil
}

// SystemConfigByL2Hash returns the [eth.SystemConfig] (matching the config updates up to and including the L1 origin) for the given L2 block hash.
// The returned [eth.SystemConfig] may not be in the canonical chain when the hash is not canonical.
func (s *L2Client) SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error) {
	if ref, ok := s.systemConfigsCache.Get(hash); ok {
		return ref, nil
	}

	envelope, err := s.PayloadByHash(ctx, hash)
	if err != nil {
		// w%: wrap to preserve ethereum.NotFound case
		return eth.SystemConfig{}, fmt.Errorf("failed to determine block-hash of hash %v, could not get payload: %w", hash, err)
	}
	cfg, err := derive.PayloadToSystemConfig(envelope.ExecutionPayload, s.rollupCfg)
	if err != nil {
		return eth.SystemConfig{}, err
	}
	s.systemConfigsCache.Add(hash, cfg)
	return cfg, nil
}
