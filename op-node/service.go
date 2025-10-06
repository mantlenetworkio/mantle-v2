package opnode

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	oppprof "github.com/ethereum-optimism/optimism/op-service/pprof"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/flags"
	"github.com/ethereum-optimism/optimism/op-node/node"
	p2pcli "github.com/ethereum-optimism/optimism/op-node/p2p/cli"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/da"
	"github.com/ethereum-optimism/optimism/op-node/rollup/driver"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
)

// NewConfig creates a Config from the provided flags or environment variables.
func NewConfig(ctx *cli.Context, log log.Logger) (*node.Config, error) {
	if err := flags.CheckRequired(ctx); err != nil {
		return nil, err
	}

	rollupConfig, err := NewRollupConfig(ctx)
	if err != nil {
		return nil, err
	}

	driverConfig := NewDriverConfig(ctx)

	p2pSignerSetup, err := p2pcli.LoadSignerSetup(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load p2p signer: %w", err)
	}

	p2pConfig, err := p2pcli.NewConfig(ctx, rollupConfig.BlockTime)
	if err != nil {
		return nil, fmt.Errorf("failed to load p2p config: %w", err)
	}

	l1Endpoint := NewL1EndpointConfig(ctx)

	l2Endpoint, err := NewL2EndpointConfig(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load l2 endpoints info: %w", err)
	}

	l2SyncEndpoint := NewL2SyncEndpointConfig(ctx)

	syncConfig := NewSyncConfig(ctx)

	daCfg, err := NewEigenDAConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load da config: %w", err)
	}

	cfg := &node.Config{
		L1:     l1Endpoint,
		L2:     l2Endpoint,
		L2Sync: l2SyncEndpoint,
		Rollup: *rollupConfig,
		Driver: *driverConfig,
		RPC: node.RPCConfig{
			ListenAddr:  ctx.String(flags.RPCListenAddr.Name),
			ListenPort:  ctx.Int(flags.RPCListenPort.Name),
			EnableAdmin: ctx.Bool(flags.RPCEnableAdmin.Name),
		},
		Metrics: node.MetricsConfig{
			Enabled:    ctx.Bool(flags.MetricsEnabledFlag.Name),
			ListenAddr: ctx.String(flags.MetricsAddrFlag.Name),
			ListenPort: ctx.Int(flags.MetricsPortFlag.Name),
		},
		Pprof: oppprof.CLIConfig{
			Enabled:    ctx.Bool(flags.PprofEnabledFlag.Name),
			ListenAddr: ctx.String(flags.PprofAddrFlag.Name),
			ListenPort: ctx.Int(flags.PprofPortFlag.Name),
		},
		P2P:                 p2pConfig,
		P2PSigner:           p2pSignerSetup,
		L1EpochPollInterval: ctx.Duration(flags.L1EpochPollIntervalFlag.Name),
		Heartbeat: node.HeartbeatConfig{
			Enabled: ctx.Bool(flags.HeartbeatEnabledFlag.Name),
			Moniker: ctx.String(flags.HeartbeatMonikerFlag.Name),
			URL:     ctx.String(flags.HeartbeatURLFlag.Name),
		},
		SafeDBPath: ctx.String(flags.SafeDBPath.Name),
		Sync:       *syncConfig,
		DA:         daCfg,
	}
	beacon := NewBeaconEndpointConfig(ctx)
	if beacon.Check() == nil {
		cfg.Beacon = beacon
	}

	if err := cfg.Check(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewL1EndpointConfig(ctx *cli.Context) *node.L1EndpointConfig {
	return &node.L1EndpointConfig{
		L1NodeAddr:       ctx.String(flags.L1NodeAddr.Name),
		L1TrustRPC:       ctx.Bool(flags.L1TrustRPC.Name),
		L1RPCKind:        sources.RPCProviderKind(strings.ToLower(ctx.String(flags.L1RPCProviderKind.Name))),
		RateLimit:        ctx.Float64(flags.L1RPCRateLimit.Name),
		BatchSize:        ctx.Int(flags.L1RPCMaxBatchSize.Name),
		HttpPollInterval: ctx.Duration(flags.L1HTTPPollInterval.Name),
	}
}

func NewBeaconEndpointConfig(ctx *cli.Context) node.L1BeaconEndpointSetup {
	return &node.L1BeaconEndpointConfig{
		BeaconAddr:             ctx.String(flags.BeaconAddr.Name),
		BeaconHeader:           ctx.String(flags.BeaconHeader.Name),
		BeaconArchiverAddr:     ctx.String(flags.BeaconArchiverAddr.Name),
		BeaconCheckIgnore:      ctx.Bool(flags.BeaconCheckIgnore.Name),
		BeaconFetchAllSidecars: ctx.Bool(flags.BeaconFetchAllSidecars.Name),
	}
}

func NewL2EndpointConfig(ctx *cli.Context, log log.Logger) (*node.L2EndpointConfig, error) {
	l2Addr := ctx.String(flags.L2EngineAddr.Name)
	fileName := ctx.String(flags.L2EngineJWTSecret.Name)
	var secret [32]byte
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, fmt.Errorf("file-name of jwt secret is empty")
	}
	if data, err := os.ReadFile(fileName); err == nil {
		jwtSecret := common.FromHex(strings.TrimSpace(string(data)))
		if len(jwtSecret) != 32 {
			return nil, fmt.Errorf("invalid jwt secret in path %s, not 32 hex-formatted bytes", fileName)
		}
		copy(secret[:], jwtSecret)
	} else {
		log.Warn("Failed to read JWT secret from file, generating a new one now. Configure L2 geth with --authrpc.jwt-secret=" + fmt.Sprintf("%q", fileName))
		if _, err := io.ReadFull(rand.Reader, secret[:]); err != nil {
			return nil, fmt.Errorf("failed to generate jwt secret: %w", err)
		}
		if err := os.WriteFile(fileName, []byte(hexutil.Encode(secret[:])), 0600); err != nil {
			return nil, err
		}
	}

	return &node.L2EndpointConfig{
		L2EngineAddr:      l2Addr,
		L2EngineJWTSecret: secret,
	}, nil
}

// NewL2SyncEndpointConfig returns a pointer to a L2SyncEndpointConfig if the
// flag is set, otherwise nil.
func NewL2SyncEndpointConfig(ctx *cli.Context) *node.L2SyncEndpointConfig {
	return &node.L2SyncEndpointConfig{
		L2NodeAddr: ctx.String(flags.BackupL2UnsafeSyncRPC.Name),
		TrustRPC:   ctx.Bool(flags.BackupL2UnsafeSyncRPCTrustRPC.Name),
	}
}

func NewDriverConfig(ctx *cli.Context) *driver.Config {
	return &driver.Config{
		VerifierConfDepth:   ctx.Uint64(flags.VerifierL1Confs.Name),
		SequencerConfDepth:  ctx.Uint64(flags.SequencerL1Confs.Name),
		SequencerEnabled:    ctx.Bool(flags.SequencerEnabledFlag.Name),
		SequencerStopped:    ctx.Bool(flags.SequencerStoppedFlag.Name),
		SequencerMaxSafeLag: ctx.Uint64(flags.SequencerMaxSafeLagFlag.Name),
	}
}

func NewRollupConfig(ctx *cli.Context) (*rollup.Config, error) {
	network := ctx.String(flags.Network.Name)
	if network != "" {
		config, err := chaincfg.GetRollupConfig(network)
		if err != nil {
			return nil, err
		}

		return &config, nil
	}

	rollupConfigPath := ctx.String(flags.RollupConfig.Name)
	file, err := os.Open(rollupConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rollup config: %w", err)
	}
	defer file.Close()

	var rollupConfig rollup.Config
	if err := json.NewDecoder(file).Decode(&rollupConfig); err != nil {
		return nil, fmt.Errorf("failed to decode rollup config: %w", err)
	}

	initMantleUpgradeConfig(&rollupConfig)

	return &rollupConfig, nil
}

func NewSnapshotLogger(ctx *cli.Context) (log.Logger, error) {
	snapshotFile := ctx.String(flags.SnapshotLog.Name)
	if snapshotFile == "" {
		return log.NewLogger(log.DiscardHandler()), nil
	}

	sf, err := os.OpenFile(snapshotFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	handler := log.JSONHandler(sf)
	return log.NewLogger(handler), nil
}

func NewSyncConfig(ctx *cli.Context) *sync.Config {
	return &sync.Config{
		EngineSync:         ctx.Bool(flags.L2EngineSyncEnabled.Name),
		SkipSyncStartCheck: ctx.Bool(flags.SkipSyncStartCheck.Name),
	}
}

func NewEigenDAConfig(ctx *cli.Context) (da.Config, error) {
	return da.Config{
		Config: eigenda.Config{
			ProxyUrl:            ctx.String(eigenda.EigenDAProxyUrlFlagName),
			DisperserUrl:        ctx.String(eigenda.EigenDADisperserUrlFlagName),
			DisperseBlobTimeout: ctx.Duration(eigenda.DisperseBlobTimeoutFlagName),
			RetrieveBlobTimeout: ctx.Duration(eigenda.RetrieveBlobTimeoutFlagName),
		},
		MantleDaIndexerSocket: ctx.String(flags.MantleDaIndexerSocketFlag.Name),
		MantleDAIndexerEnable: ctx.Bool(flags.MantleDAIndexerEnableFlag.Name),
	}, nil
}
