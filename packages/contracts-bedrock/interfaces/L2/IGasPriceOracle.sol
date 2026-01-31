// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title IGasPriceOracle
/// @notice Interface for the GasPriceOracle contract that maintains the variables responsible for
///         computing the L1 portion of the total fee charged on L2.
interface IGasPriceOracle {
    /// @notice Emitted when the token ratio is updated.
    /// @param previousTokenRatio The previous token ratio.
    /// @param newTokenRatio The new token ratio.
    event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio);

    /// @notice Emitted when ownership is transferred.
    /// @param previousOwner The previous owner address.
    /// @param newOwner The new owner address.
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    /// @notice Emitted when the operator is updated.
    /// @param previousOperator The previous operator address.
    /// @param newOperator The new operator address.
    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);

    /// @notice Number of decimals used in the scalar.
    function DECIMALS() external view returns (uint256);

    /// @notice The token ratio used for fee calculations.
    function tokenRatio() external view returns (uint256);

    /// @notice The owner address of the contract.
    function owner() external view returns (address);

    /// @notice The operator address of the contract.
    function operator() external view returns (address);

    /// @notice Indicates whether the network uses Arsia gas calculation.
    function isArsia() external view returns (bool);

    /// @notice Returns the semver contract version.
    function version() external view returns (string memory);

    /// @notice Allows the owner to modify the operator.
    /// @param _operator New operator address.
    function setOperator(address _operator) external;

    /// @notice Transfers ownership of the contract to a new account.
    /// @param _owner New owner address.
    function transferOwnership(address _owner) external;

    /// @notice Allows the operator to modify the token ratio.
    /// @param _tokenRatio New token ratio.
    function setTokenRatio(uint256 _tokenRatio) external;

    /// @notice Set chain to be Arsia chain (callable by depositor account).
    function setArsia() external;

    /// @notice Computes the L1 portion of the fee based on the size of the rlp encoded input
    ///         transaction, the current L1 base fee, and the various dynamic parameters.
    /// @param _data Unsigned fully RLP-encoded transaction to get the L1 fee for.
    /// @return L1 fee that should be paid for the tx.
    function getL1Fee(bytes memory _data) external view returns (uint256);

    /// @notice Retrieves the current gas price (base fee).
    /// @return Current L2 gas price (base fee).
    function gasPrice() external view returns (uint256);

    /// @notice Retrieves the current base fee.
    /// @return Current L2 base fee.
    function baseFee() external view returns (uint256);

    /// @notice Retrieves the current fee overhead.
    /// @return Current fee overhead.
    function overhead() external view returns (uint256);

    /// @notice Retrieves the current fee scalar.
    /// @return Current fee scalar.
    function scalar() external view returns (uint256);

    /// @notice Retrieves the latest known L1 base fee.
    /// @return Latest known L1 base fee.
    function l1BaseFee() external view returns (uint256);

    /// @notice Retrieves the current blob base fee.
    /// @return Current blob base fee.
    function blobBaseFee() external view returns (uint256);

    /// @notice Retrieves the current base fee scalar.
    /// @return Current base fee scalar.
    function baseFeeScalar() external view returns (uint32);

    /// @notice Retrieves the current blob base fee scalar.
    /// @return Current blob base fee scalar.
    function blobBaseFeeScalar() external view returns (uint32);

    /// @notice Retrieves the operator fee scalar.
    /// @return Operator fee scalar.
    function operatorFeeScalar() external view returns (uint32);

    /// @notice Retrieves the operator fee constant.
    /// @return Operator fee constant.
    function operatorFeeConstant() external view returns (uint64);

    /// @notice Computes the amount of L1 gas used for a transaction. Adds the overhead which
    ///         represents the per-transaction gas overhead of posting the transaction and state
    ///         roots to L1. Adds 68 bytes of padding to account for the fact that the input does
    ///         not have a signature.
    /// @param _data Unsigned fully RLP-encoded transaction to get the L1 gas for.
    /// @return Amount of L1 gas used to publish the transaction.
    function getL1GasUsed(bytes memory _data) external view returns (uint256);

    /// @notice Computes the operator fee for a given gas amount.
    /// @param _gasUsed Amount of gas used.
    /// @return Operator fee that should be paid.
    function getOperatorFee(uint256 _gasUsed) external view returns (uint256);

    /// @notice Retrieves the number of decimals used in the scalar.
    /// @custom:legacy
    /// @return Number of decimals used in the scalar.
    function decimals() external pure returns (uint256);
}
