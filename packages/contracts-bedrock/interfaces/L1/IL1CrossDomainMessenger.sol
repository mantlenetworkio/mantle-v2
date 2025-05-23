// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";

/// @title IL1CrossDomainMessenger
/// @notice Interface for the L1CrossDomainMessenger contract
interface IL1CrossDomainMessenger {
    /// @notice Constructor function with the same parameters as L1CrossDomainMessenger
    /// @param _portal Address of the OptimismPortal contract on this network
    /// @param l1mnt   Address of the Mantle Token on L1
    function __constructor__(IOptimismPortal _portal, address l1mnt) external;
}