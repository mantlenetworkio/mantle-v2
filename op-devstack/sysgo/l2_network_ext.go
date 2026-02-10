package sysgo

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/backend/depset"
	"github.com/ethereum/go-ethereum/params"
)

// WithL2NetworkFromExtConfig creates an L2 network using the rollup config and chain config
// provided directly in the ExtNetworkConfig, bypassing the superchain registry.
// This is useful for networks that are not in the superchain registry (e.g. Mantle).
func WithL2NetworkFromExtConfig(l2NetworkID stack.L2NetworkID, rollupCfg *rollup.Config, l2ChainConfig *params.ChainConfig) stack.Option[*Orchestrator] {
	return stack.BeforeDeploy(func(orch *Orchestrator) {
		genesis := &core.Genesis{
			Config: l2ChainConfig,
		}

		l2Net := &L2Network{
			id:        l2NetworkID,
			l1ChainID: eth.ChainIDFromBig(rollupCfg.L1ChainID),
			genesis:   genesis,
			rollupCfg: rollupCfg,
			keys:      orch.keys,
		}

		require := orch.P().WithCtx(stack.ContextWithID(orch.P().Ctx(), l2NetworkID)).Require()
		require.True(orch.l2Nets.SetIfMissing(l2NetworkID.ChainID(), l2Net),
			fmt.Sprintf("must not already exist: %s", l2NetworkID))
	})
}

// WithEmptyDepSetFromExtConfig creates an L2 network using configs provided directly
// (not from the superchain registry) and sets up an empty dependency set.
// This is useful for networks that are not in the superchain registry (e.g. Mantle).
func WithEmptyDepSetFromExtConfig(l2NetworkID stack.L2NetworkID, networkName string, rollupCfg *rollup.Config, l2ChainConfig *params.ChainConfig) stack.Option[*Orchestrator] {
	return stack.Combine(
		WithL2NetworkFromExtConfig(l2NetworkID, rollupCfg, l2ChainConfig),
		stack.BeforeDeploy(func(orch *Orchestrator) {
			clusterID := stack.ClusterID(networkName)
			cluster := &Cluster{
				id:     clusterID,
				cfgset: depset.FullConfigSetMerged{},
			}

			orch.clusters.Set(clusterID, cluster)
		}),
	)
}
