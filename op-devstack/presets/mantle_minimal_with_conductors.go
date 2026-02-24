package presets

import (
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
)

// MantleMinimalWithConductors mirrors WithMinimalWithConductors but seeds a Mantle minimal system.
func WithMantleMinimalWithConductors() stack.CommonOption {
	return stack.Combine(
		WithMantleMinimal(),
		// TODO(#16418) add sysgo conductor support; keep compatibility gate aligned with OP path
		WithCompatibleTypes(
			compat.Persistent,
			compat.Kurtosis,
		),
	)
}

type MantleMinimalWithConductors struct {
	*MantleMinimal
	ConductorSets map[stack.L2NetworkID]dsl.ConductorSet
}

// NewMantleMinimalWithConductors hydrates Mantle minimal and exposes conductor sets.
func NewMantleMinimalWithConductors(t devtest.T) *MantleMinimalWithConductors {
	system := shim.NewSystem(t)
	orch := Orchestrator()
	orch.Hydrate(system)
	chains := system.L2Networks()
	conductorSets := make(map[stack.L2NetworkID]dsl.ConductorSet)
	for _, chain := range chains {
		chainMatcher := match.L2ChainById(chain.ID())
		l2 := system.L2Network(match.Assume(t, chainMatcher))

		conductorSets[chain.ID()] = dsl.NewConductorSet(l2.Conductors())
	}
	return &MantleMinimalWithConductors{
		MantleMinimal: NewMantleMinimal(t),
		ConductorSets: conductorSets,
	}
}
