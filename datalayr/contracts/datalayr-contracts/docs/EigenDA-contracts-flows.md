
# EigenDA Contracts Flows

This outlines some of the details of contract-to-contract communications within EigenDA.

## Storing Data

### Initiating a DataStore

![Initiating a DataStore in EigenDA](images/DL_init_datastore.png?raw=true "Initiating a DataStore in EigenDA")

1. The disperser calls `DataLayrServiceManager.initDataStore`, providing relevant information about the DataStore.
2. The DataLayrServiceManager calls `DataLayrPaymentManager.payFee`, deducting from the disperser's pre-deposited payment balance. This assures the nodes in the network that payment has been provisioned for the DataStore (i.e. that they will be paid for storing the data).
3. Finally, after the DataLayrServiceManager updates its storage, it returns an index to the disperser, which specifies the location in the array `dataStoreHashesForDurationAtTimestamp[duration][block.timestamp]` at which the DataStore's hash was stored -- the index is useful in future operations that retrieve information related to the DataStore.

### Confirming a DataStore

![Confirming a DataStore in EigenDA](images/DL_confirm_datastore.png?raw=true "Confirming a DataStore in EigenDA")

1. The disperser calls `DataLayrServiceManager.confirmDataStore`, providing aggregate signatures as well as lookup data for the DataStore.
2. As the DataLayrServiceManager processes the signatures, it looks up the total stake of all operators through a call to `BLSRegistryWithBomb.getTotalStakeFromIndex`.
3. For each non-signer, the DataLayrServiceManager must also look up the individual operator's stake, retrieved through a call to `BLSRegistryWithBomb.getStakeFromPubkeyHashAndIndex`.
4. The DataLayrServiceManager must also verify the integrity of the provided aggregate public key for all EigenDA operators; to do so it consults the value returned by `BLSRegistryWithBomb.getStakeFromPubkeyHashAndIndex.getCorrectApkHash`. Finally it can verify the integrity of the provided signature and verify that it meets all requirements, and then proceed to processing the confirmation.

## Payments

### "Committing" a Payment

![Committing a Payment in EigenDA](images/DL_committing_payment.png?raw=true "Committing a Payment in EigenDA")

1. The operator calls `DataLayrPaymentManager.commitPayment`, informing the contract of the `amount` the operator *claims* they are owed, for their services up to the specified `toTaskNumber`.
2. The DataLayrPaymentManager calls `BLSRegistryWithBomb.isActiveOperator` to confirm that the initial call is indeed an active operator in EigenDA.
3. The DataLayrPaymentManager calls `DataLayrServiceManager.taskNumber` to get the current taskNumber and compare to the specified `toTaskNumber`, ensuring that no improper claim to future payments is being made.
4. The DataLayrPaymentManager calls the CollateralToken contract, transferring tokens from the operator to the DataLayrPaymentManager, to be used as the operator's collateral in the case that a payment challenge is raised.
5. *In the event that this is the operator's first payment claim*, the DataLayrPaymentManager calls `BLSRegistryWithBomb.getFromTaskNumber` in order to determine the proper taskNumber for the payment claim to begin from.

### Redeeming a Payment

![Redeeming a Payment in EigenDA](images/DL_redeeming_payment.png?raw=true "Redeeming a Payment in EigenDA")

1. After the payment fraudproof interval has elapsed, the operator calls `DataLayrPaymentManager.commitPayment` to claim their payment.
2. The DataLayrPaymentManager calls the CollateralToken contract, returning the operator's challenge collateral to them.
3. The DataLayrPaymentManager calls `EigenLayrDelegation.delegationTerms` to fetch to operator's stored `DelegationTerms`-type contract, which mediates the operator's relationship with stakers who delegate to them.
4. The DataLayrPaymentManager calls the PaymentToken contract, actually transferring the payment amount to the operator's stored `DelegationTerms`-type contract.
5. The DataLayrPaymentManager calls the `payForService` function of the operator's stored `DelegationTerms`-type contract, informing it of the payment.

### Creating a Payment Challenge

![Creating a Payment Challenge in EigenDA](images/DL_creating_payment_challenge.png?raw=true "Creating a Payment Challenge in EigenDA")

1. The challenger calls `DataLayrPaymentManager.initPaymentChallenge`
2. The DataLayrPaymentManager calls the CollateralToken contract, transferring tokens from the challenger to the DataLayrPaymentManager, to be used as the challenger's collateral throughout the challenge process.

### The Payment Challenge Process

![The Payment Challenge Process in EigenDA](images/DL_payment_challenge_process.png?raw=true "The Payment Challenge Process in EigenDA")

1 - n. The operator and challenger take turns calling `DataLayrPaymentChallenge.performBisectionStep`, until the challenge is focused down to the payment on a single DataStore.

(n+1). The operator and challenger (whoever's turn it is) calls `DataLayrPaymentChallenge.respondToPaymentChallengeFinal`, triggering a calculation of the *true* payment amount for the single DataStore in question.

(n+2). As part of verifying the integrity of the provided input data (related to the DataStore in question), the DataLayrPaymentChallenge contract calls `DataLayrServiceManager.getDataStoreHashesForDurationAtTimestamp`.

(n+3). The DataLayrPaymentChallenge contract calls `BLSRegistryWithBomb.getOperatorPubkeyHash` in order to obtain the correct public key hash for the operator.

(n+4) & (n+5). The DataLayrPaymentChallenge contract calls `BLSRegistryWithBomb.quorumBips` to understand how the payment should be split between the two quorums. The DataLayrPaymentChallenge now has enough information to calculate the true payment amount. 

(n+6). In an edge case, the DataLayrPaymentChallenge contract calls `BLSRegistryWithBomb.getFromBlockNumberForOperator` to further ensure the integrity of the provided data (note: this is to verify the case when the task was based off of stakes before the operator registered).

(n+7). The DataLayrPaymentManager calls the CollateralToken contract, sending the collateral of the 'loser' of the payment challenger process to the 'winner' of the process.