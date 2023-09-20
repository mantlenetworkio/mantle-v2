import { DeployFunction } from 'hardhat-deploy/dist/types'

import {
  assertContractVariable,
  deploy,
  getContractFromArtifact,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const OptimismPortalProxy = await getContractFromArtifact(
    hre,
    'OptimismPortalProxy'
  )
  const Proxy__L1MantleToken = await getContractFromArtifact(
    hre,
    'Proxy__L1MantleToken'
  )

  await sleep(6000)
  await deploy({
    hre,
    name: 'L1CrossDomainMessenger',
    args: [OptimismPortalProxy.address, Proxy__L1MantleToken.address],
    postDeployAction: async (contract) => {
      await assertContractVariable(
        contract,
        'PORTAL',
        OptimismPortalProxy.address
      )
      await assertContractVariable(
        contract,
        'L1_MNT_ADDRESS',
        Proxy__L1MantleToken.address
      )
    },
  })
}

deployFn.tags = ['L1CrossDomainMessengerImpl', 'setup', 'l1']

export default deployFn
