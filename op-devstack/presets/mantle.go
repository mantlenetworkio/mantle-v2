package presets

import (
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
)

func WithMantleMinimal() stack.CommonOption {
	return stack.MakeCommon(sysgo.DefaultMantleMinimalSystem(&sysgo.DefaultMinimalSystemIDs{}))
}

func WithMantleArsiaAtGenesis() stack.CommonOption {
	return stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithMantleForkAtGenesis(forks.MantleArsia)))
}
