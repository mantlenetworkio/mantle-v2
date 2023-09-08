import { DeployFunction } from 'hardhat-deploy/dist/types'
import '@eth-optimism/hardhat-deploy-config'
import 'hardhat-deploy'

import { deploy } from '../src/deploy-utils'
import {sleep} from "@eth-optimism/core-utils";

const deployFn: DeployFunction = async (hre) => {
  await sleep(5000)
  await deploy({
    hre,
    name: 'SystemDictator',
    args: [],
  })
}

deployFn.tags = ['SystemDictatorImpl', 'setup', 'l1']

export default deployFn
