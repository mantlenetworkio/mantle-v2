import { DeployFunction } from 'hardhat-deploy/dist/types'

import {
  assertContractVariable,
  deploy, deploySleepTime,
  getContractFromArtifact,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const OptimismPortalProxy = await getContractFromArtifact(
    hre,
    'OptimismPortalProxy'
  )
  const l1MantleToken = hre.deployConfig.l1MantleToken
  if (l1MantleToken === "0x000000000000000000000000000000000000000000000000") {
    throw new Error(`l1 mantle token is empty`)
  }

  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'L1CrossDomainMessenger',
    args: [OptimismPortalProxy.address, l1MantleToken],
    postDeployAction: async (contract) => {
      await assertContractVariable(
        contract,
        'PORTAL',
        OptimismPortalProxy.address
      )
      await assertContractVariable(
        contract,
        'L1_MNT_ADDRESS',
        l1MantleToken
      )
    },
  })
}

deployFn.tags = ['L1CrossDomainMessengerImpl', 'setup', 'l1']

export default deployFn
