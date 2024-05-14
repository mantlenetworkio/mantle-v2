
# EigenPods: Handling Beacon Chain ETH

## Overview

This document explains *EigenPods*, the mechanism by which EigenLayer facilitates the restaking of native beacon chain ether.

It is important to contrast this with the restaking of liquid staking derivatives (LSDs) on EigenLayer. EigenLayer will integrate with liquid staking protocols "above the hood", meaning that withdrawal credentials will be pointed to EigenLayer at the smart contract layer rather than the consensus layer. This is because liquid staking protocols need their contracts to be in possession of the withdrawal credentials in order to not have platform risk on EigenLayer. As always, this means that value of liquid staking derivatives carries a discount due to additional smart contract risk.

The architechtural design of the EigenPods system is inspired by various liquid staking protocols, particularly Rocket Pool 🚀.

## The EigenPodManager

The EigenPodManager facilitates the higher level functionality of EigenPods and their interactions with the rest of the EigenLayer smart contracts (the InvestmentManager and the InvestmentManager's owner). Stakers can call the EigenPodManager to create pods (whose addresses are determintically calculated via the Create2 OZ library) and stake on the Beacon Chain through them. The EigenPodManager also handles the cumulative paid penalties (explained later) of all EigenPods and allows the InvestmentManger's owner to redistribute them. 

## The EigenPod

The EigenPod is the contract that a staker must set their Etherum validators' withdrawal credentials to. EigenPods can be created by stakers through a call to the EigenPodManger. EigenPods are deployed using the beacon proxy pattern to have flexible global upgradability for future changes to the Ethereum specification. Stakers can stake for an Etherum validator when they create their EigenPod, through further calls to their EigenPod, and through parallel deposits to the Beacon Chain deposit contract.

### Beacon State Root Oracle

EigenPods extensively use a Beacon State Root Oracle that will bring beacon state roots into Ethereum for every [`SLOTS_PER_HISTORICAL_ROOT`](https://github.com/ethereum/consensus-specs/blob/dev/specs/phase0/beacon-chain.md#time-parameters) slots (currently 8192 slots or ~27 hours) so that all intermediate state roots can be proven against the ones posted on execution layer.

### Proof of Correctly Pointed Withdrawal Credentials

After staking an Etherum validator with its withdrawal credentials pointed to their EigenPod, a staker must prove that the new validator exists and has its withdrawal credentials pointed to the EigenPod against a beacon state root. The EigenPod will verify the proof (along with checking for replays and other conditions) and, if the ETH validator's balance is proven to be greater than `REQUIRED_BALANCE_WEI`, then the EigenPod will call the EigenPodManager to forward a call to the InvestmentManager, crediting the staker with `REQUIRED_BALANCE_WEI` shares of the virtual beacon chain ETH strategy. `REQUIRED_BALANCE_WEI` will be set to an amount of ether that a validator could get slashed down to only due to malice or negligence. The current back-of-the-envelope calculations show that 31.4 ETH is the minimum balance an offline validator can have after a week of inactivity, so it sets a good indicator for `REQUIRED_BALANCE_WEI`. For reference, there are only about 50 validators below this balance on the Ethereum beacon chain as of 12/7/2022.

### Fraud Proofs for Overcommitted Balances

If a Ethereum validator restaked on an EigenPod has a balance that falls below `REQUIRED_BALANCE_WEI`, then they are overcommitted to EigenLayer, meaning they have less stake on the beacon chain than they have restaked Eigenlayer. Any watcher can prove to EigenPods that the EigenPod has a validator that is in such a state. If proof verification and other checks succeed, then `REQUIRED_BALANCE_WEI` will be immediately decremented from the EigenPod owner's (the staker's) shares in the InvestmentManager. This overcommitment imposes a negative externality on middlewares that the staker is securing, since the middlewares a suffer sudden downgrade in security as part of the process. To punish stakers for this offense, `OVERCOMMITMENT_PENALTY_AMOUNT_GWEI` will be incremented to the penalties that the pod owner owes to EigenLayer (described later).

### Proofs of Full Withdrawals

Whenever an staker withdraws one of their validators from the beacon chain to provide liquidity, they have a few options. Stakers could keep the ETH in the EigenPod and continue staking on EigenLayer, in which case their ETH, when withdrawn to the EigenPod, will not earn any additional Ethereum staking yield, it will only earn their EigenLayer staking yield. Stakers could also queue withdrawals on EigenLayer for the virtual beacon chain ETH strategy which will be fullfilled once their staking obligations have ended and their EigenPod has enough balance to complete the withdrawal.

In this second case, in order to withdraw their balance from the EigenPod, stakers must provide a valid proof of their full withdrawal (differentiated from partial withdrawals through a simple comparison of the amount to a threshold value named `MIN_FULL_WITHDRAWAL_AMOUNT_GWEI`) against a beacon state root. Once the proof is successfully verified, if the amount withdrawn is less than `REQUIRED_BALANCE_GWEI` the validators balance is deducted from EigenLayer and the penalties are added, similar to [fraud proofs for overcommitted balances](https://github.com/Layr-Labs/eignlayr-contracts/edit/update-eigenpod-withdrawals/docs/EigenPods.md#fraud-proofs-for-overcommitted-balances). 

If the withdrawn amount is greater than `REQUIRED_BALANCE_GWEI`, then the excess is marked as instantly withdrawable after the call returns. This is fine, since the amount is not restaked on EigenLayer.

Finally, before the call returns, the EigenPod attempts to pay off any penalties it owes using the newly withdrawn amount.

Note that there exists a tail-end exploit here. If a validator was already slashed, they could get themselved slashed below `MIN_FULL_WITHDRAWAL_AMOUNT_GWEI` (very hard) and then withdraw this balance as a partial withdrawal because its amount is so little, even though it is actually a full withdrawal. This is not huge concern at the moment since it is extremely improbable and not economically advantageous. It is a known issue though.

### Partial Withdrawal Claims

One of the biggest changes that will come along with the Capella hardfork is the addition of partial withdrawals. Partial withdrawals are withdrawals on behalf of validators that aren't exiting from the beacon chain, but rather have a balance (due to yield) of greater than 32 ETH. Since these withdrawals happen every block and are often of small size, they do not make economic sense to prove individually. However, since yield is not immediately restaked on EigenLayer, the protocol allows stakers to withdraw this yield in an sensible way through an optimistic claims process. 

The balance of an EigenPod at anytime is `FULL_WITHDRAWALS + PARTIAL_WITHDRAWALS` where `FULL_WITHDRAWALS` is the balance due to full withdrawals that have not been withdrawn from the EigenPod yet and `PARTIAL_WITHDRAWALS`  is the balance due to partial withdrawals that have not been withdrawn from the EigenPod yet and includes any miscellaneous balance increases due to `selfdestruct`, etc.). The key idea in the partial withdrawal claim process is that, since full withdrawals happen much less frequently than partial withdrawals, a staker only needs to claim that they have proven all of the full withdrawals up to a certain block and record the EigenPod's balance at that block in order to calculate the ETH value of their partial withdrawals. 

In more detail, the EigenPod owner:
1. Proves all their full withdrawals up to the `currentBlockNumber`
2. Calculates the block number of the next full withdrawal occuring at or after `currentBlockNumber`. Call this `expireBlockNumber`.
3. Pings the contract with a transaction claiming that they have proven all full withdrawals until `expireBlockNumber`. The contract will note this in storage along with `partialWithdrawals = address(this).balance - FULL_WITHDRAWALS`.
4. If a watcher proves a full withdrawal for a validator restaked on the EigenPod that occured before `currentBlockNumber` that has not been proven before and this proof occurs within `PARTIAL_WITHDRAWAL_FRAUD_PROOF_PERIOD_BLOCKS` (an EigenPod contract variable) of the partial withdrawal claim, the claim is marked as failed. This means that the claim cannot be withdrawn, but the mechanism for rewarding watchers has not been fully worked out yet. It will prbably be the case that watchers will eventually end up being paid through penalty withdrawals.
5. If no such proof is provided within `PARTIAL_WITHDRAWAL_FRAUD_PROOF_PERIOD_BLOCKS`, the staker is allow withdraw `partialWithdrawals` (first attempting to pay off penalties with the ether to withdraw) and make new partial withdrawal claims

So, overall, stakers must wait for this delay period after partial withdrawals occur leading to a slight delay in partial withdrawal redemption when staking on EigenLayer vs not.

### Paying Penalties

During proofs of full withdrawals, EigenPods first attempt to pay off their penalties as much as they can from the withdrawn restaked (not instantly withdrawable balance) balance. IF that is not enough, they then attempt to pay off their penalties as much as they can from the withdrawn excess balance.

During partial withdrawal claims, before partial withdrawals are redeemed, they are used to pay off penalties to the extent they can.

Note that the withdrawn restaked balance can be overestimated through this penalty payment scheme, since they are paid off in such an aggressive manner. Penalties can come out of a staker's instantly withdrawable balance and their partial withdrawals. This means that ETH that was withdrawn to the execution layer through full withdrawals can end up staying in the contract because there aren't enough shares on EigenLayer to withdraw. In the case of such a state, stakers can roll over the overestimated restaked balance of their EigenPods into their instantly withdrawable funds. This essentially retroactively pays the penalties from their restaked balance and brings back their instantly withdrawable rewards that were previously used to pay penalties. Of course, being as pessimistic as possible, the EigenPod will attempt to pay off penalties as best as it can before allowing its owner to instantly withdraw their new instantly withdrawable funds. For more details track mentions of `rollableBalanceGwei` and the `rollOverRollableBalance` function.