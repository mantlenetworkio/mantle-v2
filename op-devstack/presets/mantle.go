package presets

import (
	bss "github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
)

func WithMantleSkadiAtGenesis() stack.CommonOption {
	return stack.Combine(
		stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithMantleForkAtGenesis(forks.MantleSkadi))),
		stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithGasLimit(1125899906842624))),
		// Skadi uses the legacy batcher which uses singular batches and zlib compression
		WithMantleLegacyBatcher(),
	)
}

func WithMantleLimbAtGenesis() stack.CommonOption {
	return stack.Combine(
		stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithMantleForkAtGenesis(forks.MantleLimb))),
		// Limb uses legacy gas limit and scalar/overhead
		stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithScalarAndOverhead(1368, 1000000))),
		stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithGasLimit(1125899906842624))),
		// Limb uses the legacy batcher which uses singular batches and zlib compression
		WithMantleLegacyBatcher(),
	)
}

func WithMantleLegacyBatcher() stack.CommonOption {
	return stack.MakeCommon(sysgo.WithBatcherOption(func(_ stack.L2BatcherID, cfg *bss.CLIConfig) {
		cfg.BatchType = derive.SingularBatchType
		cfg.CompressionAlgo = derive.Zlib
	}))
}

func WithMantleArsiaAtGenesis() stack.CommonOption {
	return stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithMantleForkAtGenesis(forks.MantleArsia)))
}
