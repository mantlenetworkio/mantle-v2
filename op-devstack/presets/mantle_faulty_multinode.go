// Package presets — fault-injection preset for the Engine API
// elfaultinjector proxy.
//
// MantleFaultyMultiNode extends MantleSingleChainMultiNode (1 sequencer +
// 1 verifier on a single L2 chain) by globally enabling the Engine API
// fault-injection proxy in sysgo's L2EL wiring. Every L2 EL gets an
// elfaultinjector.Proxy chained between its op-node-facing auth-RPC proxy
// and the underlying op-geth auth-RPC. The proxies start in inactive
// (pure pass-through) mode; the test calls Activate(rule) at runtime
// to synthesize INVALID PayloadStatusV1 responses for matching
// engine_newPayloadV{3,4} requests.
//
// This preset is sysgo-only because the fault-injection wiring is a
// sysgo-internal concept (see op-devstack/sysgo/l2_el_opgeth.go).
//
// Use this preset for tests that need to reproduce execution-layer
// divergence between an accepting EL and a rejecting EL — e.g. the
// op-conductor split-brain at unsafe head case study described in
// op-conductor/INTEGRATION.md (mitigation 2/3/4 regression target).
package presets

import (
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-service/testutils/elfaultinjector"
)

// MantleFaultyMultiNode is the hydrated test-side handle for the
// faulty-multinode preset. It exposes the underlying multi-node system
// plus a map of per-L2EL fault-injection proxies, keyed by L2ELNodeID.
type MantleFaultyMultiNode struct {
	*MantleSingleChainMultiNode
	// EngineFaultInjectors is keyed by L2ELNodeID and contains a non-nil
	// proxy for every *sysgo.OpGeth in the system. Tests look up the
	// verifier's injector via L2ELB.ID() (or the sequencer's via L2EL.ID())
	// and call Activate(rule) / Deactivate() at runtime.
	EngineFaultInjectors map[stack.L2ELNodeID]*elfaultinjector.Proxy
}

// WithMantleFaultyMultiNode wires the fault-injection proxy onto every
// L2 EL of a Mantle single-chain multinode (sequencer + verifier) system.
// Tests should pair this with NewMantleFaultyMultiNode(t).
//
// Fault rules are inactive at startup; the test must call
// EngineFaultInjectors[id].Activate(rule) to enable injection on a
// specific node.
func WithMantleFaultyMultiNode() stack.CommonOption {
	return stack.Combine(
		WithMantleSingleChainMultiNode(),
		// SysGo only: kurtosis/persistent backends would need their own
		// proxy plumbing.
		WithCompatibleTypes(compat.SysGo),
		stack.MakeCommon(sysgo.WithGlobalL2ELOption(sysgo.L2ELWithEngineFaultInjector())),
	)
}

// NewMantleFaultyMultiNode hydrates the multi-node system and harvests
// the fault-injection proxies attached to every *sysgo.OpGeth instance.
func NewMantleFaultyMultiNode(t devtest.T) *MantleFaultyMultiNode {
	base := NewMantleSingleChainMultiNode(t)

	orch, ok := Orchestrator().(*sysgo.Orchestrator)
	t.Require().True(ok, "MantleFaultyMultiNode requires the sysgo orchestrator backend")

	injectors := make(map[stack.L2ELNodeID]*elfaultinjector.Proxy)
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

	return &MantleFaultyMultiNode{
		MantleSingleChainMultiNode: base,
		EngineFaultInjectors:       injectors,
	}
}
