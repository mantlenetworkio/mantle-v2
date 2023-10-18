import { DeployFunction } from 'hardhat-deploy/dist/types'
import '@eth-optimism/hardhat-deploy-config'

import {
  assertContractVariable,
  deploy, deploySleepTime,
  getContractFromArtifact,
} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  const { deployer } = await hre.getNamedAccounts()
  const isLiveDeployer =
    deployer.toLowerCase() === hre.deployConfig.controller.toLowerCase()

  const L2OutputOracleProxy = await getContractFromArtifact(
    hre,
    'L2OutputOracleProxy'
  )

  const Artifact__SystemConfigProxy = await hre.deployments.get(
    'SystemConfigProxy'
  )

  const l1MantleToken = hre.deployConfig.l1MantleToken
  if (l1MantleToken.toString() === "0x0000000000000000000000000000000000000000") {
    throw new Error(`missing l1 mantle token address in deploy config`)
  }

  const portalGuardian = hre.deployConfig.portalGuardian
  const portalGuardianCode = await hre.ethers.provider.getCode(portalGuardian)
  if (portalGuardianCode === '0x') {
    console.log(
      `WARNING: setting OptimismPortal.GUARDIAN to ${portalGuardian} and it has no code`
    )
    if (!isLiveDeployer) {
      throw new Error(
        'Do not deploy to production networks without the GUARDIAN being a contract'
      )
    }
  }

  // Deploy the OptimismPortal implementation as paused to
  // ensure that users do not interact with it and instead
  // interact with the proxied contract.
  // The `portalGuardian` is set at the GUARDIAN.
  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'OptimismPortal',
    args: [
      L2OutputOracleProxy.address,
      portalGuardian,
      true, // paused
      Artifact__SystemConfigProxy.address,
      l1MantleToken
    ],
    postDeployAction: async (contract) => {
      await assertContractVariable(
        contract,
        'L2_ORACLE',
        L2OutputOracleProxy.address
      )
      await assertContractVariable(
        contract,
        'GUARDIAN',
        hre.deployConfig.portalGuardian
      )
      await assertContractVariable(
        contract,
        'SYSTEM_CONFIG',
        Artifact__SystemConfigProxy.address
      )
      await assertContractVariable(
        contract,
        'L1_MNT_ADDRESS',
        l1MantleToken
      )
    },
  })
}

deployFn.tags = ['OptimismPortalImpl', 'setup', 'l1']

export default deployFn
