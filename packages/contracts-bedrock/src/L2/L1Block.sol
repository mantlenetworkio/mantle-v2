// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Libraries
import { Constants } from "src/libraries/Constants.sol";

import { Semver } from "../universal/Semver.sol";

// // Interfaces
// import { ISemver } from "interfaces/universal/ISemver.sol";

/**
 * @custom:proxied
 * @custom:predeploy 0x4200000000000000000000000000000000000015
 * @title L1Block
 * @notice The L1Block predeploy gives users access to information about the last known L1 block.
 *         Values within this contract are updated once per epoch (every L1 block) and can only be
 *         set by the "depositor" account, a special system address. Depositor account transactions
 *         are created by the protocol whenever we move to a new epoch.
 */
contract L1Block is Semver {
    /// @notice Address of the special depositor account.
    function DEPOSITOR_ACCOUNT() public pure returns (address addr_) {
        addr_ = Constants.DEPOSITOR_ACCOUNT;
    }
    /**
     * @notice The latest L1 block number known by the L2 system.
     */

    uint64 public number;

    /**
     * @notice The latest L1 timestamp known by the L2 system.
     */
    uint64 public timestamp;

    /**
     * @notice The latest L1 basefee.
     */
    uint256 public basefee;

    /**
     * @notice The latest L1 blockhash.
     */
    bytes32 public hash;

    /**
     * @notice The number of L2 blocks in the same epoch.
     */
    uint64 public sequenceNumber;

    /**
     * @notice The scalar value applied to the L1 blob base fee portion of the blob-capable L1 cost func.
     */
    uint32 public blobBaseFeeScalar;

    /**
     * @notice The scalar value applied to the L1 base fee portion of the blob-capable L1 cost func.
     */
    uint32 public baseFeeScalar;

    /**
     * @notice The versioned hash to authenticate the batcher by.
     */
    bytes32 public batcherHash;

    /**
     * @notice The overhead value applied to the L1 portion of the transaction
     *         fee.
     */
    uint256 public l1FeeOverhead;

    /**
     * @notice The scalar value applied to the L1 portion of the transaction fee.
     */
    uint256 public l1FeeScalar;

    /**
     * @notice The latest L1 blob base fee.
     */
    uint256 public blobBaseFee;

    /**
     * @notice The constant value applied to the operator fee.
     */
    uint64 public operatorFeeConstant;

    /**
     * @notice The scalar value applied to the operator fee.
     */
    uint32 public operatorFeeScalar;

    /**
     * @custom:semver 1.1.0
     */
    constructor() Semver(1, 1, 0) { }

    /**
     * @notice Updates the L1 block values.
     *
     * @param _number         L1 blocknumber.
     * @param _timestamp      L1 timestamp.
     * @param _basefee        L1 basefee.
     * @param _hash           L1 blockhash.
     * @param _sequenceNumber Number of L2 blocks since epoch start.
     * @param _batcherHash    Versioned hash to authenticate batcher by.
     * @param _l1FeeOverhead  L1 fee overhead.
     * @param _l1FeeScalar    L1 fee scalar.
     */
    function setL1BlockValues(
        uint64 _number,
        uint64 _timestamp,
        uint256 _basefee,
        bytes32 _hash,
        uint64 _sequenceNumber,
        bytes32 _batcherHash,
        uint256 _l1FeeOverhead,
        uint256 _l1FeeScalar
    )
        external
    {
        require(msg.sender == DEPOSITOR_ACCOUNT(), "L1Block: only the depositor account can set L1 block values");

        number = _number;
        timestamp = _timestamp;
        basefee = _basefee;
        hash = _hash;
        sequenceNumber = _sequenceNumber;
        batcherHash = _batcherHash;
        l1FeeOverhead = _l1FeeOverhead;
        l1FeeScalar = _l1FeeScalar;
    }

    /// @notice Updates the L1 block values for Arsia upgraded chain.
    /// Params are packed and passed in as raw msg.data instead of ABI to reduce calldata size.
    /// Params are expected to be in the following order:
    ///   1. _baseFeeScalar        L1 base fee scalar (uint32)             - 4 bytes
    ///   2. _blobBaseFeeScalar    L1 blob base fee scalar (uint32)       - 4 bytes
    ///   3. _sequenceNumber       Number of L2 blocks since epoch (uint64) - 8 bytes
    ///   4. _timestamp            L1 timestamp (uint64)                   - 8 bytes
    ///   5. _number               L1 blocknumber (uint64)                 - 8 bytes
    ///   6. _basefee              L1 base fee (uint256)                   - 32 bytes
    ///   7. _blobBaseFee          L1 blob base fee (uint256)              - 32 bytes
    ///   8. _hash                 L1 blockhash (bytes32)                  - 32 bytes
    ///   9. _batcherHash          Versioned hash (bytes32)                - 32 bytes
    ///   10. _operatorFeeScalar   Operator fee scalar (uint32)            - 4 bytes
    ///   11. _operatorFeeConstant Operator fee constant (uint64)          - 8 bytes
    function setL1BlockValuesArsia() public {
        _setL1BlockValuesArsia();
    }

    /// @notice Internal function to set L1 block values for Arsia upgrade.
    ///         Uses assembly for gas optimization.
    function _setL1BlockValuesArsia() internal {
        address depositor = DEPOSITOR_ACCOUNT();
        assembly {
            // Revert if the caller is not the depositor account.
            if xor(caller(), depositor) {
                mstore(0x00, 0x3cc50b45) // 0x3cc50b45 is the 4-byte selector of "NotDepositor()"
                revert(0x1C, 0x04) // returns the stored 4-byte selector from above
            }
            // sequencenum (uint64), blobBaseFeeScalar (uint32), baseFeeScalar (uint32)
            sstore(sequenceNumber.slot, shr(128, calldataload(4)))
            // number (uint64) and timestamp (uint64)
            sstore(number.slot, shr(128, calldataload(20)))
            sstore(basefee.slot, calldataload(36)) // uint256
            sstore(blobBaseFee.slot, calldataload(68)) // uint256
            sstore(hash.slot, calldataload(100)) // bytes32
            sstore(batcherHash.slot, calldataload(132)) // bytes32

            // operatorFeeScalar (uint32), operatorFeeConstant (uint64)
            sstore(operatorFeeConstant.slot, shr(160, calldataload(164)))
        }
    }
}
