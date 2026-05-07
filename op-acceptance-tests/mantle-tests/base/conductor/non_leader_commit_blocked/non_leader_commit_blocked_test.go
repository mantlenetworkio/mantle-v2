package non_leader_commit_blocked

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

// TestConductorBlocksOrphanSequencerCommit exercises op-conductor
// README failure scenario #2:
//
//	"1 sequencer temporarily down, we transferred leadership to
//	 another sequencer, but it came back after leadership transfer
//	 succeeded and still be in sequencing mode"
//
// README solution quote:
//
//	"commit latest unsafe block to conductor, if node is not leader,
//	 commit fails and the block won't be gossiped out, this prevents
//	 any p2p blocks going out to the network."
//
// This test validates the gating mechanism — without it, an orphan
// sequencer (one whose conductor lost leadership while it kept
// sequencing) could leak unsafe blocks into p2p, causing forks at the
// network layer.
//
// Procedure:
//
//  1. Capture the current raft leader (A) and pick a transfer target (B).
//  2. Transfer leadership A → B and wait until B is the new leader.
//  3. Harvest a real, SSZ-marshallable payload envelope from B's FSM
//     (LatestUnsafePayload is leader-only via raft.Barrier — calling
//     it on B works because B is now the leader). Using a real
//     envelope ensures we exercise the leader check inside raft.Apply
//     rather than failing earlier on SSZ marshalling.
//  4. Call CommitUnsafePayload on A's still-alive conductor with that
//     envelope. A is no longer leader, so raft.Apply must return a
//     wrapped raft.ErrNotLeader. The op-conductor wrapper at
//     consensus/raft.go:299-313 surfaces this as
//     "failed to apply payload envelope: not leader" (or similar) —
//     we just assert that any error is returned. If commit succeeded,
//     the orphan-sequencer gate is broken.
//  5. Transfer leadership back B → A so the test ends in the same
//     starting layout it began in (kept for parity with the original
//     test even though each subdirectory now gets a fresh
//     orchestrator).
func TestConductorBlocksOrphanSequencerCommit(gt *testing.T) {
	t := devtest.SerialT(gt)
	logger := testlog.Logger(t, log.LevelInfo).With(
		"Test", "TestConductorBlocksOrphanSequencerCommit",
	)

	sys := presets.NewMantleMinimalWithFaultyConductors(t)
	r := t.Require()

	if len(sys.SysgoConductors) < 3 {
		t.Skipf("orphan-commit gating needs in-process conductors and "+
			">=3 voters; got %d sysgo conductors "+
			"(kurtosis/persistent backends do not expose the in-process "+
			"CommitUnsafePayload Go method)", len(sys.SysgoConductors))
	}

	for chainID, conductors := range sys.ConductorSets {
		if len(conductors) < 3 {
			t.Skipf("chain %s has %d conductors; orphan-commit test "+
				"needs >= 3 voters to pick a transfer target",
				chainID, len(conductors))
			continue
		}

		// Suite-wide baseline.
		conductorhelpers.RequireHealthyConductorCluster(t, sys.L2Chain, chainID, conductors)

		// 1. Capture the current leader.
		oldLeaderInfo := conductors[0].FetchLeader()
		r.NotNil(oldLeaderInfo,
			"chain %s: cluster has no leader before transfer", chainID)
		oldLeaderID := oldLeaderInfo.ID
		r.NotEmpty(oldLeaderID,
			"chain %s: empty leader ID before transfer", chainID)

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
			// Pick the first non-leader voter as the transfer target.
			// Skip non-voters (e.g. nonvoter members during a rolling
			// upgrade), since transferring to a nonvoter is invalid.
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

		oldLeaderSysgo, ok := sys.SysgoConductors[stack.ConductorID(oldLeaderID)]
		r.True(ok,
			"chain %s: in-process conductor for old leader %s missing",
			chainID, oldLeaderID)
		newLeaderSysgo, ok := sys.SysgoConductors[stack.ConductorID(newLeaderInfo.ID)]
		r.True(ok,
			"chain %s: in-process conductor for new leader %s missing",
			chainID, newLeaderInfo.ID)

		logger.Info("transferring leadership for orphan-commit test",
			"chain", chainID,
			"from", oldLeaderID,
			"to", newLeaderInfo.ID)

		// 2. Transfer leadership and wait until both ends agree.
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

		// 3. Harvest a real payload envelope from the new leader's FSM.
		// LatestUnsafePayload requires raft.Barrier which requires
		// leadership, so this is only callable on the leader.
		harvestCtx, cancelHarvest := context.WithTimeout(t.Ctx(), 5*time.Second)
		defer cancelHarvest()
		env, err := newLeaderSysgo.LatestUnsafePayload(harvestCtx)
		r.NoError(err,
			"chain %s: failed to read FSM unsafe payload from new "+
				"leader %s", chainID, newLeaderInfo.ID)
		r.NotNil(env,
			"chain %s: new leader %s has nil unsafe payload",
			chainID, newLeaderInfo.ID)

		// 4. Try to commit that envelope through the OLD leader (now a
		// follower). raft.Apply on a non-leader returns
		// raft.ErrNotLeader, which the op-conductor wrapper surfaces
		// as a wrapped error from CommitUnsafePayload. Any non-nil
		// error is sufficient to validate the gate; we additionally
		// log the message so failures are diagnosable.
		commitCtx, cancelCommit := context.WithTimeout(t.Ctx(), 5*time.Second)
		defer cancelCommit()
		err = oldLeaderSysgo.CommitUnsafePayload(commitCtx, env)
		r.Error(err,
			"chain %s: orphan commit on non-leader %s succeeded; "+
				"the gate that prevents orphan-sequencer p2p gossip is "+
				"broken", chainID, oldLeaderID)
		logger.Info("orphan commit correctly rejected by non-leader",
			"chain", chainID,
			"nonLeaderID", oldLeaderID,
			"err", err.Error())

		// 5. Restore: transfer leadership back to the original leader so
		// the test ends in a healthy steady state.
		newLeaderDsl.TransferLeadershipTo(*oldLeaderInfo)
		restoreDeadline := time.Now().Add(10 * time.Second)
		var restored bool
		for time.Now().Before(restoreDeadline) {
			if oldLeaderDsl.IsLeader() && !newLeaderDsl.IsLeader() {
				restored = true
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		r.True(restored,
			"chain %s: failed to restore leadership to %s within 10s",
			chainID, oldLeaderID)

		// 6. Give the restored leader's conductor action loop time to
		// drive startSequencer on its op-node before asserting chain
		// progress.
		time.Sleep(3 * time.Second)

		// 7. Final guarantee: leadership restoration didn't merely
		//    restore raft state — the chain is actually producing blocks
		//    again under the original leader. sys.L2EL was selected at
		//    preset hydration via match.WithSequencerActive against the
		//    original leader's CL, which is again the active sequencer
		//    at this point.
		conductorhelpers.AssertChainAdvances(t, sys.L2EL,
			fmt.Sprintf("chain %s: post-orphan-commit leadership restored", chainID))

		logger.Info("orphan-commit gating verified",
			"chain", chainID,
			"oldLeaderID", oldLeaderID,
			"newLeaderID", newLeaderInfo.ID)
	}
}
