// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Bridge_Initializer } from "./CommonTest.t.sol";
import { StandardBridge } from "../universal/StandardBridge.sol";
import { OptimismPortal } from "../L1/OptimismPortal.sol";
import { L2StandardBridge } from "../L2/L2StandardBridge.sol";
import { CrossDomainMessenger } from "../universal/CrossDomainMessenger.sol";
import { Predeploys } from "../libraries/Predeploys.sol";
import { AddressAliasHelper } from "../vendor/AddressAliasHelper.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { stdStorage, StdStorage } from "forge-std/Test.sol";
import { L2CrossDomainMessenger } from "../L2/L2CrossDomainMessenger.sol";
import { L1CrossDomainMessenger } from "../L1/L1CrossDomainMessenger.sol";
import { L1MantleToken } from "../local/TestMantleToken.sol";


contract L1StandardBridge_Getter_Test is Bridge_Initializer {
    function test_getters_succeeds() external {
        assert(L1Bridge.l2TokenBridge() == address(L2Bridge));
        assert(L1Bridge.OTHER_BRIDGE() == L2Bridge);
        assert(L1Bridge.messenger() == L1Messenger);
        assert(L1Bridge.MESSENGER() == L1Messenger);
        assertEq(L1Bridge.version(), "1.1.0");
    }
}

contract L1StandardBridge_Initialize_Test is Bridge_Initializer {
    function test_initialize_succeeds() external {
        assertEq(address(L1Bridge.messenger()), address(L1Messenger));

        assertEq(address(L1Bridge.OTHER_BRIDGE()), Predeploys.L2_STANDARD_BRIDGE);

        assertEq(address(L2Bridge), Predeploys.L2_STANDARD_BRIDGE);
    }
}

contract L1StandardBridge_Initialize_TestFail is Bridge_Initializer {}

contract L1StandardBridge_Receive_Test is Bridge_Initializer {
    // receive
    // - can accept ETH
    function test_receive_succeeds() external {
        assertEq(address(op).balance, 0);

        // The legacy event must be emitted for backwards compatibility
        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHDepositInitiated(alice, alice, 100, hex"");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHBridgeInitiated(alice, alice, 100, hex"");

        vm.expectCall(
            address(L1Messenger),
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                0,
                address(L2Bridge),
                abi.encodeWithSelector(
                    L2StandardBridge.finalizeBridgeETH.selector,
                    alice,
                    alice,
                    100,
                    hex""
                ),
                200_000
            )
        );

        vm.prank(alice, alice);
        (bool success, ) = address(L1Bridge).call{ value: 100 }(hex"");
        assertEq(success, true);
        assertEq(address(op).balance, 100);
    }
}

contract L1StandardBridge_Receive_TestFail {}

contract PreBridgeETH is Bridge_Initializer {
    function _preBridgeETH(bool isLegacy) internal {
        assertEq(address(op).balance, 0);
        uint256 nonce = L1Messenger.messageNonce();
        uint256 version = 0; // Internal constant in the OptimismPortal: DEPOSIT_VERSION
        address l1MessengerAliased = AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger));

        bytes memory message = abi.encodeWithSelector(
            L2StandardBridge.finalizeBridgeETH.selector,
            alice,
            alice,
            500,
            hex"dead"
        );

        if (isLegacy) {
            vm.expectCall(
                address(L1Bridge),
                500,
                abi.encodeWithSelector(L1Bridge.depositETH.selector, 50000, hex"dead")
            );
        } else {
            vm.expectCall(
                address(L1Bridge),
                500,
                abi.encodeWithSelector(L1Bridge.bridgeETH.selector, 50000, hex"dead")
            );
        }
        vm.expectCall(
            address(L1Messenger),
            500,
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                0,
                address(L2Bridge),
                message,
                50000
            )
        );

        bytes memory innerMessage = abi.encodeWithSelector(
            L2CrossDomainMessenger.relayMessage.selector,
            nonce,
            address(L1Bridge),
            address(L2Bridge),
            0,
            500,
            50000,
            message
        );

        uint64 baseGas = L1Messenger.baseGas(message, 50000);
        vm.expectCall(
            address(op),
            500,
            abi.encodeWithSelector(
                OptimismPortal.depositTransaction.selector,
                0,
                address(L2Messenger),
                0,
                baseGas,
                false,
                innerMessage
            )
        );

        bytes memory opaqueData = abi.encodePacked(
            uint256(0),
            uint256(0),
            uint256(500),
            baseGas,
            false,
            innerMessage
        );

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHDepositInitiated(alice, alice, 500, hex"dead");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHBridgeInitiated(alice, alice, 500, hex"dead");

        // OptimismPortal emits a TransactionDeposited event on `depositTransaction` call
        vm.expectEmit(true, true, true, true, address(op));
        emit TransactionDeposited(l1MessengerAliased, address(L2Messenger), version, opaqueData);

        // SentMessage event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessage(address(L2Bridge), address(L1Bridge), message, nonce, 50000);

        // SentMessageExtension1 event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessageExtension1(address(L1Bridge), 0, 500);

        vm.prank(alice, alice);
    }
}

contract L1StandardBridge_DepositETH_Test is PreBridgeETH {
    // depositETH
    // - emits ETHDepositInitiated
    // - emits ETHBridgeInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - ETH ends up in the optimismPortal
    function test_depositETH_succeeds() external {
        _preBridgeETH({ isLegacy: true });
        L1Bridge.depositETH{ value: 500 }(50000, hex"dead");
        assertEq(address(op).balance, 500);
    }
}

contract L1StandardBridge_BridgeETH_Test is PreBridgeETH {
    // BridgeETH
    // - emits ETHDepositInitiated
    // - emits ETHBridgeInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - ETH ends up in the optimismPortal
    function test_bridgeETH_succeeds() external {
        _preBridgeETH({ isLegacy: false });
        L1Bridge.bridgeETH{ value: 500 }(50000, hex"dead");
        assertEq(address(op).balance, 500);
    }
}

contract L1StandardBridge_DepositETH_TestFail is Bridge_Initializer {
    function test_depositETH_notEoa_reverts() external {
        // turn alice into a contract
        vm.etch(alice, address(L1Token).code);

        vm.expectRevert("StandardBridge: function can only be called from an EOA");
        vm.prank(alice);
        L1Bridge.depositETH{ value: 1 }(300, hex"");
    }
}

contract PreBridgeETHTo is Bridge_Initializer {
    function _preBridgeETHTo(bool isLegacy) internal {
        assertEq(address(op).balance, 0);
        uint256 nonce = L1Messenger.messageNonce();
        uint256 version = 0; // Internal constant in the OptimismPortal: DEPOSIT_VERSION
        address l1MessengerAliased = AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger));

        if (isLegacy) {
            vm.expectCall(
                address(L1Bridge),
                600,
                abi.encodeWithSelector(L1Bridge.depositETHTo.selector, bob, 60000, hex"dead")
            );
        } else {
            vm.expectCall(
                address(L1Bridge),
                600,
                abi.encodeWithSelector(L1Bridge.bridgeETHTo.selector, bob, 60000, hex"dead")
            );
        }

        bytes memory message = abi.encodeWithSelector(
            L2StandardBridge.finalizeBridgeETH.selector,
            alice,
            bob,
            600,
            hex"dead"
        );

        // the L1 bridge should call
        // L1CrossDomainMessenger.sendMessage
        vm.expectCall(
            address(L1Messenger),
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                0,
                address(L2Bridge),
                message,
                60000
            )
        );

        bytes memory innerMessage = abi.encodeWithSelector(
            L2CrossDomainMessenger.relayMessage.selector,
            nonce,
            address(L1Bridge),
            address(L2Bridge),
            0,
            600,
            60000,
            message
        );

        uint64 baseGas = L1Messenger.baseGas(message, 60000);
        vm.expectCall(
            address(op),
            600,
            abi.encodeWithSelector(
                OptimismPortal.depositTransaction.selector,
                0,
                address(L2Messenger),
                0,
                baseGas,
                false,
                innerMessage
            )
        );

        bytes memory opaqueData = abi.encodePacked(
            uint256(0),
            uint256(0),
            uint256(600),
            baseGas,
            false,
            innerMessage
        );

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHDepositInitiated(alice, bob, 600, hex"dead");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHBridgeInitiated(alice, bob, 600, hex"dead");

        // OptimismPortal emits a TransactionDeposited event on `depositTransaction` call
        vm.expectEmit(true, true, true, true, address(op));
        emit TransactionDeposited(l1MessengerAliased, address(L2Messenger), version, opaqueData);

        // SentMessage event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessage(address(L2Bridge), address(L1Bridge), message, nonce, 60000);

        // SentMessageExtension1 event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessageExtension1(address(L1Bridge), 0, 600);

        // deposit eth to bob
        vm.prank(alice, alice);
    }
}

contract L1StandardBridge_DepositETHTo_Test is PreBridgeETHTo {
    // depositETHTo
    // - emits ETHDepositInitiated
    // - calls optimismPortal.depositTransaction
    // - EOA or contract can call
    // - ETH ends up in the optimismPortal
    function test_depositETHTo_succeeds() external {
        _preBridgeETHTo({ isLegacy: true });
        L1Bridge.depositETHTo{ value: 600 }(bob, 60000, hex"dead");
        assertEq(address(op).balance, 600);
    }
}

contract L1StandardBridge_BridgeETHTo_Test is PreBridgeETHTo {
    // BridgeETHTo
    // - emits ETHDepositInitiated
    // - emits ETHBridgeInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - ETH ends up in the optimismPortal
    function test_bridgeETHTo_succeeds() external {
        _preBridgeETHTo({ isLegacy: false });
        L1Bridge.bridgeETHTo{ value: 600 }(bob, 60000, hex"dead");
        assertEq(address(op).balance, 600);
    }
}

contract L1StandardBridge_DepositETHTo_TestFail is Bridge_Initializer {}

contract L1StandardBridge_DepositERC20_Test is Bridge_Initializer {
    using stdStorage for StdStorage;

    // depositERC20
    // - updates bridge.deposits
    // - emits ERC20DepositInitiated
    // - calls optimismPortal.depositTransaction
    // - only callable by EOA
    function test_depositERC20_succeeds() external {
        uint256 nonce = L1Messenger.messageNonce();
        uint256 version = 0; // Internal constant in the OptimismPortal: DEPOSIT_VERSION
        address l1MessengerAliased = AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger));

        // Deal Alice's ERC20 State
        deal(address(L1Token), alice, 100000, true);
        vm.prank(alice);
        L1Token.approve(address(L1Bridge), type(uint256).max);

        // The L1Bridge should transfer alice's tokens to itself
        vm.expectCall(
            address(L1Token),
            abi.encodeWithSelector(ERC20.transferFrom.selector, alice, address(L1Bridge), 100)
        );

        bytes memory message = abi.encodeWithSelector(
            L2StandardBridge.finalizeBridgeERC20.selector,
            address(L2Token),
            address(L1Token),
            alice,
            alice,
            100,
            hex""
        );

        // the L1 bridge should call L1CrossDomainMessenger.sendMessage
        vm.expectCall(
            address(L1Messenger),
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                0,
                address(L2Bridge),
                message,
                10000
            )
        );

        bytes memory innerMessage = abi.encodeWithSelector(
            L2CrossDomainMessenger.relayMessage.selector,
            nonce,
            address(L1Bridge),
            address(L2Bridge),
            0,
            0,
            10000,
            message
        );

        uint64 baseGas = L1Messenger.baseGas(message, 10000);
        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                OptimismPortal.depositTransaction.selector,
                0,
                address(L2Messenger),
                0,
                baseGas,
                false,
                innerMessage
            )
        );

        bytes memory opaqueData = abi.encodePacked(
            uint256(0),
            uint256(0),
            uint256(0),
            baseGas,
            false,
            innerMessage
        );

        // Should emit both the bedrock and legacy events
        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ERC20DepositInitiated(address(L1Token), address(L2Token), alice, alice, 100, hex"");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ERC20BridgeInitiated(address(L1Token), address(L2Token), alice, alice, 100, hex"");

        // OptimismPortal emits a TransactionDeposited event on `depositTransaction` call
        vm.expectEmit(true, true, true, true, address(op));
        emit TransactionDeposited(l1MessengerAliased, address(L2Messenger), version, opaqueData);

        // SentMessage event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessage(address(L2Bridge), address(L1Bridge), message, nonce, 10000);

        // SentMessageExtension1 event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessageExtension1(address(L1Bridge), 0, 0);

        vm.prank(alice);
        L1Bridge.depositERC20(address(L1Token), address(L2Token), 100, 10000, hex"");
        assertEq(L1Bridge.deposits(address(L1Token), address(L2Token)), 100);
    }
}







contract PreBridgeMNT is Bridge_Initializer {
    function _preBridgeMNT(bool isLegacy) internal {
        assertEq(l1MNT.balanceOf(address(op)),0);
        uint256 nonce = L1Messenger.messageNonce();
        uint256 version = 0; // Internal constant in the OptimismPortal: DEPOSIT_VERSION
        address l1MessengerAliased = AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger));

        bytes memory message = abi.encodeWithSelector(
            L2StandardBridge.finalizeBridgeMNT.selector,
            alice,
            alice,
            500,
            hex"dead"
        );
        uint64 baseGas = L1Messenger.baseGas(message, 50000);

        deal(address(l1MNT), alice, 500, true);
        vm.store(address(l1MNT), bytes32(uint256(0x2)), bytes32(uint256(500))); //set total supply

        vm.prank(alice, alice);
        l1MNT.approve(address(L1Bridge), 500);
        vm.prank(alice, alice);


        if (isLegacy) {
            vm.expectCall(
                address(L1Bridge),
                abi.encodeWithSelector(L1Bridge.depositMNT.selector, 500, 50000, hex"dead")
            );
        } else {
            vm.expectCall(
                address(L1Bridge),
                abi.encodeWithSelector(L1Bridge.bridgeMNT.selector, 500, 50000, hex"dead")
            );
        }


        vm.expectCall(
            address(L1Messenger),
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                500,
                address(L2Bridge),
                message,
                50000
            )
        );


        bytes memory innerMessage = abi.encodeWithSelector(
            L2CrossDomainMessenger.relayMessage.selector,
            nonce,
            address(L1Bridge),
            address(L2Bridge),
            500,
            0,
            50000,
            message
        );

        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                OptimismPortal.depositTransaction.selector,
                500,
                address(L2Messenger),
                500,
                baseGas,
                false,
                innerMessage
            )
        );

        bytes memory opaqueData = abi.encodePacked(
            uint256(500),
            uint256(500),
            uint256(0),
            baseGas,
            false,
            innerMessage
        );


        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Transfer(alice, address(L1Bridge), 500);

        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Approval(address(L1Bridge), address(L1Messenger), 500);

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTDepositInitiated(alice, alice, 500, hex"dead");


        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTBridgeInitiated(alice, alice, 500, hex"dead");

        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Transfer(address(L1Bridge), address(L1Messenger), 500);

        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Approval(address(L1Messenger), address(op), 500);



        // OptimismPortal emits a TransactionDeposited event on `depositTransaction` call
        vm.expectEmit(true, true, true, true, address(op));
        emit TransactionDeposited(l1MessengerAliased, address(L2Messenger), version, opaqueData);

        // SentMessage event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessage(address(L2Bridge), address(L1Bridge), message, nonce, 50000);

        // SentMessageExtension1 event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessageExtension1(address(L1Bridge), 500, 0);


    }
}

contract L1StandardBridge_DepositMNT_Test is PreBridgeMNT {
    // depositMNT
    // - emits MNTDepositInitiated
    // - emits MNTBridgeInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - MNT ends up in the optimismPortal
    function test_depositMNT_succeeds() external {
        _preBridgeMNT({ isLegacy: true });
        L1Bridge.depositMNT(500, 50000, hex"dead");
        assertEq(l1MNT.balanceOf(address(op)), 500);
    }
}

contract L1StandardBridge_BridgeMNT_Test is PreBridgeMNT {
    // BridgeMNT
    // - emits MNTDepositInitiated
    // - emits MNTBridgeInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - MNT ends up in the optimismPortal
    function test_bridgeMNT_succeeds() external {
        _preBridgeMNT({ isLegacy: false });
        L1Bridge.bridgeMNT(500, 50000, hex"dead");
        assertEq(l1MNT.balanceOf(address(op)), 500);
    }
}

contract L1StandardBridge_DepositMNT_TestFail is Bridge_Initializer {
    function test_depositMNT_notEoa_reverts() external {
        // turn alice into a contract
        vm.etch(alice, address(L1Token).code);

        vm.expectRevert("StandardBridge: function can only be called from an EOA");
        vm.prank(alice);
        L1Bridge.depositMNT(1, 300, hex"");
    }
}

contract PreBridgeMNTTo is Bridge_Initializer {
    function _preBridgeMNTTo(bool isLegacy) internal {
        assertEq(l1MNT.balanceOf(address(op)), 0);
        uint256 nonce = L1Messenger.messageNonce();
        uint256 version = 0; // Internal constant in the OptimismPortal: DEPOSIT_VERSION
        address l1MessengerAliased = AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger));

        vm.deal(alice,1000);
        deal(address(l1MNT), alice, 600);
        vm.prank(alice);
        l1MNT.approve(address(L1Bridge), 600);


        if (isLegacy) {
            vm.expectCall(
                address(L1Bridge),
                abi.encodeWithSelector(L1Bridge.depositMNTTo.selector, bob, 600, 60000, hex"dead")
            );
        } else {
            vm.expectCall(
                address(L1Bridge),
                abi.encodeWithSelector(L1Bridge.bridgeMNTTo.selector, bob, 600, 60000, hex"dead")
            );
        }

        bytes memory message = abi.encodeWithSelector(
            L2StandardBridge.finalizeBridgeMNT.selector,
            alice,
            bob,
            600,
            hex"dead"
        );

        // the L1 bridge should call
        // L1CrossDomainMessenger.sendMessage
        vm.expectCall(
            address(L1Messenger),
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                600,
                address(L2Bridge),
                message,
                60000
            )
        );

        bytes memory innerMessage = abi.encodeWithSelector(
            L2CrossDomainMessenger.relayMessage.selector,
            nonce,
            address(L1Bridge),
            address(L2Bridge),
            600,
            0,
            60000,
            message
        );

        uint64 baseGas = L1Messenger.baseGas(message, 60000);
        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                OptimismPortal.depositTransaction.selector,
                600,
                address(L2Messenger),
                600,
                baseGas,
                false,
                innerMessage
            )
        );

        bytes memory opaqueData = abi.encodePacked(
            uint256(600),
            uint256(600),
            uint256(0),
            baseGas,
            false,
            innerMessage
        );


        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Transfer(alice, address(L1Bridge), 600);

        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Approval(address(L1Bridge), address(L1Messenger), 600);


        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTDepositInitiated(alice, bob, 600, hex"dead");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTBridgeInitiated(alice, bob, 600, hex"dead");

        // OptimismPortal emits a TransactionDeposited event on `depositTransaction` call
        vm.expectEmit(true, true, true, true, address(op));
        emit TransactionDeposited(l1MessengerAliased, address(L2Messenger), version, opaqueData);

        // SentMessage event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessage(address(L2Bridge), address(L1Bridge), message, nonce, 60000);

        // SentMessageExtension1 event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessageExtension1(address(L1Bridge), 600, 0);

        // deposit MNT to bob
        vm.prank(alice, alice);
    }
}

contract L1StandardBridge_DepositMNTTo_Test is PreBridgeMNTTo {
    // depositMNTTo
    // - emits MNTDepositInitiated
    // - calls optimismPortal.depositTransaction
    // - EOA or contract can call
    // - MNT ends up in the optimismPortal
    function test_depositMNTTo_succeeds() external {
        _preBridgeMNTTo({ isLegacy: true });
        L1Bridge.depositMNTTo(bob, 600, 60000, hex"dead");
        assertEq(l1MNT.balanceOf(address(op)), 600);
    }
}

contract L1StandardBridge_BridgeMNTTo_Test is PreBridgeMNTTo {
    // BridgeMNTTo
    // - emits MNTDepositInitiated
    // - emits MNTBridgeInitiated
    // - calls optimismPortal.depositTransaction
    // - only EOA
    // - MNT ends up in the optimismPortal
    function test_bridgeMNTTo_succeeds() external {
        _preBridgeMNTTo({ isLegacy: false });
        L1Bridge.bridgeMNTTo(bob, 600, 60000, hex"dead");
        assertEq(l1MNT.balanceOf(address(op)), 600);
    }
}

contract L1StandardBridge_DepositMNTTo_TestFail is Bridge_Initializer {}




contract L1StandardBridge_DepositERC20_TestFail is Bridge_Initializer {
    function test_depositERC20_notEoa_reverts() external {
        // turn alice into a contract
        vm.etch(alice, hex"ffff");

        vm.expectRevert("StandardBridge: function can only be called from an EOA");
        vm.prank(alice, alice);
        L1Bridge.depositERC20(address(0), address(0), 100, 100, hex"");
    }
}

contract L1StandardBridge_DepositERC20To_Test is Bridge_Initializer {
    // depositERC20To
    // - updates bridge.deposits
    // - emits ERC20DepositInitiated
    // - calls optimismPortal.depositTransaction
    // - callable by a contract
    function test_depositERC20To_succeeds() external {
        uint256 nonce = L1Messenger.messageNonce();
        uint256 version = 0; // Internal constant in the OptimismPortal: DEPOSIT_VERSION
        address l1MessengerAliased = AddressAliasHelper.applyL1ToL2Alias(address(L1Messenger));

        deal(address(L1Token), alice, 100000, true);
        vm.store(address(L1Token), bytes32(uint256(0x2)), bytes32(uint256(100000))); //set total supply

        vm.prank(alice);
        L1Token.approve(address(L1Bridge), type(uint256).max);

        bytes memory message = abi.encodeWithSelector(
            L2StandardBridge.finalizeBridgeERC20.selector,
            address(L2Token),
            address(L1Token),
            alice,
            bob,
            1000,
            hex""
        );

        // the L1 bridge should call L1CrossDomainMessenger.sendMessage
        vm.expectCall(
            address(L1Messenger),
            abi.encodeWithSelector(
                L1CrossDomainMessenger.sendMessageMantleBedrock.selector,
                0,
                address(L2Bridge),
                message,
                10000
            )
        );

        bytes memory innerMessage = abi.encodeWithSelector(
            L2CrossDomainMessenger.relayMessage.selector,
            nonce,
            address(L1Bridge),
            address(L2Bridge),
            0,
            0,
            10000,
            message
        );

        uint64 baseGas = L1Messenger.baseGas(message, 10000);
        vm.expectCall(
            address(op),
            abi.encodeWithSelector(
                OptimismPortal.depositTransaction.selector,
                0,
                address(L2Messenger),
                0,
                baseGas,
                false,
                innerMessage
            )
        );

        bytes memory opaqueData = abi.encodePacked(
            uint256(0),
            uint256(0),
            uint256(0),
            baseGas,
            false,
            innerMessage
        );

        // Should emit both the bedrock and legacy events
        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ERC20DepositInitiated(address(L1Token), address(L2Token), alice, bob, 1000, hex"");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ERC20BridgeInitiated(address(L1Token), address(L2Token), alice, bob, 1000, hex"");

        // OptimismPortal emits a TransactionDeposited event on `depositTransaction` call
        vm.expectEmit(true, true, true, true, address(op));
        emit TransactionDeposited(l1MessengerAliased, address(L2Messenger), version, opaqueData);

        // SentMessage event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessage(address(L2Bridge), address(L1Bridge), message, nonce, 10000);

        // SentMessageExtension1 event emitted by the CrossDomainMessenger
        vm.expectEmit(true, true, true, true, address(L1Messenger));
        emit SentMessageExtension1(address(L1Bridge), 0, 0);

        vm.expectCall(
            address(L1Token),
            abi.encodeWithSelector(ERC20.transferFrom.selector, alice, address(L1Bridge), 1000)
        );

        vm.prank(alice);
        L1Bridge.depositERC20To(address(L1Token), address(L2Token), bob, 1000, 10000, hex"");

        assertEq(L1Bridge.deposits(address(L1Token), address(L2Token)), 1000);
    }
}

contract L1StandardBridge_FinalizeETHWithdrawal_Test is Bridge_Initializer {
    using stdStorage for StdStorage;

    // finalizeETHWithdrawal
    // - emits ETHWithdrawalFinalized
    // - only callable by L2 bridge
    function test_finalizeETHWithdrawal_succeeds() external {
        uint256 aliceBalance = alice.balance;

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHWithdrawalFinalized(alice, alice, 100, hex"");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHBridgeFinalized(alice, alice, 100, hex"");

        vm.expectCall(alice, hex"");

        vm.mockCall(
            address(L1Bridge.messenger()),
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        // ensure that the messenger has ETH to call with
        vm.deal(address(L1Bridge.messenger()), 100);
        vm.prank(address(L1Bridge.messenger()));
        L1Bridge.finalizeETHWithdrawal{ value: 100 }(alice, alice, 100, hex"");

        assertEq(address(L1Bridge.messenger()).balance, 0);
        assertEq(aliceBalance + 100, alice.balance);
    }
}

contract L1StandardBridge_FinalizeETHWithdrawal_TestFail is Bridge_Initializer {}


contract L1StandardBridge_FinalizeMNTWithdrawal_Test is Bridge_Initializer {
    using stdStorage for StdStorage;

    // finalizeMNTWithdrawal
    // - emits MNTWithdrawalFinalized
    // - only callable by L2 bridge
    function test_finalizeMNTWithdrawal_succeeds() external {
        uint256 aliceBalance = l1MNT.balanceOf(alice);


        dealL1MNT(address(L1Bridge.messenger()), 100);
        vm.prank(address(L1Bridge.messenger()));
        l1MNT.approve(address(L1Bridge), 100);

        vm.mockCall(
            address(L1Bridge.messenger()),
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );



        vm.expectEmit(true, true, true, true, address(l1MNT));
        emit Transfer(address(L1Messenger), alice, 100);

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTWithdrawalFinalized(alice, alice, 100, hex"");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTBridgeFinalized(alice, alice, 100, hex"");

        // ensure that the messenger has MNT to call with
        vm.prank(address(L1Messenger));
        L1Bridge.finalizeMantleWithdrawal(alice, alice, 100, hex"");

        assertEq(l1MNT.balanceOf(address(L1Bridge.messenger())), 0);
        assertEq(aliceBalance + 100, l1MNT.balanceOf(alice));
    }
}

contract L1StandardBridge_FinalizeMNTWithdrawal_TestFail is Bridge_Initializer {}






contract L1StandardBridge_FinalizeERC20Withdrawal_Test is Bridge_Initializer {
    using stdStorage for StdStorage;

    // finalizeERC20Withdrawal
    // - updates bridge.deposits
    // - emits ERC20WithdrawalFinalized
    // - only callable by L2 bridge
    function test_finalizeERC20Withdrawal_succeeds() external {
        deal(address(L1Token), address(L1Bridge), 100, true);

        uint256 slot = stdstore
            .target(address(L1Bridge))
            .sig("deposits(address,address)")
            .with_key(address(L1Token))
            .with_key(address(L2Token))
            .find();

        // Give the L1 bridge some ERC20 tokens
        vm.store(address(L1Bridge), bytes32(slot), bytes32(uint256(100)));
        assertEq(L1Bridge.deposits(address(L1Token), address(L2Token)), 100);

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ERC20WithdrawalFinalized(address(L1Token), address(L2Token), alice, alice, 100, hex"");

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ERC20BridgeFinalized(address(L1Token), address(L2Token), alice, alice, 100, hex"");

        vm.expectCall(
            address(L1Token),
            abi.encodeWithSelector(ERC20.transfer.selector, alice, 100)
        );

        vm.mockCall(
            address(L1Bridge.messenger()),
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.prank(address(L1Bridge.messenger()));
        L1Bridge.finalizeERC20Withdrawal(
            address(L1Token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );

        assertEq(L1Token.balanceOf(address(L1Bridge)), 0);
        assertEq(L1Token.balanceOf(address(alice)), 100);
    }
}

contract L1StandardBridge_FinalizeERC20Withdrawal_TestFail is Bridge_Initializer {
    function test_finalizeERC20Withdrawal_notMessenger_reverts() external {
        vm.mockCall(
            address(L1Bridge.messenger()),
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.prank(address(28));
        vm.expectRevert("StandardBridge: function can only be called from the other bridge");
        L1Bridge.finalizeERC20Withdrawal(
            address(L1Token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );
    }

    function test_finalizeERC20Withdrawal_notOtherBridge_reverts() external {
        vm.mockCall(
            address(L1Bridge.messenger()),
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(address(0)))
        );
        vm.prank(address(L1Bridge.messenger()));
        vm.expectRevert("StandardBridge: function can only be called from the other bridge");
        L1Bridge.finalizeERC20Withdrawal(
            address(L1Token),
            address(L2Token),
            alice,
            alice,
            100,
            hex""
        );
    }
}

contract L1StandardBridge_FinalizeBridgeETH_Test is Bridge_Initializer {
    function test_finalizeBridgeETH_succeeds() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.deal(messenger, 100);
        vm.prank(messenger);

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit ETHBridgeFinalized(alice, alice, 100, hex"");

        L1Bridge.finalizeBridgeETH{ value: 100 }(alice, alice, 100, hex"");
    }
}

contract L1StandardBridge_FinalizeBridgeETH_TestFail is Bridge_Initializer {
    function test_finalizeBridgeETH_incorrectValue_reverts() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.deal(messenger, 100);
        vm.prank(messenger);
        vm.expectRevert("StandardBridge: amount sent does not match amount required");
        L1Bridge.finalizeBridgeETH{ value: 50 }(alice, alice, 100, hex"");
    }

    function test_finalizeBridgeETH_sendToSelf_reverts() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.deal(messenger, 100);
        vm.prank(messenger);
        vm.expectRevert("StandardBridge: cannot send to self");
        L1Bridge.finalizeBridgeETH{ value: 100 }(alice, address(L1Bridge), 100, hex"");
    }

    function test_finalizeBridgeETH_sendToMessenger_reverts() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.deal(messenger, 100);
        vm.prank(messenger);
        vm.expectRevert("StandardBridge: cannot send to messenger");
        L1Bridge.finalizeBridgeETH{ value: 100 }(alice, messenger, 100, hex"");
    }
}

contract L1StandardBridge_FinalizeBridgeMNT_Test is Bridge_Initializer {
    function test_finalizeBridgeMNT_succeeds() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        vm.deal(messenger, 100);
        deal(address(l1MNT),messenger,100);
        vm.prank(messenger);
        l1MNT.approve(address(L1Bridge),100);
        vm.prank(messenger);

        vm.expectEmit(true, true, true, true, address(L1Bridge));
        emit MNTBridgeFinalized(alice, alice, 100, hex"");

        L1Bridge.finalizeBridgeMNT(alice, alice, 100, hex"");
    }
}

contract L1StandardBridge_FinalizeBridgeMNT_TestFail is Bridge_Initializer {


    function test_finalizeBridgeMNT_sendToSelf_reverts() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        deal(address(l1MNT),messenger,100);
        vm.prank(messenger);
        l1MNT.approve(address(L1Bridge),100);
        vm.deal(messenger, 100);
        vm.prank(messenger);
        vm.expectRevert("StandardBridge: cannot send to self");
        L1Bridge.finalizeBridgeMNT(alice, address(L1Bridge), 100, hex"");
    }

    function test_finalizeBridgeMNT_sendToMessenger_reverts() external {
        address messenger = address(L1Bridge.messenger());
        vm.mockCall(
            messenger,
            abi.encodeWithSelector(CrossDomainMessenger.xDomainMessageSender.selector),
            abi.encode(address(L1Bridge.OTHER_BRIDGE()))
        );
        deal(address(l1MNT),messenger,100);
        vm.prank(messenger);
        l1MNT.approve(address(L1Bridge),100);
        vm.deal(messenger, 100);
        vm.prank(messenger);
        vm.expectRevert("StandardBridge: cannot send to messenger");
        L1Bridge.finalizeBridgeMNT(alice, messenger, 100, hex"");
    }

}
