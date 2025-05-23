// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

/// @title IL1StandardBridge
/// @notice Interface for the L1StandardBridge contract
interface IL1StandardBridge {
    /// @notice Constructor function with the same parameters as L1StandardBridge
    /// @param _messenger Address of the L1CrossDomainMessenger
    /// @param _l1mnt    Address of the Mantle Token on L1
    function __constructor__(address payable _messenger, address _l1mnt) external;
}