package sysgo

import (
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
)

type ControlPlane struct {
	o *Orchestrator
}

func control(lifecycle stack.Lifecycle, mode stack.ControlAction) {
	switch mode {
	case stack.Start:
		lifecycle.Start()
	case stack.Stop:
		lifecycle.Stop()
	}
}

func (c *ControlPlane) SupervisorState(id stack.SupervisorID, mode stack.ControlAction) {
	s, ok := c.o.supervisors.Get(id)
	c.o.P().Require().True(ok, "need supervisor to change state")
	control(s, mode)
}

func (c *ControlPlane) L2CLNodeState(id stack.L2CLNodeID, mode stack.ControlAction) {
	s, ok := c.o.l2CLs.Get(id)
	c.o.P().Require().True(ok, "need l2cl node to change state")
	control(s, mode)
}

func (c *ControlPlane) L2ELNodeState(id stack.L2ELNodeID, mode stack.ControlAction) {
	s, ok := c.o.l2ELs.Get(id)
	c.o.P().Require().True(ok, "need l2el node to change state")
	control(s, mode)
}

// L2ELNodeWipe wipes the data directory of the given L2 EL node, resetting it to genesis.
// Only op-reth nodes persist chain data to disk and can be wiped; for other node types
// (e.g. op-geth which uses an in-memory DB) this is a no-op with a warning.
func (c *ControlPlane) L2ELNodeWipe(id stack.L2ELNodeID) {
	s, ok := c.o.l2ELs.Get(id)
	c.o.P().Require().True(ok, "need l2el node to wipe")
	reth, ok := s.(*OpReth)
	if !ok {
		c.o.P().Logger().Warn("L2ELNodeWipe called on non-reth node; no persistent data to wipe, skipping", "id", id)
		return
	}
	reth.Wipe()
}

var _ stack.ControlPlane = (*ControlPlane)(nil)
var _ stack.WipeableControlPlane = (*ControlPlane)(nil)

func (c *ControlPlane) FakePoSState(id stack.L1CLNodeID, mode stack.ControlAction) {
	s, ok := c.o.l1CLs.Get(id)
	c.o.P().Require().True(ok, "need l1cl node to change state of fakePoS module")

	control(s.fakepos, mode)
}

func (c *ControlPlane) OPRBuilderNodeState(id stack.OPRBuilderNodeID, mode stack.ControlAction) {
	s, ok := c.o.oprbuilderNodes.Get(id)
	c.o.P().Require().True(ok, "need oprbuilder node to change state")
	control(s, mode)
}

func (c *ControlPlane) RollupBoostNodeState(id stack.RollupBoostNodeID, mode stack.ControlAction) {
	s, ok := c.o.rollupBoosts.Get(id)
	c.o.P().Require().True(ok, "need rollup boost node to change state")
	control(s, mode)
}
