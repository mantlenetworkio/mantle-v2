package sysgo

import (
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type DefaultMantleSingleChainTwoVerifiersSystemIDs struct {
	DefaultMantleSingleChainMultiNodeSystemIDs

	L2CLC stack.L2CLNodeID
	L2ELC stack.L2ELNodeID
}

func NewDefaultMantleSingleChainTwoVerifiersSystemIDs(l1ID, l2ID eth.ChainID) DefaultMantleSingleChainTwoVerifiersSystemIDs {
	return DefaultMantleSingleChainTwoVerifiersSystemIDs{
		DefaultMantleSingleChainMultiNodeSystemIDs: NewDefaultMantleSingleChainMultiNodeSystemIDs(l1ID, l2ID),
		L2CLC: stack.NewL2CLNodeID("c", l2ID),
		L2ELC: stack.NewL2ELNodeID("c", l2ID),
	}
}

func DefaultMantleSingleChainTwoVerifiersSystem(dest *DefaultMantleSingleChainTwoVerifiersSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMantleSingleChainTwoVerifiersSystemIDs(DefaultL1ID, DefaultL2AID)

	opt := stack.Combine[*Orchestrator]()
	opt.Add(DefaultMantleSingleChainMultiNodeSystem(&dest.DefaultMantleSingleChainMultiNodeSystemIDs))

	opt.Add(WithL2ELNode(ids.L2ELC))
	// Specific options are applied after global options
	opt.Add(WithL2CLNode(ids.L2CLC, ids.L1CL, ids.L1EL, ids.L2ELC, L2CLVerifierDisableUnsafeOnly()))

	opt.Add(WithL2CLP2PConnection(ids.L2CL, ids.L2CLC))
	opt.Add(WithL2ELP2PConnection(ids.L2EL, ids.L2ELC))
	opt.Add(WithL2CLP2PConnection(ids.L2CLB, ids.L2CLC))
	opt.Add(WithL2ELP2PConnection(ids.L2ELB, ids.L2ELC))

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = ids
	}))
	return opt
}
