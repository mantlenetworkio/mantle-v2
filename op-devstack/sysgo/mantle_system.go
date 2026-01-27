package sysgo

import (
	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	ps "github.com/ethereum-optimism/optimism/op-proposer/proposer"
	"github.com/ethereum/go-ethereum/common"
)

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
		WithDeployerPipelineOption(WithMantlePortalPaused(false)),
		WithDeployerPipelineOption(WithMantleForkAtGenesis(forks.MantleArsia)),
	)

	opt.Add(WithL1Nodes(ids.L1EL, ids.L1CL))

	opt.Add(WithL2ELNode(ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL, ids.L1CL, ids.L1EL, ids.L2EL, L2CLSequencer()))

	opt.Add(WithBatcher(ids.L2Batcher, ids.L1EL, ids.L2CL, ids.L2EL))
	opt.Add(WithLegacyProposer(ids.L2Proposer, ids.L1EL, &ids.L2CL, nil))

	opt.Add(WithFaucets([]stack.L1ELNodeID{ids.L1EL}, []stack.L2ELNodeID{ids.L2EL}))

	opt.Add(WithTestSequencer(ids.TestSequencer, ids.L1CL, ids.L2CL, ids.L1EL, ids.L2EL))

	opt.Add(WithL2MetricsDashboard())

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = *ids
	}))

	return opt
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
		// turn off DGF by setting relating contract address to empty string
		cfg.DGFAddress = ""
	})
}

func (d *L2Deployment) L2OOAddress() common.Address {
	return d.l2OOAddress
}
