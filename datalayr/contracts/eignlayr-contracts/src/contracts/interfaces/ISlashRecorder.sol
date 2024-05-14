// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

/**
 * @title Interface for the primary 'SlashRecorder' contract for Mantle.
 * @author mantle, Inc.
 * @notice See the `SlashRecorder` contract itself for implementation details.
 */
interface ISlashRecorder {
    enum SlashType {
        NetworkStability,
        DataValidity
    }

    struct SlashMember {
        address memberAddress;
        SlashType  evilType;
        string socket;
    }

    function addEvilMember(address _memberAddress, SlashType evilType, string calldata socket) external;
    function getSlashMemberList() external view returns (SlashMember[] memory);
    function resetSlashMemberList() external;
    function setSlasherManager(address slasherAddress) external;
}
