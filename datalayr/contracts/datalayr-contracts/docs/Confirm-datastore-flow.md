# Confirming a Datastore - The Flow

<!--add registering as on operator eventually-->

<A name="Initiate a DataStore"></A>
## Step 1: Initiate a DataStore

The first step to confirming a datastore is to initiate it.  `initDataStore` functions as a way of notifying EigenLayer that the disperser has asserted the blob of data into EigenDA, and is waiting to obtain the quorum signature of the EigenDA nodes on chain.  The fees associated with the datastore are also sent to the contract for escrow.  Another important step in this function involves posting the dataStore header on chain and verifying that the coding ratio specified in the header is adequate, i.e., it is greater than the minimum percentage of operators that must be honest signers.
```solidity
function initDataStore(
        address feePayer,
        address confirmer,
        uint8 duration,
        uint32 referenceBlockNumber,
        uint32 totalOperatorsIndex,
        bytes calldata header
    )
```


## Step 2: Confirming a Datastore

```solidity
function confirmDataStore(
    bytes calldata data, 
    DataStoreSearchData memory searchData
) external onlyWhenNotPaused(PAUSED_CONFIRM_DATASTORE) 
```

The main function of the `confirmDataStore` function is to collect and verify the aggregate BLS signature of the quorum on the datastore.  The signature verification algorithm is as follows with these inputs:

```
<
    * bytes32 msgHash, the taskHash for which disperser is calling checkSignatures
    * uint48 index of the totalStake corresponding to the dataStoreId in the 'totalStakeHistory' array of the BLSRegistryWithBomb
    * uint32 blockNumber, the blockNumber at which the task was initated
    * uint32 taskNumberToConfirm
    * uint32 numberOfNonSigners,
    * {uint256[2] pubkeyG1, uint32 stakeIndex}[numberOfNonSigners] the G1 public key and the index to query of `pubkeyHashToStakeHistory` for each nonsigner,
    * uint32 apkIndex, the index in the `apkUpdates` array at which we want to load the aggregate public key
    * uint256[2] apkG1 (G1 aggregate public key, including nonSigners),
    * uint256[4] apkG2 (G2 aggregate public key, not including nonSigners),
    * uint256[2] sigma, the aggregate signature itself
    * 
>
```
The actual verification of the aggregate BLS signature on the datastore involves computing an elliptic curve pairing.  In order to arrive at this step, there are several things to be computed:

- The first step is to calculate the aggregate nonsigner public key, `aggNonSignerPubkeyG1` by adding all the nonsigner public keys in G1.  
- We then compute the aggregate signer public key, by computing `apkG1 - aggNonSignerPubkeyG1`.
- Now we can proceed to compute the pairing. The standard BLS pairing is as follows:
$$e(\sigma, g_2) = e(H(m), pk_2)$$

where $g_2$ is the generator in G2.  This requires the public key to be in G2.  However computing ecAdd for the aggregate public key in G2 is extremely expensive.  So instead, we compute the public key aggregation in G1 with the following pairing:

$$ e(\sigma + \gamma(pk_1), -g_2) = e(\gamma(g_1) + H(m), pk_2) $$

Doing some quick math, this checks out:
$$e(\sigma + \gamma(pk_1), -g_2) = e([sk](H(m) + \gamma(g_1)), -g_2) =$$
$$e(\gamma(g_1) + H(m), [sk]g_2) =  e(\gamma(g_1) + H(m), pk_2)$$


Looking closer, this pairing is verifying two separate pairings at once:

$$e(\sigma, g_2) = e(H(m), pk_2)$$

$$e(pk_1, g_2) = e(g_1, pk_2)$$

As you can see, the first pairing is the BLS verification pairing while the seconf pairing simply verifies that the public key in $g_1$ is equivalent to the public key in $g_2$:

$$e(pk_1, g_2) = e(g_1, [sk]g_2) = e(g_1, pk_2)$$

Thus we are able to compute the aggregate public key in G1 and verify it against the public key in G2 (which is provided as an input to the function).  In parallel, we also compute the total signing stake for the datastore in question (by subtracting away nonsigner stake from total possible quorum stake). Upon verifying the signatures, we then check whether the quorum stake requirements are met for both quorums:

$$ \frac{signedStakeQuorum }{totalStakeQuorum} >= quorumThreshold$$








