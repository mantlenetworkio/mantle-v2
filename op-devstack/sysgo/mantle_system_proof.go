package sysgo

import (
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
)

// MantleProofSystem mirrors ProofSystem but uses Mantle deployer/genesis builders.
func MantleProofSystem(dest *DefaultMinimalSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMinimalSystemIDs(DefaultL1ID, DefaultL2AID)
	opt := defaultMantleMinimalSystemOpts(&ids, dest)
	opt.Add(WithCannonGameTypeAdded(ids.L1EL, ids.L2.ChainID()))
	return opt
}
