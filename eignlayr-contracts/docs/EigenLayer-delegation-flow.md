
# Delegation Flow

While delegating to an operator is designed to be a simple process from the staker's perspective, a lot happens "under the hood".

In order to be delegated *to*, an operator must have first called `EigenLayerDelegation.registerAsOperator`. If a staker tries to delegate to someone who has not previously registered as an operator, their transaction will fail.

When an operator registers in EigenLayer, the following flow of calls between contracts occurs:

![Registering as an Operator in EigenLayer](images/EL_operator_registration.png?raw=true "Registering as an Operator in EigenLayer")

1. The would-be operator calls `EigenLayerDelegation.registerAsOperator`, providing either a `DelegationTerms`-type contract or an EOA as input. The EigenLayerDelegation contract stores the `DelegationTerms`-type contract provided by the operator, which may act as an intermediary to help facilitate the relationship between the operator and any stakers who delegate to them.
All of the remaining steps (2-4) proceed as outlined in the delegation process below; the EigenLayerDelegation contract treats things as if the operator has delegated to themselves.

For a staker to delegate to an operator, the staker must either:
1. Call `EigenLayerDelegation.delegateTo` directly
OR
2. Supply an appropriate ECDSA signature, which can then be submitted by the operator (or a third party) as part of a call to `EigenLayerDelegation.delegateToBySignature`

In either case, the end result is the same, and the flow of calls between contracts looks identical:

![Delegating in EigenLayer](images/EL_delegating.png?raw=true "Delegating in EigenLayer")

1. As outlined above, either the staker themselves calls `EigenLayerDelegation.delegateTo`, or the operator (or a third party) calls `EigenLayerDelegation.delegateToBySignature`, in which case the EigenLayerDelegation contract verifies the provided ECDSA signature
2. The EigenLayerDelegation contract calls `Slasher.isFrozen` to verify that the operator being delegated to is not frozen
3. The EigenLayerDelegation contract calls `InvestmentManager.getDeposits` to get the full list of the staker (who is delegating)'s deposits. It then increases the delegated share amounts of operator (who is being delegated to) appropriately
4. The EigenLayerDelegation contract makes a call into the operator's stored `DelegationTerms`-type contract, calling the `onDelegationReceived` function to inform it of the new delegation