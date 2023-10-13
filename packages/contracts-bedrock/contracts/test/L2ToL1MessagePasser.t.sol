// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { BVMETH_Initializer } from "./CommonTest.t.sol";
import { L2ToL1MessagePasser } from "../L2/L2ToL1MessagePasser.sol";
import { Types } from "../libraries/Types.sol";
import { Hashing } from "../libraries/Hashing.sol";

contract L2ToL1MessagePasserTest is BVMETH_Initializer {

    event MessagePassed(
        uint256 indexed nonce,
        address indexed sender,
        address indexed target,
        uint256 mntValue,
        uint256 ethValue,
        uint256 gasLimit,
        bytes data,
        bytes32 withdrawalHash
    );

    event WithdrawerBalanceBurnt(uint256 indexed amount);

    function setUp() public virtual override {
        super.setUp();
    }

    function testFuzz_initiateWithdrawal_succeeds(
        uint256 _mntValue,
        uint256 _ethValue,
        uint256 _gasLimit,
        bytes memory _data
    ) external {
        uint256 nonce = messagePasser.messageNonce();

        bytes32 withdrawalHash = Hashing.hashWithdrawal(
            Types.WithdrawalTransaction({
                nonce: nonce,
                sender: alice,
                target: alice,
                mntValue: _mntValue,
                ethValue: _ethValue,
                gasLimit: _gasLimit,
                data: _data
            })
        );

        vm.deal(alice, _mntValue);
        deal(address(l2ETH), alice, _ethValue);
        vm.store(address(l2ETH), bytes32(uint256(0x2)), bytes32(_ethValue)); //set total supply

        vm.expectEmit(true, true, true, true);
        emit MessagePassed(nonce, alice, alice, _mntValue, _ethValue, _gasLimit, _data, withdrawalHash);

        vm.prank(alice);
        messagePasser.initiateWithdrawal{ value: _mntValue }(_ethValue, alice, _gasLimit, _data);

        assertEq(messagePasser.sentMessages(withdrawalHash), true);

        bytes32 slot = keccak256(bytes.concat(withdrawalHash, bytes32(0)));

        assertEq(vm.load(address(messagePasser), slot), bytes32(uint256(1)));
    }

    // Test: initiateWithdrawal should emit the correct log when called by a contract
    function test_initiateWithdrawal_fromContract_succeeds() external {
        bytes32 withdrawalHash = Hashing.hashWithdrawal(
            Types.WithdrawalTransaction(
                messagePasser.messageNonce(),
                address(this),
                address(4),
                100,
                0,
                64000,
                hex""
            )
        );

        vm.expectEmit(true, true, true, true);
        emit MessagePassed(
            messagePasser.messageNonce(),
            address(this),
            address(4),
            100,
            0,
            64000,
            hex"",
            withdrawalHash
        );

        vm.deal(address(this), 2**64);
        messagePasser.initiateWithdrawal{ value: 100 }(0, address(4), 64000, hex"");
    }

    // Test: initiateWithdrawal should emit the correct log when called by an EOA
    function test_initiateWithdrawal_fromEOA_succeeds() external {
        uint256 gasLimit = 64000;
        address target = address(4);
        uint256 mntValue = 100;
        uint256 ethValue =0;
        bytes memory data = hex"ff";
        uint256 nonce = messagePasser.messageNonce();

        // EOA emulation
        vm.prank(alice, alice);
        vm.deal(alice, 2**64);
        bytes32 withdrawalHash = Hashing.hashWithdrawal(
            Types.WithdrawalTransaction(nonce, alice, target, mntValue, ethValue, gasLimit, data)
        );

        vm.expectEmit(true, true, true, true);
        emit MessagePassed(nonce, alice, target, mntValue, ethValue, gasLimit, data, withdrawalHash);

        messagePasser.initiateWithdrawal{ value: mntValue }(ethValue, target, gasLimit, data);

        // the sent messages mapping is filled
        assertEq(messagePasser.sentMessages(withdrawalHash), true);
        // the nonce increments
        assertEq(nonce + 1, messagePasser.messageNonce());
    }

    // Test: burn should destroy the ETH held in the contract
    function test_burn_succeeds() external {
        messagePasser.initiateWithdrawal{ value: NON_ZERO_VALUE }(
            ZERO_VALUE,
            NON_ZERO_ADDRESS,
            NON_ZERO_GASLIMIT,
            NON_ZERO_DATA
        );

        assertEq(address(messagePasser).balance, NON_ZERO_VALUE);
        vm.expectEmit(true, false, false, false);
        emit WithdrawerBalanceBurnt(NON_ZERO_VALUE);
        messagePasser.burn();

        // The Withdrawer should have no balance
        assertEq(address(messagePasser).balance, 0);
    }
}
