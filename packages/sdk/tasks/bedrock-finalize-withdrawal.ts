import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'
import { Wallet, providers } from 'ethers'
import 'hardhat-deploy'
import '@nomiclabs/hardhat-ethers'

import {
  CrossChainMessenger,
  MessageStatus,
} from '../src'

task('bedrock-finalize-withdrawal', 'Finalize a withdrawal')
  .addParam(
    'transactionHash',
    'L2 Transaction hash to finalize',
    '0x414a8ffc5fb69943ee5fa5cc745e365ba5501a498eef9d73e15cda7184232219',
    types.string
  )
  .addParam('l2Url', 'L2 HTTP URL', 'http://localhost:9545', types.string)
  .setAction(async (args, hre: HardhatRuntimeEnvironment) => {
    const txHash = args.transactionHash
    if (txHash === '') {
      console.log('No tx hash')
    }

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
