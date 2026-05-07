package recovered_leader_election_during_sync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils/elfaultinjector"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestRecoveredLeaderElectionDuringSync exercises the dangerous failure
// mode that the rejoin scenario in tolerate_one_conductor_failure_test.go
// only hand-waves about: a stale-but-rejoined conductor is forced into
// leadership BEFORE its op-geth has caught up to the FSM head, and the
// chain stalls because op-conductor's compareUnsafeHead gate refuses to
// startSequencer on a stale op-geth.
//
// Why this is a real concern (not a theoretical one):
//
//   - Raft's election restriction (Raft §5.4.1) only checks raft-log
//     up-to-dateness via lastLogIndex/lastLogTerm. Raft has no awareness
//     of the op-geth chain — those are two independent state machines
//     that catch up at very different speeds:
//       Raft FSM       — a few UnsafeHead pointers, kB/s, ms latency,
//                        replicated by AppendEntries
//       op-geth chain  — full block bodies + state transitions,
//                        replicated by P2P gossip / EL sync
//   - On rejoin, a conductor's raft log is brought up to date in
//     milliseconds via AppendEntries. Its op-geth, however, must
//     download N missed blocks via P2P, which can take orders of
//     magnitude longer.
//   - In the window between "raft log caught up" and "op-geth caught
//     up", the rejoined node is fully eligible to win an election
//     even though it cannot actually sequence: compareUnsafeHead in
//     op-conductor/conductor/service.go gates startSequencer on the
//     local op-geth being within 1 block of FSM head. A gap > 1 yields
//     ErrUnsafeHeadMismatch and StartSequencer is silently NOT issued.
//   - If the rejoined node DOES win the election (e.g. the current
//     leader crashes a moment later, or an operator manually transfers
//     leadership too eagerly), it becomes leader at the raft layer but
//     never starts sequencing. Other nodes are followers, so they
//     don't sequence either. The chain stalls until op-geth catches
//     up via P2P, at which point the action loop's next tick succeeds.
//
// What this test does:
//
//  1. Identify the current leader and capture the to-be-orphan op-node,
//     op-geth and its Engine API fault injector.
//  2. Stop the leader's conductor. IMMEDIATELY activate the orphan's
//     EL fault injector (RejectFromBlock = orphanBaseline+1) before
//     the new leader is even elected, so that every engine_newPayload
//     above baseline lands as INVALID on the orphan EL no matter how
//     it arrives (gossip, sync, or the orphan op-node's own retry).
//  3. Let the survivors elect a new leader and let the chain advance
//     under that new leader for ~20s (~10 blocks at 2s block time).
//     The new leader's op-geth advances normally; the orphan's op-geth
//     stays pinned at baseline because the injector rejects every
//     engine_newPayload it sees from gossip. This builds the >1-block
//     gap that compareUnsafeHead's PostUnsafePayload bypass cannot
//     close, which is the precondition for the stall.
//  4. Restart the orphan's conductor. Raft log catches up via
//     AppendEntries (the FSM contains all the missed payloads); the
//     action loop fires StopSequencer on op-node, which then tries to
//     accept gossip/sync from the new leader. But the fault injector
//     keeps rejecting every engine_newPayload above the orphan's stuck
//     head, so op-geth stays lagging.
//  5. TransferLeadershipTo the recovered conductor. At the raft layer
//     this succeeds — the recovered server's raft log is current, and
//     raft has no visibility into op-geth state.
//  6. Observe the failure mode (stall window, several seconds):
//       a. recoveredCL's SequencerActive stays false, because every
//          action-loop tick's compareUnsafeHead returns
//          ErrUnsafeHeadMismatch (gap > 1).
//       b. No EL produces new blocks. The new-leader-no-sequencing
//          deadlock is end-to-end visible.
//  7. Deactivate the fault injector. P2P / sync now closes the gap
//     in op-geth. compareUnsafeHead transitions through ">1 gap" →
//     "gap == 1, PostUnsafePayload bypass" → "gap == 0, direct
//     StartSequencer", and sequencing resumes under the recovered
//     leader.
//  8. Verify the chain advances under the recovered leader.
//
// What this test PROVES:
//
//   - The compareUnsafeHead gate is not bypassed when leadership is
//     transferred to a stale node. (Safety: no fork, no rogue
//     sequencing on a stale op-geth.)
//   - The chain stalls for as long as op-geth is lagging. (Liveness
//     violated, as warned by op-conductor/INTEGRATION.md.)
//   - Once op-geth catches up, sequencing self-recovers without
//     operator intervention at the raft layer. (The mitigation is
//     operational: don't transfer leadership to a node that has not
//     finished syncing its op-geth.)
//
// What this test does NOT cover:
//
//   - The case where the leader transfer is involuntary (the current
//     leader crashes during the catch-up window, and the rejoined node
//     wins the election spontaneously). Constructing that race in
//     sysgo would require a way to keep the rejoined node's raft log
//     fresher than the third (caught-up) follower's, which is not
//     trivially controllable. The TransferLeadershipTo path validated
//     here exercises the same compareUnsafeHead gate; the only
//     difference is the trigger.
func TestRecoveredLeaderElectionDuringSync(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestRecoveredLeaderElectionDuringSync",
	)

	sys := presets.NewMantleMinimalWithFaultyConductors(t)
	r := t.Require()

	if len(sys.SysgoConductors) < 3 {
		t.Skipf("scenario needs in-process conductors and >=3 voters; "+
			"got %d sysgo conductors (kurtosis/persistent backends do "+
			"not expose Conductor.Stop)", len(sys.SysgoConductors))
	}
	if len(sys.EngineFaultInjectors) == 0 {
		t.Skipf("scenario needs the EL fault injector enabled; got 0 " +
			"injectors (preset wired without WithEngineFaultInjectors)")
	}

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; scenario needs >= 3",
				chainID, len(conductors))
			continue
		}

		// Suite-wide baseline: stable leader, sequencing, followers within
		// 1 block. This is non-negotiable — without it any compareUnsafeHead
		// observation downstream is meaningless.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

		// 1. Identify the leader and capture orphan handles before crash.
		leaderInfoPre := conductors[0].FetchLeader()
		r.NotNil(leaderInfoPre,
			"chain %s: cluster has no leader before crash", chainID)
		oldLeaderID := leaderInfoPre.ID
		r.NotEmpty(oldLeaderID,
			"chain %s: empty leader ID before crash", chainID)
		logger.Info("Pre-crash leader",
			"chain", chainID, "leaderID", oldLeaderID)

		leaderSysgo, ok := sys.SysgoConductors[stack.ConductorID(oldLeaderID)]
		r.True(ok,
			"chain %s: in-process conductor for leader %s missing "+
				"from sys.SysgoConductors", chainID, oldLeaderID)

		orphanEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, oldLeaderID)
		r.NotNil(orphanEL,
			"chain %s: could not locate L2EL paired with leader %s",
			chainID, oldLeaderID)
		orphanCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, oldLeaderID)
		r.NotNil(orphanCL,
			"chain %s: could not locate L2CL paired with leader %s",
			chainID, oldLeaderID)

		orphanInjector, ok := sys.EngineFaultInjectors[orphanEL.Escape().ID()]
		r.True(ok,
			"chain %s: no EL fault injector found for orphan EL %s — "+
				"expected MantleMinimalWithFaultyConductors preset to wire "+
				"one per EL", chainID, orphanEL.Escape().ID())

		orphanBaseline := orphanEL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Orphan baseline before crash",
			"chain", chainID,
			"orphanEL", orphanEL.Escape().ID(),
			"baseline", orphanBaseline)

		// 2. Stop the leader's conductor, wait for new leader, and
		//    capture handles for the new leader's CL/EL plus a survivor
		//    dsl.Conductor for FSM probes.
		var survivor *dsl.Conductor
		for _, c := range conductors {
			if c.Escape().ID() != stack.ConductorID(oldLeaderID) {
				survivor = c
				break
			}
		}
		r.NotNil(survivor,
			"chain %s: no survivor conductor (every dsl.Conductor "+
				"resolved to the leader being crashed)", chainID)

		logger.Info("Crashing leader conductor",
			"chain", chainID, "leaderID", oldLeaderID)
		leaderSysgo.Stop()

		// 2a. IMMEDIATELY activate the orphan's EL fault injector. We
		//     do this before the new leader has even been elected, let
		//     alone produced any blocks, so that EVERY engine_newPayload
		//     above orphanBaseline+1 is rejected on the orphan EL —
		//     regardless of which path delivers it (orphan op-node's
		//     own build attempt, P2P gossip from the new leader, or
		//     EL-sync). Without this, in-process libp2p gossip from the
		//     new leader's op-node into the orphan's op-node closes the
		//     gap in milliseconds, because op-node's gossip-subscriber
		//     path delivers payloads to engine_newPayload even while
		//     SequencerActive is true (op-node does NOT gate inbound
		//     unsafe-payload processing on local sequencing). The
		//     orphan-stuck invariant only holds for the brief window
		//     where the new leader hasn't started gossiping yet — too
		//     short to accumulate the >1-block lag this test needs.
		logger.Info("Activating EL fault injector on orphan",
			"chain", chainID,
			"orphanEL", orphanEL.Escape().ID(),
			"rejectFromBlock", orphanBaseline+1)
		orphanInjector.Activate(elfaultinjector.Rule{
			RejectFromBlock: orphanBaseline + 1,
		})
		// Defensive: even if the test panics later, deactivate so
		// downstream cleanup doesn't see a wedged op-geth.
		t.Cleanup(orphanInjector.Deactivate)

		var newLeaderID string
		electionDeadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(electionDeadline) {
			li := survivor.FetchLeader()
			if li != nil && li.ID != "" && li.ID != oldLeaderID {
				newLeaderID = li.ID
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NotEmpty(newLeaderID,
			"chain %s: no new leader elected within 15s after %s "+
				"crashed", chainID, oldLeaderID)
		logger.Info("New leader elected",
			"chain", chainID, "newLeaderID", newLeaderID)

		var newLeaderDsl *dsl.Conductor
		for _, c := range conductors {
			if c.Escape().ID() == stack.ConductorID(newLeaderID) {
				newLeaderDsl = c
				break
			}
		}
		r.NotNil(newLeaderDsl,
			"chain %s: dsl.Conductor for new leader %s missing",
			chainID, newLeaderID)

		newLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, newLeaderID)
		r.NotNil(newLeaderCL,
			"chain %s: cannot locate new leader CL paired with %s",
			chainID, newLeaderID)
		newLeaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, newLeaderID)
		r.NotNil(newLeaderEL,
			"chain %s: cannot locate new leader EL paired with %s",
			chainID, newLeaderID)

		// Wait until the new leader's op-node has actually started
		// sequencing. compareUnsafeHead from the recovered conductor
		// later relies on FSM head being ahead of orphan op-geth, which
		// only happens once new payloads start landing.
		const newLeaderActiveBudget = 15 * time.Second
		newLeaderActiveDeadline := time.Now().Add(newLeaderActiveBudget)
		var newLeaderActive bool
		var lastNewLeaderActiveErr error
		for time.Now().Before(newLeaderActiveDeadline) {
			probeCtx, cancelProbe := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			newLeaderActive, lastNewLeaderActiveErr = newLeaderCL.RollupAPI().
				SequencerActive(probeCtx)
			cancelProbe()
			if lastNewLeaderActiveErr == nil && newLeaderActive {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NoError(lastNewLeaderActiveErr,
			"chain %s: SequencerActive RPC failed on new leader CL %s",
			chainID, newLeaderCL.ID())
		r.True(newLeaderActive,
			"chain %s: new leader CL %s did not start sequencing within "+
				"%s — failover did not complete, test prerequisites unmet",
			chainID, newLeaderCL.ID(), newLeaderActiveBudget)

		// 3. Accumulate enough op-geth lag on the orphan that the gap
		//    is well past the 1-block PostUnsafePayload bypass.
		//    minLagBlocks * blockTime determines the wall-clock wait.
		const lagAccumulationWindow = 20 * time.Second
		const minLagBlocks uint64 = 5
		logger.Info("Accumulating op-geth lag on orphan",
			"chain", chainID, "window", lagAccumulationWindow)
		time.Sleep(lagAccumulationWindow)

		newLeaderHead := newLeaderEL.BlockRefByLabel(eth.Unsafe).Number
		orphanHeadAfterLag := orphanEL.BlockRefByLabel(eth.Unsafe).Number
		// Orphan op-geth must still be stuck near baseline — the
		// commit-then-publish gate in op-node onBuildSealed
		// (op-node/rollup/sequencing/sequencer.go) ensures the orphan
		// doesn't fork. We allow a +1 slack for an in-flight payload
		// that was already past CommitUnsafePayload at the instant the
		// conductor died.
		r.LessOrEqual(orphanHeadAfterLag, orphanBaseline+1,
			"chain %s: orphan EL %s drifted from baseline (%d -> %d) "+
				"during the lag-accumulation window — the orphan's "+
				"stuck-head invariant has broken; cannot reliably set "+
				"up the lagging-leader scenario",
			chainID, orphanEL.Escape().ID(),
			orphanBaseline, orphanHeadAfterLag)
		r.GreaterOrEqual(newLeaderHead, orphanBaseline+minLagBlocks,
			"chain %s: chain only advanced from %d to %d under new "+
				"leader during %s lag window — expected at least %d "+
				"blocks. Block production is slower than expected; "+
				"the test cannot reliably exercise the >1-block lag case",
			chainID, orphanBaseline, newLeaderHead,
			lagAccumulationWindow, minLagBlocks)
		gap := newLeaderHead - orphanHeadAfterLag
		logger.Info("Op-geth lag accumulated",
			"chain", chainID,
			"orphanHead", orphanHeadAfterLag,
			"newLeaderHead", newLeaderHead,
			"gap", gap)

		// 4. (EL fault injector was activated up in step 2a, before
		//    election, so the orphan EL has been refusing every
		//    engine_newPayload above orphanBaseline+1 throughout the lag
		//    accumulation window. No further setup needed here.)

		// 5. Restart the orphan's conductor. It rejoins as a follower
		//    (raft library invariant: every restart starts in Follower
		//    state regardless of disk-persisted role); the action loop's
		//    first tick (~1s) calls StopSequencer on the orphan op-node.
		logger.Info("Restarting orphan conductor",
			"chain", chainID, "conductorID", oldLeaderID)
		leaderSysgo.Start()

		var recoveredDsl *dsl.Conductor
		for _, c := range conductors {
			if c.Escape().ID() == stack.ConductorID(oldLeaderID) {
				recoveredDsl = c
				break
			}
		}
		r.NotNil(recoveredDsl,
			"chain %s: dsl.Conductor for recovered conductor %s missing",
			chainID, oldLeaderID)
		recoveredCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, oldLeaderID)
		r.NotNil(recoveredCL,
			"chain %s: cannot resolve recovered CL paired with %s",
			chainID, oldLeaderID)

		// Wait until raft membership confirms the recovered server AND
		// the recovered conductor reports IsLeader=false (proper rejoin
		// as follower) AND its op-node's SequencerActive flips to false
		// (action loop has fired StopSequencer). Without all three, the
		// subsequent leadership transfer might race with the action
		// loop and produce an ambiguous SequencerActive value.
		const rejoinBudget = 30 * time.Second
		rejoinDeadline := time.Now().Add(rejoinBudget)
		var recoveredServerInfo consensus.ServerInfo
		var membershipOK, isFollower, sequencerStopped bool
		for time.Now().Before(rejoinDeadline) {
			cm := newLeaderDsl.FetchClusterMembership()
			if cm != nil {
				for _, srv := range cm.Servers {
					if srv.ID == oldLeaderID {
						recoveredServerInfo = srv
						membershipOK = true
						break
					}
				}
			}
			isFollower = !recoveredDsl.IsLeader()

			probeCtx, cancelProbe := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			active, err := recoveredCL.RollupAPI().SequencerActive(probeCtx)
			cancelProbe()
			if err == nil && !active {
				sequencerStopped = true
			}

			if membershipOK && isFollower && sequencerStopped {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		r.True(membershipOK,
			"chain %s: recovered conductor %s missing from new leader's "+
				"cluster membership within %s",
			chainID, oldLeaderID, rejoinBudget)
		r.True(isFollower,
			"chain %s: recovered conductor %s did not report IsLeader=false "+
				"within %s — split-brain after rejoin",
			chainID, oldLeaderID, rejoinBudget)
		r.True(sequencerStopped,
			"chain %s: recovered op-node %s did not have SequencerActive "+
				"flip to false within %s — action loop never called "+
				"StopSequencer",
			chainID, recoveredCL.ID(), rejoinBudget)

		// 6. Force-elect the recovered (still-lagging) conductor by
		//    transferring leadership to it. At the raft layer this
		//    succeeds: raft only checks log up-to-dateness, and
		//    AppendEntries from the new leader has already brought the
		//    recovered server's raft log current. Op-geth lag is
		//    invisible to this transfer — that's the whole point of
		//    this test.
		logger.Info("Transferring leadership to recovered (lagging) conductor",
			"chain", chainID,
			"target", oldLeaderID,
			"opGethGap", newLeaderEL.BlockRefByLabel(eth.Unsafe).Number-
				orphanEL.BlockRefByLabel(eth.Unsafe).Number)
		newLeaderDsl.TransferLeadershipTo(recoveredServerInfo)

		// Confirm the transfer landed at the raft layer: recovered
		// conductor reports IsLeader=true.
		const transferBudget = 15 * time.Second
		transferDeadline := time.Now().Add(transferBudget)
		var transferred bool
		for time.Now().Before(transferDeadline) {
			if recoveredDsl.IsLeader() && !newLeaderDsl.IsLeader() {
				transferred = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(transferred,
			"chain %s: leadership transfer %s -> %s did not complete "+
				"within %s — raft layer refused the transfer (unexpected: "+
				"the recovered conductor's raft log should be current)",
			chainID, newLeaderID, oldLeaderID, transferBudget)
		logger.Info("Leadership transferred at raft layer",
			"chain", chainID, "newLeader", oldLeaderID)

		// 7. Observe the failure mode. The recovered conductor is leader
		//    at the raft layer, but compareUnsafeHead refuses
		//    StartSequencer because its op-geth is lagging by > 1 block.
		//    Two assertions:
		//      a. recoveredCL.SequencerActive stays false through the
		//         observation window.
		//      b. No EL produces new blocks during the same window.
		const stallObservationWindow = 6 * time.Second
		stallStartHead := newLeaderEL.BlockRefByLabel(eth.Unsafe).Number
		time.Sleep(stallObservationWindow)
		stallEndHead := newLeaderEL.BlockRefByLabel(eth.Unsafe).Number

		probeCtx, cancelProbe := context.WithTimeout(t.Ctx(), 2*time.Second)
		recoveredActive, recoveredActiveErr := recoveredCL.RollupAPI().
			SequencerActive(probeCtx)
		cancelProbe()
		r.NoError(recoveredActiveErr,
			"chain %s: SequencerActive RPC failed on recovered CL %s "+
				"during stall observation",
			chainID, recoveredCL.ID())
		r.False(recoveredActive,
			"chain %s: recovered CL %s reports SequencerActive=true "+
				"despite op-geth lagging by >1 block — compareUnsafeHead "+
				"gate is broken, the leader is sequencing on a stale "+
				"op-geth (this would lead to a fork)",
			chainID, recoveredCL.ID())
		r.Equal(stallStartHead, stallEndHead,
			"chain %s: chain advanced from %d to %d during the %s stall "+
				"window — somebody is sequencing despite no conductor "+
				"having a usable (caught-up) op-geth. This violates the "+
				"safety property that compareUnsafeHead protects",
			chainID, stallStartHead, stallEndHead, stallObservationWindow)
		logger.Info("Stall observed: recovered leader is not sequencing",
			"chain", chainID,
			"chainHead", stallEndHead,
			"orphanHead", orphanEL.BlockRefByLabel(eth.Unsafe).Number,
			"recoveredSequencerActive", recoveredActive)

		// 8. Deactivate the fault injector. Op-geth can now accept
		//    engine_newPayload, P2P / sync closes the gap, and the
		//    next action-loop tick's compareUnsafeHead succeeds (gap
		//    eventually drops to 0; the gap==1 PostUnsafePayload
		//    bypass may also be exercised on the very last block).
		//    StartSequencer fires, SequencerActive flips to true.
		logger.Info("Deactivating EL fault injector to allow recovery",
			"chain", chainID)
		orphanInjector.Deactivate()

		// 9. Wait for sequencing to resume under the recovered leader.
		//    The action loop ticks at 1s; op-geth needs to apply the
		//    backlog of N blocks. 30s is comfortable headroom.
		const recoveryBudget = 30 * time.Second
		recoveryDeadline := time.Now().Add(recoveryBudget)
		var resumed bool
		var lastResumedActiveErr error
		for time.Now().Before(recoveryDeadline) {
			probeCtx2, cancelProbe2 := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			active, err := recoveredCL.RollupAPI().SequencerActive(probeCtx2)
			cancelProbe2()
			lastResumedActiveErr = err
			if err == nil && active {
				resumed = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		r.NoError(lastResumedActiveErr,
			"chain %s: SequencerActive RPC failed during recovery polling",
			chainID)
		r.True(resumed,
			"chain %s: recovered leader CL %s did not resume sequencing "+
				"within %s after EL fault injector was deactivated — "+
				"op-geth either failed to catch up or compareUnsafeHead "+
				"stayed stuck (FSM head moved past op-geth without a way "+
				"to backfill?)",
			chainID, recoveredCL.ID(), recoveryBudget)

		// 10. Direct uptime check: chain is actually advancing now.
		conductorhelpers.AssertChainAdvances(t, orphanEL,
			fmt.Sprintf("chain %s: post-stall recovery under recovered leader",
				chainID))

		logger.Info("Recovered leader stalled then resumed",
			"chain", chainID,
			"recoveredID", oldLeaderID)
	}
}
