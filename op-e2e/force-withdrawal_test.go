package op_e2e

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"log"
	"math/big"
	"strings"
	"testing"
	"time"
)

var l1Erc20 = "0x89F06180e62a6d3e5ac130bbCE7bD004b434100b"
var l2Erc20 = "0x89F06180e62a6d3e5ac130bbCE7bD004b434100b"

const (
	l1url = "http://localhost:8545"
	l2url = "http://localhost:9545"

	//l1ChainId = 900
	//L2ChainId = 31337

	l1ChainId = 900
	L2ChainId = 901

	l2EthAddress    = "0xdEAddEaDdeadDEadDEADDEAddEADDEAddead1111"
	l1BridgeAddress = "0x6900000000000000000000000000000000000003"
	l2BridgeAddress = "0x4200000000000000000000000000000000000010"
	l1weth          = "0x6900000000000000000000000000000000000007"
	l1MntAddress    = "0x6900000000000000000000000000000000000020"
	l2MntAddress    = "0x0000000000000000000000000000000000000000"

	userPrivateKey = "ddf04c9058d6fac4fea241820f2fbc3b36868d33b80894ba5ff9a9baf8793e10"
	userAddress    = "0xeE3e7d56188ae7af8d5bab980908E3e91c0d7384"

	//userPrivateKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
	//userAddress    = "0x71562b71999873DB5b286dF957af199Ec94617F7"

	deployerPrivateKey = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
	deployerAddress    = "0x71562b71999873DB5b286dF957af199Ec94617F7"

	DECIMAL5    = 5000000000000000000
	DECIMAL1    = 1000000000000000000
	DECIMAL0_5  = 500000000000000000
	DECIMAL0_1  = 100000000000000000
	DECIMAL00_1 = 10000000000000000
)

func TestEnv(t *testing.T) {

	a := getETHBalanceFromL2(t, "0x4200000000000000000000000000000000000007")
	t.Logf("a : %v", a)

	a = getETHBalanceFromL2(t, "0x7A11000000000000000000000000000000001113")
	t.Logf("a : %v", a)
	//a = getETHBalanceFromL2(t, "0x6900000000000000000000000000000000000003")
	//t.Logf("a : %v", a)

	t.Log("check l1 mnt token address.....")
	checkTokenAddress(t)

	t.Log("check token bridge address.....")
	checkTokenBridge(t)

	t.Log("check balance.....")
	checkBalance(t)
	t.Log("show balance.....")

}

func TestMainProcess(t *testing.T) {
	TestEnv(t)
	TestERC20DepositAndWithdrawal(t)
	TestMNTDepositAndWithdrawal(t)
	TestETHDepositAndWithdrawal(t)
}

func checkBalance(t *testing.T) *big.Int {

	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// init balance

	l1Eth := getETHBalanceFromL1(t, userAddress)
	l1Mnt := getMNTBalanceFromL1(t, userAddress)

	t.Log("find l1 mnt balance : ", l1Mnt)
	decimal1 := big.NewInt(DECIMAL1)
	if l1Eth.Cmp(decimal1) < 0 {
		delta := big.NewInt(0)
		transferL1ETH(t, l1Client, common.HexToAddress(userAddress), delta.Sub(decimal1, l1Eth).Int64())
		l1Eth = getETHBalanceFromL1(t, userAddress)

	}
	if l1Mnt.Cmp(decimal1) < 0 {
		delta := big.NewInt(0)
		transferL1MNTFromDeployer(t, l1Client, delta.Sub(decimal1, l1Mnt).Int64())
	}
	l1Eth = getETHBalanceFromL1(t, userAddress)
	if l1Eth.Cmp(decimal1) < 0 {
		delta := big.NewInt(0)
		transferL1ETH(t, l1Client, common.HexToAddress(userAddress), delta.Sub(big.NewInt(DECIMAL1+DECIMAL00_1), l1Eth).Int64())
		l1Eth = getETHBalanceFromL1(t, userAddress)
	}
	l2Mnt := getMNTBalanceFromL2(t, userAddress)
	if l2Mnt.Cmp(decimal1) < 0 {
		delta := big.NewInt(0)
		transferL2MNT(t, l2Client, common.HexToAddress(userAddress), delta.Sub(decimal1, l2Mnt).Int64())
		l2Mnt = getMNTBalanceFromL2(t, userAddress)
	}

	l1Eth = getETHBalanceFromL1(t, userAddress)
	l1Mnt = getMNTBalanceFromL1(t, userAddress)
	l2Eth := getETHBalanceFromL2(t, userAddress)
	l2Mnt = getMNTBalanceFromL2(t, userAddress)
	//require.LessOrEqual(t, int64(DECIMAL1), l1Eth.Int64())
	//require.Equal(t, int64(DECIMAL1), l1Mnt.Int64())
	t.Log("L1 BALANCE INFO")

	t.Log("l1 balance eth: ", l1Eth)
	t.Log("l1 balance mnt: ", l1Mnt)

	t.Log("L2 BALANCE INFO")

	t.Log("l2 balance eth: ", l2Eth)
	t.Log("l2 balance mnt: ", l2Mnt)

	return l1Eth
}

func TestContractsProxy(t *testing.T) {
	t.Log("check l1 mnt token.....")
	checkTokenAddress(t)

	t.Log("check token bridge.....")
	checkTokenBridge(t)
}

func TestDeployERC20TestToken(t *testing.T) {
	t.Log("check balance.....")

	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// query eth erc20 token
	privateKey, err := crypto.HexToECDSA(userPrivateKey)
	if err != nil {
		log.Fatalln(err.Error())
	}

	nonce, err := l1Client.PendingNonceAt(context.Background(), common.HexToAddress(userAddress))
	require.NoError(t, err)

	gasPrice, err := l1Client.SuggestGasPrice(context.Background())
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(l1ChainId))
	require.NoError(t, err)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice
	t.Log("l1 gas price = ", gasPrice)
	address, tx, instance, err := bindings.DeployL1TestToken(auth, l1Client, "L1Token", "L1T")
	require.NoError(t, err)
	t.Log("tx.Hash : ", tx.Hash().Hex())
	_, err = waitForTransaction(tx.Hash(), l1Client, 100*time.Second)
	require.NoError(t, err)
	_ = instance
	t.Log("L1 Token address = ", address.Hex())

	nonce, err = l2Client.PendingNonceAt(context.Background(), common.HexToAddress(userAddress))
	require.NoError(t, err)

	gasPrice, err = l2Client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	t.Log("l2 gas price = ", gasPrice)

	auth2, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(L2ChainId))
	require.NoError(t, err)
	auth2.Nonce = big.NewInt(int64(nonce))
	auth2.Value = big.NewInt(0)      // in wei
	auth2.GasLimit = uint64(3000000) // in units
	auth2.GasPrice = gasPrice
	l2Address, tx, instance2, err := bindings.DeployL2TestToken(auth2, l2Client, address)
	require.NoError(t, err)
	t.Log("tx.Hash:", tx.Hash().Hex())
	_, err = waitForTransaction(tx.Hash(), l2Client, 100*time.Second)

	require.NoError(t, err)
	_ = instance2
	t.Log("L2 Token address = ", l2Address.Hex())

	l1Erc20 = address.Hex()
	l2Erc20 = l2Address.Hex()
}

func TestETHDepositAndWithdrawal(t *testing.T) {
	t.Log("check balance.....")

	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// query eth erc20 token
	l1Bridge, err := bindings.NewL1StandardBridge(common.HexToAddress(l1BridgeAddress), l1Client)
	require.NoError(t, err)
	l2Bridge, err := bindings.NewL2StandardBridge(common.HexToAddress(l2BridgeAddress), l2Client)
	//require.NoError(t, err)

	// TEST deposit ETH
	t.Log("----------------")
	t.Log("ETH DEPOSIT TEST")
	t.Log("----------------")
	t.Log("ETH before deposit...\\")

	beforeBalanceL1 := getETHBalanceFromL1(t, userAddress)
	beforeBalanceL2 := getETHBalanceFromL2(t, userAddress)

	t.Log("l1 eth balance: ", beforeBalanceL1)
	t.Log("l2 eth balance: ", beforeBalanceL2)
	// do deposit
	auth := buildL1Auth(t, l1Client, userPrivateKey, big.NewInt(DECIMAL0_1))
	tx, err := l1Bridge.DepositETH(auth, 2_000_000, []byte("0x"))
	_, err = waitForTransaction(tx.Hash(), l1Client, 100*time.Second)
	require.NoError(t, err)
	time.Sleep(30 * time.Second)

	t.Log("deposit eth tx hash is: ", tx.Hash())
	t.Log("ETH after deposit...\\")
	afterBalanceL1 := getETHBalanceFromL1(t, userAddress)
	afterBalanceL2 := getETHBalanceFromL2(t, userAddress)

	t.Log("l1 eth balance: ", afterBalanceL1)
	t.Log("l2 eth balance: ", afterBalanceL2)

	//require.Equal(t, getETHBalanceFromL2(t, userAddress), 0)
	t.Log("eth deposit amount: ", uint64(DECIMAL0_1))

	require.Equal(t, afterBalanceL2.Uint64()-beforeBalanceL2.Uint64(), uint64(DECIMAL0_1))

	// TEST withdraw ETH
	t.Log("-----------------")
	t.Log("ETH WITHDRAW TEST")
	t.Log("-----------------")
	t.Log("ETH before withdraw.....\\")
	setL2EthApprove(t)

	beforeBalanceL1 = getETHBalanceFromL1(t, userAddress)
	beforeBalanceL2 = getETHBalanceFromL2(t, userAddress)

	t.Log("l1 eth balance: ", beforeBalanceL1)
	t.Log("l2 eth balance: ", beforeBalanceL2)
	auth = buildL2Auth(t, l2Client, userPrivateKey, big.NewInt(0))
	tx, err = l2Bridge.Withdraw(auth, common.HexToAddress(l2EthAddress), big.NewInt(DECIMAL0_1), 300_000, []byte("0x"))
	require.NoError(t, err)
	t.Log("withdraw eth tx hash is: ", tx.Hash())
	SingleWithdrawalTx(t, tx.Hash().Hex())

	t.Log("ETH after withdraw.....\\")
	time.Sleep(10 * time.Second)

	afterBalanceL1 = getETHBalanceFromL1(t, userAddress)
	afterBalanceL2 = getETHBalanceFromL2(t, userAddress)
	t.Log("l1 eth balance: ", afterBalanceL1)
	t.Log("l2 eth balance: ", afterBalanceL2)
	t.Log("eth withdraw amount: ", DECIMAL0_1)

	require.Equal(t, uint64(DECIMAL0_1), afterBalanceL1.Uint64()-beforeBalanceL1.Uint64())

}

func TestMNTDepositAndWithdrawal(t *testing.T) {
	t.Log("check balance.....")

	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// query eth erc20 token
	l1Bridge, err := bindings.NewL1StandardBridge(common.HexToAddress(l1BridgeAddress), l1Client)
	require.NoError(t, err)
	l2Bridge, err := bindings.NewL2StandardBridge(common.HexToAddress(l2BridgeAddress), l2Client)
	//require.NoError(t, err)

	// TEST deposit ETH
	t.Log("----------------")
	t.Log("MNT DEPOSIT TEST")
	t.Log("----------------")
	t.Log("MNT before deposit...\\")

	beforeBalanceL1 := getMNTBalanceFromL1(t, userAddress)
	beforeBalanceL2 := getMNTBalanceFromL2(t, userAddress)

	t.Log("l1 mnt balance: ", beforeBalanceL1)
	t.Log("l2 mnt balance: ", beforeBalanceL2)

	setL1MntApprove(t)

	// do deposit
	auth := buildL1Auth(t, l1Client, userPrivateKey, big.NewInt(0))
	tx, err := l1Bridge.DepositMNT(auth, big.NewInt(DECIMAL0_1), 2_000_000, []byte("0x"))
	require.NoError(t, err)
	t.Log("deposit mnt tx hash is: ", tx.Hash())
	t.Log("MNT after deposit...\\")
	_, err = waitForTransaction(tx.Hash(), l1Client, 100*time.Second)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)

	afterBalanceL1 := getMNTBalanceFromL1(t, userAddress)
	afterBalanceL2 := getMNTBalanceFromL2(t, userAddress)

	t.Log("l1 mnt balance: ", afterBalanceL1)
	t.Log("l2 mnt balance: ", afterBalanceL2)
	//require.Equal(t, getETHBalanceFromL2(t, userAddress), 0)
	t.Log("mnt deposit amount: ", DECIMAL0_1)

	require.Equal(t, afterBalanceL2.Uint64()-beforeBalanceL2.Uint64(), uint64(DECIMAL0_1))

	//TEST withdraw MNT
	t.Log("-----------------")
	t.Log("MNT WITHDRAW TEST")
	t.Log("-----------------")
	t.Log("MNT before withdraw.....\\")

	beforeBalanceL1 = getMNTBalanceFromL1(t, userAddress)
	beforeBalanceL2 = getMNTBalanceFromL2(t, userAddress)

	t.Log("l1 mnt balance: ", beforeBalanceL1)
	t.Log("l2 mnt balance: ", beforeBalanceL2)

	auth = buildL2Auth(t, l2Client, userPrivateKey, big.NewInt(DECIMAL0_1))
	tx, err = l2Bridge.Withdraw(auth, common.HexToAddress("0x0"), big.NewInt(DECIMAL0_1), 300_000, []byte("0x"))
	require.NoError(t, err)
	t.Log("withdraw mnt tx hash is: ", tx.Hash())
	SingleWithdrawalTx(t, tx.Hash().Hex())

	t.Log("MNT after withdraw.....\\")
	time.Sleep(10 * time.Second)

	afterBalanceL1 = getMNTBalanceFromL1(t, userAddress)
	afterBalanceL2 = getMNTBalanceFromL2(t, userAddress)

	t.Log("l1 mnt balance: ", afterBalanceL1)
	t.Log("l2 mnt balance: ", afterBalanceL2)

	t.Log("mnt withdraw amount: ", DECIMAL0_1)
	require.Equal(t, uint64(DECIMAL0_1), afterBalanceL1.Uint64()-beforeBalanceL1.Uint64())

}

func TestERC20DepositAndWithdrawal(t *testing.T) {
	TestDeployERC20TestToken(t)

	t.Logf("l1 erc20 address : %v ,l2 erc20 address : %v", l1Erc20, l2Erc20)

	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// query eth erc20 token
	l1Bridge, err := bindings.NewL1StandardBridge(common.HexToAddress(l1BridgeAddress), l1Client)
	require.NoError(t, err)
	l2Bridge, err := bindings.NewL2StandardBridge(common.HexToAddress(l2BridgeAddress), l2Client)
	require.NoError(t, err)

	// TEST deposit ETH
	t.Log("----------------")
	t.Log("TestToken DEPOSIT TEST")
	t.Log("----------------")
	t.Log("TestToken before deposit...\\")

	beforeBalanceL1 := getTestTokenBalanceFromL1(t, userAddress)
	beforeBalanceL2 := getTestTokenBalanceFromL2(t, userAddress)

	t.Log("l1 erc20 balance: ", beforeBalanceL1)
	t.Log("l2 erc20 balance: ", beforeBalanceL2)
	// do deposit

	setL1Erc20Approve(t)
	setL2Erc20Approve(t)

	auth := buildL1Auth(t, l1Client, userPrivateKey, big.NewInt(0))
	l1Erc20Addr := common.HexToAddress(l1Erc20)
	l2Erc20Addr := common.HexToAddress(l2Erc20)

	tx, err := l1Bridge.DepositERC20(auth, l1Erc20Addr, l2Erc20Addr, big.NewInt(DECIMAL0_1), 2_000_000, []byte("0x"))
	require.NoError(t, err)
	_, err = waitForTransaction(tx.Hash(), l1Client, 100*time.Second)
	require.NoError(t, err)
	t.Log("deposit erc20 tx hash is: ", tx.Hash())
	t.Log("erc20 after deposit...\\")
	time.Sleep(10 * time.Second)

	afterBalanceL1 := getTestTokenBalanceFromL1(t, userAddress)
	afterBalanceL2 := getTestTokenBalanceFromL2(t, userAddress)

	t.Log("l1 erc20 balance: ", afterBalanceL1)
	t.Log("l2 erc20 balance: ", afterBalanceL2)
	t.Log("erc20 deposit amount: ", DECIMAL0_1)
	require.NotEqual(t, beforeBalanceL1, afterBalanceL1)
	require.Equal(t, afterBalanceL2.Uint64()+afterBalanceL1.Uint64(), beforeBalanceL2.Uint64()+beforeBalanceL1.Uint64())

	// TEST withdraw ETH
	t.Log("-----------------")
	t.Log("ERC20 WITHDRAW TEST")
	t.Log("-----------------")
	t.Log("ERC20 before withdraw.....\\")
	setL2Erc20Approve(t)
	beforeBalanceL1 = getTestTokenBalanceFromL1(t, userAddress)
	beforeBalanceL2 = getTestTokenBalanceFromL2(t, userAddress)

	t.Log("l1 erc20 balance: ", beforeBalanceL1)
	t.Log("l2 erc20 balance: ", beforeBalanceL2)

	auth = buildL2Auth(t, l2Client, userPrivateKey, big.NewInt(0))
	tx, err = l2Bridge.Withdraw(auth, common.HexToAddress(l2Erc20), big.NewInt(DECIMAL0_1), 300_000, []byte("0x"))
	require.NoError(t, err)
	t.Log("withdraw erc20 tx hash is: ", tx.Hash())

	SingleWithdrawalTx(t, tx.Hash().Hex())

	t.Log("erc20 after withdraw.....\\")

	afterBalanceL1 = getTestTokenBalanceFromL1(t, userAddress)
	afterBalanceL2 = getTestTokenBalanceFromL2(t, userAddress)

	t.Log("l1 erc20 balance: ", afterBalanceL1)
	t.Log("l2 erc20 balance: ", afterBalanceL2)
	t.Log("erc20 withdraw amount: ", DECIMAL0_1)
	require.Equal(t, afterBalanceL2.Uint64()+afterBalanceL1.Uint64(), beforeBalanceL2.Uint64()+beforeBalanceL1.Uint64())

}

func TestCheckAccountBalance(t *testing.T) {
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)
	transferL2MNT(t, l2Client, common.HexToAddress(userAddress), DECIMAL1)
}

func TestShowL1L2Balance(t *testing.T) {
	l1Eth := getETHBalanceFromL1(t, userAddress)
	l2Eth := getETHBalanceFromL2(t, userAddress)
	t.Log("l1 eth balance: ", l1Eth)
	t.Log("l2 eth balance: ", l2Eth)
	sumEth := big.NewInt(0)
	t.Log("eth sum balance is: ", sumEth.Add(l1Eth, l2Eth))

	l1Mnt := getMNTBalanceFromL1(t, userAddress)
	l2Mnt := getMNTBalanceFromL2(t, userAddress)
	t.Log("l1 mnt balance: ", l1Mnt)
	t.Log("l2 mnt balance: ", l2Mnt)
	sumMnt := big.NewInt(0)
	t.Log("mnt sum balance is: ", sumMnt.Add(l1Mnt, l2Mnt))
}

func checkTokenAddress(t *testing.T) {
	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// check l1 token address
	code, err := l1Client.CodeAt(context.Background(), common.HexToAddress(l1MntAddress), nil)
	require.NoError(t, err)
	require.True(t, len(code) > 0)
	t.Log("L1 ADDRESS INFO")
	t.Log("L1 Mnt Address: ", l1MntAddress)

	// check l2 token address

	code, err = l2Client.CodeAt(context.Background(), common.HexToAddress(l2EthAddress), nil)
	require.NoError(t, err)
	require.True(t, len(code) > 0)
	t.Log("L2 ADDRESS INFO")
	t.Log("L2 Mnt Address: ", "l2MntAddress")
	t.Log("L2 ETH Address: ", l2EthAddress)
}

func checkTokenBridge(t *testing.T) {
	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// check l1 token bridge
	code, err := l1Client.CodeAt(context.Background(), common.HexToAddress(l1BridgeAddress), nil)
	require.NoError(t, err)
	require.NotEmpty(t, code)
	t.Log("TOKEN BRIDGE INFO")
	t.Log("find l1 token bridge at: ", l1BridgeAddress)

	code, err = l1Client.CodeAt(context.Background(), common.HexToAddress(l1weth), nil)
	require.NoError(t, err)
	require.NotEmpty(t, code)
	t.Log("L1WETH INFO")
	t.Log("find l1 weth at: ", l1BridgeAddress)

	// check l2 token bridge
	code, err = l2Client.CodeAt(context.Background(), common.HexToAddress(l2BridgeAddress), nil)
	require.NoError(t, err)
	require.NotEmpty(t, code)
	t.Log("find l2 token bridge at: ", l2BridgeAddress)
}

func setL1MntApprove(t *testing.T) {
	client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l1MntInstance, err := bindings.NewL1MantleToken(common.HexToAddress(l1MntAddress), client)
	require.NoError(t, err)
	auth := buildL1Auth(t, client, userPrivateKey, big.NewInt(0))
	tx, err := l1MntInstance.Approve(auth, common.HexToAddress(l1BridgeAddress), big.NewInt(DECIMAL5))
	require.NoError(t, err)
	require.NotNil(t, tx)
	t.Log("l1 mnt approve tx = ", tx.Hash().String())
	_, err = waitForTransaction(tx.Hash(), client, 100*time.Second)
	require.NoError(t, err)

	l1MntAllowance, err := l1MntInstance.Allowance(&bind.CallOpts{}, common.HexToAddress(userAddress), common.HexToAddress(l1BridgeAddress))
	require.NoError(t, err)

	t.Log("l1mnt allowance ", l1MntAllowance)
	require.Equal(t, int64(DECIMAL5), l1MntAllowance.Int64())
}

func setL2EthApprove(t *testing.T) {
	client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l2EthInstance, err := bindings.NewBVMETH(common.HexToAddress(l2EthAddress), client)
	require.NoError(t, err)
	auth := buildL2Auth(t, client, userPrivateKey, big.NewInt(0))
	tx, err := l2EthInstance.Approve(auth, common.HexToAddress(l2BridgeAddress), big.NewInt(DECIMAL5))
	require.NoError(t, err)
	require.NotNil(t, tx)
	t.Log("approve tx = ", tx.Hash().String())
	_, err = waitForTransaction(tx.Hash(), client, 100*time.Second)
	require.NoError(t, err)

	l1MntAllowance, err := l2EthInstance.Allowance(&bind.CallOpts{}, common.HexToAddress(userAddress), common.HexToAddress(l2BridgeAddress))
	require.NoError(t, err)
	t.Log("approve amount = ", l1MntAllowance.Int64())

	//require.Equal(t, int64(DECIMAL5), l1MntAllowance.Int64())
}

func setL1Erc20Approve(t *testing.T) {
	client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l1MntInstance, err := bindings.NewL1TestToken(common.HexToAddress(l1Erc20), client)
	require.NoError(t, err)
	auth := buildL1Auth(t, client, userPrivateKey, big.NewInt(0))
	tx, err := l1MntInstance.Approve(auth, common.HexToAddress(l1BridgeAddress), big.NewInt(DECIMAL5))
	require.NoError(t, err)
	t.Log("l1 erc20 approve tx = ", tx.Hash().String())
	require.NotNil(t, tx)

	_, err = waitForTransaction(tx.Hash(), client, 100*time.Second)
	require.NoError(t, err)

	l1MntAllowance, err := l1MntInstance.Allowance(&bind.CallOpts{}, common.HexToAddress(userAddress), common.HexToAddress(l1BridgeAddress))
	require.NoError(t, err)

	t.Log("l1mnt allowance ", l1MntAllowance)
	require.Equal(t, int64(DECIMAL5), l1MntAllowance.Int64())
}

func setL2Erc20Approve(t *testing.T) {
	client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l2EthInstance, err := bindings.NewL2TestToken(common.HexToAddress(l2Erc20), client)
	require.NoError(t, err)
	auth := buildL2Auth(t, client, userPrivateKey, big.NewInt(0))
	tx, err := l2EthInstance.Approve(auth, common.HexToAddress(l2BridgeAddress), big.NewInt(DECIMAL5))
	require.NoError(t, err)
	require.NotNil(t, tx)
	t.Log("l2 erc20 approve tx = ", tx.Hash().String())
	_, err = waitForTransaction(tx.Hash(), client, 100*time.Second)
	require.NoError(t, err)

	l1MntAllowance, err := l2EthInstance.Allowance(&bind.CallOpts{}, common.HexToAddress(userAddress), common.HexToAddress(l2BridgeAddress))
	require.NoError(t, err)
	t.Log("approve amount = ", l1MntAllowance.Int64())

	//require.Equal(t, int64(DECIMAL5), l1MntAllowance.Int64())
}

func getETHBalanceFromL1(t *testing.T, address string) *big.Int {
	client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, client)

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	require.NoError(t, err)
	require.NotNil(t, balance)
	return balance
}

func getETHBalanceFromL2(t *testing.T, address string) *big.Int {
	client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l2EthInstance, err := bindings.NewBVMETH(common.HexToAddress(l2EthAddress), client)
	require.NoError(t, err)
	balance, err := l2EthInstance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	require.NoError(t, err)
	require.NotNil(t, balance)
	return balance
}

func getTestTokenBalanceFromL2(t *testing.T, address string) *big.Int {
	client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l2EthInstance, err := bindings.NewL2TestToken(common.HexToAddress(l2Erc20), client)
	require.NoError(t, err)
	balance, err := l2EthInstance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	require.NoError(t, err)
	require.NotNil(t, balance)
	return balance
}

func getTestTokenBalanceFromL1(t *testing.T, address string) *big.Int {
	client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l2EthInstance, err := bindings.NewL1TestToken(common.HexToAddress(l1Erc20), client)
	require.NoError(t, err)
	balance, err := l2EthInstance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	require.NoError(t, err)
	require.NotNil(t, balance)
	return balance
}

func getMNTBalanceFromL1(t *testing.T, address string) *big.Int {
	client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l1MntInstance, err := bindings.NewL1MantleToken(common.HexToAddress(l1MntAddress), client)
	require.NoError(t, err)
	bal, err := l1MntInstance.BalanceOf(&bind.CallOpts{}, common.HexToAddress(address))
	require.NoError(t, err)
	require.NotNil(t, bal)
	return bal
}

func getMNTBalanceFromL2(t *testing.T, address string) *big.Int {
	client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, client)

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), nil)
	require.NoError(t, err)
	require.NotNil(t, balance)
	return balance
}

func buildL1Auth(t *testing.T, client *ethclient.Client, privateKey string, amount *big.Int) *bind.TransactOpts {
	return buildAuth(t, client, privateKey, amount, big.NewInt(l1ChainId))
}

func buildL2Auth(t *testing.T, client *ethclient.Client, privateKey string, amount *big.Int) *bind.TransactOpts {
	return buildAuth(t, client, privateKey, amount, big.NewInt(L2ChainId))
}

func buildAuth(t *testing.T, client *ethclient.Client, privateKey string, amount *big.Int, chainId *big.Int) *bind.TransactOpts {
	privKey, err := crypto.HexToECDSA(privateKey)
	require.NoError(t, err)

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err)

	//gasPrice :=big.NewInt(21000)
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	if err != nil {
		require.Nil(t, err)

	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = amount             // in wei
	auth.GasLimit = uint64(3000000) // in units
	//auth.GasPrice = gasPrice
	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)
	auth.GasPrice = gasPrice
	return auth
}

func transferL1ETH(t *testing.T, client *ethclient.Client, address common.Address, amount int64) {
	privateKey, err := crypto.HexToECDSA(deployerPrivateKey)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err)

	value := big.NewInt(amount) // in wei (1 eth)
	gasLimit := uint64(21000)   // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	var data []byte
	tx := types.NewTransaction(nonce, address, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	require.NoError(t, err)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(context.Background(), signedTx)
	require.NoError(t, err)
	_, err = waitForTransaction(signedTx.Hash(), client, 100*time.Second)
	require.NoError(t, err)
}

func transferL2MNT(t *testing.T, client *ethclient.Client, address common.Address, amount int64) {
	privateKey, err := crypto.HexToECDSA(deployerPrivateKey)
	require.NoError(t, err)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	require.True(t, ok)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err)

	value := big.NewInt(amount) // in wei (1 eth)
	gasLimit := uint64(21000)   // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	var data []byte
	tx := types.NewTransaction(nonce, address, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())

	require.NoError(t, err)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	require.NoError(t, err)

	err = client.SendTransaction(context.Background(), signedTx)
	require.NoError(t, err)
	_, err = waitForTransaction(signedTx.Hash(), client, 100*time.Second)
	require.NoError(t, err)
}

func transferL1MNTFromDeployer(t *testing.T, client *ethclient.Client, amount int64) {
	l1mntToken, err := bindings.NewL1MantleToken(common.HexToAddress(l1MntAddress), client)
	require.NoError(t, err)
	auth := buildL1Auth(t, client, deployerPrivateKey, big.NewInt(0))

	public_key := common.HexToAddress(userAddress)

	tx, err := l1mntToken.Transfer(auth, public_key, big.NewInt(amount))
	require.NoError(t, err)
	require.NotNil(t, tx)
	_, err = waitForTransaction(tx.Hash(), client, 100*time.Second)
	require.NoError(t, err)
	t.Log("mnt transfer tx : ", tx.Hash().String())
}

func TestDecimal(t *testing.T) {
	client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, client)

	l2EthInstance, err := bindings.NewBVMETH(common.HexToAddress(l2EthAddress), client)
	require.NoError(t, err)

	decimal, err := l2EthInstance.Decimals(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, decimal, uint8(0x12))

	symble, err := l2EthInstance.Symbol(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, symble, "WETH")

	t.Log(decimal)
	t.Log(symble)
}

func SingleWithdrawalTx(t *testing.T, withdrawalTx string) {
	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)

	withdrawalHash := common.HexToHash(withdrawalTx)

	receipt, err := waitForTransaction(withdrawalHash, l2Client, 10*time.Duration(10)*time.Second)
	require.Nil(t, err, "withdrawal initiated on L2 sequencer")

	// Transactor Account
	ethPrivKey, err := crypto.HexToECDSA(deployerPrivateKey)
	require.NoError(t, err)

	proveReceipt, finalizeReceipt := ProveAndFinalizeWithdrawalForSingleTx(t, l2url, l1Client, ethPrivKey, receipt)
	t.Logf("proveReceipt : %v , finalizeReceipt : %v", proveReceipt.TxHash, finalizeReceipt.TxHash)

}

func TestWithdrawal(t *testing.T) {
	withdrawalTx := "0xd6b71afe402197053b5934831ac1ac9ddfc540acc1017bbeec966f2287aa3a0d"
	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)

	withdrawalHash := common.HexToHash(withdrawalTx)

	receipt, err := waitForTransaction(withdrawalHash, l2Client, 10*time.Duration(10)*time.Second)
	require.Nil(t, err, "withdrawal initiated on L2 sequencer")

	// Transactor Account
	ethPrivKey, err := crypto.HexToECDSA(deployerPrivateKey)
	require.NoError(t, err)

	proveReceipt, finalizeReceipt := ProveAndFinalizeWithdrawalForSingleTx(t, l2url, l1Client, ethPrivKey, receipt)
	t.Logf("proveReceipt : %v , finalizeReceipt : %v", proveReceipt, finalizeReceipt)

}

func TestFindDepositTx(t *testing.T) {
	//l1Client, err := ethclient.Dial(l1url)
	//require.NoError(t, err)

	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)

	bn, err := l2Client.BlockNumber(context.Background())
	require.NoError(t, err)
	t.Log("now block number", bn)
	for i := int(bn); i > 0; i-- {
		block, err := l2Client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		require.NoError(t, err)
		txs := block.Transactions()
		for _, tx := range txs {
			if tx.IsDepositTx() == true && tx.IsSystemTx() == false && tx.To().Hex() == "0x4200000000000000000000000000000000000007" {
				t.Log("block info", block.Hash().Hex())
				t.Log("block height: ", block.Number().Int64())
				t.Log("find deposit tx", tx.Hash().Hex())
				t.Log("ethvalue data = ", tx.ETHValue())
				t.Log("value data = ", tx.Value())
				t.Log("mint data = ", tx.Mint())
				jsonTx, err := tx.MarshalJSON()
				require.NoError(t, err)

				t.Logf("transaction info : %v", string(jsonTx))
				_, _, err = l2Client.TransactionByHash(context.Background(), tx.Hash())
				require.NoError(t, err)

			}
		}
	}

}

func TestETHDeposit(t *testing.T) {
	t.Log("check balance.....")

	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	require.NotNil(t, l1Client)
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)
	require.NotNil(t, l2Client)

	// query eth erc20 token
	l1Bridge, err := bindings.NewL1StandardBridge(common.HexToAddress(l1BridgeAddress), l1Client)
	require.NoError(t, err)
	//require.NoError(t, err)

	// TEST deposit ETH
	t.Log("----------------")
	t.Log("ETH DEPOSIT TEST")
	t.Log("----------------")
	t.Log("ETH before deposit...\\")

	beforeBalanceL1 := getETHBalanceFromL1(t, userAddress)
	beforeBalanceL2 := getETHBalanceFromL2(t, userAddress)

	t.Log("l1 eth balance: ", beforeBalanceL1)
	t.Log("l2 eth balance: ", beforeBalanceL2)
	// do deposit
	auth := buildL1Auth(t, l1Client, userPrivateKey, big.NewInt(DECIMAL0_1))
	tx, err := l1Bridge.DepositETH(auth, 2_000_000, []byte("0x"))
	_, err = waitForTransaction(tx.Hash(), l1Client, 100*time.Second)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)

	t.Log("deposit eth tx hash is: ", tx.Hash())
	t.Log("ETH after deposit...\\")
	afterBalanceL1 := getETHBalanceFromL1(t, userAddress)
	afterBalanceL2 := getETHBalanceFromL2(t, userAddress)

	t.Log("l1 eth balance: ", afterBalanceL1)
	t.Log("l2 eth balance: ", afterBalanceL2)

	//require.Equal(t, getETHBalanceFromL2(t, userAddress), 0)
	t.Log("eth deposit amount: ", uint64(DECIMAL0_1))

	require.Equal(t, afterBalanceL2.Uint64()-beforeBalanceL2.Uint64(), uint64(DECIMAL0_1))

}

func TestFindDepositSingleTx(t *testing.T) {
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)

	bn := big.NewInt(214)
	require.NoError(t, err)
	t.Log("now block number", bn)

	block, err := l2Client.BlockByNumber(context.Background(), bn)
	require.NoError(t, err)
	txs := block.Transactions()
	for _, tx := range txs {
		//t.Log("txs hash list : ", tx.Hash())
		if tx.IsDepositTx() == true && tx.IsSystemTx() == false && tx.To().Hex() == "0x4200000000000000000000000000000000000007" {
			//t.Log("block info", block.Hash().Hex())
			t.Log("block height: ", block.Number().Int64())
			t.Log("find deposit tx", tx.Hash().Hex())
			t.Log("ethvalue data = ", tx.ETHValue())
			t.Log("tx type", tx.Type())
			jsonTx, err := tx.MarshalJSON()
			require.NoError(t, err)

			t.Logf("transaction info : %v", string(jsonTx))
			_, _, err = l2Client.TransactionByHash(context.Background(), tx.Hash())
			require.NoError(t, err)

		}
	}

}
func TestOPTx(t *testing.T) {
	l2Client, err := ethclient.Dial("https://opt-mainnet.g.alchemy.com/v2/ZyGix8nc2yGzLuS42MG_4P9af57W_1A0")
	require.NoError(t, err)

	bn := big.NewInt(214)
	require.NoError(t, err)
	t.Log("now block number", bn)

	block, err := l2Client.BlockByHash(context.Background(), common.HexToHash("0x61ecd5f18f832a47d2751d33b9e773f7856ae1bac411dc96aa1c43b1c63e5427"))
	require.NoError(t, err)
	txs := block.Transactions()

	for _, tx := range txs {
		//t.Log("txs hash list : ", tx.Hash())
		if tx.IsDepositTx() == true && tx.IsSystemTx() == false && tx.To().Hex() == "0x4200000000000000000000000000000000000007" {
			//t.Log("block info", block.Hash().Hex())
			t.Log("block height: ", block.Number().Int64())
			t.Log("find deposit tx", tx.Hash().Hex())
			t.Log("ethvalue data = ", tx.ETHValue())
			t.Log("tx type", tx.Type())
			jsonTx, err := tx.MarshalJSON()
			require.NoError(t, err)

			t.Logf("transaction info : %v", string(jsonTx))
			_, _, err = l2Client.TransactionByHash(context.Background(), tx.Hash())
			require.NoError(t, err)

		}
	}

}

func TestTx(t *testing.T) {
	l2Client, err := ethclient.Dial(l2url)
	require.NoError(t, err)

	bn := big.NewInt(214)
	require.NoError(t, err)
	t.Log("now block number", bn)
	tx, _, err := l2Client.TransactionByHash(context.Background(), common.HexToHash("0xf9bef38d02729976f8550c692a09f7270cad6769ec8e256d7058cfdbe662f55e"))
	t.Log("ethvalue data = ", tx.ETHValue())
	t.Log(tx.MarshalJSON())
}

func TestForceWithdrawal(t *testing.T) {
	l1url := "http://localhost:8545"
	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	l1ChainId, err := l1Client.ChainID(context.Background())
	require.NoError(t, err)
	pk := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	PrivKey, err := crypto.HexToECDSA(pk)
	require.NoError(t, err)
	minGasLimit := uint32(2000000)

	l2TokenAddress := common.HexToAddress(predeploys.BVM_ETH)
	forceWithdrawalAmount := big.NewInt(10000000000000)
	depositValue := big.NewInt(100)
	l1auth := buildAuth(t, l1Client, pk, depositValue, l1ChainId)
	// l1 eth value
	callerAddress := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	nullExtraData := []byte{}

	L2StandardBridgeABI, err := abi.JSON(strings.NewReader(bindings.L2StandardBridgeMetaData.ABI))
	require.NoError(t, err)
	withdrawMessage, err := L2StandardBridgeABI.Pack("withdrawTo", l2TokenAddress, callerAddress, forceWithdrawalAmount, minGasLimit, nullExtraData)
	t.Log(common.Bytes2Hex(withdrawMessage))
	require.NoError(t, err)

	zero_mnt_amount := big.NewInt(0)

	l1XDMAddress := common.HexToAddress("0xA51c1fc2f0D1a1b8494Ed1FE312d7C3a78Ed91C0")

	l1xmsg, err := bindings.NewL1CrossDomainMessenger(l1XDMAddress, l1Client)
	require.NoError(t, err)

	tx, err := l1xmsg.SendMessage(l1auth, zero_mnt_amount,
		common.HexToAddress(predeploys.L2StandardBridge),
		withdrawMessage, minGasLimit)

	//msg := ethereum.CallMsg{From: l1auth.From, To: &to, GasPrice: l1auth.GasPrice, Value: l1auth.Value, Data: sendMessageData}
	//gasLimit, err := l1Client.EstimateGas(context.Background(), msg)
	//require.NoError(t, err)

	//tx := types.NewTx(&types.LegacyTx{
	//	Nonce:    auth.Nonce.Uint64(),
	//	GasPrice: auth.GasPrice,
	//	Gas:      gasLimit,
	//	To:       &to,
	//	Value:    auth.Value,
	//	Data:     sendMessageData,
	//})
	t.Log("sendMessageHash  ", tx.Hash().Hex())
	signer := types.LatestSignerForChainID(l1ChainId)
	tx, err = types.SignTx(tx, signer, PrivKey)
	require.NoError(t, err)
	err = l1Client.SendTransaction(context.Background(), tx)
	require.NoError(t, err)
	t.Log(tx.Hash().Hex())
	time.Sleep(time.Second * 20)
	receipt, err := l1Client.TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	t.Log(tx.Hash().Hex())

	t.Log(receipt.Status)

}

func TestForceWithdrawalFromPortal(t *testing.T) {
	l1url := "http://localhost:8545"
	l1Client, err := ethclient.Dial(l1url)
	require.NoError(t, err)
	l1ChainId, err := l1Client.ChainID(context.Background())
	require.NoError(t, err)
	pk := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	//PrivKey, err := crypto.HexToECDSA(pk)
	require.NoError(t, err)
	minGasLimit := uint32(2000000)

	l2TokenAddress := common.HexToAddress(predeploys.BVM_ETH)
	forceWithdrawalAmount := big.NewInt(10000000000000)
	depositValue := big.NewInt(100)
	l1auth := buildAuth(t, l1Client, pk, depositValue, l1ChainId)
	l1auth.Value = depositValue
	// l1 eth value
	callerAddress := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	nullExtraData := []byte{}

	L2StandardBridgeABI, err := abi.JSON(strings.NewReader(bindings.L2StandardBridgeMetaData.ABI))
	require.NoError(t, err)
	withdrawMessage, err := L2StandardBridgeABI.Pack("withdrawTo", l2TokenAddress, callerAddress, forceWithdrawalAmount, minGasLimit, nullExtraData)
	t.Log(common.Bytes2Hex(withdrawMessage))
	require.NoError(t, err)

	//zero_mnt_amount := big.NewInt(0)

	OptimismPortalAddr := common.HexToAddress("0x0B306BF915C4d645ff596e518fAf3F9669b97016")
	portal, err := bindings.NewOptimismPortal(OptimismPortalAddr, l1Client)

	tx, err := portal.DepositTransaction(l1auth, big.NewInt(0), callerAddress, big.NewInt(0), uint64(minGasLimit), false,
		withdrawMessage)
	t.Log("DepositTransaction hash on L1", tx.Hash().Hex())
	time.Sleep(time.Duration(time.Second * 20))
	receipt, err := l1Client.TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	source := derive.UserDepositSource{
		L1BlockHash: receipt.BlockHash,
		LogIndex:    uint64(receipt.TransactionIndex),
	}
	depositTx := types.DepositTx{
		SourceHash:          source.SourceHash(),
		From:                callerAddress,
		To:                  &predeploys.L2StandardBridgeAddr,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 uint64(minGasLimit),
		IsSystemTransaction: false,
		EthValue:            depositValue,
		Data:                withdrawMessage,
	}
	l2DepositTx := types.NewTx(&depositTx)
	t.Log(l2DepositTx.Hash())

	//msg := ethereum.CallMsg{From: l1auth.From, To: &to, GasPrice: l1auth.GasPrice, Value: l1auth.Value, Data: sendMessageData}
	//gasLimit, err := l1Client.EstimateGas(context.Background(), msg)
	//require.NoError(t, err)

	//tx := types.NewTx(&types.LegacyTx{
	//	Nonce:    auth.Nonce.Uint64(),
	//	GasPrice: auth.GasPrice,
	//	Gas:      gasLimit,
	//	To:       &to,
	//	Value:    auth.Value,
	//	Data:     sendMessageData,
	//})
	t.Log("DepositTransaction on L2 hash  ", l2DepositTx.Hash().Hex())
	//signer := types.LatestSignerForChainID(l1ChainId)
	//tx, err = types.SignTx(tx, signer, PrivKey)
	//require.NoError(t, err)
	//err = l1Client.SendTransaction(context.Background(), tx)
	//require.NoError(t, err)
	//t.Log(tx.Hash().Hex())
	time.Sleep(time.Second * 20)
	require.NoError(t, err)
	t.Log(tx.Hash().Hex())

	t.Log(receipt.Status)

}
