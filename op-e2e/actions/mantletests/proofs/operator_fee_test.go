package proofs

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	mantlebindings "github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_OperatorFeeConsistency(gt *testing.T) {
	type testCase int64

	const (
		NormalTx testCase = iota
		DepositTx
		StateRefund
		NotEnoughFundsInBatchMissingOpFee
		IsthmusTransitionBlock
	)

	testStorageUpdateContractAddress := common.HexToAddress("0xffffffff")
	// contract TestSetter {
	//   uint x;
	//   function set(uint _x) public { x = _x; }
	// }
	// The deployed bytecode below is from the contract above
	testStorageUpdateContractCode := common.FromHex("0x6080604052348015600e575f80fd5b50600436106026575f3560e01c806360fe47b114602a575b5f80fd5b60406004803603810190603c9190607d565b6042565b005b805f8190555050565b5f80fd5b5f819050919050565b605f81604f565b81146068575f80fd5b50565b5f813590506077816058565b92915050565b5f60208284031215608f57608e604b565b5b5f609a84828501606b565b9150509291505056fea26469706673582212201712a1e6e9c5e2ba1f8f7403f5d6e00090c6fa2b70c632beea4be8009331bd2064736f6c63430008190033")

	runJovianDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[testCase]) {
		t := actionsHelpers.NewDefaultTesting(gt)
		deployConfigOverrides := func(dp *genesis.DeployConfig) {}

		var testOperatorFeeScalar uint32
		var testOperatorFeeConstant uint64
		if testCfg.Custom == NotEnoughFundsInBatchMissingOpFee {
			testOperatorFeeScalar = 0
			testOperatorFeeConstant = 0xffff
		} else {
			testOperatorFeeScalar = 100e6
			testOperatorFeeConstant = 500
		}

		if testCfg.Custom == StateRefund {
			testCfg.Allocs = actionsHelpers.DefaultAlloc
			testCfg.Allocs.L2Alloc = make(map[common.Address]types.Account)
			testCfg.Allocs.L2Alloc[testStorageUpdateContractAddress] = types.Account{
				Code:    testStorageUpdateContractCode,
				Nonce:   1,
				Balance: new(big.Int),
			}
		}

		if testCfg.Custom == IsthmusTransitionBlock {
			deployConfigOverrides = func(dp *genesis.DeployConfig) {
				dp.L1PragueTimeOffset = ptr(hexutil.Uint64(0))
				dp.L2GenesisMantleArsiaTimeOffset = ptr(hexutil.Uint64(13))
			}
		}

		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), deployConfigOverrides)

		balanceAt := func(a common.Address) *big.Int {
			t.Helper()
			bal, err := env.Engine.EthClient().BalanceAt(t.Ctx(), a, nil)
			require.NoError(t, err)
			return bal
		}

		getCurrentBalances := func() (alice *big.Int, l1FeeVault *big.Int, baseFeeVault *big.Int, sequencerFeeVault *big.Int, operatorFeeVault *big.Int) {
			alice = balanceAt(env.Alice.Address())
			l1FeeVault = balanceAt(predeploys.L1FeeVaultAddr)
			baseFeeVault = balanceAt(predeploys.BaseFeeVaultAddr)
			sequencerFeeVault = balanceAt(predeploys.SequencerFeeVaultAddr)
			operatorFeeVault = balanceAt(predeploys.OperatorFeeVaultAddr)

			return alice, l1FeeVault, baseFeeVault, sequencerFeeVault, operatorFeeVault
		}

		setStorageInUpdateContractTo := func(value byte) {
			t.Helper()
			input := common.RightPadBytes(common.FromHex("0x60fe47b1"), 36)
			input[35] = value
			env.Sequencer.ActL2StartBlock(t)
			env.Alice.L2.ActResetTxOpts(t)
			env.Alice.L2.ActSetTxToAddr(&testStorageUpdateContractAddress)(t)
			env.Alice.L2.ActSetTxCalldata(input)(t)
			env.Alice.L2.ActMakeTx(t)
			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
			r := env.Alice.L2.LastTxReceipt(t)
			require.Equal(t, types.ReceiptStatusSuccessful, r.Status, "tx unsuccessful")
		}

		sysCfgContract, err := mantlebindings.NewSystemConfig(env.Sd.RollupCfg.L1SystemConfigAddress, env.Miner.EthClient())
		require.NoError(t, err)

		sysCfgOwner, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Deployer, env.Sd.RollupCfg.L1ChainID)
		require.NoError(t, err)

		// Update the operator fee parameters
		_, err = sysCfgContract.SetOperatorFeeScalars(sysCfgOwner, testOperatorFeeScalar, testOperatorFeeConstant)
		require.NoError(t, err)

		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTx(env.Dp.Addresses.Deployer)(t)
		l1BlockWithOpFee := env.Miner.ActL1EndBlock(t)
		t.Logf("Set operator fee in L1 block: number=%d, hash=%s", l1BlockWithOpFee.Number(), l1BlockWithOpFee.Hash().Hex())

		// sequence L2 blocks, and submit with new batcher
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActBuildToL1Head(t)
		env.BatchAndMine(t)

		env.Sequencer.ActL1HeadSignal(t)

		var aliceInitialBalance *big.Int
		var baseFeeVaultInitialBalance *big.Int
		var l1FeeVaultInitialBalance *big.Int
		var sequencerFeeVaultInitialBalance *big.Int
		var operatorFeeVaultInitialBalance *big.Int

		var receipt *types.Receipt

		switch testCfg.Custom {
		case NormalTx, IsthmusTransitionBlock:
			aliceInitialBalance, l1FeeVaultInitialBalance, baseFeeVaultInitialBalance, sequencerFeeVaultInitialBalance, operatorFeeVaultInitialBalance = getCurrentBalances()

			require.Equal(t, operatorFeeVaultInitialBalance.Sign(), 0)

			if testCfg.Custom == IsthmusTransitionBlock {
				// For Isthmus transition block test:
				// Check if Isthmus activation is also Jovian activation
				// If so, we need to:
				// 1. Build empty Jovian/Isthmus activation block (no user txs allowed)
				// 2. Build next block with user transaction (which should have operator fee)

				// Get Isthmus activation time
				isthmusTime := *env.Sd.RollupCfg.IsthmusTime
				isJovianActivation := env.Sd.RollupCfg.IsJovianActivationBlock(isthmusTime)

				if isJovianActivation {
					// Build empty blocks until we reach Jovian/Isthmus activation block
					for env.Sequencer.L2Unsafe().Time < isthmusTime {
						env.Sequencer.ActL2EmptyBlock(t)
					}

					// Verify we're at the activation block
					require.Equal(t, isthmusTime, env.Sequencer.L2Unsafe().Time)
					require.True(t, env.Sd.RollupCfg.IsIsthmusActivationBlock(env.Sequencer.L2Unsafe().Time))
					require.True(t, env.Sd.RollupCfg.IsJovianActivationBlock(env.Sequencer.L2Unsafe().Time))

					// Batch and mine the activation block to L1
					env.BatchAndMine(t)

					// Build a new L1 block to ensure L2 can reference the latest L1 config
					env.Miner.ActEmptyBlock(t)

					// Sync L1 head signal to ensure L2 gets latest L1 config
					env.Sequencer.ActL1HeadSignal(t)
					env.Sequencer.ActBuildToL1Head(t)

					// Now build the next block with a user transaction
					// This block should have operator fee applied

					// Check L1 SystemConfig to see if operator fee is set
					sysCfgContract, err := mantlebindings.NewSystemConfig(env.Sd.RollupCfg.L1SystemConfigAddress, env.Miner.EthClient())
					require.NoError(t, err)
					OperatorFeeConstant, err := sysCfgContract.OperatorFeeConstant(nil)
					require.NoError(t, err)
					OperatorFeeScalar, err := sysCfgContract.OperatorFeeScalar(nil)
					require.NoError(t, err)
					t.Logf("L1 SystemConfig OperatorFeeScalar before building block 14: %d", OperatorFeeScalar)
					t.Logf("L1 SystemConfig OperatorFeeConstant before building block 14: %d", OperatorFeeConstant)

					// Check current L1 head
					l1Head := env.Miner.L1Chain().CurrentBlock()
					t.Logf("Current L1 head: number=%d, hash=%s", l1Head.Number, l1Head.Hash().Hex())

					// Check what L1 origin the next L2 block will use
					t.Logf("Current L2 unsafe head: number=%d, time=%d, L1Origin=%s (number=%d)",
						env.Sequencer.L2Unsafe().Number,
						env.Sequencer.L2Unsafe().Time,
						env.Sequencer.L2Unsafe().L1Origin.Hash.Hex(),
						env.Sequencer.L2Unsafe().L1Origin.Number)

					// Check SystemConfig at different L1 blocks
					for i := uint64(0); i <= l1Head.Number.Uint64(); i++ {
						l1Block := env.Miner.L1Chain().GetBlockByNumber(i)
						if l1Block != nil {
							opFeeScalar, err := sysCfgContract.OperatorFeeScalar(&bind.CallOpts{BlockNumber: big.NewInt(int64(i))})
							require.NoError(t, err)
							opFeeConstant, err := sysCfgContract.OperatorFeeConstant(&bind.CallOpts{BlockNumber: big.NewInt(int64(i))})
							require.NoError(t, err)
							t.Logf("L1 block %d: OperatorFeeScalar=%d, OperatorFeeConstant=%d", i, opFeeScalar, opFeeConstant)
						}
					}

					env.Sequencer.ActL2StartBlock(t)
					env.Alice.L2.ActResetTxOpts(t)
					env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)(t)
					env.Alice.L2.ActMakeTx(t)
					env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
					env.Sequencer.ActL2EndBlock(t)
				} else {
					// Not a Jovian activation block, can include user tx in activation block
					// Build empty blocks until we're one block before activation
					for env.Sequencer.L2Unsafe().Time+2 < isthmusTime {
						env.Sequencer.ActL2EmptyBlock(t)
					}

					// Now build the activation block with a user transaction
					env.Sequencer.ActL2StartBlock(t)
					env.Alice.L2.ActResetTxOpts(t)
					env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)(t)
					env.Alice.L2.ActMakeTx(t)
					env.Engine.ActL2IncludeTxIgnoreForcedEmpty(env.Alice.Address())(t)
					env.Sequencer.ActL2EndBlock(t)
					require.True(t, env.Sd.RollupCfg.IsIsthmusActivationBlock(env.Sequencer.L2Unsafe().Time))
				}
			} else {
				// Normal tx case
				env.Sequencer.ActL2StartBlock(t)
				env.Alice.L2.ActResetTxOpts(t)
				env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)(t)
				env.Alice.L2.ActMakeTx(t)
				env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
				env.Sequencer.ActL2EndBlock(t)
			}

		case StateRefund:
			setStorageInUpdateContractTo(1)
			rSet := env.Alice.L2.LastTxReceipt(t)
			require.Equal(t, uint64(43696), rSet.GasUsed)
			aliceInitialBalance, l1FeeVaultInitialBalance, baseFeeVaultInitialBalance, sequencerFeeVaultInitialBalance, operatorFeeVaultInitialBalance = getCurrentBalances()
			setStorageInUpdateContractTo(0)
			rUnset := env.Alice.L2.LastTxReceipt(t)
			// we assert on the exact gas used to show that a refund is happening
			require.Equal(t, uint64(21784), rUnset.GasUsed)

		case DepositTx:
			aliceInitialBalance, l1FeeVaultInitialBalance, baseFeeVaultInitialBalance, sequencerFeeVaultInitialBalance, operatorFeeVaultInitialBalance = getCurrentBalances()

			// Mantle-specific deposit transaction (6 parameters)
			// Get portal address from deploy config
			portalAddr := env.Dp.DeployConfig.OptimismPortalProxy
			t.Logf("OptimismPortal address: %s", portalAddr.Hex())

			// Create Mantle OptimismPortal binding
			mantlePortal, err := mantlebindings.NewOptimismPortal(portalAddr, env.Miner.EthClient())
			require.NoError(t, err, "failed to create Mantle OptimismPortal binding")

			// Check if portal is paused
			isPaused, err := mantlePortal.Paused(&bind.CallOpts{})
			require.NoError(t, err, "failed to check if portal is paused")
			t.Logf("OptimismPortal paused status: %v", isPaused)
			require.False(t, isPaused, "OptimismPortal should not be paused")

			// Check Alice's L1 balance
			aliceL1Balance, err := env.Miner.EthClient().BalanceAt(context.Background(), env.Dp.Addresses.Alice, nil)
			require.NoError(t, err, "failed to get Alice's L1 balance")
			t.Logf("Alice's L1 balance: %s", aliceL1Balance.String())

			// Check SystemConfig ResourceConfig
			systemConfigAddr := env.Sd.RollupCfg.L1SystemConfigAddress
			systemConfig, err := mantlebindings.NewSystemConfig(systemConfigAddr, env.Miner.EthClient())
			require.NoError(t, err, "failed to create SystemConfig binding")
			resourceConfig, err := systemConfig.ResourceConfig(&bind.CallOpts{})
			require.NoError(t, err, "failed to get ResourceConfig")
			t.Logf("ResourceConfig: %+v", resourceConfig)

			// Check L1 MNT address
			l1MNTAddr, err := mantlePortal.L1MNTADDRESS(&bind.CallOpts{})
			require.NoError(t, err, "failed to get L1 MNT address")
			t.Logf("L1 MNT address: %s", l1MNTAddr.Hex())
			require.NotEqual(t, common.Address{}, l1MNTAddr, "L1 MNT address should not be zero")

			// Prepare deposit parameters
			toAddr := env.Dp.Addresses.Bob
			depositGas := uint64(2e6)

			// Create L1 transaction options using Alice's private key
			aliceTxOpts, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Alice, env.Sd.RollupCfg.L1ChainID)
			require.NoError(t, err)
			// Note: msg.value must be >= _ethTxValue to ensure L2 tx has enough ETH
			ethValue := big.NewInt(1)    // 1 wei to send
			aliceTxOpts.Value = ethValue // ETH sent to portal on L1 (msg.value)
			aliceTxOpts.NoSend = true    // Don't auto-send, we'll send manually

			// Execute deposit transaction (Mantle 7-parameter version)
			// Solidity: function depositTransaction(uint256 _ethTxValue, uint256 _mntValue, address _to, uint256 _mntTxValue, uint64 _gasLimit, bool _isCreation, bytes _data) payable
			// Parameters:
			//   _ethTxValue: ETH value to use in the L2 transaction (must be <= msg.value)
			//   _mntValue: MNT to transfer from L1 sender to portal
			//   _to: target address on L2
			//   _mntTxValue: MNT value to send in L2 transaction
			//   _gasLimit: minimum L2 gas limit
			//   _isCreation: whether this is a contract creation
			//   _data: calldata for L2 transaction
			tx, err := mantlePortal.DepositTransaction(
				aliceTxOpts,
				ethValue,      // _ethTxValue: ETH value in L2 transaction (must match msg.value)
				big.NewInt(0), // _mntValue: MNT to transfer from L1 to portal
				toAddr,        // _to: target address on L2
				big.NewInt(0), // _mntTxValue: MNT value to send in L2 transaction
				depositGas,    // _gasLimit: minimum L2 gas limit
				false,         // _isCreation: not a contract creation
				[]byte{},      // _data: empty calldata
			)
			require.NoError(t, err, "failed to create deposit tx")

			// Send the deposit transaction on L1
			err = env.Miner.EthClient().SendTransaction(t.Ctx(), tx)
			require.NoError(t, err, "must send deposit tx")

			// Mine the L1 block containing the deposit
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTx(env.Alice.Address())(t)
			env.Miner.ActL1EndBlock(t)

			// Sync sequencer and build enough blocks to adopt latest L1 origin
			env.Sequencer.ActL1HeadSignal(t)
			env.Sequencer.ActBuildToL1HeadUnsafe(t)

			// Get the deposit receipt on L2
			// We need to manually reconstruct the deposit from the L1 receipt
			l1Receipt, err := env.Miner.EthClient().TransactionReceipt(t.Ctx(), tx.Hash())
			require.NoError(t, err)
			require.Equal(t, types.ReceiptStatusSuccessful, l1Receipt.Status)

			// Reconstruct the L2 deposit transaction
			reconstructedDep, err := derive.UnmarshalDepositLogEvent(l1Receipt.Logs[0])
			require.NoError(t, err, "Could not reconstruct L2 Deposit")
			l2Tx := types.NewTx(reconstructedDep)

			// Get L2 receipt
			receipt, err = env.Engine.EthClient().TransactionReceipt(t.Ctx(), l2Tx.Hash())
			require.NoError(t, err)
			require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
			aliceInitialBalance, l1FeeVaultInitialBalance, baseFeeVaultInitialBalance, sequencerFeeVaultInitialBalance, operatorFeeVaultInitialBalance = getCurrentBalances()

			// regular Deposit, in new L1 block (using direct portal call like above)
			toAddr2 := env.Dp.Addresses.Bob
			depositGas2 := uint64(2e6)
			ethValue2 := big.NewInt(0) // No ETH value for this deposit

			aliceTxOpts2, err := bind.NewKeyedTransactorWithChainID(env.Dp.Secrets.Alice, env.Sd.RollupCfg.L1ChainID)
			require.NoError(t, err)
			aliceTxOpts2.Value = ethValue2 // ETH sent to portal on L1 (msg.value)
			aliceTxOpts2.NoSend = true     // Don't auto-send, we'll send manually

			// Execute deposit transaction (Mantle 7-parameter version)
			tx2, err := mantlePortal.DepositTransaction(
				aliceTxOpts2,
				ethValue2,     // _ethTxValue: ETH value in L2 transaction
				big.NewInt(0), // _mntValue: MNT to transfer from L1 to portal
				toAddr2,       // _to: target address on L2
				big.NewInt(0), // _mntTxValue: MNT value to send in L2 transaction
				depositGas2,   // _gasLimit: minimum L2 gas limit
				false,         // _isCreation: not a contract creation
				[]byte{},      // _data: empty calldata
			)
			require.NoError(t, err, "failed to create second deposit tx")

			// Send the deposit transaction on L1
			err = env.Miner.EthClient().SendTransaction(t.Ctx(), tx2)
			require.NoError(t, err, "must send second deposit tx")

			// Mine the L1 block containing the deposit
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTx(env.Alice.Address())(t)
			env.Miner.ActL1EndBlock(t)

			// sync sequencer build enough blocks to adopt latest L1 origin
			env.Sequencer.ActL1HeadSignal(t)
			env.Sequencer.ActBuildToL1HeadUnsafe(t)

			// Check deposit status and get L2 receipt manually (since we didn't use ActDeposit)
			depositL1Receipt := env.Alice.L1.CheckReceipt(t, true, tx2.Hash())
			require.NotNil(t, depositL1Receipt, "L1 deposit receipt should exist")
			require.Greater(t, len(depositL1Receipt.Logs), 0, "L1 deposit receipt should have logs")

			// Reconstruct the L2 deposit transaction from L1 receipt
			reconstructedDep, err = derive.UnmarshalDepositLogEvent(depositL1Receipt.Logs[0])
			require.NoError(t, err, "Could not reconstruct L2 Deposit")
			l2DepositTx := types.NewTx(reconstructedDep)

			// Get L2 receipt
			receipt = env.Alice.L2.CheckReceipt(t, true, l2DepositTx.Hash())
			require.NotNil(t, receipt, "L2 deposit receipt should exist")
		case NotEnoughFundsInBatchMissingOpFee:
			pkey, err := crypto.GenerateKey()
			require.NoError(t, err)
			address := crypto.PubkeyToAddress(pkey.PublicKey)

			// Send `address` just enough ETH to cover the gas costs of the transaction, sans the operator fee.
			// Since we're just doing a simple call to `Bob`, that should be 21000 gas at the current gas price
			// plus the L1 data fee.

			// Craft a transaction from Alice -> Bob (just to compute L1 cost, not to send.)
			env.Alice.L2.ActResetTxOpts(t)
			env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)(t)
			tx := env.Alice.L2.MakeTransaction(t)

			rlp, err := tx.MarshalBinary()
			require.NoError(t, err, "failed to RLP encode transaction")

			unsafeHeader := env.Engine.L2Chain().CurrentHeader()
			unsafeBlock := env.Engine.L2Chain().GetBlockByHash(unsafeHeader.Hash())
			nextBaseFee := eip1559.CalcBaseFee(
				env.Sd.L2Cfg.Config,
				unsafeHeader,
			)

			l1BlockInfo, err := derive.L1BlockInfoFromBytes(env.Sd.RollupCfg, unsafeHeader.Time, unsafeBlock.Transactions()[0].Data())
			require.NoError(t, err)

			daCost := fjordL1Cost(l1BlockInfo, types.NewRollupCostData(rlp))
			expectedFeePreIsthmus := nextBaseFee.Mul(nextBaseFee, big.NewInt(int64(params.TxGas)))
			expectedFeePreIsthmus.Add(expectedFeePreIsthmus, daCost)

			// Include an L2 tx, from Bob -> mock signer
			env.Bob.L2.ActResetTxOpts(t)
			env.Bob.L2.ActSetTxToAddr(&address)(t)
			env.Bob.L2.ActSetTxValue(expectedFeePreIsthmus)(t)
			env.Bob.L2.ActMakeTx(t)

			env.Sequencer.ActL2StartBlock(t)
			env.Engine.ActL2IncludeTx(env.Bob.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
			env.Bob.L2.ActCheckReceiptStatusOfLastTx(true)(t)

			// Ensure the mock signer received the funds
			require.Equal(t, expectedFeePreIsthmus, balanceAt(address))

			// Buffer the L2 block we just included
			env.Batcher.ActL2BatchBuffer(t)

			aliceInitialBalance, l1FeeVaultInitialBalance, baseFeeVaultInitialBalance, sequencerFeeVaultInitialBalance, operatorFeeVaultInitialBalance = getCurrentBalances()

			// Craft a transaction from Alice -> Bob
			env.Alice.L2.ActResetTxOpts(t)
			env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)(t)
			env.Alice.L2.ActSetTxGasLimit(params.TxGas)(t)
			env.Alice.L2.ActSetGasFeeCap(big.NewInt(1))(t)
			env.Alice.L2.ActSetGasTipCap(big.NewInt(1))(t)
			env.Alice.L2.ActMakeTx(t)

			// Include an L2 tx, from Alice -> Bob
			env.Sequencer.ActL2StartBlock(t)
			env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			env.Sequencer.ActL2EndBlock(t)
			env.Alice.L2.ActCheckReceiptStatusOfLastTx(true)(t)

			// Instruct the batcher to submit a faulty channel, with Alice's tx re-signed by a new private key.
			// This key will have 0 balance.
			env.Batcher.ActL2BatchBuffer(t, actionsHelpers.WithBlockModifier(func(block *types.Block) *types.Block {
				txs := block.Transactions()

				// Skip over any L2 blocks that don't contain user-space txs.
				if len(txs) == 1 {
					return block
				}

				// Re-sign Alice's tx with a random key.
				require.NoError(t, err, "error generating random private key")
				signer := types.LatestSigner(env.Sd.L2Cfg.Config)
				newSignedTx, err := types.SignTx(txs[1], signer, pkey)
				require.NoError(t, err, "error re-signing Alice's transaction")

				// Replace Alice's tx with the re-signed one.
				txs[1] = newSignedTx
				return block
			}))
			env.Batcher.ActL2ChannelClose(t)
			env.Batcher.ActL2BatchSubmit(t)

			// Include the batcher transaction.
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			env.Miner.ActL1EndBlock(t)
		}

		aliceFinalBalance, l1FeeVaultFinalBalance, baseFeeVaultFinalBalance, sequencerFeeVaultFinalBalance, operatorFeeVaultFinalBalance := getCurrentBalances()
		l2UnsafeHead := env.Engine.L2Chain().CurrentHeader()

		if receipt == nil {
			receipt = env.Alice.L2.LastTxReceipt(t)
		}

		// Determine if we should expect operator fee based on:
		// 1. DepositTx: never has operator fee
		// 2. IsthmusTransitionBlock with Jovian activation: tx is AFTER activation, should have operator fee
		// 3. IsthmusTransitionBlock without Jovian activation: tx is IN activation block, no operator fee
		shouldHaveOperatorFee := true
		if testCfg.Custom == DepositTx {
			shouldHaveOperatorFee = false
		} else if testCfg.Custom == IsthmusTransitionBlock {
			// Check if Isthmus activation was also Jovian activation
			// If so, the user tx is in the block AFTER activation and should have operator fee
			isMantleArsia := *env.Sd.RollupCfg.MantleArsiaTime
			isJovianActivation := env.Sd.RollupCfg.IsJovianActivationBlock(isMantleArsia)
			shouldHaveOperatorFee = isJovianActivation
		}

		if !shouldHaveOperatorFee {
			require.Nil(t, receipt.OperatorFeeScalar)
			require.Nil(t, receipt.OperatorFeeConstant)

			// Nothing should has been sent to operator fee vault
			require.Equal(t, operatorFeeVaultInitialBalance, operatorFeeVaultFinalBalance)
		} else if env.Sd.RollupCfg.IsMantleArsia(l2UnsafeHead.Time) {
			// Check that the operator fee was applied
			t.Logf("Receipt OperatorFeeScalar: %v (expected: %v)", *receipt.OperatorFeeScalar, testOperatorFeeScalar)
			t.Logf("Receipt OperatorFeeConstant: %v (expected: %v)", *receipt.OperatorFeeConstant, testOperatorFeeConstant)
			require.Equal(t, testOperatorFeeScalar, uint32(*receipt.OperatorFeeScalar))
			require.Equal(t, testOperatorFeeConstant, *receipt.OperatorFeeConstant)

			// Check that the operator fee sent to the vault is correct
			// Determine which formula to use based on whether Jovian is active
			var expectedOperatorFee *big.Int
			if env.Sd.RollupCfg.IsJovian(l2UnsafeHead.Time) {
				// Jovian formula: (gasUsed * operatorFeeScalar * 100) + operatorFeeConstant
				expectedOperatorFee = new(big.Int).Add(
					new(big.Int).Mul(
						new(big.Int).Mul(
							new(big.Int).SetUint64(receipt.GasUsed),
							new(big.Int).SetUint64(uint64(testOperatorFeeScalar)),
						),
						new(big.Int).SetUint64(100),
					),
					new(big.Int).SetUint64(testOperatorFeeConstant),
				)
			} else {
				// Isthmus formula: (gasUsed * operatorFeeScalar / 1e6) + operatorFeeConstant
				expectedOperatorFee = new(big.Int).Add(
					new(big.Int).Div(
						new(big.Int).Mul(
							new(big.Int).SetUint64(receipt.GasUsed),
							new(big.Int).SetUint64(uint64(testOperatorFeeScalar)),
						),
						new(big.Int).SetUint64(1e6),
					),
					new(big.Int).SetUint64(testOperatorFeeConstant),
				)
			}

			require.Equal(t,
				expectedOperatorFee,
				new(big.Int).Sub(operatorFeeVaultFinalBalance, operatorFeeVaultInitialBalance),
			)
		}

		if testCfg.Custom == DepositTx {
			require.Equal(t, aliceInitialBalance, aliceFinalBalance, "Alice's balance shouldn't have changed")
		} else {
			require.True(t, aliceFinalBalance.Cmp(aliceInitialBalance) < 0, "Alice's balance should decrease")
		}

		// Check that no Ether has been minted or burned
		finalTotalBalance := new(big.Int).Add(
			aliceFinalBalance,
			new(big.Int).Add(
				new(big.Int).Add(
					new(big.Int).Sub(l1FeeVaultFinalBalance, l1FeeVaultInitialBalance),
					new(big.Int).Sub(sequencerFeeVaultFinalBalance, sequencerFeeVaultInitialBalance),
				),
				new(big.Int).Add(
					new(big.Int).Sub(operatorFeeVaultFinalBalance, operatorFeeVaultInitialBalance),
					new(big.Int).Sub(baseFeeVaultFinalBalance, baseFeeVaultInitialBalance),
				),
			),
		)

		require.Equal(t, aliceInitialBalance, finalTotalBalance)

		// The NotEnoughFundsInBatchMissingOpFee case is special, as it submits its own invalid batch.
		if testCfg.Custom != NotEnoughFundsInBatchMissingOpFee {
			env.BatchAndMine(t)
		}
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Engine.L2Chain().CurrentSafeBlock()

		if testCfg.Custom == NotEnoughFundsInBatchMissingOpFee {
			// The unsafe block prior to derivation should be different from the safe block after derivation. The
			// batcher posted the block but with a different transaction, signed by a key that has no balance. This
			// should cause a reorg in the unsafe chain, and the original block should be reduced to deposits only
			// if Isthmus is active.
			require.NotEqual(t, eth.HeaderBlockID(l2SafeHead), eth.HeaderBlockID(l2UnsafeHead), "derivation should not lead to the same block")

			reorgedUnsafe := env.Engine.L2Chain().CurrentHeader()
			require.Equal(t, eth.HeaderBlockID(l2SafeHead), eth.HeaderBlockID(reorgedUnsafe), "reorged unsafe block is the same")

			safeHeadBlock := env.Engine.L2Chain().GetBlockByHash(l2SafeHead.Hash())
			if env.Sd.RollupCfg.IsMantleArsia(l2SafeHead.Time) {
				require.Equal(t, len(safeHeadBlock.Transactions()), 1)

				// Ensure that the logs contain a mention of the block being replaced _due to the signer not having enough
				// balance_.
				require.NotNil(t, env.Logs.FindLog(testlog.NewAttributesContainsFilter("err", "insufficient funds for gas * price + value")))
				require.NotNil(t, env.Logs.FindLog(testlog.NewAttributesContainsFilter("err", "have 1400000021000 want 1400000086535")))
			} else {
				require.Equal(t, len(safeHeadBlock.Transactions()), 2)
			}
		} else {
			require.Equal(t, eth.HeaderBlockID(l2SafeHead), eth.HeaderBlockID(l2UnsafeHead), "derivation leads to the same block")
		}

		//env.RunFaultProofProgramFromGenesis(t, l2SafeHead.Number.Uint64(), testCfg.CheckResult, testCfg.InputParams...)
	}

	matrix := helpers.NewMatrix[testCase]()
	matrix.AddDefaultTestCasesWithName("NormalTx", NormalTx, helpers.NewForkMatrix(helpers.MantleArsia), runJovianDerivationTest)
	matrix.AddDefaultTestCasesWithName("DepositTx", DepositTx, helpers.NewForkMatrix(helpers.MantleArsia), runJovianDerivationTest)
	matrix.AddDefaultTestCasesWithName("StateRefund", StateRefund, helpers.NewForkMatrix(helpers.MantleArsia), runJovianDerivationTest)
	matrix.AddDefaultTestCasesWithName("NotEnoughFundsInBatchMissingOpFee", NotEnoughFundsInBatchMissingOpFee, helpers.NewForkMatrix(helpers.MantleArsia), runJovianDerivationTest)
	matrix.AddDefaultTestCasesWithName("IsthmusTransitionBlock", IsthmusTransitionBlock, helpers.NewForkMatrix(helpers.MantleLimb), runJovianDerivationTest)
	matrix.Run(gt)
}

func fjordL1Cost(l1BlockInfo *derive.L1BlockInfo, rollupCostData types.RollupCostData) *big.Int {
	costFunc := types.NewL1CostFuncFjord(
		l1BlockInfo.BaseFee,
		l1BlockInfo.BlobBaseFee,
		new(big.Int).SetUint64(uint64(l1BlockInfo.BaseFeeScalar)),
		new(big.Int).SetUint64(uint64(l1BlockInfo.BlobBaseFeeScalar)))

	fee, _ := costFunc(rollupCostData)
	return fee
}

func ptr[T any](v T) *T {
	return &v
}
