// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";

/// @title IOptimismPortal
/// @notice Interface for the OptimismPortal contract
interface IOptimismPortal {
    /// @notice Constructor function with the same parameters as OptimismPortal
    /// @param _l2Oracle Address of the L2OutputOracle contract
    /// @param _guardian Address that can pause deposits and withdrawals
    /// @param _paused   Sets the contract's pausability state
    /// @param _config   Address of the SystemConfig contract
    /// @param _l1MNT    Address of the L1 Mantle Token
    function __constructor__(
        IL2OutputOracle _l2Oracle,
        address _guardian,
        bool _paused,
        ISystemConfig _config,
        address _l1MNT
    )
        external;

    function L2_ORACLE() external view returns (IL2OutputOracle);
    function GUARDIAN() external view returns (address);
    function SYSTEM_CONFIG() external view returns (ISystemConfig);
    function L1_MNT_ADDRESS() external view returns (address);
}
