import { DeployFunction } from 'hardhat-deploy/dist/types'

import { assertContractVariable, deploy } from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const { deployer } = await hre.getNamedAccounts()

  await sleep(5000)
  await deploy({
    hre,
    name: 'L1ERC721BridgeProxy',
    contract: 'Proxy',
    args: [deployer],
    postDeployAction: async (contract) => {
      await assertContractVariable(contract, 'admin', deployer)
    },
  })
}

deployFn.tags = ['L1ERC721BridgeProxy', 'setup', 'l1']

export default deployFn
