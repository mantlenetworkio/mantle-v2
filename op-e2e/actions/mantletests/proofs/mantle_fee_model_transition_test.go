package proofs

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	mantlebindings "github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_MantleFeeModelTransition(gt *testing.T) {

	run := func(gt *testing.T, testCfg *helpers.TestCfg[any]) {
		t := actionsHelpers.NewDefaultTesting(gt)

		deployConfigOverrides := func(dp *genesis.DeployConfig) {
			// Fork configuration: Limb at genesis, Arsia at block 15
			dp.ActivateMantleForkAtGenesis(forks.MantleLimb)
			dp.ActivateMantleForkAtOffset(forks.MantleArsia, 15)

			// L2 genesis block parameters
			// dp.L2GenesisBlockGasLimit = 200_000_000_000
			dp.L2GenesisBlockBaseFeePerGas = (*hexutil.Big)(big.NewInt(1 * 1e9)) // 1 gwei

			// EIP-1559 parameters
			dp.EIP1559Denominator = 50
			dp.EIP1559Elasticity = 4

			// Mantle-specific: Token Ratio
			dp.GasPriceOracleTokenRatio = 2

			// Operator fee parameters (Arsia only)
			dp.GasPriceOracleOperatorFeeScalar = 345e6   // 345 (DECIMALS=6)
			dp.GasPriceOracleOperatorFeeConstant = 123e4 // 1,230,000 wei

			// MinBaseFee (Limb: fixed 0.02 gwei, Arsia: dynamic, min 10 gwei)
			dp.MinBaseFee = 10 * 1e9 // 10 gwei

			// L1 fee scalars (Arsia uses Ecotone/Fjord formula)
			dp.GasPriceOracleBaseFeeScalar = 1368
			dp.GasPriceOracleBlobBaseFeeScalar = 810949

			// Enable MNT as gas token
			dp.UseCustomGasToken = true
			dp.GasPayingTokenName = "MNT"
			dp.GasPayingTokenSymbol = "MNT"
			dp.NativeAssetLiquidityAmount = (*hexutil.Big)(new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e18)))
		}

		// Create test environment
		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), deployConfigOverrides)

		gpo, err := mantlebindings.NewGasPriceOracleCaller(predeploys.GasPriceOracleAddr, env.Engine.EthClient())
		require.NoError(t, err)

		// Phase 0: Setup TokenRatio

		t.Log("=== Phase 0: Setting up TokenRatio ===")

		// Get GPO owner and create transactor
		gpoOwnerAddr, err := gpo.Owner(nil)
		require.NoError(t, err)

		var ownerKey *ecdsa.PrivateKey
		switch gpoOwnerAddr {
		case env.Dp.Addresses.Alice:
			ownerKey = env.Dp.Secrets.Alice
		case env.Dp.Addresses.Bob:
			ownerKey = env.Dp.Secrets.Bob
		case env.Dp.Addresses.Deployer:
			ownerKey = env.Dp.Secrets.Deployer
		case env.Dp.Addresses.SysCfgOwner:
			ownerKey = env.Dp.Secrets.SysCfgOwner
		default:
			t.Fatalf("Unknown GPO owner: %s", gpoOwnerAddr.String())
		}

		gpoOwnerTransactor, err := bind.NewKeyedTransactorWithChainID(ownerKey, env.Sd.L2Cfg.Config.ChainID)
		require.NoError(t, err)
		gpoOwnerTransactor.GasLimit = 500_000

		gpoContract, err := mantlebindings.NewGasPriceOracle(predeploys.GasPriceOracleAddr, env.Engine.EthClient())
		require.NoError(t, err)

		// Set operator to Alice
		_, err = gpoContract.SetOperator(gpoOwnerTransactor, env.Dp.Addresses.Alice)
		require.NoError(t, err)
		env.Sequencer.ActL2StartBlock(t)
		env.Engine.ActL2IncludeTx(gpoOwnerAddr)(t)
		env.Sequencer.ActL2EndBlock(t)

		// Set TokenRatio by operator
		gpoOperator, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Alice, env.Sd.L2Cfg.Config.ChainID)
		require.NoError(t, err)
		gpoOperator.GasLimit = 500_000

		tokenRatioValue := new(big.Int).SetUint64(env.Dp.DeployConfig.GasPriceOracleTokenRatio)
		_, err = gpoContract.SetTokenRatio(gpoOperator, tokenRatioValue)
		require.NoError(t, err)
		env.Sequencer.ActL2StartBlock(t)
		env.Engine.ActL2IncludeTx(env.Dp.Addresses.Alice)(t)
		env.Sequencer.ActL2EndBlock(t)

		actualTokenRatio, err := gpo.TokenRatio(nil)
		require.NoError(t, err)
		require.Equal(t, tokenRatioValue.String(), actualTokenRatio.String())

		// Phase 1: Test Limb Fee Model

		t.Log("=== Phase 1: Testing Limb fee model ===")

		isArsia, err := gpo.IsArsia(nil)
		require.NoError(t, err)
		require.False(t, isArsia, "Arsia should not be active at genesis")

		// Helper: Get account balance at specific block
		balanceAt := func(a common.Address, blockNumber *big.Int) *big.Int {
			bal, err := env.Engine.EthClient().BalanceAt(t.Ctx(), a, blockNumber)
			require.NoError(t, err)
			return bal
		}

		// Helper: Get all fee vault balances
		getCurrentBalances := func(blockNumberU64 uint64) (alice, l1FeeVault, baseFeeVault, sequencerFeeVault, operatorFeeVault *big.Int) {
			blockNumber := new(big.Int).SetUint64(blockNumberU64)
			alice = balanceAt(env.Alice.Address(), blockNumber)
			l1FeeVault = balanceAt(predeploys.L1FeeVaultAddr, blockNumber)
			baseFeeVault = balanceAt(predeploys.BaseFeeVaultAddr, blockNumber)
			sequencerFeeVault = balanceAt(predeploys.SequencerFeeVaultAddr, blockNumber)
			operatorFeeVault = balanceAt(predeploys.OperatorFeeVaultAddr, blockNumber)
			return
		}

		// Helper: Build L2 block with single transaction
		buildL2BlockWithSingleTx := func() (eth.L2BlockRef, *types.Receipt) {
			env.Sequencer.ActL2StartBlock(t)

			nonce := env.Alice.L2.PendingNonce(t)
			tx := types.MustSignNewTx(env.Alice.L2.Secret(), env.Alice.L2.Signer(), &types.DynamicFeeTx{
				ChainID:   env.Sd.L2Cfg.Config.ChainID,
				Nonce:     nonce,
				To:        &env.Dp.Addresses.Bob,
				Value:     big.NewInt(0),
				Gas:       50000,
				GasFeeCap: big.NewInt(100 * 1e9),
				GasTipCap: big.NewInt(2 * 1e9),
				Data:      []byte{},
			})

			err := env.Engine.EthClient().SendTransaction(t.Ctx(), tx)
			require.NoError(t, err)

			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			unsafeL2Head := env.Sequencer.ActL2EndBlock(t)

			receipt, err := env.Engine.EthClient().TransactionReceipt(t.Ctx(), tx.Hash())
			require.NoError(t, err)

			return unsafeL2Head, receipt
		}

		// Send transaction and verify Limb fee structure
		limbUnsafeHead := env.Sequencer.SyncStatus().UnsafeL2
		aliceInitialLimb, l1FeeVaultInitialLimb, baseFeeVaultInitialLimb, sequencerFeeVaultInitialLimb, operatorFeeVaultInitialLimb := getCurrentBalances(limbUnsafeHead.Number)

		require.Zero(t, operatorFeeVaultInitialLimb.Sign(), "Operator fee vault should be zero in Limb")

		limbUnsafeHead, limbReceipt := buildL2BlockWithSingleTx()

		aliceFinalLimb, l1FeeVaultFinalLimb, baseFeeVaultFinalLimb, sequencerFeeVaultFinalLimb, operatorFeeVaultFinalLimb := getCurrentBalances(limbUnsafeHead.Number)

		// Calculate fees
		limbL1Fee := new(big.Int).Sub(l1FeeVaultFinalLimb, l1FeeVaultInitialLimb)
		baseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultFinalLimb, baseFeeVaultInitialLimb)
		sequencerFeeVaultIncrease := new(big.Int).Sub(sequencerFeeVaultFinalLimb, sequencerFeeVaultInitialLimb)
		limbL2Fee := new(big.Int).Add(baseFeeVaultIncrease, sequencerFeeVaultIncrease)
		limbOperatorFee := new(big.Int).Sub(operatorFeeVaultFinalLimb, operatorFeeVaultInitialLimb)
		limbTotalFee := new(big.Int).Sub(aliceInitialLimb, aliceFinalLimb)

		require.Zero(t, limbOperatorFee.Sign(), "Limb should have zero operator fee")

		// Verify Limb base fee (fixed at 1 gwei)
		limbBlock, err := env.Engine.EthClient().BlockByNumber(t.Ctx(), new(big.Int).SetUint64(limbUnsafeHead.Number))
		require.NoError(t, err)
		limbBlockBaseFee := limbBlock.BaseFee()
		expectedLimbBaseFee := big.NewInt(1 * 1e9)
		require.Equal(t, expectedLimbBaseFee, limbBlockBaseFee, "Limb base fee should be fixed at 1 gwei")

		// Verify L1 Fee calculation (Bedrock formula)
		// Fee^L1_MNT = Fee^L1_ETH × R
		// Fee^L1_ETH = (Gas^L1_ETH + Overhead) × Scalar × BaseFee^L1_ETH
		require.NotNil(t, limbReceipt.L1GasPrice)
		require.NotNil(t, limbReceipt.L1GasUsed)
		require.NotNil(t, limbReceipt.L1Fee)
		require.NotNil(t, limbReceipt.FeeScalar)

		l1GasCost := new(big.Int).Mul(limbReceipt.L1GasPrice, limbReceipt.L1GasUsed)
		l1FeeFloat := new(big.Float).Mul(new(big.Float).SetInt(l1GasCost), limbReceipt.FeeScalar)
		if limbReceipt.TokenRatio != nil && limbReceipt.TokenRatio.Sign() > 0 {
			l1FeeFloat.Mul(l1FeeFloat, new(big.Float).SetInt(limbReceipt.TokenRatio))
		}
		expectedL1Fee, _ := l1FeeFloat.Int(nil)
		require.Equal(t, expectedL1Fee.String(), limbReceipt.L1Fee.String(),
			"Receipt L1 fee should match formula")

		// IMPORTANT: In Limb, L1 fee is NOT sent to L1FeeVault
		// It's included in gas used calculation and sent to BaseFeeVault
		require.Zero(t, limbL1Fee.Sign(), "Limb L1FeeVault should be zero")

		// Verify L2 Fee calculation: L2 Fee = gasUsed × (baseFee + priorityFee)
		// NOTE: In Limb, L2 Fee includes L1 Fee
		gasUsed := new(big.Int).SetUint64(limbReceipt.GasUsed)
		effectiveGasPrice := limbReceipt.EffectiveGasPrice
		expectedLimbL2Fee := new(big.Int).Mul(gasUsed, effectiveGasPrice)
		require.Equal(t, expectedLimbL2Fee.String(), limbL2Fee.String())

		// Verify total fee = L2 fee (L1 fee is included in L2 fee)
		expectedLimbTotalFee := limbL2Fee
		require.Equal(t, expectedLimbTotalFee.String(), limbTotalFee.String())

		aliceDeducted := new(big.Int).Sub(aliceInitialLimb, aliceFinalLimb)
		require.Equal(t, expectedLimbTotalFee.String(), aliceDeducted.String())

		// Phase 2: Activate Arsia Fork

		t.Log("=== Phase 2: Activating Arsia fork ===")

		arsiaActivationBlock := env.Sequencer.ActBuildL2ToMantleFork(t, forks.MantleArsia)

		isArsiaAfterFork, err := gpo.IsArsia(nil)
		require.NoError(t, err)
		require.True(t, isArsiaAfterFork, "Arsia should be active now")

		arsiaBlock := env.Engine.L2Chain().GetBlockByNumber(arsiaActivationBlock.Number)
		require.Equal(t, 8, len(arsiaBlock.Transactions()),
			"Arsia activation block should have 8 transactions (1 set-L1-info + 7 upgrade txs)")

		// Phase 3: Configure L1 SystemConfig

		t.Log("=== Phase 3: Setting parameters via L1 SystemConfig ===")

		sysCfgContract, err := mantlebindings.NewSystemConfig(env.Sd.RollupCfg.L1SystemConfigAddress, env.Miner.EthClient())
		require.NoError(t, err)

		sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Deployer, env.Sd.RollupCfg.L1ChainID)
		require.NoError(t, err)
		sysCfgOwner.GasLimit = 500_000

		// Set MinBaseFee
		minBaseFee := uint64(10 * 1e9)
		_, err = sysCfgContract.SetMinBaseFee(sysCfgOwner, minBaseFee)
		require.NoError(t, err)
		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTx(env.Dp.Addresses.Deployer)(t)
		env.Miner.ActL1EndBlock(t)

		// Set Ecotone Gas Config (L1 base fee scalar and blob base fee scalar)
		newBaseFeeScalar := uint32(1368)
		newBlobBaseFeeScalar := uint32(810949)
		_, err = sysCfgContract.SetGasConfigArsia(sysCfgOwner, newBaseFeeScalar, newBlobBaseFeeScalar)
		require.NoError(t, err)
		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTx(env.Dp.Addresses.Deployer)(t)
		env.Miner.ActL1EndBlock(t)

		// Set Operator Fee parameters
		newOperatorFeeScalar := uint32(345e6)
		newOperatorFeeConstant := uint64(123e4)
		_, err = sysCfgContract.SetOperatorFeeScalars(sysCfgOwner, newOperatorFeeScalar, newOperatorFeeConstant)
		require.NoError(t, err)
		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTx(env.Dp.Addresses.Deployer)(t)
		env.Miner.ActL1EndBlock(t)

		// L2 syncs L1 head but doesn't adopt new parameters yet
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActBuildToL1HeadExcl(t)

		// Phase 4: Test Arsia Fee Model

		t.Log("=== Phase 4: Testing Arsia fee model ===")

		arsiaUnsafeHead := env.Sequencer.SyncStatus().UnsafeL2
		aliceInitialArsia, l1FeeVaultInitialArsia, baseFeeVaultInitialArsia, sequencerFeeVaultInitialArsia, operatorFeeVaultInitialArsia := getCurrentBalances(arsiaUnsafeHead.Number)

		arsiaUnsafeHead, arsiaReceipt := buildL2BlockWithSingleTx()

		// Verify receipt contains operator fee parameters
		require.NotNil(t, arsiaReceipt.OperatorFeeScalar)
		require.NotNil(t, arsiaReceipt.OperatorFeeConstant)
		require.Equal(t, uint32(345e6), uint32(*arsiaReceipt.OperatorFeeScalar))
		require.Equal(t, uint64(123e4), *arsiaReceipt.OperatorFeeConstant)

		aliceFinalArsia, l1FeeVaultFinalArsia, baseFeeVaultFinalArsia, sequencerFeeVaultFinalArsia, operatorFeeVaultFinalArsia := getCurrentBalances(arsiaUnsafeHead.Number)

		// Calculate fees
		arsiaL1Fee := new(big.Int).Sub(l1FeeVaultFinalArsia, l1FeeVaultInitialArsia)
		arsiaBaseFeeVaultIncrease := new(big.Int).Sub(baseFeeVaultFinalArsia, baseFeeVaultInitialArsia)
		arsiaSequencerFeeVaultIncrease := new(big.Int).Sub(sequencerFeeVaultFinalArsia, sequencerFeeVaultInitialArsia)
		arsiaL2Fee := new(big.Int).Add(arsiaBaseFeeVaultIncrease, arsiaSequencerFeeVaultIncrease)
		arsiaOperatorFee := new(big.Int).Sub(operatorFeeVaultFinalArsia, operatorFeeVaultInitialArsia)
		arsiaTotalFee := new(big.Int).Sub(aliceInitialArsia, aliceFinalArsia)

		// Verify Operator Fee (Jovian formula): (gasUsed × scalar × 100) + constant
		expectedOperatorFee := uint64(724_500_001_230_000)
		require.Equal(t, expectedOperatorFee, arsiaOperatorFee.Uint64())

		gpoOperatorFee, err := gpo.GetOperatorFee(nil, new(big.Int).SetUint64(arsiaReceipt.GasUsed))
		require.NoError(t, err)
		require.Equal(t, expectedOperatorFee, gpoOperatorFee.Uint64())

		// Verify Arsia base fee >= minBaseFee (10 gwei)
		arsiaBlock, err = env.Engine.EthClient().BlockByNumber(t.Ctx(), new(big.Int).SetUint64(arsiaUnsafeHead.Number))
		require.NoError(t, err)
		arsiaBlockBaseFee := arsiaBlock.BaseFee()
		expectedMinBaseFee := big.NewInt(10 * 1e9)
		require.True(t, arsiaBlockBaseFee.Cmp(expectedMinBaseFee) >= 0)

		// Verify L1 Fee calculation (Ecotone/Fjord formula)
		// Fee^L1_MNT = Fee^L1_ETH × R
		// Fee^L1_ETH = Size × (Scalar^L1_basefee × BaseFee^L1_ETH × 16 + Scalar^L1_blobfee × BlobFee^L1_ETH) / 1e12
		require.NotNil(t, arsiaReceipt.L1BaseFeeScalar)
		require.NotNil(t, arsiaReceipt.L1BlobBaseFeeScalar)
		require.NotNil(t, arsiaReceipt.L1BlobBaseFee)
		require.NotNil(t, arsiaReceipt.L1Fee)

		require.Equal(t, uint64(1368), *arsiaReceipt.L1BaseFeeScalar)
		require.Equal(t, uint64(810949), *arsiaReceipt.L1BlobBaseFeeScalar)

		require.Equal(t, arsiaReceipt.L1Fee.String(), arsiaL1Fee.String(),
			"L1 Fee vault should receive the L1 fee from receipt")

		// Verify L2 Fee calculation: L2 Fee = gasUsed × (baseFee + priorityFee)
		arsiaGasUsed := new(big.Int).SetUint64(arsiaReceipt.GasUsed)
		arsiaEffectiveGasPrice := arsiaReceipt.EffectiveGasPrice
		expectedArsiaL2Fee := new(big.Int).Mul(arsiaGasUsed, arsiaEffectiveGasPrice)
		require.Equal(t, expectedArsiaL2Fee.String(), arsiaL2Fee.String())

		// Verify total fee = L1 fee + L2 fee + Operator fee
		expectedArsiaTotalFee := new(big.Int).Add(arsiaL1Fee, arsiaL2Fee)
		expectedArsiaTotalFee.Add(expectedArsiaTotalFee, arsiaOperatorFee)
		require.Equal(t, expectedArsiaTotalFee.String(), arsiaTotalFee.String())

		arsiaAliceDeducted := new(big.Int).Sub(aliceInitialArsia, aliceFinalArsia)
		require.Equal(t, expectedArsiaTotalFee.String(), arsiaAliceDeducted.String())

		// Arsia total fee should be much higher than Limb
		require.True(t, arsiaTotalFee.Cmp(limbTotalFee) > 0)

		// Phase 5: Compare Limb vs Arsia

		t.Log("=== Phase 5: Comparing Limb vs Arsia ===")

		l2FeeIncrease := new(big.Int).Sub(arsiaL2Fee, limbL2Fee)
		operatorFeeIncrease := arsiaOperatorFee
		totalFeeIncrease := new(big.Int).Sub(arsiaTotalFee, limbTotalFee)

		require.True(t, l2FeeIncrease.Sign() > 0, "L2 fee should increase")
		require.True(t, operatorFeeIncrease.Cmp(l2FeeIncrease) > 0, "Operator fee is the main increase")

		expectedTotalIncrease := new(big.Int).Add(l2FeeIncrease, operatorFeeIncrease)
		require.InDelta(t, expectedTotalIncrease.Uint64(), totalFeeIncrease.Uint64(),
			float64(expectedTotalIncrease.Uint64())*0.1)

		t.Log("Mantle fee model transition test completed successfully!")
	}

	matrix := helpers.NewMatrix[any]()
	matrix.AddDefaultTestCases(nil, helpers.NewForkMatrix(helpers.MantleLimb), run)
	matrix.Run(gt)
}
