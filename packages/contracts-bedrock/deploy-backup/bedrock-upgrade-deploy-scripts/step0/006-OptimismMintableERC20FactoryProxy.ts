import { DeployFunction } from 'hardhat-deploy/dist/types'

import {
  assertContractVariable,
  deploy, deploySleepTime,
  getDeploymentAddress,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const proxyAdmin = await getDeploymentAddress(hre, 'ProxyAdmin')

  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'OptimismMintableERC20FactoryProxy',
    contract: 'Proxy',
    args: [proxyAdmin],
    postDeployAction: async (contract) => {
      await assertContractVariable(contract, 'admin', proxyAdmin)
    },
  })
}

deployFn.tags = ['OptimismMintableERC20FactoryProxy', 'setup', 'l1']

export default deployFn
