// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "../permissions/Pausable.sol";
import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import "../interfaces/ISlashRecorder.sol";

contract SlashRecorder is Initializable, OwnableUpgradeable, ISlashRecorder {
    address public slasherManager;

    SlashMember[] slashMemberList;

    constructor() {
        _disableInitializers();
    }

    modifier onlySlasherManager() {
        require(msg.sender == slasherManager, "Only the slasher manager can do this action");
        _;
    }

    function initialize(address slasherAddress, address initialOwner) public initializer {
        slasherManager = slasherAddress;
        _transferOwnership(initialOwner);
    }

    function addEvilMember(address _memberAddress, SlashType evilType, string calldata socket) external onlySlasherManager {
        SlashMember memory slashMember = SlashMember(
            _memberAddress,
            evilType,
            socket
        );
        slashMemberList.push(slashMember);
    }

    function getSlashMemberList() external view returns (SlashMember[] memory) {
        return slashMemberList;
    }

    function resetSlashMemberList() external onlySlasherManager {
        delete slashMemberList;
    }

    function setSlasherManager(address slasherAddress) external onlyOwner {
        require(slasherAddress != address(0), "slasherAddress is the zero address");
        slasherManager = slasherAddress;
    }

}
