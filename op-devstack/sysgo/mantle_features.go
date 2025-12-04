package sysgo

import (
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum/go-ethereum/common"
)

var DefaultL1MNT = common.HexToAddress("0x8000000000000000000000000000000000000000")

func DefaultMantleMinimalSystem(dest *DefaultMinimalSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMinimalSystemIDs(DefaultL1ID, DefaultL2AID)
	opt := defaultMinimalSystemOpts(&ids, dest)
	opt.Add(WithDeployerPipelineOption(WithMantleL2()))
	return opt
}

func WithMantleL2() DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "with-mantle-l2").Info("Setting L1MNT", "L1MNT", DefaultL1MNT)
		for _, l2 := range intent.Chains {
			l2.L1MNT = DefaultL1MNT
		}
	}
}
