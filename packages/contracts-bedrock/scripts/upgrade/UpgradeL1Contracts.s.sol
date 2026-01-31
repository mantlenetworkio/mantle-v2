// SPDX-License-Identifier: MIT
pragma solidity ^0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { SystemConfig } from "src/L1/SystemConfig.sol";
import { ResourceMetering } from "src/L1/ResourceMetering.sol";

/// @title UpgradeL1Contracts
/// @notice Script to upgrade L1 SystemConfig contract for Mantle Limb to Arsia upgrade
/// @dev This script only upgrades SystemConfig (1.3.0 → 1.4.0)
///      L2OutputOracle and OptimismPortal remain unchanged as their versions are the same
contract UpgradeL1Contracts is Script {
    /// @notice Deploys new SystemConfig implementation and upgrades proxy (all-in-one with default params)
    /// @param proxyAdmin Address of the ProxyAdmin contract
    /// @param systemConfigProxy Address of the SystemConfig proxy to read current config
    /// @dev Constructor values default to 0. Use setter methods (setBaseFee, setGasConfigArsia) after upgrade to set
    /// actual values and emit ConfigUpdate events for op-node.
    function run(address proxyAdmin, address systemConfigProxy) public {
        // Use default values (0) - setter methods should be called after upgrade
        run(proxyAdmin, systemConfigProxy, 0, 0, 0);
    }

    /// @notice Deploys new SystemConfig implementation and upgrades proxy (all-in-one with custom params)
    /// @param proxyAdmin Address of the ProxyAdmin contract
    /// @param systemConfigProxy Address of the SystemConfig proxy to read current config
    /// @param baseFeeScalar Base fee scalar for Arsia (default: 0, should be set via setGasConfigArsia after upgrade)
    /// @param blobBaseFeeScalar Blob base fee scalar for Arsia (default: 0, should be set via setGasConfigArsia after
    /// upgrade)
    /// @param baseFee Initial base fee in wei (default: 0, should be set via setBaseFee after upgrade)
    /// @dev Note: Constructor values are stored but do not emit ConfigUpdate events.
    ///      Call setBaseFee() and setGasConfigArsia() after upgrade to emit events for op-node to detect changes.
    function run(
        address proxyAdmin,
        address systemConfigProxy,
        uint32 baseFeeScalar,
        uint32 blobBaseFeeScalar,
        uint256 baseFee
    )
        public
    {
        require(proxyAdmin != address(0), "UpgradeL1Contracts: Invalid ProxyAdmin address");
        require(systemConfigProxy != address(0), "UpgradeL1Contracts: Invalid SystemConfig proxy");

        console.log("=== Mantle Limb to Arsia L1 Contract Upgrade (All-in-One) ===");
        console.log("ProxyAdmin:", proxyAdmin);
        console.log("SystemConfig Proxy:", systemConfigProxy);
        console.log("");

        // Read current configuration from the proxy
        SystemConfig currentConfig = SystemConfig(systemConfigProxy);

        address owner = currentConfig.owner();
        bytes32 batcherHash = currentConfig.batcherHash();
        uint64 gasLimit = currentConfig.gasLimit();
        address unsafeBlockSigner = currentConfig.unsafeBlockSigner();
        ResourceMetering.ResourceConfig memory resourceConfig = currentConfig.resourceConfig();

        console.log("Current Configuration:");
        console.log("  owner:", owner);
        console.log("  batcherHash:", uint256(batcherHash));
        console.log("  gasLimit:", gasLimit);
        console.log("  unsafeBlockSigner:", unsafeBlockSigner);
        console.log("");

        console.log("Arsia Configuration (constructor values):");
        console.log("  baseFeeScalar:", baseFeeScalar);
        console.log("  blobBaseFeeScalar:", blobBaseFeeScalar);
        console.log("  baseFee:", baseFee);
        if (baseFeeScalar == 0 && blobBaseFeeScalar == 0 && baseFee == 0) {
            console.log(
                "  Note: Values are 0. Call setBaseFee() and setGasConfigArsia() after upgrade to set actual values and emit ConfigUpdate events."
            );
        }
        console.log("  ResourceConfig: Preserved from current configuration");
        console.log("");

        vm.startBroadcast();

        // Deploy new SystemConfig implementation
        console.log("Deploying new SystemConfig implementation...");
        SystemConfig newImpl = new SystemConfig({
            _owner: owner,
            _basefeeScalar: baseFeeScalar,
            _blobbasefeeScalar: blobBaseFeeScalar,
            _batcherHash: batcherHash,
            _gasLimit: gasLimit,
            _baseFee: baseFee,
            _unsafeBlockSigner: unsafeBlockSigner,
            _config: resourceConfig
        });

        console.log("New SystemConfig Implementation:", address(newImpl));
        console.log("");

        // Upgrade the proxy
        IProxyAdmin admin = IProxyAdmin(proxyAdmin);
        address oldImpl = _getProxyImplementation(proxyAdmin, systemConfigProxy);
        console.log("Old Implementation:", oldImpl);
        console.log("Upgrading proxy to new implementation...");
        admin.upgrade(payable(systemConfigProxy), address(newImpl));
        console.log("SystemConfig upgraded successfully!");

        vm.stopBroadcast();

        console.log("");
        console.log("=== Upgrade Complete ===");
        console.log("Note: L2OutputOracle and OptimismPortal were NOT upgraded (versions unchanged)");
        console.log("");
        console.log("Important: Constructor values are stored but do not emit ConfigUpdate events.");
        console.log("To make op-node detect the changes, call setter methods after upgrade:");
        console.log("  - systemConfig.setBaseFee(<value>)");
        console.log("  - systemConfig.setGasConfigArsia(<baseFeeScalar>, <blobBaseFeeScalar>)");
        console.log("");

        // Verify upgrade
        address actualImpl = _getProxyImplementation(proxyAdmin, systemConfigProxy);
        string memory version = currentConfig.version();

        console.log("Verification:");
        console.log("  New Implementation:", actualImpl);
        console.log("  Expected:", address(newImpl));
        console.log("  Version:", version);

        if (actualImpl == address(newImpl)) {
            console.log("  Status: SUCCESS");
        } else {
            console.log("  Status: FAILED");
            revert("UpgradeL1Contracts: Implementation mismatch after upgrade");
        }
    }

    /// @notice Upgrades SystemConfig implementation through ProxyAdmin (existing implementation)
    /// @param proxyAdmin Address of the ProxyAdmin contract
    /// @param systemConfigProxy Address of the SystemConfig proxy
    /// @param systemConfigImpl Address of the new SystemConfig implementation
    function run(address proxyAdmin, address systemConfigProxy, address systemConfigImpl) public {
        // Validate inputs
        require(proxyAdmin != address(0), "UpgradeL1Contracts: Invalid ProxyAdmin address");
        require(systemConfigProxy != address(0), "UpgradeL1Contracts: Invalid SystemConfig proxy");
        require(systemConfigImpl != address(0), "UpgradeL1Contracts: Invalid SystemConfig impl");

        console.log("=== Mantle Limb to Arsia L1 Contract Upgrade ===");
        console.log("ProxyAdmin:", proxyAdmin);
        console.log("SystemConfig Proxy:", systemConfigProxy);
        console.log("SystemConfig New Impl:", systemConfigImpl);
        console.log("");

        // Query old implementation before upgrade
        IProxyAdmin admin = IProxyAdmin(proxyAdmin);
        address oldImpl = _getProxyImplementation(proxyAdmin, systemConfigProxy);
        console.log("SystemConfig Old Impl:", oldImpl);
        console.log("");

        vm.startBroadcast();

        // Only upgrade SystemConfig (1.3.0 -> 1.4.0)
        console.log("Upgrading SystemConfig...");
        admin.upgrade(payable(systemConfigProxy), systemConfigImpl);
        console.log("SystemConfig upgraded successfully!");

        vm.stopBroadcast();

        console.log("");
        console.log("=== Upgrade Complete ===");
        console.log("Note: L2OutputOracle and OptimismPortal were NOT upgraded (versions unchanged)");
        console.log("");

        // Verify upgrade
        address newImpl = _getProxyImplementation(proxyAdmin, systemConfigProxy);
        console.log("Verification:");
        console.log("  New Implementation:", newImpl);
        console.log("  Expected:", systemConfigImpl);

        if (newImpl == systemConfigImpl) {
            console.log("  Status: SUCCESS");
        } else {
            console.log("  Status: FAILED");
            revert("UpgradeL1Contracts: Implementation mismatch after upgrade");
        }
    }

    /// @notice Internal helper to get proxy implementation address via ProxyAdmin
    /// @param _proxyAdmin Address of the ProxyAdmin contract
    /// @param proxy Address of the proxy
    /// @return impl Address of the current implementation
    function _getProxyImplementation(address _proxyAdmin, address proxy) internal view returns (address impl) {
        // Try to get implementation through ProxyAdmin first (more reliable)
        IProxyAdmin admin = IProxyAdmin(_proxyAdmin);

        // ProxyAdmin has getProxyImplementation(address) function
        try admin.getProxyImplementation(proxy) returns (address implementation) {
            return implementation;
        } catch {
            // Fallback: Try ERC1967 implementation slot
            bytes32 slot = 0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc;
            assembly {
                impl := sload(slot)
            }
        }
    }
}
