package config

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/log"
)

type L2SyncEndpointSetup interface {
	// Setup a RPC client to another L2 node to sync L2 blocks from.
	// It may return a nil client with nil error if RPC based sync is not enabled.
	Setup(ctx context.Context, log log.Logger, rollupCfg *rollup.Config) (cl client.RPC, rpcCfg *sources.SyncClientConfig, err error)
	Check() error
}

// L2SyncEndpointConfig contains configuration for the fallback sync endpoint
type L2SyncEndpointConfig struct {
	Enabled bool
	// Address of the L2 RPC to use for backup sync, may be empty if RPC alt-sync is disabled.
	L2NodeAddr string
	TrustRPC   bool
}

var _ L2SyncEndpointSetup = (*L2SyncEndpointConfig)(nil)

// Setup creates an RPC client to sync from.
// It will return nil without error if no sync method is configured.
func (cfg *L2SyncEndpointConfig) Setup(ctx context.Context, log log.Logger, rollupCfg *rollup.Config) (client.RPC, *sources.SyncClientConfig, error) {
	if !cfg.Enabled {
		return nil, nil, nil
	}
	if cfg.L2NodeAddr == "" {
		return nil, nil, nil
	}
	l2Node, err := client.NewRPC(ctx, log, cfg.L2NodeAddr)
	if err != nil {
		return nil, nil, err
	}

	return l2Node, sources.SyncClientDefaultConfig(rollupCfg, cfg.TrustRPC), nil
}

func (cfg *L2SyncEndpointConfig) Check() error {
	// empty addr is valid, as it is optional.
	return nil
}

type PreparedL2SyncEndpoint struct {
	// RPC endpoint to use for syncing, may be nil if RPC alt-sync is disabled.
	Client   client.RPC
	TrustRPC bool
}

var _ L2SyncEndpointSetup = (*PreparedL2SyncEndpoint)(nil)

func (cfg *PreparedL2SyncEndpoint) Setup(ctx context.Context, log log.Logger, rollupCfg *rollup.Config) (client.RPC, *sources.SyncClientConfig, error) {
	return cfg.Client, sources.SyncClientDefaultConfig(rollupCfg, cfg.TrustRPC), nil
}

func (cfg *PreparedL2SyncEndpoint) Check() error {
	return nil
}
