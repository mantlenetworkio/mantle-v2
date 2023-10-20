import {subtask, task} from 'hardhat/config'
import 'hardhat-deploy'
import '@nomiclabs/hardhat-ethers'
import {ethers, Wallet} from "ethers";
import {CONTRACT_ADDRESSES,L1ChainID,L2ChainID} from "../src";
import {sleep} from "@eth-optimism/core-utils";

import {
  CrossChainMessenger,
  MessageStatus,
} from '../src'
import {assert} from "chai";

let l1CustomERC20Address: string
let l2CustomERC20Address: string
let l1Wallet: Wallet
let l2Wallet: Wallet
let crossChainMessenger:CrossChainMessenger

async function initializeCrossChainMessenger() {
    const contractAddrs = CONTRACT_ADDRESSES[L2ChainID.OPTIMISM_BEDROCK_LOCAL_DEVNET]
    const l1Cid = L1ChainID.BEDROCK_LOCAL_DEVNET
    const l2Cid = L2ChainID.OPTIMISM_BEDROCK_LOCAL_DEVNET
    const privateKey = 'b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291'
    const l1RpcProvider = new ethers.providers.JsonRpcProvider('http://localhost:8545')
    l1Wallet = new ethers.Wallet(privateKey, l1RpcProvider)
    const l2RpcProvider = new ethers.providers.JsonRpcProvider('http://localhost:9545')
    l2Wallet = new ethers.Wallet(privateKey, l2RpcProvider)
    crossChainMessenger = new CrossChainMessenger({
        l1ChainId: l1Cid,
        l2ChainId: l2Cid,
        l1SignerOrProvider: l1Wallet,
        l2SignerOrProvider: l2Wallet,
        bedrock: true,
        contracts: contractAddrs
    })
}



/**
 * yarn hardhat depositWithdrawEthV2 --damount 5000000000000000 --wamount 1000000000000000
 */
task("depositWithdrawEthV2", "depositWithdrawEthV2")
    .addParam('damount', "deposi eth amount")
    .addParam('wamount', "withdraw eth amount")
    .setAction(async (taskArgs, hre) => {
        await initializeCrossChainMessenger()
        const balance1 = await l1Wallet.getBalance()
        console.log(`l1 eth balance（wei）：${balance1.toString()}`);
        const approvalTx = await crossChainMessenger.approveERC20(
            process.env.Proxy__L1MantleToken,
            process.env.l2MntIsEthAddress,
            hre.ethers.constants.MaxUint256
        )
        console.log(`Deposit eth approvalTx hash (on L1): ${approvalTx.hash}`)

        const response =await crossChainMessenger.depositETH(taskArgs.damount)
        console.log(`Deposit ETH transaction hash (on L1): ${response.hash}`)
        await response.wait()
        console.log("Waiting for status to change to RELAYED")
        await crossChainMessenger.waitForMessageStatus(response.hash, MessageStatus.RELAYED)

        crossChainMessenger.approval(ethers.constants.AddressZero, '0xdEAddEaDdeadDEadDEADDEAddEADDEAddead1111')
        const wresponse = await crossChainMessenger.withdrawETH(taskArgs.wamount)
        console.log(`Withdraw ETH Transaction hash (on L2): ${wresponse.hash}`)
        await wresponse.wait()

        let status = await crossChainMessenger.getMessageStatus(wresponse.hash);
        console.log(`Status 1: ${MessageStatus[status]}`);
        await crossChainMessenger.waitForMessageStatus(wresponse.hash,MessageStatus.READY_TO_PROVE)
        console.log(`Status 2: ${MessageStatus[status]}`);

    });









export async function delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}


