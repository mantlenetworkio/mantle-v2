//! Mantle Hoodi chain constants

/// Mantle Hoodi chain ID.
///
/// Mantle Hoodi's mainline fork times come from its genesis and rollup configuration. Only
/// `HoodiBvmUpgrade` has a bundled activation time, see [`MANTLE_HOODI_BVM_UPGRADE_TIMESTAMP`].
pub const MANTLE_HOODI_CHAIN_ID: u64 = 50002;

/// Mantle Hoodi L2 block time in seconds.
///
/// `HoodiBvmUpgrade` writes its code in the first block at or after the activation timestamp,
/// detected as "active at the block timestamp but not one block time earlier". If Hoodi's block
/// time changes before the fork activates, update this value or pick an activation timestamp that
/// equals a block timestamp.
pub const MANTLE_HOODI_BLOCK_TIME: u64 = 2;

/// `HoodiBvmUpgrade` activation timestamp for Mantle Hoodi. `None` until scheduled.
///
/// Bundled into op-reth's built-in `mantle-hoodi` chain spec. Must equal op-geth's
/// `MantleHoodiUpgradeConfig.MantleHoodiBvmUpgradeTime` and op-node's
/// `rollup.BundledMantleHoodiBVMUpgradeTime`.
pub const MANTLE_HOODI_BVM_UPGRADE_TIMESTAMP: Option<u64> = None;
