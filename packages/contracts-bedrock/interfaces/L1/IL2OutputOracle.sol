// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

/// @title IL2OutputOracle
/// @notice Interface for the L2OutputOracle contract
interface IL2OutputOracle {
    /// @notice Constructor function with the same parameters as L2OutputOracle
    /// @param _submissionInterval  Interval in blocks at which checkpoints must be submitted
    /// @param _l2BlockTime         The time per L2 block, in seconds
    /// @param _startingBlockNumber The number of the first L2 block
    /// @param _startingTimestamp   The timestamp of the first L2 block
    /// @param _proposer            The address of the proposer
    /// @param _challenger          The address of the challenger
    /// @param _finalizationPeriodSeconds The period in seconds that must elapse before a withdrawal can be finalized
    function __constructor__(
        uint256 _submissionInterval,
        uint256 _l2BlockTime,
        uint256 _startingBlockNumber,
        uint256 _startingTimestamp,
        address _proposer,
        address _challenger,
        uint256 _finalizationPeriodSeconds
    )
        external;

    function SUBMISSION_INTERVAL() external view returns (uint256);
    function L2_BLOCK_TIME() external view returns (uint256);
    function PROPOSER() external view returns (address);
    function CHALLENGER() external view returns (address);
    function FINALIZATION_PERIOD_SECONDS() external view returns (uint256);
}
