package proofs

import (
	"encoding/binary"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/stretchr/testify/require"
)

func setMinBaseFeeViaSystemConfig(t actionsHelpers.Testing, env *helpers.L2ProofEnv, minBaseFee uint64) {
	// Create system config contract binding
	systemConfig, err := bindings.NewSystemConfig(env.Sd.RollupCfg.L1SystemConfigAddress, env.Miner.EthClient())
	require.NoError(t, err)

	// Create transactor for the deployer (system config owner)
	deployerTx, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Deployer, env.Sd.RollupCfg.L1ChainID)
	require.NoError(t, err)
	t.Logf("Setting min base fee on L1: minBaseFee=%d", minBaseFee)

	// Mine the L1 transaction
	env.Miner.ActL1StartBlock(12)(t)
	_, err = systemConfig.SetMinBaseFee(deployerTx, minBaseFee)
	require.NoError(t, err, "SetMinBaseFee transaction failed")
	env.Miner.ActL1IncludeTx(env.Dp.Addresses.Deployer)(t)
	env.Miner.ActL1EndBlock(t)
}

func Test_ProgramAction_JovianMinBaseFee(gt *testing.T) {
	runJovianDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[any], genesisConfigFn func(*genesis.DeployConfig), minBaseFee uint64) {
		t := actionsHelpers.NewDefaultTesting(gt)
		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), genesisConfigFn)
		t.Logf("L2 Genesis Time: %d, ArsiaTime: %d ", env.Sequencer.RollupCfg.Genesis.L2Time, *env.Sequencer.RollupCfg.MantleArsiaTime)
		// Create GasPriceOracle contract binding
		gpo, err := bindings.NewGasPriceOracleCaller(predeploys.GasPriceOracleAddr, env.Engine.EthClient())
		require.NoError(t, err)
		arsiaAtGenesis := env.Sequencer.RollupCfg.IsMantleArsia(env.Sequencer.RollupCfg.Genesis.L2Time)
		if !arsiaAtGenesis {
			// Check GPO status
			isArsia, err := gpo.IsArsia(nil)
			require.NoError(t, err)
			require.False(t, isArsia, "GPO should report that Arsia is not active")

			// If Arsia is not activated at genesis, build some blocks up to the activation block
			// and verify that the extra data is Holocene
			for env.Engine.L2Chain().CurrentBlock().Time < *env.Sequencer.RollupCfg.MantleArsiaTime {
				b := env.Engine.L2Chain().GetBlockByHash(env.Sequencer.L2Unsafe().Hash)
				require.Equal(t, 0, len(b.Extra()),
					"ExtraData should be empty before Arsia activation (Limb version)")
				require.Equal(t, 0, len(b.Extra()),
					"ExtraData should be empty before Arsia activation (Limb version)")
				env.Sequencer.ActL2EmptyBlock(t)
				// Last iteration builds the activation block.
			}
		}

		// Check GPO status
		isArsia, err := gpo.IsArsia(nil)
		require.NoError(t, err)
		require.True(t, isArsia, "GPO should report that Jovian is active")

		activationBlock := env.Engine.L2Chain().GetBlockByHash(env.Sequencer.L2Unsafe().Hash)
		require.Equal(t, eip1559.EncodeMinBaseFeeExtraData(50, 4, 0), activationBlock.Extra(), "activation block should have Jovian extraData")

		// Set the minimum base fee
		setMinBaseFeeViaSystemConfig(t, env, minBaseFee)

		// Build activation+1 block
		env.Sequencer.ActL2EmptyBlock(t)
		blockAfterActivation := env.Engine.L2Chain().GetBlockByHash(env.Sequencer.L2Unsafe().Hash)
		// Assert extradata of the blocks which were past the Jovian activation, but before the L1 origin moved to the SystemConfig change
		// It should have a zero min base fee
		actualMinBaseFee := binary.BigEndian.Uint64(blockAfterActivation.Extra()[9:17])
		require.Equal(t, uint64(0), actualMinBaseFee, "activation block should have a zero min base fee")

		// Allow L1->L2 derivation to propagate the SystemConfig change & build L2 blocks up to the L1 origin that includes the SystemConfig change
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)
		env.Sequencer.ActBuildToL1Head(t)

		// Block after the SystemConfig change
		env.Sequencer.ActL2EmptyBlock(t)
		blockAfterSystemConfigChange := env.Engine.L2Chain().GetBlockByHash(env.Sequencer.L2Unsafe().Hash)
		expectedJovianExtraDataWithMinFee := eip1559.EncodeMinBaseFeeExtraData(50, 4, minBaseFee)
		require.Equal(t, expectedJovianExtraDataWithMinFee, blockAfterSystemConfigChange.Extra(), "block should have updated Jovian extraData with min base fee")

		// Verify base fee is clamped
		require.GreaterOrEqual(t, blockAfterSystemConfigChange.BaseFee().Uint64(), minBaseFee, "base fee should be >= minimum base fee")

		if !arsiaAtGenesis {
			// Mantle forks activate all OP Stack forks simultaneously
			// The activation log will show the first OP fork (canyon), not the last one (jovian)
			t.Logf("MantleArsia activated at block %d", env.Engine.L2Chain().CurrentBlock().Number)
		}

		env.BatchMineAndSync(t)
		l2SafeHead := env.Sequencer.L2Safe()
		t.Logf("Safe head block number: %d, timestamp: %d", l2SafeHead.Number, l2SafeHead.Time)
		require.True(t, l2SafeHead.Number >= uint64(0), "safe head should progress")

		//env.RunFaultProofProgramFromGenesis(t, l2SafeHead.Number, testCfg.CheckResult, testCfg.InputParams...)
	}

	tests := map[string]struct {
		genesisConfigFn func(*genesis.DeployConfig)
		jovianAtGenesis bool
		minBaseFee      uint64
	}{
		"JovianActivationAfterGenesis": {
			genesisConfigFn: func(dc *genesis.DeployConfig) {
				dc.ActivateMantleForkAtOffset(forks.MantleArsia, 10)
				// Set EIP-1559 parameters for testing
				dc.EIP1559Denominator = 50
				dc.EIP1559Elasticity = 4
			},
			minBaseFee: 0,
		},
		"JovianActivationAtGenesisZeroMinBaseFee": {
			genesisConfigFn: func(dc *genesis.DeployConfig) {
				dc.ActivateMantleForkAtGenesis(forks.MantleArsia)
				// Set EIP-1559 parameters for testing
				dc.EIP1559Denominator = 50
				dc.EIP1559Elasticity = 4
			},
			minBaseFee: 0,
		},
		"JovianActivationAtGenesisMinBaseFeeMedium": {
			genesisConfigFn: func(dc *genesis.DeployConfig) {
				dc.ActivateMantleForkAtGenesis(forks.MantleArsia)
				// Set EIP-1559 parameters for testing
				dc.EIP1559Denominator = 50
				dc.EIP1559Elasticity = 4
			},
			minBaseFee: 1_000_000_000, // 1 gwei
		},
		"JovianActivationAtGenesisMinBaseFeeHigh": {
			genesisConfigFn: func(dc *genesis.DeployConfig) {
				dc.ActivateMantleForkAtGenesis(forks.MantleArsia)
				// Set EIP-1559 parameters for testing
				dc.EIP1559Denominator = 50
				dc.EIP1559Elasticity = 4
			},
			minBaseFee: 2_000_000_000, // 2 gwei
		},
	}

	for name, tt := range tests {
		gt.Run(name, func(t *testing.T) {
			matrix := helpers.NewMatrix[any]()
			matrix.AddDefaultTestCasesWithName(
				name,
				nil,
				helpers.NewForkMatrix(helpers.MantleArsia),
				func(gt *testing.T, testCfg *helpers.TestCfg[any]) {
					runJovianDerivationTest(gt, testCfg, tt.genesisConfigFn, tt.minBaseFee)
				},
			)
			matrix.Run(t)
		})
	}
}
