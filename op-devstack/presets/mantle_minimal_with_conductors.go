package presets

import (
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-service/testutils/elfaultinjector"
)

// MantleMinimalWithConductors mirrors WithMinimalWithConductors but seeds a Mantle minimal system.
//
// On the sysgo backend, this dispatches to DefaultMantleConductorSystem, which
// composes the 3-sequencer-3-conductor topology used to reproduce the
// op-conductor split-brain at unsafe head case study (closes #16418).
// On Persistent/Kurtosis backends, the sysgo-specific option is a no-op and
// conductor topology is provided by the external orchestration layer.
func WithMantleMinimalWithConductors() stack.CommonOption {
	return stack.Combine(
		stack.MakeCommon(sysgo.DefaultMantleConductorSystem(&sysgo.DefaultMantleConductorSystemIDs{})),
		WithCompatibleTypes(
			compat.Persistent,
			compat.Kurtosis,
			compat.SysGo,
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

// WithMantleMinimalWithFaultyConductors layers the sysgo Engine API
// fault-injection proxy onto the Mantle minimal-with-conductors topology.
//
// On sysgo, WithMantleMinimalWithConductors composes a 3-sequencer-3-conductor
// topology (DefaultMantleConductorSystem) and L2ELWithEngineFaultInjector wraps
// each L2EL's auth RPC with an Engine-API fault injection proxy. On
// Persistent/Kurtosis the sysgo-specific options are no-ops and
// EngineFaultInjectors will be empty; tests that rely on injection should skip
// when len(EngineFaultInjectors) < 2.
func WithMantleMinimalWithFaultyConductors() stack.CommonOption {
	return stack.Combine(
		WithMantleMinimalWithConductors(),
		stack.MakeCommon(sysgo.WithGlobalL2ELOption(sysgo.L2ELWithEngineFaultInjector())),
	)
}

// MantleMinimalWithFaultyConductors extends MantleMinimalWithConductors with
// per-L2EL Engine API fault-injection proxies, keyed by L2ELNodeID. The map
// is non-nil but may be empty if the active backend doesn't expose injectors
// (see WithMantleMinimalWithFaultyConductors for current backend status).
//
// SysgoConductors exposes the in-process *sysgo.Conductor wrappers when the
// active backend is sysgo. This lets tests reach Go-only conductor APIs that
// upstream does not register on the public RPC interface — most importantly
// LatestUnsafePayload, which reads the raft FSM directly (the public RPC
// only exposes CommitUnsafePayload). On non-sysgo backends the map is
// non-nil but empty; tests should fall back to indirect FSM-health proofs
// (cluster membership, leadership stability) in that case.
type MantleMinimalWithFaultyConductors struct {
	*MantleMinimalWithConductors
	EngineFaultInjectors map[stack.L2ELNodeID]*elfaultinjector.Proxy
	SysgoConductors      map[stack.ConductorID]*sysgo.Conductor
}

// NewMantleMinimalWithFaultyConductors hydrates the conductor system and
// harvests EL fault injectors from the sysgo orchestrator when available.
// On non-sysgo orchestrators, EngineFaultInjectors is populated as an empty
// (non-nil) map.
func NewMantleMinimalWithFaultyConductors(t devtest.T) *MantleMinimalWithFaultyConductors {
	base := NewMantleMinimalWithConductors(t)

	injectors := make(map[stack.L2ELNodeID]*elfaultinjector.Proxy)
	sysgoConductors := make(map[stack.ConductorID]*sysgo.Conductor)
	if orch, ok := Orchestrator().(*sysgo.Orchestrator); ok {
		orch.RangeL2ELs(func(id stack.L2ELNodeID, n sysgo.L2ELNode) bool {
			g, isOpGeth := n.(*sysgo.OpGeth)
			if !isOpGeth {
				return true
			}
			if inj := g.EngineFaultInjector(); inj != nil {
				injectors[id] = inj
			}
			return true
		})
		orch.RangeConductors(func(id stack.ConductorID, c *sysgo.Conductor) bool {
			sysgoConductors[id] = c
			return true
		})
	}

	return &MantleMinimalWithFaultyConductors{
		MantleMinimalWithConductors: base,
		EngineFaultInjectors:        injectors,
		SysgoConductors:             sysgoConductors,
	}
}
