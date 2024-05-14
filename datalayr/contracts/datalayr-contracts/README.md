<a name="introduction"/></a>
# EigenDA
EigenDA (formerly 'DataLayr') is a Data Availability network built on top of EigenLayer.
For more on EigenLayer, see the [EigenLayer documentation](link-to-be-added).

Click the links in the Table of Contents below to access more specific documentation. We recommend starting with the [EigenDA Contracts Technical Specification](docs/EigenDA-contracts-tech-spec.md) to get a better overview before diving into any of the other docs.

## Table of Contents  
* [Introduction](#introduction)
* [Installation and Running Tests / Analyzers](#installation)
* [EigenDA Contracts Technical Specification](docs/EigenDA-contracts-tech-spec.md)
* [An Introduction to Proofs of Custody](docs/Proofs-of-Custody.md)
* [EigenDA Registration Flow](docs/DataLayr-registration-flow.md)
* [Low Degree Challenge Deep Dive](docs/LowDegreenessChallenge-overview.md)

<a name="installation"/></a>
## Getting Started
The EigenDA smart contracts utilize the EigenLayer smart contracts as a library/submodule.  In order to resolve relevant imports during compilation, follow these steps:
1. Launch a github codespace (preferably a 4-core or 8-core) in the datalayr repo
2. Within the repo, navigate to the eigenlayer-contracts repo: `cd contracts/eigenlayer-contracts`, a submodule pointing to the eigenlayer repo.
3. Checkout the branch of eigenlayer-contracts that you would like your datalayr branch to read from: `git checkout [BRANCH_NAME]`.
4. Then navigate to the datalayr-contracts repo: `cd contracts/datalayr-contracts`
5. Then run `./watch.sh start` to copy over the contents of the eigenlayer-contracts repo into `/lib`.  Now we can resolve eigenlayer contract imports in EigenDA contracts.
6. Run `foundryup` followed by `forge build` to ensure the contracts are compiling correctly.

## Pushing Changes
1.  If you intend to make changes to the contract, checkout the `contract-dev` branch, and make a branch from there.
2. If your commit involves changes to EigenLayer contracts, ensure that you merge those changes to `master`.
3.  Then go ahead and create a PR in the datalayr contracts, merging your branch into `contract-dev`
4.  Now you're all set!



### Run Tests

`forge test -vv`

### Run Static Analysis

`solhint 'src/contracts/**/*.sol'`

`slither .`

### Generate Inheritance and Control-Flow Graphs

first [install surya](https://github.com/ConsenSys/surya/)

then run

`surya inheritance ./src/contracts/**/*.sol | dot -Tpng > InheritanceGraph.png`

and/or

`surya graph ./src/contracts/middleware/*.sol | dot -Tpng > MiddlewareControlFlowGraph.png`

and/or

`surya mdreport surya_report.md ./src/contracts/**/*.sol`
