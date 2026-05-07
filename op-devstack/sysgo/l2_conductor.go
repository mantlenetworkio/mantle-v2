package sysgo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum-optimism/optimism/op-conductor/conductor"
	"github.com/ethereum-optimism/optimism/op-conductor/consensus"
	"github.com/ethereum-optimism/optimism/op-conductor/health"
	"github.com/ethereum-optimism/optimism/op-conductor/metrics"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
)

// Conductor is the sysgo lifecycle wrapper around an in-process
// op-conductor instance. It keeps the underlying *conductor.OpConductor
// alive while the orchestrator is running and exposes the helpers tests
// need to assemble a multi-server raft cluster (AddVoter, Resume,
// ConsensusEndpoint, HTTPEndpoint).
//
// Tests typically don't construct this directly; they use the
// WithSysgoConductor(...) AfterDeploy option, then layer the
// cluster-formation calls on top.
type Conductor struct {
	mu sync.Mutex

	id      stack.ConductorID
	chainID eth.ChainID
	cfg     *conductor.Config
	p       devtest.P
	logger  log.Logger

	httpURL string
	service *conductor.OpConductor
}

var _ stack.Lifecycle = (*Conductor)(nil)

// ID returns the conductor's stack ID.
func (c *Conductor) ID() stack.ConductorID { return c.id }

// ChainID returns the L2 chain this conductor manages.
func (c *Conductor) ChainID() eth.ChainID { return c.chainID }

// HTTPEndpoint returns the HTTP RPC URL of the underlying op-conductor,
// or the empty string if the conductor hasn't started yet.
func (c *Conductor) HTTPEndpoint() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.httpURL
}

// ConsensusEndpoint returns the raft consensus address (host:port) that
// other servers should dial. Empty if the conductor isn't running.
func (c *Conductor) ConsensusEndpoint() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.service == nil {
		return ""
	}
	return c.service.ConsensusEndpoint()
}

// AddVoter forwards to the underlying raft consensus on this conductor.
// Used for cluster formation: the bootstrap conductor calls this for
// the non-bootstrap conductors after they've started listening.
func (c *Conductor) AddVoter(id, addr string, version uint64) error {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return errors.New("conductor not started")
	}
	return svc.AddServerAsVoter(context.Background(), id, addr, version)
}

// AddNonvoter forwards to the underlying raft consensus, registering the
// peer as a non-voting follower. The recommended raft membership-change
// protocol (also used by upstream sysgo) is to AddNonvoter first, wait
// for the follower to catch up, then promote with AddVoter — this avoids
// the new voter contributing to a quorum it has not yet caught up to.
func (c *Conductor) AddNonvoter(id, addr string, version uint64) error {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return errors.New("conductor not started")
	}
	return svc.AddServerAsNonvoter(context.Background(), id, addr, version)
}

// CommitUnsafePayload forwards a payload envelope to the conductor's raft
// FSM. Used at bootstrap to seed the FSM with the L2 genesis payload so
// that startSequencer's compareUnsafeHead succeeds (otherwise the FSM is
// empty and the leader's startSequencer returns ErrNoUnsafeHead in a
// deadlock loop, since the FSM only ever gets populated AFTER the
// sequencer produces its first payload).
func (c *Conductor) CommitUnsafePayload(ctx context.Context, payload *eth.ExecutionPayloadEnvelope) error {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return errors.New("conductor not started")
	}
	return svc.CommitUnsafePayload(ctx, payload)
}

// LatestUnsafePayload returns the latest unsafe payload envelope from this
// conductor's raft FSM. The OpConductor Go struct exposes this directly
// (op-conductor/conductor/service.go:LatestUnsafePayload) but upstream
// does NOT register it on the public RPC API (rpc/api.go) — only
// CommitUnsafePayload is exposed there. Since sysgo holds the in-process
// conductor, we can call the Go method directly. Tests use this to assert
// that the raft FSM keeps advancing under EL fault injection (the
// "FSM keeps committing payloads" half of the split-brain at unsafe head
// case study), which would otherwise be unobservable from outside.
//
// Caveat: this RPC is leader-only — RaftConsensus.LatestUnsafePayload
// (op-conductor/consensus/raft.go) issues raft.Barrier() to get a
// strongly-consistent FSM read, and Barrier requires leadership. Calling
// this on a follower returns an error. Querying the leader is enough to
// prove FSM advancement on the cluster: by raft's quorum invariant the
// leader's FSM only advances on quorum-applied log entries, so a
// follower's FSM has also applied them.
func (c *Conductor) LatestUnsafePayload(ctx context.Context) (*eth.ExecutionPayloadEnvelope, error) {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return nil, errors.New("conductor not started")
	}
	return svc.LatestUnsafePayload(ctx)
}

// IsLeader reports whether this conductor's raft node is the current leader.
func (c *Conductor) IsLeader(ctx context.Context) bool {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return false
	}
	return svc.Leader(ctx)
}

// ClusterMembership returns the current raft cluster membership as observed
// by this conductor.
func (c *Conductor) ClusterMembership(ctx context.Context) (*consensus.ClusterMembership, error) {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return nil, errors.New("conductor not started")
	}
	return svc.ClusterMembership(ctx)
}

// SequencerHealthy reports whether this conductor regards its op-node /
// op-geth pair as healthy. With the no-op health monitor injected by
// Start, this stays true for the lifetime of the conductor.
func (c *Conductor) SequencerHealthy(ctx context.Context) bool {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return false
	}
	return svc.SequencerHealthy(ctx)
}

// ServerID returns the raft server ID of this conductor (== string(c.ID())
// for sysgo conductors). Convenience accessor for cluster-formation code
// that wants the bare raft identity.
func (c *Conductor) ServerID() string {
	return string(c.id)
}

// Resume lifts the auto-pause that the bootstrap conductor enters after
// raft cluster bootstrap. Must be called once the cluster has its
// expected voter count.
func (c *Conductor) Resume() error {
	c.mu.Lock()
	svc := c.service
	c.mu.Unlock()
	if svc == nil {
		return errors.New("conductor not started")
	}
	return svc.Resume(context.Background())
}

// hydrate registers the conductor on the active stack frontend by
// dialling its HTTP RPC and wrapping it in a shim.Conductor.
func (c *Conductor) hydrate(system stack.ExtensibleSystem) {
	require := system.T().Require()
	cl, err := rpc.DialContext(system.T().Ctx(), c.httpURL)
	require.NoError(err, "failed to dial conductor RPC at %s", c.httpURL)
	system.T().Cleanup(cl.Close)

	sysCond := shim.NewConductor(shim.ConductorConfig{
		CommonConfig: shim.NewCommonConfig(system.T()),
		ID:           c.id,
		Client:       cl,
	})
	l2Net := system.L2Network(stack.L2NetworkID(c.chainID))
	l2Net.(stack.ExtensibleL2Network).AddConductor(sysCond)
}

// Start brings up the in-process op-conductor service. Idempotent — if
// the conductor is already started, the second call is a no-op (with a
// warning log).
func (c *Conductor) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.service != nil {
		c.logger.Warn("conductor already started")
		return
	}
	c.logger.Info("starting op-conductor", "id", c.id)
	// We inject a no-op health monitor so the conductor's leader-election
	// logic does not depend on op-node P2P peer count, EL P2P peers, or
	// safe-head progression — none of which are reliably present in a
	// single-process sysgo topology. The rest of the conductor stack
	// (sequencer control, raft, RPC) runs unmodified.
	hmon := newNoopHealthMonitor()
	svc, err := conductor.NewOpConductor(
		context.Background(),
		c.cfg,
		c.logger,
		metrics.NewMetrics(),
		"dev",
		nil, // sequencer control: built from cfg.NodeRPC + cfg.ExecutionRPC
		nil, // consensus: built from cfg consensus fields
		hmon,
	)
	c.p.Require().NoError(err, "failed to construct op-conductor %s", c.id)
	if err := svc.Start(context.Background()); err != nil {
		_ = svc.Stop(context.Background())
		c.p.Require().NoError(err, "failed to start op-conductor %s", c.id)
	}
	c.service = svc
	c.httpURL = svc.HTTPEndpoint()

	// Pin the OS-allocated ports back into cfg so subsequent Stop/Start
	// cycles bind the SAME HTTP and consensus addresses. This is what
	// makes the dsl.Conductor wrapper survive a conductor restart: its
	// shim caches an *rpc.Client dialled to the original c.httpURL
	// (see hydrate above), so a recovered conductor that came up on a
	// fresh ephemeral port would be unreachable from previously-built
	// dsl wrappers. The same logic applies to the raft consensus port:
	// other voters know the recovered server by its original address
	// (raft membership records it), and re-binding to a different port
	// on restart would break AppendEntries/RequestVote.
	if c.cfg.RPC.ListenPort == 0 {
		if u, err := url.Parse(c.httpURL); err == nil {
			if portStr := u.Port(); portStr != "" {
				if port, perr := strconv.Atoi(portStr); perr == nil {
					c.cfg.RPC.ListenPort = port
				}
			}
		}
	}
	if c.cfg.ConsensusPort == 0 {
		// ConsensusEndpoint() returns "host:port" (no scheme).
		if _, portStr, err := net.SplitHostPort(svc.ConsensusEndpoint()); err == nil {
			if port, perr := strconv.Atoi(portStr); perr == nil {
				c.cfg.ConsensusPort = port
			}
		}
	}

	c.logger.Info("started op-conductor",
		"id", c.id,
		"http", c.httpURL,
		"raft", svc.ConsensusEndpoint(),
	)
}

// IsRunning reports whether the underlying op-conductor service is
// currently up. Tests use this to skip already-stopped conductors after
// destructive scenarios — calling RPC on a stopped conductor will hang
// or error since its HTTP server is gone.
func (c *Conductor) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.service != nil
}

// Stop shuts down the underlying conductor service. Safe to call
// multiple times.
func (c *Conductor) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.service == nil {
		c.logger.Warn("conductor already stopped")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.logger.Info("stopping op-conductor", "id", c.id)
	if err := c.service.Stop(ctx); err != nil {
		c.logger.Warn("conductor stop returned error", "err", err)
	}
	c.service = nil
}

// WithSysgoConductor instantiates a single in-process op-conductor and
// registers it on the orchestrator under conductorID, wiring it to the
// given L2 CL/EL pair.
//
// Cluster formation is the caller's responsibility: exactly one
// conductor in the cluster should set bootstrap=true; the others
// (bootstrap=false) need to be added as voters via the bootstrap
// conductor's AddVoter API once they're listening, then the bootstrap
// conductor needs to be Resume'd to lift the auto-pause raft applies on
// successful bootstrap.
func WithSysgoConductor(
	conductorID stack.ConductorID,
	l2NetID stack.L2NetworkID,
	l2CLID stack.L2CLNodeID,
	l2ELID stack.L2ELNodeID,
	bootstrap bool,
) stack.Option[*Orchestrator] {
	return stack.AfterDeploy(func(orch *Orchestrator) {
		p := orch.P().WithCtx(stack.ContextWithID(orch.P().Ctx(), conductorID))
		require := p.Require()

		l2Net, ok := orch.l2Nets.Get(l2NetID.ChainID())
		require.True(ok, "l2 network required")

		l2EL, ok := orch.GetL2EL(l2ELID)
		require.True(ok, "l2 EL required")

		l2CL, ok := orch.l2CLs.Get(l2CLID)
		require.True(ok, "l2 CL required")

		storageDir := filepath.Join(p.TempDir(),
			fmt.Sprintf("conductor-%s", conductorID))

		cfg := &conductor.Config{
			ConsensusAddr:           "127.0.0.1",
			ConsensusPort:           0, // OS-allocated
			ConsensusAdvertisedAddr: "",
			RaftServerID:            string(conductorID),
			RaftStorageDir:          storageDir,
			RaftBootstrap:           bootstrap,
			RaftSnapshotInterval:    1 * time.Second,
			RaftSnapshotThreshold:   1024,
			RaftTrailingLogs:        128,
			RaftHeartbeatTimeout:    1000 * time.Millisecond,
			RaftLeaderLeaseTimeout:  500 * time.Millisecond,
			NodeRPC:                 l2CL.UserRPC(),
			ExecutionRPC:            l2EL.UserRPC(),
			// Followers boot paused so they don't try to take leader
			// actions on a stopped sequencer mid-bootstrap. The cluster
			// formation Finally hook calls Resume on every conductor
			// once raft membership has converged. The bootstrap
			// conductor (raft single-voter at boot) starts unpaused.
			Paused: !bootstrap,
			HealthCheck: conductor.HealthCheckConfig{
				// MinPeerCount must be > 0 to satisfy cfg.Check();
				// runtime peer-count enforcement is bypassed by the
				// no-op health monitor injected in Conductor.Start().
				Interval:       1,
				UnsafeInterval: 60,
				SafeEnabled:    false,
				SafeInterval:   30,
				MinPeerCount:   1,
			},
			RollupCfg:      *l2Net.rollupCfg,
			RPCEnableProxy: false,
			LogConfig: oplog.CLIConfig{
				Level:  log.LevelInfo,
				Format: oplog.FormatText,
			},
			MetricsConfig: opmetrics.CLIConfig{Enabled: false},
			PprofConfig:   oppprof.CLIConfig{ListenEnabled: false},
			RPC: oprpc.CLIConfig{
				ListenAddr:  "127.0.0.1",
				ListenPort:  0, // OS-allocated
				EnableAdmin: true,
			},
		}

		cnode := &Conductor{
			id:      conductorID,
			chainID: l2NetID.ChainID(),
			cfg:     cfg,
			p:       p,
			logger:  p.Logger(),
		}
		require.True(orch.conductors.SetIfMissing(conductorID, cnode),
			fmt.Sprintf("conductor %s already exists", conductorID))
		cnode.Start()
		p.Cleanup(cnode.Stop)
	})
}

// noopHealthMonitor is a health.HealthMonitor that never produces
// updates. It satisfies the interface and lets the conductor stack run
// without needing real op-node P2P peers, safe-head progression, or EL
// P2P endpoints — none of which are reliable in a single-process sysgo
// devstack.
//
// The conductor initialises its internal `healthy` atomic to true on
// construction and only flips it on a value received from this channel.
// Since we never send, it stays true for the lifetime of the conductor.
type noopHealthMonitor struct {
	ch chan error
}

func newNoopHealthMonitor() *noopHealthMonitor {
	return &noopHealthMonitor{ch: make(chan error)}
}

func (n *noopHealthMonitor) Subscribe() <-chan error         { return n.ch }
func (n *noopHealthMonitor) Start(ctx context.Context) error { return nil }
func (n *noopHealthMonitor) Stop() error {
	// Channel is intentionally not closed: closing would cause the
	// conductor's loop to receive a zero-value (nil error => "healthy")
	// repeatedly. The conductor's Stop tears down its own goroutines
	// via shutdownCtx, so an unread channel is harmless.
	return nil
}

var _ health.HealthMonitor = (*noopHealthMonitor)(nil)
