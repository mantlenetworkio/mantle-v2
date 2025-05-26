// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
/// @title ISystemConfig
/// @notice Interface for the SystemConfig contract

interface ISystemConfig {
    /// @notice Constructor function with the same parameters as SystemConfig
    /// @param _owner             Initial owner of the contract
    /// @param _overhead          Initial overhead value
    /// @param _scalar            Initial scalar value
    /// @param _batcherHash       Initial batcher hash
    /// @param _gasLimit          Initial gas limit
    /// @param _baseFee           Initial base fee
    /// @param _unsafeBlockSigner Initial unsafe block signer address
    /// @param _config            Initial resource config
    function __constructor__(
        address _owner,
        uint256 _overhead,
        uint256 _scalar,
        bytes32 _batcherHash,
        uint64 _gasLimit,
        uint256 _baseFee,
        address _unsafeBlockSigner,
        IResourceMetering.ResourceConfig memory _config
    )
        external;

    function owner() external view returns (address);
    function overhead() external view returns (uint256);
    function scalar() external view returns (uint256);
    function batcherHash() external view returns (bytes32);
    function gasLimit() external view returns (uint64);
    function baseFee() external view returns (uint256);
    function unsafeBlockSigner() external view returns (address);
    function resourceConfig() external view returns (IResourceMetering.ResourceConfig memory);
}
