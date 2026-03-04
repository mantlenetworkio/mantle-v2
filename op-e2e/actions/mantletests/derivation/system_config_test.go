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
	"github.com/ethereum/go-ethereum/params"
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
	t.Logf("  Initial L1 Fee: %s MNT", initialL1Fee.String())
	t.Logf("  After baseFeeScalar doubled: %s MNT", receipt2.L1Fee.String())
	t.Logf("  After L1 origin change: %s MNT (persistent)", receipt3.L1Fee.String())
	t.Logf("  Operator Fee: %s MNT (consistent)", operatorFee3.String())
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

	// Record balances BEFORE the block containing 4 transactions
	aliceBalanceBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(gt, err)
	bobBalanceBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(gt, err)

	// Record vault balances before (for verification)
	baseFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000019")
	seqFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000011")
	l1FeeVault := common.HexToAddress("0x420000000000000000000000000000000000001a")
	opFeeVault := common.HexToAddress("0x420000000000000000000000000000000000001b")

	baseFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), baseFeeVault, nil)
	require.NoError(gt, err)
	seqFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), seqFeeVault, nil)
	require.NoError(gt, err)
	l1FeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), l1FeeVault, nil)
	require.NoError(gt, err)
	opFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), opFeeVault, nil)
	require.NoError(gt, err)

	// Start a new L2 block that will contain 4 transactions
	sequencer.ActL2StartBlock(gt)

	// Transaction 1: Alice sends a transaction (should use old tokenRatio)
	alice.ActResetTxOpts(gt)
	alice.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Transaction 2: Update tokenRatio to 4000 MNT per ETH (should still use old tokenRatio)
	newTokenRatio := uint64(4000 * 1e6) // Double the ratio (2000 -> 4000)
	newTokenRatioValue := new(big.Int).SetUint64(newTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, newTokenRatioValue)
	require.NoError(gt, err)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Transaction 3: Bob sends a transaction (should use new tokenRatio)
	bob.ActResetTxOpts(gt)
	bob.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(gt)

	// Transaction 4: Alice sends another transaction (should also use new tokenRatio)
	alice.ActResetTxOpts(gt)
	alice.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// End the block
	sequencer.ActL2EndBlock(gt)

	// Get receipts from block transactions (more reliable than LastTxReceipt for multiple txs)
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
	require.Equal(gt, 4, len(userTxs), "Block should contain exactly 4 user transactions (excluding deposits)")

	// Get receipts for all four transactions
	receipt1, err := seqEngine.EthClient().TransactionReceipt(gt.Ctx(), userTxs[0].Hash())
	require.NoError(gt, err)
	receipt2, err := seqEngine.EthClient().TransactionReceipt(gt.Ctx(), userTxs[1].Hash())
	require.NoError(gt, err)
	receipt3, err := seqEngine.EthClient().TransactionReceipt(gt.Ctx(), userTxs[2].Hash())
	require.NoError(gt, err)
	receipt4, err := seqEngine.EthClient().TransactionReceipt(gt.Ctx(), userTxs[3].Hash())
	require.NoError(gt, err)

	// Record vault balances after
	baseFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), baseFeeVault, nil)
	require.NoError(gt, err)
	seqFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), seqFeeVault, nil)
	require.NoError(gt, err)
	l1FeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), l1FeeVault, nil)
	require.NoError(gt, err)
	opFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), opFeeVault, nil)
	require.NoError(gt, err)

	// Verify tokenRatio is updated in contract
	updatedRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(gt, err)
	require.Equal(gt, newTokenRatioValue, updatedRatio)

	// Key verification: Transaction 3's L1 fee should be approximately double of Transaction 1's
	// (because tokenRatio doubled, and L1 fee is proportional to tokenRatio)
	// Allow some variance due to different transaction sizes
	if receipt1.L1Fee.Sign() > 0 && receipt3.L1Fee.Sign() > 0 {
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt3.L1Fee),
			new(big.Float).SetInt(receipt1.L1Fee),
		)
		gt.Logf("L1 Fee ratio (Tx3/Tx1): %s", ratio.String())

		ratioFloat, _ := ratio.Float64()
		require.Greater(gt, ratioFloat, 1.8, "Tx3 L1 fee should be ~2x Tx1 (got %.2f)", ratioFloat)
		require.Less(gt, ratioFloat, 2.2, "Tx3 L1 fee should be ~2x Tx1 (got %.2f)", ratioFloat)
	}

	// Verify Transaction 4 also uses new tokenRatio (~2x of Tx1)
	if receipt1.L1Fee.Sign() > 0 && receipt4.L1Fee.Sign() > 0 {
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt4.L1Fee),
			new(big.Float).SetInt(receipt1.L1Fee),
		)
		gt.Logf("L1 Fee ratio (Tx4/Tx1): %s", ratio.String())

		ratioFloat, _ := ratio.Float64()
		require.Greater(gt, ratioFloat, 1.8, "Tx4 L1 fee should be ~2x Tx1 (got %.2f)", ratioFloat)
		require.Less(gt, ratioFloat, 2.2, "Tx4 L1 fee should be ~2x Tx1 (got %.2f)", ratioFloat)
	}

	// Verify Transaction 2 uses old tokenRatio (similar to Tx1)
	if receipt1.L1Fee.Sign() > 0 && receipt2.L1Fee.Sign() > 0 {
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt2.L1Fee),
			new(big.Float).SetInt(receipt1.L1Fee),
		)
		gt.Logf("L1 Fee ratio (Tx2/Tx1): %s", ratio.String())

		ratioFloat, _ := ratio.Float64()
		require.Greater(gt, ratioFloat, 0.5, "Tx2 should use old tokenRatio (got %.2f)", ratioFloat)
		require.Less(gt, ratioFloat, 1.5, "Tx2 should use old tokenRatio (got %.2f)", ratioFloat)
	}

	// ========== CRITICAL: Verify actual balance deductions match receipt fees ==========
	// This catches the bug where L1CostFunc uses cached (old) tokenRatio for actual deduction
	// while receipt shows the correct (new) tokenRatio.

	gt.Log("Verifying actual balance deductions match receipt fees...")

	// Record balances AFTER the block
	aliceBalanceAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(gt, err)
	bobBalanceAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(gt, err)
	gt.Logf("Balance after block: Alice=%s, Bob=%s", aliceBalanceAfter.String(), bobBalanceAfter.String())

	// --- Verify Bob's deduction (Tx3 only, simplest case) ---
	// Bob only sent Tx3, so his balance change = Tx3 total fee
	// Total fee = L2 execution fee + L1 data fee + Operator fee
	// L2 execution fee = gasUsed * effectiveGasPrice
	bobActualDeduction := new(big.Int).Sub(bobBalanceBefore, bobBalanceAfter)
	// Subtract tx value (default is 0 for ActResetTxOpts)
	bobTxValue := big.NewInt(0) // ActResetTxOpts sets Value to 0
	bobActualFee := new(big.Int).Sub(bobActualDeduction, bobTxValue)

	bobL2Fee := new(big.Int).Mul(
		new(big.Int).SetUint64(receipt3.GasUsed),
		receipt3.EffectiveGasPrice,
	)
	bobOperatorFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt3.GasUsed))
	require.NoError(gt, err)

	// Expected fee based on receipt (uses new tokenRatio in receipt.L1Fee)
	bobExpectedFee := new(big.Int).Add(bobL2Fee, receipt3.L1Fee)
	bobExpectedFee.Add(bobExpectedFee, bobOperatorFee)

	// Derive actual L1 fee from balance: actualL1Fee = actualFee - L2Fee - operatorFee
	bobActualL1Fee := new(big.Int).Sub(bobActualFee, bobL2Fee)
	bobActualL1Fee.Sub(bobActualL1Fee, bobOperatorFee)

	gt.Logf("Bob (Tx3) verification:")
	gt.Logf("  Receipt L1Fee: %s", receipt3.L1Fee.String())
	gt.Logf("  Actual L1Fee (from balance): %s", bobActualL1Fee.String())
	gt.Logf("  Expected total: %s, Actual total: %s", bobExpectedFee.String(), bobActualFee.String())

	// Check if Tx3 actual deduction matches receipt
	tx3Match := bobExpectedFee.Cmp(bobActualFee) == 0
	if !tx3Match {
		gt.Logf("  MISMATCH!")
	} else {
		gt.Logf("  Match!")
	}

	// --- Verify Alice's deduction (Tx1 + Tx2 + Tx4) ---
	aliceActualDeduction := new(big.Int).Sub(aliceBalanceBefore, aliceBalanceAfter)
	aliceActualFee := new(big.Int).Set(aliceActualDeduction) // tx values are 0

	// Tx1 expected fee
	tx1L2Fee := new(big.Int).Mul(
		new(big.Int).SetUint64(receipt1.GasUsed),
		receipt1.EffectiveGasPrice,
	)
	tx1OperatorFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt1.GasUsed))
	require.NoError(gt, err)
	tx1ExpectedFee := new(big.Int).Add(tx1L2Fee, receipt1.L1Fee)
	tx1ExpectedFee.Add(tx1ExpectedFee, tx1OperatorFee)

	// Tx2 expected fee
	tx2L2Fee := new(big.Int).Mul(
		new(big.Int).SetUint64(receipt2.GasUsed),
		receipt2.EffectiveGasPrice,
	)
	tx2OperatorFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
	require.NoError(gt, err)
	tx2ExpectedFee := new(big.Int).Add(tx2L2Fee, receipt2.L1Fee)
	tx2ExpectedFee.Add(tx2ExpectedFee, tx2OperatorFee)

	// Tx4 expected fee
	tx4L2Fee := new(big.Int).Mul(
		new(big.Int).SetUint64(receipt4.GasUsed),
		receipt4.EffectiveGasPrice,
	)
	tx4OperatorFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt4.GasUsed))
	require.NoError(gt, err)
	tx4ExpectedFee := new(big.Int).Add(tx4L2Fee, receipt4.L1Fee)
	tx4ExpectedFee.Add(tx4ExpectedFee, tx4OperatorFee)

	aliceExpectedFee := new(big.Int).Add(tx1ExpectedFee, tx2ExpectedFee)
	aliceExpectedFee.Add(aliceExpectedFee, tx4ExpectedFee)

	// Derive Alice's actual L1 fees from balance diff
	aliceActualL1Fees := new(big.Int).Sub(aliceActualFee, tx1L2Fee)
	aliceActualL1Fees.Sub(aliceActualL1Fees, tx1OperatorFee)
	aliceActualL1Fees.Sub(aliceActualL1Fees, tx2L2Fee)
	aliceActualL1Fees.Sub(aliceActualL1Fees, tx2OperatorFee)
	aliceActualL1Fees.Sub(aliceActualL1Fees, tx4L2Fee)
	aliceActualL1Fees.Sub(aliceActualL1Fees, tx4OperatorFee)
	aliceReceiptL1Fees := new(big.Int).Add(receipt1.L1Fee, receipt2.L1Fee)
	aliceReceiptL1Fees.Add(aliceReceiptL1Fees, receipt4.L1Fee)

	gt.Logf("Alice (Tx1+Tx2+Tx4) verification:")
	gt.Logf("  Receipt L1Fees total: %s", aliceReceiptL1Fees.String())
	gt.Logf("  Actual L1Fees (from balance): %s", aliceActualL1Fees.String())
	gt.Logf("  Expected total: %s, Actual total: %s", aliceExpectedFee.String(), aliceActualFee.String())

	aliceMatch := aliceExpectedFee.Cmp(aliceActualFee) == 0
	if !aliceMatch {
		gt.Logf("  MISMATCH!")
	} else {
		gt.Logf("  Match!")
	}

	require.True(gt, tx3Match,
		"Tx3 (Bob): actual balance deduction must match receipt fees. "+
			"Actual L1 fee=%s, Receipt L1 fee=%s. "+
			"This means the actual L1 cost deduction used wrong tokenRatio.",
		bobActualL1Fee.String(), receipt3.L1Fee.String())

	require.True(gt, aliceMatch,
		"Alice (Tx1+Tx2+Tx4): actual balance deduction must match receipt fees. "+
			"Actual=%s, Expected=%s, ActualL1Fees=%s, ReceiptL1Fees=%s",
		aliceActualFee.String(), aliceExpectedFee.String(),
		aliceActualL1Fees.String(), aliceReceiptL1Fees.String())

	// ========== Verify Vault Balances (Arsia Mode) ==========
	// Calculate vault increases (independent data source from blockchain)
	baseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultAfter, baseFeeVaultBefore)
	seqFeeVaultIncrease := new(big.Int).Sub(seqFeeVaultAfter, seqFeeVaultBefore)
	l1FeeVaultIncrease := new(big.Int).Sub(l1FeeVaultAfter, l1FeeVaultBefore)
	opFeeVaultIncrease := new(big.Int).Sub(opFeeVaultAfter, opFeeVaultBefore)

	gt.Logf("Vault verification:")

	// In Arsia: L1FeeVault should receive the sum of all L1Fees
	totalL1Fee := new(big.Int).Add(receipt1.L1Fee, receipt2.L1Fee)
	totalL1Fee.Add(totalL1Fee, receipt3.L1Fee)
	totalL1Fee.Add(totalL1Fee, receipt4.L1Fee)

	gt.Logf("  L1FeeVault: expected=%s, actual=%s", totalL1Fee.String(), l1FeeVaultIncrease.String())
	require.Equal(gt, totalL1Fee.String(), l1FeeVaultIncrease.String(),
		"[Arsia] L1FeeVault should receive exactly the sum of all L1Fees")

	// Verify OperatorFeeVault received the sum of all OpFees
	totalOpFee := new(big.Int).Add(tx1OperatorFee, tx2OperatorFee)
	totalOpFee.Add(totalOpFee, bobOperatorFee) // tx3
	totalOpFee.Add(totalOpFee, tx4OperatorFee)

	gt.Logf("  OperatorFeeVault: expected=%s, actual=%s", totalOpFee.String(), opFeeVaultIncrease.String())
	require.Equal(gt, totalOpFee.String(), opFeeVaultIncrease.String(),
		"[Arsia] OperatorFeeVault should receive exactly the sum of all OpFees")

	// CRITICAL VERIFICATION: Total vault increase (from blockchain) == Total user deduction (from blockchain)
	// This uses two independent data sources to prevent false positives
	totalUserDeduction := new(big.Int).Add(aliceActualDeduction, bobActualDeduction)
	totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
	totalVaultIncrease.Add(totalVaultIncrease, l1FeeVaultIncrease)
	totalVaultIncrease.Add(totalVaultIncrease, opFeeVaultIncrease)

	gt.Logf("  Total UserDeduction (blockchain): %s", totalUserDeduction.String())
	gt.Logf("  Total VaultIncrease (blockchain): %s", totalVaultIncrease.String())

	require.Equal(gt, totalUserDeduction.String(), totalVaultIncrease.String(),
		"[Arsia] Total vault increase should equal total user deduction")

	gt.Log(" TestTokenRatioInBlockChange passed!")

}

// TestMultipleTokenRatioChangesInBlock tests that when token ratio is updated multiple times
// within a single block, each change is correctly applied to subsequent transactions.
//
// Block layout (6 transactions):
// - Tx1 (Alice): normal tx, uses initial tokenRatio (1000*1e6)
// - Tx2 (Alice): SetTokenRatio to 3000*1e6
// - Tx3 (Bob): normal tx, should use 3000*1e6
// - Tx4 (Alice): SetTokenRatio to 6000*1e6
// - Tx5 (Mallory): normal tx, should use 6000*1e6
// - Tx6 (Bob): normal tx, should use 6000*1e6
//
// This test exposes the caching bug in NewL1CostFuncArsia where Tx3 and Tx5
// (first non-modifier tx after each ratio change) use the old cached tokenRatio.
func TestMultipleTokenRatioChangesInBlock(t *testing.T) {
	gt := actionsHelpers.NewDefaultTesting(t)

	dp := e2eutils.MakeMantleDeployParams(gt, actionsHelpers.DefaultRollupTestParams())

	// Enable MNT as gas token
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set initial tokenRatio: 1 ETH = 1000 MNT
	initialTokenRatio := uint64(1000 * 1e6)
	dp.DeployConfig.GasPriceOracleTokenRatio = initialTokenRatio

	// Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000

	// Set MinBaseFee
	dp.DeployConfig.MinBaseFee = 10 * 1e9
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9))

	// Activate Arsia fork at genesis
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(gt, dp, actionsHelpers.DefaultAlloc)
	logger := testlog.Logger(gt, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(gt, sd, logger)

	// Setup users: Alice, Bob, Mallory
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

	mallory := actionsHelpers.NewBasicUser[any](logger, dp.Secrets.Mallory, rand.New(rand.NewSource(9012)))
	mallory.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
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

	// Setup GPO operator
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

	// Verify initial tokenRatio
	currentRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(gt, err)
	require.Equal(gt, tokenRatioValue, currentRatio)
	gt.Logf("Initial tokenRatio: %s (1 ETH = %s MNT)",
		currentRatio.String(),
		new(big.Int).Div(currentRatio, big.NewInt(1e6)).String())

	// Record balances BEFORE the block
	aliceBalanceBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(gt, err)
	bobBalanceBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(gt, err)
	malloryBalanceBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Mallory, nil)
	require.NoError(gt, err)

	// Record vault balances before (for verification)
	baseFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000019")
	seqFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000011")
	l1FeeVault := common.HexToAddress("0x420000000000000000000000000000000000001a")
	opFeeVault := common.HexToAddress("0x420000000000000000000000000000000000001b")

	baseFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), baseFeeVault, nil)
	require.NoError(gt, err)
	seqFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), seqFeeVault, nil)
	require.NoError(gt, err)
	l1FeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), l1FeeVault, nil)
	require.NoError(gt, err)
	opFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), opFeeVault, nil)
	require.NoError(gt, err)

	// ========== Build block with 6 transactions ==========
	sequencer.ActL2StartBlock(gt)

	// Tx1: Alice normal tx (uses initial ratio 1000)
	alice.ActResetTxOpts(gt)
	alice.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Tx2: SetTokenRatio to 3000 (uses initial ratio 1000)
	ratio2 := new(big.Int).SetUint64(uint64(3000 * 1e6))
	_, err = gpoContract.SetTokenRatio(gpoOperator, ratio2)
	require.NoError(gt, err)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Tx3: Bob normal tx (should use new ratio 3000)
	bob.ActResetTxOpts(gt)
	bob.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(gt)

	// Tx4: SetTokenRatio to 6000 (should use ratio 3000)
	ratio4 := new(big.Int).SetUint64(uint64(6000 * 1e6))
	_, err = gpoContract.SetTokenRatio(gpoOperator, ratio4)
	require.NoError(gt, err)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(gt)

	// Tx5: Mallory normal tx (should use new ratio 6000)
	mallory.ActResetTxOpts(gt)
	mallory.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Mallory)(gt)

	// Tx6: Bob normal tx (should also use ratio 6000)
	bob.ActResetTxOpts(gt)
	bob.ActMakeTx(gt)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(gt)

	sequencer.ActL2EndBlock(gt)

	// Get receipts from block
	block := seqEngine.L2Chain().CurrentBlock()
	blockNum := block.Number.Uint64()
	fullBlock, err := seqEngine.EthClient().BlockByNumber(gt.Ctx(), new(big.Int).SetUint64(blockNum))
	require.NoError(gt, err)

	var userTxs []*types.Transaction
	for _, tx := range fullBlock.Transactions() {
		if tx.Type() != types.DepositTxType {
			userTxs = append(userTxs, tx)
		}
	}
	require.Equal(gt, 6, len(userTxs), "Block should contain exactly 6 user transactions")

	receipts := make([]*types.Receipt, 6)
	for i := 0; i < 6; i++ {
		receipts[i], err = seqEngine.EthClient().TransactionReceipt(gt.Ctx(), userTxs[i].Hash())
		require.NoError(gt, err)
	}

	// ========== Log receipt data ==========
	txLabels := []string{
		"Tx1 (Alice, normal, ratio=1000)",
		"Tx2 (Alice, SetRatio→3000)",
		"Tx3 (Bob, normal, expect ratio=3000)",
		"Tx4 (Alice, SetRatio→6000)",
		"Tx5 (Mallory, normal, expect ratio=6000)",
		"Tx6 (Bob, normal, expect ratio=6000)",
	}
	for i, r := range receipts {
		gt.Logf("%s: L1Fee=%s, GasUsed=%d", txLabels[i], r.L1Fee.String(), r.GasUsed)
	}

	// Verify receipt L1Fee ratios relative to Tx1
	// Tx1 uses ratio 1000, Tx2 uses 1000, Tx3 should use 3000 (3x), Tx4 uses 3000 (3x),
	// Tx5 should use 6000 (6x), Tx6 should use 6000 (6x)
	expectedMultipliers := []float64{1.0, 1.0, 3.0, 3.0, 6.0, 6.0}
	baseFee := receipts[0].L1Fee
	for i := 1; i < 6; i++ {
		if baseFee.Sign() > 0 && receipts[i].L1Fee.Sign() > 0 {
			ratio := new(big.Float).Quo(
				new(big.Float).SetInt(receipts[i].L1Fee),
				new(big.Float).SetInt(baseFee),
			)
			ratioFloat, _ := ratio.Float64()
			gt.Logf("L1Fee ratio (%s / Tx1): %.2f (expected ~%.1f)", txLabels[i], ratioFloat, expectedMultipliers[i])
		}
	}

	// ========== Balance verification ==========
	gt.Log("Verifying actual balance deductions match receipt fees...")

	aliceBalanceAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(gt, err)
	bobBalanceAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(gt, err)
	malloryBalanceAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), dp.Addresses.Mallory, nil)
	require.NoError(gt, err)

	// Record vault balances after
	baseFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), baseFeeVault, nil)
	require.NoError(gt, err)
	seqFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), seqFeeVault, nil)
	require.NoError(gt, err)
	l1FeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), l1FeeVault, nil)
	require.NoError(gt, err)
	opFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(gt.Ctx(), opFeeVault, nil)
	require.NoError(gt, err)

	// Helper: compute expected fee for a receipt
	computeExpectedFee := func(r *types.Receipt) *big.Int {
		l2Fee := new(big.Int).Mul(new(big.Int).SetUint64(r.GasUsed), r.EffectiveGasPrice)
		opFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(r.GasUsed))
		require.NoError(gt, err)
		total := new(big.Int).Add(l2Fee, r.L1Fee)
		total.Add(total, opFee)
		return total
	}

	// --- Mallory: only Tx5 (isolated, best for detecting bug) ---
	malloryActualDeduction := new(big.Int).Sub(malloryBalanceBefore, malloryBalanceAfter)
	malloryExpectedFee := computeExpectedFee(receipts[4])

	malloryL2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipts[4].GasUsed), receipts[4].EffectiveGasPrice)
	malloryOpFee, _ := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipts[4].GasUsed))
	malloryActualL1Fee := new(big.Int).Sub(malloryActualDeduction, malloryL2Fee)
	malloryActualL1Fee.Sub(malloryActualL1Fee, malloryOpFee)

	gt.Logf("Mallory (Tx5 only) fee breakdown:")
	gt.Logf("  Actual total fee:   %s", malloryActualDeduction.String())
	gt.Logf("  Expected total fee: %s", malloryExpectedFee.String())
	gt.Logf("  Receipt L1 fee:     %s", receipts[4].L1Fee.String())
	gt.Logf("  Actual L1 fee:      %s", malloryActualL1Fee.String())

	tx5Match := malloryExpectedFee.Cmp(malloryActualDeduction) == 0
	if !tx5Match {
		gt.Logf("  Tx5 MISMATCH: actual L1=%s, receipt L1=%s", malloryActualL1Fee.String(), receipts[4].L1Fee.String())
	} else {
		gt.Logf("  Tx5 OK")
	}

	// --- Bob: Tx3 + Tx6 ---
	bobActualDeduction := new(big.Int).Sub(bobBalanceBefore, bobBalanceAfter)
	bobExpectedFee := new(big.Int).Add(computeExpectedFee(receipts[2]), computeExpectedFee(receipts[5]))

	bobReceiptL1Fees := new(big.Int).Add(receipts[2].L1Fee, receipts[5].L1Fee)
	bobL2Fees := new(big.Int).Add(
		new(big.Int).Mul(new(big.Int).SetUint64(receipts[2].GasUsed), receipts[2].EffectiveGasPrice),
		new(big.Int).Mul(new(big.Int).SetUint64(receipts[5].GasUsed), receipts[5].EffectiveGasPrice),
	)
	bobOpFee3, _ := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipts[2].GasUsed))
	bobOpFee6, _ := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipts[5].GasUsed))
	bobOpFees := new(big.Int).Add(bobOpFee3, bobOpFee6)
	bobActualL1Fees := new(big.Int).Sub(bobActualDeduction, bobL2Fees)
	bobActualL1Fees.Sub(bobActualL1Fees, bobOpFees)

	gt.Logf("Bob (Tx3+Tx6) fee breakdown:")
	gt.Logf("  Actual total fee:   %s", bobActualDeduction.String())
	gt.Logf("  Expected total fee: %s", bobExpectedFee.String())
	gt.Logf("  Receipt L1 fees:    %s (Tx3=%s + Tx6=%s)", bobReceiptL1Fees.String(), receipts[2].L1Fee.String(), receipts[5].L1Fee.String())
	gt.Logf("  Actual L1 fees:     %s", bobActualL1Fees.String())

	bobMatch := bobExpectedFee.Cmp(bobActualDeduction) == 0
	if !bobMatch {
		gt.Logf("  Bob MISMATCH: actual L1=%s, receipt L1=%s", bobActualL1Fees.String(), bobReceiptL1Fees.String())
	} else {
		gt.Logf("  Bob OK")
	}

	// --- Alice: Tx1 + Tx2 + Tx4 ---
	aliceActualDeduction := new(big.Int).Sub(aliceBalanceBefore, aliceBalanceAfter)
	aliceExpectedFee := new(big.Int).Add(
		computeExpectedFee(receipts[0]),
		new(big.Int).Add(computeExpectedFee(receipts[1]), computeExpectedFee(receipts[3])),
	)

	gt.Logf("Alice (Tx1+Tx2+Tx4) fee breakdown:")
	gt.Logf("  Actual total fee:   %s", aliceActualDeduction.String())
	gt.Logf("  Expected total fee: %s", aliceExpectedFee.String())

	aliceMatch := aliceExpectedFee.Cmp(aliceActualDeduction) == 0
	if !aliceMatch {
		gt.Logf("  Alice MISMATCH")
	} else {
		gt.Logf("  Alice OK")
	}

	// ========== Final Results ==========
	gt.Log("========== Final Results ==========")
	gt.Logf("  Tx5 (Mallory, 1st tx after 2nd ratio change): match=%v", tx5Match)
	gt.Logf("  Bob  (Tx3+Tx6):                                match=%v", bobMatch)
	gt.Logf("  Alice (Tx1+Tx2+Tx4):                           match=%v", aliceMatch)

	require.True(gt, tx5Match,
		"Tx5 (Mallory): actual balance deduction must match receipt fees. "+
			"Actual L1=%s, Receipt L1=%s. "+
			"This means the actual L1 cost deduction used wrong tokenRatio after 2nd ratio change.",
		malloryActualL1Fee.String(), receipts[4].L1Fee.String())

	require.True(gt, bobMatch,
		"Bob (Tx3+Tx6): actual balance deduction must match receipt fees. "+
			"Actual L1=%s, Receipt L1=%s. "+
			"Tx3 likely used wrong tokenRatio after 1st ratio change.",
		bobActualL1Fees.String(), bobReceiptL1Fees.String())

	require.True(gt, aliceMatch,
		"Alice (Tx1+Tx2+Tx4): actual balance deduction must match receipt fees. "+
			"Actual=%s, Expected=%s",
		aliceActualDeduction.String(), aliceExpectedFee.String())

	// ========== Verify Vault Balances (Arsia Mode) ==========
	// Calculate vault increases (independent data source from blockchain)
	baseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultAfter, baseFeeVaultBefore)
	seqFeeVaultIncrease := new(big.Int).Sub(seqFeeVaultAfter, seqFeeVaultBefore)
	l1FeeVaultIncrease := new(big.Int).Sub(l1FeeVaultAfter, l1FeeVaultBefore)
	opFeeVaultIncrease := new(big.Int).Sub(opFeeVaultAfter, opFeeVaultBefore)

	gt.Log("========== Vault Verification ==========")

	// Calculate total L1Fees from all receipts
	totalL1Fee := big.NewInt(0)
	for i := 0; i < 6; i++ {
		totalL1Fee.Add(totalL1Fee, receipts[i].L1Fee)
	}

	// Calculate total OpFees from all receipts
	totalOpFee := big.NewInt(0)
	for i := 0; i < 6; i++ {
		opFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipts[i].GasUsed))
		require.NoError(gt, err)
		totalOpFee.Add(totalOpFee, opFee)
	}

	// Verify L1FeeVault received the sum of all L1Fees
	gt.Logf("  L1FeeVault: expected=%s, actual=%s", totalL1Fee.String(), l1FeeVaultIncrease.String())
	require.Equal(gt, totalL1Fee.String(), l1FeeVaultIncrease.String(),
		"[Arsia] L1FeeVault should receive exactly the sum of all L1Fees")

	// Verify OperatorFeeVault received the sum of all OpFees
	gt.Logf("  OperatorFeeVault: expected=%s, actual=%s", totalOpFee.String(), opFeeVaultIncrease.String())
	require.Equal(gt, totalOpFee.String(), opFeeVaultIncrease.String(),
		"[Arsia] OperatorFeeVault should receive exactly the sum of all OpFees")

	// CRITICAL VERIFICATION: Total vault increase (from blockchain) == Total user deduction (from blockchain)
	// This uses two independent data sources to prevent false positives
	totalUserDeduction := new(big.Int).Add(aliceActualDeduction, bobActualDeduction)
	totalUserDeduction.Add(totalUserDeduction, malloryActualDeduction)

	totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
	totalVaultIncrease.Add(totalVaultIncrease, l1FeeVaultIncrease)
	totalVaultIncrease.Add(totalVaultIncrease, opFeeVaultIncrease)

	gt.Logf("  Total UserDeduction (blockchain): %s", totalUserDeduction.String())
	gt.Logf("  Total VaultIncrease (blockchain): %s", totalVaultIncrease.String())

	require.Equal(gt, totalUserDeduction.String(), totalVaultIncrease.String(),
		"[Arsia] Total vault increase should equal total user deductions")

	gt.Log(" TestMultipleTokenRatioChangesInBlock passed!")
}

// ============================================================================
// Scenario A: Pure Fee Params Change (no ratio change)
// ============================================================================
// Block: Tx1(L1BlockInfo deposit - updates fee params) → Tx2(user tx)
//
// Both Limb and Arsia:
//   Once L1BlockInfo updates fee params, subsequent transactions in the same block
//   immediately apply the NEW fee params for both deduction and receipt.
//   Tx2 (after fee change): deduction=new fee, receipt=new fee
//
// Note: In Limb, L1 cost is collected by inflating gas (l1Gas added to intrinsic gas).
//       The receipt.L1Fee field is for display only and should NOT be separately added
//       to the total fee calculation, as it's already embedded in gasUsed.
// ============================================================================

func TestFeeParamsInBlockChange(t *testing.T) {
	t.Run("Limb", func(t *testing.T) { testFeeParamsInBlockChange(t, "Limb") })
	t.Run("Arsia", func(t *testing.T) { testFeeParamsInBlockChange(t, "Arsia") })
}

func testFeeParamsInBlockChange(gt *testing.T, forkName string) {
	t := actionsHelpers.NewDefaultTesting(gt)
	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())

	// Enable MNT as gas token
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set tokenRatio=1 to simplify (1 ETH = 1 MNT for easier calculation)
	dp.DeployConfig.GasPriceOracleTokenRatio = 1 * 1e6

	// Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000

	// Set MinBaseFee
	dp.DeployConfig.MinBaseFee = 10 * 1e9
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9))

	// Configure fork
	isLimb := forkName == "Limb"
	switch forkName {
	case "Limb":
		limbOffset := hexutil.Uint64(0)
		upgradesHelpers.ApplyLimbTimeOffset(dp, &limbOffset)
		// Limb intrinsic gas includes L1 cost, need very large block gas limit
		dp.DeployConfig.L2GenesisBlockGasLimit = 1_000_000_000_000 // 1 trillion gas
	case "Arsia":
		arsiaOffset := hexutil.Uint64(0)
		upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaOffset)
	}

	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	logger := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, logger)
	batcher := actionsHelpers.NewL2Batcher(logger, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	alice := actionsHelpers.NewBasicUser[any](logger, dp.Secrets.Alice, rand.New(rand.NewSource(1234)))
	alice.SetUserEnv(&actionsHelpers.BasicUserEnv[any]{
		EthCl:  seqEngine.EthClient(),
		Signer: types.LatestSigner(sd.L2Cfg.Config),
	})

	sequencer.ActL2PipelineFull(t)

	// Build initial L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Get contract bindings
	gpoContract, err := bindings.NewGasPriceOracle(
		common.HexToAddress("0x420000000000000000000000000000000000000F"),
		seqEngine.EthClient(),
	)
	require.NoError(t, err)

	sysCfgContract, err := bindings.NewSystemConfig(sd.RollupCfg.L1SystemConfigAddress, miner.EthClient())
	require.NoError(t, err)

	sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.RollupCfg.L1ChainID)
	require.NoError(t, err)

	// ========== DIAGNOSTIC: Print initial fee parameters ==========
	gt.Logf("[%s] ========== DIAGNOSTIC: Initial Fee Parameters ==========", forkName)
	diagBaseFeeScalar, _ := gpoContract.BaseFeeScalar(&bind.CallOpts{})
	diagBlobBaseFeeScalar, _ := gpoContract.BlobBaseFeeScalar(&bind.CallOpts{})
	diagL1BaseFee, _ := gpoContract.L1BaseFee(&bind.CallOpts{})
	diagTokenRatio, _ := gpoContract.TokenRatio(&bind.CallOpts{})
	diagOverhead, _ := gpoContract.Overhead(&bind.CallOpts{})
	diagScalar, _ := gpoContract.Scalar(&bind.CallOpts{})

	gt.Logf("[%s] baseFeeScalar=%d, blobBaseFeeScalar=%d", forkName, diagBaseFeeScalar, diagBlobBaseFeeScalar)
	gt.Logf("[%s] L1BaseFee=%s, TokenRatio=%s", forkName, diagL1BaseFee.String(), diagTokenRatio.String())
	gt.Logf("[%s] Overhead=%s, Scalar=%s", forkName, diagOverhead.String(), diagScalar.String())
	gt.Logf("[%s] ==========================================================", forkName)

	// ========== Reference Transaction (with old fee params) ==========
	gt.Logf("[%s] === Starting Reference Transaction ===", forkName)

	// First, get the current fee parameters BEFORE creating the transaction
	gt.Logf("[%s] Reading BaseFeeScalar...", forkName)
	preBaseFeeScalar, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
	if err != nil {
		gt.Logf("[%s] ERROR reading BaseFeeScalar: %v", forkName, err)
	} else {
		gt.Logf("[%s] baseFeeScalar=%d", forkName, preBaseFeeScalar)
	}
	require.NoError(t, err)

	gt.Logf("[%s] Reading BlobBaseFeeScalar...", forkName)
	preBlobBaseFeeScalar, err := gpoContract.BlobBaseFeeScalar(&bind.CallOpts{})
	if err != nil {
		gt.Logf("[%s] ERROR reading BlobBaseFeeScalar: %v", forkName, err)
	} else {
		gt.Logf("[%s] blobBaseFeeScalar=%d", forkName, preBlobBaseFeeScalar)
	}
	require.NoError(t, err)

	gt.Logf("[%s] Reading L1BaseFee...", forkName)
	preL1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
	if err != nil {
		gt.Logf("[%s] ERROR reading L1BaseFee: %v", forkName, err)
	} else {
		gt.Logf("[%s] L1BaseFee=%s", forkName, preL1BaseFee.String())
	}
	require.NoError(t, err)

	gt.Logf("[%s] Reading TokenRatio...", forkName)
	preTokenRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	if err != nil {
		gt.Logf("[%s] ERROR reading TokenRatio: %v", forkName, err)
	} else {
		gt.Logf("[%s] TokenRatio=%s", forkName, preTokenRatio.String())
	}
	require.NoError(t, err)

	gt.Logf("[%s] === Fee Parameters Summary ===", forkName)
	gt.Logf("[%s] baseFeeScalar=%d, blobBaseFeeScalar=%d", forkName, preBaseFeeScalar, preBlobBaseFeeScalar)
	gt.Logf("[%s] L1BaseFee=%s, TokenRatio=%s", forkName, preL1BaseFee.String(), preTokenRatio.String())

	alice.ActResetTxOpts(t)
	if isLimb {
		// Limb fork includes L1 cost in intrinsic gas, need sufficient gas limit and gas price
		// Set to ~90% of block gas limit to leave room for deposit tx
		alice.ActSetTxGasLimit(270_000_000_000)(t) // 90% of block limit
		// Get current header to check base fee
		latestHeader, err := seqEngine.EthClient().HeaderByNumber(t.Ctx(), nil)
		require.NoError(t, err)
		gt.Logf("[%s] Reference tx - Current BaseFee: %s", forkName, latestHeader.BaseFee.String())

		// Set a very high fixed gas price to avoid l1Gas calculation issues
		gasTipCap := big.NewInt(1000 * params.GWei)
		gasFeeCap := big.NewInt(1000 * params.GWei)
		gt.Logf("[%s] Reference tx - Setting GasTipCap: %s, GasFeeCap: %s", forkName, gasTipCap.String(), gasFeeCap.String())
		alice.ActSetGasFeeCap(gasFeeCap)(t)
		alice.ActSetGasTipCap(gasTipCap)(t)
	}
	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	refReceipt := alice.LastTxReceipt(t)
	require.Equal(t, types.ReceiptStatusSuccessful, refReceipt.Status)
	gt.Logf("[%s] Reference tx L1Fee=%s", forkName, refReceipt.L1Fee.String())

	// Record old fee params based on fork type
	var oldBaseFeeScalar uint32
	var oldBlobBaseFeeScalar uint32
	var oldOverhead *big.Int
	var oldScalar *big.Int

	if isLimb {
		// Limb uses Bedrock params (Overhead + Scalar)
		oldOverhead, err = gpoContract.Overhead(&bind.CallOpts{})
		require.NoError(t, err)
		oldScalar, err = gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		gt.Logf("[%s] Old Bedrock fee params: Overhead=%s, Scalar=%s", forkName, oldOverhead.String(), oldScalar.String())
	} else {
		// Arsia uses Ecotone params (baseFeeScalar + blobBaseFeeScalar)
		oldBaseFeeScalar, err = gpoContract.BaseFeeScalar(&bind.CallOpts{})
		require.NoError(t, err)
		oldBlobBaseFeeScalar, err = gpoContract.BlobBaseFeeScalar(&bind.CallOpts{})
		require.NoError(t, err)
		gt.Logf("[%s] Old Ecotone fee params: baseFeeScalar=%d, blobBaseFeeScalar=%d", forkName, oldBaseFeeScalar, oldBlobBaseFeeScalar)
	}

	// Get L1 base fee and token ratio to understand L1 cost calculation
	l1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	tokenRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(t, err)
	gt.Logf("[%s] L1BaseFee=%s, TokenRatio=%s", forkName, l1BaseFee.String(), tokenRatio.String())

	// ========== Change Fee Params on L1 ==========
	// Submit L2 batch to L1 first
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// Update fee params based on fork type
	if isLimb {
		// Limb uses Bedrock method SetGasConfig(overhead, scalar)
		// Increase Scalar by 20% to make fee params change observable
		newOverhead := oldOverhead
		newScalar := new(big.Int).Mul(oldScalar, big.NewInt(6))
		newScalar.Div(newScalar, big.NewInt(5)) // 1.2x

		gt.Logf("[%s] Updating Bedrock params: Overhead=%s->%s, Scalar=%s->%s",
			forkName, oldOverhead.String(), newOverhead.String(), oldScalar.String(), newScalar.String())

		_, err = sysCfgContract.SetGasConfig(sysCfgOwner, newOverhead, newScalar)
		require.NoError(t, err)
		miner.ActL1StartBlock(12)(t)
		miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
		miner.ActL1EndBlock(t)

		gt.Logf("[%s] New Bedrock fee params: Overhead=%s, Scalar=%s", forkName, newOverhead.String(), newScalar.String())
	} else {
		// Arsia uses Ecotone method SetGasConfigArsia(baseFeeScalar, blobBaseFeeScalar)
		// Double the baseFeeScalar
		newBaseFeeScalar := oldBaseFeeScalar * 2
		newBlobBaseFeeScalar := oldBlobBaseFeeScalar
		if newBlobBaseFeeScalar == 0 {
			newBlobBaseFeeScalar = 1
		}

		gt.Logf("[%s] Updating Ecotone params: baseFeeScalar=%d->%d, blobBaseFeeScalar=%d->%d",
			forkName, oldBaseFeeScalar, newBaseFeeScalar, oldBlobBaseFeeScalar, newBlobBaseFeeScalar)

		_, err = sysCfgContract.SetGasConfigArsia(sysCfgOwner, newBaseFeeScalar, newBlobBaseFeeScalar)
		require.NoError(t, err)
		miner.ActL1StartBlock(12)(t)
		miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
		miner.ActL1EndBlock(t)

		gt.Logf("[%s] New Ecotone fee params: baseFeeScalar=%d, blobBaseFeeScalar=%d", forkName, newBaseFeeScalar, newBlobBaseFeeScalar)
	}

	// ========== Build L2 block that adopts the new L1 origin ==========
	// Build L2 blocks up to but excluding the block that adopts the GPO change
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1HeadExcl(t)

	// Verify L2 hasn't adopted the GPO change yet
	l2BaseFeeScalarBefore, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, oldBaseFeeScalar, l2BaseFeeScalarBefore,
		"L2 baseFeeScalar should not be updated before adopting new L1 origin")

	// Record Alice balance before the critical block
	aliceBalanceBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(t, err)

	// Build L2 block that adopts new L1 origin (L1BlockInfo deposit tx updates fee params)
	// AND include Alice's user tx in the same block
	alice.ActResetTxOpts(t)
	if isLimb {
		// Limb fork includes L1 cost in intrinsic gas, so we need a much larger gas limit
		// Also increase GasFeeCap to ensure effective gas price is high enough for l1Gas calculation
		// Set to ~90% of block gas limit to leave room for deposit tx
		alice.ActSetTxGasLimit(270_000_000_000)(t) // 90% of block limit
		// Get current header to check base fee
		latestHeader, err := seqEngine.EthClient().HeaderByNumber(t.Ctx(), nil)
		require.NoError(t, err)
		gt.Logf("[%s] Current BaseFee: %s", forkName, latestHeader.BaseFee.String())

		// Set a very high fixed gas price to avoid l1Gas calculation issues
		// Use 1000 gwei to ensure msg.GasPrice is large enough for l1Cost / msg.GasPrice
		gasTipCap := big.NewInt(1000 * params.GWei)
		gasFeeCap := big.NewInt(1000 * params.GWei)
		gt.Logf("[%s] Setting GasTipCap: %s, GasFeeCap: %s", forkName, gasTipCap.String(), gasFeeCap.String())
		alice.ActSetGasFeeCap(gasFeeCap)(t)
		alice.ActSetGasTipCap(gasTipCap)(t)
	}

	// Record vault balances before transaction
	baseFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000019")
	seqFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000011")
	l1FeeVault := common.HexToAddress("0x420000000000000000000000000000000000001a")
	opFeeVault := common.HexToAddress("0x420000000000000000000000000000000000001b")

	baseFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), baseFeeVault, nil)
	require.NoError(t, err)
	seqFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), seqFeeVault, nil)
	require.NoError(t, err)
	l1FeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), l1FeeVault, nil)
	require.NoError(t, err)
	opFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), opFeeVault, nil)
	require.NoError(t, err)

	alice.ActMakeTx(t)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	// Record Alice balance after
	aliceBalanceAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(t, err)

	// Record vault balances after transaction
	baseFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), baseFeeVault, nil)
	require.NoError(t, err)
	seqFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), seqFeeVault, nil)
	require.NoError(t, err)
	l1FeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), l1FeeVault, nil)
	require.NoError(t, err)
	opFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), opFeeVault, nil)
	require.NoError(t, err)

	// Verify fee params are now updated in L2
	if isLimb {
		// For Limb, verify Scalar is updated (Bedrock params)
		updatedScalar, err := gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		expectedScalar := new(big.Int).Mul(oldScalar, big.NewInt(6))
		expectedScalar.Div(expectedScalar, big.NewInt(5)) // 1.2x
		require.Equal(t, expectedScalar.String(), updatedScalar.String(),
			"L2 Scalar should be updated after adopting new L1 origin")
		gt.Logf("[%s] L2 Scalar updated: %s → %s", forkName, oldScalar.String(), updatedScalar.String())
	} else {
		// For Arsia, verify baseFeeScalar is updated (Ecotone params)
		updatedBaseFeeScalar, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
		require.NoError(t, err)
		expectedBaseFeeScalar := oldBaseFeeScalar * 2
		require.Equal(t, expectedBaseFeeScalar, updatedBaseFeeScalar,
			"L2 baseFeeScalar should be updated after adopting new L1 origin")
		gt.Logf("[%s] L2 baseFeeScalar updated: %d → %d", forkName, oldBaseFeeScalar, updatedBaseFeeScalar)
	}

	// Get receipt for Alice's tx
	receipt := alice.LastTxReceipt(t)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)

	// Calculate vault balance changes (independent data source from blockchain)
	baseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultAfter, baseFeeVaultBefore)
	seqFeeVaultIncrease := new(big.Int).Sub(seqFeeVaultAfter, seqFeeVaultBefore)
	l1FeeVaultIncrease := new(big.Int).Sub(l1FeeVaultAfter, l1FeeVaultBefore)
	opFeeVaultIncrease := new(big.Int).Sub(opFeeVaultAfter, opFeeVaultBefore)

	// Compute actual deduction from user (independent data source from blockchain)
	actualDeduction := new(big.Int).Sub(aliceBalanceBefore, aliceBalanceAfter)

	// Get current block header to check base fee
	latestBlock, err := seqEngine.EthClient().BlockByNumber(t.Ctx(), nil)
	require.NoError(t, err)
	baseFee := latestBlock.BaseFee()

	gt.Logf("[%s] ========== Fee Verification ==========", forkName)
	gt.Logf("[%s] Receipt GasUsed: %d", forkName, receipt.GasUsed)
	gt.Logf("[%s] Receipt EffectiveGasPrice: %s", forkName, receipt.EffectiveGasPrice.String())
	gt.Logf("[%s] Block BaseFee: %s", forkName, baseFee.String())
	gt.Logf("[%s] Receipt L1Fee (display): %s", forkName, receipt.L1Fee.String())
	gt.Logf("[%s] User balance deduction (blockchain): %s", forkName, actualDeduction.String())
	gt.Logf("[%s] BaseFeeVault increase: %s", forkName, baseFeeVaultIncrease.String())
	gt.Logf("[%s] SeqFeeVault increase: %s", forkName, seqFeeVaultIncrease.String())
	gt.Logf("[%s] L1FeeVault increase: %s", forkName, l1FeeVaultIncrease.String())
	gt.Logf("[%s] OperatorFeeVault increase: %s", forkName, opFeeVaultIncrease.String())

	if isLimb {
		// ========== Limb Mode Verification ==========
		// In Limb: L1 cost and OpFee are embedded in gasUsed (through l1Gas inflation)
		// Total fee = gasUsed * gasPrice (no separate L1Fee or OpFee)
		// Vaults: BaseFeeVault + SeqFeeVault receive all fees, L1FeeVault and OperatorFeeVault receive nothing

		expectedL2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)

		// Verify user deduction matches L2 fee (from receipt data)
		require.Equal(t, expectedL2Fee.String(), actualDeduction.String(),
			"[Limb] User deduction should equal gasUsed * gasPrice (L1 cost embedded in gas)")

		// Verify L1FeeVault didn't receive anything (L1 cost is in gas, not separately deducted)
		require.Equal(t, "0", l1FeeVaultIncrease.String(),
			"[Limb] L1FeeVault should not receive anything (L1 cost embedded in gas)")

		// Verify OperatorFeeVault didn't receive anything
		require.Equal(t, "0", opFeeVaultIncrease.String(),
			"[Limb] OperatorFeeVault should not receive anything (OpFee embedded in gas)")

		// CRITICAL VERIFICATION: Total vault increase (from blockchain) == User deduction (from blockchain)
		// This uses two independent data sources to prevent false positives
		totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
		gt.Logf("[Limb] Total VaultIncrease (blockchain): %s", totalVaultIncrease.String())
		require.Equal(t, actualDeduction.String(), totalVaultIncrease.String(),
			"[Limb] Total vault increase should equal user deduction")

		// ========== Verify receipt.L1Fee matches expected L1 cost calculation ==========
		// Get the transaction to calculate rollup data gas
		tx, _, err := seqEngine.EthClient().TransactionByHash(t.Ctx(), receipt.TxHash)
		require.NoError(t, err)

		// RLP encode the transaction to count zero/non-zero bytes
		txRLP, err := tx.MarshalBinary()
		require.NoError(t, err)

		// Count zero and non-zero bytes
		var zeroes, ones uint64
		for _, b := range txRLP {
			if b == 0 {
				zeroes++
			} else {
				ones++
			}
		}

		// Calculate rollup data gas (Bedrock/Regolith formula)
		// gas = zeroes * 4 + ones * 16
		const TxDataZeroGas = 4
		const TxDataNonZeroGasEIP2028 = 16
		rollupDataGas := zeroes*TxDataZeroGas + ones*TxDataNonZeroGasEIP2028

		// Get current fee params (these are the NEW params after L1BlockInfo update)
		currentOverhead, err := gpoContract.Overhead(&bind.CallOpts{})
		require.NoError(t, err)
		currentScalar, err := gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		currentL1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
		require.NoError(t, err)
		currentTokenRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
		require.NoError(t, err)

		// Calculate expected L1 cost using Bedrock formula:
		// l1Cost = (rollupDataGas + overhead) * l1BaseFee * scalar * tokenRatio / 1e6
		gasWithOverhead := new(big.Int).SetUint64(rollupDataGas)
		gasWithOverhead.Add(gasWithOverhead, currentOverhead)

		expectedL1Cost := new(big.Int).Mul(gasWithOverhead, currentL1BaseFee)
		expectedL1Cost.Mul(expectedL1Cost, currentScalar)
		expectedL1Cost.Mul(expectedL1Cost, currentTokenRatio)
		expectedL1Cost.Div(expectedL1Cost, big.NewInt(1e6))

		gt.Logf("[Limb] Transaction RLP size: %d bytes (zeroes=%d, ones=%d)", len(txRLP), zeroes, ones)
		gt.Logf("[Limb] Rollup data gas: %d", rollupDataGas)
		gt.Logf("[Limb] Fee params: Overhead=%s, Scalar=%s", currentOverhead.String(), currentScalar.String())
		gt.Logf("[Limb] L1BaseFee=%s, TokenRatio=%s", currentL1BaseFee.String(), currentTokenRatio.String())
		gt.Logf("[Limb] Expected L1 cost: %s", expectedL1Cost.String())
		gt.Logf("[Limb] Receipt L1Fee:    %s", receipt.L1Fee.String())

		// Verify receipt.L1Fee matches expected L1 cost
		require.Equal(t, expectedL1Cost.String(), receipt.L1Fee.String(),
			"[Limb] receipt.L1Fee should match calculated L1 cost using new fee params")

		gt.Logf("[Limb]  Fee structure verified: L1 cost embedded in gas, receipt.L1Fee matches calculation")

	} else {
		// ========== Arsia Mode Verification ==========
		// In Arsia: L1Fee and OpFee are separately deducted
		// Total fee = gasUsed * gasPrice + L1Fee + OpFee
		// Vaults: BaseFeeVault + SeqFeeVault receive L2 fee, L1FeeVault receives L1Fee, OperatorFeeVault receives OpFee

		l2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)
		opFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt.GasUsed))
		require.NoError(t, err)

		expectedTotalFee := new(big.Int).Add(l2Fee, receipt.L1Fee)
		expectedTotalFee.Add(expectedTotalFee, opFee)

		// Verify user deduction (from blockchain) matches total fee (from receipt)
		require.Equal(t, expectedTotalFee.String(), actualDeduction.String(),
			"[Arsia] User deduction should equal gasUsed * gasPrice + L1Fee + OpFee")

		// Verify L1FeeVault received the L1Fee
		require.Equal(t, receipt.L1Fee.String(), l1FeeVaultIncrease.String(),
			"[Arsia] L1FeeVault should receive exactly receipt.L1Fee")

		// Verify OperatorFeeVault received the OpFee
		gt.Logf("[Arsia] OperatorFeeVault: expected=%s, actual=%s", opFee.String(), opFeeVaultIncrease.String())
		require.Equal(t, opFee.String(), opFeeVaultIncrease.String(),
			"[Arsia] OperatorFeeVault should receive exactly the OpFee")

		// CRITICAL VERIFICATION: Total vault increase (from blockchain) == User deduction (from blockchain)
		// This uses two independent data sources to prevent false positives
		totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
		totalVaultIncrease.Add(totalVaultIncrease, l1FeeVaultIncrease)
		totalVaultIncrease.Add(totalVaultIncrease, opFeeVaultIncrease)

		gt.Logf("[Arsia] L2 Fee: %s, L1Fee: %s, OpFee: %s", l2Fee.String(), receipt.L1Fee.String(), opFee.String())
		gt.Logf("[Arsia] Total VaultIncrease (blockchain): %s", totalVaultIncrease.String())
		require.Equal(t, actualDeduction.String(), totalVaultIncrease.String(),
			"[Arsia] Total vault increase should equal user deduction")

		gt.Logf("[Arsia]  Fee structure verified: L1Fee and OpFee separately deducted, all vaults match")
	}

	// Note: We don't verify L1Fee ratio here because L1BaseFee changes dynamically
	// during L1 block production, which affects the final L1Cost calculation.
	// The Scalar/baseFeeScalar update itself is already verified above.

	gt.Logf("[%s] Scenario A test passed!", forkName)
}

// ============================================================================
// Scenario B: Pure Token Ratio Change (no fee params change)
// ============================================================================
// Block: Tx1(user tx) → Tx2(SetTokenRatio) → Tx3(user tx)
//
// Both Limb and Arsia:
//   Tx1: uses old tokenRatio
//   Tx2: SetTokenRatio itself uses old tokenRatio for its own fee calculation
//   Tx3: uses NEW tokenRatio (change applies immediately after Tx2 completes)
// ============================================================================

func TestTokenRatioInBlockChange_Forks(t *testing.T) {
	t.Run("Limb", func(t *testing.T) { testTokenRatioInBlockChange_Forks(t, "Limb") })
	t.Run("Arsia", func(t *testing.T) { testTokenRatioInBlockChange_Forks(t, "Arsia") })
}

func testTokenRatioInBlockChange_Forks(gt *testing.T, forkName string) {
	t := actionsHelpers.NewDefaultTesting(gt)
	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())

	// Enable MNT as gas token
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set initial tokenRatio = 1 MNT/ETH (simplified for Limb gas calculation)
	// Will be changed to higher value (e.g., 2000) later in the test
	initialTokenRatio := uint64(1 * 1e6)
	dp.DeployConfig.GasPriceOracleTokenRatio = initialTokenRatio

	// Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000

	// Set MinBaseFee
	dp.DeployConfig.MinBaseFee = 10 * 1e9
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9))

	// Configure fork
	isLimb := forkName == "Limb"
	switch forkName {
	case "Limb":
		limbOffset := hexutil.Uint64(0)
		upgradesHelpers.ApplyLimbTimeOffset(dp, &limbOffset)
		// Limb intrinsic gas includes L1 cost, need larger block gas limit
		dp.DeployConfig.L2GenesisBlockGasLimit = 300_000_000_000 // 300B gas
	case "Arsia":
		arsiaOffset := hexutil.Uint64(0)
		upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaOffset)
	}

	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	logger := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, logger)

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

	sequencer.ActL2PipelineFull(t)

	// Build initial L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Get GPO contract binding
	gpoContract, err := bindings.NewGasPriceOracle(
		common.HexToAddress("0x420000000000000000000000000000000000000F"),
		seqEngine.EthClient(),
	)
	require.NoError(t, err)

	// Setup GPO operator
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
		gt.Fatalf("Unknown GPO owner: %s", gpoOwnerAddr.String())
	}

	gpoOwner, err := bind.NewKeyedTransactorWithChainID(ownerKey, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)

	// For Limb, set high gas limit AND gas price (L1 cost in intrinsic gas)
	if isLimb {
		gpoOwner.GasLimit = 90_000_000_000 // ~30% of block limit (3 txs in block)
		// High gas price reduces l1Gas = l1Cost / gasPrice
		gpoOwner.GasFeeCap = big.NewInt(3000 * params.GWei)
		gpoOwner.GasTipCap = big.NewInt(3000 * params.GWei)
	}

	// Set operator to Alice
	_, err = gpoContract.SetOperator(gpoOwner, dp.Addresses.Alice)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(gpoOwnerAddr)(t)
	sequencer.ActL2EndBlock(t)

	// Set initial tokenRatio
	gpoOperator, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Alice, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)

	// For Limb, set high gas limit AND gas price (L1 cost in intrinsic gas)
	if isLimb {
		gpoOperator.GasLimit = 90_000_000_000 // ~30% of block limit (3 txs in block)
		// High gas price reduces l1Gas = l1Cost / gasPrice
		gpoOperator.GasFeeCap = big.NewInt(3000 * params.GWei)
		gpoOperator.GasTipCap = big.NewInt(3000 * params.GWei)
	}

	tokenRatioValue := new(big.Int).SetUint64(initialTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, tokenRatioValue)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	// Verify initial tokenRatio
	currentRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, tokenRatioValue.Uint64(), currentRatio.Uint64(), "initial tokenRatio mismatch")
	gt.Logf("[%s] Initial tokenRatio: %d", forkName, currentRatio.Uint64())

	// ========== Build block with 3 transactions ==========
	// Tx1: Alice normal tx (uses old ratio)
	// Tx2: SetTokenRatio (uses old ratio)
	// Tx3: Bob normal tx (uses new ratio)

	newTokenRatio := uint64(15 * 1e5) // 1.5x the ratio (1 -> 1.5)

	// Record balances before
	aliceBalBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(t, err)
	bobBalBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(t, err)

	// Record vault balances before (for verification)
	baseFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000019")
	seqFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000011")
	l1FeeVault := common.HexToAddress("0x420000000000000000000000000000000000001a")
	opFeeVault := common.HexToAddress("0x420000000000000000000000000000000000001b")

	baseFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), baseFeeVault, nil)
	require.NoError(t, err)
	seqFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), seqFeeVault, nil)
	require.NoError(t, err)
	l1FeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), l1FeeVault, nil)
	require.NoError(t, err)
	opFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), opFeeVault, nil)
	require.NoError(t, err)

	sequencer.ActL2StartBlock(t)

	// Tx1: Alice normal tx
	alice.ActResetTxOpts(t)
	if isLimb {
		// Limb needs high gas limit and gas price for L1 cost in intrinsic gas
		alice.ActSetTxGasLimit(90_000_000_000)(t) // ~30% of block limit (3 txs in block)
		gasTipCap := big.NewInt(3000 * params.GWei)
		gasFeeCap := big.NewInt(3000 * params.GWei)
		alice.ActSetGasFeeCap(gasFeeCap)(t)
		alice.ActSetGasTipCap(gasTipCap)(t)
	}
	alice.ActMakeTx(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)

	// Tx2: SetTokenRatio (Alice is operator)
	newTokenRatioValue := new(big.Int).SetUint64(newTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, newTokenRatioValue)
	require.NoError(t, err)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)

	// Tx3: Bob normal tx (should use NEW tokenRatio)
	bob.ActResetTxOpts(t)
	if isLimb {
		// Limb needs high gas limit and gas price for L1 cost in intrinsic gas
		bob.ActSetTxGasLimit(90_000_000_000)(t) // ~30% of block limit (3 txs in block)
		gasTipCap := big.NewInt(3000 * params.GWei)
		gasFeeCap := big.NewInt(3000 * params.GWei)
		bob.ActSetGasFeeCap(gasFeeCap)(t)
		bob.ActSetGasTipCap(gasTipCap)(t)
	}
	bob.ActMakeTx(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(t)

	sequencer.ActL2EndBlock(t)

	// Record balances after
	aliceBalAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(t, err)
	bobBalAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(t, err)

	// Record vault balances after
	baseFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), baseFeeVault, nil)
	require.NoError(t, err)
	seqFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), seqFeeVault, nil)
	require.NoError(t, err)
	l1FeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), l1FeeVault, nil)
	require.NoError(t, err)
	opFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), opFeeVault, nil)
	require.NoError(t, err)

	// Verify tokenRatio is updated
	updatedRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, newTokenRatio, updatedRatio.Uint64(), "tokenRatio should be updated")
	gt.Logf("[%s] TokenRatio updated: %d → %d", forkName, initialTokenRatio, updatedRatio.Uint64())

	// Get block and receipts
	latestBlock, err := seqEngine.EthClient().BlockByNumber(t.Ctx(), nil)
	require.NoError(t, err)
	txs := latestBlock.Transactions()
	// Block should have: deposit tx + Tx1 + Tx2 + Tx3 = 4 txs
	require.GreaterOrEqual(t, len(txs), 4, "block should have at least 4 transactions")

	// Get receipts for user txs (skip deposit tx at index 0)
	receipt1, err := seqEngine.EthClient().TransactionReceipt(t.Ctx(), txs[1].Hash())
	require.NoError(t, err)
	receipt2, err := seqEngine.EthClient().TransactionReceipt(t.Ctx(), txs[2].Hash())
	require.NoError(t, err)
	receipt3, err := seqEngine.EthClient().TransactionReceipt(t.Ctx(), txs[3].Hash())
	require.NoError(t, err)

	require.Equal(t, types.ReceiptStatusSuccessful, receipt1.Status, "Tx1 should succeed")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt2.Status, "Tx2 should succeed")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt3.Status, "Tx3 should succeed")

	gt.Logf("[%s] Tx1 (Alice normal) L1Fee: %s", forkName, receipt1.L1Fee.String())
	gt.Logf("[%s] Tx2 (SetTokenRatio) L1Fee: %s", forkName, receipt2.L1Fee.String())
	gt.Logf("[%s] Tx3 (Bob normal) L1Fee: %s", forkName, receipt3.L1Fee.String())

	// ========== Verify Tx2 (SetTokenRatio): deduction=old ratio, receipt=old ratio ==========
	// Alice's total deduction covers Tx1 + Tx2
	aliceTotalDeduction := new(big.Int).Sub(aliceBalBefore, aliceBalAfter)

	// Calculate expected fees based on fork type
	var tx1ExpectedFee, tx2ExpectedFee *big.Int

	if isLimb {
		// Limb: L1 cost embedded in gas, no separate L1Fee or OpFee
		tx1L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt1.GasUsed), receipt1.EffectiveGasPrice)
		tx1ExpectedFee = tx1L2Fee

		tx2L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt2.GasUsed), receipt2.EffectiveGasPrice)
		tx2ExpectedFee = tx2L2Fee
	} else {
		// Arsia: L1Fee and OpFee separately deducted
		tx1L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt1.GasUsed), receipt1.EffectiveGasPrice)
		tx1OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt1.GasUsed))
		require.NoError(t, err)
		tx1ExpectedFee = new(big.Int).Add(tx1L2Fee, receipt1.L1Fee)
		tx1ExpectedFee.Add(tx1ExpectedFee, tx1OpFee)

		tx2L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt2.GasUsed), receipt2.EffectiveGasPrice)
		tx2OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
		require.NoError(t, err)
		tx2ExpectedFee = new(big.Int).Add(tx2L2Fee, receipt2.L1Fee)
		tx2ExpectedFee.Add(tx2ExpectedFee, tx2OpFee)
	}

	aliceExpectedTotal := new(big.Int).Add(tx1ExpectedFee, tx2ExpectedFee)

	gt.Logf("[%s] Alice expected total (from receipts): %s", forkName, aliceExpectedTotal.String())
	gt.Logf("[%s] Alice actual total deduction:         %s", forkName, aliceTotalDeduction.String())

	// Both Limb and Arsia: Alice's deduction should match receipts (Tx1 old ratio, Tx2 old ratio)
	require.Equal(t, aliceExpectedTotal.Cmp(aliceTotalDeduction), 0,
		"[%s] Alice total deduction should match receipt-based expected. "+
			"Expected=%s, Actual=%s", forkName, aliceExpectedTotal.String(), aliceTotalDeduction.String())

	// ========== Verify Tx3 (Bob normal): deduction=new ratio, receipt=new ratio ==========
	bobDeduction := new(big.Int).Sub(bobBalBefore, bobBalAfter)

	var tx3ExpectedFee *big.Int
	if isLimb {
		// Limb: L1 cost embedded in gas, no separate L1Fee or OpFee
		tx3L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt3.GasUsed), receipt3.EffectiveGasPrice)
		tx3ExpectedFee = tx3L2Fee
	} else {
		// Arsia: L1Fee and OpFee separately deducted
		tx3L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt3.GasUsed), receipt3.EffectiveGasPrice)
		tx3OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt3.GasUsed))
		require.NoError(t, err)
		tx3ExpectedFee = new(big.Int).Add(tx3L2Fee, receipt3.L1Fee)
		tx3ExpectedFee.Add(tx3ExpectedFee, tx3OpFee)
	}

	gt.Logf("[%s] Bob expected (from receipt): %s", forkName, tx3ExpectedFee.String())
	gt.Logf("[%s] Bob actual deduction:        %s", forkName, bobDeduction.String())

	// Both Limb and Arsia: Bob's deduction should match receipt (both use new ratio)
	require.Equal(t, tx3ExpectedFee.Cmp(bobDeduction), 0,
		"[%s] Bob deduction should match receipt-based expected (new ratio). "+
			"Expected=%s, Actual=%s", forkName, tx3ExpectedFee.String(), bobDeduction.String())

	// Verify Tx3 L1Fee is roughly 1.5x Tx1 L1Fee (since ratio increased 1.5x)
	if receipt1.L1Fee.Sign() > 0 && receipt3.L1Fee.Sign() > 0 {
		ratio := new(big.Float).Quo(
			new(big.Float).SetInt(receipt3.L1Fee),
			new(big.Float).SetInt(receipt1.L1Fee),
		)
		gt.Logf("[%s] Tx3/Tx1 L1Fee ratio: %s (expected ~1.5 since tokenRatio increased 1.5x)", forkName, ratio.String())
	}

	// ========== Verify Vault Balances ==========
	// Calculate vault increases (independent data source from blockchain)
	baseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultAfter, baseFeeVaultBefore)
	seqFeeVaultIncrease := new(big.Int).Sub(seqFeeVaultAfter, seqFeeVaultBefore)
	l1FeeVaultIncrease := new(big.Int).Sub(l1FeeVaultAfter, l1FeeVaultBefore)
	opFeeVaultIncrease := new(big.Int).Sub(opFeeVaultAfter, opFeeVaultBefore)

	if isLimb {
		// ========== Limb Mode Vault Verification ==========
		// In Limb: L1 cost and OpFee are embedded in gasUsed (through l1Gas inflation)
		// Vaults: BaseFeeVault + SeqFeeVault receive all fees, L1FeeVault and OperatorFeeVault receive nothing

		gt.Logf("[%s] Vault verification:", forkName)
		gt.Logf("[%s]   L1FeeVault increase: %s (expected: 0)", forkName, l1FeeVaultIncrease.String())
		gt.Logf("[%s]   OperatorFeeVault increase: %s (expected: 0)", forkName, opFeeVaultIncrease.String())

		// Verify L1FeeVault didn't receive anything
		require.Equal(t, "0", l1FeeVaultIncrease.String(),
			"[Limb] L1FeeVault should not receive anything (L1 cost embedded in gas)")

		// Verify OperatorFeeVault didn't receive anything
		require.Equal(t, "0", opFeeVaultIncrease.String(),
			"[Limb] OperatorFeeVault should not receive anything (OpFee embedded in gas)")

		// CRITICAL VERIFICATION: Total vault increase (from blockchain) == Total user deduction (from blockchain)
		// This uses two independent data sources to prevent false positives
		totalUserDeduction := new(big.Int).Add(aliceTotalDeduction, bobDeduction)
		totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
		gt.Logf("[%s]   Total UserDeduction (blockchain): %s", forkName, totalUserDeduction.String())
		gt.Logf("[%s]   Total VaultIncrease (blockchain): %s", forkName, totalVaultIncrease.String())
		require.Equal(t, totalUserDeduction.String(), totalVaultIncrease.String(),
			"[Limb] Total vault increase should equal total user deductions")

		// ========== Manually Calculate and Verify L1 Cost for Each Transaction ==========
		// Get current fee params (after all updates)
		currentOverhead, err := gpoContract.Overhead(&bind.CallOpts{})
		require.NoError(t, err)
		currentScalar, err := gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		currentL1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
		require.NoError(t, err)

		// Verify Tx1 L1Fee (uses old tokenRatio = 1e6)
		tx1, _, err := seqEngine.EthClient().TransactionByHash(t.Ctx(), receipt1.TxHash)
		require.NoError(t, err)
		tx1RLP, err := tx1.MarshalBinary()
		require.NoError(t, err)

		var tx1Zeroes, tx1Ones uint64
		for _, b := range tx1RLP {
			if b == 0 {
				tx1Zeroes++
			} else {
				tx1Ones++
			}
		}
		tx1RollupDataGas := tx1Zeroes*4 + tx1Ones*16

		tx1GasWithOverhead := new(big.Int).SetUint64(tx1RollupDataGas)
		tx1GasWithOverhead.Add(tx1GasWithOverhead, currentOverhead)
		tx1ExpectedL1Cost := new(big.Int).Mul(tx1GasWithOverhead, currentL1BaseFee)
		tx1ExpectedL1Cost.Mul(tx1ExpectedL1Cost, currentScalar)
		tx1ExpectedL1Cost.Mul(tx1ExpectedL1Cost, big.NewInt(int64(initialTokenRatio))) // Old ratio
		tx1ExpectedL1Cost.Div(tx1ExpectedL1Cost, big.NewInt(1e6))

		require.Equal(t, tx1ExpectedL1Cost.String(), receipt1.L1Fee.String(),
			"[Limb] Tx1 receipt.L1Fee should match calculated L1 cost (old tokenRatio)")

		// Verify Tx3 L1Fee (uses new tokenRatio = 2e6)
		tx3, _, err := seqEngine.EthClient().TransactionByHash(t.Ctx(), receipt3.TxHash)
		require.NoError(t, err)
		tx3RLP, err := tx3.MarshalBinary()
		require.NoError(t, err)

		var tx3Zeroes, tx3Ones uint64
		for _, b := range tx3RLP {
			if b == 0 {
				tx3Zeroes++
			} else {
				tx3Ones++
			}
		}
		tx3RollupDataGas := tx3Zeroes*4 + tx3Ones*16

		tx3GasWithOverhead := new(big.Int).SetUint64(tx3RollupDataGas)
		tx3GasWithOverhead.Add(tx3GasWithOverhead, currentOverhead)
		tx3ExpectedL1Cost := new(big.Int).Mul(tx3GasWithOverhead, currentL1BaseFee)
		tx3ExpectedL1Cost.Mul(tx3ExpectedL1Cost, currentScalar)
		tx3ExpectedL1Cost.Mul(tx3ExpectedL1Cost, big.NewInt(int64(newTokenRatio))) // New ratio
		tx3ExpectedL1Cost.Div(tx3ExpectedL1Cost, big.NewInt(1e6))

		require.Equal(t, tx3ExpectedL1Cost.String(), receipt3.L1Fee.String(),
			"[Limb] Tx3 receipt.L1Fee should match calculated L1 cost (new tokenRatio)")

	} else {
		// ========== Arsia Mode Vault Verification ==========
		// In Arsia: L1Fee and OpFee are separately deducted
		// Vaults: BaseFeeVault + SeqFeeVault receive L2 fee, L1FeeVault receives L1Fee, OperatorFeeVault receives OpFee

		gt.Logf("[%s] Vault verification:", forkName)

		// Calculate expected L1Fees and OpFees from receipts
		totalL1Fee := new(big.Int).Add(receipt1.L1Fee, receipt2.L1Fee)
		totalL1Fee.Add(totalL1Fee, receipt3.L1Fee)

		tx1OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt1.GasUsed))
		require.NoError(t, err)
		tx2OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
		require.NoError(t, err)
		tx3OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt3.GasUsed))
		require.NoError(t, err)
		totalOpFee := new(big.Int).Add(tx1OpFee, tx2OpFee)
		totalOpFee.Add(totalOpFee, tx3OpFee)

		// Verify L1FeeVault received the sum of all L1Fees
		gt.Logf("[%s]   L1FeeVault: expected=%s, actual=%s", forkName, totalL1Fee.String(), l1FeeVaultIncrease.String())
		require.Equal(t, totalL1Fee.String(), l1FeeVaultIncrease.String(),
			"[Arsia] L1FeeVault should receive sum of all L1Fees")

		// Verify OperatorFeeVault received the sum of all OpFees
		gt.Logf("[%s]   OperatorFeeVault: expected=%s, actual=%s", forkName, totalOpFee.String(), opFeeVaultIncrease.String())
		require.Equal(t, totalOpFee.String(), opFeeVaultIncrease.String(),
			"[Arsia] OperatorFeeVault should receive sum of all OpFees")

		// CRITICAL VERIFICATION: Total vault increase (from blockchain) == Total user deduction (from blockchain)
		// This uses two independent data sources to prevent false positives
		totalUserDeduction := new(big.Int).Add(aliceTotalDeduction, bobDeduction)
		totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
		totalVaultIncrease.Add(totalVaultIncrease, l1FeeVaultIncrease)
		totalVaultIncrease.Add(totalVaultIncrease, opFeeVaultIncrease)

		gt.Logf("[%s]   Total UserDeduction (blockchain): %s", forkName, totalUserDeduction.String())
		gt.Logf("[%s]   Total VaultIncrease (blockchain): %s", forkName, totalVaultIncrease.String())
		require.Equal(t, totalUserDeduction.String(), totalVaultIncrease.String(),
			"[Arsia] Total vault increase should equal total user deductions")
	}

	gt.Logf("[%s]  Scenario B test passed!", forkName)
}

// ============================================================================
// Scenario C: Fee Params + Token Ratio Combined Change
// ============================================================================
// Block: Tx1(L1BlockInfo deposit - updates fee params) → Tx2(user tx) → Tx3(SetTokenRatio) → Tx4(user tx)
//
// Both Limb and Arsia:
//   Tx1 (L1BlockInfo): updates fee params → immediately effective for Tx2
//   Tx2: uses new fee params + old tokenRatio
//   Tx3 (SetTokenRatio): uses new fee params + old tokenRatio (for its own fee)
//                        updates tokenRatio → immediately effective for Tx4
//   Tx4: uses new fee params + new tokenRatio
//
// Note: Both L1BlockInfo and SetTokenRatio apply changes immediately after execution.
//       Subsequent transactions in the same block use the new values.
// ============================================================================

func TestFeeParamsAndRatioInBlockChange(t *testing.T) {
	t.Run("Limb", func(t *testing.T) { testFeeParamsAndRatioInBlockChange(t, "Limb") })
	t.Run("Arsia", func(t *testing.T) { testFeeParamsAndRatioInBlockChange(t, "Arsia") })
}

func testFeeParamsAndRatioInBlockChange(gt *testing.T, forkName string) {
	t := actionsHelpers.NewDefaultTesting(gt)
	dp := e2eutils.MakeMantleDeployParams(t, actionsHelpers.DefaultRollupTestParams())

	// Enable MNT as gas token
	dp.DeployConfig.UseCustomGasToken = true
	dp.DeployConfig.GasPayingTokenName = "MNT"
	dp.DeployConfig.GasPayingTokenSymbol = "MNT"
	dp.DeployConfig.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
	dp.DeployConfig.LiquidityControllerOwner = dp.Addresses.Deployer

	// Set initial tokenRatio = 1 MNT/ETH (simplified for Limb gas calculation)
	// Will be changed to higher value (e.g., 2000) later in the test
	initialTokenRatio := uint64(1 * 1e6)
	dp.DeployConfig.GasPriceOracleTokenRatio = initialTokenRatio

	// Set Operator Fee parameters
	dp.DeployConfig.GasPriceOracleOperatorFeeScalar = 1000
	dp.DeployConfig.GasPriceOracleOperatorFeeConstant = 1000

	// Set MinBaseFee
	dp.DeployConfig.MinBaseFee = 10 * 1e9
	dp.DeployConfig.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(10 * 1e9))

	// Configure fork
	isLimb := forkName == "Limb"
	switch forkName {
	case "Limb":
		limbOffset := hexutil.Uint64(0)
		upgradesHelpers.ApplyLimbTimeOffset(dp, &limbOffset)
		// Limb intrinsic gas includes L1 cost, need larger block gas limit
		dp.DeployConfig.L2GenesisBlockGasLimit = 300_000_000_000 // 300B gas
	case "Arsia":
		arsiaOffset := hexutil.Uint64(0)
		upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaOffset)
	}

	sd := e2eutils.SetupMantleNormal(t, dp, actionsHelpers.DefaultAlloc)
	logger := testlog.Logger(t, log.LevelDebug)
	miner, seqEngine, sequencer := actionsHelpers.SetupSequencerTest(t, sd, logger)
	batcher := actionsHelpers.NewL2Batcher(logger, sd.RollupCfg, actionsHelpers.MantleSingularBatcherCfg(dp),
		sequencer.RollupClient(), miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

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

	sequencer.ActL2PipelineFull(t)

	// Build initial L1 and L2 blocks
	miner.ActEmptyBlock(t)
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1Head(t)

	// Get contract bindings
	gpoContract, err := bindings.NewGasPriceOracle(
		common.HexToAddress("0x420000000000000000000000000000000000000F"),
		seqEngine.EthClient(),
	)
	require.NoError(t, err)

	sysCfgContract, err := bindings.NewSystemConfig(sd.RollupCfg.L1SystemConfigAddress, miner.EthClient())
	require.NoError(t, err)

	sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Deployer, sd.RollupCfg.L1ChainID)
	require.NoError(t, err)

	// Setup GPO operator
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
		gt.Fatalf("Unknown GPO owner: %s", gpoOwnerAddr.String())
	}

	gpoOwner, err := bind.NewKeyedTransactorWithChainID(ownerKey, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)

	// For Limb, set high gas limit AND gas price (L1 cost in intrinsic gas)
	if isLimb {
		gpoOwner.GasLimit = 90_000_000_000 // ~30% of block limit (3 txs in block)
		// High gas price reduces l1Gas = l1Cost / gasPrice
		gpoOwner.GasFeeCap = big.NewInt(3000 * params.GWei)
		gpoOwner.GasTipCap = big.NewInt(3000 * params.GWei)
	}

	// Set operator to Alice
	_, err = gpoContract.SetOperator(gpoOwner, dp.Addresses.Alice)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(gpoOwnerAddr)(t)
	sequencer.ActL2EndBlock(t)

	// Set initial tokenRatio
	gpoOperator, err := bind.NewKeyedTransactorWithChainID(dp.Secrets.Alice, sd.L2Cfg.Config.ChainID)
	require.NoError(t, err)

	// For Limb, set high gas limit AND gas price (L1 cost in intrinsic gas)
	if isLimb {
		gpoOperator.GasLimit = 90_000_000_000 // ~30% of block limit (3 txs in block)
		// High gas price reduces l1Gas = l1Cost / gasPrice
		gpoOperator.GasFeeCap = big.NewInt(3000 * params.GWei)
		gpoOperator.GasTipCap = big.NewInt(3000 * params.GWei)
	}

	tokenRatioValue := new(big.Int).SetUint64(initialTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, tokenRatioValue)
	require.NoError(t, err)
	sequencer.ActL2StartBlock(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)
	sequencer.ActL2EndBlock(t)

	// Record old fee params based on fork type
	var oldBaseFeeScalar uint32
	var oldBlobBaseFeeScalar uint32
	var oldOverhead *big.Int
	var oldScalar *big.Int

	if isLimb {
		// Limb uses Bedrock params (Overhead + Scalar)
		oldOverhead, err = gpoContract.Overhead(&bind.CallOpts{})
		require.NoError(t, err)
		oldScalar, err = gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		gt.Logf("[%s] Old Bedrock fee params: Overhead=%s, Scalar=%s", forkName, oldOverhead.String(), oldScalar.String())
	} else {
		// Arsia uses Ecotone params (baseFeeScalar + blobBaseFeeScalar)
		oldBaseFeeScalar, err = gpoContract.BaseFeeScalar(&bind.CallOpts{})
		require.NoError(t, err)
		oldBlobBaseFeeScalar, err = gpoContract.BlobBaseFeeScalar(&bind.CallOpts{})
		require.NoError(t, err)
		gt.Logf("[%s] Old Ecotone fee params: baseFeeScalar=%d, blobBaseFeeScalar=%d", forkName, oldBaseFeeScalar, oldBlobBaseFeeScalar)
	}

	// Get L1 base fee and token ratio to understand L1 cost calculation
	l1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
	require.NoError(t, err)
	tokenRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(t, err)
	gt.Logf("[%s] L1BaseFee=%s, TokenRatio=%s", forkName, l1BaseFee.String(), tokenRatio.String())
	gt.Logf("[%s] Initial tokenRatio: %d", forkName, initialTokenRatio)

	// ========== Change Fee Params on L1 ==========
	batcher.ActSubmitAll(t)
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTx(dp.Addresses.Batcher)(t)
	miner.ActL1EndBlock(t)

	// Update fee params based on fork type
	if isLimb {
		// Limb uses Bedrock method SetGasConfig(overhead, scalar)
		// Increase Scalar by 20% to make fee params change observable
		newOverhead := oldOverhead
		newScalar := new(big.Int).Mul(oldScalar, big.NewInt(6))
		newScalar.Div(newScalar, big.NewInt(5)) // 1.2x

		gt.Logf("[%s] Updating Bedrock params: Overhead=%s->%s, Scalar=%s->%s",
			forkName, oldOverhead.String(), newOverhead.String(), oldScalar.String(), newScalar.String())

		_, err = sysCfgContract.SetGasConfig(sysCfgOwner, newOverhead, newScalar)
		require.NoError(t, err)
		miner.ActL1StartBlock(12)(t)
		miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
		miner.ActL1EndBlock(t)

		gt.Logf("[%s] New Bedrock fee params: Overhead=%s, Scalar=%s", forkName, newOverhead.String(), newScalar.String())
	} else {
		// Arsia uses Ecotone method SetGasConfigArsia(baseFeeScalar, blobBaseFeeScalar)
		// Double the baseFeeScalar
		newBaseFeeScalar := oldBaseFeeScalar * 2
		newBlobBaseFeeScalar := oldBlobBaseFeeScalar
		if newBlobBaseFeeScalar == 0 {
			newBlobBaseFeeScalar = 1
		}

		gt.Logf("[%s] Updating Ecotone params: baseFeeScalar=%d->%d, blobBaseFeeScalar=%d->%d",
			forkName, oldBaseFeeScalar, newBaseFeeScalar, oldBlobBaseFeeScalar, newBlobBaseFeeScalar)

		_, err = sysCfgContract.SetGasConfigArsia(sysCfgOwner, newBaseFeeScalar, newBlobBaseFeeScalar)
		require.NoError(t, err)
		miner.ActL1StartBlock(12)(t)
		miner.ActL1IncludeTx(dp.Addresses.Deployer)(t)
		miner.ActL1EndBlock(t)

		gt.Logf("[%s] New Ecotone fee params: baseFeeScalar=%d, blobBaseFeeScalar=%d", forkName, newBaseFeeScalar, newBlobBaseFeeScalar)
	}

	// ========== Build L2 block that adopts new L1 origin ==========
	// Build L2 blocks up to but excluding the block that adopts the GPO change
	sequencer.ActL1HeadSignal(t)
	sequencer.ActBuildToL1HeadExcl(t)

	// New tokenRatio = 1.5 MNT/ETH (1.5x increase)
	newTokenRatio := uint64(15 * 1e5)

	// Record balances before the critical block
	aliceBalBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(t, err)
	bobBalBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(t, err)

	// Record vault balances before (for verification)
	baseFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000019")
	seqFeeVault := common.HexToAddress("0x4200000000000000000000000000000000000011")
	l1FeeVault := common.HexToAddress("0x420000000000000000000000000000000000001a")
	opFeeVault := common.HexToAddress("0x420000000000000000000000000000000000001b")

	baseFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), baseFeeVault, nil)
	require.NoError(t, err)
	seqFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), seqFeeVault, nil)
	require.NoError(t, err)
	l1FeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), l1FeeVault, nil)
	require.NoError(t, err)
	opFeeVaultBefore, err := seqEngine.EthClient().BalanceAt(t.Ctx(), opFeeVault, nil)
	require.NoError(t, err)

	// Build L2 block: Tx1(L1BlockInfo deposit) → Tx2(Alice user tx) → Tx3(SetTokenRatio) → Tx4(Bob user tx)
	sequencer.ActL2StartBlock(t)

	// Tx2: Alice normal tx
	alice.ActResetTxOpts(t)
	if isLimb {
		// Limb needs high gas limit and gas price for L1 cost in intrinsic gas
		alice.ActSetTxGasLimit(90_000_000_000)(t) // ~25% of block limit (4 txs in block)
		gasTipCap := big.NewInt(3000 * params.GWei)
		gasFeeCap := big.NewInt(3000 * params.GWei)
		alice.ActSetGasFeeCap(gasFeeCap)(t)
		alice.ActSetGasTipCap(gasTipCap)(t)
	}
	alice.ActMakeTx(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)

	// Tx3: SetTokenRatio (Alice is operator)
	newTokenRatioValue := new(big.Int).SetUint64(newTokenRatio)
	_, err = gpoContract.SetTokenRatio(gpoOperator, newTokenRatioValue)
	require.NoError(t, err)
	seqEngine.ActL2IncludeTx(dp.Addresses.Alice)(t)

	// Tx4: Bob normal tx
	bob.ActResetTxOpts(t)
	if isLimb {
		// Limb needs high gas limit and gas price for L1 cost in intrinsic gas
		bob.ActSetTxGasLimit(90_000_000_000)(t) // ~25% of block limit (4 txs in block)
		gasTipCap := big.NewInt(3000 * params.GWei)
		gasFeeCap := big.NewInt(3000 * params.GWei)
		bob.ActSetGasFeeCap(gasFeeCap)(t)
		bob.ActSetGasTipCap(gasTipCap)(t)
	}
	bob.ActMakeTx(t)
	seqEngine.ActL2IncludeTx(dp.Addresses.Bob)(t)

	sequencer.ActL2EndBlock(t)

	// Record balances after
	aliceBalAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Alice, nil)
	require.NoError(t, err)
	bobBalAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), dp.Addresses.Bob, nil)
	require.NoError(t, err)

	// Record vault balances after
	baseFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), baseFeeVault, nil)
	require.NoError(t, err)
	seqFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), seqFeeVault, nil)
	require.NoError(t, err)
	l1FeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), l1FeeVault, nil)
	require.NoError(t, err)
	opFeeVaultAfter, err := seqEngine.EthClient().BalanceAt(t.Ctx(), opFeeVault, nil)
	require.NoError(t, err)

	// Verify fee params and tokenRatio are updated
	if isLimb {
		// For Limb, verify Scalar is updated (Bedrock params)
		updatedScalar, err := gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		expectedScalar := new(big.Int).Mul(oldScalar, big.NewInt(6))
		expectedScalar.Div(expectedScalar, big.NewInt(5)) // 1.2x
		require.Equal(t, expectedScalar.String(), updatedScalar.String(), "Scalar should be updated")
		gt.Logf("[%s] L2 Scalar updated: %s → %s", forkName, oldScalar.String(), updatedScalar.String())
	} else {
		// For Arsia, verify baseFeeScalar is updated (Ecotone params)
		updatedBaseFeeScalar, err := gpoContract.BaseFeeScalar(&bind.CallOpts{})
		require.NoError(t, err)
		expectedBaseFeeScalar := oldBaseFeeScalar * 2
		require.Equal(t, expectedBaseFeeScalar, updatedBaseFeeScalar, "baseFeeScalar should be updated")
		gt.Logf("[%s] L2 baseFeeScalar updated: %d → %d", forkName, oldBaseFeeScalar, updatedBaseFeeScalar)
	}

	updatedRatio, err := gpoContract.TokenRatio(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, newTokenRatio, updatedRatio.Uint64(), "tokenRatio should be updated")
	gt.Logf("[%s] TokenRatio updated: %d → %d", forkName, initialTokenRatio, updatedRatio.Uint64())

	// Get block and receipts
	latestBlock, err := seqEngine.EthClient().BlockByNumber(t.Ctx(), nil)
	require.NoError(t, err)
	txs := latestBlock.Transactions()
	// Block should have: deposit tx + Tx2 + Tx3 + Tx4 = 4 txs
	require.GreaterOrEqual(t, len(txs), 4, "block should have at least 4 transactions")

	receipt2, err := seqEngine.EthClient().TransactionReceipt(t.Ctx(), txs[1].Hash())
	require.NoError(t, err)
	receipt3, err := seqEngine.EthClient().TransactionReceipt(t.Ctx(), txs[2].Hash())
	require.NoError(t, err)
	receipt4, err := seqEngine.EthClient().TransactionReceipt(t.Ctx(), txs[3].Hash())
	require.NoError(t, err)

	require.Equal(t, types.ReceiptStatusSuccessful, receipt2.Status, "Tx2 should succeed")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt3.Status, "Tx3 should succeed")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt4.Status, "Tx4 should succeed")

	gt.Logf("[%s] Tx2 (Alice normal) L1Fee: %s", forkName, receipt2.L1Fee.String())
	gt.Logf("[%s] Tx3 (SetTokenRatio) L1Fee: %s", forkName, receipt3.L1Fee.String())
	gt.Logf("[%s] Tx4 (Bob normal) L1Fee: %s", forkName, receipt4.L1Fee.String())

	// ========== Verify Tx2 + Tx3 (Alice): fee params and ratio behavior ==========
	aliceTotalDeduction := new(big.Int).Sub(aliceBalBefore, aliceBalAfter)

	// Calculate expected fees based on fork type
	var tx2ExpectedFee, tx3ExpectedFee *big.Int

	if isLimb {
		// Limb: L1 cost embedded in gas, no separate L1Fee or OpFee
		tx2L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt2.GasUsed), receipt2.EffectiveGasPrice)
		tx2ExpectedFee = tx2L2Fee

		tx3L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt3.GasUsed), receipt3.EffectiveGasPrice)
		tx3ExpectedFee = tx3L2Fee
	} else {
		// Arsia: L1Fee and OpFee separately deducted
		tx2L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt2.GasUsed), receipt2.EffectiveGasPrice)
		tx2OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
		require.NoError(t, err)
		tx2ExpectedFee = new(big.Int).Add(tx2L2Fee, receipt2.L1Fee)
		tx2ExpectedFee.Add(tx2ExpectedFee, tx2OpFee)

		tx3L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt3.GasUsed), receipt3.EffectiveGasPrice)
		tx3OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt3.GasUsed))
		require.NoError(t, err)
		tx3ExpectedFee = new(big.Int).Add(tx3L2Fee, receipt3.L1Fee)
		tx3ExpectedFee.Add(tx3ExpectedFee, tx3OpFee)
	}

	aliceExpectedTotal := new(big.Int).Add(tx2ExpectedFee, tx3ExpectedFee)

	gt.Logf("[%s] Alice expected total (from receipts): %s", forkName, aliceExpectedTotal.String())
	gt.Logf("[%s] Alice actual total deduction:         %s", forkName, aliceTotalDeduction.String())
	gt.Logf("[%s] Tx2 L1Fee (display): %s, Tx3 L1Fee (display): %s", forkName, receipt2.L1Fee.String(), receipt3.L1Fee.String())

	aliceDeductionMatchesReceipt := aliceExpectedTotal.Cmp(aliceTotalDeduction) == 0

	// Both Limb and Arsia: Tx2/Tx3 use NEW fee params + OLD ratio for both deduction and receipt
	gt.Logf("[%s] Alice deduction matches receipt: %v (expected: true)", forkName, aliceDeductionMatchesReceipt)
	require.True(t, aliceDeductionMatchesReceipt,
		"[%s] Alice deduction should match receipt: both use new fee params + old ratio. "+
			"Actual=%s, Receipt-based=%s", forkName, aliceTotalDeduction.String(), aliceExpectedTotal.String())

	// ========== Verify Tx4 (Bob): should use new fee + new ratio ==========
	bobDeduction := new(big.Int).Sub(bobBalBefore, bobBalAfter)

	var tx4ExpectedFee *big.Int
	if isLimb {
		// Limb: L1 cost embedded in gas, no separate L1Fee or OpFee
		tx4L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt4.GasUsed), receipt4.EffectiveGasPrice)
		tx4ExpectedFee = tx4L2Fee
	} else {
		// Arsia: L1Fee and OpFee separately deducted
		tx4L2Fee := new(big.Int).Mul(new(big.Int).SetUint64(receipt4.GasUsed), receipt4.EffectiveGasPrice)
		tx4OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt4.GasUsed))
		require.NoError(t, err)
		tx4ExpectedFee = new(big.Int).Add(tx4L2Fee, receipt4.L1Fee)
		tx4ExpectedFee.Add(tx4ExpectedFee, tx4OpFee)
	}

	gt.Logf("[%s] Bob expected (from receipt): %s", forkName, tx4ExpectedFee.String())
	gt.Logf("[%s] Bob actual deduction:        %s", forkName, bobDeduction.String())

	// Both Limb and Arsia: Tx4 uses new fee + new ratio, deduction should match receipt
	require.Equal(t, tx4ExpectedFee.Cmp(bobDeduction), 0,
		"[%s] Tx4 (Bob) deduction should match receipt: both use new fee params + new ratio. "+
			"Expected=%s, Actual=%s", forkName, tx4ExpectedFee.String(), bobDeduction.String())

	// ========== Verify Vault Balances ==========
	// Calculate vault increases (independent data source from blockchain)
	baseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultAfter, baseFeeVaultBefore)
	seqFeeVaultIncrease := new(big.Int).Sub(seqFeeVaultAfter, seqFeeVaultBefore)
	l1FeeVaultIncrease := new(big.Int).Sub(l1FeeVaultAfter, l1FeeVaultBefore)
	opFeeVaultIncrease := new(big.Int).Sub(opFeeVaultAfter, opFeeVaultBefore)

	if isLimb {
		// ========== Limb Mode Vault Verification ==========
		// In Limb: L1 cost and OpFee are embedded in gasUsed (through l1Gas inflation)
		// Vaults: BaseFeeVault + SeqFeeVault receive all fees, L1FeeVault and OperatorFeeVault receive nothing

		gt.Logf("[%s] Vault verification:", forkName)
		gt.Logf("[%s]   L1FeeVault increase: %s (expected: 0)", forkName, l1FeeVaultIncrease.String())
		gt.Logf("[%s]   OperatorFeeVault increase: %s (expected: 0)", forkName, opFeeVaultIncrease.String())

		// Verify L1FeeVault didn't receive anything
		require.Equal(t, "0", l1FeeVaultIncrease.String(),
			"[Limb] L1FeeVault should not receive anything (L1 cost embedded in gas)")

		// Verify OperatorFeeVault didn't receive anything
		require.Equal(t, "0", opFeeVaultIncrease.String(),
			"[Limb] OperatorFeeVault should not receive anything (OpFee embedded in gas)")

		// CRITICAL VERIFICATION: Total vault increase (from blockchain) == Total user deduction (from blockchain)
		// This uses two independent data sources to prevent false positives
		totalUserDeduction := new(big.Int).Add(aliceTotalDeduction, bobDeduction)
		totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
		gt.Logf("[%s]   Total UserDeduction (blockchain): %s", forkName, totalUserDeduction.String())
		gt.Logf("[%s]   Total VaultIncrease (blockchain): %s", forkName, totalVaultIncrease.String())
		require.Equal(t, totalUserDeduction.String(), totalVaultIncrease.String(),
			"[Limb] Total vault increase should equal total user deductions")

		gt.Logf("[Limb]  Vault verification passed: L1FeeVault=0, OperatorFeeVault=0, Total vaults = User deductions")

		// ========== Manually Calculate and Verify L1 Cost for Transactions ==========
		// Get current fee params (after all updates - new params)
		currentOverhead, err := gpoContract.Overhead(&bind.CallOpts{})
		require.NoError(t, err)
		currentScalar, err := gpoContract.Scalar(&bind.CallOpts{})
		require.NoError(t, err)
		currentL1BaseFee, err := gpoContract.L1BaseFee(&bind.CallOpts{})
		require.NoError(t, err)

		// Verify Tx2 L1Fee (uses new fee params + old tokenRatio = 1e6)
		tx2, _, err := seqEngine.EthClient().TransactionByHash(t.Ctx(), receipt2.TxHash)
		require.NoError(t, err)
		tx2RLP, err := tx2.MarshalBinary()
		require.NoError(t, err)

		var tx2Zeroes, tx2Ones uint64
		for _, b := range tx2RLP {
			if b == 0 {
				tx2Zeroes++
			} else {
				tx2Ones++
			}
		}
		tx2RollupDataGas := tx2Zeroes*4 + tx2Ones*16

		tx2GasWithOverhead := new(big.Int).SetUint64(tx2RollupDataGas)
		tx2GasWithOverhead.Add(tx2GasWithOverhead, currentOverhead)
		tx2ExpectedL1Cost := new(big.Int).Mul(tx2GasWithOverhead, currentL1BaseFee)
		tx2ExpectedL1Cost.Mul(tx2ExpectedL1Cost, currentScalar)
		tx2ExpectedL1Cost.Mul(tx2ExpectedL1Cost, big.NewInt(int64(initialTokenRatio))) // Old ratio
		tx2ExpectedL1Cost.Div(tx2ExpectedL1Cost, big.NewInt(1e6))

		require.Equal(t, tx2ExpectedL1Cost.String(), receipt2.L1Fee.String(),
			"[Limb] Tx2 receipt.L1Fee should match calculated L1 cost (new fee params + old tokenRatio)")

		// Verify Tx4 L1Fee (uses new fee params + new tokenRatio = 1.5e6)
		tx4, _, err := seqEngine.EthClient().TransactionByHash(t.Ctx(), receipt4.TxHash)
		require.NoError(t, err)
		tx4RLP, err := tx4.MarshalBinary()
		require.NoError(t, err)

		var tx4Zeroes, tx4Ones uint64
		for _, b := range tx4RLP {
			if b == 0 {
				tx4Zeroes++
			} else {
				tx4Ones++
			}
		}
		tx4RollupDataGas := tx4Zeroes*4 + tx4Ones*16

		tx4GasWithOverhead := new(big.Int).SetUint64(tx4RollupDataGas)
		tx4GasWithOverhead.Add(tx4GasWithOverhead, currentOverhead)
		tx4ExpectedL1Cost := new(big.Int).Mul(tx4GasWithOverhead, currentL1BaseFee)
		tx4ExpectedL1Cost.Mul(tx4ExpectedL1Cost, currentScalar)
		tx4ExpectedL1Cost.Mul(tx4ExpectedL1Cost, big.NewInt(int64(newTokenRatio))) // New ratio
		tx4ExpectedL1Cost.Div(tx4ExpectedL1Cost, big.NewInt(1e6))

		require.Equal(t, tx4ExpectedL1Cost.String(), receipt4.L1Fee.String(),
			"[Limb] Tx4 receipt.L1Fee should match calculated L1 cost (new fee params + new tokenRatio)")

		// Verify Tx4 L1Fee is roughly 1.5x Tx2 L1Fee (fee params same, but tokenRatio 1.5x)
		if receipt2.L1Fee.Sign() > 0 && receipt4.L1Fee.Sign() > 0 {
			ratio := new(big.Float).Quo(
				new(big.Float).SetInt(receipt4.L1Fee),
				new(big.Float).SetInt(receipt2.L1Fee),
			)
			gt.Logf("[Limb] Tx4/Tx2 L1Fee ratio: %s (expected ~1.5 due to tokenRatio increase)", ratio.String())
		}

	} else {
		// ========== Arsia Mode Vault Verification ==========
		// In Arsia: L1Fee and OpFee are separately deducted
		// Vaults: BaseFeeVault + SeqFeeVault receive L2 fee, L1FeeVault receives L1Fee, OperatorFeeVault receives OpFee

		// Verify Tx4/Tx2 L1Fee ratio is roughly 1.5x (fee params same, but tokenRatio 1.5x)
		if receipt2.L1Fee.Sign() > 0 && receipt4.L1Fee.Sign() > 0 {
			ratio := new(big.Float).Quo(
				new(big.Float).SetInt(receipt4.L1Fee),
				new(big.Float).SetInt(receipt2.L1Fee),
			)
			gt.Logf("[Arsia] Tx4/Tx2 L1Fee ratio: %s (expected ~1.5 due to tokenRatio increase)", ratio.String())
		}

		gt.Logf("[%s] Vault verification:", forkName)

		// Calculate expected L1Fees and OpFees from receipts
		totalL1Fee := new(big.Int).Add(receipt2.L1Fee, receipt3.L1Fee)
		totalL1Fee.Add(totalL1Fee, receipt4.L1Fee)

		tx2OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt2.GasUsed))
		require.NoError(t, err)
		tx3OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt3.GasUsed))
		require.NoError(t, err)
		tx4OpFee, err := gpoContract.GetOperatorFee(&bind.CallOpts{}, new(big.Int).SetUint64(receipt4.GasUsed))
		require.NoError(t, err)
		totalOpFee := new(big.Int).Add(tx2OpFee, tx3OpFee)
		totalOpFee.Add(totalOpFee, tx4OpFee)

		// Verify L1FeeVault received the sum of all L1Fees
		gt.Logf("[%s]   L1FeeVault: expected=%s, actual=%s", forkName, totalL1Fee.String(), l1FeeVaultIncrease.String())
		require.Equal(t, totalL1Fee.String(), l1FeeVaultIncrease.String(),
			"[Arsia] L1FeeVault should receive sum of all L1Fees")

		// Verify OperatorFeeVault received the sum of all OpFees
		gt.Logf("[%s]   OperatorFeeVault: expected=%s, actual=%s", forkName, totalOpFee.String(), opFeeVaultIncrease.String())
		require.Equal(t, totalOpFee.String(), opFeeVaultIncrease.String(),
			"[Arsia] OperatorFeeVault should receive sum of all OpFees")

		// CRITICAL VERIFICATION: Total vault increase (from blockchain) == Total user deduction (from blockchain)
		// This uses two independent data sources to prevent false positives
		totalUserDeduction := new(big.Int).Add(aliceTotalDeduction, bobDeduction)
		totalVaultIncrease := new(big.Int).Add(baseFeeVaultIncrease, seqFeeVaultIncrease)
		totalVaultIncrease.Add(totalVaultIncrease, l1FeeVaultIncrease)
		totalVaultIncrease.Add(totalVaultIncrease, opFeeVaultIncrease)

		gt.Logf("[%s]   Total UserDeduction (blockchain): %s", forkName, totalUserDeduction.String())
		gt.Logf("[%s]   Total VaultIncrease (blockchain): %s", forkName, totalVaultIncrease.String())
		require.Equal(t, totalUserDeduction.String(), totalVaultIncrease.String(),
			"[Arsia] Total vault increase should equal total user deductions")
	}

	gt.Logf("[%s]  Scenario C test passed!", forkName)
}
