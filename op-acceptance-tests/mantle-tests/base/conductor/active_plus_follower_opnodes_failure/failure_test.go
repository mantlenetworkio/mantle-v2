package active_plus_follower_opnodes_failure

import (
	"context"
	"strings"
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

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestActivePlusFollowerOpNodesFailureAndRecovery is the entry point
// for the paired 2-of-3 op-node failure + recovery scenario. Both
// halves share a single orchestrator boot. "Failure" stops the active
// sequencer's op-node + one follower op-node and asserts that the
// surviving voter (C) ends up actively sequencing under the
// production health monitor + escape-hatch path — see the docstring
// on runFailure for the full reasoning, including the sysgo-specific
// raft-scheduling caveat that requires an explicit leadership-transfer
// nudge to land leadership on C. "Recovery" restarts both crashed
// op-nodes and asserts the cluster returns to the pre-failure healthy
// 3-member baseline.
//
// Subtests run in declaration order; "Recovery" requires the post-
// "Failure" state and must not be invoked in isolation. To run the
// recovery half:
//
//	go test -run TestActivePlusFollowerOpNodesFailureAndRecovery
//
// which executes both subtests in order.
func TestActivePlusFollowerOpNodesFailureAndRecovery(gt *testing.T) {
	parentT := devtest.SerialT(gt)
	sys := presets.NewMantleMinimalWithFaultyConductors(parentT)

	// Suite-wide pre-failure baseline.
	for chainID, conductors := range sys.ConductorSets {
		conductorhelpers.RequireHealthyConductorCluster(parentT, sys.L2Chain, chainID, conductors)
	}

	// State carried across subtests: which follower CL was crashed
	// alongside sys.L2CL during Failure, so Recovery can restart the
	// same one. Keyed by chain ID for multi-chain robustness.
	deadFollowerCLs := map[stack.L2NetworkID]stack.L2CLNodeID{}

	gt.Run("Failure", func(sgt *testing.T) {
		runFailure(devtest.SerialT(sgt), sys, deadFollowerCLs)
	})
	if gt.Failed() {
		return
	}
	gt.Run("Recovery", func(sgt *testing.T) {
		runRecovery(devtest.SerialT(sgt), sys, deadFollowerCLs)
	})
}

// runFailure exercises op-conductor README failure scenario #3,
// sub-scenario "Scenario #1" (op-conductor/README.md ~line 65):
//
//	"2 or more sequencer down (regardless of if active sequencer is
//	 healthy): Scenario #1: active sequencer and 1 follower are down.
//
//	 Solution: At this point, the Conductor will notice the sequencer
//	 not being healthy and start to transfer leadership to another
//	 node."
//
// What this test pins down — sysgo-faithful interpretation
//
// Under a real-world P2P topology (validators, batchers, RPC peers,
// etc.), the surviving healthy follower would have peers from sources
// other than the two crashed sequencer op-nodes, so its
// SequencerHealthMonitor's MinPeerCount check would still pass and
// the README's "transfer leadership to another node" promise would
// resolve to that healthy follower starting sequencing.
//
// Under the sysgo static mesh (op-devstack/sysgo/mantle_system_conductor.go
// wires only A↔B, A↔C, B↔C), killing op-nodes A and B leaves C with
// zero peers. With cfg.HealthCheck.MinPeerCount=1
// (op-devstack/sysgo/l2_conductor.go), C's conductor's
// SequencerHealthMonitor flips C to unhealthy on its next tick.
//
// Structural fix-point we want to land on
//
// We want C to end up actively sequencing — the same end state as
// scenario two_follower_opnodes_failure's survivor (A in that test):
//
//   - C's prevState becomes (F,F,F) once C's monitor flips C
//     unhealthy after its op-node loses both static-mesh peers
//     (peer count → 0; hcerr=ErrSequencerNotHealthy).
//   - When raft leadership lands on C, C's status transitions to
//     (T,F,F): leader=T from the raft event, healthy=F (still 0
//     peers), active=F (not sequencing yet).
//   - Action loop dispatch on (T,F,F) at op-conductor service.go:747
//     evaluates the escape-hatch guard:
//       !prev.leader && !prev.active && hcerr != ConnectionDown
//     With prev=(F,F,F) and hcerr=ErrSequencerNotHealthy, all three
//     conjuncts are true → startSequencer fires. C's op-node is
//     alive, so startSequencer succeeds; C latches into active.
//   - On every subsequent tick, status=(T,F,T) with frozen
//     prevState=(T,F,F) (the early-return-on-err at service.go:803
//     prevents prev from advancing past (T,F,F) once
//     shouldWaitForHealthRecovery sets err on the (T,F,T) branch) →
//     shouldWaitForHealthRecovery returns TRUE → C remains
//     sequencing despite peer-count=0. The chain advances on C's EL.
//
// Why we drive leadership to C explicitly
//
// The structural fix-point above relies on raft's TransferLeader
// eventually round-robining through the dead-op-node voters and
// landing on C. Under sysgo's deterministic raft scheduling, this
// has been observed to ping-pong between the two dead-op-node voters
// (A and B) every ~10 s — each of A's and B's stopSequencer calls
// against its dead op-node times out, then the (leader,!healthy,active)
// multierror branch at service.go:761–786 fires transferLeader which
// hashicorp raft routes to the alphabetically-next voter. With the
// 3-voter mesh A,B,C in our config, raft picks {A↔B} and
// never organically lands on C within practical test budgets.
//
// To still verify the escape-hatch fix-point on C, we explicitly
// drive raft leadership to C via TransferLeaderToServer once C's
// prevState has settled to (F,F,F). This is the same intervention an
// operator would apply in production after observing the cluster
// oscillating between dead-op-node voters without converging — it
// surfaces the structural property of the production action loop
// (escape hatch + shouldWaitForHealthRecovery latch), independent of
// raft's particular scheduling choices.
//
// State left for the recovery subtest:
//   - sys.L2CL: stopped (was the active sequencer at test start).
//   - One follower op-node: stopped (the dead follower we picked).
//     Recorded in deadFollowerCLs[chainID] so Recovery knows which
//     one to restart.
//   - Conductors: all 3 still running (we only killed op-nodes).
//   - C is latched into escape-hatch sequencing; the chain is
//     advancing on its EL.
func runFailure(t devtest.T, sys *presets.MantleMinimalWithFaultyConductors, deadFollowerCLs map[stack.L2NetworkID]stack.L2CLNodeID) {
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestActivePlusFollowerOpNodesFailureAndRecovery/Failure",
	)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; 2-of-3 failover test "+
				"needs >= 3 to exercise the peer-degraded state",
				chainID, len(conductors))
			continue
		}

		// 1+2. Identify A (active sequencer) and sanity-probe its state.
		activeCLID := sys.L2CL.Escape().ID()
		activeProbeCtx, cancelActiveProbe := context.WithTimeout(
			t.Ctx(), 3*time.Second)
		isActive, err := sys.L2CL.Escape().RollupAPI().SequencerActive(
			activeProbeCtx)
		cancelActiveProbe()
		r.NoError(err,
			"chain %s: SequencerActive RPC failed on supposed leader %s",
			chainID, activeCLID)
		r.True(isActive,
			"chain %s: sys.L2CL %s reports SequencerActive=false at "+
				"test start; the failover scenario requires A to be "+
				"actually sequencing before we kill it",
			chainID, activeCLID)

		oldLeaderInfo := conductors[0].FetchLeader()
		r.NotNil(oldLeaderInfo,
			"chain %s: cluster has no leader before failover", chainID)
		oldLeaderID := oldLeaderInfo.ID
		r.NotEmpty(oldLeaderID, "chain %s: empty leader ID", chainID)

		// Find the dsl.Conductor for the old leader and pick ONE
		// follower as the "dead" arm. The remaining voter becomes the
		// "lone survivor" — under the sysgo static mesh it will lose
		// both peers and be reported unhealthy by its own health
		// monitor (MinPeerCount=1).
		membership := conductors[0].FetchClusterMembership()

		var oldLeaderDsl, deadFollowerDsl, survivorDsl *dsl.Conductor
		var deadFollowerID, survivorID string
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			if id == oldLeaderID {
				oldLeaderDsl = c
				continue
			}
			isVoter := false
			for _, mi := range membership.Servers {
				if mi.ID == id && mi.Suffrage == consensus.Voter {
					isVoter = true
					break
				}
			}
			if !isVoter {
				continue
			}
			if deadFollowerDsl == nil {
				deadFollowerDsl = c
				deadFollowerID = id
				continue
			}
			if survivorDsl == nil {
				survivorDsl = c
				survivorID = id
				continue
			}
		}
		r.NotNil(oldLeaderDsl,
			"chain %s: dsl.Conductor for old leader %s not found",
			chainID, oldLeaderID)
		r.NotNil(deadFollowerDsl,
			"chain %s: could not find a follower to label as the "+
				"'dead' arm of the 2-of-3 failure", chainID)
		r.NotNil(survivorDsl,
			"chain %s: could not find a second follower as 'lone "+
				"survivor' — need exactly 1 leader + 1 dead follower "+
				"+ 1 lone survivor", chainID)

		// Resolve the survivor's raft ServerInfo (ID + Addr) from
		// membership. We need the Addr to call TransferLeaderToServer
		// against it later.
		var survivorServerInfo consensus.ServerInfo
		for _, mi := range membership.Servers {
			if mi.ID == survivorID {
				survivorServerInfo = mi
				break
			}
		}
		r.NotEmpty(survivorServerInfo.ID,
			"chain %s: survivor %s not found in raft membership; "+
				"cannot resolve its consensus address",
			chainID, survivorID)
		r.NotEmpty(survivorServerInfo.Addr,
			"chain %s: survivor %s has empty raft address in "+
				"membership; cannot target leadership transfer",
			chainID, survivorID)

		// 3. Resolve the (CL, EL) pairs we need.
		deadFollowerCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, deadFollowerID)
		r.NotNil(deadFollowerCL,
			"chain %s: could not locate L2CL paired with dead "+
				"follower conductor %s", chainID, deadFollowerID)
		survivorCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, survivorID)
		r.NotNil(survivorCL,
			"chain %s: could not locate L2CL paired with survivor %s",
			chainID, survivorID)
		survivorEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, survivorID)
		r.NotNil(survivorEL,
			"chain %s: could not locate L2EL paired with survivor %s",
			chainID, survivorID)

		deadFollowerCLDsl := dsl.NewL2CLNode(deadFollowerCL, sys.ControlPlane)

		baseline := sys.L2EL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("Pre-failure cluster state",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"deadFollowerID", deadFollowerID,
			"survivorID", survivorID,
			"survivorRaftAddr", survivorServerInfo.Addr,
			"activeCL", activeCLID,
			"baselineUnsafe", baseline)

		// 4. Stop A's op-node AND the chosen dead-follower's op-node.
		//    We deliberately do NOT register a t.Cleanup restart —
		//    the recovery subtest owns the restart-and-rejoin proof.
		sys.L2CL.Stop()
		deadFollowerCLDsl.Stop()
		// Record the dead follower CL ID for Recovery to pick up.
		// Done immediately after Stop so that even if subsequent
		// assertions fail, the parent test's gt.Failed() short-circuit
		// will skip Recovery (no risk of running Recovery against a
		// poisoned cluster).
		deadFollowerCLs[chainID] = deadFollowerCL.ID()

		// 5. Wait for C's prevState to settle to (F,F,F). C's monitor
		//    ticks every cfg.HealthCheck.Interval (1s in sysgo). After
		//    A's and B's op-nodes are stopped, C's static-mesh peer
		//    count drops to 0; the next monitor tick flips C unhealthy
		//    (hcerr=ErrSequencerNotHealthy), the next action tick
		//    advances C's prev from (F,T,F) to (F,F,F). Use
		//    SequencerHealthy=false on C as a proxy observable — by the
		//    time the RPC returns false, at least one monitor tick has
		//    fired with the unhealthy verdict.
		const cSettleBudget = 15 * time.Second
		cSettleEnd := time.Now().Add(cSettleBudget)
		var cSettled bool
		for time.Now().Before(cSettleEnd) {
			cHealthCtx, cancelCHealth := context.WithTimeout(
				t.Ctx(), 1*time.Second)
			cHealthy, herr := survivorDsl.Escape().RpcAPI().
				SequencerHealthy(cHealthCtx)
			cancelCHealth()
			if herr == nil && !cHealthy {
				cSettled = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		r.True(cSettled,
			"chain %s: survivor conductor %s never reported "+
				"SequencerHealthy=false within %s after killing both "+
				"sequencer-arm op-nodes — peer-count cascade did not "+
				"reach the survivor; without (F,F,F) prev state the "+
				"escape hatch on C cannot fire",
			chainID, survivorID, cSettleBudget)

		// 6. Drive raft leadership to C explicitly. We retry within a
		//    budget because:
		//      - the current leader (raft) ping-pongs between A and B
		//        each ~10 s as their stopSequencer calls time out;
		//      - the conductor we observed as leader at FetchLeader
		//        time may already have lost leadership by the time the
		//        TransferLeaderToServer RPC reaches it (ErrNotLeader);
		//      - the actual transfer applied via raft can also race
		//        with a re-election triggered by an A↔B oscillation.
		//    We swallow individual TransferLeaderToServer errors and
		//    re-poll FetchLeader on the next iteration.
		const transferBudget = 60 * time.Second
		transferEnd := time.Now().Add(transferBudget)
		var cBecameLeader bool
		var lastTransferErr error
		for time.Now().Before(transferEnd) {
			leaderCtx, cancelLeaderCtx := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			curLeader, lerr := conductors[0].Escape().RpcAPI().
				LeaderWithID(leaderCtx)
			cancelLeaderCtx()
			if lerr != nil || curLeader == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if curLeader.ID == survivorID {
				cBecameLeader = true
				break
			}
			var leaderDsl *dsl.Conductor
			for _, c := range conductors {
				id := strings.TrimPrefix(
					c.String(), stack.ConductorKind.String()+"-")
				if id == curLeader.ID {
					leaderDsl = c
					break
				}
			}
			if leaderDsl == nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			transferCtx, cancelTransfer := context.WithTimeout(
				t.Ctx(), 5*time.Second)
			lastTransferErr = leaderDsl.Escape().RpcAPI().
				TransferLeaderToServer(transferCtx,
					survivorServerInfo.ID, survivorServerInfo.Addr)
			cancelTransfer()
			// Allow raft a moment to apply the transfer before we
			// re-poll. raft.LeadershipTransferToServer blocks until
			// the transfer either succeeds or times out, so this sleep
			// is conservative — usually IsLeader on C is already true
			// by the time the RPC returns.
			time.Sleep(1 * time.Second)
			isLeaderCtx, cancelIsLeader := context.WithTimeout(
				t.Ctx(), 1*time.Second)
			isCLeader, ilErr := survivorDsl.Escape().RpcAPI().
				Leader(isLeaderCtx)
			cancelIsLeader()
			if ilErr == nil && isCLeader {
				cBecameLeader = true
				break
			}
		}
		r.True(cBecameLeader,
			"chain %s: failed to drive raft leadership to survivor %s "+
				"within %s — last TransferLeaderToServer error: %v. "+
				"Without leadership on C the escape-hatch path at "+
				"op-conductor service.go:747 cannot fire and C will "+
				"never start sequencing",
			chainID, survivorID, transferBudget, lastTransferErr)

		// 7. Wait for the chain to advance on C's EL. Once C is leader
		//    with prev=(F,F,F) and hcerr=ErrSequencerNotHealthy, the
		//    escape-hatch guard at service.go:747 fires startSequencer
		//    on C's next action tick. C's op-node is alive →
		//    startSequencer succeeds → C produces blocks → C's EL
		//    advances. We give this one block-time worth of margin
		//    plus a 30s budget for action-loop timing.
		const advanceBudget = 30 * time.Second
		advanceEnd := time.Now().Add(advanceBudget)
		baselineProgress := survivorEL.BlockRefByLabel(eth.Unsafe).Number
		var observedProgress uint64
		for time.Now().Before(advanceEnd) {
			observedProgress = survivorEL.BlockRefByLabel(eth.Unsafe).Number
			if observedProgress >= baselineProgress+1 {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		r.GreaterOrEqual(observedProgress-baselineProgress, uint64(1),
			"chain %s: survivor EL %s did not advance past baseline+1 "+
				"(baseline=%d, observed=%d) within %s after raft "+
				"leadership landed on the survivor (%s) — the "+
				"escape-hatch path at op-conductor service.go:747 "+
				"should fire startSequencer once C's prev=(F,F,F) and "+
				"hcerr=ErrSequencerNotHealthy. If this assertion "+
				"fails, the action loop is not taking the expected "+
				"branch on C's first leader tick",
			chainID, survivorEL.Escape().ID(),
			baselineProgress, observedProgress, advanceBudget,
			survivorID)

		// Cross-check via the survivor's RollupAPI that
		// SequencerActive==true at the latched moment. This is a
		// direct probe of op-node's driver.IsActive — independent of
		// whether the EL was advancing for some other reason.
		probeCtx, cancelProbe := context.WithTimeout(
			t.Ctx(), 3*time.Second)
		survivorActive, surErr := survivorCL.RollupAPI().
			SequencerActive(probeCtx)
		cancelProbe()
		r.NoError(surErr,
			"chain %s: SequencerActive RPC failed on survivor CL %s",
			chainID, survivorCL.ID())
		r.True(survivorActive,
			"chain %s: survivor CL %s reports SequencerActive=false "+
				"after observing chain advancement on its EL — RollupAPI "+
				"and EL views disagree on whether the survivor is "+
				"sequencing; this would indicate the EL advanced from "+
				"some other source (impossible: every other op-node is "+
				"dead) or that op-node already stopped between the EL "+
				"poll and this probe (would imply the latch is unstable)",
			chainID, survivorCL.ID())

		// 8. Final state log so Recovery diagnostics have a reference.
		head := survivorEL.BlockRefByLabel(eth.Unsafe).Number
		logger.Info("README scenario #3a verified (sysgo-faithful, "+
			"with explicit transferLeader→C nudge): production health "+
			"monitor detected op-node failures, leadership was driven "+
			"to the lone live-op-node survivor, and the survivor "+
			"latched into escape-hatch sequencing via "+
			"shouldWaitForHealthRecovery — the chain advances on the "+
			"survivor's EL despite peer-count=0",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"deadFollowerID", deadFollowerID,
			"survivorID", survivorID,
			"baselineUnsafe", baseline,
			"survivorUnsafe", head)
	}
}
