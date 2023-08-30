// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "./CommonTest.t.sol";
import { LegacyERC20MNT } from "../legacy/LegacyERC20MNT.sol";
import { Predeploys } from "../libraries/Predeploys.sol";
import { BridgeConstants }from "../libraries/BridgeConstants.sol";
contract LegacyERC20MNT_Test is CommonTest {
    LegacyERC20MNT mnt;

    function setUp() public virtual override {
        super.setUp();
        mnt = new LegacyERC20MNT();
    }

    function test_metadata_succeeds() external {
        assertEq(mnt.name(), "Mantle Token");
        assertEq(mnt.symbol(), "MNT");
        assertEq(mnt.decimals(), 18);
    }

    function test_crossDomain_succeeds() external {
        assertEq(mnt.l2Bridge(), Predeploys.L2_STANDARD_BRIDGE);
        assertEq(mnt.l1Token(), BridgeConstants.L1_MNT);
    }

    function test_transfer_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: transfer is disabled");
        mnt.transfer(alice, 100);
    }

    function test_approve_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: approve is disabled");
        mnt.approve(alice, 100);
    }

    function test_transferFrom_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: transferFrom is disabled");
        mnt.transferFrom(bob, alice, 100);
    }

    function test_increaseAllowance_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: increaseAllowance is disabled");
        mnt.increaseAllowance(alice, 100);
    }

    function test_decreaseAllowance_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: decreaseAllowance is disabled");
        mnt.decreaseAllowance(alice, 100);
    }

    function test_mint_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: mint is disabled");
        mnt.mint(alice, 100);
    }

    function test_burn_doesNotExist_reverts() external {
        vm.expectRevert("LegacyERC20MNT: burn is disabled");
        mnt.burn(alice, 100);
    }
}
