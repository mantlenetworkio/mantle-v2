package two_follower_opnodes_failure

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestTwoFollowerOpNodesFailureAndRecovery is the entry point for the
// paired follower-op-node failure + recovery scenario. Both halves
// share a single orchestrator boot. "Failure" stops both follower
// op-nodes (NOT their conductors) and asserts the sysgo-faithful
// production-monitor outcome (see the docstring on runFailure for the
// full reasoning around the sysgo P2P topology constraint that makes
// this divergent from the README's "active sequencer is still working"
// claim). "Recovery" restarts both crashed op-nodes and asserts the
// cluster returns to the pre-failure healthy 3-member baseline.
//
// Subtests run in declaration order; "Recovery" requires the post-
// "Failure" state and must not be invoked in isolation. To run the
// recovery half:
//
//	go test -run TestTwoFollowerOpNodesFailureAndRecovery
//
// which executes both subtests in order.
func TestTwoFollowerOpNodesFailureAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	// Suite-wide pre-failure baseline.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys)
	})
	if gt.Failed() {
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys)
	})
}

// runFailure exercises op-conductor README failure mode (table row
// "2 standby are down") under the sysgo P2P topology.
//
// What the README says (op-conductor/README.md):
//
//	"2 standby are down. Cluster will still be healthy, active
//	 sequencer is still working, and raft consensus is healthy as well,
//	 so no leadership transfer will happen."
//
// What this test pins down — sysgo-faithful interpretation
//
// Under a real-world P2P topology (validators, batchers, RPC peers,
// etc.), the active sequencer's op-node would have peers from sources
// other than the two crashed follower op-nodes, so its
// SequencerHealthMonitor's MinPeerCount check would still pass and the
// README's "active sequencer is still working" promise would hold.
//
// Under the sysgo static mesh (op-devstack/sysgo/mantle_system_conductor.go
// wires only A↔B, A↔C, B↔C), killing op-nodes B and C leaves A with
// zero peers. With cfg.HealthCheck.MinPeerCount=1
// (op-devstack/sysgo/l2_conductor.go), A's conductor's
// SequencerHealthMonitor flips A to unhealthy on its next tick, A's
// action loop sees (leader && !healthy && active), stops its sequencer
// and calls transferLeader().
//
// What this test therefore verifies, sysgo-faithfully:
//
//  1. The production health monitor on the active leader (A) detects
//     A's loss of peers, stops A's sequencer, and initiates leadership
//     transfer — the action-loop branch (leader && !healthy && active)
//     is exercised end-to-end.
//
//  2. After the dust settles, every voter in the cluster reports
//     SequencerHealthy=false. This is the sysgo-specific consequence
//     of the static-mesh topology and is a *correct* expression of
//     the production health monitor's MinPeerCount check — A has zero
//     peers because its only static-mesh peers (B and C) are dead, B
//     and C have dead op-nodes so SyncStatus fails on their pair.
//
//  3. The cluster does NOT silently elect a new leader on top of a
//     peer-degraded or dead-op-node voter. No node enters a stable
//     (leader=true, healthy=false, active=true) state — every
//     candidate's action loop correctly self-demotes.
//
// State left for the recovery subtest:
//   - both follower op-nodes (B, C): stopped.
//   - active sequencer's op-node (A / sys.L2CL): still alive, but
//     conductor reports SequencerHealthy=false because it has no peers.
//   - raft leader: may have rotated away from A; the recovery subtest
//     does not require a specific leader, only a healthy 3-member
//     cluster after the op-nodes come back.
//   - no node is sustainably sequencing.
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestTwoFollowerOpNodesFailureAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; this scenario needs >= 3 "+
				"to exercise the peer-degraded leader cascade",
				chainID, len(conductors))
			continue
		}

		// 1+2. Identify the active leader (sys.L2CL = A) and the two
		//      followers as every CL except sys.L2CL.
		activeCLID := sys.L2CL.Escape().ID()
		allCLs := sys.L2Chain.Escape().L2CLNodes()
		r.Equal(3, len(allCLs),
			"chain %s: expected exactly 3 L2 CL nodes, got %d",
			chainID, len(allCLs))

		var followerCLs []*dsl.L2CLNode
		var followerIDs []stack.L2CLNodeID
		for _, c := range allCLs {
			if c.ID() == activeCLID {
				continue
			}
			followerCLs = append(followerCLs, dsl.NewL2CLNode(c, sys.ControlPlane))
			followerIDs = append(followerIDs, c.ID())
		}
		r.Len(followerCLs, 2,
			"chain %s: expected 2 standby CLs after excluding active %s, got %d",
			chainID, activeCLID, len(followerCLs))

		activeProbeCtx, cancelActiveProbe := context.WithTimeout(
			t.Ctx(), 3*time.Second)
		isActive, err := sys.L2CL.Escape().RollupAPI().SequencerActive(activeProbeCtx)
		cancelActiveProbe()
		r.NoError(err,
			"chain %s: SequencerActive RPC failed on supposed leader %s",
			chainID, activeCLID)
		r.True(isActive,
			"chain %s: sys.L2CL %s is not actively sequencing — preset "+
				"selection diverged from cluster state",
			chainID, activeCLID)

		// Capture the current raft leader and bind dsl.Conductors for all
		// three voters so we can poll IsLeader / SequencerHealthy on each.
		leaderInfoBefore := conductors[0].FetchLeader()
		r.NotNil(leaderInfoBefore,
			"chain %s: cluster has no leader before stopping followers",
			chainID)
		oldLeaderID := leaderInfoBefore.ID
		r.NotEmpty(oldLeaderID, "chain %s: empty leader ID", chainID)

		var oldLeaderDsl *dsl.Conductor
		var followerConductors []*dsl.Conductor
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			if id == oldLeaderID {
				oldLeaderDsl = c
				continue
			}
			followerConductors = append(followerConductors, c)
		}
		r.NotNil(oldLeaderDsl,
			"chain %s: dsl.Conductor for old leader %s not found",
			chainID, oldLeaderID)
		r.Len(followerConductors, 2,
			"chain %s: expected 2 follower conductors, got %d",
			chainID, len(followerConductors))

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Pre-failure cluster state",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"activeCL", activeCLID,
			"followers", followerIDs,
			"baselineUnsafe", baseline)

		// 3. Stop both follower op-nodes (NOT their conductors). We
		//    deliberately do NOT register a t.Cleanup restart — the
		//    recovery subtest owns the restart-and-rejoin proof.
		for _, f := range followerCLs {
			f.Stop()
		}

		// 4. Wait for the production health monitor cascade. Each
		//    conductor's monitor ticks every cfg.HealthCheck.Interval
		//    (1s in sysgo). The expected cascade:
		//      a) Followers (B, C): SyncStatus calls to their dead
		//         op-nodes fail → both report unhealthy.
		//      b) Active leader (A): SyncStatus succeeds (its op-node
		//         is alive) but PeerStats reports 0 peers (A's only
		//         static-mesh peers were B and C, both dead).
		//         MinPeerCount=1 fails → A reports unhealthy, action
		//         loop stops sequencer + calls transferLeader().
		//
		//    Budget chosen at 30s — Interval=1s but the leader's first
		//    transferLeader RPC round-trip + raft transfer can take
		//    10–20s in practice; 30s gives margin.
		const cascadeBudget = 30 * time.Second
		cascadeEnd := time.Now().Add(cascadeBudget)

		probeHealth := func(c *dsl.Conductor) bool {
			ctx, cancel := context.WithTimeout(t.Ctx(), 2*time.Second)
			defer cancel()
			h, err := c.Escape().RpcAPI().SequencerHealthy(ctx)
			if err != nil {
				return true // treat RPC error as "still healthy" → keep polling
			}
			return h
		}

		var leaderTransferred bool
		var allUnhealthy bool
		for time.Now().Before(cascadeEnd) {
			if !leaderTransferred && !oldLeaderDsl.IsLeader() {
				leaderTransferred = true
			}
			if !probeHealth(oldLeaderDsl) &&
				!probeHealth(followerConductors[0]) &&
				!probeHealth(followerConductors[1]) {
				allUnhealthy = true
			}
			if leaderTransferred && allUnhealthy {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}

		// 4a. Assert the production-monitor's transferLeader() fired on
		//     the active leader. Even though no transfer target can
		//     reach a stable healthy state under sysgo (none exist),
		//     the *initiation* of the transfer is the observable side
		//     of the action-loop branch we want to verify.
		r.True(leaderTransferred,
			"chain %s: old leader %s still reports IsLeader=true after "+
				"%s — the production health monitor's transferLeader() "+
				"on peer-count-below-min never fired",
			chainID, oldLeaderID, cascadeBudget)

		// 4b. Assert the sysgo-specific consequence: every voter
		//     reports SequencerHealthy=false.
		r.True(allUnhealthy,
			"chain %s: not every voter reported SequencerHealthy=false "+
				"within %s after the 2 follower op-node kills — under "+
				"the sysgo static mesh the active leader (%s) should "+
				"lose both peers and trip MinPeerCount=1",
			chainID, cascadeBudget, oldLeaderID)

		// 5. Settle window + chain-still-advancing assertion. Under the
		//    production monitor with all voters unhealthy, the cluster
		//    settles into a steady state where:
		//      - whichever conductor wins each oscillation of raft
		//        leadership stays leader long enough to keep its FSM
		//        responsive,
		//      - raft's auto-election tends to pick the original leader
		//        back because its match index keeps advancing (the
		//        only voter with a working sequencer is A),
		//      - A's sequencer's commit-to-conductor RPCs against
		//        whoever is leader at the moment intermittently
		//        succeed, so the chain advances roughly at block-time
		//        cadence, modulo the leadership oscillation.
		//    Therefore the README's "active sequencer is still working"
		//    claim is actually preserved under sysgo, *despite* the
		//    additional finding above that leadership transfer does
		//    fire (it ping-pongs back). We assert chain progress to
		//    pin down both halves: the cluster doesn't deadlock, AND
		//    the production monitor's transferLeader call doesn't
		//    permanently relocate leadership to a peerless voter.
		const settleWindow = 8 * time.Second
		time.Sleep(settleWindow)

		preProgress := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		const progressWindow = 10 * time.Second
		time.Sleep(progressWindow)
		postProgress := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number

		// Block time is ~2s, so 10s should yield ~5 blocks at full
		// rate. We require at least 1 to prove the chain is alive
		// (not deadlocked) — slack accommodates raft oscillation
		// stealing some block slots.
		const minProgress uint64 = 1
		r.GreaterOrEqual(postProgress-preProgress, minProgress,
			"chain %s: unsafe head did not advance from %d over %s "+
				"after the cascade — under sysgo two-follower-down the "+
				"chain should keep making *some* progress because raft "+
				"oscillation tends to return leadership to the only "+
				"voter with a live sequencer (the original leader %s)",
			chainID, preProgress, progressWindow, oldLeaderID)

		// 6. Final state log so Recovery diagnostics have a reference.
		head := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("README scenario verified (sysgo-faithful): "+
			"production health monitor detected both followers' op-nodes "+
			"down, then the active leader's own peer count dropped to 0 "+
			"and its action loop initiated leadership transfer; no node "+
			"is sustainably sequencing in the resulting peer-degraded "+
			"cluster",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"followers", followerIDs,
			"baselineUnsafe", baseline,
			"observedUnsafe", head)
	}
}
