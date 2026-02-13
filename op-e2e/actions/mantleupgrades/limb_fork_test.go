package mantleupgrades

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	mantleHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestLimbActivationAtGenesis tests that Limb fork can be activated at genesis
func TestLimbActivationAtGenesis(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Testing Limb activation in Genesis ==========")

	// Create base deploy params
	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Apply Limb activation at genesis
	offset := hexutil.Uint64(0)
	mantleHelpers.ApplyLimbTimeOffset(dp, &offset)

	// Create test environment with Mantle-specific setup
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	logger.Info("Test environment created",
		"L1ChainID", sd.L1Cfg.Config.ChainID,
		"L2ChainID", sd.L2Cfg.Config.ChainID)

	// Verify Limb fork configuration
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.NotNil(t, sd.RollupCfg.MantleLimbTime, "Limb fork should already be configured.")

	// When offset=0, fork time is set to 0 (special marker for "genesis activation")
	// Any actual genesis timestamp (2025) is >= 0, so the fork will be active
	require.Equal(t, uint64(0), *sd.RollupCfg.MantleLimbTime, "Limb fork offset=0 should be marked as timestamp 0.")

	// Verify Limb is active at genesis
	isLimbActive := sd.RollupCfg.IsMantleLimb(genesisTime)
	require.True(t, isLimbActive, "Limb fork should be active at genesis.")
	logger.Info("Limb fork activated at genesis.")

	// Verify Arsia is NOT active
	if sd.RollupCfg.MantleArsiaTime != nil {
		isArsiaActive := sd.RollupCfg.IsMantleArsia(genesisTime)
		require.False(t, isArsiaActive, "Arsia fork should not be active at genesis.")
		logger.Info("Arsia fork not activated at genesis.")
	}

	// Verify EIP-1559 parameters for Limb version
	require.NotNil(t, sd.RollupCfg.ChainOpConfig, "ChainOpConfig should not be nil.")
	elasticity := sd.RollupCfg.ChainOpConfig.EIP1559Elasticity
	denominator := sd.RollupCfg.ChainOpConfig.EIP1559Denominator

	logger.Info("EIP-1559 parameters for Limb version",
		"Elasticity", elasticity,
		"Denominator", denominator)

	require.Equal(t, uint64(4), elasticity, "Limb version Elasticity should be 4.")
	require.Equal(t, uint64(50), denominator, "Limb version Denominator should be 50.")

	logger.Info("EIP-1559 parameters for Limb version are correct.")

	// Create actors and verify chain works
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, logger)

	sequencer.ActL2PipelineFull(t)
	require.Equal(t, uint64(0), sequencer.SyncStatus().UnsafeL2.Number, "Initial L2 block number should be 0.")

	// Build L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2StartBlock(t)
	sequencer.ActL2EndBlock(t)

	l2Status := sequencer.SyncStatus()
	require.Greater(t, l2Status.UnsafeL2.Number, uint64(0), "L2 block number should be greater than 0.")

	// Verify block ExtraData format (Limb version should have 0 bytes)
	l2Block := seqEngine.L2Chain().GetBlockByNumber(l2Status.UnsafeL2.Number)
	require.NotNil(t, l2Block, "L2 block should not be nil.")

	extraData := l2Block.Extra()
	require.Equal(t, 0, len(extraData), "ExtraData length for Limb version should be 0 bytes (non-Holocene).")

	logger.Info("ExtraData length for Limb version is correct (0 bytes).")
	logger.Info("Limb fork genesis activation test passed.")
}

// TestLimbLateActivation tests that Limb fork can be activated after genesis
func TestLimbLateActivation(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Testing Limb Delayed Activation ==========")

	// Configure to activate Limb at offset 24 (after first L1 block)
	limbOffset := hexutil.Uint64(24)

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Apply Limb activation at offset 24
	mantleHelpers.ApplyLimbTimeOffset(dp, &limbOffset)

	// Create test environment with Mantle-specific setup
	// Use SetupMantleNormal to preserve manual fork configuration
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	logger.Info("Test environment created",
		"L1ChainID", sd.L1Cfg.Config.ChainID,
		"L2ChainID", sd.L2Cfg.Config.ChainID,
		"LimbOffset", limbOffset)

	// Verify Limb fork configuration
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.NotNil(t, sd.RollupCfg.MantleLimbTime, "Limb fork should already be configured.")

	// When offset=24, fork time is genesisTime + 24 (absolute timestamp)
	expectedLimbTime := genesisTime + uint64(limbOffset)
	require.Equal(t, expectedLimbTime, *sd.RollupCfg.MantleLimbTime, "The Limb fork should activate after 24 seconds of genesis.")

	// Verify Limb is NOT active at genesis
	isLimbActiveAtGenesis := sd.RollupCfg.IsMantleLimb(genesisTime)
	require.False(t, isLimbActiveAtGenesis, "Limb fork should not be active at genesis.")
	logger.Info("Limb fork not activated at genesis.")

	// Create actors
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, logger)

	sequencer.ActL2PipelineFull(t)

	// Build L1 block to activate Limb
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)

	// Build L2 blocks until Limb activation
	// With offset=24 and L2 block time=1s, need at least 24 blocks to reach activation
	for i := 0; i < 30; i++ {
		sequencer.ActL2StartBlock(t)
		sequencer.ActL2EndBlock(t)

		currentTime := sequencer.SyncStatus().UnsafeL2.Time
		isLimbActive := sd.RollupCfg.IsMantleLimb(currentTime)

		logger.Info("L2 block generated",
			"number", sequencer.SyncStatus().UnsafeL2.Number,
			"time", currentTime,
			"isLimbActive", isLimbActive)

		if isLimbActive {
			logger.Info("Limb fork activated",
				"blockNumber", sequencer.SyncStatus().UnsafeL2.Number,
				"blockTime", currentTime)
			break
		}
	}

	// Verify Limb is now active
	finalTime := sequencer.SyncStatus().UnsafeL2.Time
	isLimbActiveFinal := sd.RollupCfg.IsMantleLimb(finalTime)
	require.True(t, isLimbActiveFinal, "Limb fork should be active after 24 seconds of genesis.")

	// Build more blocks after activation
	sequencer.ActL2StartBlock(t)
	sequencer.ActL2EndBlock(t)

	l2Status := sequencer.SyncStatus()
	l2Block := seqEngine.L2Chain().GetBlockByNumber(l2Status.UnsafeL2.Number)
	require.NotNil(t, l2Block, "L2 block should not be nil.")

	// Verify ExtraData is still 0 bytes (Limb doesn't include Holocene)
	extraData := l2Block.Extra()
	require.Equal(t, 0, len(extraData), "ExtraData length for Limb version should be 0 bytes (non-Holocene).")

	logger.Info("Limb fork late activation test passed.")
}

// TestLimbForkOrdering tests that Mantle forks activate in the correct order
func TestLimbForkOrdering(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Test Mantle Fork order ==========")

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Activate Limb at genesis
	offset := hexutil.Uint64(0)
	mantleHelpers.ApplyLimbTimeOffset(dp, &offset)

	// Use SetupMantleNormal to preserve manual fork configuration
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	// Verify fork time ordering
	logger.Info("Mantle Fork time configuration",
		"BaseFee", sd.RollupCfg.MantleBaseFeeTime,
		"Everest", sd.RollupCfg.MantleEverestTime,
		"Euboea", sd.RollupCfg.MantleEuboeaTime,
		"Skadi", sd.RollupCfg.MantleSkadiTime,
		"Limb", sd.RollupCfg.MantleLimbTime,
		"Arsia", sd.RollupCfg.MantleArsiaTime)

	// Verify Mantle fork time ordering
	if sd.RollupCfg.MantleBaseFeeTime != nil && sd.RollupCfg.MantleEverestTime != nil {
		require.LessOrEqual(t, *sd.RollupCfg.MantleBaseFeeTime, *sd.RollupCfg.MantleEverestTime,
			"BaseFee should be active before or at the same time as Everest.")
	}

	if sd.RollupCfg.MantleEverestTime != nil && sd.RollupCfg.MantleEuboeaTime != nil {
		require.LessOrEqual(t, *sd.RollupCfg.MantleEverestTime, *sd.RollupCfg.MantleEuboeaTime,
			"Everest should be active before or at the same time as Euboea")
	}

	if sd.RollupCfg.MantleEuboeaTime != nil && sd.RollupCfg.MantleSkadiTime != nil {
		require.LessOrEqual(t, *sd.RollupCfg.MantleEuboeaTime, *sd.RollupCfg.MantleSkadiTime,
			"Euboea should be active before or at the same time as Skadi")
	}

	if sd.RollupCfg.MantleSkadiTime != nil && sd.RollupCfg.MantleLimbTime != nil {
		require.LessOrEqual(t, *sd.RollupCfg.MantleSkadiTime, *sd.RollupCfg.MantleLimbTime,
			"Skadi should be active before or at the same time as Limb")
	}

	if sd.RollupCfg.MantleLimbTime != nil && sd.RollupCfg.MantleArsiaTime != nil {
		require.LessOrEqual(t, *sd.RollupCfg.MantleLimbTime, *sd.RollupCfg.MantleArsiaTime,
			"Limb should be active before or at the same time as Arsia")
	}

	logger.Info("Mantle Fork time ordering test passed.")
}
