package sysgo

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

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

// enodeNodeID derives the canonical 64-hex keccak256 devp2p node ID from an
// enode URL. This is the authoritative identifier that admin_peers consistently
// returns (modulo an optional "0x" prefix in op-reth).
//
// Using the enode URL rather than admin_nodeInfo.id avoids a cross-client format
// mismatch:
//   - op-geth admin_nodeInfo.id  → 64-hex keccak256 hash, no prefix
//   - op-reth admin_nodeInfo.id  → 66-hex compressed public key, no prefix
//   - op-geth admin_peers[].id   → 64-hex keccak256 hash, no prefix
//   - op-reth admin_peers[].id   → 64-hex keccak256 hash, "0x" prefix
func enodeNodeID(require *testreq.Assertions, enodeURL string) string {
	node, err := enode.ParseV4(enodeURL)
	require.NoError(err, "failed to parse enode URL")
	return node.ID().String() // 64 lowercase hex chars, no "0x"
}

// matchesPeerID reports whether the peer ID returned by admin_peers matches the
// given canonical node ID. It strips an optional "0x" prefix to handle the
// format difference between op-geth (no prefix) and op-reth (with "0x" prefix).
func matchesPeerID(peerID, targetNodeID string) bool {
	return strings.EqualFold(strings.TrimPrefix(peerID, "0x"), targetNodeID)
}

// ConnectP2P creates a p2p peer connection between node1 and node2.
//
// Both nodes call admin_addPeer on each other (bidirectional) so that either
// side can initiate the TCP dial. This is required for op-reth with
// --disable-discovery, which does not actively dial admin_addPeer targets.
//
// Peer verification uses the canonical 64-hex keccak256 node ID derived from
// the enode URL. This avoids a cross-client ID encoding mismatch where
// op-reth's admin_nodeInfo.id is a compressed public key (different format)
// while op-reth's admin_peers[].id is a keccak256 hash with a "0x" prefix.
func ConnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var initiatorInfo, acceptorInfo p2p.NodeInfo
	require.NoError(initiator.CallContext(ctx, &initiatorInfo, "admin_nodeInfo"), "get initiator node info")
	require.NoError(acceptor.CallContext(ctx, &acceptorInfo, "admin_nodeInfo"), "get acceptor node info")

	// Derive canonical node IDs from the enode URLs (consistent across EL clients).
	initiatorNodeID := enodeNodeID(require, initiatorInfo.Enode)
	acceptorNodeID := enodeNodeID(require, acceptorInfo.Enode)

	// Both nodes add each other as peers so that either side can initiate the
	// connection. Required for reth with --disable-discovery.
	var peerAdded bool
	require.NoError(initiator.CallContext(ctx, &peerAdded, "admin_addPeer", acceptorInfo.Enode), "initiator add peer")
	require.True(peerAdded, "initiator should have added peer successfully")
	require.NoError(acceptor.CallContext(ctx, &peerAdded, "admin_addPeer", initiatorInfo.Enode), "acceptor add peer")
	require.True(peerAdded, "acceptor should have added peer successfully")

	// Wait up to 90 seconds for the peer to appear on either side.
	// 30 seconds is sufficient for geth (synchronous P2P). The extended timeout
	// accommodates reth with --disable-discovery: after both nodes call admin_addPeer
	// on each other, the actual TCP handshake is initiated by the acceptor side, which
	// may take additional time before the devp2p listener processes the outbound dial
	// request. In the worst case (CI under load) this can exceed 30 seconds for reth.
	connCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	err := wait.For(connCtx, time.Second, func() (bool, error) {
		var peers []peer
		// Check initiator's peer list for acceptor
		if err := initiator.CallContext(connCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		if slices.ContainsFunc(peers, func(p peer) bool {
			return matchesPeerID(p.ID, acceptorNodeID)
		}) {
			return true, nil
		}
		// Check acceptor's peer list for initiator
		if err := acceptor.CallContext(connCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		return slices.ContainsFunc(peers, func(p peer) bool {
			return matchesPeerID(p.ID, initiatorNodeID)
		}), nil
	})
	require.NoError(err, "The peer was not connected")
}

// DisconnectP2P disconnects a p2p peer connection between node1 and node2.
func DisconnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var targetInfo p2p.NodeInfo
	require.NoError(acceptor.CallContext(ctx, &targetInfo, "admin_nodeInfo"), "get node info")

	// Derive canonical node ID from the enode URL (consistent across EL clients).
	targetNodeID := enodeNodeID(require, targetInfo.Enode)

	var peerRemoved bool
	require.NoError(initiator.CallContext(ctx, &peerRemoved, "admin_removePeer", targetInfo.ENR), "remove peer")
	require.True(peerRemoved, "should have removed peer successfully")

	waitCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := wait.For(waitCtx, time.Second, func() (bool, error) {
		var peers []peer
		if err := initiator.CallContext(waitCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		return !slices.ContainsFunc(peers, func(p peer) bool {
			return matchesPeerID(p.ID, targetNodeID)
		}), nil
	})
	require.NoError(err, "The peer was not removed")
}

type peer struct {
	ID string `json:"id"`
}
