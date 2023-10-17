import { DeployFunction } from 'hardhat-deploy/dist/types'

import {
  assertContractVariable,
  deploy, deploySleepTime,
  getContractFromArtifact,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const L1StandardBridgeProxy = await getContractFromArtifact(
    hre,
    'Proxy__BVM_L1StandardBridge'
  )

  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'OptimismMintableERC20Factory',
    args: [L1StandardBridgeProxy.address],
    postDeployAction: async (contract) => {
      await assertContractVariable(
        contract,
        'BRIDGE',
        L1StandardBridgeProxy.address
      )
    },
  })
}

deployFn.tags = ['OptimismMintableERC20FactoryImpl', 'setup', 'l1']

export default deployFn
