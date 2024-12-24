import {
  Contract,
  Overrides,
  Signer,
  BigNumber,
  CallOverrides,
  PayableOverrides,
} from 'ethers'
import {
  BlockTag,
  TransactionRequest,
  TransactionResponse,
} from '@ethersproject/abstract-provider'

import { NumberLike, AddressLike, TokenBridgeMessage } from './types'
import { CrossChainMessenger } from '../cross-chain-messenger'

/**
 * Represents an adapter for an L1<>L2 token bridge. Each custom bridge currently needs its own
 * adapter because the bridge interface is not standardized. This may change in the future.
 */
export interface IERC721BridgeAdapter {
  /**
   * Provider used to make queries related to cross-chain interactions.
   */
  messenger: CrossChainMessenger

  /**
   * L1 bridge contract.
   */
  l1Bridge: Contract

  /**
   * L2 bridge contract.
   */
  l2Bridge: Contract

  /**
   * Checks whether the given token pair is supported by the bridge.
   *
   * @param l1Token The L1 token address.
   * @param l2Token The L2 token address.
   * @returns Whether the given token pair is supported by the bridge.
   */
  supportsTokenPair(
    l1Token: AddressLike,
    l2Token: AddressLike
  ): Promise<boolean>

  /**
   * Queries the account's approval amount for a given L1 token.
   *
   * @param l1Token The L1 token address.
   * @param l2Token The L2 token address.
   * @param signer Signer to query the approval for.
   * @returns Amount of tokens approved for deposits from the account.
   */
  approval(
    l1Token: AddressLike,
    l2Token: AddressLike,
    signer: Signer
  ): Promise<boolean>

  /**
   * Approves a deposit into the L2 chain.
   *
   * @param l1Token The L1 token address.
   * @param l2Token The L2 token address.
   * @param amount Amount of the token to approve.
   * @param signer Signer used to sign and send the transaction.
   * @param opts Additional options.
   * @param opts.overrides Optional transaction overrides.
   * @returns Transaction response for the approval transaction.
   */
  approve(
    l1Token: AddressLike,
    l2Token: AddressLike,
    signer: Signer,
    opts?: {
      overrides?: Overrides
    }
  ): Promise<TransactionResponse>

  /**
   * Deposits some tokens into the L2 chain.
   *
   * @param l1Token The L1 token address.
   * @param l2Token The L2 token address.
   * @param amount Amount of the token to deposit.
   * @param signer Signer used to sign and send the transaction.
   * @param opts Additional options.
   * @param opts.recipient Optional address to receive the funds on L2. Defaults to sender.
   * @param opts.l2GasLimit Optional gas limit to use for the transaction on L2.
   * @param opts.overrides Optional transaction overrides.
   * @returns Transaction response for the deposit transaction.
   */
  deposit(
    l1Token: AddressLike,
    l2Token: AddressLike,
    tokenId: NumberLike,
    signer: Signer,
    opts?: {
      recipient?: AddressLike
      l2GasLimit?: NumberLike
      overrides?: Overrides
    }
  ): Promise<TransactionResponse>

  getWithdrawalsByAddress(
    address: AddressLike,
    opts?: {
      fromBlock?: BlockTag
      toBlock?: BlockTag
    }
  ): Promise<TokenBridgeMessage[]>

  getDepositsByAddress(
    address: AddressLike,
    opts?: {
      fromBlock?: BlockTag
      toBlock?: BlockTag
    }
  ): Promise<TokenBridgeMessage[]>

  /**
   * Withdraws some tokens back to the L1 chain.
   *
   * @param l1Token The L1 token address.
   * @param l2Token The L2 token address.
   * @param amount Amount of the token to withdraw.
   * @param signer Signer used to sign and send the transaction.
   * @param opts Additional options.
   * @param opts.recipient Optional address to receive the funds on L1. Defaults to sender.
   * @param opts.overrides Optional transaction overrides.
   * @returns Transaction response for the withdraw transaction.
   */
  withdraw(
    l1Token: AddressLike,
    l2Token: AddressLike,
    tokenId: NumberLike,
    signer: Signer,
    opts?: {
      recipient?: AddressLike
      overrides?: Overrides
    }
  ): Promise<TransactionResponse>

  /**
   * Object that holds the functions that generate transactions to be signed by the user.
   * Follows the pattern used by ethers.js.
   */
  populateTransaction: {
    /**
     * Generates a transaction for approving some tokens to deposit into the L2 chain.
     *
     * @param l1Token The L1 token address.
     * @param l2Token The L2 token address.
     * @param opts Additional options.
     * @param opts.overrides Optional transaction overrides.
     * @returns Transaction that can be signed and executed to deposit the tokens.
     */
    approve(
      l1Token: AddressLike,
      l2Token: AddressLike,
      opts?: {
        overrides?: Overrides
      }
    ): Promise<TransactionRequest>

    /**
     * Generates a transaction for depositing some tokens into the L2 chain.
     *
     * @param l1Token The L1 token address.
     * @param l2Token The L2 token address.
     * @param amount Amount of the token to deposit.
     * @param opts Additional options.
     * @param opts.recipient Optional address to receive the funds on L2. Defaults to sender.
     * @param opts.l2GasLimit Optional gas limit to use for the transaction on L2.
     * @param opts.overrides Optional transaction overrides.
     * @returns Transaction that can be signed and executed to deposit the tokens.
     */
    deposit(
      l1Token: AddressLike,
      l2Token: AddressLike,
      tokenId: NumberLike,
      opts?: {
        recipient?: AddressLike
        l2GasLimit?: NumberLike
        overrides?: PayableOverrides
      }
    ): Promise<TransactionRequest>

    /**
     * Generates a transaction for withdrawing some tokens back to the L1 chain.
     *
     * @param l1Token The L1 token address.
     * @param l2Token The L2 token address.
     * @param amount Amount of the token to withdraw.
     * @param opts Additional options.
     * @param opts.recipient Optional address to receive the funds on L1. Defaults to sender.
     * @param opts.overrides Optional transaction overrides.
     * @returns Transaction that can be signed and executed to withdraw the tokens.
     */
    withdraw(
      l1Token: AddressLike,
      l2Token: AddressLike,
      tokenId: NumberLike,
      opts?: {
        recipient?: AddressLike
        overrides?: Overrides
      }
    ): Promise<TransactionRequest>
  }

  /**
   * Object that holds the functions that estimates the gas required for a given transaction.
   * Follows the pattern used by ethers.js.
   */
  estimateGas: {
    /**
     * Estimates gas required to approve some tokens to deposit into the L2 chain.
     *
     * @param l1Token The L1 token address.
     * @param l2Token The L2 token address.
     * @param opts Additional options.
     * @param opts.overrides Optional transaction overrides.
     * @returns Gas estimate for the transaction.
     */
    approve(
      l1Token: AddressLike,
      l2Token: AddressLike,
      opts?: {
        overrides?: CallOverrides
      }
    ): Promise<BigNumber>

    /**
     * Estimates gas required to deposit some tokens into the L2 chain.
     *
     * @param l1Token The L1 token address.
     * @param l2Token The L2 token address.
     * @param amount Amount of the token to deposit.
     * @param opts Additional options.
     * @param opts.recipient Optional address to receive the funds on L2. Defaults to sender.
     * @param opts.l2GasLimit Optional gas limit to use for the transaction on L2.
     * @param opts.overrides Optional transaction overrides.
     * @returns Gas estimate for the transaction.
     */
    deposit(
      l1Token: AddressLike,
      l2Token: AddressLike,
      tokenId: NumberLike,
      opts?: {
        recipient?: AddressLike
        l2GasLimit?: NumberLike
        overrides?: CallOverrides
      }
    ): Promise<BigNumber>

    /**
     * Estimates gas required to withdraw some tokens back to the L1 chain.
     *
     * @param l1Token The L1 token address.
     * @param l2Token The L2 token address.
     * @param amount Amount of the token to withdraw.
     * @param opts Additional options.
     * @param opts.recipient Optional address to receive the funds on L1. Defaults to sender.
     * @param opts.overrides Optional transaction overrides.
     * @returns Gas estimate for the transaction.
     */
    withdraw(
      l1Token: AddressLike,
      l2Token: AddressLike,
      tokenId: NumberLike,
      opts?: {
        recipient?: AddressLike
        overrides?: CallOverrides
      }
    ): Promise<BigNumber>
  }
}
