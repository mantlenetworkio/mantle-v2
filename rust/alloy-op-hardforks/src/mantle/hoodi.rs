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
/// Must equal op-geth's `params.BundledMantleHoodiBvmUpgradeTime` and op-node's
/// `rollup.BundledMantleHoodiBVMUpgradeTime`.
pub const MANTLE_HOODI_BVM_UPGRADE_TIMESTAMP: Option<u64> = None;

/// Returns the `HoodiBvmUpgrade` activation timestamp of a chain.
///
/// On Mantle Hoodi a scheduled [`MANTLE_HOODI_BVM_UPGRADE_TIMESTAMP`] takes precedence over
/// the chain's configuration, as bundled Mantle mainnet and Sepolia fork times do; otherwise the
/// configured timestamp is used.
pub const fn mantle_hoodi_bvm_upgrade_time(chain_id: u64, configured: Option<u64>) -> Option<u64> {
    match MANTLE_HOODI_BVM_UPGRADE_TIMESTAMP {
        Some(timestamp) if chain_id == MANTLE_HOODI_CHAIN_ID => Some(timestamp),
        _ => configured,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn hoodi_bvm_upgrade_time_on_hoodi() {
        let expected = |configured| MANTLE_HOODI_BVM_UPGRADE_TIMESTAMP.or(configured);
        assert_eq!(
            mantle_hoodi_bvm_upgrade_time(MANTLE_HOODI_CHAIN_ID, Some(7)),
            expected(Some(7))
        );
        assert_eq!(mantle_hoodi_bvm_upgrade_time(MANTLE_HOODI_CHAIN_ID, None), expected(None));
    }

    #[test]
    fn hoodi_bvm_upgrade_time_on_other_chains_is_configured() {
        assert_eq!(mantle_hoodi_bvm_upgrade_time(5003, Some(7)), Some(7));
        assert_eq!(mantle_hoodi_bvm_upgrade_time(5000, None), None);
        assert_eq!(mantle_hoodi_bvm_upgrade_time(5003, None), None);
    }
}
