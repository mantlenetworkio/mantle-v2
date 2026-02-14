// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

import { Preinstalls } from "src/libraries/Preinstalls.sol";
import { MantlePreinstalls } from "src/libraries/MantlePreinstalls.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";

/// @title SetPreinstalls
/// @notice Sets all preinstalls in the VM state. There is no default "run()" entrypoint,
/// as this is used in L2Genesis.s.sol, and standalone in the Go test setup for L1 state.
contract SetPreinstalls is Script {
    /// @notice Sets all the preinstalls.
    ///         Warning: the creator-accounts of the preinstall contracts have 0 nonce values.
    ///         When performing a regular user-initiated contract-creation of a preinstall,
    ///         the creation will fail (but nonce will be bumped and not blocked).
    ///         The preinstalls themselves are all inserted with a nonce of 1, reflecting regular user execution.
    function setPreinstalls() public {
        _setPreinstallCode(Preinstalls.MultiCall3);
        _setPreinstallCode(Preinstalls.Create2Deployer);
        _setPreinstallCode(Preinstalls.Safe_v130);
        _setPreinstallCode(Preinstalls.SafeL2_v130);
        _setPreinstallCode(Preinstalls.MultiSendCallOnly_v130);
        _setPreinstallCode(Preinstalls.SafeSingletonFactory);
        _setPreinstallCode(Preinstalls.DeterministicDeploymentProxy);
        _setPreinstallCode(Preinstalls.MultiSend_v130);
        _setPreinstallCode(Preinstalls.Permit2);
        _setPreinstallCode(Preinstalls.SenderCreator_v060); // ERC 4337 v0.6.0
        _setPreinstallCode(Preinstalls.EntryPoint_v060); // ERC 4337 v0.6.0
        _setPreinstallCode(Preinstalls.SenderCreator_v070); // ERC 4337 v0.7.0
        _setPreinstallCode(Preinstalls.EntryPoint_v070); // ERC 4337 v0.7.0
        _setPreinstallCode(Preinstalls.BeaconBlockRoots);
        _setPreinstallCode(Preinstalls.HistoryStorage); // EIP-2935
        _setPreinstallCode(Preinstalls.CreateX);
        // 4788 sender nonce must be incremented, since it's part of later upgrade-transactions.
        // For the upgrade-tx to not create a contract that conflicts with an already-existing copy,
        // the nonce must be bumped.
        vm.setNonce(Preinstalls.BeaconBlockRootsSender, 1);
        vm.setNonce(Preinstalls.HistoryStorageSender, 1);
    }

    // @notice Sets the mantle preinstalls on L1 for testing.
    // @dev Since token ratio is embedded too deeply in mantle geth, in order to new a L1 EL node by mantle geth,
    //      we need to set a test-only gas price oracle on L1 and set token ratio to 1.
    function setMantlePreinstalls() public {
        _setMantlePreinstallCode(MantlePreinstalls.L1MNTProxy);
        _setMantlePreinstallCode(MantlePreinstalls.L1MNTImpl);

        // Set the gas price oracle code and set token ratio to 1.
        string memory cname = Predeploys.getName(Predeploys.GAS_PRICE_ORACLE);
        vm.etch(Predeploys.GAS_PRICE_ORACLE, vm.getDeployedCode(string.concat(cname, ".sol:", cname)));
        vm.store(Predeploys.GAS_PRICE_ORACLE, bytes32(uint256(0)), bytes32(uint256(1)));
    }

    /// @notice Sets the bytecode in state
    function _setPreinstallCode(address _addr) internal {
        string memory cname = Preinstalls.getName(_addr);
        console.log("Setting %s preinstall code at: %s", cname, _addr);
        vm.etch(_addr, Preinstalls.getDeployedCode(_addr, block.chainid));
        // during testing in a shared L1/L2 account namespace some preinstalls may already have been inserted and used.
        if (vm.getNonce(_addr) == 0) {
            vm.setNonce(_addr, 1);
        }
    }

    function _setMantlePreinstallCode(address _addr) internal {
        string memory cname = MantlePreinstalls.getName(_addr);
        console.log("Setting %s mantle preinstall code at: %s", cname, _addr);
        vm.etch(_addr, MantlePreinstalls.getDeployedCode(_addr, block.chainid));
        // during testing in a shared L1/L2 account namespace some preinstalls may already have been inserted and used.
        if (vm.getNonce(_addr) == 0) {
            vm.setNonce(_addr, 1);
        }
        // Set storage slots if any are defined
        MantlePreinstalls.StorageSlot[] memory slots = MantlePreinstalls.getStorageSlots(_addr);
        for (uint256 i = 0; i < slots.length; i++) {
            vm.store(_addr, slots[i].key, slots[i].value);
        }
    }
}
