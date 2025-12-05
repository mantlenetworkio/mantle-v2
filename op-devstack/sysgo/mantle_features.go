package sysgo

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/intentbuilder"
	ps "github.com/ethereum-optimism/optimism/op-proposer/proposer"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
)

var DefaultL1MNT = common.HexToAddress("0x8000000000000000000000000000000000000000")
var DefaultOperatorFeeVaultRecipient = common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9")

func DefaultMantleMinimalSystem(dest *DefaultMinimalSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMinimalSystemIDs(DefaultL1ID, DefaultL2AID)
	return defaultMantleMinimalSystemOpts(&ids, dest)
}

func defaultMantleMinimalSystemOpts(ids *DefaultMinimalSystemIDs, dest *DefaultMinimalSystemIDs) stack.CombinedOption[*Orchestrator] {
	opt := stack.Combine[*Orchestrator]()
	opt.Add(stack.BeforeDeploy(func(o *Orchestrator) {
		o.P().Logger().Info("Setting up")
	}))

	opt.Add(WithMnemonicKeys(devkeys.TestMnemonic))

	opt.Add(WithMantleDeployer(),
		WithDeployerOptions(
			WithLocalContractSources(),
			WithCommons(ids.L1.ChainID()),
			WithPrefundedL2(ids.L1.ChainID(), ids.L2.ChainID()),
		),
		WithDeployerPipelineOption(WithL1MNT(DefaultL1MNT)),
		WithDeployerPipelineOption(WithOperatorFeeVaultRecipient(DefaultOperatorFeeVaultRecipient)),
	)

	opt.Add(WithL1Nodes(ids.L1EL, ids.L1CL))

	opt.Add(WithL2ELNode(ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL, ids.L1CL, ids.L1EL, ids.L2EL, L2CLSequencer()))

	opt.Add(WithBatcher(ids.L2Batcher, ids.L1EL, ids.L2CL, ids.L2EL))
	opt.Add(WithLegacyProposer(ids.L2Proposer, ids.L1EL, &ids.L2CL, nil))

	opt.Add(WithFaucets([]stack.L1ELNodeID{ids.L1EL}, []stack.L2ELNodeID{ids.L2EL}))

	opt.Add(WithTestSequencer(ids.TestSequencer, ids.L1CL, ids.L2CL, ids.L1EL, ids.L2EL))

	// opt.Add(WithL2Challenger(ids.L2Challenger, ids.L1EL, ids.L1CL, nil, nil, &ids.L2CL, []stack.L2ELNodeID{
	// 	ids.L2EL,
	// }))

	opt.Add(WithL2MetricsDashboard())

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = *ids
	}))

	return opt
}

// An alternative way to set the L1MNT and OperatorFeeVaultRecipient is to use the WithDeployerOption.
// It requires we extend the L2Configurator interface to include WithL1MNT and WithOperatorFeeVaultRecipient.
// Since MNT token address is a Mantle-only feature, directly modifying deployer pipeline seems cleaner.
func WithL1MNT(l1MNT common.Address) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-l1-mnt").Info("L1MNT", l1MNT.Hex())
		cfg.Logger.New("stage", "with-mantle-l2").Info("Setting OperatorFeeVaultRecipient", "OperatorFeeVaultRecipient", DefaultOperatorFeeVaultRecipient)
		for _, l2 := range intent.Chains {
			l2.L1MNT = l1MNT
			l2.OperatorFeeVaultRecipient = DefaultOperatorFeeVaultRecipient
		}
	}
}

func WithOperatorFeeVaultRecipient(recipient common.Address) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-operator-fee-vault-recipient").Info("OperatorFeeVaultRecipient", recipient.Hex())
		for _, l2 := range intent.Chains {
			l2.OperatorFeeVaultRecipient = recipient
		}
	}
}

func WithLegacyProposer(proposerID stack.L2ProposerID, l1ELID stack.L1ELNodeID,
	l2CLID *stack.L2CLNodeID, supervisorID *stack.SupervisorID) stack.Option[*Orchestrator] {
	return stack.AfterDeploy(func(orch *Orchestrator) {
		WithLegacyProposerOption(orch, proposerID)
		WithProposerPostDeploy(orch, proposerID, l1ELID, l2CLID, supervisorID)
	})
}

func WithLegacyProposerOption(orch *Orchestrator, proposerID stack.L2ProposerID) {
	orch.proposerOptions = append(orch.proposerOptions, func(id stack.L2ProposerID, cfg *ps.CLIConfig) {
		ctx := orch.P().Ctx()
		ctx = stack.ContextWithID(ctx, proposerID)
		p := orch.P().WithCtx(ctx)

		require := p.Require()
		l2Net, ok := orch.l2Nets.Get(proposerID.ChainID())
		require.True(ok)
		cfg.L2OOAddress = l2Net.deployment.L2OOAddress().Hex()
		fmt.Println("dgf address", cfg.DGFAddress)
		cfg.DGFAddress = ""
	})
}

/////////////////////////////////////////////////////////////
// Deployer
/////////////////////////////////////////////////////////////

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

/////////////////////////////////////////////////////////////
// world builder
/////////////////////////////////////////////////////////////

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

	// TODO-ARSIA add a builder option or new a mantle world builder
	err = deployer.MantleApplyPipeline(wb.p.Ctx(), pipelineOpts)
	wb.require.NoError(err)

	wb.require.NotNil(wb.output, "expected state-write to output")

	for _, id := range wb.output.Chains {
		chainID := eth.ChainIDFromBytes32(id.ID)
		wb.l2Chains = append(wb.l2Chains, chainID)
	}

	wb.buildL1Genesis()
	wb.buildL2Genesis()
	wb.buildL2DeploymentOutputs()
	wb.buildFullConfigSet()
}
