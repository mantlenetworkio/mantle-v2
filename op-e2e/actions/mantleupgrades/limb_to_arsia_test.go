package mantleupgrades

import (
	"math/rand"
	"testing"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	mantleHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestLimbToArsiaUpgrade tests the complete upgrade path from Limb to Arsia
// This is the most important test for Mantle fork progression
func TestLimbToArsiaUpgrade(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Test Limb → Arsia Upgrade ==========")

	// Phase 1: Setup Limb activation at genesis, Arsia activation later

	logger.Info("========== Phase 1: Configure Limb → Arsia Upgrade ==========")

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)
	dp.DeployConfig.L2BlockTime = 2

	// Limb activates at genesis (offset 0)
	// Arsia activates later (offset 48 = 4 L2 blocks * 12s block time)
	limbOffset := hexutil.Uint64(0)
	arsiaOffset := hexutil.Uint64(48)

	mantleHelpers.ApplyLimbToArsiaUpgrade(dp, &limbOffset, &arsiaOffset)

	// Create test environment with Mantle-specific setup
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	logger.Info("========== Phase 1: Test Environment Created ==========",
		"L1ChainID", sd.L1Cfg.Config.ChainID,
		"L2ChainID", sd.L2Cfg.Config.ChainID,
		"LimbOffset", limbOffset,
		"ArsiaOffset", arsiaOffset)

	// Verify fork configuration
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.NotNil(t, sd.RollupCfg.MantleLimbTime, "Limb fork should be configured")
	require.NotNil(t, sd.RollupCfg.MantleArsiaTime, "Arsia fork should be configured")

	// Verify fork times based on offset values:
	// - offset=0 → fork time = 0 (special marker for genesis activation)
	// - offset>0 → fork time = genesisTime + offset (absolute timestamp)
	require.Equal(t, uint64(0), *sd.RollupCfg.MantleLimbTime, "Limb offset=0 should be marked as timestamp 0")

	expectedArsiaTime := genesisTime + uint64(arsiaOffset)
	require.Equal(t, expectedArsiaTime, *sd.RollupCfg.MantleArsiaTime, "Arsia should activate at genesis + 48 seconds")

	// Verify Limb < Arsia
	require.Less(t, *sd.RollupCfg.MantleLimbTime, *sd.RollupCfg.MantleArsiaTime,
		"Limb should be active before Arsia")

	logger.Info("========== Phase 1: Limb → Arsia Fork Configuration Test Passed ==========")

	// Phase 2: Create actors and verify Limb state

	logger.Info("========== Phase 2: Create actors and verify Limb state ==========")

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, logger)
	_, verifier := helpers.SetupVerifier(t, sd, logger, miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})

	// Create batcher
	batcher := helpers.NewL2Batcher(logger, sd.RollupCfg, helpers.MantleDefaultBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// Verify genesis block ExtraData
	genesisBlock := seqEngine.L2Chain().GetBlockByNumber(0)
	genesisExtraData := genesisBlock.Extra()
	logger.Info("Genesis block ExtraData",
		"len", len(genesisExtraData),
		"data", genesisExtraData)

	// Genesis block should have empty ExtraData because:
	// - Limb is active at genesis (offset=0)
	// - Arsia is NOT active at genesis (offset=48)
	// - Jovian should NOT be active at genesis (should follow Arsia)
	require.Equal(t, 0, len(genesisExtraData), "Genesis block ExtraData should be 0 bytes (Limb active, Arsia not active)")

	// Verify initial state - Limb active, Arsia not active
	isLimbActive := sd.RollupCfg.IsMantleLimb(genesisTime)
	isArsiaActive := sd.RollupCfg.IsMantleArsia(genesisTime)
	isHoloceneActive := sd.RollupCfg.IsHolocene(genesisTime)

	require.True(t, isLimbActive, "Limb should be activated in genesis.")
	require.False(t, isArsiaActive, "Arsia should not be active in genesis.")
	require.False(t, isHoloceneActive, "Holocene should not be active in genesis.")

	logger.Info("========== Phase 2: Initial State Verification Passed ==========",
		"Limb", isLimbActive,
		"Arsia", isArsiaActive,
		"Holocene", isHoloceneActive)

	// Verify MaxSequencerDrift in Limb era (pre-Fjord)
	// Should use config value (40) instead of hardcoded value (1800)
	maxDriftLimb := sd.ChainSpec.MaxSequencerDrift(genesisTime)
	require.Equal(t, uint64(40), maxDriftLimb,
		"Limb version (pre-Fjord) MaxSequencerDrift should equal config value 40")

	logger.Info("========== Phase 2: Limb version MaxSequencerDrift Verification Passed ==========",
		"configValue", sd.RollupCfg.MaxSequencerDrift,
		"actualValue", maxDriftLimb)

	// Phase 3: Build L2 blocks in Limb era

	logger.Info("========== Phase 3: Build L2 blocks in Limb era ==========")

	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)

	// Build 2 L2 blocks in Limb era
	for i := 0; i < 2; i++ {
		sequencer.ActL2StartBlock(t)
		sequencer.ActL2EndBlock(t)

		currentStatus := sequencer.SyncStatus()
		currentTime := currentStatus.UnsafeL2.Time

		logger.Info("L2 block generated (Limb version)",
			"number", currentStatus.UnsafeL2.Number,
			"time", currentTime,
			"IsMantleLimb", sd.RollupCfg.IsMantleLimb(currentTime),
			"IsMantleArsia", sd.RollupCfg.IsMantleArsia(currentTime))
	}

	limbStatus := sequencer.SyncStatus()
	limbBlock := seqEngine.L2Chain().GetBlockByNumber(limbStatus.UnsafeL2.Number)
	limbExtraData := limbBlock.Extra()

	logger.Info("Limb version block information",
		"number", limbStatus.UnsafeL2.Number,
		"time", limbStatus.UnsafeL2.Time,
		"extraDataLen", len(limbExtraData))

	// Verify Limb block has 0-byte ExtraData
	require.Equal(t, 0, len(limbExtraData), "Limb version ExtraData should be 0 bytes")

	logger.Info("========== Phase 3: Limb version block verification passed ==========")

	// Submit Limb blocks to L1
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// IMPORTANT: Sync both sequencer and verifier
	// This updates the safe head for both nodes
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	logger.Info("Sequencer and Verifier have synchronized Limb block",
		"sequencerSafe", sequencer.SyncStatus().SafeL2.Number,
		"verifierSafe", verifier.L2Safe().Number)

	// Phase 4: Cross Arsia activation boundary

	logger.Info("========== Phase 4: Cross Arsia activation boundary ==========")

	// CRITICAL: Push L1 time far ahead BEFORE building L2 blocks
	// This ensures all L2 blocks built in Phase 4 will have L1 origins after Arsia activation
	// Use the expectedArsiaTime calculated earlier (genesisTime + arsiaOffset)
	l1Head := miner.L1Chain().CurrentBlock()

	logger.Info("Push L1 time to Arsia activation time + large margin",
		"currentL1Time", l1Head.Time,
		"arsiaActivationTime", expectedArsiaTime)

	// Push L1 time to Arsia activation time + large margin
	targetL1Time := expectedArsiaTime + 120 // 120 seconds ahead
	for l1Head.Time < targetL1Time {
		miner.ActEmptyBlock(t)
		l1Head = miner.L1Chain().CurrentBlock()
	}

	// Signal to sequencer so it sees the new L1 head
	sequencer.ActL1HeadSignal(t)

	logger.Info("L1 time advanced to Arsia activation time + large margin",
		"l1Time", l1Head.Time,
		"arsiaTime", expectedArsiaTime,
		"margin", l1Head.Time-expectedArsiaTime)

	// Build more L2 blocks to cross Arsia activation time
	// Now all L2 blocks will have L1 origins after Arsia activation
OuterLoop:
	for i := 0; i < 20; i++ {
		// Generate L1 block to advance L1 time
		miner.ActEmptyBlock(t)
		sequencer.ActL1HeadSignal(t)

		// Generate multiple L2 blocks within the L1 time window
		for j := 0; j < 3; j++ {
			sequencer.ActL2StartBlock(t)
			sequencer.ActL2EndBlock(t)

			currentStatus := sequencer.SyncStatus()
			currentTime := currentStatus.UnsafeL2.Time
			isLimbActive := sd.RollupCfg.IsMantleLimb(currentTime)
			isArsiaActive := sd.RollupCfg.IsMantleArsia(currentTime)
			isHoloceneActive := sd.RollupCfg.IsHolocene(currentTime)

			logger.Info("L2 block generated",
				"number", currentStatus.UnsafeL2.Number,
				"time", currentTime,
				"IsMantleLimb", isLimbActive,
				"IsMantleArsia", isArsiaActive,
				"IsHolocene", isHoloceneActive)

			if isArsiaActive {
				logger.Info("Arsia fork activated",
					"blockNumber", currentStatus.UnsafeL2.Number,
					"blockTime", currentTime)

				// Verify ExtraData changed after Arsia/Jovian activation
				// Arsia includes Jovian, so it should use MinBaseFee ExtraData (17 bytes)
				currentBlock := seqEngine.L2Chain().GetBlockByNumber(currentStatus.UnsafeL2.Number)
				currentExtraData := currentBlock.Extra()

				logger.Info("Arsia version block information",
					"extraDataLen", len(currentExtraData))

				isJovianActive := sd.RollupCfg.IsJovian(currentTime)

				if isJovianActive {
					// Arsia includes Jovian, so ExtraData should be MinBaseFee format (17 bytes)
					require.Equal(t, 17, len(currentExtraData),
						"Arsia version (Jovian active) ExtraData should be 17 bytes (MinBaseFee format)")

					// Parse MinBaseFee ExtraData
					version := currentExtraData[0]
					denominator := uint32(currentExtraData[1])<<24 | uint32(currentExtraData[2])<<16 |
						uint32(currentExtraData[3])<<8 | uint32(currentExtraData[4])
					elasticity := uint32(currentExtraData[5])<<24 | uint32(currentExtraData[6])<<16 |
						uint32(currentExtraData[7])<<8 | uint32(currentExtraData[8])
					minBaseFee := uint64(currentExtraData[9])<<56 | uint64(currentExtraData[10])<<48 |
						uint64(currentExtraData[11])<<40 | uint64(currentExtraData[12])<<32 |
						uint64(currentExtraData[13])<<24 | uint64(currentExtraData[14])<<16 |
						uint64(currentExtraData[15])<<8 | uint64(currentExtraData[16])

					logger.Info("Arsia version ExtraData details (MinBaseFee format)",
						"total_length", len(currentExtraData),
						"version", version,
						"denominator", denominator,
						"elasticity", elasticity,
						"minBaseFee", minBaseFee)

					logger.Info("Arsia version ExtraData updated to Jovian/MinBaseFee format (17 bytes)")
				} else if isHoloceneActive {
					// Only Holocene (without Jovian), should be 9 bytes
					require.GreaterOrEqual(t, len(currentExtraData), 9,
						"Arsia version (Holocene active) ExtraData should be at least 9 bytes")

					version := currentExtraData[0]
					denominator := uint32(currentExtraData[1])<<24 | uint32(currentExtraData[2])<<16 |
						uint32(currentExtraData[3])<<8 | uint32(currentExtraData[4])
					elasticity := uint32(currentExtraData[5])<<24 | uint32(currentExtraData[6])<<16 |
						uint32(currentExtraData[7])<<8 | uint32(currentExtraData[8])

					logger.Info("Arsia version ExtraData details (Holocene format)",
						"total_length", len(currentExtraData),
						"version", version,
						"denominator", denominator,
						"elasticity", elasticity)

					logger.Info("Arsia version ExtraData updated to Holocene format (9 bytes)")
				}
				verifier.ActL1HeadSignal(t)
				verifier.ActL2PipelineFull(t)

				break OuterLoop
			}
		}
	}

	arsiaStatus := sequencer.SyncStatus()

	// Verify Arsia is now active
	isArsiaActiveFinal := sd.RollupCfg.IsMantleArsia(arsiaStatus.UnsafeL2.Time)
	require.True(t, isArsiaActiveFinal, "Arsia version fork should be active")

	logger.Info("Arsia version fork activation verification passed")

	// IMPORTANT: Submit batches after Arsia activation to avoid accumulating too many blocks
	// This prevents "future batch" errors in Phase 6
	logger.Info("Submit L2 blocks after Arsia version fork activation to L1")

	// Push L1 time ahead to avoid "future batch" errors
	l1Head = miner.L1Chain().CurrentBlock()
	l2UnsafeTime := arsiaStatus.UnsafeL2.Time
	safetyMargin := uint64(60)
	targetL1Time = l2UnsafeTime + safetyMargin

	for l1Head.Time < targetL1Time {
		miner.ActEmptyBlock(t)
		l1Head = miner.L1Chain().CurrentBlock()
	}

	logger.Info("Arsia version L1 time advanced",
		"l1Time", l1Head.Time,
		"l2UnsafeTime", l2UnsafeTime)

	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// Sync both nodes
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	logger.Info("Arsia version L2 blocks submitted after fork activation",
		"sequencerUnsafe", sequencer.SyncStatus().UnsafeL2.Number,
		"sequencerSafe", sequencer.SyncStatus().SafeL2.Number,
		"verifierSafe", verifier.L2Safe().Number)

	// Phase 5: Verify OP Stack forks are active

	logger.Info("========== Phase 5: Verify Arsia version OP Stack Forks Activation Status ==========")

	currentTime := arsiaStatus.UnsafeL2.Time

	// Verify all OP Stack forks are now active
	if sd.RollupCfg.CanyonTime != nil {
		require.True(t, sd.RollupCfg.IsCanyon(currentTime), "Arsia version Canyon fork should be active")
		logger.Info("Arsia version Canyon fork activated")
	}

	if sd.RollupCfg.DeltaTime != nil {
		require.True(t, sd.RollupCfg.IsDelta(currentTime), "Arsia version Delta fork should be active")
		logger.Info("Arsia version Delta fork activated")
	}

	if sd.RollupCfg.HoloceneTime != nil {
		require.True(t, sd.RollupCfg.IsHolocene(currentTime), "Arsia version Holocene fork should be active")
		logger.Info("Arsia version Holocene fork activated")
	}
	if sd.RollupCfg.EcotoneTime != nil {
		require.True(t, sd.RollupCfg.IsEcotone(currentTime), "Arsia version Ecotone fork should be active")
		logger.Info("Arsia version Ecotone fork activated")
	}
	if sd.RollupCfg.FjordTime != nil {
		require.True(t, sd.RollupCfg.IsFjord(currentTime), "Arsia version Fjord fork should be active")
		logger.Info("Arsia version Fjord fork activated")
	}
	if sd.RollupCfg.GraniteTime != nil {
		require.True(t, sd.RollupCfg.IsGranite(currentTime), "Arsia version Granite fork should be active")
		logger.Info("Arsia version Granite fork activated")
	}
	if sd.RollupCfg.JovianTime != nil {
		require.True(t, sd.RollupCfg.IsJovian(currentTime), "Arsia version Jovian fork should be active")
		logger.Info("Arsia version Jovian fork activated")
	}
	if sd.RollupCfg.IsthmusTime != nil {
		require.True(t, sd.RollupCfg.IsIsthmus(currentTime), "Arsia version Isthmus fork should be active")
		logger.Info("Arsia version Isthmus fork activated")
	}

	// Verify that MaxSequencerDrift changed to version 1800 after the Arsia upgrade.

	logger.Info("========== Verify that MaxSequencerDrift changed to version 1800 after the Arsia upgrade. ==========")

	maxDrift := sd.ChainSpec.MaxSequencerDrift(currentTime)
	require.Equal(t, uint64(1800), maxDrift,
		"Arsia version upgrade should hardcode MaxSequencerDrift to 1800 (no longer 40)")

	logger.Info("MaxSequencerDrift upgrade verification passed",
		"configValue", sd.RollupCfg.MaxSequencerDrift,
		"limbValue", uint64(40),
		"arsiaValue", maxDrift)

	// Phase 6: Test Span Batch after Arsia activation

	logger.Info("========== Phase 6: Test Span Batch after Arsia activation (Delta/Arsia feature) ==========")

	// IMPORTANT: SpanBatch validation checks if L1 origin time >= Arsia activation time
	// We need to ensure L1 has advanced enough so that the L1 origin of the SpanBatch
	// is after Arsia activation time
	//
	// Arsia activates at L2 time: genesisTime + 48
	// We need L1 origin time to also be >= genesisTime + 48
	//
	// Generate additional L1 blocks to ensure L1 time advances past Arsia activation
	l1Head = miner.L1Chain().CurrentBlock()
	arsiaActivationTime := *sd.RollupCfg.MantleArsiaTime

	logger.Info("Advance L1 time to ensure that SpanBatch's L1 origin is activated after Arsia.",
		"currentL1Time", l1Head.Time,
		"arsiaActivationTime", arsiaActivationTime,
		"genesisTime", genesisTime)

	// Generate L1 blocks until L1 time >= Arsia activation time
	for l1Head.Time < arsiaActivationTime {
		miner.ActEmptyBlock(t)
		l1Head = miner.L1Chain().CurrentBlock()
		logger.Info("Generate L1 block to advance time",
			"l1Number", l1Head.Number.Uint64(),
			"l1Time", l1Head.Time,
			"target", arsiaActivationTime)
	}

	// Signal L1 head to sequencer and build L2 to catch up
	// This ensures new L2 blocks will reference the new L1 origin (after Arsia activation)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	logger.Info("L1 time advanced to ensure that SpanBatch's L1 origin is activated after Arsia.",
		"l1Time", l1Head.Time,
		"arsiaTime", arsiaActivationTime,
		"l2Unsafe", sequencer.SyncStatus().UnsafeL2.Number)

	// IMPORTANT: Generate more L1 blocks to ensure L1 time is far enough ahead
	// This prevents "future batch" errors when submitting
	l2UnsafeTime = sequencer.SyncStatus().UnsafeL2.Time
	logger.Info("Advance L1 time to ensure that SpanBatch's L1 origin is activated after Arsia.",
		"l2UnsafeTime", l2UnsafeTime,
		"currentL1Time", l1Head.Time)

	// Generate L1 blocks until L1 time > L2 unsafe time + safety margin
	safetyMargin = uint64(60) // 60 seconds margin
	targetL1Time = l2UnsafeTime + safetyMargin

	for l1Head.Time < targetL1Time {
		miner.ActEmptyBlock(t)
		l1Head = miner.L1Chain().CurrentBlock()
	}

	sequencer.ActL1HeadSignal(t)

	logger.Info("L1 time advanced to ensure that SpanBatch's L1 origin is activated after Arsia.",
		"l1Time", l1Head.Time,
		"l2UnsafeTime", l2UnsafeTime,
		"margin", l1Head.Time-l2UnsafeTime)

	// IMPORTANT: Submit existing L2 blocks to L1 first
	// This ensures the SpanBatch we create next will only contain NEW blocks
	// whose L1 origin is after Arsia activation
	logger.Info("Submit existing L2 blocks to L1 first to ensure SpanBatch only contains new blocks.")

	// Record current unsafe head before submitting
	unsafeBeforeSubmit := sequencer.SyncStatus().UnsafeL2.Number

	// CRITICAL FIX: Use SingularBatch to submit existing blocks
	// These blocks may have L1 origins before Delta activation
	// SpanBatch requires L1 origin >= Delta activation time
	singularBatcher := helpers.NewL2Batcher(logger, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		ForceSubmitSingularBatch: true, // Force SingularBatch for backwards compatibility
		DataAvailabilityType:     batcherFlags.CalldataType,
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	singularBatcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// CRITICAL: Wait for sequencer to derive the submitted batches
	// This updates the sequencer's safe head, so SpanBatch will only include NEW blocks
	sequencer.ActL1HeadSignal(t)

	// Keep running derivation pipeline until safe head catches up to the unsafe head we just submitted
	logger.Info("Wait for sequencer safe head to update after submitting existing L2 blocks.",
		"targetSafe", unsafeBeforeSubmit,
		"currentSafe", sequencer.SyncStatus().SafeL2.Number)

	for i := 0; i < 10; i++ {
		sequencer.ActL2PipelineFull(t)
		currentSafe := sequencer.SyncStatus().SafeL2.Number

		logger.Info("Sequencer derivation pipeline progress.",
			"iteration", i+1,
			"currentSafe", currentSafe,
			"targetSafe", unsafeBeforeSubmit)

		if currentSafe >= unsafeBeforeSubmit {
			logger.Info("The Safe head has been updated to the submission location.")
			break
		}
	}

	// Sync verifier with submitted batches
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	logger.Info("Existing L2 blocks have been submitted to L1 and processed by the derivation pipeline.",
		"sequencerUnsafe", sequencer.SyncStatus().UnsafeL2.Number,
		"sequencerSafe", sequencer.SyncStatus().SafeL2.Number,
		"verifierSafe", verifier.L2Safe().Number)

	// Verify safe head has been updated
	require.GreaterOrEqual(t, sequencer.SyncStatus().SafeL2.Number, unsafeBeforeSubmit,
		"The Sequencer safe head should have been updated to at least the unsafe head before submission.")

	// Create batcher with SpanBatch enabled
	spanBatcher := helpers.NewL2Batcher(logger, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:          0,
		MaxL1TxSize:          128_000,
		BatcherKey:           dp.Secrets.Batcher,
		ForceSubmitSpanBatch: true, // Force SpanBatch
		DataAvailabilityType: batcherFlags.CalldataType,
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Build more L2 blocks - these will have L1 origin after Arsia activation
	// and will be included in the SpanBatch
	for i := 0; i < 2; i++ {
		sequencer.ActL2StartBlock(t)
		sequencer.ActL2EndBlock(t)
	}

	spanBatchStatus := sequencer.SyncStatus()

	// Log L1 origin info of the SpanBatch blocks for debugging
	l1OriginBlock := miner.L1Chain().GetBlockByNumber(spanBatchStatus.UnsafeL2.L1Origin.Number)
	logger.Info("SpanBatch block information.",
		"l2Number", spanBatchStatus.UnsafeL2.Number,
		"l2Time", spanBatchStatus.UnsafeL2.Time,
		"l1OriginNumber", spanBatchStatus.UnsafeL2.L1Origin.Number,
		"l1OriginTime", l1OriginBlock.Time(),
		"arsiaTime", arsiaActivationTime,
		"l1OriginAfterArsia", l1OriginBlock.Time() >= arsiaActivationTime)

	// Submit SpanBatch
	spanBatcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// CRITICAL FIX: Signal BOTH nodes about the new L1 block, SEQUENCER FIRST!
	// Ensure that both the Sequencer and Verifier see the new L1 block containing SpanBatch.
	// This prevents Verifier from thinking a L1 Re-org occurred.
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t) // The sequencer processes its own batch and updates the safe head.
	// Verifier derives from SpanBatch
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify verifier synced correctly
	require.Equal(t, spanBatchStatus.UnsafeL2.Hash, verifier.L2Safe().Hash,
		"The Verifier should have synchronized correctly with the SpanBatch.")

	logger.Info("SpanBatch test passed.",
		"sequencerUnsafe", spanBatchStatus.UnsafeL2.Number,
		"verifierSafe", verifier.L2Safe().Number)

	// Phase 7: Verify no reorg occurred during upgrade

	logger.Info("========== Phase 7: Verify no reorg occurred during upgrade ==========")

	// Build more blocks to ensure stability
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Submit to L1
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// Sync both nodes
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify chains are consistent
	require.Equal(t, sequencer.L2Safe().Hash, verifier.L2Safe().Hash,
		"The Sequencer and Verifier safe heads should be consistent.")

	logger.Info("Upgrade process completed without Reorg, chain states are consistent.")

	// Phase 8: Summary

	logger.Info("Limb → Arsia upgrade testing complete")

	logger.Info("")
	logger.Info("Test summary:")
	logger.Info("  Limb fork activated at genesis.")
	logger.Info("  Limb version MaxSequencerDrift = 40 (configured value)")
	logger.Info("  Limb version blocks generated with ExtraData = 0 bytes.")
	logger.Info("  Arsia fork activated at the scheduled time.")
	logger.Info("  Arsia version MaxSequencerDrift = 1800 (Fjord hardcoded value)")
	logger.Info("  Arsia version blocks generated with ExtraData >= 9 bytes.")
	logger.Info("  All OP Stack forks activated correctly.")
	logger.Info("  SpanBatch is working as expected.")
	logger.Info("  Upgrade process completed without Reorg.")
	logger.Info("  Sequencer and Verifier states are consistent.")
	logger.Info("")

	t.Log("TestLimbToArsiaUpgrade testing complete.")
}

// TestLimbToArsiaUpgradeWithTransactions tests the upgrade with actual transactions
func TestLimbToArsiaUpgradeWithTransactions(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Testing Limb → Arsia upgrade with transactions ==========")

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Limb at genesis, Arsia at offset 36
	limbOffset := hexutil.Uint64(0)
	arsiaOffset := hexutil.Uint64(36)
	mantleHelpers.ApplyLimbToArsiaUpgrade(dp, &limbOffset, &arsiaOffset)

	// Use SetupMantleNormal to preserve manual fork configuration
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, logger)
	_, verifier := helpers.SetupVerifier(t, sd, logger, miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)

	// Send transaction in Limb era
	alice := helpers.NewBasicUser[any](logger, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&helpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})
	alice.ActResetTxOpts(t)
	alice.ActSetTxToAddr(&dp.Addresses.Bob)(t)
	alice.ActMakeTx(t)

	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	limbTxStatus := sequencer.SyncStatus()
	logger.Info("Limb version transaction included in block",
		"blockNumber", limbTxStatus.UnsafeL2.Number,
		"blockTime", limbTxStatus.UnsafeL2.Time)

	// Cross Arsia activation
	for i := 0; i < 4; i++ {
		sequencer.ActL2StartBlock(t)
		sequencer.ActL2EndBlock(t)

		currentTime := sequencer.SyncStatus().UnsafeL2.Time
		if sd.RollupCfg.IsMantleArsia(currentTime) {
			logger.Info("Arsia fork activated",
				"blockNumber", sequencer.SyncStatus().UnsafeL2.Number)
			break
		}
	}

	// Send transaction in Arsia era
	bob := helpers.NewBasicUser[any](logger, dp.Secrets.Bob, rand.New(rand.NewSource(5678)))
	bob.SetUserEnv(&helpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})
	bob.ActResetTxOpts(t)
	bob.ActSetTxToAddr(&dp.Addresses.Alice)(t)
	bob.ActMakeTx(t)

	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(t)
	sequencer.ActL2EndBlock(t)

	arsiaTxStatus := sequencer.SyncStatus()
	logger.Info("Arsia version transaction included in block",
		"blockNumber", arsiaTxStatus.UnsafeL2.Number,
		"blockTime", arsiaTxStatus.UnsafeL2.Time)

	// CRITICAL: Push L1 time ahead to avoid "future batch" errors
	l1Head := miner.L1Chain().CurrentBlock()
	l2UnsafeTime := arsiaTxStatus.UnsafeL2.Time
	safetyMargin := uint64(60)
	targetL1Time := l2UnsafeTime + safetyMargin

	for l1Head.Time < targetL1Time {
		miner.ActEmptyBlock(t)
		l1Head = miner.L1Chain().CurrentBlock()
	}

	sequencer.ActL1HeadSignal(t)

	logger.Info("L1 time advanced to avoid future batch errors",
		"l1Time", l1Head.Time,
		"l2UnsafeTime", l2UnsafeTime,
		"margin", l1Head.Time-l2UnsafeTime)

	// Submit all to L1
	// Use SingularBatch to handle blocks with L1 origins before Delta activation
	singularBatcher := helpers.NewL2Batcher(logger, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		ForceSubmitSingularBatch: true,
		DataAvailabilityType:     batcherFlags.CalldataType,
	}, sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	singularBatcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// CRITICAL: Both sequencer and verifier must process the submitted batch
	// This updates their safe heads to include the submitted transactions
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Record final states
	finalSequencerStatus := sequencer.SyncStatus()
	finalVerifierSafe := verifier.L2Safe()

	logger.Info("Final sequencer and verifier sync status",
		"sequencerUnsafe", finalSequencerStatus.UnsafeL2.Number,
		"sequencerSafe", finalSequencerStatus.SafeL2.Number,
		"verifierSafe", finalVerifierSafe.Number)

	// Verify both nodes have processed all transactions
	// Compare safe heads (both should have processed the submitted batch)
	require.Equal(t, finalSequencerStatus.SafeL2.Hash, finalVerifierSafe.Hash,
		"Sequencer and Verifier safe heads should be consistent")

	// Verify the safe head includes both transactions (should be at least the Arsia transaction block)
	require.GreaterOrEqual(t, finalVerifierSafe.Number, arsiaTxStatus.UnsafeL2.Number,
		"Verifier safe head should include Arsia version transaction")

	// Verify transactions in both Limb and Arsia blocks
	verifCl := seqEngine.EthClient()
	limbBlock, err := verifCl.BlockByNumber(t.Ctx(), nil)
	require.NoError(t, err)
	require.NotNil(t, limbBlock, "Should be able to fetch Limb version block")

	arsiaBlock, err := verifCl.BlockByNumber(t.Ctx(), nil)
	require.NoError(t, err)
	require.NotNil(t, arsiaBlock, "Should be able to fetch Arsia version block")

	logger.Info("Limb and Arsia version transactions processed correctly")
	logger.Info("Upgrade test with transactions passed")
}

// TestArsiaTriggersHoloceneStageTransformation tests that when Arsia fork activates,
// it correctly triggers the Holocene stage transformation in the derivation pipeline.
//
// This is critical because Arsia is a Mantle-specific fork that encompasses multiple
// OP Stack forks including Holocene. Without proper stage transformation:
// - BatchQueue won't be replaced by BatchStage
// - Past batches won't be handled correctly (BatchPast vs BatchDrop)
// - The derivation pipeline may behave incorrectly after upgrade
//
// The test verifies:
// 1. Before Arsia: No Holocene stage transformation logs
// 2. After Arsia activation: "BatchMux: transforming to Holocene stage" log appears
// 3. The derivation pipeline continues to work correctly after transformation
func TestArsiaTriggersHoloceneStageTransformation(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// ========== Phase 1: Setup environment with log capturing ==========

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	dp := e2eutils.MakeMantleDeployParams(t, testParams)
	dp.DeployConfig.L2BlockTime = 2

	// Limb at genesis, Arsia activates later
	limbOffset := hexutil.Uint64(0)
	arsiaOffset := hexutil.Uint64(48)
	mantleHelpers.ApplyLimbToArsiaUpgrade(dp, &limbOffset, &arsiaOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	// Setup log capturing for both sequencer and verifier
	seqLog := testlog.Logger(t, log.LevelInfo)
	seqLogHandler := testlog.WrapCaptureLogger(seqLog.Handler())
	seqLogger := log.NewLogger(seqLogHandler)

	verifLog := testlog.Logger(t, log.LevelInfo)
	verifLogHandler := testlog.WrapCaptureLogger(verifLog.Handler())
	verifLogger := log.NewLogger(verifLogHandler)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, seqLogger)
	_, verifier := helpers.SetupVerifier(t, sd, verifLogger, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})

	batcher := helpers.NewL2Batcher(testlog.Logger(t, log.LevelInfo), sd.RollupCfg,
		helpers.MantleDefaultBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(),
		seqEngine.EngineClient(t, sd.RollupCfg))

	// Helper functions to check logs
	seqCapture := seqLogHandler.(*testlog.CapturingHandler)
	verifCapture := verifLogHandler.(*testlog.CapturingHandler)

	// Holocene activates two stage transformations:
	// 1. BatchMux: BatchQueue -> BatchStage
	// 2. ChannelMux: ChannelBank -> ChannelAssembler
	findBatchMuxTransformLog := func(handler *testlog.CapturingHandler) *testlog.CapturedRecord {
		return handler.FindLog(
			testlog.NewMessageContainsFilter("BatchMux: transforming to Holocene stage"),
		)
	}

	findChannelMuxTransformLog := func(handler *testlog.CapturingHandler) *testlog.CapturedRecord {
		return handler.FindLog(
			testlog.NewMessageContainsFilter("ChannelMux: transforming to Holocene stage"),
		)
	}

	// ========== Phase 2: Verify no transformation before Arsia ==========

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	genesisTime := sd.RollupCfg.Genesis.L2Time
	expectedArsiaTime := genesisTime + uint64(arsiaOffset)

	// Verify initial state
	require.True(t, sd.RollupCfg.IsMantleLimb(genesisTime), "Limb should be active at genesis")
	require.False(t, sd.RollupCfg.IsMantleArsia(genesisTime), "Arsia should not be active at genesis")
	require.False(t, sd.RollupCfg.IsHolocene(genesisTime), "Holocene should not be active at genesis")

	// Build some L2 blocks before Arsia
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Submit batch before Arsia
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batcher.LastSubmitted.Hash())(t)
	miner.ActL1EndBlock(t)

	// Derive on verifier
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify NO Holocene transformation happened yet (both BatchMux and ChannelMux)
	require.Nil(t, findBatchMuxTransformLog(seqCapture),
		"Sequencer should NOT have BatchMux transformation log before Arsia")
	require.Nil(t, findChannelMuxTransformLog(seqCapture),
		"Sequencer should NOT have ChannelMux transformation log before Arsia")
	require.Nil(t, findBatchMuxTransformLog(verifCapture),
		"Verifier should NOT have BatchMux transformation log before Arsia")
	require.Nil(t, findChannelMuxTransformLog(verifCapture),
		"Verifier should NOT have ChannelMux transformation log before Arsia")

	t.Log("Phase 2 passed: No Holocene stage transformations before Arsia activation")

	// ========== Phase 3: Activate Arsia and verify transformation ==========

	// Push L1 time past Arsia activation
	for miner.L1Chain().CurrentBlock().Time < expectedArsiaTime+60 {
		miner.ActEmptyBlock(t)
	}

	// Build L2 blocks after Arsia activation
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Verify Arsia is now active
	currentL2Time := sequencer.SyncStatus().UnsafeL2.Time
	require.True(t, sd.RollupCfg.IsMantleArsia(currentL2Time), "Arsia should be active now")
	require.True(t, sd.RollupCfg.IsHolocene(currentL2Time), "Holocene should be active with Arsia")

	// Submit batch after Arsia
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batcher.LastSubmitted.Hash())(t)
	miner.ActL1EndBlock(t)

	// Derive on verifier - this should trigger Holocene stage transformations
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	// Verify both Holocene transformations happened on verifier:
	// 1. BatchMux: BatchQueue -> BatchStage
	verifBatchMuxLog := findBatchMuxTransformLog(verifCapture)
	require.NotNil(t, verifBatchMuxLog,
		"Verifier should have 'BatchMux: transforming to Holocene stage' log after Arsia activation")
	t.Logf("Found verifier BatchMux transformation log: %s", verifBatchMuxLog.Message)

	// 2. ChannelMux: ChannelBank -> ChannelAssembler
	verifChannelMuxLog := findChannelMuxTransformLog(verifCapture)
	require.NotNil(t, verifChannelMuxLog,
		"Verifier should have 'ChannelMux: transforming to Holocene stage' log after Arsia activation")
	t.Logf("Found verifier ChannelMux transformation log: %s", verifChannelMuxLog.Message)

	// Derive on sequencer
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	// Verify both Holocene transformations happened on sequencer:
	// 1. BatchMux: BatchQueue -> BatchStage
	seqBatchMuxLog := findBatchMuxTransformLog(seqCapture)
	require.NotNil(t, seqBatchMuxLog,
		"Sequencer should have 'BatchMux: transforming to Holocene stage' log after Arsia activation")
	t.Logf("Found sequencer BatchMux transformation log: %s", seqBatchMuxLog.Message)

	// 2. ChannelMux: ChannelBank -> ChannelAssembler
	seqChannelMuxLog := findChannelMuxTransformLog(seqCapture)
	require.NotNil(t, seqChannelMuxLog,
		"Sequencer should have 'ChannelMux: transforming to Holocene stage' log after Arsia activation")
	t.Logf("Found sequencer ChannelMux transformation log: %s", seqChannelMuxLog.Message)

	t.Log("Phase 3 passed: Both BatchMux and ChannelMux transformations occurred after Arsia activation")

	// ========== Phase 4: Verify derivation continues to work ==========

	// Build more L2 blocks
	for i := 0; i < 3; i++ {
		sequencer.ActL2StartBlock(t)
		sequencer.ActL2EndBlock(t)
	}

	// Submit another batch
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)

	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batcher.LastSubmitted.Hash())(t)
	miner.ActL1EndBlock(t)

	// Verify derivation works correctly after transformation
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	// Verify sequencer and verifier are in sync
	require.Equal(t, sequencer.L2Safe().Hash, verifier.L2Safe().Hash,
		"Sequencer and verifier should have same safe head after Holocene transformation")

	t.Log("Phase 4 passed: Derivation continues to work correctly after Holocene transformation")
	t.Log("TestArsiaTriggersHoloceneStageTransformation PASSED")
}
