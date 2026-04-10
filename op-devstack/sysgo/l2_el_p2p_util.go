package sysgo

import (
	"context"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/p2p"

	enginekind "github.com/ethereum-optimism/optimism/op-node/rollup/engine"
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
// The initiator always calls admin_addPeer on the acceptor to trigger the outbound
// dial. For op-reth with --disable-discovery, the acceptor must also call
// admin_addPeer on the initiator because reth does not actively dial admin_addPeer
// targets; the acceptor must initiate the TCP handshake. For op-geth, the
// initiator's addPeer is sufficient and bidirectional addPeer is intentionally
// avoided: simultaneous dials cause geth's devp2p scheduler to add repeated
// dial-history entries on the "loser" side, making subsequent reconnects unreliable.
func ConnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var initiatorInfo, acceptorInfo p2p.NodeInfo
	require.NoError(initiator.CallContext(ctx, &initiatorInfo, "admin_nodeInfo"), "get initiator node info")
	require.NoError(acceptor.CallContext(ctx, &acceptorInfo, "admin_nodeInfo"), "get acceptor node info")

	// Initiator always dials the acceptor.
	var peerAdded bool
	require.NoError(initiator.CallContext(ctx, &peerAdded, "admin_addPeer", acceptorInfo.Enode), "initiator add peer")
	require.True(peerAdded, "initiator should have added peer successfully")

	// reth with --disable-discovery does not actively dial admin_addPeer targets.
	// Have the acceptor also call addPeer so it can initiate the TCP handshake.
	// op-geth dials synchronously from the initiator side, so bidirectional addPeer
	// is not needed and intentionally skipped to avoid dial-history re-dial issues.
	if devstackL2ELKind() == enginekind.Reth {
		require.NoError(acceptor.CallContext(ctx, &peerAdded, "admin_addPeer", initiatorInfo.Enode), "acceptor add peer")
		require.True(peerAdded, "acceptor should have added peer successfully")
	}

	// Wait for the peer connection to appear.
	// geth: 30 s is sufficient for the initiator-side synchronous dial.
	// reth with --disable-discovery needs more time: the TCP handshake is
	// initiated by the acceptor side and may be delayed while the devp2p listener
	// processes the outbound dial request. In CI under load this can exceed 30 s.
	connTimeout := 30 * time.Second
	if devstackL2ELKind() == enginekind.Reth {
		connTimeout = 90 * time.Second
	}
	connCtx, cancel := context.WithTimeout(context.Background(), connTimeout)
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
// For op-reth, both nodes call admin_removePeer on each other (bidirectional)
// to mirror the bidirectional admin_addPeer done by ConnectP2P. For op-geth,
// only the initiator calls removePeer; the acceptor never had the initiator as a
// static peer (unidirectional addPeer), so there is nothing to clean up on the
// acceptor side.
func DisconnectP2P(ctx context.Context, require *testreq.Assertions, initiator RpcCaller, acceptor RpcCaller) {
	var initiatorInfo, acceptorInfo p2p.NodeInfo
	require.NoError(acceptor.CallContext(ctx, &acceptorInfo, "admin_nodeInfo"), "get acceptor node info")

	var peerRemoved bool
	require.NoError(initiator.CallContext(ctx, &peerRemoved, "admin_removePeer", acceptorInfo.ENR), "initiator remove peer")
	require.True(peerRemoved, "initiator should have removed peer successfully")

	// For reth: also remove acceptor-side static peer to mirror bidirectional ConnectP2P.
	if devstackL2ELKind() == enginekind.Reth {
		require.NoError(initiator.CallContext(ctx, &initiatorInfo, "admin_nodeInfo"), "get initiator node info")
		require.NoError(acceptor.CallContext(ctx, &peerRemoved, "admin_removePeer", initiatorInfo.ENR), "acceptor remove peer")
		require.True(peerRemoved, "acceptor should have removed peer successfully")
	}

	// Wait for both sides (or just the initiator side for geth) to no longer see each other.
	waitCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := wait.For(waitCtx, time.Second, func() (bool, error) {
		var peers []peer
		if err := initiator.CallContext(waitCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		if slices.ContainsFunc(peers, func(p peer) bool {
			return p.ID == acceptorInfo.ID
		}) {
			return false, nil
		}
		if devstackL2ELKind() != enginekind.Reth {
			return true, nil
		}
		if err := acceptor.CallContext(waitCtx, &peers, "admin_peers"); err != nil {
			return false, err
		}
		return !slices.ContainsFunc(peers, func(p peer) bool {
			return p.ID == initiatorInfo.ID
		}), nil
	})
	require.NoError(err, "The peer was not disconnected")
}

type peer struct {
	ID string `json:"id"`
}
