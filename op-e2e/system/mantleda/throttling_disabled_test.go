package mantleda

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
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

const (
	// bigTxSize is the amount of incompressible calldata in each "big" transaction
	bigTxSize = 10000 // 10KB per transaction
	// numBigTxs is the number of big transactions to send to create sufficient backlog
	// 400 * 10KB = 4MB > default threshold 3.2MB
	numBigTxs = 400
)

// TestThrottlingDisabledWithBacklog tests that when OP_BATCHER_THROTTLE_UNSAFE_DA_BYTES_LOWER_THRESHOLD
// is set to 0 (i.e., throttling disabled), even if L2 batches are backlogged in memory,
// op-batcher will not make any throttling-related Engine API calls (miner_setMaxDASize) to op-geth.
//
// Test steps:
// 1. Set LowerThreshold = 0 to disable throttling
// 2. Send many transactions to create DA data backlog (exceeding default threshold 3.2MB)
// 3. Start batcher
// 4. Verify that large transactions can still be included normally without being throttled
// 5. Verify that throttling loop is not started (via logs or behavior verification)
func TestThrottlingDisabledWithBacklog(t *testing.T) {
	op_e2e.InitParallel(t)

	// Setup test environment with LowerThreshold = 0 to disable throttling
	cfg, rollupClient, l2Seq, _, batcher, logHandler := setupTestWithDisabledThrottling(t)

	t.Log("Step 1: Send many transactions to create DA data backlog")
	// Send enough transactions to create backlog exceeding default threshold
	// Default threshold is 3.2MB, we send 4MB of data
	for i := 0; i < numBigTxs; i++ {
		hash := sendTx(t, cfg.Secrets.Alice, uint64(i), bigTxSize, cfg.L2ChainIDBig(), l2Seq)
		// Wait for transaction to be included in block
		if i%50 == 0 { // Check every 50 transactions to avoid excessive waiting
			waitForReceipt(t, hash, l2Seq)
			t.Logf("Sent %d/%d big transactions", i+1, numBigTxs)
		}
	}

	t.Log("Step 2: Wait for all transactions to be included in blocks")
	time.Sleep(2 * time.Second)

	t.Log("Step 3: Start batcher")
	err := batcher.StartBatchSubmitting()
	require.NoError(t, err, "failed to start batcher")

	t.Log("Step 4: Wait sufficient time for batcher to process backlogged data")
	// Wait enough time for batcher to process the 4MB backlog
	// This ensures the batcher is ready to process new transactions
	// If throttling were enabled, miner_setMaxDASize would be called during this period
	time.Sleep(30 * time.Second)

	t.Log("Step 5: Verify that large transactions can still be included normally without throttling")
	// Send a new large transaction and verify it can be included quickly
	// If throttling were enabled, this transaction would be delayed
	startTime := time.Now()
	hash := sendTx(t, cfg.Secrets.Bob, 0, bigTxSize, cfg.L2ChainIDBig(), l2Seq)
	receipt := waitForReceipt(t, hash, l2Seq)
	elapsed := time.Since(startTime)

	require.NotNil(t, receipt, "transaction should be included")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "transaction should succeed")

	// With throttling disabled, transaction should be included relatively quickly
	// We allow up to 30 seconds to account for normal batcher processing time
	require.Less(t, elapsed, 30*time.Second, "transaction should not be delayed by throttling")

	t.Logf("Test passed: transaction included in %v without throttling", elapsed)

	// Wait for transaction to be submitted to L1 and confirmed on verifier
	waitForSafeBlock(t, receipt.BlockNumber, rollupClient)

	t.Log("Step 6: Verify throttling loop was not started via log messages")
	// Verify that the "Throttling loop is DISABLED" warning was logged
	logHandler.RequireMessageContained(t, "Throttling loop is DISABLED due to 0 throttle-threshold")

	// Verify that throttling loop was NOT started
	throttlingStartLog := logHandler.FindLog(testlog.NewMessageContainsFilter("Starting DA throttling loop"))
	require.Nil(t, throttlingStartLog, "Throttling loop should NOT be started when LowerThreshold=0")

	// Verify that miner_setMaxDASize was NOT called
	setMaxDASizeLog := logHandler.FindLog(testlog.NewMessageContainsFilter("Setting max DA size on endpoint"))
	require.Nil(t, setMaxDASizeLog, "miner_setMaxDASize should NOT be called when throttling is disabled")

	// Verify that no throttling warnings were logged
	throttlingWarningLog := logHandler.FindLog(testlog.NewMessageContainsFilter("unsafe bytes above threshold"))
	require.Nil(t, throttlingWarningLog, "No throttling warnings should be logged when throttling is disabled")

	t.Log("All verifications passed: throttling is correctly disabled")
}

// setupTestWithDisabledThrottling sets up a test environment with throttling disabled.
func setupTestWithDisabledThrottling(t *testing.T) (e2esys.SystemConfig, *sources.RollupClient, *ethclient.Client, *ethclient.Client, *batcher.TestBatchSubmitter, *testlog.CapturingHandler) {
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
	// Disable automatic batcher, we will start it manually
	cfg.DisableBatcher = true

	// Set LowerThreshold = 0 to disable throttling
	// Other parameters set to 0 means no limits
	sys, err := cfg.StartMantle(t,
		e2esys.WithBatcherThrottling(500*time.Millisecond, 0, 0, 0))
	require.NoError(t, err, "failed to start system")

	rollupClient := sys.RollupClient("verifier")
	l2Seq := sys.NodeClient("sequencer")
	l2Verif := sys.NodeClient("verifier")

	batcher := sys.BatchSubmitter.ThrottlingTestDriver()

	// Verify throttling configuration
	require.Equal(t, uint64(0), batcher.Config.ThrottleParams.LowerThreshold,
		"LowerThreshold should be 0")
	t.Log("Throttling disabled: LowerThreshold = 0")

	return cfg, rollupClient, l2Seq, l2Verif, batcher, logHandler
}
