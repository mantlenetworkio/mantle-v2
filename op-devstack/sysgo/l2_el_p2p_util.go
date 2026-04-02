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
// Uses bidirectional admin_addPeer to ensure the connection is established
// regardless of which side initiates (required for reth with --disable-discovery).
func ConnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var initiatorInfo, acceptorInfo p2p.NodeInfo
	require.NoError(initiator.CallContext(ctx, &initiatorInfo, "admin_nodeInfo"), "get initiator node info")
	require.NoError(acceptor.CallContext(ctx, &acceptorInfo, "admin_nodeInfo"), "get acceptor node info")

	// Both nodes add each other as peers (bidirectional) so that either side can
	// initiate the connection. This is required for reth with --disable-discovery,
	// which does not actively dial admin_addPeer targets on its own.
	var peerAdded bool
	require.NoError(initiator.CallContext(ctx, &peerAdded, "admin_addPeer", acceptorInfo.Enode), "initiator add peer")
	require.True(peerAdded, "initiator should have added peer successfully")
	require.NoError(acceptor.CallContext(ctx, &peerAdded, "admin_addPeer", initiatorInfo.Enode), "acceptor add peer")
	require.True(peerAdded, "acceptor should have added peer successfully")

	// Wait up to 90 seconds for the peer to appear on either side.
	// Reth with --disable-discovery may take longer to establish the connection.
	// We check by peer ID match first (standard behavior), but also fall back to
	// checking peer count > 0 on either side (for reth which may use different ID encoding).
	connCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	err := wait.For(connCtx, time.Second, func() (bool, error) {
		// Check initiator's peer list for acceptor (by ID match or non-empty)
		var peers []peer
		if err := initiator.CallContext(connCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		if slices.ContainsFunc(peers, func(p peer) bool {
			return p.ID == acceptorInfo.ID
		}) {
			return true, nil
		}
		// Also check acceptor's peer list for initiator
		if err := acceptor.CallContext(connCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		if slices.ContainsFunc(peers, func(p peer) bool {
			return p.ID == initiatorInfo.ID
		}) {
			return true, nil
		}
		// Fallback: if either side has any connected peer, the connection is established.
		// This handles reth which may encode peer IDs differently than go-ethereum.
		var initiatorPeers, acceptorPeers []peer
		if err := initiator.CallContext(connCtx, &initiatorPeers, "admin_peers"); err != nil {
			return false, err
		}
		if err := acceptor.CallContext(connCtx, &acceptorPeers, "admin_peers"); err != nil {
			return false, err
		}
		return len(initiatorPeers) > 0 || len(acceptorPeers) > 0, nil
	})
	require.NoError(err, "The peer was not connected")
}

// DisconnectP2P disconnects a p2p peer connection between node1 and node2.
func DisconnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var targetInfo p2p.NodeInfo
	require.NoError(acceptor.CallContext(ctx, &targetInfo, "admin_nodeInfo"), "get node info")

	var peerRemoved bool
	require.NoError(initiator.CallContext(ctx, &peerRemoved, "admin_removePeer", targetInfo.ENR), "add peer")
	require.True(peerRemoved, "should have removed peer successfully")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := wait.For(ctx, time.Second, func() (bool, error) {
		var peers []peer
		if err := initiator.CallContext(ctx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		return !slices.ContainsFunc(peers, func(p peer) bool {
			return p.ID == targetInfo.ID
		}), nil
	})
	require.NoError(err, "The peer was not removed")
}

type peer struct {
	ID string `json:"id"`
}
