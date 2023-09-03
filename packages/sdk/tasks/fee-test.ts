import { task, types } from 'hardhat/config'
import '@nomiclabs/hardhat-ethers'
import 'hardhat-deploy'
import {providers, ethers, BigNumber} from 'ethers'
import {sleep} from "@eth-optimism/core-utils";


task('fee-test', 'Test bedrock fee')
  .addParam(
    'l2ProviderUrl',
    'L2 provider URL.',
    'http://localhost:9545',
    types.string
  )
  .setAction(async (args, hre) => {
    const signers = await hre.ethers.getSigners()
    if (signers.length === 0) {
      throw new Error('No configured signers')
    }

    const signer = signers[0]
    const address = await signer.getAddress()
    console.log(`Using signer ${address}`)

    const balance = await signer.getBalance()
    if (balance.eq(0)) {
      throw new Error('Signer has no balance')
    }

    const l2Provider = new providers.StaticJsonRpcProvider(args.l2ProviderUrl)

    const l2Signer = new hre.ethers.Wallet(
      hre.network.config.accounts[0],
      l2Provider
    )

    const privateKey1 = ethers.utils.randomBytes(32);
    const wallet1 = new ethers.Wallet(privateKey1,l2Provider);
    const address1 = wallet1.address;
    const privateKey2 = ethers.utils.randomBytes(32);
    const wallet2 = new ethers.Wallet(privateKey2,l2Provider);
    const address2 = wallet2.address;
    const privateKey3 = ethers.utils.randomBytes(32);
    const wallet3 = new ethers.Wallet(privateKey3,l2Provider);
    const address3 = wallet3.address;

     // await testBaseFee(l2Signer,address3)

    const tx1 = {
      to: address1,
      value: ethers.utils.parseEther('10000'),
    };
    await l2Signer.sendTransaction(tx1);
    const tx2 = {
      to: address2,
      value: ethers.utils.parseEther('10000'),
    };
    await l2Signer.sendTransaction(tx2);
    console.log('Wait for transaction to be confirmed')
    await sleep(30000)
    console.log('Initialize account done')

    const priorityFees = ['50000', '30000', '40000'];
    const senders = [l2Signer,wallet1,wallet2]
    const promises = priorityFees.map(async (priorityFee, index) => {
      const sender = senders[index];
      return sendTransactionWithGasPrice(priorityFee, sender, address3);
    });
    await Promise.all(promises);
    getBalance(wallet3);
    await sleep(30000)
  })

const sendTransactionWithGasPrice = async (priorityFee: string,sender,to) => {

  const tx = {
    to,
    value: ethers.utils.parseEther('1'),
    maxPriorityFeePerGas:priorityFee,
    maxFeePerGas:10000000000
  };
  const res = await sender.sendTransaction(tx);
  console.log(`transaction response , ${JSON.stringify(res)}`)
  console.log('priorityFee:', priorityFee);
};

const getBalance = async (wallet3) => {

  const interval = setInterval(async () => {
    const b = await wallet3.getBalance();
    const result: BigNumber = b.div(BigNumber.from(10).pow(18));
    console.log(`Address: ${wallet3.address}, Balance: ${result} ETH` );
  }, 100);

  // await new Promise((resolve) => setTimeout(resolve, 5000));
  // clearInterval(interval);

}

const testBaseFee = async (l2Signer,address3) =>{
  for (let i = 0; i < 100; i++) {
    const tx = {
      to: address3,
      value: ethers.utils.parseEther('1'),
      gasLimit: 210000,
      gasPrice: ethers.utils.parseUnits('200', 'gwei'),
    };
    const res = await l2Signer.sendTransaction(tx);
    console.log(`transaction response , ${JSON.stringify(res)}`)
  }
}
