package sync_tester_hfs

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
)

func TestMain(m *testing.M) {
	presets.DoMain(m, presets.WithMantleSimpleWithSyncTester(),
		presets.WithCompatibleTypes(compat.SysGo),
		presets.WithMantleHardforkSequentialActivation(forks.MantleSkadi, forks.MantleArsia, 6),
		presets.WithNoDiscovery(),
		stack.Combine(
			stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithScalarAndOverhead(1368, 1000000))),
			stack.MakeCommon(sysgo.WithDeployerPipelineOption(sysgo.WithGasLimit(1125899906842624))),
			presets.WithMantleLegacyBatcher(),
		))
}
