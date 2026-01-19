package node

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-node/config"
	"github.com/ethereum-optimism/optimism/op-node/p2p"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

func (n *OpNode) initRPCSync(ctx context.Context, cfg *config.Config) error {
	rpcSyncClient, rpcCfg, err := cfg.L2Sync.Setup(ctx, n.log, &cfg.Rollup)
	if err != nil {
		return fmt.Errorf("failed to setup L2 execution-engine RPC client for backup sync: %w", err)
	}
	if rpcSyncClient == nil { // if no RPC client is configured to sync from, then don't add the RPC sync client
		return nil
	}
	rec := p2p.NewBlockReceiver(n.log, n.metrics, n.l2Driver.SyncDeriver, n.cfg.Tracer)
	syncClient, err := sources.NewSyncClient(rec.OnUnsafeL2Payload, rpcSyncClient, n.log, n.metrics.L2SourceCache, rpcCfg)
	if err != nil {
		return fmt.Errorf("failed to create sync client: %w", err)
	}

	if err := syncClient.Start(); err != nil {
		n.log.Error("Could not start the backup sync client", "err", err)
		return err
	}
	n.log.Info("Started L2-RPC sync service")
	n.rpcSync = syncClient
	return nil
}
