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
    /// @notice Deploys new SystemConfig implementation and upgrades proxy
    /// @param proxyAdmin Address of the ProxyAdmin contract
    /// @param systemConfigProxy Address of the SystemConfig proxy
    function run(address proxyAdmin, address systemConfigProxy) public {
        // Deploy implementation and upgrade in order
        address impl = deployImplementation();
        upgrade(proxyAdmin, systemConfigProxy, impl);
    }

    /// @notice Deploys new SystemConfig implementation with minimal values
    /// @return impl Address of the deployed implementation
    /// @dev Implementation storage is never used by proxy (delegatecall uses proxy's storage).
    ///      Uses minimal values to satisfy validation and minimize deployment gas cost.
    function deployImplementation() public returns (address impl) {
        console.log("=== Deploy SystemConfig Implementation ===");

        vm.startBroadcast();

        // Deploy with minimal values that satisfy constructor requirements:
        // - baseFeeMaxChangeDenominator > 1 (min: 2)
        // - elasticityMultiplier > 0 (min: 1)
        SystemConfig newImpl = new SystemConfig({
            _owner: address(1),
            _basefeeScalar: 0,
            _blobbasefeeScalar: 0,
            _batcherHash: bytes32(0),
            _gasLimit: 0,
            _baseFee: 0,
            _unsafeBlockSigner: address(0),
            _config: ResourceMetering.ResourceConfig({
                maxResourceLimit: 0,
                elasticityMultiplier: 1,
                baseFeeMaxChangeDenominator: 2,
                minimumBaseFee: 0,
                systemTxMaxGas: 0,
                maximumBaseFee: 0
            })
        });

        impl = address(newImpl);
        console.log("New SystemConfig Implementation:", impl);

        vm.stopBroadcast();
    }

    /// @notice Upgrades SystemConfig proxy to new implementation
    /// @param proxyAdmin Address of the ProxyAdmin contract
    /// @param systemConfigProxy Address of the SystemConfig proxy
    /// @param systemConfigImpl Address of the new SystemConfig implementation
    function upgrade(address proxyAdmin, address systemConfigProxy, address systemConfigImpl) public {
        require(proxyAdmin != address(0), "UpgradeL1Contracts: Invalid ProxyAdmin address");
        require(systemConfigProxy != address(0), "UpgradeL1Contracts: Invalid SystemConfig proxy");
        require(systemConfigImpl != address(0), "UpgradeL1Contracts: Invalid SystemConfig impl");

        console.log("=== Upgrade SystemConfig Proxy ===");
        console.log("ProxyAdmin:", proxyAdmin);
        console.log("SystemConfig Proxy:", systemConfigProxy);
        console.log("SystemConfig New Impl:", systemConfigImpl);
        console.log("");

        // Query old implementation before upgrade
        address oldImpl = _getProxyImplementation(proxyAdmin, systemConfigProxy);
        console.log("SystemConfig Old Impl:", oldImpl);
        console.log("");

        vm.startBroadcast();

        // Upgrade SystemConfig (1.3.0 -> 1.4.0)
        console.log("Upgrading SystemConfig...");
        IProxyAdmin(proxyAdmin).upgrade(payable(systemConfigProxy), systemConfigImpl);
        console.log("SystemConfig upgraded successfully!");

        vm.stopBroadcast();

        console.log("");
        console.log("=== Upgrade Complete ===");
        console.log("Note: L2OutputOracle and OptimismPortal were NOT upgraded (versions unchanged)");
        console.log("");

        // Verify upgrade
        _verifyUpgrade(proxyAdmin, systemConfigProxy, systemConfigImpl);
    }

    /// @notice Verifies the upgrade was successful
    /// @param proxyAdmin Address of the ProxyAdmin contract
    /// @param systemConfigProxy Address of the SystemConfig proxy
    /// @param expectedImpl Expected implementation address
    function _verifyUpgrade(address proxyAdmin, address systemConfigProxy, address expectedImpl) internal view {
        address actualImpl = _getProxyImplementation(proxyAdmin, systemConfigProxy);
        string memory version = SystemConfig(systemConfigProxy).version();

        console.log("Verification:");
        console.log("  New Implementation:", actualImpl);
        console.log("  Expected:", expectedImpl);
        console.log("  Version:", version);

        if (actualImpl == expectedImpl) {
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
