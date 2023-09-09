import { DeployFunction } from 'hardhat-deploy/dist/types'

import { assertContractVariable, deploy } from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const { deployer } = await hre.getNamedAccounts()

  await sleep(6000)
  await deploy({
    hre,
    name: 'Lib_AddressManager',
    contract: 'AddressManager',
    args: [],
    postDeployAction: async (contract) => {
      // Owner is temporarily set to the deployer.
      await assertContractVariable(contract, 'owner', deployer)
    },
  })
}

deployFn.tags = ['AddressManager', 'setup', 'l1']

export default deployFn
