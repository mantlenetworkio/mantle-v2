// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

/// @title IOptimismMintableERC20Factory
/// @notice Interface for the OptimismMintableERC20Factory contract.
interface IOptimismMintableERC20Factory {
    /// @notice Constructor function with the same parameters as SystemConfig
    /// @param _bridge Address of the StandardBridge on this chain.
    function __constructor__(address _bridge) external;

    function BRIDGE() external view returns (address);
}
