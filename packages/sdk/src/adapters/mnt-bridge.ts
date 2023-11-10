/* eslint-disable @typescript-eslint/no-unused-vars */
import {ethers, Overrides, BigNumber, Contract} from 'ethers'
import { TransactionRequest, BlockTag } from '@ethersproject/abstract-provider'
import { hexStringEquals } from '@ethan-bedrock/core-utils'
import {
  NumberLike,
  AddressLike,
  TokenBridgeMessage,
  MessageDirection,
} from '../interfaces'
import { toAddress, omit } from '../utils'
import { StandardBridgeAdapter } from './standard-bridge'
import { getContractInterface } from '@ethan-bedrock/contracts-bedrock'

/**
 * Bridge adapter for the MNT bridge.
 */
export class MNTBridgeAdapter extends StandardBridgeAdapter {
  public async approval(
    l1Token: AddressLike,
    l2Token: AddressLike,
    signer: ethers.Signer
  ): Promise<BigNumber> {
    if (!(await this.supportsTokenPair(l1Token, l2Token))) {
      throw new Error(`token pair not supported by bridge`)
    }

    const token = new Contract(
      toAddress(l1Token),
      getContractInterface('OptimismMintableERC20'), // Any ERC20 will do
      this.messenger.l1Provider
    )

    return token.allowance(await signer.getAddress(), this.l1Bridge.address)
  }

  public async getDepositsByAddress(
    address: AddressLike,
    opts?: {
      fromBlock?: BlockTag
      toBlock?: BlockTag
    }
  ): Promise<TokenBridgeMessage[]> {
    const events = await this.l1Bridge.queryFilter(
      this.l1Bridge.filters.MNTDepositInitiated(address),
      opts?.fromBlock,
      opts?.toBlock
    )
    const l1_mnt = this.l1Bridge.L1_MNT_ADDRESS()
    return events
      .map((event) => {
        return {
          direction: MessageDirection.L1_TO_L2,
          from: event.args.from,
          to: event.args.to,
          l1Token: l1_mnt,
          l2Token: ethers.constants.AddressZero,
          amount: event.args.amount,
          data: event.args.extraData,
          logIndex: event.logIndex,
          blockNumber: event.blockNumber,
          transactionHash: event.transactionHash,
        }
      })
      .sort((a, b) => {
        // Sort descending by block number
        return b.blockNumber - a.blockNumber
      })
  }

  public async getWithdrawalsByAddress(
    address: AddressLike,
    opts?: {
      fromBlock?: BlockTag
      toBlock?: BlockTag
    }
  ): Promise<TokenBridgeMessage[]> {
    const events = await this.l2Bridge.queryFilter(
      this.l2Bridge.filters.WithdrawalInitiated(undefined, undefined, address),
      opts?.fromBlock,
      opts?.toBlock
    )

    return events
      .filter((event) => {
        // Only find MNT withdrawals.
        return (
          hexStringEquals(event.args.l2Token, ethers.constants.AddressZero)
        )
      })
      .map((event) => {
        return {
          direction: MessageDirection.L2_TO_L1,
          from: event.args.from,
          to: event.args.to,
          l1Token: event.args.l1Token,
          l2Token: event.args.l2Token,
          amount: event.args.amount,
          data: event.args.extraData,
          logIndex: event.logIndex,
          blockNumber: event.blockNumber,
          transactionHash: event.transactionHash,
        }
      })
      .sort((a, b) => {
        // Sort descending by block number
        return b.blockNumber - a.blockNumber
      })
  }

  public async supportsTokenPair(
    l1Token: AddressLike,
    l2Token: AddressLike
  ): Promise<boolean> {
    // Only support MNT deposits and withdrawals.
    return (
      hexStringEquals(toAddress(l2Token), ethers.constants.AddressZero)
    )
  }

  populateTransaction = {
    approve: async (
      l1Token: AddressLike,
      l2Token: AddressLike,
      amount: NumberLike,
      opts?: {
        overrides?: Overrides
      }
    ): Promise<TransactionRequest> => {
      if (!(await this.supportsTokenPair(l1Token, l2Token))) {
        throw new Error(`token pair not supported by bridge`)
      }

      const token = new Contract(
        toAddress(l1Token),
        getContractInterface('OptimismMintableERC20'), // Any ERC20 will do
        this.messenger.l1Provider
      )

      return token.populateTransaction.approve(
        this.l1Bridge.address,
        amount,
        opts?.overrides || {}
      )
    },

    deposit: async (
      l1Token: AddressLike,
      l2Token: AddressLike,
      amount: NumberLike,
      opts?: {
        recipient?: AddressLike
        l2GasLimit?: NumberLike
        overrides?: Overrides
      }
    ): Promise<TransactionRequest> => {
      if (!(await this.supportsTokenPair(l1Token, l2Token))) {
        throw new Error(`token pair not supported by bridge`)
      }

      if (opts?.recipient === undefined) {
        return this.l1Bridge.populateTransaction.depositMNT(
          amount,
          opts?.l2GasLimit || 200_000, // Default to 200k gas limit.
          '0x', // No data.
        )
      } else {
        return this.l1Bridge.populateTransaction.depositMNTTo(
          toAddress(opts.recipient),
          amount,
          opts?.l2GasLimit || 200_000, // Default to 200k gas limit.
          '0x', // No data.
        )
      }
    },

    withdraw: async (
      l1Token: AddressLike,
      l2Token: AddressLike,
      amount: NumberLike,
      opts?: {
        recipient?: AddressLike
        overrides?: Overrides
      }
    ): Promise<TransactionRequest> => {
      if (!(await this.supportsTokenPair(l1Token, l2Token))) {
        throw new Error(`token pair not supported by bridge`)
      }

      if (opts?.recipient === undefined) {
        return this.l2Bridge.populateTransaction.withdraw(
          toAddress(l2Token),
          amount,
          0, // L1 gas not required.
          '0x', // No data.
          {
            ...omit(opts?.overrides || {}, 'value'),
            value: this.messenger.bedrock ? amount : 0,
          }
        )
      } else {
        return this.l2Bridge.populateTransaction.withdrawTo(
          toAddress(l2Token),
          toAddress(opts.recipient),
          amount,
          0, // L1 gas not required.
          '0x', // No data.
          {
            ...omit(opts?.overrides || {}, 'value'),
            value: this.messenger.bedrock ? amount : 0,
          }
        )
      }
    },
  }
}
