# Proofs of Custody

## Introduction

This explainer is meant to describe our implementation of the bomb, a concept introduced by Dankrad in this [article](https://dankradfeist.de/ethereum/2021/09/30/proofs-of-custody.html), as a way of verifying proof of custody.  In the context of DataLayr, the idea of a bomb is to allow a challenger to verify whether or not a DataLayr Node is storing datastores in the correct manner.  If the DLN signs a given block with a bomb, a challenger will be able to prove that the DLN was behaving maliciously.

## Implementation

### Ephemeral Keys
The proof of custody game works as follows: A given blob contains a bomb that is "detonated" if that blob is signed on by a DLN, resulting in slashing. Thus, the DLN must avoid signing that blob, forcing them to download and store the blob correctly to detect the bomb's presence. 

Whether or not a blob contains a bomb is determined by a unique value we call an ephemeral key.  This ephemeral key is an arbitrary 32-byte value, unique to a given DLN, allowing them to detect the presence of a bomb.  This happens as follows:

1) Upon registering, the DLN generates a random ephemeral key (EK) and posts a commitment to it on chain.  After a fixed period of time, the DLN reveals the EK and posts a commitment to a new EK.  
2) During this disclosure period, a challenger can check for the presence of a bomb in that DLN's datastores.  If a bomb is found and the DLN signed the block, the DLN is slashed!
3) There are several additonal slashing conditions.  The first is when a DLN fails to reveal the EK they committed to within a certain time frame, they are slashed.  They are also slashed if their ephemeral key is revealed by a third party on chain before the disclosure period starts.  


### Bomb Function Implementation
The bomb function  is implemented as follows:

Inputs:
* *k* is the number of active chunks stored by the DLN
* *h* is the blockhash
* *T* is the threshold to evaluate the bomb function against

Evaluation:
$hash(chunk(h \text{ mod } k), EK, h) < T$


The function $chunk(d)$ returns the $d^{th}$ datastore.   If the expression evaluates to $true$, the bomb detonates and if $false$, the bomb is diffused.  The threshold probabilistically ensures that the bomb will only detonate a certain percentage of the time.  

There is a possible attack here, due to the partially deterministic nature of the bomb function.  A malicious DLN will be able to obtain the datastore from an external source and compute the bomb without ever having downloaded the datastore.  To remediate this, we force the DLN to compute the bomb function $b$ times, where: 
$$b * \text{datastoreRetreivalCost} >> \text{datastoreStorageCost}$$

This way, the cost of being malicious outweighs the cost of acting honest, making the system cryptoeconomically secure.

### Challenger Fraud Proof

In order to fraud proof the DLN, a challenger needs a list of all the blobs stored by the DLN for the bomb window.  To get such information onchain is expensive and so we present the following solution:

* Datastores in DataLayr can only be stores for discrete periods of time, between 1 and 14 days.  Let us consider the case for Datastores of length $n$, for $1\leq n \leq 14$.  We want the earliest chunk ID and the latest chunk ID within $n$ days of the challenge period start $t$.  
*  Off chain, we iterate through timestamps starting at time $t-7$ days until we find the first chunk, $id_x$.  Similarly, we find the last chunk prior to or at time $t$, $id_y$.  On chain, we need to verify that these are the first and last chunks within the bomb window.  All the challenger has to do is prove that $id_{x-1}$ ends before time $t-7$ days and that $id_{y+1}$ starts after time $t$.  We repeat this process for all values of $n$.
* Thus we have $id_x$ and $id_y$ for each value of $n$, giving us the ability to calcualte the $(h \text{ mod } k)^{th}$ chunk on chain.  Thus the challenger can now compute the bomb function and detect a malicious DLN.






