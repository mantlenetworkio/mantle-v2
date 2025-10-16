// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IResolvedDelegateProxy } from "interfaces/legacy/IResolvedDelegateProxy.sol";
import { IL1ChugSplashProxy } from "interfaces/legacy/IL1ChugSplashProxy.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";
import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";
import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IOwnable } from "interfaces/universal/IOwnable.sol";
import { IAddressManager } from "interfaces/legacy/IAddressManager.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

contract DeployProxies is Script {
    struct Output {
        IAddressManager addressManager;
        IProxyAdmin proxyAdmin;
        IL1StandardBridge l1StandardBridgeProxy;
        IL2OutputOracle l2OutputOracleProxy;
        IL1CrossDomainMessenger l1CrossDomainMessengerProxy;
        IOptimismPortal optimismPortalProxy;
        IOptimismMintableERC20Factory optimismMintableERC20FactoryProxy;
        IL1ERC721Bridge l1ERC721BridgeProxy;
        ISystemConfig systemConfigProxy;
    }

    function run(address _deployer) public returns (Output memory output_) {
        vm.startBroadcast(_deployer);
        // deploy AddressManager
        output_.addressManager = IAddressManager(deployAddressManager());

        // deploy ProxyAdmin
        output_.proxyAdmin = IProxyAdmin(deployProxyAdmin(_deployer));
        output_.proxyAdmin.setAddressManager(output_.addressManager);

        // deploy ERC1967Proxies
        output_.l2OutputOracleProxy = IL2OutputOracle(deployERC1967Proxy(output_.proxyAdmin));
        output_.optimismPortalProxy = IOptimismPortal(payable(deployERC1967Proxy(output_.proxyAdmin)));
        output_.optimismMintableERC20FactoryProxy =
            IOptimismMintableERC20Factory(deployERC1967Proxy(output_.proxyAdmin));
        output_.l1ERC721BridgeProxy = IL1ERC721Bridge(deployERC1967Proxy(output_.proxyAdmin));
        output_.systemConfigProxy = ISystemConfig(deployERC1967Proxy(output_.proxyAdmin));

        // deploy legacy proxies
        output_.l1StandardBridgeProxy = IL1StandardBridge(deployL1ChugSplashProxy(output_.proxyAdmin));
        output_.proxyAdmin.setProxyType(address(output_.l1StandardBridgeProxy), IProxyAdmin.ProxyType.CHUGSPLASH);

        output_.l1CrossDomainMessengerProxy =
            IL1CrossDomainMessenger(deployResolvedDelegateProxy(output_.addressManager, "BVM_L1CrossDomainMessenger"));
        output_.proxyAdmin.setProxyType(address(output_.l1CrossDomainMessengerProxy), IProxyAdmin.ProxyType.RESOLVED);
        output_.proxyAdmin.setImplementationName(
            address(output_.l1CrossDomainMessengerProxy), "BVM_L1CrossDomainMessenger"
        );

        // transfer ownership of AddressManager to ProxyAdmin
        IOwnable(address(output_.addressManager)).transferOwnership(address(output_.proxyAdmin));
        vm.stopBroadcast();

        vm.label(address(output_.proxyAdmin), "ProxyAdmin");
        vm.label(address(output_.addressManager), "AddressManager");
        vm.label(address(output_.l1StandardBridgeProxy), "L1StandardBridgeProxy");
        vm.label(address(output_.l2OutputOracleProxy), "L2OutputOracleProxy");
        vm.label(address(output_.l1CrossDomainMessengerProxy), "L1CrossDomainMessengerProxy");
        vm.label(address(output_.optimismPortalProxy), "OptimismPortalProxy");
        vm.label(address(output_.optimismMintableERC20FactoryProxy), "OptimismMintableERC20FactoryProxy");
        vm.label(address(output_.l1ERC721BridgeProxy), "L1ERC721BridgeProxy");
        vm.label(address(output_.systemConfigProxy), "SystemConfigProxy");

        return output_;
    }

    function deployProxyAdmin(address _owner) private returns (IProxyAdmin proxyAdmin_) {
        proxyAdmin_ = IProxyAdmin(
            DeployUtils.create1({
                _name: "ProxyAdmin",
                _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxyAdmin.__constructor__, (_owner)))
            })
        );
    }

    function deployERC1967Proxy(IProxyAdmin _proxyAdmin) private returns (address proxy_) {
        proxy_ = DeployUtils.create1({
            _name: "Proxy",
            _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (address(_proxyAdmin))))
        });
    }

    function deployL1ChugSplashProxy(IProxyAdmin _proxyAdmin) private returns (address proxy_) {
        proxy_ = DeployUtils.create1({
            _name: "L1ChugSplashProxy",
            _args: DeployUtils.encodeConstructor(abi.encodeCall(IL1ChugSplashProxy.__constructor__, (address(_proxyAdmin))))
        });
    }

    function deployResolvedDelegateProxy(
        IAddressManager _addressManager,
        string memory _implementationName
    )
        private
        returns (address proxy_)
    {
        proxy_ = DeployUtils.create1({
            _name: "ResolvedDelegateProxy",
            _args: DeployUtils.encodeConstructor(
                abi.encodeCall(IResolvedDelegateProxy.__constructor__, (_addressManager, _implementationName))
            )
        });
    }

    function deployAddressManager() private returns (address addressManager_) {
        addressManager_ = DeployUtils.create1({ _name: "AddressManager", _args: new bytes(0) });
    }
}
