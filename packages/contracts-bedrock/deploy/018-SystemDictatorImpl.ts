import { DeployFunction } from 'hardhat-deploy/dist/types'
import '@eth-optimism/hardhat-deploy-config'
import 'hardhat-deploy'

import {deploy, deploySleepTime} from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  await sleep(deploySleepTime)
  await deploy({
    hre,
    name: 'SystemDictator',
    args: [],
  })
}

deployFn.tags = ['SystemDictatorImpl', 'setup', 'l1']

export default deployFn
