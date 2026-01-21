package presets

import (
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
)

type MantleSimpleWithSyncTester struct {
	MantleMinimal

	SyncTester *dsl.SyncTester
}

func WithMantleSimpleWithSyncTester() stack.CommonOption {
	return stack.MakeCommon(sysgo.DefaultMantleSimpleSystemWithSyncTester(&sysgo.DefaultSimpleSystemWithSyncTesterIDs{}))
}

func NewMantleSimpleWithSyncTester(t devtest.T) *MantleSimpleWithSyncTester {
	system := shim.NewSystem(t)
	orch := Orchestrator()
	orch.Hydrate(system)
	minimal := mantleMinimalFromSystem(t, system, orch)
	l2 := system.L2Network(match.Assume(t, match.L2ChainA))
	syncTester := l2.SyncTester(match.Assume(t, match.FirstSyncTester))
	return &MantleSimpleWithSyncTester{
		MantleMinimal: *minimal,
		SyncTester:    dsl.NewSyncTester(syncTester),
	}
}
