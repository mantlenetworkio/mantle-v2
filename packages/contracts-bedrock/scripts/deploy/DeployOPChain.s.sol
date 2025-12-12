// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";

// Libraries
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { Constants } from "src/libraries/Constants.sol";

// Interfaces
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
import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";

contract DeployOPChain is Script {
    struct Input {
        IProxyAdmin proxyAdmin;
        IOptimismPortal optimismPortalImpl;
        IOptimismPortal optimismPortalProxy;
        ISystemConfig systemConfigImpl;
        ISystemConfig systemConfigProxy;
        IL1CrossDomainMessenger l1CrossDomainMessengerImpl;
        IL1CrossDomainMessenger l1CrossDomainMessengerProxy;
        IL1ERC721Bridge l1ERC721BridgeImpl;
        IL1ERC721Bridge l1ERC721BridgeProxy;
        IL1StandardBridge l1StandardBridgeImpl;
        IL1StandardBridge l1StandardBridgeProxy;
        IOptimismMintableERC20Factory optimismMintableERC20FactoryImpl;
        IOptimismMintableERC20Factory optimismMintableERC20FactoryProxy;
        IL2OutputOracle l2OutputOracleImpl;
        IL2OutputOracle l2OutputOracleProxy;
        address finalSystemOwner;
        uint32 basefeeScalar;
        uint32 blobbasefeeScalar;
        address batchSenderAddress;
        uint64 l2GenesisBlockGasLimit;
        uint256 l2GenesisBlockBaseFeePerGas;
        address p2pSequencerAddress;
        uint256 l2OutputOracleStartingBlockNumber;
        uint256 l2OutputOracleStartingTimestamp;
    }

    function run(Input memory _input) public {
        _run(_input, msg.sender);
    }

    function runWithDeployer(Input memory _input, address _deployer) public {
        _run(_input, _deployer);
    }

    function _run(Input memory _input, address _deployer) internal {
        _initializeProxies(_input, _deployer);

        address proxyAdmin = address(_input.proxyAdmin);
        address finalSystemOwner = _input.finalSystemOwner;
        vm.broadcast(_deployer);
        _transferOwnership(proxyAdmin, finalSystemOwner);
    }

    function _initializeProxies(Input memory _input, address _deployer) internal {
        console.log("Initializing proxies");

        vm.startBroadcast(_deployer);
        initializeSystemConfig(_input);
        initializeL1StandardBridge(_input);
        initializeL1ERC721Bridge(_input);
        initializeOptimismMintableERC20Factory(_input);
        initializeL1CrossDomainMessenger(_input);
        initializeL2OutputOracle(_input);
        initializeOptimismPortal(_input);
        vm.stopBroadcast();
    }

    function _transferOwnership(address _contract, address _newOwner) internal {
        IOwnable(_contract).transferOwnership(_newOwner);
    }

    /// @notice Returns the default resource config. We encourage using interface instead of the original contract.
    function defaultResourceConfig() public view returns (IResourceMetering.ResourceConfig memory) {
        return abi.decode(abi.encode(Constants.DEFAULT_RESOURCE_CONFIG()), (IResourceMetering.ResourceConfig));
    }

    ////////////////////////////////////////////////////////////////
    //                Proxy Initialization Functions              //
    ////////////////////////////////////////////////////////////////

    function initializeSystemConfig(Input memory _input) private {
        console.log("Initializing SystemConfig");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.systemConfigProxy);
        address impl = address(_input.systemConfigImpl);

        bytes memory data = abi.encodeCall(
            ISystemConfig.initialize,
            (
                _input.finalSystemOwner,
                _input.basefeeScalar,
                _input.blobbasefeeScalar,
                bytes32(uint256(uint160(_input.batchSenderAddress))),
                _input.l2GenesisBlockGasLimit,
                _input.l2GenesisBlockBaseFeePerGas,
                _input.p2pSequencerAddress,
                defaultResourceConfig()
            )
        );
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    function initializeL1StandardBridge(Input memory _input) private {
        console.log("Initializing L1StandardBridge");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.l1StandardBridgeProxy);
        address impl = address(_input.l1StandardBridgeImpl);

        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeL1ERC721Bridge(Input memory _input) private {
        console.log("Initializing L1ERC721Bridge");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.l1ERC721BridgeProxy);
        address impl = address(_input.l1ERC721BridgeImpl);

        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeOptimismMintableERC20Factory(Input memory _input) private {
        console.log("Initializing OptimismMintableERC20Factory");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.optimismMintableERC20FactoryProxy);
        address impl = address(_input.optimismMintableERC20FactoryImpl);

        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeL1CrossDomainMessenger(Input memory _input) private {
        console.log("Initializing L1CrossDomainMessenger");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.l1CrossDomainMessengerProxy);
        address impl = address(_input.l1CrossDomainMessengerImpl);

        bytes memory data = abi.encodeCall(IL1CrossDomainMessenger.initialize, ());
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    function initializeL2OutputOracle(Input memory _input) private {
        console.log("Initializing L2OutputOracle");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.l2OutputOracleProxy);
        address impl = address(_input.l2OutputOracleImpl);

        bytes memory data = abi.encodeCall(
            IL2OutputOracle.initialize,
            (_input.l2OutputOracleStartingBlockNumber, _input.l2OutputOracleStartingTimestamp)
        );
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    function initializeOptimismPortal(Input memory _input) private {
        console.log("Initializing OptimismPortal");
        IProxyAdmin proxyAdmin = _input.proxyAdmin;
        address proxy = address(_input.optimismPortalProxy);
        address impl = address(_input.optimismPortalImpl);

        bytes memory data = abi.encodeCall(IOptimismPortal.initialize, (false));
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }
}
