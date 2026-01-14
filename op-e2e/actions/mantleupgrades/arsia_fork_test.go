package mantleupgrades

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	mantleHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestArsiaActivationAtGenesis tests that Arsia fork activates at genesis
// and all OP Stack forks are mapped to Arsia
func TestArsiaActivationAtGenesis(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Testing Arsia in Genesis activation ==========")

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

	// Apply Arsia activation at genesis
	offset := hexutil.Uint64(0)
	mantleHelpers.ApplyArsiaTimeOffset(dp, &offset)

	// Create test environment with Mantle-specific setup
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	logger.Info("The test environment has been created.",
		"L1ChainID", sd.L1Cfg.Config.ChainID,
		"L2ChainID", sd.L2Cfg.Config.ChainID)

	// Verify Arsia fork configuration
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.NotNil(t, sd.RollupCfg.MantleArsiaTime, "Arsia fork should be configured")

	// When offset=0, fork time is set to 0 (special marker for "genesis activation")
	require.Equal(t, uint64(0), *sd.RollupCfg.MantleArsiaTime, "Arsia fork offset=0 should be marked as timestamp 0")

	// Verify Arsia is active at genesis
	isArsiaActive := sd.RollupCfg.IsMantleArsia(genesisTime)
	require.True(t, isArsiaActive, "Arsia fork should be active at genesis")
	logger.Info("Arsia fork in genesis has been activated")

	// Verify all OP Stack forks are mapped to Arsia
	logger.Info("========== Verifying OP Stack Fork mappings ==========")

	logger.Info("OP Stack Fork time configurations",
		"Canyon", sd.RollupCfg.CanyonTime,
		"Delta", sd.RollupCfg.DeltaTime,
		"Ecotone", sd.RollupCfg.EcotoneTime,
		"Fjord", sd.RollupCfg.FjordTime,
		"Granite", sd.RollupCfg.GraniteTime,
		"Holocene", sd.RollupCfg.HoloceneTime,
		"Isthmus", sd.RollupCfg.IsthmusTime,
		"Jovian", sd.RollupCfg.JovianTime)

	// Verify each OP Stack fork is mapped to Arsia time
	if sd.RollupCfg.CanyonTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.CanyonTime,
			"Canyon should be mapped to Arsia")
	}
	if sd.RollupCfg.DeltaTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.DeltaTime,
			"Delta should be mapped to Arsia")
	}
	if sd.RollupCfg.EcotoneTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.EcotoneTime,
			"Ecotone should be mapped to Arsia")
	}
	if sd.RollupCfg.FjordTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.FjordTime,
			"Fjord should be mapped to Arsia")
	}
	if sd.RollupCfg.GraniteTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.GraniteTime,
			"Granite should be mapped to Arsia")
	}
	if sd.RollupCfg.HoloceneTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.HoloceneTime,
			"Holocene should be mapped to Arsia")
	}
	if sd.RollupCfg.IsthmusTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.IsthmusTime,
			"Isthmus should be mapped to Arsia")
	}
	if sd.RollupCfg.JovianTime != nil {
		require.Equal(t, *sd.RollupCfg.MantleArsiaTime, *sd.RollupCfg.JovianTime,
			"Jovian should be mapped to Arsia")
	}

	logger.Info(" All OP Stack forks are mapped to Arsia")

	// Verify EIP-1559 parameters
	logger.Info("========== Verifying EIP-1559 Parameters ==========")

	require.NotNil(t, sd.RollupCfg.ChainOpConfig, "ChainOpConfig should not be nil")
	elasticity := sd.RollupCfg.ChainOpConfig.EIP1559Elasticity
	denominator := sd.RollupCfg.ChainOpConfig.EIP1559Denominator

	logger.Info("EIP-1559 Parameters",
		"Elasticity", elasticity,
		"Denominator", denominator)

	require.Equal(t, uint64(4), elasticity, "Arsia version Elasticity should be 4")
	require.Equal(t, uint64(50), denominator, "Arsia version Denominator should be 50")

	logger.Info("EIP-1559 Parameters match Arsia version")

	// Verify that MaxSequencerDrift is hardcoded as 1800.
	logger.Info("========== Verifying MaxSequencerDrift (Fjord Feature) ==========")

	maxDrift := sd.ChainSpec.MaxSequencerDrift(genesisTime)

	// Arsia includes Fjord, so MaxSequencerDrift should be 1800
	require.Equal(t, uint64(1800), maxDrift,
		"Arsia version (includes Fjord) MaxSequencerDrift should be hardcoded as 1800")

	logger.Info("MaxSequencerDrift is hardcoded as 1800 (Fjord Feature)",
		"configValue", sd.RollupCfg.MaxSequencerDrift,
		"actualValue", maxDrift)

	// Create actors and verify chain works
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, logger)

	sequencer.ActL2PipelineFull(t)
	require.Equal(t, uint64(0), sequencer.SyncStatus().UnsafeL2.Number, "Initial L2 block should be 0")

	// Build L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2StartBlock(t)
	sequencer.ActL2EndBlock(t)

	l2Status := sequencer.SyncStatus()
	require.Greater(t, l2Status.UnsafeL2.Number, uint64(0), "L2 block should be generated")

	// Verify block ExtraData format (Arsia version should have MinBaseFee ExtraData)
	// Arsia includes Jovian fork which uses MinBaseFee ExtraData (17 bytes)
	logger.Info("========== Verify the ExtraData format of the block. ==========")

	l2Block := seqEngine.L2Chain().GetBlockByNumber(l2Status.UnsafeL2.Number)
	require.NotNil(t, l2Block, "L2 blocks should not be empty.")

	extraData := l2Block.Extra()
	logger.Info("L2 block ExtraData",
		"length", len(extraData))

	currentTime := l2Status.UnsafeL2.Time
	isJovian := sd.RollupCfg.IsJovian(currentTime)
	isHolocene := sd.RollupCfg.IsHolocene(currentTime)

	if isJovian {
		// Jovian/Arsia is active, ExtraData should be MinBaseFee format (17 bytes)
		require.Equal(t, 17, len(extraData), "Arsia version (Jovian active) ExtraData should be 17 bytes (MinBaseFee format)")

		// Parse MinBaseFee ExtraData: [1 byte version] + [4 bytes denominator] + [4 bytes elasticity] + [8 bytes minBaseFee]
		version := extraData[0]
		denominator := uint32(extraData[1])<<24 | uint32(extraData[2])<<16 |
			uint32(extraData[3])<<8 | uint32(extraData[4])
		elasticity := uint32(extraData[5])<<24 | uint32(extraData[6])<<16 |
			uint32(extraData[7])<<8 | uint32(extraData[8])
		minBaseFee := uint64(extraData[9])<<56 | uint64(extraData[10])<<48 |
			uint64(extraData[11])<<40 | uint64(extraData[12])<<32 |
			uint64(extraData[13])<<24 | uint64(extraData[14])<<16 |
			uint64(extraData[15])<<8 | uint64(extraData[16])

		logger.Info("ExtraData Details (MinBaseFee format)",
			"total_length", len(extraData),
			"version", version,
			"denominator", denominator,
			"elasticity", elasticity,
			"minBaseFee", minBaseFee)

		logger.Info("ExtraData format is MinBaseFee (17 bytes)")
	} else if isHolocene {
		// Only Holocene is active (without Jovian), ExtraData should be 9 bytes
		require.Equal(t, len(extraData), 9, "Holocene ExtraData should be 9 bytes")

		version := extraData[0]
		denominator := uint32(extraData[1])<<24 | uint32(extraData[2])<<16 |
			uint32(extraData[3])<<8 | uint32(extraData[4])
		elasticity := uint32(extraData[5])<<24 | uint32(extraData[6])<<16 |
			uint32(extraData[7])<<8 | uint32(extraData[8])

		logger.Info("ExtraData Details (Holocene format)",
			"total_length", len(extraData),
			"version", version,
			"denominator", denominator,
			"elasticity", elasticity)

		logger.Info("ExtraData format is Holocene (9 bytes)")
	} else {
		require.Equal(t, 0, len(extraData), "Non-Holocene/Jovian version ExtraData should be 0 bytes")
		logger.Info("ℹ️  Holocene/Jovian not active - ExtraData is 0 bytes")
	}

	logger.Info("Arsia fork genesis activation test passed")
}

// TestArsiaLateActivation tests that Arsia fork can be activated after genesis
func TestArsiaLateActivation(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Testing delayed activation of Arsia fork ==========")

	// Configure to activate Arsia at offset 24 (after first L1 block)
	arsiaOffset := hexutil.Uint64(24)

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Apply Arsia activation at offset 24
	mantleHelpers.ApplyArsiaTimeOffset(dp, &arsiaOffset)

	// Create test environment with Mantle-specific setup
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	logger.Info("The test environment has been created.",
		"L1ChainID", sd.L1Cfg.Config.ChainID,
		"L2ChainID", sd.L2Cfg.Config.ChainID,
		"ArsiaOffset", arsiaOffset)

	// Verify Arsia fork configuration
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.NotNil(t, sd.RollupCfg.MantleArsiaTime, "Arsia fork should have been configured.")
	expect_time := genesisTime + uint64(24)
	require.Equal(t, expect_time, *sd.RollupCfg.MantleArsiaTime, "Arsia fork should be activated at offset 24.")

	// Verify Arsia is NOT active at genesis
	isArsiaActiveAtGenesis := sd.RollupCfg.IsMantleArsia(genesisTime)
	require.False(t, isArsiaActiveAtGenesis, "Arsia fork should not be active at genesis.")
	logger.Info("Arsia fork should not be active at genesis.")

	// Verify OP Stack forks are also not active at genesis
	isHoloceneAtGenesis := sd.RollupCfg.IsHolocene(genesisTime)
	require.False(t, isHoloceneAtGenesis, "Holocene should not be active at genesis.")
	logger.Info("Holocene should not be active at genesis.")

	// Create actors
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, logger)

	sequencer.ActL2PipelineFull(t)

	// Build L1 block to activate Arsia
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)

	// Track pre-activation ExtraData
	sequencer.ActL2StartBlock(t)
	sequencer.ActL2EndBlock(t)

	preActivationStatus := sequencer.SyncStatus()
	preActivationBlock := seqEngine.L2Chain().GetBlockByNumber(preActivationStatus.UnsafeL2.Number)
	preActivationExtraData := preActivationBlock.Extra()

	logger.Info("Pre-activation block",
		"number", preActivationStatus.UnsafeL2.Number,
		"time", preActivationStatus.UnsafeL2.Time,
		"extraDataLen", len(preActivationExtraData))

	// Build L2 blocks until Arsia activation
	for i := 0; i < 30; i++ {
		sequencer.ActL2StartBlock(t)
		sequencer.ActL2EndBlock(t)

		currentTime := sequencer.SyncStatus().UnsafeL2.Time
		isArsiaActive := sd.RollupCfg.IsMantleArsia(currentTime)
		isHoloceneActive := sd.RollupCfg.IsHolocene(currentTime)

		logger.Info("L2 block generation",
			"number", sequencer.SyncStatus().UnsafeL2.Number,
			"time", currentTime,
			"isArsiaActive", isArsiaActive,
			"isHoloceneActive", isHoloceneActive)

		if isArsiaActive {
			logger.Info("Arsia fork is activated",
				"blockNumber", sequencer.SyncStatus().UnsafeL2.Number,
				"blockTime", currentTime)

			// Verify ExtraData changed after activation
			currentBlock := seqEngine.L2Chain().GetBlockByNumber(sequencer.SyncStatus().UnsafeL2.Number)
			currentExtraData := currentBlock.Extra()

			logger.Info("Post-activation block",
				"extraDataLen", len(currentExtraData))

			if isHoloceneActive {
				require.NotEqual(t, len(preActivationExtraData), len(currentExtraData),
					"ExtraData length should change after Arsia/Holocene activation.")
				require.GreaterOrEqual(t, len(currentExtraData), 9,
					"Holocene ExtraData should have at least 9 bytes.")
			}

			break
		}
	}

	// Verify Arsia is now active
	finalTime := sequencer.SyncStatus().UnsafeL2.Time
	isArsiaActiveFinal := sd.RollupCfg.IsMantleArsia(finalTime)
	require.True(t, isArsiaActiveFinal, "Arsia fork should be active after activation.")

	logger.Info("Arsia fork late activation test passed")
}

// TestArsiaIncludesAllOPStackForks verifies that activating Arsia
// automatically activates all included OP Stack forks
func TestArsiaIncludesAllOPStackForks(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== The test for Arsia includes all OP stack forks. ==========")

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	// Use MakeMantleDeployParams for Mantle-specific configuration
	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Activate Arsia at genesis
	offset := hexutil.Uint64(0)
	mantleHelpers.ApplyArsiaTimeOffset(dp, &offset)

	// Use SetupMantleNormal to preserve manual fork configuration
	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)

	genesisTime := sd.RollupCfg.Genesis.L2Time

	// Verify all OP Stack forks are active at genesis
	logger.Info("Verify OP Stack fork activation status at genesis.")

	// Canyon
	if sd.RollupCfg.CanyonTime != nil {
		isActive := sd.RollupCfg.IsCanyon(genesisTime)
		require.True(t, isActive, "Canyon should be active at genesis.")
		logger.Info("Canyon is active at genesis.")
	}

	// Delta
	if sd.RollupCfg.DeltaTime != nil {
		isActive := sd.RollupCfg.IsDelta(genesisTime)
		require.True(t, isActive, "Delta should be active at genesis.")
		logger.Info("Delta is active at genesis.")
	}

	// Ecotone
	if sd.RollupCfg.EcotoneTime != nil {
		isActive := sd.RollupCfg.IsEcotone(genesisTime)
		require.True(t, isActive, "Ecotone should be active at genesis.")
		logger.Info("Ecotone is active at genesis.")
	}

	// Fjord
	if sd.RollupCfg.FjordTime != nil {
		isActive := sd.RollupCfg.IsFjord(genesisTime)
		require.True(t, isActive, "Fjord should be active at genesis.")
		logger.Info("Fjord is active at genesis.")
	}

	// Granite
	if sd.RollupCfg.GraniteTime != nil {
		isActive := sd.RollupCfg.IsGranite(genesisTime)
		require.True(t, isActive, "Granite should be active at genesis.")
		logger.Info("Granite is active at genesis.")
	}

	// Holocene
	if sd.RollupCfg.HoloceneTime != nil {
		isActive := sd.RollupCfg.IsHolocene(genesisTime)
		require.True(t, isActive, "Holocene should be active at genesis.")
		logger.Info("Holocene is active at genesis.")
	}

	// Isthmus
	if sd.RollupCfg.IsthmusTime != nil {
		isActive := sd.RollupCfg.IsIsthmus(genesisTime)
		require.True(t, isActive, "Isthmus should be active at genesis.")
		logger.Info("Isthmus is active at genesis.")
	}

	// Jovian
	if sd.RollupCfg.JovianTime != nil {
		isActive := sd.RollupCfg.IsJovian(genesisTime)
		require.True(t, isActive, "Jovian should be active at genesis.")
		logger.Info("Jovian is active at genesis.")
	}

	logger.Info("Arsia includes all OP Stack forks test passed.")
}

func TestArsiaInvalidPayload(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	env := helpers.SetupEnvMantle(t, helpers.WithActiveMantleGenesisFork(forks.MantleArsia))
	ctx := context.Background()

	requireDepositOnlyLogs := func(role string, expNumLogs int) {
		t.Helper()
		recs := env.Logs.FindLogs(testlog.NewMessageContainsFilter("deposits-only attributes"), testlog.NewAttributesFilter("role", role))
		require.Len(t, recs, expNumLogs)
	}

	// Start op-nodes
	env.Seq.ActL2PipelineFull(t)

	// generate and batch buffer two empty blocks
	env.Seq.ActL2EmptyBlock(t) // 1 - genesis is 0
	env.Batcher.ActL2BatchBuffer(t)
	env.Seq.ActL2EmptyBlock(t) // 2
	env.Batcher.ActL2BatchBuffer(t)

	// send and include a single transaction
	env.Alice.L2.ActResetTxOpts(t)
	env.Alice.L2.ActSetTxToAddr(&env.DeployParams.Addresses.Bob)
	env.Alice.L2.ActMakeTx(t)

	env.Seq.ActL2StartBlock(t)
	env.SeqEngine.ActL2IncludeTx(env.Alice.Address())(t)
	env.Seq.ActL2EndBlock(t) // 3
	env.Alice.L2.ActCheckReceiptStatusOfLastTx(true)(t)
	l2Unsafe := env.Seq.L2Unsafe()
	const invalidNum = 3
	require.EqualValues(t, invalidNum, l2Unsafe.Number)
	b, err := env.SeqEngine.EthClient().BlockByNumber(ctx, big.NewInt(invalidNum))
	require.NoError(t, err)
	require.Len(t, b.Transactions(), 2)

	// buffer into the batcher, invalidating the tx via signature zeroing
	env.Batcher.ActL2BatchBuffer(t, helpers.WithBlockModifier(func(block *types.Block) *types.Block {
		// Replace the tx with one that has a bad signature.
		txs := block.Transactions()
		newTx, err := txs[1].WithSignature(env.Alice.L2.Signer(), make([]byte, 65))
		require.NoError(t, err)
		txs[1] = newTx
		return block
	}))

	// generate two more empty blocks
	env.Seq.ActL2EmptyBlock(t) // 4
	env.Seq.ActL2EmptyBlock(t) // 5
	require.EqualValues(t, 5, env.Seq.L2Unsafe().Number)

	// submit it all
	env.ActBatchSubmitAllAndMine(t)

	// derive chain on sequencer
	env.Seq.ActL1HeadSignal(t)
	env.Seq.ActL2PipelineFull(t)

	l2Safe := env.Seq.L2Safe()
	require.EqualValues(t, invalidNum, l2Safe.Number)
	require.NotEqual(t, l2Safe.Hash, l2Unsafe.Hash, // old L2Unsafe above
		"block-3 should have been replaced by deposit-only version")
	requireDepositOnlyLogs(e2esys.RoleSeq, 2)
	require.Equal(t, l2Safe, env.Seq.L2Unsafe(), "unsafe chain should have reorg'd")
	b, err = env.SeqEngine.EthClient().BlockByNumber(ctx, big.NewInt(invalidNum))
	require.NoError(t, err)
	require.Len(t, b.Transactions(), 1)

	// test that building on top of reorg'd chain and deriving further works

	env.Seq.ActL2EmptyBlock(t) // 4
	env.Seq.ActL2EmptyBlock(t) // 5
	l2Unsafe = env.Seq.L2Unsafe()
	require.EqualValues(t, 5, l2Unsafe.Number)

	env.Batcher.Reset() // need to reset batcher to become aware of reorg
	env.ActBatchSubmitAllAndMine(t)
	env.Seq.ActL1HeadSignal(t)
	env.Seq.ActL2PipelineFull(t)
	require.Equal(t, l2Unsafe, env.Seq.L2Safe())
}

// TestArsiaNetworkUpgradeTransactions tests that the Arsia network upgrade
// transactions are properly executed during the Arsia activation block.
// Mantle Arsia includes 7 upgrade transactions that deploy and upgrade:
// - L1Block (with operator fee support)
// - GasPriceOracle (with Arsia fee calculation)
// - OperatorFeeVault (Mantle-specific feature)
func TestArsiaNetworkUpgradeTransactions(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	logger := testlog.Logger(t, log.LevelInfo)

	logger.Info("========== Testing the Arsia network upgrade transaction ==========")

	testParams := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}

	dp := e2eutils.MakeMantleDeployParams(t, testParams)

	// Activate all Mantle forks before Arsia at genesis, schedule Arsia for block after genesis
	arsiaOffset := hexutil.Uint64(2)
	mantleHelpers.ApplyArsiaTimeOffset(dp, &arsiaOffset)

	require.NoError(t, dp.DeployConfig.MantleCheck(logger), "must have valid config")

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	_, _, _, sequencer, engine, verifier, _, _ := helpers.SetupReorgTestActors(t, dp, sd, logger)
	ethCl := engine.EthClient()

	// Start op-nodes
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	logger.Info("Verifying initial state")

	// Get initial implementation addresses for L1Block, GasPriceOracle, and OperatorFeeVault
	initialL1BlockAddress, err := ethCl.StorageAt(context.Background(), predeploys.L1BlockAddr, genesis.ImplementationSlot, nil)
	require.NoError(t, err)

	initialGasPriceOracleAddress, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, genesis.ImplementationSlot, nil)
	require.NoError(t, err)

	initialOperatorFeeVaultAddress, err := ethCl.StorageAt(context.Background(), predeploys.OperatorFeeVaultAddr, genesis.ImplementationSlot, nil)
	require.NoError(t, err)

	logger.Info("Initial state recorded",
		"L1BlockImpl", common.BytesToAddress(initialL1BlockAddress).Hex(),
		"GasPriceOracleImpl", common.BytesToAddress(initialGasPriceOracleAddress).Hex(),
		"OperatorFeeVaultImpl", common.BytesToAddress(initialOperatorFeeVaultAddress).Hex())

	// DEBUG: Check storage before Arsia activation
	logger.Info("========== DEBUG:  GasPriceOracle storage before Arsia activation ==========")
	slot0, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(0)), nil)
	require.NoError(t, err)
	slot1, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(1)), nil)
	require.NoError(t, err)
	slot2, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(2)), nil)
	require.NoError(t, err)
	slot3, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(3)), nil)
	require.NoError(t, err)

	// IMPORTANT: Solidity packs variables in BIG-ENDIAN order!
	// In Ethereum storage, rightmost bytes = low-order bytes
	// slot 2 layout (big-endian): [unused 11 bytes (indices 0-10)] [isArsia 1 byte (index 11)] [operator 20 bytes (indices 12-31)]
	operatorAddr := common.BytesToAddress(slot2[12:32]) // bytes 12-31 (rightmost 20 bytes)
	isArsiaFromSlot2 := slot2[11] != 0                  // byte 11 (the byte just left of operator)

	logger.Info("GasPriceOracle storage before Arsia activation",
		"slot0_tokenRatio", common.BytesToHash(slot0).Hex(),
		"slot1_owner", common.BytesToAddress(slot1[:]).Hex(),
		"slot2_raw", common.BytesToHash(slot2).Hex(),
		"slot2_operator", operatorAddr.Hex(),
		"slot2_isArsia_byte11", slot2[11],
		"slot2_isArsia_bool", isArsiaFromSlot2,
		"slot3", common.BytesToHash(slot3).Hex())

	// Build to Arsia activation block
	logger.Info("Generating L1 and L2 blocks until Arsia activation")

	// DEBUG: Record block number before building
	blockNumberBefore := sequencer.L2Unsafe().Number
	logger.Info("Block number before building", "blockNumber", blockNumberBefore)

	sequencer.ActBuildL2ToArsia(t)

	// DEBUG: Record block number after building
	blockNumberAfter := sequencer.L2Unsafe().Number
	logger.Info("Block number after building", "blockNumber", blockNumberAfter, "newBlocks", blockNumberAfter-blockNumberBefore)

	// Get latest block (should be the Arsia activation block)
	latestBlock, err := ethCl.BlockByNumber(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, sequencer.L2Unsafe().Number, latestBlock.Number().Uint64())

	logger.Info("Arsia activation block generated",
		"blockNumber", latestBlock.Number().Uint64(),
		"txCount", len(latestBlock.Transactions()))

	// DEBUG: Check if there are other recent blocks with upgrade transactions
	if blockNumberAfter > blockNumberBefore+1 {
		logger.Info("========== DEBUG: Found multiple new blocks, checking each block ==========")
		for bn := blockNumberBefore + 1; bn <= blockNumberAfter; bn++ {
			block, err := ethCl.BlockByNumber(context.Background(), big.NewInt(int64(bn)))
			require.NoError(t, err)
			logger.Info("Block details",
				"blockNumber", bn,
				"timestamp", block.Time(),
				"txCount", len(block.Transactions()),
				"hash", block.Hash().Hex())
		}
	}

	transactions := latestBlock.Transactions()
	// Expected: 1 set-L1-info + 7 upgrade transactions
	// See [derive.MantleArsiaNetworkUpgradeTransactions]:
	// 1. Deploy L1Block implementation
	// 2. Deploy GasPriceOracle implementation
	// 3. Deploy OperatorFeeVault implementation
	// 4. Upgrade L1Block proxy
	// 5. Upgrade GasPriceOracle proxy
	// 6. Upgrade OperatorFeeVault proxy
	// 7. Enable Arsia mode in GasPriceOracle
	require.Equal(t, 8, len(transactions), "There should be 8 transactions: 1 set-L1-info + 7 upgrade txs")

	logger.Info("Transaction volume verification passed", "count", len(transactions))

	// DEBUG: Check if there are any duplicate setArsia calls
	logger.Info("========== DEBUG: Check for duplicate setArsia() calls ==========")
	setArsiaSelector := common.FromHex("0x8f018a7b")
	setArsiaCount := 0
	for i, tx := range transactions {
		if tx.To() != nil && *tx.To() == predeploys.GasPriceOracleAddr && len(tx.Data()) >= 4 {
			if common.BytesToHash(tx.Data()[:4]) == common.BytesToHash(setArsiaSelector) {
				setArsiaCount++
				logger.Info("setArsia() call found",
					"index", i,
					"hash", tx.Hash().Hex(),
					"from", tx.From().Hex(),
					"to", tx.To().Hex())
			}
		}
	}
	logger.Info("Total setArsia() calls found", "count", setArsiaCount)

	//  DEBUG: Check implementation after transaction 5 (GasPriceOracle proxy upgrade)
	logger.Info("========== DEBUG: Check implementation address after transaction 5 (GasPriceOracle proxy upgrade) ==========")

	// Verify all upgrade transactions are successful
	logger.Info("Verify all upgrade transactions are successful")
	for i := 1; i < 8; i++ {
		txn := transactions[i]
		receipt, err := ethCl.TransactionReceipt(context.Background(), txn.Hash())
		require.NoError(t, err)

		// DEBUG: Immediately check isArsia value after each transaction
		slot2AfterTx, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(2)), latestBlock.Number())
		require.NoError(t, err)
		isArsiaAfterTx := slot2AfterTx[11] != 0
		logger.Info("isArsia status after transaction", "txIndex", i, "isArsia", isArsiaAfterTx, "slot2_byte11", slot2AfterTx[11])

		// DEBUG: Print detailed info for each transaction
		toAddr := "nil (contract deployment)"
		if txn.To() != nil {
			toAddr = txn.To().Hex()
		}
		dataHex := hexutil.Encode(txn.Data())
		if len(dataHex) > 66 {
			dataHex = dataHex[:66] + "..."
		}
		logger.Info("Transaction details",
			"txIndex", i,
			"hash", txn.Hash().Hex(),
			"from", txn.From().Hex(),
			"to", toAddr,
			"data", dataHex,
			"status", receipt.Status,
			"gasUsed", receipt.GasUsed)

		if receipt.Status != types.ReceiptStatusSuccessful {
			logger.Error("Transaction failed",
				"txIndex", i,
				"status", receipt.Status,
				"expectedStatus", types.ReceiptStatusSuccessful)
		}

		require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status,
			"Transaction %d should be successful", i)
		require.NotEmpty(t, txn.Data(), "Transaction must provide input data")

		// DEBUG: After transaction 5, check if GasPriceOracle proxy upgraded
		if i == 5 {
			implAfterTx5, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, genesis.ImplementationSlot, latestBlock.Number())
			require.NoError(t, err)
			logger.Info("GasPriceOracle implementation address after transaction 5",
				"implementation", common.BytesToAddress(implAfterTx5).Hex(),
				"expected", crypto.CreateAddress(derive.GasPriceOracleArsiaDeployerAddress, 0).Hex())

			// Check storage again
			slot3After, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(3)), latestBlock.Number())
			require.NoError(t, err)
			logger.Info("slot3 value after transaction 5", "slot3", common.BytesToHash(slot3After).Hex())
		}

		// DEBUG: Before verifying transaction 7, check slot2 and slot3 value
		if i == 6 {
			logger.Info("========== DEBUG: Check storage after transaction 6 ==========")
			slot2AfterTx6, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(2)), latestBlock.Number())
			require.NoError(t, err)
			slot3AfterTx6, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(3)), latestBlock.Number())
			require.NoError(t, err)

			// Print all non-zero bytes to find where isArsia actually is
			logger.Info("========== DEBUG: Slot 2 all bytes ==========")
			for i := 0; i < 32; i++ {
				if slot2AfterTx6[i] != 0 {
					logger.Info("Non-zero byte", "index", i, "value", slot2AfterTx6[i])
				}
			}

			// According to Forge: operator at offset 0 (20 bytes), isArsia at offset 20 (1 byte)
			// In Solidity packing: offset from RIGHT (low bytes)
			// So operator = bytes 12-31, isArsia = byte 11
			operatorAfter := common.BytesToAddress(slot2AfterTx6[12:32])
			isArsiaByte11 := slot2AfterTx6[11] != 0
			isArsiaByte10 := slot2AfterTx6[10] != 0

			logger.Info("Storage after transaction 6",
				"slot2_raw", common.BytesToHash(slot2AfterTx6).Hex(),
				"slot2_operator", operatorAfter.Hex(),
				"slot2_byte10", slot2AfterTx6[10],
				"slot2_byte11", slot2AfterTx6[11],
				"isArsia_if_byte10", isArsiaByte10,
				"isArsia_if_byte11", isArsiaByte11,
				"slot3", common.BytesToHash(slot3AfterTx6).Hex())

			// Call isArsia() getter to see what the contract returns
			logger.Info("========== DEBUG: Call isArsia() getter function ==========")
			isArsiaSelector := common.FromHex("0x1e2b6e7b") // isArsia() function selector
			isArsiaCallMsg := ethereum.CallMsg{
				To:   &predeploys.GasPriceOracleAddr,
				Gas:  100000,
				Data: isArsiaSelector,
			}
			isArsiaResult, err := ethCl.CallContract(context.Background(), isArsiaCallMsg, latestBlock.Number())
			if err != nil {
				logger.Error("isArsia() call failed", "error", err.Error())
			} else {
				isArsiaValue := new(big.Int).SetBytes(isArsiaResult).Uint64() != 0
				logger.Info("isArsia() return value",
					"rawResult", hexutil.Encode(isArsiaResult),
					"boolValue", isArsiaValue)
			}

			// Try to simulate the call to see what would happen
			logger.Info("========== DEBUG: Simulate setArsia() call to get revert reason ==========")
			setArsiaData := common.FromHex("0x8f018a7b") // setArsia() selector
			callMsg := ethereum.CallMsg{
				From: common.HexToAddress("0xDeaDDEaDDeAdDeAdDEAdDEaddeAddEAdDEAd0001"),
				To:   &predeploys.GasPriceOracleAddr,
				Gas:  300000,
				Data: setArsiaData,
			}

			result, err := ethCl.CallContract(context.Background(), callMsg, latestBlock.Number())
			if err != nil {
				logger.Error("setArsia() simulation call failed",
					"error", err.Error(),
					"result", hexutil.Encode(result))
			} else {
				logger.Info("setArsia() simulation call successful",
					"result", hexutil.Encode(result))
			}
		}

		// 🔍 DEBUG: After verifying transaction 7, check what slot3 is
		if i == 7 {
			logger.Info("========== DEBUG: Check slot3 value after transaction 7 ==========")
			slot3AfterTx7, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, common.BigToHash(big.NewInt(3)), latestBlock.Number())
			require.NoError(t, err)
			logger.Info("slot3 value after transaction 7",
				"slot3", common.BytesToHash(slot3AfterTx7).Hex(),
				"bool", new(big.Int).SetBytes(slot3AfterTx7).Uint64() != 0,
				"status", receipt.Status)
		}
	}

	logger.Info("All upgrade transactions executed successfully")

	// Verify proxy addresses are updated
	logger.Info("Verify proxy addresses are updated")

	expectedL1BlockAddress := crypto.CreateAddress(derive.L1BlockArsiaDeployerAddress, 0)
	updatedL1BlockAddress, err := ethCl.StorageAt(context.Background(), predeploys.L1BlockAddr, genesis.ImplementationSlot, latestBlock.Number())
	require.NoError(t, err)
	require.Equal(t, expectedL1BlockAddress, common.BytesToAddress(updatedL1BlockAddress))
	require.NotEqualf(t, initialL1BlockAddress, updatedL1BlockAddress, "L1Block Proxy address should have been updated")

	logger.Info("L1Block proxy address updated",
		"expected", expectedL1BlockAddress.Hex(),
		"actual", common.BytesToAddress(updatedL1BlockAddress).Hex())

	expectedGasPriceOracleAddress := crypto.CreateAddress(derive.GasPriceOracleArsiaDeployerAddress, 0)
	updatedGasPriceOracleAddress, err := ethCl.StorageAt(context.Background(), predeploys.GasPriceOracleAddr, genesis.ImplementationSlot, latestBlock.Number())
	require.NoError(t, err)
	require.Equal(t, expectedGasPriceOracleAddress, common.BytesToAddress(updatedGasPriceOracleAddress))
	require.NotEqualf(t, initialGasPriceOracleAddress, updatedGasPriceOracleAddress, "GasPriceOracle Proxy address should have been updated")

	logger.Info("GasPriceOracle proxy address updated",
		"expected", expectedGasPriceOracleAddress.Hex(),
		"actual", common.BytesToAddress(updatedGasPriceOracleAddress).Hex())

	expectedOperatorFeeVaultAddress := crypto.CreateAddress(derive.OperatorFeeVaultArsiaDeployerAddress, 0)
	updatedOperatorFeeVaultAddress, err := ethCl.StorageAt(context.Background(), predeploys.OperatorFeeVaultAddr, genesis.ImplementationSlot, latestBlock.Number())
	require.NoError(t, err)
	require.Equal(t, expectedOperatorFeeVaultAddress, common.BytesToAddress(updatedOperatorFeeVaultAddress))
	require.NotEqualf(t, initialOperatorFeeVaultAddress, updatedOperatorFeeVaultAddress, "OperatorFeeVault Proxy address should have been updated")

	logger.Info("OperatorFeeVault proxy address updated",
		"expected", expectedOperatorFeeVaultAddress.Hex(),
		"actual", common.BytesToAddress(updatedOperatorFeeVaultAddress).Hex())

	logger.Info("All proxy contract addresses updated")

	logger.Info("========================================")
	logger.Info("Arsia network upgrade transaction test completed")
	logger.Info("========================================")
	logger.Info("")
	logger.Info("Test summary:")
	logger.Info("  Arsia activation block contains 8 transactions (1 set-L1-info + 7 upgrade txs)")
	logger.Info("  All upgrade transactions executed successfully")
	logger.Info("  L1Block proxy address updated")
	logger.Info("  GasPriceOracle proxy address updated")
	logger.Info("  OperatorFeeVault proxy address updated")
	logger.Info("")
}
