/*
Package deposit documents Mantle L1->L2 deposit behaviors used by the tests.

Portal deposits
  - OptimismPortal.depositTransaction encodes msg.value and _ethTxValue into the
    deposit payload. msg.value controls the L2 mint, _ethTxValue controls the L2
    transfer executed after minting.
  - op-geth core/state_transition.go:
  - mintBVMETH: mints BVM_ETH on L2 from msg.value.
  - transferBVMETH: executes ethTxValue by transferring from msg.From to msg.To.
    If msg.value is zero and msg.From has no BVM_ETH, the L2 deposit fails.

MNT deposits
  - _mntValue mints on L2 (via ERC20 transfer on L1).
  - mntTxValue is an L2 native-value transfer from msg.From to msg.To; it fails
    if msg.From lacks L2 MNT, even when _mntValue is zero.

Bridge deposits
  - L1StandardBridge handles ERC20/MNT bridging and expects msg.value=0 (non-zero
    msg.value should revert). ETH bridge deposits depend on msg.value to affect L2
    balances.

Other notes
  - Portal does not mint/bridge ERC20 directly; tests use the bridge path.
  - waitForL2Deposit reconstructs the L2 deposit tx from the L1
    TransactionDeposited log and waits for its L2 receipt. The WithoutMsgValue
    ETH case asserts a successful receipt with unchanged balance when to=from.
*/
package deposit

import (
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl/contract"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/errutil"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var bvmETHAddr = common.HexToAddress("0xdEAddEaDdeadDEadDEADDEAddEADDEAddead1111")

type depositPath string

const (
	depositByPortal depositPath = "portal"
	depositByBridge depositPath = "bridge"
)

type depositAsset string

const (
	assetETH   depositAsset = "eth"
	assetMNT   depositAsset = "mnt"
	assetERC20 depositAsset = "erc20"
)

type msgValueMode string

const (
	msgValueZero    msgValueMode = "zero"
	msgValueWithout msgValueMode = "without-msgvalue"
	msgValueWith    msgValueMode = "with-msgvalue"
)

func runDepositCase(gt *testing.T, path depositPath, asset depositAsset, mode msgValueMode) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	sys.L1Network.WaitForOnline()

	l1User := sys.FunderL1.NewFundedEOA(eth.OneTenthEther)
	l2User := l1User.AsEL(sys.L2EL)
	sys.FunderL2.FundAtLeast(l2User, eth.OneTenthEther)

	depositAmount := eth.GWei(1_000_000)
	if mode == msgValueZero {
		depositAmount = eth.ZeroWei
	}

	msgValue := eth.ZeroWei
	if mode == msgValueWith {
		msgValue = depositAmount
	}

	switch asset {
	case assetETH:
		runETHDepositCase(t, sys, l1User, l2User, path, depositAmount, msgValue, mode)
	case assetMNT:
		runMNTDepositCase(t, sys, l1User, l2User, path, depositAmount, msgValue)
	case assetERC20:
		runERC20DepositCase(t, sys, l1User, l2User, path, depositAmount, msgValue)
	default:
		t.Require().Fail("unknown asset type", "asset", asset)
	}
}

func runETHDepositCase(t devtest.T, sys *presets.MantleMinimal, l1User *dsl.EOA, l2User *dsl.EOA, path depositPath, amount, msgValue eth.ETH, mode msgValueMode) {
	initial := l2User.GetTokenBalance(bvmETHAddr)

	if path == depositByPortal {
		portalAddr := sys.L2Chain.Escape().RollupConfig().DepositContractAddress
		portal := bindings.NewBindings[bindings.MantleOptimismPortal](
			bindings.WithTest(t),
			bindings.WithClient(sys.L1EL.EthClient()),
			bindings.WithTo(portalAddr),
		)
		call := portal.DepositTransaction(amount, eth.ZeroWei, l1User.Address(), eth.ZeroWei, 300_000, false, []byte{})
		receipt := writeDepositTx(t, l1User, call, msgValue)
		if mode == msgValueWithout && amount != eth.ZeroWei {
			waitForL2DepositStatus(t, sys, receipt, types.ReceiptStatusSuccessful)
			l2User.WaitForTokenBalance(bvmETHAddr, initial)
			return
		}
		waitForL2Deposit(t, sys, receipt)

		expected := initial.Add(amount)
		l2User.WaitForTokenBalance(bvmETHAddr, expected)
		return
	}

	bridgeAddr := sys.L2Chain.Escape().Deployment().L1StandardBridgeProxyAddr()
	bridge := bindings.NewBindings[bindings.MantleL1StandardBridge](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(bridgeAddr),
	)

	call := bridge.DepositETH(200_000, []byte{})
	if mode == msgValueWithout && amount != eth.ZeroWei {
		// ETH bridge deposits depend on msg.value. With zero msg.value the L2 balance won't change.
		amount = eth.ZeroWei
	}

	receipt := writeDepositTx(t, l1User, call, msgValue)
	waitForL2Deposit(t, sys, receipt)

	expected := initial.Add(amount)
	l2User.WaitForTokenBalance(bvmETHAddr, expected)
}

func runMNTDepositCase(t devtest.T, sys *presets.MantleMinimal, l1User *dsl.EOA, l2User *dsl.EOA, path depositPath, amount, msgValue eth.ETH) {
	l1MNTAddr := sysgo.DefaultL1MNT
	fundMNT(t, sys, l1User, amount)
	initial := l2User.GetBalance()

	if path == depositByPortal {
		portalAddr := sys.L2Chain.Escape().RollupConfig().DepositContractAddress
		l1User.ApproveToken(l1MNTAddr, portalAddr, amount)
		portal := bindings.NewBindings[bindings.MantleOptimismPortal](
			bindings.WithTest(t),
			bindings.WithClient(sys.L1EL.EthClient()),
			bindings.WithTo(portalAddr),
		)
		call := portal.DepositTransaction(eth.ZeroWei, amount, l1User.Address(), eth.ZeroWei, 300_000, false, []byte{})
		receipt := writeDepositTx(t, l1User, call, msgValue)
		waitForL2Deposit(t, sys, receipt)

		expected := initial.Add(amount)
		l2User.WaitForBalance(expected)
		return
	}

	bridgeAddr := sys.L2Chain.Escape().Deployment().L1StandardBridgeProxyAddr()
	l1User.ApproveToken(l1MNTAddr, bridgeAddr, amount)
	bridge := bindings.NewBindings[bindings.MantleL1StandardBridge](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(bridgeAddr),
	)
	call := bridge.DepositMNT(amount, 200_000, []byte{})
	if msgValue != eth.ZeroWei {
		_, err := writeDepositTxAllowError(t, l1User, call, msgValue)
		t.Require().Error(err, "MNT bridge deposit should revert with msg.value")
		return
	}
	receipt := writeDepositTx(t, l1User, call, msgValue)
	waitForL2Deposit(t, sys, receipt)

	expected := initial.Add(amount)
	l2User.WaitForBalance(expected)
}

func runERC20DepositCase(t devtest.T, sys *presets.MantleMinimal, l1User *dsl.EOA, l2User *dsl.EOA, path depositPath, amount, msgValue eth.ETH) {
	if path == depositByPortal {
		t.Skip("portal deposits do not mint or bridge ERC20 tokens directly; use bridge path")
	}

	l1TokenAddr, l2TokenAddr := setupERC20Bridge(t, sys, l1User, l2User, amount)
	l1BridgeAddr := sys.L2Chain.Escape().Deployment().L1StandardBridgeProxyAddr()
	l1User.ApproveToken(l1TokenAddr, l1BridgeAddr, amount)
	initial := l2User.GetTokenBalance(l2TokenAddr)

	bridge := bindings.NewBindings[bindings.MantleL1StandardBridge](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(l1BridgeAddr),
	)
	call := bridge.DepositERC20(l1TokenAddr, l2TokenAddr, amount, 200_000, []byte{})

	if msgValue != eth.ZeroWei {
		_, err := writeDepositTxAllowError(t, l1User, call, msgValue)
		t.Require().Error(err, "ERC20 bridge deposit should revert with msg.value")
		return
	}

	receipt := writeDepositTx(t, l1User, call, msgValue)
	waitForL2Deposit(t, sys, receipt)
	l2User.WaitForTokenBalance(l2TokenAddr, initial.Add(amount))
}

func fundMNT(t devtest.T, sys *presets.MantleMinimal, to *dsl.EOA, amount eth.ETH) {
	if amount == eth.ZeroWei {
		return
	}
	holderPriv := sys.L2Chain.Escape().Keys().Secret(devkeys.UserKey(0))
	holder := dsl.NewKey(t, holderPriv).User(sys.L1EL)
	sys.FunderL1.FundAtLeast(holder, eth.OneTenthEther)

	mntFunder := dsl.NewMNTFunder(t, sysgo.DefaultL1MNT, holder)
	mntFunder.FundAtLeast(to, amount)
	to.WaitForTokenBalance(sysgo.DefaultL1MNT, amount)
}

func setupERC20Bridge(t devtest.T, sys *presets.MantleMinimal, l1User *dsl.EOA, l2User *dsl.EOA, amount eth.ETH) (common.Address, common.Address) {
	l1TokenAddr := l1User.DeployWETH()

	weth := bindings.NewBindings[bindings.WETH](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(l1TokenAddr),
	)
	contract.Write(l1User, weth.Deposit(), txplan.WithValue(amount))
	l1User.WaitForTokenBalance(l1TokenAddr, amount)

	bridge := sys.StandardBridge()
	l2TokenAddr := bridge.CreateL2Token(l1TokenAddr, "L2 WETH", "L2WETH", l2User)
	l2User.WaitForTokenBalance(l2TokenAddr, eth.ZeroWei)
	return l1TokenAddr, l2TokenAddr
}

func writeDepositTx(t devtest.T, user *dsl.EOA, call bindings.TypedCall[any], msgValue eth.ETH) *types.Receipt {
	receipt, err := writeDepositTxAllowError(t, user, call, msgValue)
	t.Require().NoError(err, "deposit tx failed: %v", errutil.TryAddRevertReason(err))
	t.Require().Equal(types.ReceiptStatusSuccessful, receipt.Status, "deposit tx failed")
	return receipt
}

func writeDepositTxAllowError(t devtest.T, user *dsl.EOA, call bindings.TypedCall[any], msgValue eth.ETH) (*types.Receipt, error) {
	plan, err := contractio.Plan(call)
	if err != nil {
		return nil, err
	}
	opts := []txplan.Option{user.Plan()}
	if msgValue != eth.ZeroWei {
		opts = append(opts, txplan.WithValue(msgValue))
	}
	opts = append(opts, txplan.WithRetryInclusion(call.Client(), 10, retry.Exponential()))
	tx := txplan.NewPlannedTx(plan, txplan.Combine(opts...))

	signed, err := tx.Signed.Eval(t.Ctx())
	if err != nil {
		return nil, err
	}
	t.Logger().Info("deposit tx signed", "hash_full", signed.Hash().Hex(), "from", signed.From)

	receipt, err := tx.Included.Eval(t.Ctx())
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func waitForL2Deposit(t devtest.T, sys *presets.MantleMinimal, receipt *types.Receipt) {
	waitForL2DepositStatus(t, sys, receipt, types.ReceiptStatusSuccessful)
}

func waitForL2DepositStatus(t devtest.T, sys *presets.MantleMinimal, receipt *types.Receipt, expectedStatus uint64) {
	var l2DepositTx *types.DepositTx
	for _, log := range receipt.Logs {
		if dep, err := derive.UnmarshalDepositLogEvent(log); err == nil {
			l2DepositTx = dep
			break
		}
	}
	t.Require().NotNil(l2DepositTx, "could not reconstruct L2 deposit transaction")

	hash := types.NewTx(l2DepositTx).Hash()
	sequencingWindow := time.Duration(sys.L2Chain.Escape().RollupConfig().SeqWindowSize) * sys.L1EL.EstimateBlockTime()
	var l2Receipt *types.Receipt
	t.Require().Eventually(func() bool {
		var err error
		l2Receipt, err = sys.L2EL.Escape().EthClient().TransactionReceipt(t.Ctx(), hash)
		return err == nil
	}, sequencingWindow, 500*time.Millisecond, "L2 deposit never found")
	t.Require().Equal(expectedStatus, l2Receipt.Status, "L2 deposit status mismatch")
}

func runPortalETHTxValueTransferCase(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	sys.L1Network.WaitForOnline()

	l1Sender := sys.FunderL1.NewFundedEOA(eth.OneTenthEther)
	l2Sender := l1Sender.AsEL(sys.L2EL)
	l2Recipient := sys.FunderL2.NewFundedEOA(eth.ZeroWei)

	bridge := sys.MantleBridge()
	fundAmount := eth.GWei(2_000_000)
	initialSender := l2Sender.GetTokenBalance(bvmETHAddr)
	bridge.DepositETH(fundAmount, l1Sender)
	l2Sender.WaitForTokenBalance(bvmETHAddr, initialSender.Add(fundAmount))

	transferAmount := eth.GWei(1_000_000)
	initialRecipient := l2Recipient.GetTokenBalance(bvmETHAddr)

	portalAddr := sys.L2Chain.Escape().RollupConfig().DepositContractAddress
	portal := bindings.NewBindings[bindings.MantleOptimismPortal](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(portalAddr),
	)
	call := portal.DepositTransaction(transferAmount, eth.ZeroWei, l2Recipient.Address(), eth.ZeroWei, 300_000, false, []byte{})
	receipt := writeDepositTx(t, l1Sender, call, eth.ZeroWei)
	waitForL2Deposit(t, sys, receipt)

	l2Recipient.WaitForTokenBalance(bvmETHAddr, initialRecipient.Add(transferAmount))
	l2Sender.WaitForTokenBalance(bvmETHAddr, initialSender.Add(fundAmount).Sub(transferAmount))
}

func runPortalMNTTxValueTransferCase(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	sys.L1Network.WaitForOnline()

	l1Sender := sys.FunderL1.NewFundedEOA(eth.OneTenthEther)
	l2Sender := l1Sender.AsEL(sys.L2EL)
	l2Recipient := sys.FunderL2.NewFundedEOA(eth.ZeroWei)

	transferAmount := eth.GWei(1_000_000)
	sys.FunderL2.FundAtLeast(l2Sender, transferAmount.Mul(2))

	initialSender := l2Sender.GetBalance()
	initialRecipient := l2Recipient.GetBalance()

	portalAddr := sys.L2Chain.Escape().RollupConfig().DepositContractAddress
	portal := bindings.NewBindings[bindings.MantleOptimismPortal](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(portalAddr),
	)
	call := portal.DepositTransaction(eth.ZeroWei, eth.ZeroWei, l2Recipient.Address(), transferAmount, 300_000, false, []byte{})
	receipt := writeDepositTx(t, l1Sender, call, eth.ZeroWei)
	waitForL2Deposit(t, sys, receipt)

	l2Recipient.WaitForBalance(initialRecipient.Add(transferAmount))
	l2Sender.WaitForBalance(initialSender.Sub(transferAmount))
}
