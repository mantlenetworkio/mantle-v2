package withdrawal

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/custom_gas_token"
	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	gameTypes "github.com/ethereum-optimism/optimism/op-challenger/game/types"
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl/contract"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum/go-ethereum/common"
)

var bvmETHAddr = common.HexToAddress("0xdEAddEaDdeadDEadDEADDEAddEADDEAddead1111")

func InitWithGameType(m *testing.M, gameType gameTypes.GameType) {
	_ = gameType
	presets.DoMain(m,
		presets.WithCompatibleTypes(compat.SysGo),
		presets.WithMantleMinimal(),
		presets.WithTimeTravel(),
		presets.WithFinalizationPeriodSeconds(1),
	)
}

func TestWithdrawalMNT(gt *testing.T, gameType gameTypes.GameType) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	_ = gameType

	custom_gas_token.SkipIfCGT(t, sys)
	sys.L1Network.WaitForOnline()

	bridge := sys.MantleBridge()

	initialL1Balance := eth.OneThirdEther
	l1User := sys.FunderL1.NewFundedEOA(initialL1Balance)
	l2User := l1User.AsEL(sys.L2EL)

	depositAmount := eth.OneTenthEther
	withdrawalAmount := eth.OneHundredthEther

	l1MNTAddr := sysgo.DefaultL1MNT
	l1BridgeAddr := sys.L2Chain.Escape().Deployment().L1StandardBridgeProxyAddr()

	mntFunderKey := dsl.NewKey(t, sys.L2Chain.Escape().Keys().Secret(devkeys.UserKey(0)))
	mntFunder := mntFunderKey.User(sys.L1EL)
	sys.FunderL1.FundAtLeast(mntFunder, eth.OneTenthEther)

	mntToken := bindings.NewBindings[bindings.OptimismMintableERC20](
		bindings.WithTest(t),
		bindings.WithClient(sys.L1EL.EthClient()),
		bindings.WithTo(l1MNTAddr),
	)

	funderMNTBalance := contract.Read(mntToken.BalanceOf(mntFunder.Address()))
	t.Require().True(funderMNTBalance.Gt(eth.ZeroWei), "L1 MNT funder has no balance")
	contract.Write(mntFunder, mntToken.Transfer(l1User.Address(), depositAmount))
	l1User.WaitForTokenBalance(l1MNTAddr, depositAmount)
	initialL1MNTBalance := l1User.GetTokenBalance(l1MNTAddr)

	approveReceipt := contract.Write(l1User, mntToken.Approve(l1BridgeAddr, depositAmount))
	expectedL1Balance := initialL1Balance.Sub(bridge.L1GasCost(approveReceipt))

	initialL2Balance := l2User.GetBalance()
	deposit := bridge.DepositMNT(depositAmount, l1User)
	expectedL1Balance = expectedL1Balance.Sub(deposit.GasCost())
	l1User.VerifyBalanceExact(expectedL1Balance)
	l1User.WaitForTokenBalance(l1MNTAddr, initialL1MNTBalance.Sub(depositAmount))

	expectedL2Balance := initialL2Balance.Add(depositAmount)
	l2User.WaitForBalance(expectedL2Balance)

	withdrawal := bridge.InitiateWithdrawalMNT(withdrawalAmount, l2User)
	expectedL2Balance = expectedL2Balance.Sub(withdrawalAmount).Sub(withdrawal.InitiateGasCost())
	l2User.VerifyBalanceExact(expectedL2Balance)

	withdrawal.Prove(l1User)
	expectedL1Balance = expectedL1Balance.Sub(withdrawal.ProveGasCost())
	l1User.VerifyBalanceExact(expectedL1Balance)

	sys.AdvanceTime(bridge.WithdrawalDelay() + time.Second)
	t.Logger().Info("Attempting to finalize", "finalizationDelay", bridge.WithdrawalDelay())
	withdrawal.Finalize(l1User)
	expectedL1Balance = expectedL1Balance.Sub(withdrawal.FinalizeGasCost())
	l1User.VerifyBalanceExact(expectedL1Balance)

	expectedL1MNTBalance := initialL1MNTBalance.Sub(depositAmount).Add(withdrawalAmount)
	l1User.WaitForTokenBalance(l1MNTAddr, expectedL1MNTBalance)
}

func TestWithdrawalETH(gt *testing.T, gameType gameTypes.GameType) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	_ = gameType

	custom_gas_token.SkipIfCGT(t, sys)
	sys.L1Network.WaitForOnline()

	bridge := sys.MantleBridge()

	initialL1Balance := eth.OneThirdEther
	l1User := sys.FunderL1.NewFundedEOA(initialL1Balance)
	l2User := l1User.AsEL(sys.L2EL)

	// Ensure the L2 user has MNT to pay for L2 gas.
	sys.FunderL2.FundAtLeast(l2User, eth.OneTenthEther)

	depositAmount := eth.OneTenthEther
	withdrawalAmount := eth.OneHundredthEther

	initialL2BVMETHBalance := l2User.GetTokenBalance(bvmETHAddr)
	deposit := bridge.DepositETH(depositAmount, l1User)
	expectedL1Balance := initialL1Balance.Sub(depositAmount).Sub(deposit.GasCost())
	l1User.VerifyBalanceExact(expectedL1Balance)
	l2User.WaitForTokenBalance(bvmETHAddr, initialL2BVMETHBalance.Add(depositAmount))

	withdrawal := bridge.InitiateWithdrawalETH(withdrawalAmount, l1User.Address(), l2User)
	l2User.WaitForTokenBalance(bvmETHAddr, initialL2BVMETHBalance.Add(depositAmount).Sub(withdrawalAmount))

	withdrawal.Prove(l1User)
	expectedL1Balance = expectedL1Balance.Sub(withdrawal.ProveGasCost())
	l1User.VerifyBalanceExact(expectedL1Balance)

	sys.AdvanceTime(bridge.WithdrawalDelay() + time.Second)
	t.Logger().Info("Attempting to finalize", "finalizationDelay", bridge.WithdrawalDelay())
	withdrawal.Finalize(l1User)
	expectedL1Balance = expectedL1Balance.Sub(withdrawal.FinalizeGasCost()).Add(withdrawalAmount)
	l1User.VerifyBalanceExact(expectedL1Balance)
}
