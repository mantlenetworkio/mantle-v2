package dsl

import (
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

func (n *L2Network) IsMantleForkActive(fork forks.MantleForkName) bool {
	el := NewL2ELNode(n.inner.L2ELNode(match.FirstL2EL), n.control)
	timestamp := el.BlockRefByLabel(eth.Unsafe).Time
	return n.IsMantleForkActiveAt(fork, timestamp)
}

func (n *L2Network) IsMantleForkActiveAt(forkName forks.MantleForkName, timestamp uint64) bool {
	return n.Escape().RollupConfig().IsMantleForkActive(forkName, timestamp)
}
