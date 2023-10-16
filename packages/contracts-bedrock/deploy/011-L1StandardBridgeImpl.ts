import { DeployFunction } from 'hardhat-deploy/dist/types'

import { predeploys } from '../src'
import {
  assertContractVariable,
  deploy, deploySleepTime,
  getContractFromArtifact,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const L1CrossDomainMessengerProxy = await getContractFromArtifact(
    hre,
    'Proxy__OVM_L1CrossDomainMessenger'
  )
  const l1MantleToken = hre.deployConfig.l1MantleToken
  if (l1MantleToken === "0x000000000000000000000000000000000000000000000000") {
    throw new Error(`l1 mantle token is empty`)
  }

  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'L1StandardBridge',
    args: [L1CrossDomainMessengerProxy.address, l1MantleToken],
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
        l1MantleToken
      )
    },
  })
}

deployFn.tags = ['L1StandardBridgeImpl', 'setup', 'l1']

export default deployFn
