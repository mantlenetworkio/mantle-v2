---
title: LowDegreenessChallenge Overview
tags: high-level-docs
---

# Low Degreeness: Fraud Proofs for Low Degree Headers

## Background - Low degreeness checks

Datastores in EigenDA are commited using the KZG commitment scheme.  The prompt behind KZG hidden reveals is, given a KZG commitment to a polynomial $p(x)$, a prover, $\mathcal{P}$, to convince a verifier, $\mathcal{V}$ he knows $p(\alpha)$ for some public value of $\alpha$ without revealing $p(\alpha)$ or $p(x)$. Read more about it [here](https://hackmd.io/9N1L3VS3RtCJKW--tq8voA?both).

Powers of tau is a setup allowing for commitments to a polynomial, up to a certain degree.  Assume that the max degree of the powers of tau is $n$. This means that one cannot commit to a polynomial of degree greater than $n$.  THe goal of a low degreeness challenge is to prove that the polynomial being committed to is of a certain degree.

## Initiating the Low Degreeness Challenge
A challenger can call the `challengeLowDegreenessProof` function by providing the disputed header.  The function then ensures that the header matches the stored hash for that datastore as well verifying the signatory record of that datastore. Once all the inputs provided by the challenger are verified against the onchain record, the actualy header's low-degreeness proof is verified.
## The Pairing
Verifying the "low-degreeness" of a KZG commitment is acheived using a pairing.  The prover, $\mathcal{P}$ provides $([p(x)]_1, [x^{n-m}p(x)]_1)$, and $\mathcal{V}$ verifies that 

$$e([p(x)]_1, [x^{n-m}]_2) = e([x^{n-m}p(x)]_1, [1]_2)$$

This check verifies that $[x^{n-m}p(x)]_1$ is, in fact, the KZG commitment of $x^{n-m}p(x)$ - this is referred to as the `lowDegreenessProof`. If $p(x)$ has degree $> m$, then $x^{n-m}p(x)$ has degree $>n$, which means that it can't be committed to using the powers of tau. Since the degree check shows that $\mathcal{P}$ has committed to $x^{n-m}p(x)$, so $\mathcal{V}$ knows $p(x)$ has degree $\le m$. It is important to note that $\mathcal{J}$ learns nothing about $p(x)$ but the degree bound in this process. 

[p(x)]_1 is parsed from the header, referred to as `dskzgMetadata.c` and [x^{n-m}]_2 is the `potElement` and provided by the challenger.  We verify that the exponent of the `potElement` is verified against the header, by ensuring verifying its inclusion in the powers of Tau merkle tree (Note: the Low Degreeness Challenge assumes integrity of all data in the header (eg. numSys and degree) except the kzg committment).

The low-degreeness verification fails if 1) the inputs to the pairing are invalid or 2) the pairing itself fails.  In either scenario, the challenger has proven fraud and freezing is initiated

## Freezing the Operator

Once the low-degreeness fraud proof is successful, all the signing operators on that datastore are frozen.  The challenger submits `nonSignerExclusionProofs` which is an array containing the `signerAddress` and the `operatorHistoryIndex` of every signing operator on that datastore.  The function verifies that the provided `nonSignerExclusionProofs` are valid and freezes each one of those operators.  


