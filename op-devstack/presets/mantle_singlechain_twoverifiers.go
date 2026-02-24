package presets

import (
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

type MantleSingleChainTwoVerifiers struct {
	MantleSingleChainMultiNode

	L2ELC *dsl.L2ELNode
	L2CLC *dsl.L2CLNode
}

func WithMantleSingleChainTwoVerifiers() stack.CommonOption {
	return stack.MakeCommon(sysgo.DefaultMantleSingleChainTwoVerifiersSystem(&sysgo.DefaultMantleSingleChainTwoVerifiersSystemIDs{}))
}

func NewMantleSingleChainTwoVerifiers(t devtest.T) *MantleSingleChainTwoVerifiers {
	system := shim.NewSystem(t)
	orch := Orchestrator()
	orch.Hydrate(system)
	multi := mantleMinimalFromSystem(t, system, orch)
	l2 := system.L2Network(match.Assume(t, match.L2ChainA))

	verifierCL := l2.L2CLNode(match.Assume(t,
		match.And(
			match.Not(match.WithSequencerActive(t.Ctx())),
			match.Not[stack.L2CLNodeID, stack.L2CLNode](multi.L2CL.ID()),
		)))
	verifierEL := l2.L2ELNode(match.Assume(t,
		match.And(
			match.EngineFor(verifierCL),
			match.Not[stack.L2ELNodeID, stack.L2ELNode](multi.L2EL.ID()))))

	verifierCL2 := l2.L2CLNode(match.Assume(t,
		match.And(
			match.Not(match.WithSequencerActive(t.Ctx())),
			match.Not[stack.L2CLNodeID, stack.L2CLNode](multi.L2CL.ID()),
			match.Not[stack.L2CLNodeID, stack.L2CLNode](verifierCL.ID()),
		)))
	verifierEL2 := l2.L2ELNode(match.Assume(t,
		match.And(
			match.EngineFor(verifierCL2),
			match.Not[stack.L2ELNodeID, stack.L2ELNode](multi.L2EL.ID()),
			match.Not[stack.L2ELNodeID, stack.L2ELNode](verifierEL.ID()),
		)))

	preset := &MantleSingleChainTwoVerifiers{
		MantleSingleChainMultiNode: *NewMantleSingleChainMultiNodeWithoutCheck(t),
		L2ELC:                      dsl.NewL2ELNode(verifierEL2, orch.ControlPlane()),
		L2CLC:                      dsl.NewL2CLNode(verifierCL2, orch.ControlPlane()),
	}

	dsl.CheckAll(t,
		preset.L2CLB.MatchedFn(preset.L2CL, types.CrossSafe, 30),
		preset.L2CLB.MatchedFn(preset.L2CL, types.LocalUnsafe, 30),
		preset.L2CLC.MatchedFn(preset.L2CL, types.CrossSafe, 30),
		preset.L2CLC.MatchedFn(preset.L2CL, types.LocalUnsafe, 30),
	)
	return preset
}
