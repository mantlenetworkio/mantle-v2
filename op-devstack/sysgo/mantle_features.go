package sysgo

import (
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum/go-ethereum/common"
)

var DefaultL1MNT = common.HexToAddress("0x8000000000000000000000000000000000000000")
var DefaultOperatorFeeVaultRecipient = common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9")

func DefaultMantleMinimalSystem(dest *DefaultMinimalSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMinimalSystemIDs(DefaultL1ID, DefaultL2AID)
	opt := defaultMinimalSystemOpts(&ids, dest)
	opt.Add(WithDeployerPipelineOption(WithMantleL2()))
	return opt
}

func WithMantleL2() DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		// An alternative way to set the L1MNT and OperatorFeeVaultRecipient is to use the WithDeployerOption.
		// It requires we extend the L2Configurator interface to include WithL1MNT and WithOperatorFeeVaultRecipient.
		// Since MNT token address is a Mantle-only feature, directly modifying deployer pipeline seems cleaner.
		cfg.Logger.New("stage", "with-mantle-l2").Info("Setting L1MNT", "L1MNT", DefaultL1MNT)
		cfg.Logger.New("stage", "with-mantle-l2").Info("Setting OperatorFeeVaultRecipient", "OperatorFeeVaultRecipient", DefaultOperatorFeeVaultRecipient)
		for _, l2 := range intent.Chains {
			l2.L1MNT = DefaultL1MNT
			l2.OperatorFeeVaultRecipient = DefaultOperatorFeeVaultRecipient
		}
	}
}
