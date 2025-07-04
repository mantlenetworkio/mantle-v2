package node

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/p2p"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/da"
	"github.com/ethereum-optimism/optimism/op-node/rollup/driver"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	oppprof "github.com/ethereum-optimism/optimism/op-service/pprof"
)

type Config struct {
	L1     L1EndpointSetup
	L2     L2EndpointSetup
	L2Sync L2SyncEndpointSetup

	Beacon L1BeaconEndpointSetup

	Driver driver.Config

	Rollup rollup.Config

	// P2PSigner will be used for signing off on published content
	// if the node is sequencing and if the p2p stack is enabled
	P2PSigner p2p.SignerSetup

	RPC RPCConfig

	P2P p2p.SetupP2P

	Metrics MetricsConfig

	Pprof oppprof.CLIConfig

	// Used to poll the L1 for new finalized or safe blocks
	L1EpochPollInterval time.Duration

	// Optional
	Tracer    Tracer
	Heartbeat HeartbeatConfig

	Sync sync.Config

	// Path to store safe head database. Disabled when set to empty string
	SafeDBPath string
	DA         da.Config
}

type RPCConfig struct {
	ListenAddr  string
	ListenPort  int
	EnableAdmin bool
}

func (cfg *RPCConfig) HttpEndpoint() string {
	return fmt.Sprintf("http://%s:%d", cfg.ListenAddr, cfg.ListenPort)
}

type MetricsConfig struct {
	Enabled    bool
	ListenAddr string
	ListenPort int
}

func (m MetricsConfig) Check() error {
	if !m.Enabled {
		return nil
	}

	if m.ListenPort < 0 || m.ListenPort > math.MaxUint16 {
		return errors.New("invalid metrics port")
	}

	return nil
}

type HeartbeatConfig struct {
	Enabled bool
	Moniker string
	URL     string
}

// Check verifies that the given configuration makes sense
func (cfg *Config) Check() error {
	if err := cfg.L2.Check(); err != nil {
		return fmt.Errorf("l2 endpoint config error: %w", err)
	}
	if err := cfg.L2Sync.Check(); err != nil {
		return fmt.Errorf("sync config error: %w", err)
	}
	if err := cfg.Rollup.Check(); err != nil {
		return fmt.Errorf("rollup config error: %w", err)
	}
	if err := cfg.Metrics.Check(); err != nil {
		return fmt.Errorf("metrics config error: %w", err)
	}
	if err := cfg.Pprof.Check(); err != nil {
		return fmt.Errorf("pprof config error: %w", err)
	}
	if cfg.P2P != nil {
		if err := cfg.P2P.Check(); err != nil {
			return fmt.Errorf("p2p config error: %w", err)
		}
	}
	if cfg.Beacon != nil {
		if err := cfg.Beacon.Check(); err != nil {
			return fmt.Errorf("beacon config error: %w", err)
		}
	}

	if cfg.Rollup.MantleDaSwitch && !cfg.DA.MantleDAIndexerEnable {
		if cfg.DA.ProxyUrl != "" && cfg.Beacon == nil {
			return errors.New("beacon config must be set when using eigenda")
		}
	}

	return nil
}
