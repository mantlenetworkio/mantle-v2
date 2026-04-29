package sysgo

import (
	"os"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	nodeConfig "github.com/ethereum-optimism/optimism/op-node/config"
	nodeSync "github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type L2CLNode interface {
	hydrate(system stack.ExtensibleSystem)
	stack.Lifecycle
	UserRPC() string
	InteropRPC() (endpoint string, jwtSecret eth.Bytes32)
}

type L2CLConfig struct {
	// SyncMode to run, if this is a sequencer
	SequencerSyncMode nodeSync.Mode
	// SyncMode to run, if this is a verifier
	VerifierSyncMode nodeSync.Mode

	// SafeDBPath is the path to the safe DB to use. Disabled if empty.
	SafeDBPath string

	IsSequencer  bool
	IndexingMode bool

	// EnableReqRespSync is the flag to enable/disable req-resp sync.
	EnableReqRespSync bool

	// UseReqRespSync controls whether to use the req-resp sync protocol. EnableReqRespSync == false && UseReqRespSync == true is not allowed, and node will fail to start.
	UseReqRespSync bool

	// NoDiscovery is the flag to enable/disable discovery
	NoDiscovery bool

	// UnsafeOnly is the flag to disable derivation
	SequencerUnsafeOnly bool
	VerifierUnsafeOnly  bool

	// SequencerStopped controls whether the sequencer starts in the stopped
	// state. Required to be true when ConductorEnabled is set so the
	// conductor (not the op-node) decides which voter actually sequences.
	SequencerStopped bool

	// ConductorEnabled mirrors op-node config.Config.ConductorEnabled.
	// When true, IsSequencer must also be true and SequencerStopped should be
	// true at startup; the op-conductor will start/stop the sequencer based
	// on raft leadership.
	ConductorEnabled bool

	// ConductorRpc is the lazy resolver for the op-conductor RPC endpoint
	// this op-node should consult. The resolver is invoked lazily by the
	// op-node ConductorClient on its first leader query, with retry, so the
	// conductor does not have to be running yet at op-node start.
	ConductorRpc nodeConfig.ConductorRPCFunc

	// ConductorRpcTimeout caps the per-call timeout the op-node uses when
	// reaching out to the conductor RPC.
	ConductorRpcTimeout time.Duration
}

func L2CLSequencer() L2CLOption {
	return L2CLOptionFn(func(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
		cfg.IsSequencer = true
	})
}

func L2CLIndexing() L2CLOption {
	return L2CLOptionFn(func(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
		cfg.IndexingMode = true
	})
}

func L2CLVerifierDisableUnsafeOnly() L2CLOption {
	return L2CLOptionFn(func(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
		cfg.VerifierUnsafeOnly = false
	})
}

// L2CLSequencerStopped marks the op-node so it boots in stopped-sequencer
// mode. Combined with L2CLWithConductor, this is the standard configuration
// for a conductor cluster member: the conductor RPC drives sequencer
// start/stop transitions instead of the op-node config.
func L2CLSequencerStopped() L2CLOption {
	return L2CLOptionFn(func(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
		cfg.SequencerStopped = true
	})
}

// L2CLWithConductor wires the op-node to the op-conductor RPC endpoint
// resolved lazily by getter. Sets ConductorEnabled=true and the matching
// ConductorRpcTimeout (defaults to 1s if timeout==0). Must be combined with
// L2CLSequencer (op-node enforces SequencerEnabled when ConductorEnabled).
func L2CLWithConductor(getter nodeConfig.ConductorRPCFunc, timeout time.Duration) L2CLOption {
	return L2CLOptionFn(func(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
		cfg.ConductorEnabled = true
		cfg.ConductorRpc = getter
		if timeout <= 0 {
			timeout = time.Second
		}
		cfg.ConductorRpcTimeout = timeout
	})
}

func DefaultL2CLConfig() *L2CLConfig {
	return &L2CLConfig{
		SequencerSyncMode:   nodeSync.CLSync,
		VerifierSyncMode:    nodeSync.CLSync,
		SafeDBPath:          "",
		IsSequencer:         false,
		IndexingMode:        false,
		EnableReqRespSync:   true,
		UseReqRespSync:      true,
		NoDiscovery:         false,
		SequencerUnsafeOnly: false,
		VerifierUnsafeOnly:  false,
	}
}

type L2CLOption interface {
	Apply(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig)
}

// WithGlobalL2CLOption applies the L2CLOption to all L2CLNode instances in this orchestrator
func WithGlobalL2CLOption(opt L2CLOption) stack.Option[*Orchestrator] {
	return stack.BeforeDeploy(func(o *Orchestrator) {
		o.l2CLOptions = append(o.l2CLOptions, opt)
	})
}

type L2CLOptionFn func(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig)

var _ L2CLOption = L2CLOptionFn(nil)

func (fn L2CLOptionFn) Apply(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
	fn(p, id, cfg)
}

// L2CLOptionBundle a list of multiple L2CLOption, to all be applied in order.
type L2CLOptionBundle []L2CLOption

var _ L2CLOption = L2CLOptionBundle(nil)

func (l L2CLOptionBundle) Apply(p devtest.P, id stack.L2CLNodeID, cfg *L2CLConfig) {
	for _, opt := range l {
		p.Require().NotNil(opt, "cannot Apply nil L2CLOption")
		opt.Apply(p, id, cfg)
	}
}

// WithL2CLNode adds the default type of L2 CL node.
// The default can be configured with DEVSTACK_L2CL_KIND.
// Tests that depend on specific types can use options like WithKonaNode and WithOpNode directly.
func WithL2CLNode(l2CLID stack.L2CLNodeID, l1CLID stack.L1CLNodeID, l1ELID stack.L1ELNodeID, l2ELID stack.L2ELNodeID, opts ...L2CLOption) stack.Option[*Orchestrator] {
	switch os.Getenv("DEVSTACK_L2CL_KIND") {
	case "kona":
		return WithKonaNode(l2CLID, l1CLID, l1ELID, l2ELID, opts...)
	case "supernode":
		return WithSuperNode(l2CLID, l1CLID, l1ELID, l2ELID, opts...)
	default:
		return WithOpNode(l2CLID, l1CLID, l1ELID, l2ELID, opts...)
	}
}
