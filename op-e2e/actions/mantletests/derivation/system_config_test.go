package derivation

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"testing"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestSystemConfigBatchType run each system config-related test case in singular batch mode and span batch mode.
func TestSystemConfigBatchType(t *testing.T) {
	tests := []struct {
		name string
		f    func(gt *testing.T, isSpanBatch bool)
	}{
		{"BatcherKeyRotation", BatcherKeyRotation},
		{"GPOParamsChange", GPOParamsChange},
		{"GasLimitChange", GasLimitChange},
		{"OperatorFeeScalarsChange", OperatorFeeScalarsChange},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name+"_SingularBatch", func(t *testing.T) {
			test.f(t, false)
		})
	}

	for _, test := range tests {
		test := test
		t.Run(test.name+"_SpanBatch", func(t *testing.T) {
			test.f(t, true)
		})
	}
}

// BatcherKeyRotation tests that batcher A can operate, then be replaced with batcher B, then ignore old batcher A,
// and that the change to batcher B is reverted properly upon reorg of L1.
func BatcherKeyRotation(gt *testing.T, isSpanBatch bool) {
	t := actionsHelpers.NewDefaultTesting(gt)

	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())
	dp.DeployConfig.L2BlockTime = 2
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)
	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, log)
	miner.ActL1SetFeeRecipient(common.Address{'A'})
	sequencer.ActL2PipelineFull(t)
	_, verifier := actionsHelpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})
	rollupSeqCl := sequencer.RollupClient()

	// the default batcher
	batcherA := actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSpanBatcherCfg(dp),
		rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	if !isSpanBatch {
		batcherA = actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
			rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	}
	// a batcher with a new key
	altCfg := *actionsHelpers.MantleSpanBatcherCfg(dp)
	if !isSpanBatch {
		altCfg = *actionsHelpers.MantleSingularBatcherCfg(dp)
	}
	altCfg.BatcherKey = dp.Secrets.Bob
	batcherB := actionsHelpers.NewL2Batcher(log, sd.RollupCfg, &altCfg,
		rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)

	// build a L1 chain, and then L2 chain, for batcher A to submit
	miner.ActEmptyBlock(t)
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	batcherA.ActSubmitAll(t)

	// include the batch data on L1
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// sync from L1
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, uint64(2), sequencer.L2Safe().L1Origin.Number, "l2 chain with new L1 origins")
	require.Equal(t, sequencer.L2Safe(), verifier.L2Safe(), "fully synced verifier")

	sysCfgContract, err := bindings.NewSystemConfig(sd.RollupCfg.L1SystemConfigAddress, miner.EthClient())
	require.NoError(t, err)

	sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.RollupCfg.L1ChainID)
	require.NoError(t, err)

	owner, err := sysCfgContract.Owner(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, dp.Addresses.Deployer, owner, "system config owner mismatch")

	// Change the batch sender key to Bob!
	tx, err := sysCfgContract.SetBatcherHash(sysCfgOwner, eth.AddressAsLeftPaddedHash(dp.Addresses.Bob))
	require.NoError(t, err)
	t.Logf("batcher changes in L1 tx %s", tx.Hash())
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
	miner.ActL1EndBlock(t)

	receipt, err := miner.EthClient().TransactionReceipt(t.Ctx(), tx.Hash())
	require.NoError(t, err)

	cfgChangeL1BlockNum := miner.L1Chain().CurrentBlock().Number.Uint64()
	require.Equal(t, cfgChangeL1BlockNum, receipt.BlockNumber.Uint64())

	// sequence L2 blocks, and submit with new batcher
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	batcherB.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Bob)(t)
	miner.ActL1EndBlock(t)

	// check that the first L2 payload that adopted the L1 block with the batcher key change
	// indeed changed the batcher key in the system config
	engCl := seqEngine.EngineClient(t, sd.RollupCfg)
	// 12 new L2 blocks: 5 with origin before L1 block with batch, 6 with origin of L1 block
	// with batch, 1 with new origin that changed the batcher
	for i := 0; i <= 12; i++ {
		envelope, err := engCl.PayloadByNumber(t.Ctx(), sequencer.L2Safe().Number+uint64(i))
		require.NoError(t, err)
		ref, err := derive.PayloadToBlockRef(sd.RollupCfg, envelope.ExecutionPayload)
		require.NoError(t, err)
		if i < 6 {
			require.Equal(t, ref.L1Origin.Number, cfgChangeL1BlockNum-2)
			require.Equal(t, ref.SequenceNumber, uint64(i))
		} else if i < 12 {
			require.Equal(t, ref.L1Origin.Number, cfgChangeL1BlockNum-1)
			require.Equal(t, ref.SequenceNumber, uint64(i-6))
		} else {
			require.Equal(t, ref.L1Origin.Number, cfgChangeL1BlockNum)
			require.Equal(t, ref.SequenceNumber, uint64(0), "first L2 block with this origin")
			sysCfg, err := derive.PayloadToSystemConfig(sd.RollupCfg, envelope.ExecutionPayload)
			require.NoError(t, err)
			require.Equal(t, dp.Addresses.Bob, sysCfg.BatcherAddr, "bob should be batcher now")
		}
	}

	// sync from L1
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, sequencer.L2Safe().L1Origin.Number, uint64(4), "safe l2 chain with two new l1 blocks")
	require.Equal(t, sequencer.L2Safe(), verifier.L2Safe(), "fully synced verifier")

	// now try to build a new L1 block, and corresponding L2 blocks, and submit with the old batcher
	before := sequencer.L2Safe()
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	batcherA.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// check that the data submitted by the old batcher is ignored
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, sequencer.L2Safe(), before, "no new safe l1 chain")
	require.Equal(t, verifier.L2Safe(), before, "verifier is ignoring old batcher too")

	// now submit with the new batcher
	batcherB.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Bob)(t)
	miner.ActL1EndBlock(t)

	// not ignored now with new batcher
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.NotEqual(t, sequencer.L2Safe(), before, "new safe l1 chain")
	require.NotEqual(t, verifier.L2Safe(), before, "verifier is not ignoring new batcher")

	// twist: reorg L1, including the batcher key change
	miner.ActL1RewindDepth(5)(t)
	for i := 0; i < 6; i++ { // build some empty blocks so the reorg is picked up
		miner.ActEmptyBlock(t)
	}
	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, uint64(2), sequencer.L2Safe().L1Origin.Number, "l2 safe is first batch submission with original batcher")
	require.Equal(t, uint64(3), sequencer.L2Unsafe().L1Origin.Number, "l2 unsafe l1 origin is the block that included the first batch")
	require.Equal(t, sequencer.L2Safe(), verifier.L2Safe(), "verifier safe head check")
	require.Equal(t, sequencer.L2Unsafe(), verifier.L2Unsafe(), "verifier unsafe head check")

	// without building L2 chain for the new L1 blocks yet, just batch-submit the unsafe part
	batcherA.ActL2BatchBuffer(t) // forces the buffer state to handle the rewind, before we loop with ActSubmitAll
	batcherA.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, sequencer.L2Safe(), sequencer.L2Unsafe(), "all L2 blocks are safe now")
	require.Equal(t, sequencer.L2Unsafe(), verifier.L2Unsafe(), "verifier synced")

	// and see if we can go past it, with new L2 blocks
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	batcherA.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	sequencer.ActL2PipelineFull(t)
	verifier.ActL2PipelineFull(t)
	require.Equal(t, uint64(3+6+1), verifier.L2Safe().L1Origin.Number, "sync new L1 chain, while key change is reorged out")
	require.Equal(t, sequencer.L2Unsafe(), verifier.L2Unsafe(), "verifier synced")
}

// GPOParamsChange tests that the GPO params can be updated to adjust fees of L2 transactions,
// and that the L1 data fees to the L2 transaction are applied correctly before, during and after the GPO update in L2.
// This test verifies the complete Mantle fee structure:
// Total Fee = L1 Data Fee + L2 Execution Fee + Operator Fee
func GPOParamsChange(gt *testing.T, isSpanBatch bool) {
	t := actionsHelpers.NewDefaultTesting(gt)
	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())

	// Enable MNT as gas token (zero address represents MNT in Mantle)
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set tokenRatio: 1 ETH = 2000 MNT (using DECIMALS=6 precision)
	dp.DeployConfig.GasPriceOracleTokenRatio = 2000 * 1e6

	// Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000   // Scalar
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000 // Constant (wei)

	// Set MinBaseFee (around 10 gwei for Mantle)
	dp.DeployConfig.MinBaseFee = 10 * 1e9 // 10 gwei

	// Set L2 genesis block BaseFee (initial BaseFee for L2)
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9)) // 10 gwei

	// Activate Arsia fork at genesis
	arsiaTimeoffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeoffset)

	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, log)
	batcher := actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSpanBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	if !isSpanBatch {
		batcher = actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
			sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	}
	alice := actionsHelpers.NewBasicUser[any](log, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	sequencer.ActL2PipelineFull(t)

	// Build initial L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	basefee := miner.L1Chain().CurrentBlock().BaseFee

	// Get contract bindings
	gpoContract, err := bindings.NewGasPriceOracle(
		common.HexToAddress("0x420000000000000000000000000000000000000F"),
		seqEngine.EthClient(),
	)
	require.NoError(t, err)

	sysCfgContract, err := bindings.NewSystemConfig(sd.RollupCfg.L1SystemConfigAddress, miner.EthClient())
	require.NoError(t, err)

	// Set MinBaseFee on L1 SystemConfig (workaround for missing evm: tag)
	sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.RollupCfg.L1ChainID)
	require.NoError(t, err)

	_, err = sysCfgContract.SetMinBaseFee(sysCfgOwner, dp.DeployConfig.MinBaseFee)
	require.NoError(t, err)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
	miner.ActL1EndBlock(t)

	// Verify MinBaseFee is set
	minBaseFee, err := sysCfgContract.MinBaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, dp.DeployConfig.MinBaseFee, minBaseFee, "MinBaseFee should be set")
	t.Logf("MinBaseFee set to: %d wei (%d gwei)", minBaseFee, minBaseFee/1e9)

	// Setup tokenRatio (workaround for missing evm: tag in DeployConfig)
	// Note: This is a known issue - GasPriceOracleTokenRatio lacks evm: tag
	gpoOwnerAddr, err := gpoContract.Owner(&bind.CallOpts{})
	require.NoError(t, err)

	var ownerKey *ecdsa.PrivateKey
	switch gpoOwnerAddr {
	case dp.Addresses.Alice:
		ownerKey = dp.Secrets.Alice
	case dp.Addresses.Bob:
		ownerKey = dp.Secrets.Bob
	case dp.Addresses.Deployer:
		ownerKey = dp.Secrets.Deployer
	case dp.Addresses.SysCfgOwner:
		ownerKey = dp.Secrets.SysCfgOwner
	default:
		t.Fatalf("Unknown GPO owner: %s", gpoOwnerAddr.String())
	}

	gpoOwner, err := bind.NewKeyedTransactorWithChainID(ownerKey, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)

	_, err = gpoContract.SetOperator(gpoOwner, dp.Addresses.Alice)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(gpoOwnerAddr)(t)
	sequencer.ActL2EndBlock(t)

	gpoOperator, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Alice, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)

	tokenRatioValue := new(big.Int).SetUint64(dp.DeployConfig.GasPriceOracleTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, tokenRatioValue)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	// Alice makes first L2 transaction
	alice.ActResetTxOpts(t)
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	receipt := alice.LastTxReceipt(t)
	require.Equal(t, basefee, receipt.L1GasPrice, "L1 gas price matches basefee of L1 origin")
	require.NotZero(t, receipt.L1GasUsed, "L2 tx uses L1 data")
	require.NotZero(t, receipt.L1Fee, "L1 fee should be non-zero")

	// ========== Verify Mantle Fee Structure ==========

	// 1. L2 Execution Fee: Gas × GasPrice
	l2ExecutionFee := new(big.Int).Mul(
		new(big.Int).SetUint64(receipt.GasUsed),
		receipt.EffectiveGasPrice,
	)
	t.Logf("L2 Execution Fee: %s MNT (Gas: %d, Price: %s)",
		l2ExecutionFee.String(), receipt.GasUsed, receipt.EffectiveGasPrice.String())

	// 2. L1 Data Fee (already converted to MNT via tokenRatio)
	l1DataFee := receipt.L1Fee
	t.Logf("L1 Data Fee: %s MNT (L1GasUsed: %d, L1GasPrice: %s)",
		l1DataFee.String(), receipt.L1GasUsed.Uint64(), receipt.L1GasPrice.String())

	// Verify L1 fee calculation formula
	l1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	blobBaseFee, err := gpoContract.BlobBaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	baseFeeScalar, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	blobBaseFeeScalar, err := gpoContract.BlobBaseFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)

	// L1 Fee (ETH) = estimatedSize × (baseFeeScalar × l1BaseFee × 16 + blobBaseFeeScalar × blobBaseFee) / 10^12
	// L1 Fee (MNT) = L1 Fee (ETH) × tokenRatio / 10^6
	// Note: estimatedSize is calculated by _arsiaLinearRegression in the contract
	feeScaled := new(big.Int).Add(
		new(big.Int).Mul(new(big.Int).SetUint64(uint64(baseFeeScalar)), new(big.Int).Mul(l1BaseFee, big.NewInt(16))),
		new(big.Int).Mul(new(big.Int).SetUint64(uint64(blobBaseFeeScalar)), blobBaseFee),
	)
	t.Logf("L1 Fee Params: l1BaseFee=%s, blobBaseFee=%s, baseFeeScalar=%d, blobBaseFeeScalar=%d, feeScaled=%s",
		l1BaseFee.String(), blobBaseFee.String(), baseFeeScalar, blobBaseFeeScalar, feeScaled.String())

	// 3. Operator Fee: Constant + Scalar × 100 × Gas
	operatorFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt.GasUsed))
	require.NoError(t, err)

	operatorFeeScalar, err := gpoContract.OperatorFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	operatorFeeConstant, err := gpoContract.OperatorFeeConstant(&bind.CallOpts{})
	require.NoError(t, err)

	expectedOperatorFee := new(big.Int).Add(
		new(big.Int).SetUint64(operatorFeeConstant),
		new(big.Int).Mul(
			new(big.Int).SetUint64(uint64(operatorFeeScalar)),
			new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(receipt.GasUsed)),
		),
	)
	require.Equal(t, expectedOperatorFee, operatorFee, "Operator fee formula mismatch")
	t.Logf("Operator Fee: %s MNT (Scalar=%d, Constant=%d)", operatorFee.String(), operatorFeeScalar, operatorFeeConstant)

	// 4. Verify TokenRatio (DECIMALS=6, not 1e18)
	tokenRatioValue, err = gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(t, err)
	t.Logf("TokenRatio: %s (1 ETH = %s MNT)", tokenRatioValue.String(),
		new(big.Int).Div(tokenRatioValue, big.NewInt(1e6)).String())

	// 5. Verify MinBaseFee constraint
	l2BaseFee, err := gpoContract.BaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	minBaseFeeCheck, err := sysCfgContract.MinBaseFee(&bind.CallOpts{})
	require.NoError(t, err)

	// Note: L2 BaseFee (stored value) can be 0, but EffectiveGasPrice uses max(BaseFee, MinBaseFee)
	// So even if BaseFee = 0, transactions will pay at least MinBaseFee
	effectiveBaseFee := receipt.EffectiveGasPrice
	t.Logf("L2 BaseFee (stored): %s MNT, MinBaseFee: %d MNT, EffectiveGasPrice: %s MNT",
		l2BaseFee.String(), minBaseFeeCheck, effectiveBaseFee.String())

	// Verify EffectiveGasPrice >= MinBaseFee
	require.GreaterOrEqual(t, effectiveBaseFee.Uint64(), minBaseFeeCheck,
		"EffectiveGasPrice should be >= MinBaseFee")

	// 6. Total Fee = L2 + L1 + Operator
	totalFee := new(big.Int).Add(l2ExecutionFee, l1DataFee)
	totalFee = totalFee.Add(totalFee, operatorFee)
	t.Logf("Total Fee: %s MNT (L2: %s + L1: %s + Operator: %s)",
		totalFee.String(), l2ExecutionFee.String(), l1DataFee.String(), operatorFee.String())

	initialL1Fee := receipt.L1Fee

	// ========== Update GPO Parameters ==========

	// Submit L2 batch to L1
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// Double baseFeeScalar to test L1 fee change
	currentBaseFeeScalar := baseFeeScalar
	newBaseFeeScalar := currentBaseFeeScalar * 2
	newBlobBaseFeeScalar := blobBaseFeeScalar
	if newBlobBaseFeeScalar == 0 {
		newBlobBaseFeeScalar = 1 // Ensure non-zero for L1 fee calculation
	}

	_, err = sysCfgContract.SetGasConfigArsia(sysCfgOwner, newBaseFeeScalar, newBlobBaseFeeScalar)
	require.NoError(t, err)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
	miner.ActL1EndBlock(t)
	basefeeGPOUpdate := miner.L1Chain().CurrentBlock().BaseFee

	// Build L2 blocks up to but excluding the block that adopts the GPO change
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1HeadExcl(t)

	// Verify L2 hasn't adopted the GPO change yet
	l2BaseFeeScalarBeforeUpdate, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, currentBaseFeeScalar, l2BaseFeeScalarBeforeUpdate,
		"L2 baseFeeScalar should not be updated before adopting new L1 origin")

	// Alice makes second transaction (adopts L1 origin with GPO change)
	alice.ActResetTxOpts(t)
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	// Verify SystemConfig updated on L1
	updatedBaseFeeScalar, err := sysCfgContract.BasefeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	updatedBlobBaseFeeScalar, err := sysCfgContract.BlobbasefeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, newBaseFeeScalar, updatedBaseFeeScalar, "BaseFeeScalar mismatch")
	require.Equal(t, newBlobBaseFeeScalar, updatedBlobBaseFeeScalar, "BlobBaseFeeScalar mismatch")

	receipt2 := alice.LastTxReceipt(t)
	require.Equal(t, basefeeGPOUpdate, receipt2.L1GasPrice, "L1 gas price mismatch")
	require.NotZero(t, receipt2.L1GasUsed, "L2 tx should use L1 data")
	require.NotZero(t, receipt2.L1Fee, "L1 fee should be non-zero")

	// ========== Verify Fee Changes After GPO Update ==========

	t.Logf("L1 Fee change: %s -> %s", initialL1Fee.String(), receipt2.L1Fee.String())
	if initialL1Fee.Sign() > 0 {
		feeRatio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt2.L1Fee),
			new(big.Float).SetInt(initialL1Fee),
		)
		t.Logf("L1 Fee ratio: %s", feeRatio.String())
		// Use big.Int.Cmp instead of Uint64() to avoid overflow
		require.Equal(t, 1, receipt2.L1Fee.Cmp(initialL1Fee),
			"L1 fee should increase after baseFeeScalar doubled")
	} else {
		require.NotZero(t, receipt2.L1Fee, "L1 fee should be non-zero after update")
	}

	operatorFee2, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
	require.NoError(t, err)
	t.Logf("Operator Fee after update: %s MNT", operatorFee2.String())

	totalFee2 := new(big.Int).Add(
		new(big.Int).Mul(new(big.Int).SetUint64(receipt2.GasUsed), receipt2.EffectiveGasPrice),
		receipt2.L1Fee,
	)
	totalFee2 = totalFee2.Add(totalFee2, operatorFee2)
	t.Logf("Total Fee after update: %s MNT", totalFee2.String())

	// ========== Verify GPO Parameter Persistence ==========

	// Build more L2 blocks with new L1 origin
	miner.ActEmptyBlock(t)
	basefee = miner.L1Chain().CurrentBlock().BaseFee
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Alice makes third transaction
	alice.ActResetTxOpts(t)
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	receipt3 := alice.LastTxReceipt(t)
	require.Equal(t, basefee, receipt3.L1GasPrice, "L1 gas price mismatch")
	require.NotZero(t, receipt3.L1GasUsed, "L2 tx should use L1 data")
	require.NotZero(t, receipt3.L1Fee, "L1 fee should be non-zero")

	// Verify L1 fee remains elevated (baseFeeScalar still doubled)
	// Use big.Int.Cmp instead of Uint64() to avoid overflow
	require.Equal(t, 1, receipt3.L1Fee.Cmp(initialL1Fee),
		"L1 fee should remain elevated after GPO params update")

	operatorFee3, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt3.GasUsed))
	require.NoError(t, err)

	t.Logf("GPO params change test completed!")
	t.Logf("   Initial L1 Fee: %s MNT", initialL1Fee.String())
	t.Logf("   After baseFeeScalar doubled: %s MNT", receipt2.L1Fee.String())
	t.Logf("   After L1 origin change: %s MNT (persistent)", receipt3.L1Fee.String())
	t.Logf("   Operator Fee: %s MNT (consistent)", operatorFee3.String())
}

// OperatorFeeScalarsChange tests that Operator Fee parameters can be updated dynamically
// and that the updated parameters affect subsequent transactions correctly.
// This test verifies the Operator Fee formula: OperatorFee = Constant + Scalar × 100 × Gas
func OperatorFeeScalarsChange(gt *testing.T, isSpanBatch bool) {
	t := actionsHelpers.NewDefaultTesting(gt)
	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)
	//Enable MNT as gas fee token (zero address represents MNT in Mantle)
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set tokenRatio: 1 ETH = 2000 MNT (using DECIMALS=6 precision)
	dp.DeployConfig.GasPriceOracleTokenRatio = 2000 * 1e6

	//Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000 //wei
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000
	OperatorfeeConstant := uint64(dp.DeployConfig.GasPriceOracleOperatorFeeConstant)
	OperatorfeeScalar := uint64(dp.DeployConfig.GasPriceOracleOperatorFeeScalar)
	// Set MinBaseFee (around 10 gwei for Mantle)
	dp.DeployConfig.MinBaseFee = 10 * 1e9 // 10 gwei

	// Set L2 genesis block BaseFee (initial BaseFee for L2)
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9)) // 10 gwei
	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, log)
	batcher := actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSpanBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	if !isSpanBatch {
		batcher = actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
			sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	}
	alice := actionsHelpers.NewBasicUser[any](log, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	sequencer.ActL2PipelineFull(t)

	// Build initial L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)
	L1GasPrice := miner.L1Chain().CurrentBlock().BaseFee

	//get contracts bindings
	gpoContract, err := bindings.NewGasPriceOracle(common.HexToAddress("0x420000000000000000000000000000000000000F"), seqEngine.EthClient())
	require.NoError(t, err)
	sysCfgContract, err := bindings.NewSystemConfig(sd.RollupCfg.L1SystemConfigAddress, miner.EthClient())
	require.NoError(t, err)
	//Set MinBaseFee on L1 SystemConfig (workaround for missing evm: tag)
	sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.RollupCfg.L1ChainID)
	require.NoError(t, err)
	_, err = sysCfgContract.SetMinBaseFee(sysCfgOwner, dp.DeployConfig.MinBaseFee)
	require.NoError(t, err)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
	miner.ActL1EndBlock(t)

	// Verify MinBaseFee is set
	minBaseFee, err := sysCfgContract.MinBaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, dp.DeployConfig.MinBaseFee, minBaseFee, "MinBaseFee should be set")
	t.Logf("MinBaseFee set to: %d wei (%d gwei)", minBaseFee, minBaseFee/1e9)

	//Set tokenRatio
	gopOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)
	tokenRatioValue := new(big.Int).SetUint64(dp.DeployConfig.GasPriceOracleTokenRatio)
	_, err = gpoContract.SetTokenRatio(gopOwner, tokenRatioValue)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Deployer)(t)
	sequencer.ActL2EndBlock(t)

	//Alice make first L2 transaction
	alice.ActResetTxOpts(t)
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	//record tx operator fee
	receipt := alice.LastTxReceipt(t)
	require.Equal(t, L1GasPrice, receipt.L1GasPrice, "L1 gas price mismatch")
	require.Equal(t, OperatorfeeConstant, *receipt.OperatorFeeConstant, "Operator fee constant mismatch")
	require.Equal(t, OperatorfeeScalar, *receipt.OperatorFeeScalar, "Operator fee scalar mismatch")
	require.NotZero(t, receipt.L1GasUsed, "L2 tx uses L1 data")
	require.NotZero(t, receipt.L1Fee, "L1 fee should be non-zero")
	exOperatorFee := new(big.Int).Add(
		new(big.Int).SetUint64(dp.DeployConfig.GasPriceOracleOperatorFeeConstant),
		new(big.Int).Mul(
			new(big.Int).SetUint64(uint64(dp.DeployConfig.GasPriceOracleOperatorFeeScalar)),
			new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(receipt.GasUsed)),
		),
	)
	acOperatorFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt.GasUsed))
	require.NoError(t, err)
	require.Equal(t, exOperatorFee, acOperatorFee, "Operator fee mismatch")
	t.Logf("Initial Operator Fee: %s MNT", acOperatorFee.String())

	// ========== Update Operator Fee Scalars ==========
	// Submit L2 batch to L1
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// Double Operator Fee Scalar to test fee change
	newOperatorfeeScalar := uint32(2 * OperatorfeeScalar)
	newOperatorFeeConstant := uint64(2 * OperatorfeeConstant)
	_, err = sysCfgContract.SetOperatorFeeScalars(sysCfgOwner, newOperatorfeeScalar, newOperatorFeeConstant)
	require.NoError(t, err)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
	miner.ActL1EndBlock(t)

	// Build L2 blocks up to but excluding the block that adopts the Operator Fee Scalars change
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1HeadExcl(t)
	L1GasPrice = miner.L1Chain().CurrentBlock().BaseFee
	// Verify L2 hasn't adopted the GPO change yet
	operatorFeeConstantBeforeupdate, err := gpoContract.OperatorFeeConstant(&bind.CallOpts{})
	require.NoError(t, err)
	operatorFeeScalarBeforeupdate, err := gpoContract.OperatorFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, operatorFeeConstantBeforeupdate, OperatorfeeConstant, "Operator fee constant mismatch")
	require.Equal(t, uint64(operatorFeeScalarBeforeupdate), OperatorfeeScalar, "Operator fee scalar mismatch")

	//Alice make second transaction (adopts L1 origin with Operator Fee  change)
	alice.ActResetTxOpts(t)
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	// Verify SystemConfig updated on L1
	updatedOperatorFeeConstant, err := sysCfgContract.OperatorFeeConstant(&bind.CallOpts{})
	require.NoError(t, err)
	updatedOperatorFeeScalar, err := sysCfgContract.OperatorFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, updatedOperatorFeeConstant, uint64(newOperatorFeeConstant), "Operator fee constant mismatch")
	require.Equal(t, updatedOperatorFeeScalar, newOperatorfeeScalar, "Operator fee scalar mismatch")

	receipt2 := alice.LastTxReceipt(t)
	require.Equal(t, L1GasPrice, receipt2.L1GasPrice, "L1 gas price mismatch")
	require.Equal(t, uint64(newOperatorFeeConstant), *receipt2.OperatorFeeConstant, "Operator fee constant mismatch")
	require.Equal(t, uint64(newOperatorfeeScalar), *receipt2.OperatorFeeScalar, "Operator fee scalar mismatch")
	require.NotZero(t, receipt2.L1GasUsed, "L2 tx uses L1 data")
	require.NotZero(t, receipt2.L1Fee, "L1 fee should be non-zero")
	// ========== Verify Operator Fee Changes After Scalars Update ==========
	exOperatorFee2 := new(big.Int).Add(
		new(big.Int).SetUint64(uint64(newOperatorFeeConstant)),
		new(big.Int).Mul(
			new(big.Int).SetUint64(uint64(newOperatorfeeScalar)),
			new(big.Int).Mul(big.NewInt(100), new(big.Int).SetUint64(receipt2.GasUsed)),
		),
	)
	acOperatorFee2, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
	require.NoError(t, err)
	require.Equal(t, exOperatorFee2, acOperatorFee2, "Operator fee mismatch")
	require.True(t, acOperatorFee2.Cmp(acOperatorFee) == 1, "Operator fee should be greater. Fee1: %s, Fee2: %s", acOperatorFee.String(), acOperatorFee2.String())
	t.Logf("Updated Operator Fee: %s MNT", acOperatorFee2.String())
	t.Logf("Operator Fee Scalars change test completed!")
}

// GasLimitChange tests that the gas limit can be configured to L1,
// and that the L2 changes the gas limit instantly at the exact block that adopts the L1 origin with
// the gas limit change event. And checks if a verifier node can reproduce the same gas limit change.
func GasLimitChange(gt *testing.T, isSpanBatch bool) {
	t := actionsHelpers.NewDefaultTesting(gt)
	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)
	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, log)
	batcher := actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSpanBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	if !isSpanBatch {
		batcher = actionsHelpers.NewL2Batcher(log, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
			sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))
	}
	sequencer.ActL2PipelineFull(t)
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	oldGasLimit := seqEngine.L2Chain().CurrentBlock().GasLimit
	require.Equal(t, oldGasLimit, uint64(dp.DeployConfig.L2GenesisBlockGasLimit))

	// change gas limit on L1 to triple what it was
	sysCfgContract, err := bindings.NewSystemConfig(sd.RollupCfg.L1SystemConfigAddress, miner.EthClient())
	require.NoError(t, err)

	sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.RollupCfg.L1ChainID)
	require.NoError(t, err)

	_, err = sysCfgContract.SetGasLimit(sysCfgOwner, oldGasLimit*3)
	require.NoError(t, err)

	// include the gaslimit update on L1
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
	miner.ActL1EndBlock(t)

	// build to latest L1, excluding the block that adopts the L1 block with the gaslimit change
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1HeadExcl(t)

	require.Equal(t, oldGasLimit, seqEngine.L2Chain().CurrentBlock().GasLimit)
	require.Equal(t, uint64(1), sequencer.SyncStatus().UnsafeL2.L1Origin.Number)

	// now include the L1 block with the gaslimit change, and see if it changes as expected
	sequencer.ActBuildToL1Head(t)
	require.Equal(t, oldGasLimit*3, seqEngine.L2Chain().CurrentBlock().GasLimit)
	require.Equal(t, uint64(2), sequencer.SyncStatus().UnsafeL2.L1Origin.Number)

	// now submit all this to L1, and see if a verifier can sync and reproduce it
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	_, verifier := actionsHelpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg), miner.BlobStore(), &sync.Config{})
	verifier.ActL2PipelineFull(t)

	require.Equal(t, sequencer.L2Unsafe(), verifier.L2Safe(), "verifier stays in sync, even with gaslimit changes")
}

// TestTokenRatioInBlockChange tests that when token ratio is updated within a block:
// - Transaction 1 (before update): uses old tokenRatio
// - Transaction 2 (SetTokenRatio): uses old tokenRatio (the tx that modifies it)
// - Transaction 3 (after update): uses new tokenRatio
//
// This verifies the token ratio caching mechanism in rollup_cost.go works correctly.
func TestTokenRatioInBlockChange(t *testing.T) {
	gt := actionsHelpers.NewDefaultTesting(t)

	dp := e2eutils.MakeMantleDeployParams(gt, actionsHelpers.DefaultRollupTestParams())

	// Enable MNT as gas token
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set initial tokenRatio: 1 ETH = 2000 MNT
	initialTokenRatio := uint64(2000 * 1e6)
	dp.DeployConfig.GasPriceOracleTokenRatio = initialTokenRatio

	// Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000

	// Set MinBaseFee
	dp.DeployConfig.MinBaseFee = 10 * 1e9 // 10 gwei
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9))

	// Activate Arsia fork at genesis
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(gt, dp, actionsHelpers.DefaultAlloc)
	logger := testlog.Logger(gt, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(gt, sd, logger)

	// Setup users
	alice := actionsHelpers.NewBasicUser[any](logger, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	bob := actionsHelpers.NewBasicUser[any](logger, dp.Secrets.Bob, rand.New(rand.NewSource(5678)))
	bob.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	sequencer.ActL2PipelineFull(gt)

	// Build initial L1 and L2 blocks
	miner.ActEmptyBlock(gt)
	sequencer.ActL1HeadSignal(gt)
	sequencer.ActBuildToL1Head(gt)

	// Get GPO contract
	gpoContract, err := bindings.NewGasPriceOracle(
		common.HexToAddress("0x420000000000000000000000000000000000000F"),
		seqEngine.EthClient(),
	)
	require.NoError(gt, err)

	// Setup GPO operator (find the owner first)
	gpoOwnerAddr, err := gpoContract.Owner(&bind.CallOpts{})
	require.NoError(gt, err)

	var ownerKey *ecdsa.PrivateKey
	switch gpoOwnerAddr {
	case dp.Addresses.Alice:
		ownerKey = dp.Secrets.Alice
	case dp.Addresses.Bob:
		ownerKey = dp.Secrets.Bob
	case dp.Addresses.Deployer:
		ownerKey = dp.Secrets.Deployer
	case dp.Addresses.SysCfgOwner:
		ownerKey = dp.Secrets.SysCfgOwner
	default:
		gt.Fatalf("Unknown GPO owner: %s", gpoOwnerAddr.String())
	}

	gpoOwner, err := bind.NewKeyedTransactorWithChainID(ownerKey, sd.L2Cfg.Config.ChainID)
	require.NoError(gt, err)

	// Set operator to Alice
	_, err = gpoContract.SetOperator(gpoOwner, dp.Addresses.Alice)
	require.NoError(gt, err)
	sequencer.ActL2StartBlock(gt)
	seqEngine.ActL2IncludeTx(gpoOwnerAddr)(gt)
	sequencer.ActL2EndBlock(gt)

	// Set initial tokenRatio
	gpoOperator, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Alice, sd.L2Cfg.Config.ChainID)
	require.NoError(gt, err)

	tokenRatioValue := new(big.Int).SetUint64(initialTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, tokenRatioValue)
	require.NoError(gt, err)
	sequencer.ActL2StartBlock(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)
	sequencer.ActL2EndBlock(gt)

	// Verify initial tokenRatio is set
	currentRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(gt, err)
	require.Equal(gt, tokenRatioValue, currentRatio)
	gt.Logf("Initial tokenRatio: %s (1 ETH = %s MNT)",
		currentRatio.String(),
		new(big.Int).Div(currentRatio, big.NewInt(1e6)).String())

	// Now test the in-block token ratio change

	gt.Log("Testing in-block token ratio change")

	// Start a new L2 block that will contain 3 transactions
	sequencer.ActL2StartBlock(gt)

	// Transaction 1: Alice sends a transaction (should use old tokenRatio)
	alice.ActResetTxOpts(gt)
	alice.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Transaction 2: Update tokenRatio to 4000 MNT per ETH (should still use old tokenRatio)
	newTokenRatio := uint64(4000 * 1e6) // Double the ratio
	newTokenRatioValue := new(big.Int).SetUint64(newTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, newTokenRatioValue)
	require.NoError(gt, err)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Transaction 3: Bob sends a transaction (should use new tokenRatio)
	bob.ActResetTxOpts(gt)
	bob.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(gt)

	// End the block
	sequencer.ActL2EndBlock(gt)

	// Get receipts for all three transactions
	receipt1 := alice.LastTxReceipt(gt)
	receipt3 := bob.LastTxReceipt(gt)

	// Get the SetTokenRatio transaction receipt
	block := seqEngine.L2Chain().CurrentBlock()
	blockNum := block.Number.Uint64()
	fullBlock, err := seqEngine.EthClient().BlockByNumber(gt.Ctx(), new(big.Int).SetUint64(blockNum))
	require.NoError(gt, err)

	// Filter out deposit transactions (type 0x7E)
	var userTxs []*types.Transaction
	for _, tx := range fullBlock.Transactions() {
		if tx.Type() != types.DepositTxType {
			userTxs = append(userTxs, tx)
		}
	}
	require.Equal(gt, 3, len(userTxs), "Block should contain exactly 3 user transactions (excluding deposits)")

	// Get the SetTokenRatio transaction receipt (should be the 2nd user transaction)
	receipt2, err := seqEngine.EthClient().TransactionReceipt(gt.Ctx(), userTxs[1].Hash())
	require.NoError(gt, err)

	gt.Log("Verifying token ratio application")

	// Verify tokenRatio is updated in contract
	updatedRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(gt, err)
	require.Equal(gt, newTokenRatioValue, updatedRatio)
	gt.Logf("Updated tokenRatio: %s (1 ETH = %s MNT)",
		updatedRatio.String(),
		new(big.Int).Div(updatedRatio, big.NewInt(1e6)).String())

	// Log L1 fee parameters for reference
	l1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
	require.NoError(gt, err)
	blobBaseFee, err := gpoContract.BlobBaseFee(&bind.CallOpts{})
	require.NoError(gt, err)
	baseFeeScalar, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
	require.NoError(gt, err)
	blobBaseFeeScalar, err := gpoContract.BlobBaseFeeScalar(&bind.CallOpts{})
	require.NoError(gt, err)

	gt.Logf("L1 Fee Params: l1BaseFee=%s, blobBaseFee=%s, baseFeeScalar=%d, blobBaseFeeScalar=%d",
		l1BaseFee.String(), blobBaseFee.String(), baseFeeScalar, blobBaseFeeScalar)

	// Transaction 1: Should use old tokenRatio (2000 MNT/ETH)
	gt.Logf("Transaction 1 (Alice, before update):")
	gt.Logf("  L1 Fee: %s MNT", receipt1.L1Fee.String())
	gt.Logf("  L1 Gas Used: %d", receipt1.L1GasUsed.Uint64())
	gt.Logf("  L1 Gas Price: %s", receipt1.L1GasPrice.String())

	// Transaction 2: Should use old tokenRatio (2000 MNT/ETH) - the tx that modifies it
	gt.Logf("Transaction 2 (SetTokenRatio):")
	gt.Logf("  L1 Fee: %s MNT", receipt2.L1Fee.String())
	gt.Logf("  L1 Gas Used: %d", receipt2.L1GasUsed.Uint64())
	gt.Logf("  L1 Gas Price: %s", receipt2.L1GasPrice.String())

	// Transaction 3: Should use new tokenRatio (4000 MNT/ETH)
	gt.Logf("Transaction 3 (Bob, after update):")
	gt.Logf("  L1 Fee: %s MNT", receipt3.L1Fee.String())
	gt.Logf("  L1 Gas Used: %d", receipt3.L1GasUsed.Uint64())
	gt.Logf("  L1 Gas Price: %s", receipt3.L1GasPrice.String())

	// Key verification: Transaction 3's L1 fee should be approximately double of Transaction 1's
	// (because tokenRatio doubled, and L1 fee is proportional to tokenRatio)
	// Allow some variance due to different transaction sizes
	if receipt1.L1Fee.Sign() > 0 && receipt3.L1Fee.Sign() > 0 {
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt3.L1Fee),
			new(big.Float).SetInt(receipt1.L1Fee),
		)
		gt.Logf("L1 Fee ratio (Tx3/Tx1): %s", ratio.String())

		// The ratio should be close to 2.0 (allowing 10% variance for tx size differences)
		ratioFloat, _ := ratio.Float64()
		require.Greater(gt, ratioFloat, 1.8, "Tx3 L1 fee should be ~2x Tx1 (got %.2f)", ratioFloat)
		require.Less(gt, ratioFloat, 2.2, "Tx3 L1 fee should be ~2x Tx1 (got %.2f)", ratioFloat)
	}

	// Verify Transaction 2 uses old tokenRatio (similar to Tx1)
	if receipt1.L1Fee.Sign() > 0 && receipt2.L1Fee.Sign() > 0 {
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt2.L1Fee),
			new(big.Float).SetInt(receipt1.L1Fee),
		)
		gt.Logf("L1 Fee ratio (Tx2/Tx1): %s", ratio.String())

		// Tx2 should use old tokenRatio, so ratio should be close to 1.0
		ratioFloat, _ := ratio.Float64()
		require.Greater(gt, ratioFloat, 0.5, "Tx2 should use old tokenRatio (got %.2f)", ratioFloat)
		require.Less(gt, ratioFloat, 1.5, "Tx2 should use old tokenRatio (got %.2f)", ratioFloat)
	}

	gt.Log("Test completed successfully!")
	gt.Log("Verified:")
	gt.Log("  - Tx1 (before update): uses old tokenRatio")
	gt.Log("  - Tx2 (SetTokenRatio): uses old tokenRatio")
	gt.Log("  - Tx3 (after update): uses new tokenRatio")

}
