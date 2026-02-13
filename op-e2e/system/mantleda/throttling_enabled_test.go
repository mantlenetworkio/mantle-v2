package mantleda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"

	"github.com/ethereum-optimism/optimism/op-batcher/batcher"
	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestThrottlingEnabledFailure verifies that when throttling is enabled (LowerThreshold > 0),
// the batcher fails to start because op-geth does not support the miner_setMaxDASize RPC method.
//
// Background:
// - The miner_setMaxDASize RPC method has been removed from op-geth
// - When throttling is enabled, batcher attempts to call this method
// - The RPC call fails with "method does not exist" error
// - Batcher automatically shuts down when it detects the missing method
//
// Test steps:
// 1. Setup with LowerThreshold = 1 to enable throttling
// 2. Send a transaction to create DA data
// 3. Start batcher (should start successfully initially)
// 4. Wait for batcher to detect missing RPC method and shut down
// 5. Verify error logs contain expected messages
func TestThrottlingEnabledFailure(t *testing.T) {
	op_e2e.InitParallel(t)

	// Setup test environment with LowerThreshold = 1 to enable throttling
	cfg, _, l2Seq, _, batcher, logHandler := setupTestWithEnabledThrottling(t)

	t.Log("Step 1: Send a transaction to create some DA data")
	// Send a transaction before starting batcher
	receipt := sendAndWaitForReceipt(t, cfg.Secrets.Alice, 0, bigTxSize, cfg.L2ChainIDBig(), l2Seq)
	require.NotNil(t, receipt, "transaction should be included before batcher starts")

	t.Log("Step 2: Start batcher with throttling enabled (LowerThreshold = 1)")
	err := batcher.StartBatchSubmitting()
	require.NoError(t, err, "batcher should start initially")

	t.Log("Step 3: Wait for batcher to detect RPC method unavailable and shut down")
	// The batcher will:
	// 1. Start the throttling loop
	// 2. Attempt to call miner_setMaxDASize
	// 3. Receive "method does not exist" error
	// 4. Shut down automatically

	// Wait for the error to occur and batcher to shut down
	time.Sleep(15 * time.Second)

	t.Log("Step 4: Verify error logs")
	// Verify that throttling loop was started
	logHandler.RequireMessageContained(t, "Starting DA throttling loop")

	// Verify that the SetMaxDASize RPC was attempted
	logHandler.RequireMessageContained(t, "Setting max DA size on endpoint")

	// Verify that the RPC method unavailable error was logged
	// The error message is in the "err" field of the log
	logHandler.RequireMessageContained(t, "SetMaxDASize RPC method unavailable on endpoint")

	// Verify that throttling was NOT disabled (it was enabled but failed)
	disabledLog := logHandler.FindLog(testlog.NewMessageContainsFilter("Throttling loop is DISABLED"))
	require.Nil(t, disabledLog, "Throttling loop should NOT show as disabled - it was enabled but failed")

	t.Log("Test passed: Batcher correctly detected missing RPC method and shut down")
}

// setupTestWithEnabledThrottling sets up a test environment with throttling enabled.
func setupTestWithEnabledThrottling(t *testing.T) (e2esys.SystemConfig, *sources.RollupClient, *ethclient.Client, *ethclient.Client, *batcher.TestBatchSubmitter, *testlog.CapturingHandler) {
	// Create a capturing logger for the batcher to verify log messages
	// Use LevelDebug to capture all throttling-related logs including "Setting max DA size on endpoint"
	batcherLog, logHandler := testlog.CaptureLogger(t, log.LevelDebug)

	// Use Mantle Arsia configuration
	arsiaTimeOffset := hexutil.Uint64(0)
	cfg := e2esys.MantleArsiaSystemConfigP2PGossip(t, &arsiaTimeOffset)

	// Replace the default batcher logger with our capturing logger
	cfg.Loggers["batcher"] = batcherLog.New("role", "batcher")

	cfg.GethOptions["sequencer"] = append(cfg.GethOptions["sequencer"], []geth.GethOption{
		func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
			ethCfg.Miner.GasCeil = 30_000_000
			return nil
		},
	}...)
	// Disable automatic batcher
	cfg.DisableBatcher = true

	// Set LowerThreshold = 1 to enable throttling
	// maxTxSize = 100 means only very small transactions are allowed
	sys, err := cfg.StartMantle(t,
		e2esys.WithBatcherThrottling(500*time.Millisecond, 1, 100, 0))
	require.NoError(t, err, "failed to start system")

	rollupClient := sys.RollupClient("verifier")
	l2Seq := sys.NodeClient("sequencer")
	l2Verif := sys.NodeClient("verifier")

	batcher := sys.BatchSubmitter.ThrottlingTestDriver()

	// Verify throttling configuration
	require.Equal(t, uint64(1), batcher.Config.ThrottleParams.LowerThreshold,
		"LowerThreshold should be 1")
	require.Equal(t, uint64(100), batcher.Config.ThrottleParams.TxSizeLowerLimit,
		"TxSizeLowerLimit should be 100")
	t.Log("Throttling enabled: LowerThreshold = 1, TxSizeLowerLimit = 100")

	return cfg, rollupClient, l2Seq, l2Verif, batcher, logHandler
}
