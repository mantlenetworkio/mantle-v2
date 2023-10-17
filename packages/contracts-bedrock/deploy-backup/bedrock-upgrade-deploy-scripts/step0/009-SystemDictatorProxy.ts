import { DeployFunction } from 'hardhat-deploy/dist/types'

import {assertContractVariable, deploy, deploySleepTime} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const { deployer } = await hre.getNamedAccounts()

  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'SystemDictatorProxy',
    contract: 'Proxy',
    args: [deployer],
    postDeployAction: async (contract) => {
      await assertContractVariable(contract, 'admin', deployer)
    },
  })
}

deployFn.tags = ['SystemDictatorProxy', 'setup', 'l1']

export default deployFn
