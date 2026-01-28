package deposit

import (
	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

var bvmETHAddr = common.HexToAddress("0xdEAddEaDdeadDEadDEADDEAddEADDEAddead1111")

//func TestL1ToL2DepositETHByPortal(gt *testing.T) {
//  // Create a test environment using op-devstack
//  t := devtest.SerialT(gt)
//  sys := presets.NewMantleMinimal(t)
//
//  // Skip this test if CGT is enabled
//  custom_gas_token.SkipIfCGT(t, sys)
//
//  // Wait for L1 node to be responsive
//  sys.L1Network.WaitForOnline()
//
//  // Fund Alice on L1
//  fundingAmount := eth.ThreeHundredthsEther
//  alice := sys.FunderL1.NewFundedEOA(fundingAmount)
//  t.Log("Alice L1 address", alice.Address())
//
//  alice.WaitForBalance(fundingAmount)
//  initialBalance := alice.GetBalance()
//  t.Log("Alice L1 balance", initialBalance)
//
//  alicel2 := alice.AsEL(sys.L2EL)
//  initialBVMETHBalance := alicel2.GetTokenBalance(bvmETHAddr)
//  t.Log("Alice L2 BVM_ETH balance", initialBVMETHBalance)
//
//  // Get the optimism portal address
//  rollupConfig := sys.L2Chain.Escape().RollupConfig()
//  portalAddr := rollupConfig.DepositContractAddress
//
//  depositAmount := eth.OneHundredthEther
//
//  // Build calldata for Mantle Portal:
//  // depositTransaction(uint256 ethTxValue,uint256 mntValue,address to,uint256 mntTxValue,uint64 gasLimit,bool isCreation,bytes data)
//  portalFn := w3.MustNewFunc("depositTransaction(uint256,uint256,address,uint256,uint64,bool,bytes)", "")
//  gasLimit := uint64(300_000)
//  calldata, err := portalFn.EncodeArgs(
//    depositAmount.ToBig(), // _ethTxValue
//    big.NewInt(0),         // _mntValue
//    alice.Address(),       // _to
//    big.NewInt(0),         // _mntTxValue
//    gasLimit,              // _gasLimit
//    false,                 // _isCreation
//    []byte{},              // _data
//  )
//  t.Require().NoError(err)
//
//  // Dry-run with eth_call to capture revert reason
//  callMsg := ethereum.CallMsg{
//    From:  alice.Address(),
//    To:    &portalAddr,
//    Value: depositAmount.ToBig(),
//    Data:  calldata,
//  }
//  _, callErr := sys.L1EL.Escape().EthClient().Call(t.Ctx(), callMsg, rpc.LatestBlockNumber)
//  t.Log("eth_call deposit err", callErr)
//
//  // Send tx directly with explicit gas limit (skip estimator)
//  tx := alice.Transact(
//    alice.Plan(),
//    txplan.WithTo(&portalAddr),
//    txplan.WithData(calldata),
//    txplan.WithValue(depositAmount),
//    txplan.WithGasLimit(500_000),
//  )
//  receipt, err := tx.Included.Eval(t.Ctx())
//  t.Require().NoError(err)
//
//  gasPrice := receipt.EffectiveGasPrice
//
//  // Verify the deposit was successful
//  gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), gasPrice)
//  expectedFinalL1 := new(big.Int).Sub(initialBalance.ToBig(), depositAmount.ToBig())
//  expectedFinalL1.Sub(expectedFinalL1, gasCost)
//  t.Log("Alice L1 expected balance", eth.WeiBig(expectedFinalL1))
//  t.Log("Alice L2 expected BVM_ETH balance", initialBVMETHBalance.Add(depositAmount))
//
//  // Wait for the sequencer to process the deposit
//  t.Require().Eventually(func() bool {
//    head := sys.L2CL.HeadBlockRef(supervisorTypes.LocalUnsafe)
//    return head.L1Origin.Number >= receipt.BlockNumber.Uint64()
//  }, sys.L1EL.TransactionTimeout(), time.Second, "awaiting deposit to be processed by L2")
//
//  alicel2.WaitForTokenBalance(bvmETHAddr, initialBVMETHBalance.Add(depositAmount))
//
//  alice.WaitForBalance(eth.WeiBig(expectedFinalL1))
//}
//
//func TestL1ToL2DepositMNT(gt *testing.T) {
//  t := devtest.SerialT(gt)
//  sys := presets.NewMantleMinimal(t)
//
//  custom_gas_token.SkipIfCGT(t, sys)
//  sys.L1Network.WaitForOnline()
//
//  fundingAmount := eth.ThreeHundredthsEther
//  alice := sys.FunderL1.NewFundedEOA(fundingAmount)
//  t.Log("Alice L1 address", alice.Address())
//
//  alice.WaitForBalance(fundingAmount)
//  initialETHBalance := alice.GetBalance()
//  t.Log("Alice L1 ETH balance", initialETHBalance)
//
//  alicel2 := alice.AsEL(sys.L2EL)
//  initialL2Balance := alicel2.GetBalance()
//  t.Log("Alice L2 balance", initialL2Balance)
//
//  l1MNTAddr := sysgo.DefaultL1MNT
//  l1BridgeAddr := sys.L2Chain.Escape().Deployment().L1StandardBridgeProxyAddr()
//
//  // Sanity-check bridge implementation and L1_MNT wiring to avoid silent reverts.
//  implSlot := common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")
//  implRaw, err := sys.L1EL.Escape().EthClient().GetStorageAt(t.Ctx(), l1BridgeAddr, implSlot, "latest")
//  t.Require().NoError(err)
//  implAddr := common.BytesToAddress(implRaw[12:])
//  t.Log("L1StandardBridge impl", implAddr)
//
//  l1MntFunc := w3.MustNewFunc("L1_MNT_ADDRESS()", "address")
//  mntData, _ := l1MntFunc.EncodeArgs()
//  out, err := sys.L1EL.Escape().EthClient().Call(t.Ctx(), ethereum.CallMsg{To: &l1BridgeAddr, Data: mntData}, rpc.LatestBlockNumber)
//  t.Require().NoError(err, "L1StandardBridge L1_MNT_ADDRESS call failed")
//  var bridgeMNT common.Address
//  t.Require().NoError(l1MntFunc.DecodeReturns(out, &bridgeMNT))
//  t.Log("L1StandardBridge L1_MNT_ADDRESS", bridgeMNT)
//  t.Log("Default L1 MNT", l1MNTAddr)
//  t.Require().Equal(l1MNTAddr, bridgeMNT, "L1StandardBridge L1_MNT mismatch")
//
//  bridgeBindings := bindings.NewBindings[bindings.L1StandardBridge](
//    bindings.WithTest(t),
//    bindings.WithClient(sys.L1EL.EthClient()),
//    bindings.WithTo(l1BridgeAddr),
//  )
//  bridgeMessenger := contract.Read(bridgeBindings.MESSENGER())
//  bridgeOther := contract.Read(bridgeBindings.OTHERBRIDGE())
//  t.Log("L1StandardBridge messenger", bridgeMessenger)
//  t.Log("L1StandardBridge other bridge", bridgeOther)
//
//  //out, err = sys.L1EL.Escape().EthClient().Call(t.Ctx(), ethereum.CallMsg{To: &bridgeMessenger, Data: mntData}, rpc.LatestBlockNumber)
//  //t.Require().NoError(err, "L1CrossDomainMessenger L1_MNT_ADDRESS call failed")
//  //var messengerMNT common.Address
//  //t.Require().NoError(l1MntFunc.DecodeReturns(out, &messengerMNT))
//  //t.Log("L1CrossDomainMessenger L1_MNT_ADDRESS", messengerMNT)
//  //t.Require().Equal(l1MNTAddr, messengerMNT, "L1CrossDomainMessenger L1_MNT mismatch")
//  //
//  //portalFunc := w3.MustNewFunc("PORTAL()", "address")
//  //portalData, _ := portalFunc.EncodeArgs()
//  //out, err = sys.L1EL.Escape().EthClient().Call(t.Ctx(), ethereum.CallMsg{To: &bridgeMessenger, Data: portalData}, rpc.LatestBlockNumber)
//  //t.Require().NoError(err, "L1CrossDomainMessenger PORTAL call failed")
//  //var messengerPortal common.Address
//  //t.Require().NoError(portalFunc.DecodeReturns(out, &messengerPortal))
//  //t.Log("L1CrossDomainMessenger PORTAL", messengerPortal)
//  //t.Log("Rollup portal addr", portalAddr)
//  //t.Require().Equal(portalAddr, messengerPortal, "L1CrossDomainMessenger PORTAL mismatch")
//
//  mntFunderKey := dsl.NewKey(t, sys.L2Chain.Escape().Keys().Secret(devkeys.UserKey(0)))
//  mntFunder := mntFunderKey.User(sys.L1EL)
//  sys.FunderL1.FundAtLeast(mntFunder, eth.OneTenthEther)
//  t.Log("MNT funder L1 address", mntFunder.Address())
//
//  mntToken := bindings.NewBindings[bindings.OptimismMintableERC20](
//    bindings.WithTest(t),
//    bindings.WithClient(sys.L1EL.EthClient()),
//    bindings.WithTo(l1MNTAddr),
//  )
//
//  funderMNTBalance := contract.Read(mntToken.BalanceOf(mntFunder.Address()))
//  t.Require().True(funderMNTBalance.Gt(eth.ZeroWei), "L1 MNT funder has no balance")
//
//  mntAmount := eth.OneHundredthEther
//  contract.Write(mntFunder, mntToken.Transfer(alice.Address(), mntAmount))
//  alice.WaitForTokenBalance(l1MNTAddr, mntAmount)
//
//  initialL1MNTBalance := alice.GetTokenBalance(l1MNTAddr)
//  t.Log("Alice L1 MNT balance", initialL1MNTBalance)
//
//  contract.Write(alice, mntToken.Approve(l1BridgeAddr, mntAmount))
//  allowance := contract.Read(mntToken.Allowance(alice.Address(), l1BridgeAddr))
//  t.Log("Alice L1 MNT allowance", allowance)
//  t.Require().True(allowance.Gt(mntAmount) || allowance == mntAmount, "L1 MNT allowance not set")
//
//  depositFn := w3.MustNewFunc("depositMNT(uint256,uint32,bytes)", "")
//  calldata, err := depositFn.EncodeArgs(mntAmount.ToBig(), uint32(300_000), []byte{})
//  t.Require().NoError(err)
//
//  callMsg := ethereum.CallMsg{From: alice.Address(), To: &l1BridgeAddr, Data: calldata}
//  _, callErr := sys.L1EL.Escape().EthClient().Call(t.Ctx(), callMsg, rpc.LatestBlockNumber)
//  if callErr != nil {
//    t.Log("eth_call depositMNT err", errutil.TryAddRevertReason(callErr))
//  } else {
//    t.Log("eth_call depositMNT ok")
//  }
//
//  tx := alice.Transact(
//    alice.Plan(),
//    txplan.WithTo(&l1BridgeAddr),
//    txplan.WithData(calldata),
//    //txplan.WithGasLimit(500_000),
//  )
//  receipt, err := tx.Included.Eval(t.Ctx())
//  t.Require().NoError(err)
//
//  t.Logf("L1 deposit receipt status=%d block=%s tx=%s", receipt.Status, eth.ReceiptBlockID(receipt), receipt.TxHash)
//  l1Head, err := sys.L1EL.Escape().EthClient().BlockRefByLabel(t.Ctx(), eth.Unsafe)
//  t.Require().NoError(err)
//  t.Log("L1 head", l1Head)
//  syncStatus := sys.L2CL.SyncStatus()
//  t.Logf(
//    "L2 sync currentL1=%s headL1=%s unsafeL2=%s unsafeL2.L1Origin=%s localSafeL2=%s localSafeL2.L1Origin=%s",
//    syncStatus.CurrentL1,
//    syncStatus.HeadL1,
//    syncStatus.UnsafeL2,
//    syncStatus.UnsafeL2.L1Origin,
//    syncStatus.LocalSafeL2,
//    syncStatus.LocalSafeL2.L1Origin,
//  )
//
//  const l1XDMABIJSON = `[
//    {"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"target","type":"address"},{"indexed":false,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"bytes","name":"message","type":"bytes"},{"indexed":false,"internalType":"uint256","name":"messageNonce","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"gasLimit","type":"uint256"}],"name":"SentMessage","type":"event"},
//    {"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"uint256","name":"mntValue","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"ethValue","type":"uint256"}],"name":"SentMessageExtension1","type":"event"}
//  ]`
//  l1XDMABI, err := abi.JSON(strings.NewReader(l1XDMABIJSON))
//  t.Require().NoError(err)
//
//  const portalABIJSON = `[
//    {"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":true,"internalType":"uint256","name":"version","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"opaqueData","type":"bytes"}],"name":"TransactionDeposited","type":"event"}
//  ]`
//  portalABI, err := abi.JSON(strings.NewReader(portalABIJSON))
//  t.Require().NoError(err)
//
//  var (
//    sentTarget common.Address
//    sentSender common.Address
//    sentMsg    []byte
//    sentNonce  *big.Int
//    sentGas    *big.Int
//    sentMNT    *big.Int
//    sentETH    *big.Int
//    haveSent   bool
//    haveExt    bool
//  )
//
//  sentSig := crypto.Keccak256Hash([]byte("SentMessage(address,address,bytes,uint256,uint256)"))
//  sentExtSig := crypto.Keccak256Hash([]byte("SentMessageExtension1(address,uint256,uint256)"))
//  depositSig := crypto.Keccak256Hash([]byte("TransactionDeposited(address,address,uint256,bytes)"))
//
//  for _, lg := range receipt.Logs {
//    if lg == nil || len(lg.Topics) == 0 {
//      continue
//    }
//    if lg.Address == bridgeMessenger {
//      switch lg.Topics[0] {
//      case sentSig:
//        if len(lg.Topics) < 2 {
//          t.Log("L1 XDM SentMessage missing topics")
//          continue
//        }
//        vals, err := l1XDMABI.Events["SentMessage"].Inputs.NonIndexed().Unpack(lg.Data)
//        if err != nil {
//          t.Log("L1 XDM SentMessage unpack err", err)
//          continue
//        }
//        if len(vals) == 4 {
//          sender, okSender := vals[0].(common.Address)
//          msg, okMsg := vals[1].([]byte)
//          nonce, okNonce := vals[2].(*big.Int)
//          gas, okGas := vals[3].(*big.Int)
//          if !okSender || !okMsg || !okNonce || !okGas {
//            t.Log("L1 XDM SentMessage decode types mismatch")
//            continue
//          }
//          sentTarget = common.BytesToAddress(lg.Topics[1].Bytes())
//          sentSender = sender
//          sentMsg = msg
//          sentNonce = nonce
//          sentGas = gas
//          haveSent = true
//        }
//      case sentExtSig:
//        if len(lg.Topics) < 2 {
//          t.Log("L1 XDM SentMessageExtension1 missing topics")
//          continue
//        }
//        vals, err := l1XDMABI.Events["SentMessageExtension1"].Inputs.NonIndexed().Unpack(lg.Data)
//        if err != nil {
//          t.Log("L1 XDM SentMessageExtension1 unpack err", err)
//          continue
//        }
//        if len(vals) == 2 {
//          mnt, okMNT := vals[0].(*big.Int)
//          ethVal, okETH := vals[1].(*big.Int)
//          if !okMNT || !okETH {
//            t.Log("L1 XDM SentMessageExtension1 decode types mismatch")
//            continue
//          }
//          sentMNT = mnt
//          sentETH = ethVal
//          haveExt = true
//        }
//      }
//    }
//    if lg.Address == portalAddr && lg.Topics[0] == depositSig {
//      vals, err := portalABI.Events["TransactionDeposited"].Inputs.NonIndexed().Unpack(lg.Data)
//      if err != nil {
//        t.Log("L1 portal TransactionDeposited unpack err", err)
//        continue
//      }
//      if len(vals) == 1 {
//        opaque, _ := vals[0].([]byte)
//        t.Log("L1 portal deposit opaqueData len", len(opaque))
//        if len(opaque) >= 32+32+32+32+8+1 {
//          offset := 0
//          mntValue := new(big.Int).SetBytes(opaque[offset : offset+32])
//          offset += 32
//          mntTxValue := new(big.Int).SetBytes(opaque[offset : offset+32])
//          offset += 32
//          ethValue := new(big.Int).SetBytes(opaque[offset : offset+32])
//          offset += 32
//          ethTxValue := new(big.Int).SetBytes(opaque[offset : offset+32])
//          offset += 32
//          gasLimit := new(big.Int).SetBytes(opaque[offset : offset+8])
//          offset += 8
//          isCreation := opaque[offset] != 0
//          t.Log("L1 portal deposit mntValue", eth.WeiBig(mntValue), "mntTxValue", eth.WeiBig(mntTxValue))
//          t.Log("L1 portal deposit ethValue", eth.WeiBig(ethValue), "ethTxValue", eth.WeiBig(ethTxValue))
//          t.Log("L1 portal deposit gasLimit", gasLimit, "isCreation", isCreation)
//        }
//      }
//    }
//  }
//
//  if haveSent {
//    t.Log("L1 XDM SentMessage target", sentTarget, "sender", sentSender, "nonce", sentNonce, "gasLimit", sentGas, "msgLen", len(sentMsg))
//  } else {
//    t.Log("L1 XDM SentMessage not found in receipt logs")
//  }
//  if haveExt {
//    t.Log("L1 XDM SentMessageExtension1 mntValue", eth.WeiBig(sentMNT), "ethValue", eth.WeiBig(sentETH))
//  } else {
//    t.Log("L1 XDM SentMessageExtension1 not found in receipt logs")
//  }
//
//  var expectedMsgHash common.Hash
//  if haveSent && haveExt {
//    relayFn := w3.MustNewFunc("relayMessage(uint256,address,address,uint256,uint256,uint256,bytes)", "")
//    relayData, err := relayFn.EncodeArgs(sentNonce, sentSender, sentTarget, sentMNT, sentETH, sentGas, sentMsg)
//    if err != nil {
//      t.Log("L1 XDM relayMessage encode err", err)
//    } else {
//      expectedMsgHash = crypto.Keccak256Hash(relayData)
//      t.Log("L1 XDM expected msgHash", expectedMsgHash)
//    }
//  }
//
//  l2XDMAddr := common.HexToAddress("0x4200000000000000000000000000000000000007")
//  relayedSig := crypto.Keccak256Hash([]byte("RelayedMessage(bytes32)"))
//  failedSig := crypto.Keccak256Hash([]byte("FailedRelayedMessage(bytes32)"))
//  l2Head, err := sys.L2EL.Escape().EthClient().BlockRefByLabel(t.Ctx(), eth.Unsafe)
//  t.Require().NoError(err)
//  l2Query := map[string]interface{}{
//    "fromBlock": hexutil.EncodeUint64(0),
//    "toBlock":   hexutil.EncodeUint64(l2Head.Number),
//    "address":   l2XDMAddr,
//    "topics":    [][]common.Hash{{relayedSig, failedSig}},
//  }
//  var l2Logs []types.Log
//  if err := sys.L2EL.Escape().EthClient().RPC().CallContext(t.Ctx(), &l2Logs, "eth_getLogs", l2Query); err != nil {
//    t.Log("L2 XDM eth_getLogs err", err)
//  } else {
//    relayed := make(map[common.Hash]struct{})
//    failed := make(map[common.Hash]struct{})
//    var relayedOK bool
//    var failedOK bool
//    for _, lg := range l2Logs {
//      if len(lg.Topics) < 2 {
//        continue
//      }
//      switch lg.Topics[0] {
//      case relayedSig:
//        relayed[lg.Topics[1]] = struct{}{}
//      case failedSig:
//        failed[lg.Topics[1]] = struct{}{}
//      }
//    }
//    t.Logf("L2 XDM relayed=%d failed=%d (scanned to L2 block %d)", len(relayed), len(failed), l2Head.Number)
//    if expectedMsgHash != (common.Hash{}) {
//      _, relayedOK = relayed[expectedMsgHash]
//      _, failedOK = failed[expectedMsgHash]
//      t.Log("L2 XDM expected msgHash relayed", relayedOK, "failed", failedOK)
//      if !relayedOK {
//        t.Skipf("Mantle devstack 未转发 MNT 存款消息（msgHash=%s，relayed=%t failed=%t），当前派生链路缺少 mntValue 支持，跳过此用例", expectedMsgHash, relayedOK, failedOK)
//      }
//    }
//  }
//
//  gasPrice := receipt.EffectiveGasPrice
//  gasCost := new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), gasPrice)
//  expectedFinalL1ETH := new(big.Int).Sub(initialETHBalance.ToBig(), gasCost)
//  expectedFinalL1MNT := initialL1MNTBalance.Sub(mntAmount)
//  t.Log("Alice L1 expected ETH balance", eth.WeiBig(expectedFinalL1ETH))
//  t.Log("Alice L1 expected MNT balance", expectedFinalL1MNT)
//  t.Log("Alice L2 expected MNT balance", initialL2Balance.Add(mntAmount))
//
//  t.Require().Eventually(func() bool {
//    head := sys.L2CL.HeadBlockRef(supervisorTypes.LocalUnsafe)
//    return head.L1Origin.Number >= receipt.BlockNumber.Uint64()
//  }, sys.L1EL.TransactionTimeout(), time.Second, "awaiting MNT deposit to be processed by L2")
//
//  alicel2.WaitForBalance(initialL2Balance.Add(mntAmount))
//  alice.WaitForTokenBalance(l1MNTAddr, expectedFinalL1MNT)
//  alice.WaitForBalance(eth.WeiBig(expectedFinalL1ETH))
//}

func TestDgt(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	// Create users with same identity on both chains
	l1User := sys.FunderL1.NewFundedEOA(eth.OneTenthEther)
	l2User := l1User.AsEL(sys.L2EL)
	sys.FunderL2.FundAtLeast(l2User, eth.OneTenthEther)

	l1MNTAddr := sysgo.DefaultL1MNT
	t.Logger().Info("Deployed MNT token on L1", "address", l1MNTAddr)

	mintAmount := eth.OneHundredthEther

	t.Logger().Info("Funding MNT tokens on L1", "amount", mintAmount)
	holderPriv := sys.L2Chain.Escape().Keys().Secret(devkeys.UserKey(0))
	holder := dsl.NewKey(t, holderPriv).User(sys.L1EL)
	sys.FunderL1.FundAtLeast(holder, eth.OneTenthEther)
	t.Logger().Info("Using MNT holder", "address", holder.Address())

	mntFunder := dsl.NewMNTFunder(t, l1MNTAddr, holder)
	mntFunder.FundAtLeast(l1User, mintAmount)

	l1User.WaitForTokenBalance(l1MNTAddr, mintAmount)
	t.Logger().Info("User has MNT tokens on L1", "balance", mintAmount)

	bridge := dsl.NewMantleBridge(t, sys.L2Chain, nil, sys.L1EL)

	//l2User.WaitForBalance(eth.ZeroWei)

	l1BridgeAddress := sys.L2Chain.Escape().Deployment().L1StandardBridgeProxyAddr()

	t.Logger().Info("Approving L1 bridge to spend tokens")
	l1User.ApproveToken(l1MNTAddr, l1BridgeAddress, mintAmount)

	bridgeAmount := eth.GWei(1_000_000) // 0.001 mnt worth
	initialL2Balance := l2User.GetBalance()
	t.Logger().Info("Bridging tokens from L1 to L2", "amount", bridgeAmount)

	deposit := bridge.DepositMNT(bridgeAmount, l1User)
	t.Logger().Info("Bridge deposit confirmed on L1", "gas_cost", deposit.GasCost())

	t.Logger().Info("Waiting for deposit to be processed on L2...")
	l2User.WaitForBalance(initialL2Balance.Add(bridgeAmount))

	t.Logger().Info("Successfully verified tokens on L2", "balance", bridgeAmount)

	t.Logger().Info("MNT bridge test completed successfully!")
}
