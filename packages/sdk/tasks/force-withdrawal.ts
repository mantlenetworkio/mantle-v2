import {task, types} from 'hardhat/config'
import {HardhatRuntimeEnvironment} from 'hardhat/types'
import {ethers, providers, Wallet} from 'ethers'
import 'hardhat-deploy'
import '@nomiclabs/hardhat-ethers'
import {depositTx } from '@ethan-bedrock/core-utils'
import {predeploys} from "@ethan-bedrock/contracts-bedrock"
import {CrossChainMessenger, MessageStatus,} from '../src'




task('force-withdrawal', 'force withdraw')
  .addParam('l2Url', 'L2 HTTP URL', 'http://localhost:9545', types.string)
  .setAction(async (args, hre: HardhatRuntimeEnvironment) => {


    const signers = await hre.ethers.getSigners()
    if (signers.length === 0) {
      throw new Error('No configured signers')
    }
    const signer = signers[0]
    const address = await signer.getAddress()
    console.log(`Using signer: ${address}`)

    const l2Provider = new providers.StaticJsonRpcProvider(args.l2Url)
    const l2Signer = new Wallet("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291", l2Provider)

    let Deployment__L1StandardBridgeProxy = await hre.deployments.getOrNull(
      'L1StandardBridgeProxy'
    )
    if (Deployment__L1StandardBridgeProxy === undefined) {
      Deployment__L1StandardBridgeProxy = await hre.deployments.getOrNull(
        'Proxy__BVM_L1StandardBridge'
      )
    }

    let Deployment__L1CrossDomainMessengerProxy =
      await hre.deployments.getOrNull('L1CrossDomainMessengerProxy')
    if (Deployment__L1CrossDomainMessengerProxy === undefined) {
      Deployment__L1CrossDomainMessengerProxy = await hre.deployments.getOrNull(
        'Proxy__BVM_L1CrossDomainMessenger'
      )
    }

    const transferAmt = BigInt(0.01 * 1e18)

    const bedrockContracts = require("@ethan-bedrock/contracts-bedrock")
    const optimismPortalData = bedrockContracts.getContractDefinition("OptimismPortal")
    const optimismPortal = new ethers.Contract(bedrockContracts.predeploys.OptimismPortal, optimismPortalData.abi, signer)
    const txn = await optimismPortal.depositTransaction(
      bedrockContracts.predeploys.L2StandardBridge,
      transferAmt,
      1e6, false, []
    )
    const receipt = await txn.wait()


    const optimismCoreUtils = require("@ethan-bedrock/core-utils")
    const withdrawalData = new optimismCoreUtils.DepositTx({
      from: signer.address,
      to: bedrockContracts.predeploys.L2StandardBridge,
      mint: 0,
      value: ethers.BigNumber.from(transferAmt),
      ethValue: 0,
      gas: 1e6,
      isSystemTransaction: false,
      data: "",
      domain: optimismCoreUtils.SourceHashDomain.UserDeposit,
      l1BlockHash: receipt.blockHash,
      logIndex: receipt.logs[0].logIndex,
    })
    const txHash = withdrawalData.hash()



    const messenger = new CrossChainMessenger({
      l1SignerOrProvider: signer,
      l2SignerOrProvider: l2Signer,
      l1ChainId: await signer.getChainId(),
      l2ChainId: await l2Signer.getChainId(),
      bedrock: true,
    })

    console.log(`Fetching message status for ${txHash}`)
    const status = await messenger.getMessageStatus(txHash)
    console.log(`Status: ${MessageStatus[status]}`)

    if (status === MessageStatus.READY_TO_PROVE) {
      const proveTx = await messenger.proveMessage(txHash)
      const proveReceipt = await proveTx.wait()
      console.log('Prove receipt', proveReceipt)

      const finalizeInterval = setInterval(async () => {
        const currentStatus = await messenger.getMessageStatus(txHash)
        console.log(`Message status: ${MessageStatus[currentStatus]}`)
      }, 3000)

      try {
        await messenger.waitForMessageStatus(
          txHash,
          MessageStatus.READY_FOR_RELAY
        )
      } finally {
        clearInterval(finalizeInterval)
      }

      const tx = await messenger.finalizeMessage(txHash)
      const receipt = await tx.wait()
      console.log(receipt)
      console.log('Finalized withdrawal')
    }
  })
