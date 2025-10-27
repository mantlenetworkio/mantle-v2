// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
/// @title IL1Block
/// @notice Interface for the L1Block contract that provides information about the last known L1 block.

interface IL1Block {
    /// @notice Address of the special depositor account.
    function DEPOSITOR_ACCOUNT() external view returns (address);
    /// @notice The latest L1 block number known by the L2 system.
    function number() external view returns (uint64);
    /// @notice The latest L1 timestamp known by the L2 system.
    function timestamp() external view returns (uint64);
    /// @notice The latest L1 basefee.
    function basefee() external view returns (uint256);
    /// @notice The latest L1 blockhash.
    function hash() external view returns (bytes32);
    /// @notice The number of L2 blocks in the same epoch.
    function sequenceNumber() external view returns (uint64);
    /// @notice The versioned hash to authenticate the batcher by.
    function batcherHash() external view returns (bytes32);
    /// @notice The overhead value applied to the L1 portion of the transaction fee.
    function l1FeeOverhead() external view returns (uint256);
    /// @notice The scalar value applied to the L1 portion of the transaction fee.
    function l1FeeScalar() external view returns (uint256);
    /// @notice The scalar value applied to the L1 base fee portion of the blob-capable L1 cost func.
    function baseFeeScalar() external view returns (uint32);
    /// @notice The scalar value applied to the L1 blob base fee portion of the blob-capable L1 cost func.
    function blobBaseFeeScalar() external view returns (uint32);
    /// @notice The latest L1 blob base fee.
    function blobBaseFee() external view returns (uint256);
    /// @notice The constant value applied to the operator fee.
    function operatorFeeConstant() external view returns (uint64);
    /// @notice The scalar value applied to the operator fee.
    function operatorFeeScalar() external view returns (uint32);
    /// @notice Returns the semver contract version.
    function version() external view returns (string memory);
    /// @notice Updates the L1 block values.
    /// @param _number         L1 blocknumber.
    /// @param _timestamp      L1 timestamp.
    /// @param _basefee        L1 basefee.
    /// @param _hash           L1 blockhash.
    /// @param _sequenceNumber Number of L2 blocks since epoch start.
    /// @param _batcherHash    Versioned hash to authenticate batcher by.
    /// @param _l1FeeOverhead  L1 fee overhead.
    /// @param _l1FeeScalar    L1 fee scalar.
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
        external;
    /// @notice Updates the L1 block values for Arsia upgraded chain.
    /// @dev Params are packed and passed in as raw msg.data instead of ABI to reduce calldata size.
    ///      Params are expected to be in the following order:
    ///      1. _baseFeeScalar        L1 base fee scalar (uint32)             - 4 bytes
    ///      2. _blobBaseFeeScalar    L1 blob base fee scalar (uint32)       - 4 bytes
    ///      3. _sequenceNumber       Number of L2 blocks since epoch (uint64) - 8 bytes
    ///      4. _timestamp            L1 timestamp (uint64)                   - 8 bytes
    ///      5. _number               L1 blocknumber (uint64)                 - 8 bytes
    ///      6. _basefee              L1 base fee (uint256)                   - 32 bytes
    ///      7. _blobBaseFee          L1 blob base fee (uint256)              - 32 bytes
    ///      8. _hash                 L1 blockhash (bytes32)                  - 32 bytes
    ///      9. _batcherHash          Versioned hash (bytes32)                - 32 bytes
    ///      10. _operatorFeeScalar   Operator fee scalar (uint32)            - 4 bytes
    ///      11. _operatorFeeConstant Operator fee constant (uint64)          - 8 bytes
    function setL1BlockValuesArsia() external;
}
