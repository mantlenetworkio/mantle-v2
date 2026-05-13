package dsl

import (
	"context"
	"time"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

func (n *L2Network) IsMantleForkActive(fork forks.MantleForkName) bool {
	el := n.inner.L2ELNode(match.FirstL2EL)
	// Use InfoByLabel instead of L2BlockRefByLabel to avoid parsing L1 info
	// deposit tx, which may have an incompatible function signature on Mantle.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	info, err := el.L2EthClient().InfoByLabel(ctx, eth.Unsafe)
	if err != nil {
		// Fallback: try L2BlockRefByLabel if InfoByLabel fails
		elNode := NewL2ELNode(el, n.control)
		return n.IsMantleForkActiveAt(fork, elNode.BlockRefByLabel(eth.Unsafe).Time)
	}
	return n.IsMantleForkActiveAt(fork, info.Time())
}

func (n *L2Network) IsMantleForkActiveAt(forkName forks.MantleForkName, timestamp uint64) bool {
	return n.Escape().RollupConfig().IsMantleForkActive(forkName, timestamp)
}
