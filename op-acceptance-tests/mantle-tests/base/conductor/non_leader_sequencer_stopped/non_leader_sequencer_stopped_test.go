package non_leader_sequencer_stopped

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/testlog"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/conductor/conductorhelpers"
)

// TestNonLeaderHealthySequencerIsStopped exercises the *second* half of
// op-conductor README scenario #2 (op-conductor/README.md:64):
//
//	"1 sequencer temporarily down, we transferred leadership to another
//	 sequencer, but it came back after leadership transfer succeeded and
//	 still be in sequencing mode"
//
// README solution quote — the half this test covers:
//
//	"for control loop health update handling logic, stop sequencer when
//	 it's not leader but healthy."
//
// What this test pins down
//
// The exact case-statement at op-conductor/conductor/service.go:740-742:
//
//	case !status.leader && status.healthy && status.active:
//	    // sequencer is not leader, healthy, and active, stop it.
//	    err = oc.stopSequencer()
//
// Procedure
//
//  1. Identify the current active sequencer (sys.L2CL via the preset's
//     match.WithSequencerActive) and the current raft leader.
//  2. Sanity-probe that the active sequencer's op-node really does
//     report SequencerActive=true.
//  3. Pick a healthy voter follower as the leadership transfer target.
//  4. Transfer leadership old-leader → new-leader and wait for raft to
//     commit the change on both ends.
//  5. Poll the OLD leader's op-node SequencerActive until it flips to
//     false. This is the README guarantee.
//  6. Sanity-check that the NEW leader's SequencerActive flipped to
//     true.
func TestNonLeaderHealthySequencerIsStopped(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestNonLeaderHealthySequencerIsStopped",
	)

	sys := presets.NewMantleMinimalWithFaultyConductors(t)
	r := t.Require()

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; orphan-stop test needs "+
				">= 3 voters to transfer leadership to a different node",
				chainID, len(conductors))
			continue
		}

		// Suite-wide baseline.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

		// 1+2. Identify and sanity-probe the active sequencer.
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
				"test start; the orphan-stop scenario requires the old "+
				"leader to actually be sequencing before we transfer "+
				"leadership away from it",
			chainID, activeCLID)

		oldLeaderInfo := conductors[0].FetchLeader()
		r.NotNil(oldLeaderInfo,
			"chain %s: cluster has no leader before transfer", chainID)
		oldLeaderID := oldLeaderInfo.ID
		r.NotEmpty(oldLeaderID, "chain %s: empty leader ID", chainID)

		// 3. Pick a voter follower as the transfer target.
		membership := conductors[0].FetchClusterMembership()

		var oldLeaderDsl, newLeaderDsl *dsl.Conductor
		var newLeaderInfo consensus.ServerInfo
		for _, c := range conductors {
			id := strings.TrimPrefix(
				c.String(), stack.ConductorKind.String()+"-")
			if id == oldLeaderID {
				oldLeaderDsl = c
				continue
			}
			if newLeaderDsl != nil {
				continue
			}
			for _, mi := range membership.Servers {
				if mi.ID == id && mi.Suffrage == consensus.Voter {
					newLeaderInfo = mi
					newLeaderDsl = c
					break
				}
			}
		}
		r.NotNil(oldLeaderDsl,
			"chain %s: dsl.Conductor for old leader %s not found",
			chainID, oldLeaderID)
		r.NotNil(newLeaderDsl,
			"chain %s: no eligible voter to transfer leadership to",
			chainID)

		oldLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, oldLeaderID)
		r.NotNil(oldLeaderCL,
			"chain %s: could not locate L2CL paired with old leader %s",
			chainID, oldLeaderID)
		newLeaderCL := conductorhelpers.CLPairedWithConductor(sys.L2Chain, newLeaderInfo.ID)
		r.NotNil(newLeaderCL,
			"chain %s: could not locate L2CL paired with new leader %s",
			chainID, newLeaderInfo.ID)

		logger.Info("Pre-transfer cluster state",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderInfo.ID,
			"activeCL", activeCLID,
			"oldLeaderCL", oldLeaderCL.ID(),
			"newLeaderCL", newLeaderCL.ID())

		// 4. Transfer leadership and wait for raft to settle on both ends.
		oldLeaderDsl.TransferLeadershipTo(newLeaderInfo)
		transferDeadline := time.Now().Add(10 * time.Second)
		var transferred bool
		for time.Now().Before(transferDeadline) {
			if newLeaderDsl.IsLeader() && !oldLeaderDsl.IsLeader() {
				transferred = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(transferred,
			"chain %s: leadership transfer %s -> %s did not complete "+
				"within 10s", chainID, oldLeaderID, newLeaderInfo.ID)

		// 5. The README guarantee: poll the OLD leader's
		//    SequencerActive until it flips to false.
		const stopDeadlineDur = 10 * time.Second
		stopDeadline := time.Now().Add(stopDeadlineDur)
		var oldStillActive bool = true
		var lastOldErr error
		for time.Now().Before(stopDeadline) {
			probeCtx, cancelProbe := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			oldStillActive, lastOldErr = oldLeaderCL.RollupAPI().
				SequencerActive(probeCtx)
			cancelProbe()
			if lastOldErr == nil && !oldStillActive {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NoError(lastOldErr,
			"chain %s: SequencerActive RPC failed on old leader CL %s "+
				"during stop-poll", chainID, oldLeaderCL.ID())
		r.False(oldStillActive,
			"chain %s, old leader CL %s: SequencerActive=true %s after "+
				"leadership transfer — README scenario #2 requires the "+
				"control-loop case (!leader, healthy, active) at "+
				"op-conductor/conductor/service.go:740-742 to call "+
				"stopSequencer(). Without it, the old node becomes an "+
				"orphan sequencer that keeps building blocks no quorum "+
				"will ever apply.",
			chainID, oldLeaderCL.ID(), stopDeadlineDur)

		// 6. Sanity: the new leader took over and is actually sequencing.
		const startDeadlineDur = 10 * time.Second
		startDeadline := time.Now().Add(startDeadlineDur)
		var newActive bool
		var lastNewErr error
		for time.Now().Before(startDeadline) {
			probeCtx, cancelProbe := context.WithTimeout(
				t.Ctx(), 2*time.Second)
			newActive, lastNewErr = newLeaderCL.RollupAPI().
				SequencerActive(probeCtx)
			cancelProbe()
			if lastNewErr == nil && newActive {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.NoError(lastNewErr,
			"chain %s: SequencerActive RPC failed on new leader CL %s",
			chainID, newLeaderCL.ID())
		r.True(newActive,
			"chain %s, new leader CL %s: SequencerActive=false %s after "+
				"leadership transfer — the chain is now leaderless. Either "+
				"raft believes %s is leader but its conductor never called "+
				"StartSequencer, or the wire-up between conductor and "+
				"op-node is broken.",
			chainID, newLeaderCL.ID(), startDeadlineDur, newLeaderInfo.ID)

		// 7. Final guarantee: SequencerActive=true on the new leader is
		//    necessary but not sufficient. Verify the new leader's EL is
		//    actually producing blocks.
		newLeaderEL := conductorhelpers.ELPairedWithConductor(sys.L2Chain, newLeaderInfo.ID)
		r.NotNil(newLeaderEL,
			"chain %s: could not locate L2EL paired with new leader %s",
			chainID, newLeaderInfo.ID)
		conductorhelpers.AssertChainAdvances(t, newLeaderEL,
			fmt.Sprintf("chain %s: post-orphan-stop new-leader sequencing", chainID))

		logger.Info("README scenario #2 (control-loop half) verified: "+
			"non-leader healthy sequencer was stopped; new leader is "+
			"sequencing",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderInfo.ID,
			"oldLeaderCL", oldLeaderCL.ID(),
			"newLeaderCL", newLeaderCL.ID(),
			"oldSequencerActive", oldStillActive,
			"newSequencerActive", newActive)
	}
}
