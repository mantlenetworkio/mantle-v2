package sysgo

import (
	"context"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/p2p"

	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/testreq"
)

func WithL2ELP2PConnection(l2EL1ID, l2EL2ID stack.L2ELNodeID) stack.Option[*Orchestrator] {
	return stack.AfterDeploy(func(orch *Orchestrator) {
		require := orch.P().Require()

		l2EL1, ok := orch.l2ELs.Get(l2EL1ID)
		require.True(ok, "looking for L2 EL node 1 to connect p2p")
		l2EL2, ok := orch.l2ELs.Get(l2EL2ID)
		require.True(ok, "looking for L2 EL node 2 to connect p2p")
		require.Equal(l2EL1ID.ChainID(), l2EL2ID.ChainID(), "must be same l2 chain")

		ctx := orch.P().Ctx()
		logger := orch.P().Logger()

		rpc1, err := dial.DialRPCClientWithTimeout(ctx, logger, l2EL1.UserRPC())
		require.NoError(err, "failed to connect to el1 rpc")
		defer rpc1.Close()
		rpc2, err := dial.DialRPCClientWithTimeout(ctx, logger, l2EL2.UserRPC())
		require.NoError(err, "failed to connect to el2 rpc")
		defer rpc2.Close()

		ConnectP2P(orch.P().Ctx(), require, rpc1, rpc2)
	})
}

type RpcCaller interface {
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
}

// ConnectP2P creates a p2p peer connection between node1 and node2.
//
// Peer IDs from admin_peers[].id are compared directly against admin_nodeInfo.id.
// Since mantle-xyz/reth v2.2.1 (commit e166f3a9, porting paradigmxyz/reth#23318
// and #23319), both fields return the same 64-hex keccak256 node ID without a
// "0x" prefix, matching go-ethereum's format. Prior to that fix,
// admin_nodeInfo.id returned a 66-hex compressed public key and
// admin_peers[].id included a "0x" prefix, requiring client-side normalization.
//
// Only the initiator calls admin_addPeer. Both op-geth and op-reth dial static
// peers added via admin_addPeer, so a single unidirectional call is sufficient.
// The difference is timing: op-geth's dialsched.addStatic triggers an immediate
// synchronous dial, while op-reth's PeersManager schedules the dial asynchronously
// via fill_outbound_slots, which polls every 5 s (refill_slots_interval). The 45 s
// timeout accommodates this delay with margin for CI load.
func ConnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var initiatorInfo, acceptorInfo p2p.NodeInfo
	require.NoError(initiator.CallContext(ctx, &initiatorInfo, "admin_nodeInfo"), "get initiator node info")
	require.NoError(acceptor.CallContext(ctx, &acceptorInfo, "admin_nodeInfo"), "get acceptor node info")

	var peerAdded bool
	require.NoError(initiator.CallContext(ctx, &peerAdded, "admin_addPeer", acceptorInfo.Enode), "initiator add peer")
	require.True(peerAdded, "initiator should have added peer successfully")

	// 45 s timeout: 30 s base + 5 s for reth's async dial scheduler + 10 s CI margin.
	connCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	// Require both sides to see each other: an AND condition prevents a half-open
	// devp2p connection (only one peer list updated) from being accepted as success,
	// which would cause subsequent sync steps to flake.
	err := wait.For(connCtx, time.Second, func() (bool, error) {
		var initiatorPeers, acceptorPeers []peer
		if err := initiator.CallContext(connCtx, &initiatorPeers, "admin_peers"); err != nil {
			return false, err
		}
		if err := acceptor.CallContext(connCtx, &acceptorPeers, "admin_peers"); err != nil {
			return false, err
		}
		initiatorSeesAcceptor := slices.ContainsFunc(initiatorPeers, func(p peer) bool {
			return p.ID == acceptorInfo.ID
		})
		acceptorSeesInitiator := slices.ContainsFunc(acceptorPeers, func(p peer) bool {
			return p.ID == initiatorInfo.ID
		})
		return initiatorSeesAcceptor && acceptorSeesInitiator, nil
	})
	require.NoError(err, "The peer was not connected")
}

// DisconnectP2P disconnects a p2p peer connection between node1 and node2.
//
// Only the initiator calls admin_removePeer, matching the unidirectional
// admin_addPeer in ConnectP2P. Both sides are polled to confirm full teardown.
func DisconnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var initiatorInfo, acceptorInfo p2p.NodeInfo
	require.NoError(initiator.CallContext(ctx, &initiatorInfo, "admin_nodeInfo"), "get initiator node info")
	require.NoError(acceptor.CallContext(ctx, &acceptorInfo, "admin_nodeInfo"), "get acceptor node info")

	var peerRemoved bool
	require.NoError(initiator.CallContext(ctx, &peerRemoved, "admin_removePeer", acceptorInfo.ENR), "initiator remove peer")
	require.True(peerRemoved, "initiator should have removed peer successfully")

	// Wait for both sides to no longer see each other.
	waitCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := wait.For(waitCtx, time.Second, func() (bool, error) {
		var initiatorPeers, acceptorPeers []peer
		if err := initiator.CallContext(waitCtx, &initiatorPeers, "admin_peers"); err != nil {
			return false, err
		}
		if slices.ContainsFunc(initiatorPeers, func(p peer) bool {
			return p.ID == acceptorInfo.ID
		}) {
			return false, nil
		}
		if err := acceptor.CallContext(waitCtx, &acceptorPeers, "admin_peers"); err != nil {
			return false, err
		}
		return !slices.ContainsFunc(acceptorPeers, func(p peer) bool {
			return p.ID == initiatorInfo.ID
		}), nil
	})
	require.NoError(err, "The peer was not disconnected")
}

type peer struct {
	ID string `json:"id"`
}
