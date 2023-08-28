import assert from 'assert'

import { ethers } from 'ethers'
import { DeployFunction } from 'hardhat-deploy/dist/types'
import { awaitCondition } from '@eth-optimism/core-utils'
import '@eth-optimism/hardhat-deploy-config'
import 'hardhat-deploy'
import '@nomiclabs/hardhat-ethers'

import {
  assertContractVariable,
  getContractsFromArtifacts,
  printJsonTransaction,
  isStep,
  printTenderlySimulationLink,
  printCastCommand,
  liveDeployer,
  doPhase,
  isStartOfPhase,
} from '../src/deploy-utils'

const deployFn: DeployFunction = async (hre) => {
  const { deployer } = await hre.getNamedAccounts()

  // Set up required contract references.
  const [
    SystemDictator,
    ProxyAdmin,
    AddressManager,
    L1CrossDomainMessenger,
    L1StandardBridgeProxy,
    L1StandardBridge,
    L2OutputOracle,
    OptimismPortal,
    OptimismMintableERC20Factory,
    L1ERC721BridgeProxy,
    L1ERC721Bridge,
  ] = await getContractsFromArtifacts(hre, [
    {
      name: 'SystemDictatorProxy',
      iface: 'SystemDictator',
      signerOrProvider: deployer,
    },
    {
      name: 'ProxyAdmin',
      signerOrProvider: deployer,
    },
    {
      name: 'Lib_AddressManager',
      signerOrProvider: deployer,
    },
    {
      name: 'Proxy__OVM_L1CrossDomainMessenger',
      iface: 'L1CrossDomainMessenger',
      signerOrProvider: deployer,
    },
    {
      name: 'Proxy__OVM_L1StandardBridge',
    },
    {
      name: 'Proxy__OVM_L1StandardBridge',
      iface: 'L1StandardBridge',
      signerOrProvider: deployer,
    },
    {
      name: 'L2OutputOracleProxy',
      iface: 'L2OutputOracle',
      signerOrProvider: deployer,
    },
    {
      name: 'OptimismPortalProxy',
      iface: 'OptimismPortal',
      signerOrProvider: deployer,
    },
    {
      name: 'OptimismMintableERC20FactoryProxy',
      iface: 'OptimismMintableERC20Factory',
      signerOrProvider: deployer,
    },
    {
      name: 'L1ERC721BridgeProxy',
    },
    {
      name: 'L1ERC721BridgeProxy',
      iface: 'L1ERC721Bridge',
      signerOrProvider: deployer,
    },
  ])

  // If we have the key for the controller then we don't need to wait for external txns.
  // Set the DISABLE_LIVE_DEPLOYER=true in the env to ensure the script will pause to simulate scenarios
  // where the controller is not the deployer.
  const isLiveDeployer = await liveDeployer({
    hre,
    disabled: process.env.DISABLE_LIVE_DEPLOYER,
  })

  // Make sure the dynamic system configuration has been set.
  if (
    (await isStartOfPhase(SystemDictator, 2)) &&
    !(await SystemDictator.dynamicConfigSet())
  ) {
    console.log(`
      You must now set the dynamic L2OutputOracle configuration by calling the function
      updateL2OutputOracleDynamicConfig. You will need to provide the
      l2OutputOracleStartingBlockNumber and the l2OutputOracleStartingTimestamp which can both be
      found by querying the last finalized block in the L2 node.
    `)

    if (isLiveDeployer) {
      console.log(`Updating dynamic oracle config...`)

      // Use default starting time if not provided
      let deployL2StartingTimestamp =
        hre.deployConfig.l2OutputOracleStartingTimestamp
      if (deployL2StartingTimestamp < 0) {
        const l1StartingBlock = await hre.ethers.provider.getBlock(
          hre.deployConfig.l1StartingBlockTag
        )
        if (l1StartingBlock === null) {
          throw new Error(
            `Cannot fetch block tag ${hre.deployConfig.l1StartingBlockTag}`
          )
        }
        deployL2StartingTimestamp = l1StartingBlock.timestamp
      }

      await SystemDictator.updateDynamicConfig(
        {
          l2OutputOracleStartingBlockNumber:
          hre.deployConfig.l2OutputOracleStartingBlockNumber,
          l2OutputOracleStartingTimestamp: deployL2StartingTimestamp,
        },
        false // do not pause the the OptimismPortal when initializing
      )
    } else {
      // pause the OptimismPortal when initializing
      const optimismPortalPaused = true
      const tx = await SystemDictator.populateTransaction.updateDynamicConfig(
        {
          l2OutputOracleStartingBlockNumber:
          hre.deployConfig.l2OutputOracleStartingBlockNumber,
          l2OutputOracleStartingTimestamp:
          hre.deployConfig.l2OutputOracleStartingTimestamp,
        },
        optimismPortalPaused
      )
      console.log(`Please update dynamic oracle config...`)
      console.log(
        JSON.stringify(
          {
            l2OutputOracleStartingBlockNumber:
            hre.deployConfig.l2OutputOracleStartingBlockNumber,
            l2OutputOracleStartingTimestamp:
            hre.deployConfig.l2OutputOracleStartingTimestamp,
            optimismPortalPaused,
          },
          null,
          2
        )
      )
      console.log(`MSD address: ${SystemDictator.address}`)
      printJsonTransaction(tx)
      printCastCommand(tx)
      await printTenderlySimulationLink(SystemDictator.provider, tx)
    }

    await awaitCondition(
      async () => {
        return SystemDictator.dynamicConfigSet()
      },
      5000,
      1000
    )
  }
}

deployFn.tags = ['SystemDictatorSteps', 'update', 'dynamicconfig']

export default deployFn
