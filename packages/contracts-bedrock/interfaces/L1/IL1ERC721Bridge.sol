// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

/// @title IL1ERC721Bridge
/// @notice Interface for the L1ERC721Bridge contract
interface IL1ERC721Bridge {
    /// @notice Constructor function with the same parameters as L1ERC721Bridge
    /// @param _messenger   Address of the CrossDomainMessenger on this network
    /// @param _otherBridge Address of the ERC721 bridge on the other network
    function __constructor__(address _messenger, address _otherBridge) external;

    function MESSENGER() external view returns (address);
    function OTHER_BRIDGE() external view returns (address);
}
