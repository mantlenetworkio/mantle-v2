package opreth

import (
	"context"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/l2backend"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/reth"
	"github.com/ethereum-optimism/optimism/op-e2e/enginetest"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// NewRethEngine creates an OpEngine backed by an external op-reth process.
//
// It requires OP_E2E_L2_BIN to point to the op-reth binary; otherwise the
// test is automatically skipped via l2backend.RethBinPath.
func NewRethEngine(t testing.TB, ctx context.Context, cfg *e2esys.SystemConfig) (*enginetest.OpEngine, error) {
	t.Helper()

	binPath := l2backend.RethBinPath(t) // skips if OP_E2E_L2_BIN is not set

	l1Genesis, err := genesis.BuildL1DeveloperGenesis(cfg.DeployConfig, config.L1Allocs(config.DefaultAllocType), config.L1Deployments(config.DefaultAllocType))
	require.NoError(t, err)
	l1Block := l1Genesis.ToBlock()
	allocsMode := e2eutils.GetL2AllocsMode(cfg.DeployConfig, l1Block.Time())
	l2Allocs := config.L2Allocs(config.DefaultAllocType, allocsMode)
	l2Genesis, err := genesis.BuildL2Genesis(cfg.DeployConfig, l2Allocs, eth.BlockRefFromHeader(l1Block.Header()))
	require.NoError(t, err)
	l2GenesisBlock := l2Genesis.ToBlock()

	rollupGenesis := rollup.Genesis{
		L1:           eth.BlockID{Hash: l1Block.Hash(), Number: l1Block.NumberU64()},
		L2:           eth.BlockID{Hash: l2GenesisBlock.Hash(), Number: l2GenesisBlock.NumberU64()},
		L2Time:       l2GenesisBlock.Time(),
		SystemConfig: e2eutils.SystemConfigFromDeployConfig(cfg.DeployConfig),
	}

	rethNode, err := reth.InitReth(t, "l2", l2Genesis, cfg.JWTFilePath, binPath)
	require.NoError(t, err)
	require.NoError(t, rethNode.Start())

	rollupCfg, err := cfg.DeployConfig.RollupConfig(eth.BlockRefFromHeader(l1Block.Header()), l2GenesisBlock.Hash(), l2GenesisBlock.NumberU64())
	require.NoError(t, err)
	rollupCfg.Genesis = rollupGenesis

	// Note: Using CanyonTime here because for OP Stack chains, Shanghai must be activated at the same time as Canyon.
	chainCfg := params.ChainConfig{
		CanyonTime: cfg.DeployConfig.CanyonTime(l2GenesisBlock.Time()),
	}
	genesisPayload, err := eth.BlockAsPayload(l2GenesisBlock, &chainCfg)
	require.NoError(t, err)

	return enginetest.NewOpEngine(t, ctx, cfg, enginetest.OpEngineConfig{
		Node:           rethNode,
		RollupCfg:      rollupCfg,
		RollupGenesis:  rollupGenesis,
		L1ChainConfig:  l1Genesis.Config,
		L2ChainConfig:  l2Genesis.Config,
		L1Head:         eth.BlockToInfo(l1Block),
		GenesisPayload: genesisPayload,
	})
}
