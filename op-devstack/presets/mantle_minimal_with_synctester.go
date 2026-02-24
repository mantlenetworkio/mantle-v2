package presets

import (
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type MantleMinimalWithSyncTester struct {
	MantleMinimal

	SyncTester *dsl.SyncTester
}

func WithMantleMinimalWithSyncTester(fcu eth.FCUState) stack.CommonOption {
	return stack.MakeCommon(sysgo.DefaultMantleMinimalSystemWithSyncTester(&sysgo.DefaultMinimalSystemWithSyncTesterIDs{}, fcu))
}

func NewMantleMinimalWithSyncTester(t devtest.T) *MantleMinimalWithSyncTester {
	system := shim.NewSystem(t)
	orch := Orchestrator()
	orch.Hydrate(system)
	minimal := mantleMinimalFromSystem(t, system, orch)
	l2 := system.L2Network(match.Assume(t, match.L2ChainA))
	syncTester := l2.SyncTester(match.Assume(t, match.FirstSyncTester))
	return &MantleMinimalWithSyncTester{
		MantleMinimal: *minimal,
		SyncTester:    dsl.NewSyncTester(syncTester),
	}
}
