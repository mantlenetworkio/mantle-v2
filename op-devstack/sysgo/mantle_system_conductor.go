package sysgo

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	nodeConfig "github.com/ethereum-optimism/optimism/op-node/config"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/params/forks"
)

// DefaultMantleConductorSystemIDs identifies every component of the
// 3-sequencer-3-conductor Mantle topology used to reproduce the
// op-conductor split-brain at unsafe head case study
// (op-conductor/INTEGRATION.md). The three conductor nodes form a single
// raft cluster that drives the three op-nodes; only the raft leader
// actually sequences. All three EL/CL pairs share the same chain genesis
// and the same SequencerP2PRole devkey (derived from chain ID), so
// payloads gossiped by the leader's op-node are accepted by every
// follower's op-node.
//
// The topology + bootstrap dance mirrors upstream sysgo's
// NewMinimalWithConductorsRuntime in op-devstack/sysgo/singlechain_variants.go,
// adapted from upstream's procedural Runtime style to Mantle's
// orchestrator-option style.
type DefaultMantleConductorSystemIDs struct {
	L1   stack.L1NetworkID
	L1EL stack.L1ELNodeID
	L1CL stack.L1CLNodeID

	L2 stack.L2NetworkID

	// Three sequencer-eligible EL/CL pairs. "A" is the raft bootstrap
	// node and the canonical batcher / proposer target — leadership
	// starts on A and the split-brain test never rotates it.
	L2ELA stack.L2ELNodeID
	L2CLA stack.L2CLNodeID
	L2ELB stack.L2ELNodeID
	L2CLB stack.L2CLNodeID
	L2ELC stack.L2ELNodeID
	L2CLC stack.L2CLNodeID

	// Three op-conductor instances forming a 3-voter raft cluster.
	CondA stack.ConductorID
	CondB stack.ConductorID
	CondC stack.ConductorID

	// Optional 4th sequencer-eligible EL/CL pair + paused op-conductor.
	// Populated unconditionally so callers can refer to the IDs, but
	// the actual sysgo nodes are only spun up when
	// DefaultMantleConductorSystemWithSpare is used. The spare's
	// conductor never gets AddVoter'd during cluster bootstrap; it
	// stays out of the raft cluster until a test explicitly adds it.
	L2ELD stack.L2ELNodeID
	L2CLD stack.L2CLNodeID
	CondD stack.ConductorID

	L2Batcher  stack.L2BatcherID
	L2Proposer stack.L2ProposerID
}

// NewDefaultMantleConductorSystemIDs allocates IDs for the conductor
// topology. The L2EL / L2CL "sequencer" key (no suffix) is reused for
// node A so that downstream code paths that look up the canonical
// sequencer (e.g. presets.NewMantleMinimal's match.WithSequencerActive)
// converge on A while it holds raft leadership.
func NewDefaultMantleConductorSystemIDs(l1ID, l2ID eth.ChainID) DefaultMantleConductorSystemIDs {
	return DefaultMantleConductorSystemIDs{
		L1:         stack.L1NetworkID(l1ID),
		L1EL:       stack.NewL1ELNodeID("l1", l1ID),
		L1CL:       stack.NewL1CLNodeID("l1", l1ID),
		L2:         stack.L2NetworkID(l2ID),
		L2ELA:      stack.NewL2ELNodeID("sequencer", l2ID),
		L2CLA:      stack.NewL2CLNodeID("sequencer", l2ID),
		L2ELB:      stack.NewL2ELNodeID("sequencer-b", l2ID),
		L2CLB:      stack.NewL2CLNodeID("sequencer-b", l2ID),
		L2ELC:      stack.NewL2ELNodeID("sequencer-c", l2ID),
		L2CLC:      stack.NewL2CLNodeID("sequencer-c", l2ID),
		L2ELD:      stack.NewL2ELNodeID("sequencer-d", l2ID),
		L2CLD:      stack.NewL2CLNodeID("sequencer-d", l2ID),
		CondA:      stack.ConductorID("a"),
		CondB:      stack.ConductorID("b"),
		CondC:      stack.ConductorID("c"),
		CondD:      stack.ConductorID("d"),
		L2Batcher:  stack.NewL2BatcherID("main", l2ID),
		L2Proposer: stack.NewL2ProposerID("main", l2ID),
	}
}

// DefaultMantleConductorSystem composes a sysgo orchestrator option that
// stands up a 3-sequencer-3-conductor Mantle minimal topology suitable
// for op-conductor cluster tests (split-brain reproducer + leadership
// transfer).
//
// Topology:
//   - 1× L1 EL + L1 CL (geth + faux beacon)
//   - 3× L2 EL + L2 CL (op-geth + op-node), all flagged as sequencer
//     candidates with SequencerStopped=true so the conductor decides
//     who actually sequences.
//   - 3× op-conductor (raft cluster). A is bootstrap=Paused:false.
//     B and C come up paused; they're added to the cluster via
//     AddNonvoter → AddVoter, and Resume() is called on every conductor
//     once cluster membership has converged. This matches upstream's
//     bootstrap dance exactly and avoids racing a stopped follower
//     sequencer during cluster formation.
//   - Static P2P peering between all three op-nodes so blocks the
//     leader gossips reach every follower's op-node and its op-geth.
//   - 1× batcher + 1× legacy proposer, both pinned to A. Tests that
//     rotate leadership should not assert on batching/proposing.
//   - 1× faucet on the L1 EL + the L2 EL of A.
func DefaultMantleConductorSystem(dest *DefaultMantleConductorSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMantleConductorSystemIDs(DefaultL1ID, DefaultL2AID)
	return defaultMantleConductorSystemOpts(&ids, dest, false)
}

// DefaultMantleConductorSystemWithSpare is identical to
// DefaultMantleConductorSystem but additionally provisions a 4th
// sequencer-eligible EL/CL pair plus a 4th op-conductor instance
// ("d"). The spare conductor is started up and reaches a paused steady
// state, but is NOT seated in the raft cluster during bootstrap — its
// AddVoter must be issued by the test. This is the topology used to
// exercise raft membership-change scenarios such as
// "start a 4th sequencer and replace one of the existing 3".
func DefaultMantleConductorSystemWithSpare(dest *DefaultMantleConductorSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMantleConductorSystemIDs(DefaultL1ID, DefaultL2AID)
	return defaultMantleConductorSystemOpts(&ids, dest, true)
}

func defaultMantleConductorSystemOpts(ids *DefaultMantleConductorSystemIDs, dest *DefaultMantleConductorSystemIDs, withSpare bool) stack.CombinedOption[*Orchestrator] {
	opt := stack.Combine[*Orchestrator]()
	opt.Add(stack.BeforeDeploy(func(o *Orchestrator) {
		o.P().Logger().Info("Setting up mantle conductor cluster",
			"l1", ids.L1.ChainID(), "l2", ids.L2.ChainID())
	}))

	opt.Add(WithMnemonicKeys(devkeys.TestMnemonic))

	opt.Add(WithMantleDeployer(),
		WithDeployerOptions(
			WithLocalContractSources(),
			WithCommons(ids.L1.ChainID()),
			WithDefaultBPOBlobSchedule,
			WithForkAtL1Genesis(forks.BPO2),
			WithPrefundedL2(ids.L1.ChainID(), ids.L2.ChainID()),
		),
		WithDeployerPipelineOption(WithL1MNT(DefaultL1MNT)),
		WithDeployerPipelineOption(WithOperatorFeeVaultRecipient(DefaultOperatorFeeVaultRecipient)),
		WithDeployerPipelineOption(WithMantlePortalPaused(false)),
	)

	opt.Add(WithL1Nodes(ids.L1EL, ids.L1CL))

	// All three L2 ELs run the same L2 chain genesis.
	opt.Add(WithL2ELNode(ids.L2ELA))
	opt.Add(WithL2ELNode(ids.L2ELB))
	opt.Add(WithL2ELNode(ids.L2ELC))
	if withSpare {
		opt.Add(WithL2ELNode(ids.L2ELD))
	}

	// Per-conductor RPC URL slots. Each op-node gets a lazy resolver that
	// blocks until its conductor's Start() populates the slot. This
	// matches the upstream pattern (sysgo singlechain_variants.go:242-256):
	// the resolver waits — with ctx-aware backoff — until the conductor
	// has bound its HTTP listener, then returns the URL.
	rpcSlotA := &atomic.Value{}
	rpcSlotB := &atomic.Value{}
	rpcSlotC := &atomic.Value{}
	rpcSlotA.Store("")
	rpcSlotB.Store("")
	rpcSlotC.Store("")
	var rpcSlotD *atomic.Value
	if withSpare {
		rpcSlotD = &atomic.Value{}
		rpcSlotD.Store("")
	}

	makeResolver := func(slot *atomic.Value, condID stack.ConductorID) nodeConfig.ConductorRPCFunc {
		return func(ctx context.Context) (string, error) {
			for {
				if v, _ := slot.Load().(string); v != "" {
					return v, nil
				}
				select {
				case <-ctx.Done():
					return "", fmt.Errorf("waiting for conductor %s rpc: %w", condID, ctx.Err())
				case <-time.After(100 * time.Millisecond):
				}
			}
		}
	}

	clOpts := func(slot *atomic.Value, condID stack.ConductorID) []L2CLOption {
		return []L2CLOption{
			L2CLSequencer(),
			L2CLSequencerStopped(),
			L2CLWithConductor(makeResolver(slot, condID), 5*time.Second),
		}
	}

	opt.Add(WithL2CLNode(ids.L2CLA, ids.L1CL, ids.L1EL, ids.L2ELA, clOpts(rpcSlotA, ids.CondA)...))
	opt.Add(WithL2CLNode(ids.L2CLB, ids.L1CL, ids.L1EL, ids.L2ELB, clOpts(rpcSlotB, ids.CondB)...))
	opt.Add(WithL2CLNode(ids.L2CLC, ids.L1CL, ids.L1EL, ids.L2ELC, clOpts(rpcSlotC, ids.CondC)...))
	if withSpare {
		opt.Add(WithL2CLNode(ids.L2CLD, ids.L1CL, ids.L1EL, ids.L2ELD, clOpts(rpcSlotD, ids.CondD)...))
	}

	// Stand up the three op-conductor instances. Only A bootstraps; B and
	// C come up paused. WithSysgoConductor sets Paused=!bootstrap, so
	// followers don't try to direct a stopped sequencer mid-bootstrap.
	opt.Add(WithSysgoConductor(ids.CondA, ids.L2, ids.L2CLA, ids.L2ELA, true))
	opt.Add(WithSysgoConductor(ids.CondB, ids.L2, ids.L2CLB, ids.L2ELB, false))
	opt.Add(WithSysgoConductor(ids.CondC, ids.L2, ids.L2CLC, ids.L2ELC, false))
	if withSpare {
		// The spare comes up paused like B and C, but is NOT added to
		// the raft cluster during bootstrap (see Finally below). It
		// stays a non-member with raft.NumServers=1 (only itself, from
		// its own bootstrap=false config) and a stopped sequencer
		// until a test promotes it via AddNonvoter+AddVoter and
		// Resume.
		opt.Add(WithSysgoConductor(ids.CondD, ids.L2, ids.L2CLD, ids.L2ELD, false))
	}

	// Populate the per-op-node RPC slots once each conductor is up. This
	// must run after WithSysgoConductor (which is AfterDeploy) but before
	// the cluster-bootstrap Finally hook below — Finally runs strictly
	// after AfterDeploy, so chaining a Finally that publishes URLs and
	// then a second Finally that bootstraps the cluster is safe.
	opt.Add(stack.Finally(func(orch *Orchestrator) {
		require := orch.P().Require()
		publish := func(slot *atomic.Value, condID stack.ConductorID) {
			cnode, ok := orch.conductors.Get(condID)
			require.True(ok, "conductor %s missing from registry", condID)
			url := cnode.HTTPEndpoint()
			require.NotEmpty(url, "conductor %s has empty HTTPEndpoint", condID)
			slot.Store(url)
		}
		publish(rpcSlotA, ids.CondA)
		publish(rpcSlotB, ids.CondB)
		publish(rpcSlotC, ids.CondC)
		if withSpare {
			publish(rpcSlotD, ids.CondD)
		}
	}))

	// Pairwise static P2P peering between the three op-nodes. Without
	// this, follower op-nodes never receive blocks gossiped by the
	// leader; follower op-geth never sees engine_newPayload calls; and
	// the split-brain test's totalInjections > 0 assertion fails because
	// no payloads were ever offered for the injectors to reject.
	opt.Add(WithL2CLP2PConnection(ids.L2CLA, ids.L2CLB))
	opt.Add(WithL2CLP2PConnection(ids.L2CLA, ids.L2CLC))
	opt.Add(WithL2CLP2PConnection(ids.L2CLB, ids.L2CLC))
	if withSpare {
		// The spare must also receive gossip from whichever node is
		// the active sequencer; otherwise after AddVoter+Resume its
		// EL would be too far behind to ever match the unsafe-head
		// gate inside compareUnsafeHead.
		opt.Add(WithL2CLP2PConnection(ids.L2CLA, ids.L2CLD))
		opt.Add(WithL2CLP2PConnection(ids.L2CLB, ids.L2CLD))
		opt.Add(WithL2CLP2PConnection(ids.L2CLC, ids.L2CLD))
	}

	// Batcher and proposer pinned to A. Correct as long as raft
	// leadership stays on A; tests that rotate leadership should not
	// assert on batching / proposing.
	opt.Add(WithBatcher(ids.L2Batcher, ids.L1EL, ids.L2CLA, ids.L2ELA))
	opt.Add(WithLegacyProposer(ids.L2Proposer, ids.L1EL, &ids.L2CLA, nil))

	opt.Add(WithFaucets([]stack.L1ELNodeID{ids.L1EL}, []stack.L2ELNodeID{ids.L2ELA}))

	opt.Add(WithL2MetricsDashboard())

	// Cluster-formation Finally. Mirrors upstream startConductorCluster
	// in op-devstack/sysgo/singlechain_variants.go:321-376:
	//   1. Wait for bootstrap to actually be raft leader.
	//   2. AddNonvoter → AddVoter for each follower (recommended raft
	//      membership-change protocol; non-voter is a sync step).
	//   3. Wait for cluster membership to converge to N voters.
	//   4. Resume every conductor (including the bootstrap, which is in
	//      auto-pause after single-voter bootstrap).
	//   5. Wait for SequencerHealthy on every conductor.
	opt.Add(stack.Finally(func(orch *Orchestrator) {
		require := orch.P().Require()
		ctx, cancel := context.WithTimeout(orch.P().Ctx(), 90*time.Second)
		defer cancel()

		bootstrap, ok := orch.conductors.Get(ids.CondA)
		require.True(ok, "bootstrap conductor %s missing", ids.CondA)
		nodeB, ok := orch.conductors.Get(ids.CondB)
		require.True(ok, "conductor %s missing", ids.CondB)
		nodeC, ok := orch.conductors.Get(ids.CondC)
		require.True(ok, "conductor %s missing", ids.CondC)
		members := []*Conductor{nodeB, nodeC}
		cluster := []*Conductor{bootstrap, nodeB, nodeC}

		// 1. Wait for the bootstrap conductor to actually be leader.
		err := retry.Do0(ctx, 90, retry.Fixed(500*time.Millisecond), func() error {
			if !bootstrap.IsLeader(ctx) {
				return errors.New("bootstrap conductor is not leader yet")
			}
			return nil
		})
		require.NoError(err, "bootstrap conductor never became leader")

		// 1.5. Seed the leader's FSM with the L2 genesis payload so that
		// startSequencer's compareUnsafeHead doesn't deadlock on an empty
		// FSM. Without this, the leader's action loop sees
		// (leader && healthy && !active) and tries to start the
		// sequencer, but cons.LatestUnsafePayload() returns nil and
		// compareUnsafeHead bails with ErrNoUnsafeHead. The FSM only
		// gets seeded *after* the sequencer produces its first payload,
		// which the sequencer can't do until startSequencer succeeds —
		// classic chicken-and-egg.
		//
		// We seed with the deterministic L2 genesis block: every node
		// already has it locally as its head at startup (forkchoice
		// update applied by op-node init), so when raft applies the
		// payload to followers' FSMs they accept it without needing to
		// execute it. Once leadership transfer happens later in tests,
		// the new leader's FSM already has a valid unsafeHead from raft
		// replication.
		l2Net, ok := orch.l2Nets.Get(ids.L2.ChainID())
		require.True(ok, "l2 network missing for FSM seed")
		genesisBlock := l2Net.genesis.ToBlock()
		genesisEnvelope, err := eth.BlockAsPayloadEnv(genesisBlock, l2Net.genesis.Config)
		require.NoError(err, "failed to convert L2 genesis block to payload envelope")
		err = retry.Do0(ctx, 40, retry.Fixed(250*time.Millisecond), func() error {
			return bootstrap.CommitUnsafePayload(ctx, genesisEnvelope)
		})
		require.NoError(err, "failed to seed bootstrap FSM with L2 genesis payload")

		// 2. AddNonvoter then AddVoter for each follower.
		for _, m := range members {
			err = retry.Do0(ctx, 40, retry.Fixed(250*time.Millisecond), func() error {
				return bootstrap.AddNonvoter(m.ServerID(), m.ConsensusEndpoint(), 0)
			})
			require.NoError(err, "failed to add %s as non-voter", m.ServerID())

			err = retry.Do0(ctx, 40, retry.Fixed(250*time.Millisecond), func() error {
				return bootstrap.AddVoter(m.ServerID(), m.ConsensusEndpoint(), 0)
			})
			require.NoError(err, "failed to add %s as voter", m.ServerID())
		}

		// 3. Wait for membership convergence.
		expected := len(cluster)
		err = retry.Do0(ctx, 90, retry.Fixed(500*time.Millisecond), func() error {
			m, err := bootstrap.ClusterMembership(ctx)
			if err != nil {
				return err
			}
			if len(m.Servers) != expected {
				return fmt.Errorf("expected %d voters in cluster, got %d", expected, len(m.Servers))
			}
			return nil
		})
		require.NoError(err, "cluster did not converge to %d voters", expected)

		// 4. Resume every conductor.
		for _, c := range cluster {
			err = retry.Do0(ctx, 40, retry.Fixed(250*time.Millisecond), func() error {
				return c.Resume()
			})
			require.NoError(err, "failed to resume conductor %s", c.ServerID())
		}

		// 5. Wait for SequencerHealthy on every conductor. The production
		// SequencerHealthMonitor wired in Conductor.Start() needs at
		// least Interval (1s) to tick, plus time for op-node SyncStatus
		// to succeed and the static-mesh peer count to reach
		// MinPeerCount=1; we poll generously to allow for it.
		for _, c := range cluster {
			err = retry.Do0(ctx, 60, retry.Fixed(500*time.Millisecond), func() error {
				if !c.SequencerHealthy(ctx) {
					return fmt.Errorf("conductor %s sequencer not healthy yet", c.ServerID())
				}
				return nil
			})
			require.NoError(err, "conductor %s never became healthy", c.ServerID())
		}

		// 6. Wait for the leader's op-node to actually report
		// SequencerActive==true. After Resume the leader conductor
		// enqueues an action that calls startSequencer on its op-node
		// (op-conductor/conductor/service.go action() leader/healthy/!active
		// branch). That call is async, so without this wait, downstream
		// hydration code that runs match.WithSequencerActive (e.g.
		// presets.NewMantleMinimal) can race the start and find no
		// active sequencer, skipping the test.
		l2CLs := []L2CLNode{}
		for _, id := range []stack.L2CLNodeID{ids.L2CLA, ids.L2CLB, ids.L2CLC} {
			cl, ok := orch.l2CLs.Get(id)
			require.True(ok, "l2 CL %s missing from registry", id)
			l2CLs = append(l2CLs, cl)
		}
		rollupClients := make([]*sources.RollupClient, 0, len(l2CLs))
		for _, cl := range l2CLs {
			rpcCl, err := client.NewRPC(ctx, orch.P().Logger(), cl.UserRPC(), client.WithLazyDial())
			require.NoError(err, "failed to dial l2 CL %s for SequencerActive poll", cl.UserRPC())
			rollupClients = append(rollupClients, sources.NewRollupClient(rpcCl))
		}
		err = retry.Do0(ctx, 60, retry.Fixed(500*time.Millisecond), func() error {
			for _, rc := range rollupClients {
				active, qerr := rc.SequencerActive(ctx)
				if qerr != nil {
					return qerr
				}
				if active {
					return nil
				}
			}
			return errors.New("no sequencer is active yet")
		})
		require.NoError(err, "no L2CL ever reported SequencerActive after cluster bootstrap")

		orch.P().Logger().Info("conductor cluster formed",
			"bootstrap", ids.CondA,
			"voters", []stack.ConductorID{ids.CondA, ids.CondB, ids.CondC})
	}))

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = *ids
	}))

	return opt
}
