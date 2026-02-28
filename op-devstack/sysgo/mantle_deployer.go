package sysgo

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/intentbuilder"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
)

var DefaultL1MNT = common.HexToAddress("0x8000000000000000000000000000000000000000")
var DefaultOperatorFeeVaultRecipient = common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9")

// WithL1MNT An alternative way to set the L1MNT. The other way is to use the WithDeployerOption.
// It requires we extend the L2Configurator interface to include WithL1MNT.
// Since MNT token address is a Mantle-only feature, directly modifying deployer pipeline seems cleaner.
func WithL1MNT(l1MNT common.Address) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting L1MNT", "l1MNT", l1MNT.Hex())
		for _, l2 := range intent.Chains {
			l2.L1MNT = l1MNT
		}
	}
}

func WithOperatorFeeVaultRecipient(recipient common.Address) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting OperatorFeeVaultRecipient", "recipient", recipient.Hex())
		for _, l2 := range intent.Chains {
			l2.OperatorFeeVaultRecipient = recipient
		}
	}
}

func WithGasPriceOracleTokenRatio(tokenRatio uint64) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting GasPriceOracleTokenRatio", "tokenRatio", tokenRatio)
		for _, l2 := range intent.Chains {
			l2.GasPriceOracleTokenRatio = tokenRatio
		}
	}
}

// WithMantlePortalPaused sets the OptimismPortalPaused override (defaults to true in Mantle artifacts).
func WithMantlePortalPaused(paused bool) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting OptimismPortalPaused", "paused", paused)
		if intent.GlobalDeployOverrides == nil {
			intent.GlobalDeployOverrides = make(map[string]any)
		}
		intent.GlobalDeployOverrides["OptimismPortalPaused"] = paused
	}
}

func WithScalarAndOverhead(scalar uint32, overhead uint64) DeployerPipelineOption {
	return func(wb *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting Scalar and Overhead", "scalar", scalar, "overhead", overhead)
		if intent.GlobalDeployOverrides == nil {
			intent.GlobalDeployOverrides = make(map[string]any)
		}
		intent.GlobalDeployOverrides["gasPriceOracleScalar"] = scalar
		intent.GlobalDeployOverrides["gasPriceOracleOverhead"] = overhead
	}
}

func WithGasLimit(gasLimit uint64) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting GasLimit", "gasLimit", gasLimit)
		for _, l2 := range intent.Chains {
			l2.GasLimit = gasLimit
		}
	}
}

func WithMantleForkAtGenesis(fork forks.MantleForkName) DeployerPipelineOption {
	return func(wb *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		wb.require.True(forks.IsValidMantleFork(fork), "invalid mantle fork: %s", string(fork))
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting MantleFork at genesis", "fork", string(fork))
		var opts []DeployerPipelineOption
		var future bool
		for _, refFork := range forks.AllMantleForks {
			if future {
				opts = append(opts, WithMantleForkAtOffset(refFork, nil))
			} else {
				opts = append(opts, WithMantleForkAtOffset(refFork, new(uint64)))
			}

			if refFork == fork {
				future = true
			}
		}

		for _, opt := range opts {
			opt(wb, intent, cfg)
		}
	}
}

func WithMantleForkAtOffset(fork forks.MantleForkName, offset *uint64) DeployerPipelineOption {
	return func(wb *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		wb.require.True(forks.IsValidMantleFork(fork), "invalid mantle fork: %s", string(fork))
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting MantleFork at offset", "fork", string(fork), "offset", offset)
		for _, l2 := range intent.Chains {
			key := fmt.Sprintf("l2Genesis%sTimeOffset", string(fork))
			if offset == nil {
				l2.DeployOverrides[key] = nil // Explicitly set to nil
			} else {
				// The typing is important, or op-deployer merge-JSON tricks will fail
				l2.DeployOverrides[key] = (*hexutil.Uint64)(offset)
			}
		}
	}
}

func WithMantleHardforkSequentialActivation(startFork, endFork forks.MantleForkName, delta uint64) DeployerPipelineOption {
	return func(wb *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting MantleHardforkSequentialActivation", "startFork", string(startFork), "endFork", string(endFork), "delta", delta)
		var opts []DeployerPipelineOption
		opts = append(opts, WithMantleForkAtGenesis(startFork))
		var activateWithOffset bool
		var deactivate bool
		var relativeIdx uint64
		for _, fork := range forks.AllMantleForks {
			if deactivate {
				opts = append(opts, WithMantleForkAtOffset(fork, nil))
				continue
			}
			if activateWithOffset {
				offset := delta * relativeIdx
				opts = append(opts, WithMantleForkAtOffset(fork, &offset))
				relativeIdx++
			}
			if fork == startFork {
				activateWithOffset = true
			}
			if fork == endFork {
				deactivate = true
			}
		}

		for _, opt := range opts {
			opt(wb, intent, cfg)
		}
	}
}

// WithMantleDeployer swaps in the Mantle-specific deployer pipeline.
func WithMantleDeployer() stack.Option[*Orchestrator] {
	return stack.FnOption[*Orchestrator]{
		BeforeDeployFn: func(o *Orchestrator) {
			o.P().Require().Nil(o.wb, "must not already have a world builder")
			o.wb = &worldBuilder{
				p:       o.P(),
				logger:  o.P().Logger(),
				require: o.P().Require(),
				keys:    o.keys,
				builder: intentbuilder.New(),
			}
		},
		DeployFn: func(o *Orchestrator) {
			o.P().Require().NotNil(o.wb, "must have a world builder")
			o.wb.deployerPipelineOptions = o.deployerPipelineOptions
			o.wb.BuildMantle()
		},
		AfterDeployFn: func(o *Orchestrator) {
			wb := o.wb
			require := o.P().Require()
			require.NotNil(o.wb, "must have a world builder")

			l1ID := stack.L1NetworkID(eth.ChainIDFromUInt64(wb.output.AppliedIntent.L1ChainID))
			superchainID := stack.SuperchainID("main")
			clusterID := stack.ClusterID("main")

			l1Net := &L1Network{
				id:        l1ID,
				genesis:   wb.outL1Genesis,
				blockTime: 6,
			}
			o.l1Nets.Set(l1ID.ChainID(), l1Net)

			o.superchains.Set(superchainID, &Superchain{
				id:         superchainID,
				deployment: wb.outSuperchainDeployment,
			})
			o.clusters.Set(clusterID, &Cluster{
				id:     clusterID,
				cfgset: wb.outFullCfgSet,
			})

			for _, chainID := range wb.l2Chains {
				l2Genesis, ok := wb.outL2Genesis[chainID]
				require.True(ok, "L2 genesis must exist")
				l2RollupCfg, ok := wb.outL2RollupCfg[chainID]
				require.True(ok, "L2 rollup config must exist")
				l2Dep, ok := wb.outL2Deployment[chainID]
				require.True(ok, "L2 deployment must exist")

				l2ID := stack.L2NetworkID(chainID)
				l2Net := &L2Network{
					id:         l2ID,
					l1ChainID:  l1ID.ChainID(),
					genesis:    l2Genesis,
					rollupCfg:  l2RollupCfg,
					deployment: l2Dep,
					keys:       o.keys,
				}
				o.l2Nets.Set(l2ID.ChainID(), l2Net)
			}
		},
	}
}

// BuildMantle runs the Mantle deployer pipeline and captures outputs.
func (wb *worldBuilder) BuildMantle() {
	st := &state.State{
		Version: 1,
	}

	// Work-around of op-deployer design issue.
	// We use the same deployer key for all L1 and L2 chains we deploy here.
	deployerKey, err := wb.keys.Secret(devkeys.DeployerRole.Key(big.NewInt(0)))
	wb.require.NoError(err, "need deployer key")

	intent, err := wb.builder.Build()
	wb.require.NoError(err)

	pipelineOpts := deployer.ApplyPipelineOpts{
		DeploymentTarget:   deployer.DeploymentTargetGenesis,
		L1RPCUrl:           "",
		DeployerPrivateKey: deployerKey,
		Intent:             intent,
		State:              st,
		Logger:             wb.logger,
		StateWriter:        wb, // direct output back here
	}
	for _, opt := range wb.deployerPipelineOptions {
		opt(wb, intent, &pipelineOpts)
	}

	err = deployer.MantleApplyPipeline(wb.p.Ctx(), pipelineOpts)
	wb.require.NoError(err)

	wb.require.NotNil(wb.output, "expected state-write to output")

	for _, id := range wb.output.Chains {
		chainID := eth.ChainIDFromBytes32(id.ID)
		wb.l2Chains = append(wb.l2Chains, chainID)
	}

	wb.buildL1Genesis()
	wb.buildMantleL2Genesis()
	wb.buildL2DeploymentOutputs()
	wb.buildFullConfigSet()
}

func (wb *worldBuilder) buildMantleL2Genesis() {
	wb.outL2Genesis = make(map[eth.ChainID]*core.Genesis)
	wb.outL2RollupCfg = make(map[eth.ChainID]*rollup.Config)
	for _, ch := range wb.output.Chains {
		l2Genesis, l2RollupCfg, err := inspect.MantleGenesisAndRollup(wb.output, ch.ID)
		wb.require.NoError(err, "need L2 genesis and rollup")
		id := eth.ChainIDFromBytes32(ch.ID)
		wb.outL2Genesis[id] = l2Genesis
		wb.outL2RollupCfg[id] = l2RollupCfg
	}
}
