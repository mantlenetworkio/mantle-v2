import { task, types } from 'hardhat/config'
import { HardhatRuntimeEnvironment } from 'hardhat/types'
import '@nomiclabs/hardhat-ethers'
import 'hardhat-deploy'
import {Event, Contract, Wallet, providers, ethers} from 'ethers'
import {
  predeploys,
  getContractDefinition,
} from '@eth-optimism/contracts-bedrock'

const deployerc721 = async (
  hre: HardhatRuntimeEnvironment,
): Promise<Contract> => {
  const signers = await hre.ethers.getSigners()
  const signer = signers[0]

  const Artifact__ERC721 = await getContractDefinition('ERC721')
  const Factory__ERC721 = new hre.ethers.ContractFactory(
    Artifact__ERC721.abi,
    Artifact__ERC721.bytecode,
    signer
  )
  console.log('Sending deployment transaction')
  const ERC721 = await Factory__ERC721.deploy();

  const receipt = await ERC721.deployTransaction.wait()
  console.log(`ERC721 deployed: ${receipt.transactionHash}`)

  return ERC721
}

const createOptimismMintableERC721 = async (
  hre: HardhatRuntimeEnvironment,
  L1ERC721: Contract,
  l2Signer: Wallet
): Promise<Contract> => {
  const Artifact__OptimismMintableERC721Token = await getContractDefinition(
    'OptimismMintableERC721'
  )

  const Artifact__OptimismMintableERC721TokenFactory =
    await getContractDefinition('OptimismMintableERC721Factory')

  const OptimismMintableERC721TokenFactory = new Contract(
    predeploys.OptimismMintableERC721Factory,
    Artifact__OptimismMintableERC721TokenFactory.abi,
    l2Signer
  )

  const name = await L1ERC721.name()
  const symbol = await L1ERC721.symbol()
  console.log(`name : ${name}`)
  console.log(`symbol : ${symbol}`)

  const tx =
    await OptimismMintableERC721TokenFactory.createOptimismMintableERC721(
      L1ERC721.address,
      `L2 ${name}`,
      `L2-${symbol}`
    )

  const receipt = await tx.wait()
  const event = receipt.events.find(
    (e: Event) => e.event === 'OptimismMintableERC721Created'
  )

  if (!event) {
    throw new Error('Unable to find OptimismMintableERC721Created event')
  }

  const l2erc721Address = event.args.localToken
  console.log(`Deployed to ${l2erc721Address}`)

  return new Contract(
    l2erc721Address,
    Artifact__OptimismMintableERC721Token.abi,
    l2Signer
  )
}


task('transfer-erc721', 'transfer erc721 onto L2.')
  .addParam(
    'l2ProviderUrl',
    'L2 provider URL.',
    'http://localhost:9545',
    types.string
  )
  .addParam(
    'opNodeProviderUrl',
    'op-node provider URL',
    'http://localhost:7545',
    types.string
  )
  .addOptionalParam(
    'l1ContractsJsonPath',
    'Path to a JSON with L1 contract addresses in it',
    '',
    types.string
  )
  .setAction(async (args, hre) => {
    const signers = await hre.ethers.getSigners()
    if (signers.length === 0) {
      throw new Error('No configured signers')
    }
    // Use the first configured signer for simplicity
    const signer = signers[0]
    const address = await signer.getAddress()
    console.log(`Using signer ${address}`)

    // Ensure that the signer has a balance before trying to
    // do anything
    const balance = await signer.getBalance()
    if (balance.eq(0)) {
      throw new Error('Signer has no balance')
    }

    const l2Provider = new providers.StaticJsonRpcProvider(args.l2ProviderUrl)

    const l2Signer = new hre.ethers.Wallet(
      hre.network.config.accounts[0],
      l2Provider
    )

    console.log('Deploying erc721 to L1')
    const erc721 = await deployerc721(hre)
    console.log(`Deployed to ${erc721.address}`)

    const gasLimit = 60000;
    const txa = await erc721.approve(l2Signer.address,1,{ gasLimit })
    console.log(`ERC721 approve: ${txa.hash}`)

    console.log('Creating L2 erc721')
    const OptimismMintableERC721 = await createOptimismMintableERC721(
      hre,
      erc721,
      l2Signer
    )

    const privateKey = ethers.utils.randomBytes(32);
    const wallet = new ethers.Wallet(privateKey);
    const address2 = wallet.address;

    const tx = await OptimismMintableERC721.transferFrom(l2Signer.address,address2,1,{ gasLimit });

    console.log(`send erc721 to : ${address2}`)
    console.log('Transaction hash:', tx.hash);

  })
