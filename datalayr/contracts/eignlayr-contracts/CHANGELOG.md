# Changelog
## [0.1.0] - 2023-06-08

### Features
- Add staker staking and unstaking permission mechanism, Implementing a mechanism in the Delegation that directly triggers a stake update on the BLSRegistry, add SlashRecorder contract for on chain record operators evil([#1](https://github.com/mantlenetworkio/eignlayr-contracts/pull/1))

### Improvements
- Token switch from BIT to MNT([#4](https://github.com/mantlenetworkio/eignlayr-contracts/pull/4))

### Bug Fixes
- Modified the unit test script for the contract([#2](https://github.com/mantlenetworkio/eignlayr-contracts/pull/2))

### Deprecated
- Remove PaymentManager contracts for handling payments from the AVS to operators, remove Slashing mechanism for slashing operators evil([#5](https://github.com/mantlenetworkio/eignlayr-contracts/pull/5))