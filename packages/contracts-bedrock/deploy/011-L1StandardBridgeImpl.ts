import { DeployFunction } from 'hardhat-deploy/dist/types'

import { predeploys } from '../src'
import {
  assertContractVariable,
  deploy,
  getContractFromArtifact,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const L1CrossDomainMessengerProxy = await getContractFromArtifact(
    hre,
    'Proxy__OVM_L1CrossDomainMessenger'
  )
  const Proxy__L1MantleToken = await getContractFromArtifact(
    hre,
    'Proxy__L1MantleToken'
  )

  await sleep(6000)
  await deploy({
    hre,
    name: 'L1StandardBridge',
    args: [L1CrossDomainMessengerProxy.address, Proxy__L1MantleToken.address],
    postDeployAction: async (contract) => {
      await assertContractVariable(
        contract,
        'MESSENGER',
        L1CrossDomainMessengerProxy.address
      )
      await assertContractVariable(
        contract,
        'OTHER_BRIDGE',
        predeploys.L2StandardBridge
      )
      await assertContractVariable(
        contract,
        'L1_MNT_ADDRESS',
        Proxy__L1MantleToken.address
      )
    },
  })
}

deployFn.tags = ['L1StandardBridgeImpl', 'setup', 'l1']

export default deployFn
