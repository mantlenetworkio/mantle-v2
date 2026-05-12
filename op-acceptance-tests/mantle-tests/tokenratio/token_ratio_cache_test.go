package tokenratio

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/apis"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	txib "github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
)

// A/B/C + critical double-setRatio scenario in black-box style.
// Besides receipt-level fee checks, each scenario verifies real vault balance growth.
func TestTokenRatioAndFeeParamsCharging(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia), "test requires Arsia fork active")

	l2Client := sys.L2EL.Escape().EthClient()
	gpoRO := txib.NewGasPriceOracle(
		txib.WithClient(l2Client),
		txib.WithTo(predeploys.GasPriceOracleAddr),
		txib.WithTest(t),
	)

	operatorAddr, err := contractio.Read(gpoRO.Operator(), ctx)
	require.NoError(err)
	operatorPriv := findPrivKeyByAddress(t, sys, operatorAddr)
	require.NotNil(operatorPriv, "cannot find private key for GPO operator %s", operatorAddr.Hex())
	sender := dsl.NewKey(t, operatorPriv).User(sys.L2EL)
	receiver := sys.FunderL2.NewFundedEOA(eth.ZeroWei)
	sys.FunderL2.FundAtLeast(sender, eth.OneEther)

	tokenRatioStart, err := contractio.Read(gpoRO.TokenRatio(), ctx)
	require.NoError(err)
	operatorFee := dsl.NewOperatorFee(t, sys.L2Chain, sys.L1EL)
	operatorFee.CheckCompatibility()

	t.Cleanup(func() {
		_ = sendSetTokenRatioTx(t, gpoRO, sender, tokenRatioStart)
		operatorFee.RestoreOriginalConfig()
	})

	t.Run("ScenarioA_FeeParamsOnly", func(t devtest.T) {
		newScalar := uint32(123)
		newConstant := uint64(456)
		operatorFee.SetOperatorFee(newScalar, newConstant)
		operatorFee.WaitForL2Sync(newScalar, newConstant)

		receipts := submitSameSenderTxs(t, gpoRO, sender, receiver.Address(),
			txAction{kind: actionTransfer, name: "normal-1", value: big.NewInt(1)},
			txAction{kind: actionTransfer, name: "normal-2", value: big.NewInt(2)},
		)
		assertSameBlock(t, receipts)

		for _, rcpt := range receipts {
			t.Require().NotNil(rcpt.OperatorFeeScalar)
			t.Require().NotNil(rcpt.OperatorFeeConstant)
			t.Require().Equal(newScalar, uint32(*rcpt.OperatorFeeScalar))
			t.Require().Equal(newConstant, *rcpt.OperatorFeeConstant)
		}
		assertVaultGrowthMatchesBlockFees(t, l2Client, receipts[0].BlockHash)
	})

	t.Run("ScenarioB_TokenRatioOnly", func(t devtest.T) {
		oldRatio, err := contractio.Read(gpoRO.TokenRatio(), ctx)
		t.Require().NoError(err)
		newRatio := new(big.Int).Add(new(big.Int).Set(oldRatio), big.NewInt(333_333))

		receipts := submitSameSenderTxs(t, gpoRO, sender, receiver.Address(),
			txAction{kind: actionTransfer, name: "tx1-normal", value: big.NewInt(3)},
			txAction{kind: actionSetTokenRatio, name: "tx2-setRatio", ratio: newRatio},
			txAction{kind: actionTransfer, name: "tx3-normal", value: big.NewInt(4)},
		)
		assertSameBlock(t, receipts)

		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[0], oldRatio)
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[1], oldRatio)
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[2], newRatio)
		assertVaultGrowthMatchesBlockFees(t, l2Client, receipts[0].BlockHash)
	})

	t.Run("ScenarioC_FeeParamsAndTokenRatio", func(t devtest.T) {
		newScalar := uint32(321)
		newConstant := uint64(654)
		operatorFee.SetOperatorFee(newScalar, newConstant)
		operatorFee.WaitForL2Sync(newScalar, newConstant)

		oldRatio, err := contractio.Read(gpoRO.TokenRatio(), ctx)
		t.Require().NoError(err)
		newRatio := new(big.Int).Add(new(big.Int).Set(oldRatio), big.NewInt(444_444))

		receipts := submitSameSenderTxs(t, gpoRO, sender, receiver.Address(),
			txAction{kind: actionTransfer, name: "tx2-normal", value: big.NewInt(5)},
			txAction{kind: actionSetTokenRatio, name: "tx3-setRatio", ratio: newRatio},
			txAction{kind: actionTransfer, name: "tx4-normal", value: big.NewInt(6)},
		)
		assertSameBlock(t, receipts)

		for _, rcpt := range receipts {
			t.Require().NotNil(rcpt.OperatorFeeScalar)
			t.Require().NotNil(rcpt.OperatorFeeConstant)
			t.Require().Equal(newScalar, uint32(*rcpt.OperatorFeeScalar))
			t.Require().Equal(newConstant, *rcpt.OperatorFeeConstant)
		}
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[0], oldRatio)
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[1], oldRatio)
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[2], newRatio)
		assertVaultGrowthMatchesBlockFees(t, l2Client, receipts[0].BlockHash)
	})

	t.Run("ScenarioD_DoubleSetTokenRatio", func(t devtest.T) {
		baseRatio, err := contractio.Read(gpoRO.TokenRatio(), ctx)
		t.Require().NoError(err)
		r1 := new(big.Int).Add(new(big.Int).Set(baseRatio), big.NewInt(111_111))
		r2 := new(big.Int).Add(new(big.Int).Set(baseRatio), big.NewInt(222_222))

		receipts := submitSameSenderTxs(t, gpoRO, sender, receiver.Address(),
			txAction{kind: actionSetTokenRatio, name: "tx1-setR1", ratio: r1},
			txAction{kind: actionSetTokenRatio, name: "tx2-setR2", ratio: r2},
			txAction{kind: actionTransfer, name: "tx3-normal", value: big.NewInt(7)},
		)
		assertSameBlock(t, receipts)

		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[0], baseRatio)
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[1], r1)
		assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[2], r2)
		assertVaultGrowthMatchesBlockFees(t, l2Client, receipts[0].BlockHash)
	})

	t.Run("ScenarioE_SameBlockFeeParamAndTokenRatioTransition", func(t devtest.T) {
		const maxAttempts = 4
		matched := false
		for i := 0; i < maxAttempts; i++ {
			newScalar := uint32(700 + i)
			newConstant := uint64(900 + i)
			operatorFee.SetOperatorFee(newScalar, newConstant)

			oldRatio, err := contractio.Read(gpoRO.TokenRatio(), ctx)
			t.Require().NoError(err)
			newRatio := new(big.Int).Add(new(big.Int).Set(oldRatio), big.NewInt(int64(555_000+i)))

			receipts := submitSameSenderTxs(t, gpoRO, sender, receiver.Address(),
				txAction{kind: actionSetTokenRatio, name: "tx1-setRatio", ratio: newRatio},
				txAction{kind: actionTransfer, name: "tx2-normal", value: big.NewInt(int64(8 + i))},
			)
			assertSameBlock(t, receipts)

			blockScalar, blockConstant := readOperatorFeeParamsAtBlock(t, l2Client, receipts[0].BlockHash)
			info, err := l2Client.InfoByHash(ctx, receipts[0].BlockHash)
			t.Require().NoError(err)
			parentScalar, parentConstant := readOperatorFeeParamsAtBlock(t, l2Client, info.ParentHash())

			isTransition := blockScalar == newScalar && blockConstant == newConstant &&
				!(parentScalar == newScalar && parentConstant == newConstant)
			if !isTransition {
				continue
			}

			for _, rcpt := range receipts {
				t.Require().NotNil(rcpt.OperatorFeeScalar)
				t.Require().NotNil(rcpt.OperatorFeeConstant)
				t.Require().Equal(newScalar, uint32(*rcpt.OperatorFeeScalar))
				t.Require().Equal(newConstant, *rcpt.OperatorFeeConstant)
			}

			assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[0], oldRatio)
			assertReceiptL1FeeUsesRatio(t, l2Client, gpoRO, receipts[1], newRatio)
			assertVaultGrowthMatchesBlockFees(t, l2Client, receipts[0].BlockHash)
			matched = true
			break
		}
		t.Require().True(matched, "failed to hit a same-block fee-param-transition + tokenRatio-update case within retry budget")
	})
}

type actionKind int

const (
	actionTransfer actionKind = iota
	actionSetTokenRatio
)

type txAction struct {
	kind  actionKind
	name  string
	value *big.Int
	ratio *big.Int
}

func submitSameSenderTxs(t devtest.T, gpo *txib.GasPriceOracle, sender *dsl.EOA, transferTo common.Address, actions ...txAction) []*types.Receipt {
	require := t.Require()
	ctx := t.Ctx()
	nonce := sender.PendingNonce()

	planned := make([]*txplan.PlannedTx, 0, len(actions))
	for i, a := range actions {
		opts := []txplan.Option{
			sender.Plan(),
			txplan.WithNonce(nonce + uint64(i)),
			txplan.WithGasLimit(200_000),
		}
		switch a.kind {
		case actionTransfer:
			opts = append(opts, txplan.WithTo(&transferTo), txplan.WithEth(zeroIfNil(a.value)))
		case actionSetTokenRatio:
			planOpt, err := contractio.Plan(gpo.SetTokenRatio(a.ratio))
			require.NoError(err)
			opts = append(opts, planOpt)
		default:
			require.FailNow("unknown tx action kind")
		}
		p := txplan.NewPlannedTx(opts...)
		planned = append(planned, p)
	}

	for _, p := range planned {
		_, err := p.Submitted.Eval(ctx)
		require.NoError(err)
	}

	receipts := make([]*types.Receipt, 0, len(planned))
	for _, p := range planned {
		rcpt, err := p.Included.Eval(ctx)
		require.NoError(err)
		require.Equal(types.ReceiptStatusSuccessful, rcpt.Status)
		receipts = append(receipts, rcpt)
	}
	return receipts
}

func assertSameBlock(t devtest.T, receipts []*types.Receipt) {
	require := t.Require()
	require.NotEmpty(receipts)
	b := receipts[0].BlockNumber.Uint64()
	for i := 1; i < len(receipts); i++ {
		require.Equalf(b, receipts[i].BlockNumber.Uint64(), "tx[%d] is not in same block", i)
	}
}

func assertReceiptL1FeeUsesRatio(t devtest.T, client apis.EthClient, gpo *txib.GasPriceOracle, receipt *types.Receipt, ratio *big.Int) {
	require := t.Require()
	require.NotNil(receipt)
	require.NotNil(receipt.L1Fee)

	unsigned, err := unsignedTxBytesByReceipt(t, client, receipt)
	require.NoError(err)
	baseL1Fee, err := dsl.ReadGasPriceOracleL1FeeAt(t.Ctx(), client, gpo, unsigned, receipt.BlockHash)
	require.NoError(err)

	expected := new(big.Int).Mul(new(big.Int).Set(baseL1Fee), ratio)
	require.Equal(0, receipt.L1Fee.Cmp(expected), "receipt L1Fee mismatch with expected token ratio")
}

func unsignedTxBytesByReceipt(t devtest.T, client apis.EthClient, receipt *types.Receipt) ([]byte, error) {
	signed, err := dsl.FindSignedTransactionFromReceipt(t.Ctx(), client, receipt)
	if err != nil {
		return nil, err
	}
	unsigned, err := dsl.CreateUnsignedTransactionFromSigned(signed)
	if err != nil {
		return nil, err
	}
	return unsigned.MarshalBinary()
}

func assertVaultGrowthMatchesBlockFees(t devtest.T, client apis.EthClient, blockHash common.Hash) {
	require := t.Require()
	ctx := t.Ctx()

	info, txs, err := client.InfoAndTxsByHash(ctx, blockHash)
	require.NoError(err)
	require.Greater(info.NumberU64(), uint64(0), "block number must be > 0")

	currNum := new(big.Int).SetUint64(info.NumberU64())
	prevNum := new(big.Int).SetUint64(info.NumberU64() - 1)

	baseBefore := balanceAt(t, client, predeploys.BaseFeeVaultAddr, prevNum)
	baseAfter := balanceAt(t, client, predeploys.BaseFeeVaultAddr, currNum)
	l1Before := balanceAt(t, client, predeploys.L1FeeVaultAddr, prevNum)
	l1After := balanceAt(t, client, predeploys.L1FeeVaultAddr, currNum)
	seqBefore := balanceAt(t, client, predeploys.SequencerFeeVaultAddr, prevNum)
	seqAfter := balanceAt(t, client, predeploys.SequencerFeeVaultAddr, currNum)
	opBefore := balanceAt(t, client, predeploys.OperatorFeeVaultAddr, prevNum)
	opAfter := balanceAt(t, client, predeploys.OperatorFeeVaultAddr, currNum)

	actualBase := new(big.Int).Sub(baseAfter, baseBefore)
	actualL1 := new(big.Int).Sub(l1After, l1Before)
	actualSeq := new(big.Int).Sub(seqAfter, seqBefore)
	actualOp := new(big.Int).Sub(opAfter, opBefore)

	expectedBase := big.NewInt(0)
	expectedSeq := big.NewInt(0)
	expectedOp := big.NewInt(0)

	for _, tx := range txs {
		if tx.Type() == types.DepositTxType {
			continue
		}
		rcpt, err := client.TransactionReceipt(ctx, tx.Hash())
		require.NoError(err)

		gasUsed := big.NewInt(int64(rcpt.GasUsed))
		baseFee := new(big.Int).Mul(info.BaseFee(), gasUsed)
		totalGasFee := new(big.Int).Mul(rcpt.EffectiveGasPrice, gasUsed)
		priorityFee := new(big.Int).Sub(totalGasFee, baseFee)

		expectedBase.Add(expectedBase, baseFee)
		expectedSeq.Add(expectedSeq, priorityFee)
		expectedOp.Add(expectedOp, arsiaOperatorFeeFromReceipt(rcpt))
	}

	require.Equal(0, expectedBase.Cmp(actualBase), "BaseFeeVault growth mismatch")
	// L1 fee accounting includes system-level effects in this environment; enforce non-negative growth only.
	require.Truef(actualL1.Sign() >= 0, "L1FeeVault growth must be non-negative: %s", actualL1.String())
	delta := new(big.Int).Sub(l1After, l1Before)
	exp, err := sumBlockL1Fee(ctx, client, info.NumberU64())
	require.NoError(err)
	require.Equal(t, 0, delta.Cmp(exp), "L1FeeVault delta must equal block summed L1Fee")
	require.Equal(0, expectedSeq.Cmp(actualSeq), "SequencerFeeVault growth mismatch")
	require.Equal(0, expectedOp.Cmp(actualOp), "OperatorFeeVault growth mismatch")
}

func sumBlockL1Fee(ctx context.Context, c apis.EthClient, bn uint64) (*big.Int, error) {
	_, txs, err := c.InfoAndTxsByNumber(ctx, bn)
	if err != nil {
		return nil, err
	}
	total := new(big.Int)
	for _, tx := range txs {
		// Skip deposit transactions
		if tx.Type() == types.DepositTxType {
			continue
		}
		rcpt, err := c.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, err
		}
		if rcpt.L1Fee != nil {
			total.Add(total, rcpt.L1Fee)
		}
	}
	return total, nil
}

func balanceAt(t devtest.T, client apis.EthClient, addr common.Address, blockNum *big.Int) *big.Int {
	bal, err := client.BalanceAt(t.Ctx(), addr, blockNum)
	t.Require().NoError(err)
	return bal
}

func arsiaOperatorFeeFromReceipt(rcpt *types.Receipt) *big.Int {
	if rcpt.OperatorFeeScalar == nil || rcpt.OperatorFeeConstant == nil {
		return big.NewInt(0)
	}
	fee := new(big.Int).Mul(big.NewInt(int64(rcpt.GasUsed)), big.NewInt(int64(*rcpt.OperatorFeeScalar)))
	fee.Mul(fee, big.NewInt(100))
	fee.Add(fee, new(big.Int).SetUint64(*rcpt.OperatorFeeConstant))
	return fee
}

func readOperatorFeeParamsAtBlock(t devtest.T, client apis.EthClient, blockHash common.Hash) (uint32, uint64) {
	l1Block := txib.NewL1Block(
		txib.WithClient(client),
		txib.WithTo(predeploys.L1BlockAddr),
		txib.WithTest(t),
	)
	overrideBlockOpt := func(ptx *txplan.PlannedTx) {
		ptx.AgainstBlock.Fn(func(ctx context.Context) (eth.BlockInfo, error) {
			return client.InfoByHash(ctx, blockHash)
		})
	}
	scalar, err := contractio.Read(l1Block.OperatorFeeScalar(), t.Ctx(), overrideBlockOpt)
	t.Require().NoError(err)
	constant, err := contractio.Read(l1Block.OperatorFeeConstant(), t.Ctx(), overrideBlockOpt)
	t.Require().NoError(err)
	return scalar, constant
}

func sendSetTokenRatioTx(t devtest.T, gpo *txib.GasPriceOracle, from *dsl.EOA, ratio *big.Int) error {
	_, err := contractio.Write(gpo.SetTokenRatio(ratio), t.Ctx(), from.Plan(), txplan.WithGasLimit(200_000))
	return err
}

func zeroIfNil(v *big.Int) *big.Int {
	if v == nil {
		return big.NewInt(0)
	}
	return v
}

func findPrivKeyByAddress(t devtest.T, sys *presets.MantleMinimal, addr common.Address) *ecdsa.PrivateKey {
	keys := sys.L2Chain.Escape().Keys()
	chainID := sys.L2Chain.ChainID().ToBig()

	for role := devkeys.ChainOperatorRole(0); role <= devkeys.ChainFeesRecipientRole; role++ {
		priv := keys.Secret(role.Key(chainID))
		if dsl.NewKey(t, priv).Address() == addr {
			return priv
		}
	}
	for i := uint64(0); i < 40; i++ {
		priv := keys.Secret(devkeys.UserKey(i))
		if dsl.NewKey(t, priv).Address() == addr {
			return priv
		}
	}
	return nil
}
