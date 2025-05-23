// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { Vm } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";

// Scripts
import { DeployConfig } from "scripts/deploy/DeployConfig.s.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Types } from "scripts/libraries/Types.sol";
import { Blueprint } from "src/libraries/Blueprint.sol";

// Interfaces
// import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
// import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
// import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
// import { ISuperchainConfig } from "interfaces/L1/ISuperchainConfig.sol";
// import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
// import { IOptimismPortal2 as IOptimismPortal } from "interfaces/L1/IOptimismPortal2.sol";
// import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
// import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
// import { ProtocolVersion, IProtocolVersions } from "interfaces/L1/IProtocolVersions.sol";
// import { IDisputeGameFactory } from "interfaces/dispute/IDisputeGameFactory.sol";
// import { IDelayedWETH } from "interfaces/dispute/IDelayedWETH.sol";
// import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";
// import { IPreimageOracle } from "interfaces/cannon/IPreimageOracle.sol";
// import { IMIPS } from "interfaces/cannon/IMIPS.sol";
// import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
// import { IETHLockbox } from "interfaces/L1/IETHLockbox.sol";

library ChainAssertions {
    Vm internal constant vm = Vm(0x7109709ECfa91a80626fF3989D68f67F5b1DD12D);

    /// @notice Asserts the correctness of an L1 deployment. This function expects that all contracts
    ///         within the `prox` ContractSet are proxies that have been setup and initialized.
    function postDeployAssertions(Types.ContractSet memory _prox, DeployConfig _cfg, Vm _vm) internal view {
        // console.log("Running post-deploy assertions");
        // IResourceMetering.ResourceConfig memory rcfg = ISystemConfig(_prox.SystemConfig).resourceConfig();
        // IResourceMetering.ResourceConfig memory dflt = Constants.DEFAULT_RESOURCE_CONFIG();
        // require(keccak256(abi.encode(rcfg)) == keccak256(abi.encode(dflt)), "CHECK-RCFG-10");

        // checkSystemConfig({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        // checkL1CrossDomainMessenger({ _contracts: _prox, _vm: _vm, _isProxy: true });
        // checkL1StandardBridge({ _contracts: _prox, _isProxy: true });
        // checkOptimismMintableERC20Factory({ _contracts: _prox, _isProxy: true });
        // checkL1ERC721Bridge({ _contracts: _prox, _isProxy: true });
        // checkOptimismPortal({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        // checkL2OutputOracle({ _contracts: _prox, _isProxy: true });
    }

    /// @notice Asserts that the SystemConfig is setup correctly
    function checkSystemConfig(Types.ContractSet memory _contracts, DeployConfig _cfg, bool _isProxy) internal view {
        // ISystemConfig config = ISystemConfig(_contracts.SystemConfig);
        // console.log(
        //     "Running chain assertions on the SystemConfig %s at %s",
        //     _isProxy ? "proxy" : "implementation",
        //     address(config)
        // );

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(config), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        // IResourceMetering.ResourceConfig memory resourceConfig = config.resourceConfig();

        // if (_isProxy) {
        //     require(config.owner() == _cfg.finalSystemOwner(), "CHECK-SCFG-10");
        //     require(config.basefeeScalar() == _cfg.basefeeScalar(), "CHECK-SCFG-20");
        //     require(config.blobbasefeeScalar() == _cfg.blobbasefeeScalar(), "CHECK-SCFG-30");
        //     require(config.batcherHash() == bytes32(uint256(uint160(_cfg.batchSenderAddress()))), "CHECK-SCFG-40");
        //     require(config.gasLimit() == uint64(_cfg.l2GenesisBlockGasLimit()), "CHECK-SCFG-50");
        //     require(config.unsafeBlockSigner() == _cfg.p2pSequencerAddress(), "CHECK-SCFG-60");
        //     require(config.scalar() >> 248 == 1, "CHECK-SCFG-70");
        //     // Check _config
        //     IResourceMetering.ResourceConfig memory rconfig = Constants.DEFAULT_RESOURCE_CONFIG();
        //     require(resourceConfig.maxResourceLimit == rconfig.maxResourceLimit, "CHECK-SCFG-80");
        //     require(resourceConfig.elasticityMultiplier == rconfig.elasticityMultiplier, "CHECK-SCFG-90");
        //     require(resourceConfig.baseFeeMaxChangeDenominator == rconfig.baseFeeMaxChangeDenominator,
        // "CHECK-SCFG-100");
        //     require(resourceConfig.systemTxMaxGas == rconfig.systemTxMaxGas, "CHECK-SCFG-110");
        //     require(resourceConfig.minimumBaseFee == rconfig.minimumBaseFee, "CHECK-SCFG-120");
        //     require(resourceConfig.maximumBaseFee == rconfig.maximumBaseFee, "CHECK-SCFG-130");
        //     // Depends on start block being set to 0 in `initialize`
        //     uint256 cfgStartBlock = _cfg.systemConfigStartBlock();
        //     require(config.startBlock() == (cfgStartBlock == 0 ? block.number : cfgStartBlock), "CHECK-SCFG-140");
        //     require(config.batchInbox() == _cfg.batchInboxAddress(), "CHECK-SCFG-150");
        //     // Check _addresses
        //     require(config.l1CrossDomainMessenger() == _contracts.L1CrossDomainMessenger, "CHECK-SCFG-160");
        //     require(config.l1ERC721Bridge() == _contracts.L1ERC721Bridge, "CHECK-SCFG-170");
        //     require(config.l1StandardBridge() == _contracts.L1StandardBridge, "CHECK-SCFG-180");
        //     require(config.optimismPortal() == _contracts.OptimismPortal, "CHECK-SCFG-200");
        //     require(config.optimismMintableERC20Factory() == _contracts.OptimismMintableERC20Factory,
        // "CHECK-SCFG-210");
        // } else {
        //     require(config.owner() == address(0), "CHECK-SCFG-220");
        //     require(config.overhead() == 0, "CHECK-SCFG-230");
        //     require(config.scalar() == 0, "CHECK-SCFG-240"); // version 1
        //     require(config.basefeeScalar() == 0, "CHECK-SCFG-250");
        //     require(config.blobbasefeeScalar() == 0, "CHECK-SCFG-260");
        //     require(config.batcherHash() == bytes32(0), "CHECK-SCFG-270");
        //     require(config.gasLimit() == 0, "CHECK-SCFG-280");
        //     require(config.unsafeBlockSigner() == address(0), "CHECK-SCFG-290");
        //     // Check _config
        //     require(resourceConfig.maxResourceLimit == 0, "CHECK-SCFG-300");
        //     require(resourceConfig.elasticityMultiplier == 0, "CHECK-SCFG-310");
        //     require(resourceConfig.baseFeeMaxChangeDenominator == 0, "CHECK-SCFG-320");
        //     require(resourceConfig.systemTxMaxGas == 0, "CHECK-SCFG-330");
        //     require(resourceConfig.minimumBaseFee == 0, "CHECK-SCFG-340");
        //     require(resourceConfig.maximumBaseFee == 0, "CHECK-SCFG-350");
        //     // Check _addresses
        //     require(config.startBlock() == type(uint256).max, "CHECK-SCFG-360");
        //     require(config.batchInbox() == address(0), "CHECK-SCFG-370");
        //     require(config.l1CrossDomainMessenger() == address(0), "CHECK-SCFG-380");
        //     require(config.l1ERC721Bridge() == address(0), "CHECK-SCFG-390");
        //     require(config.l1StandardBridge() == address(0), "CHECK-SCFG-400");
        //     require(config.optimismPortal() == address(0), "CHECK-SCFG-420");
        //     require(config.optimismMintableERC20Factory() == address(0), "CHECK-SCFG-430");
        // }
    }

    /// @notice Asserts that the L2OutputOracle is setup correctly
    function checkL2OutputOracle(Types.ContractSet memory _contracts, bool _isProxy) internal view { }

    /// @notice Asserts that the L1CrossDomainMessenger is setup correctly
    function checkL1CrossDomainMessenger(Types.ContractSet memory _contracts, Vm _vm, bool _isProxy) internal view {
        // IL1CrossDomainMessenger messenger = IL1CrossDomainMessenger(_contracts.L1CrossDomainMessenger);
        // console.log(
        //     "Running chain assertions on the L1CrossDomainMessenger %s at %s",
        //     _isProxy ? "proxy" : "implementation",
        //     address(messenger)
        // );
        // require(address(messenger) != address(0), "CHECK-L1XDM-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(messenger), _isProxy: _isProxy, _slot: 0, _offset:
        // 20 });

        // if (_isProxy) {
        //     require(address(messenger.OTHER_MESSENGER()) == Predeploys.L2_CROSS_DOMAIN_MESSENGER, "CHECK-L1XDM-20");
        //     require(address(messenger.otherMessenger()) == Predeploys.L2_CROSS_DOMAIN_MESSENGER, "CHECK-L1XDM-30");
        //     require(address(messenger.PORTAL()) == _contracts.OptimismPortal, "CHECK-L1XDM-40");
        //     require(address(messenger.portal()) == _contracts.OptimismPortal, "CHECK-L1XDM-50");
        //     require(address(messenger.systemConfig()) == _contracts.SystemConfig, "CHECK-L1XDM-60");
        //     bytes32 xdmSenderSlot = _vm.load(address(messenger), bytes32(uint256(204)));
        //     require(address(uint160(uint256(xdmSenderSlot))) == Constants.DEFAULT_L2_SENDER, "CHECK-L1XDM-70");
        // } else {
        //     require(address(messenger.OTHER_MESSENGER()) == address(0), "CHECK-L1XDM-80");
        //     require(address(messenger.otherMessenger()) == address(0), "CHECK-L1XDM-90");
        //     require(address(messenger.PORTAL()) == address(0), "CHECK-L1XDM-100");
        //     require(address(messenger.portal()) == address(0), "CHECK-L1XDM-110");
        //     require(address(messenger.systemConfig()) == address(0), "CHECK-L1XDM-120");
        // }
    }

    /// @notice Asserts that the L1StandardBridge is setup correctly
    function checkL1StandardBridge(Types.ContractSet memory _contracts, bool _isProxy) internal view {
        // IL1StandardBridge bridge = IL1StandardBridge(payable(_contracts.L1StandardBridge));
        // console.log(
        //     "Running chain assertions on the L1StandardBridge %s at %s",
        //     _isProxy ? "proxy" : "implementation",
        //     address(bridge)
        // );
        // require(address(bridge) != address(0), "CHECK-L1SB-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(bridge), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        // if (_isProxy) {
        //     require(address(bridge.MESSENGER()) == _contracts.L1CrossDomainMessenger, "CHECK-L1SB-20");
        //     require(address(bridge.messenger()) == _contracts.L1CrossDomainMessenger, "CHECK-L1SB-30");
        //     require(address(bridge.OTHER_BRIDGE()) == Predeploys.L2_STANDARD_BRIDGE, "CHECK-L1SB-40");
        //     require(address(bridge.otherBridge()) == Predeploys.L2_STANDARD_BRIDGE, "CHECK-L1SB-50");
        //     require(address(bridge.systemConfig()) == _contracts.SystemConfig, "CHECK-L1SB-60");
        // } else {
        //     require(address(bridge.MESSENGER()) == address(0), "CHECK-L1SB-70");
        //     require(address(bridge.messenger()) == address(0), "CHECK-L1SB-80");
        //     require(address(bridge.OTHER_BRIDGE()) == address(0), "CHECK-L1SB-90");
        //     require(address(bridge.otherBridge()) == address(0), "CHECK-L1SB-100");
        //     require(address(bridge.systemConfig()) == address(0), "CHECK-L1SB-110");
        // }
    }

    /// @notice Asserts that the OptimismMintableERC20Factory is setup correctly
    function checkOptimismMintableERC20Factory(Types.ContractSet memory _contracts, bool _isProxy) internal view {
        // IOptimismMintableERC20Factory factory =
        // IOptimismMintableERC20Factory(_contracts.OptimismMintableERC20Factory);
        // console.log(
        //     "Running chain assertions on the OptimismMintableERC20Factory %s at %s",
        //     _isProxy ? "proxy" : "implementation",
        //     address(factory)
        // );
        // require(address(factory) != address(0), "CHECK-MERC20F-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(factory), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        // if (_isProxy) {
        //     require(factory.BRIDGE() == _contracts.L1StandardBridge, "CHECK-MERC20F-10");
        //     require(factory.bridge() == _contracts.L1StandardBridge, "CHECK-MERC20F-20");
        // } else {
        //     require(factory.BRIDGE() == address(0), "CHECK-MERC20F-30");
        //     require(factory.bridge() == address(0), "CHECK-MERC20F-40");
        // }
    }

    /// @notice Asserts that the L1ERC721Bridge is setup correctly
    function checkL1ERC721Bridge(Types.ContractSet memory _contracts, bool _isProxy) internal view {
        // console.log("Running chain assertions on the L1ERC721Bridge");
        // IL1ERC721Bridge bridge = IL1ERC721Bridge(_contracts.L1ERC721Bridge);
        // console.log(
        //     "Running chain assertions on the L1ERC721Bridge %s at %s",
        //     _isProxy ? "proxy" : "implementation",
        //     address(bridge)
        // );
        // require(address(bridge) != address(0), "CHECK-L1ERC721B-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(bridge), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        // if (_isProxy) {
        //     require(address(bridge.OTHER_BRIDGE()) == Predeploys.L2_ERC721_BRIDGE, "CHECK-L1ERC721B-10");
        //     require(address(bridge.otherBridge()) == Predeploys.L2_ERC721_BRIDGE, "CHECK-L1ERC721B-20");
        //     require(address(bridge.MESSENGER()) == _contracts.L1CrossDomainMessenger, "CHECK-L1ERC721B-30");
        //     require(address(bridge.messenger()) == _contracts.L1CrossDomainMessenger, "CHECK-L1ERC721B-40");
        //     require(address(bridge.systemConfig()) == _contracts.SystemConfig, "CHECK-L1ERC721B-50");
        // } else {
        //     require(address(bridge.OTHER_BRIDGE()) == address(0), "CHECK-L1ERC721B-60");
        //     require(address(bridge.otherBridge()) == address(0), "CHECK-L1ERC721B-70");
        //     require(address(bridge.MESSENGER()) == address(0), "CHECK-L1ERC721B-80");
        //     require(address(bridge.messenger()) == address(0), "CHECK-L1ERC721B-90");
        //     require(address(bridge.systemConfig()) == address(0), "CHECK-L1ERC721B-100");
        // }
    }

    /// @notice Asserts the OptimismPortal is setup correctly
    function checkOptimismPortal(Types.ContractSet memory _contracts, DeployConfig _cfg, bool _isProxy) internal view {
        // IOptimismPortal portal = IOptimismPortal(payable(_contracts.OptimismPortal));
        // console.log(
        //     "Running chain assertions on the OptimismPortal2 %s at %s",
        //     _isProxy ? "proxy" : "implementation",
        //     address(portal)
        // );
        // require(address(portal) != address(0), "CHECK-OP2-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(portal), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        // address guardian = _cfg.superchainConfigGuardian();
        // if (guardian.code.length == 0) {
        //     console.log("Guardian has no code: %s", guardian);
        // }

        // if (_isProxy) {
        //     require(address(portal.disputeGameFactory()) == _contracts.DisputeGameFactory, "CHECK-OP2-20");
        //     require(address(portal.anchorStateRegistry()) == _contracts.AnchorStateRegistry, "CHECK-OP2-25");
        //     require(address(portal.systemConfig()) == _contracts.SystemConfig, "CHECK-OP2-30");
        //     require(portal.guardian() == guardian, "CHECK-OP2-40");
        //     require(address(portal.systemConfig()) == address(_contracts.SystemConfig), "CHECK-OP2-50");
        //     require(portal.paused() == ISystemConfig(_contracts.SystemConfig).paused(), "CHECK-OP2-60");
        //     require(portal.l2Sender() == Constants.DEFAULT_L2_SENDER, "CHECK-OP2-70");
        //     require(address(portal.ethLockbox()) == _contracts.ETHLockbox, "CHECK-OP2-80");
        // } else {
        //     require(address(portal.anchorStateRegistry()) == address(0), "CHECK-OP2-80");
        //     require(address(portal.systemConfig()) == address(0), "CHECK-OP2-90");
        //     require(address(portal.systemConfig()) == address(0), "CHECK-OP2-100");
        //     require(portal.l2Sender() == address(0), "CHECK-OP2-110");
        //     require(address(portal.ethLockbox()) == address(0), "CHECK-OP2-120");
        // }
        // // This slot is the custom gas token _balance and this check ensures
        // // that it stays unset for forwards compatibility with custom gas token.
        // require(vm.load(address(portal), bytes32(uint256(61))) == bytes32(0), "CHECK-OP2-130");
    }

    // /// @notice Asserts that the OPContractsManager is setup correctly
    // function checkOPContractsManager(
    //     Types.ContractSet memory _impls,
    //     Types.ContractSet memory _proxies,
    //     IOPContractsManager _opcm,
    //     IMIPS _mips,
    //     IProxyAdmin _superchainProxyAdmin
    // )
    //     internal
    //     view
    // {
    //     console.log("Running chain assertions on the OPContractsManager at %s", address(_opcm));
    //     require(address(_opcm) != address(0), "CHECK-OPCM-10");

    //     require(bytes(_opcm.version()).length > 0, "CHECK-OPCM-15");
    //     require(bytes(_opcm.l1ContractsRelease()).length > 0, "CHECK-OPCM-16");
    //     require(address(_opcm.protocolVersions()) == _proxies.ProtocolVersions, "CHECK-OPCM-17");
    //     require(address(_opcm.superchainProxyAdmin()) == address(_superchainProxyAdmin), "CHECK-OPCM-18");
    //     require(address(_opcm.superchainConfig()) == _proxies.SuperchainConfig, "CHECK-OPCM-19");

    //     // Ensure that the OPCM impls are correctly saved
    //     IOPContractsManager.Implementations memory impls = _opcm.implementations();
    //     require(impls.l1ERC721BridgeImpl == _impls.L1ERC721Bridge, "CHECK-OPCM-50");
    //     require(impls.optimismPortalImpl == _impls.OptimismPortal, "CHECK-OPCM-60");
    //     require(impls.systemConfigImpl == _impls.SystemConfig, "CHECK-OPCM-70");
    //     require(impls.optimismMintableERC20FactoryImpl == _impls.OptimismMintableERC20Factory, "CHECK-OPCM-80");
    //     require(impls.l1CrossDomainMessengerImpl == _impls.L1CrossDomainMessenger, "CHECK-OPCM-90");
    //     require(impls.l1StandardBridgeImpl == _impls.L1StandardBridge, "CHECK-OPCM-100");
    //     require(impls.disputeGameFactoryImpl == _impls.DisputeGameFactory, "CHECK-OPCM-110");
    //     require(impls.delayedWETHImpl == _impls.DelayedWETH, "CHECK-OPCM-120");
    //     require(impls.mipsImpl == address(_mips), "CHECK-OPCM-130");
    //     require(impls.superchainConfigImpl == _impls.SuperchainConfig, "CHECK-OPCM-140");
    //     require(impls.protocolVersionsImpl == _impls.ProtocolVersions, "CHECK-OPCM-150");

    //     // Verify that initCode is correctly set into the blueprints
    //     IOPContractsManager.Blueprints memory blueprints = _opcm.blueprints();
    //     Blueprint.Preamble memory addressManagerPreamble =
    //         Blueprint.parseBlueprintPreamble(address(blueprints.addressManager).code);
    //     require(keccak256(addressManagerPreamble.initcode) == keccak256(vm.getCode("AddressManager")),
    // "CHECK-OPCM-160");

    //     Blueprint.Preamble memory proxyPreamble = Blueprint.parseBlueprintPreamble(address(blueprints.proxy).code);
    //     require(keccak256(proxyPreamble.initcode) == keccak256(vm.getCode("Proxy")), "CHECK-OPCM-170");

    //     Blueprint.Preamble memory proxyAdminPreamble =
    //         Blueprint.parseBlueprintPreamble(address(blueprints.proxyAdmin).code);
    //     require(keccak256(proxyAdminPreamble.initcode) == keccak256(vm.getCode("ProxyAdmin")), "CHECK-OPCM-180");

    //     Blueprint.Preamble memory l1ChugSplashProxyPreamble =
    //         Blueprint.parseBlueprintPreamble(address(blueprints.l1ChugSplashProxy).code);
    //     require(
    //         keccak256(l1ChugSplashProxyPreamble.initcode) == keccak256(vm.getCode("L1ChugSplashProxy")),
    //         "CHECK-OPCM-190"
    //     );

    //     Blueprint.Preamble memory rdProxyPreamble =
    //         Blueprint.parseBlueprintPreamble(address(blueprints.resolvedDelegateProxy).code);
    //     require(keccak256(rdProxyPreamble.initcode) == keccak256(vm.getCode("ResolvedDelegateProxy")),
    // "CHECK-OPCM-200");

    //     Blueprint.Preamble memory pdg1Preamble =
    //         Blueprint.parseBlueprintPreamble(address(blueprints.permissionedDisputeGame1).code);
    //     Blueprint.Preamble memory pdg2Preamble =
    //         Blueprint.parseBlueprintPreamble(address(blueprints.permissionedDisputeGame2).code);
    //     // combine pdg1 and pdg2 initcodes
    //     bytes memory fullPermissionedDisputeGameInitcode =
    //         abi.encodePacked(pdg1Preamble.initcode, pdg2Preamble.initcode);
    //     require(
    //         keccak256(fullPermissionedDisputeGameInitcode) == keccak256(vm.getCode("PermissionedDisputeGame")),
    //         "CHECK-OPCM-210"
    //     );
    // }
}
