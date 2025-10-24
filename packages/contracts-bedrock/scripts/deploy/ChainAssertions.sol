// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { Vm } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";

// Scripts
import { DeployConfig } from "scripts/deploy/DeployConfig.s.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { DeployImplementations } from "scripts/deploy/DeployImplementations.s.sol";

// Libraries
import { Constants } from "src/libraries/Constants.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Types } from "scripts/libraries/Types.sol";
import { Blueprint } from "src/libraries/Blueprint.sol";

// Interfaces
// import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";
import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";
import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IAddressManager } from "interfaces/legacy/IAddressManager.sol";

// Contracts
import { ResourceMetering } from "src/L1/ResourceMetering.sol";

library ChainAssertions {
    Vm internal constant vm = Vm(0x7109709ECfa91a80626fF3989D68f67F5b1DD12D);

    /// @notice Asserts the correctness of an L1 deployment. This function expects that all contracts
    ///         within the `prox` ContractSet are proxies that have been setup and initialized.
    function postDeployAssertions(Types.ContractSet memory _prox, DeployConfig _cfg) internal view {
        console.log("Running post-deploy assertions");
        IResourceMetering.ResourceConfig memory rcfg = ISystemConfig(_prox.SystemConfig).resourceConfig();
        ResourceMetering.ResourceConfig memory dflt = Constants.DEFAULT_RESOURCE_CONFIG();
        require(keccak256(abi.encode(rcfg)) == keccak256(abi.encode(dflt)), "CHECK-RCFG-10");

        checkProxyAdmin({ _contracts: _prox, _cfg: _cfg });
        checkAddressManager({ _contracts: _prox, _cfg: _cfg });
        checkSystemConfig({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        checkL1CrossDomainMessenger({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        checkL1StandardBridge({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        checkOptimismMintableERC20Factory({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        checkL1ERC721Bridge({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        checkOptimismPortal({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
        checkL2OutputOracle({ _contracts: _prox, _cfg: _cfg, _isProxy: true });
    }

    /// @notice Asserts that the ProxyAdmin is setup correctly
    function checkProxyAdmin(Types.ContractSet memory _contracts, DeployConfig _cfg) internal view {
        IProxyAdmin admin = IProxyAdmin(_contracts.ProxyAdmin);
        console.log("Running chain assertions on the ProxyAdmin %s", address(admin));

        require(address(admin) != address(0), "CHECK-PROXY-ADMIN-10");
        require(address(admin.addressManager()) == _contracts.AddressManager, "CHECK-PROXY-ADMIN-20");
        require(
            keccak256(abi.encodePacked(admin.implementationName(_contracts.L1CrossDomainMessenger)))
                == keccak256(abi.encodePacked("BVM_L1CrossDomainMessenger")),
            "CHECK-PROXY-ADMIN-30"
        );
        require(uint8(admin.proxyType(_contracts.L1CrossDomainMessenger)) == 2, "CHECK-PROXY-ADMIN-40");
        require(uint8(admin.proxyType(_contracts.L1StandardBridge)) == 1, "CHECK-PROXY-ADMIN-50");
    }

    /// @notice Asserts that the AddressManager is setup correctly
    function checkAddressManager(Types.ContractSet memory _contracts, DeployConfig _cfg) internal view {
        IAddressManager manager = IAddressManager(_contracts.AddressManager);
        console.log("Running chain assertions on the AddressManager %s", address(manager));

        require(address(manager) != address(0), "CHECK-ADDR-MGR-10");
    }

    /// @notice Asserts that the SystemConfig is setup correctly
    function checkSystemConfig(Types.ContractSet memory _contracts, DeployConfig _cfg, bool _isProxy) internal view {
        ISystemConfig config = ISystemConfig(_isProxy ? _contracts.SystemConfig : _contracts.SystemConfigImpl);
        console.log(
            "Running chain assertions on the SystemConfig %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(config)
        );
        require(address(config) != address(0), "CHECK-SCFG-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(config), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        IResourceMetering.ResourceConfig memory resourceConfig = config.resourceConfig();

        if (_isProxy) {
            require(config.owner() == _cfg.finalSystemOwner(), "CHECK-SCFG-10");
            require(config.overhead() == 0, "CHECK-SCFG-20");
            require(config.scalar() == 0, "CHECK-SCFG-30");
            require(config.batcherHash() == bytes32(uint256(uint160(_cfg.batchSenderAddress()))), "CHECK-SCFG-40");
            require(config.gasLimit() == uint64(_cfg.l2GenesisBlockGasLimit()), "CHECK-SCFG-50");
            require(config.baseFee() == _cfg.l2GenesisBlockBaseFeePerGas(), "CHECK-SCFG-60");
            require(config.unsafeBlockSigner() == _cfg.p2pSequencerAddress(), "CHECK-SCFG-70");
            // Check _config
            require(resourceConfig.maxResourceLimit == 20_000_000, "CHECK-SCFG-80");
            require(resourceConfig.elasticityMultiplier == 10, "CHECK-SCFG-90");
            require(resourceConfig.baseFeeMaxChangeDenominator == 8, "CHECK-SCFG-100");
            require(resourceConfig.systemTxMaxGas == 1_000_000, "CHECK-SCFG-110");
            require(resourceConfig.minimumBaseFee == 1 gwei, "CHECK-SCFG-120");
            require(resourceConfig.maximumBaseFee == type(uint128).max, "CHECK-SCFG-130");
        } else {
            require(config.owner() == _cfg.finalSystemOwner(), "CHECK-SCFG-210");
            require(config.overhead() == 0, "CHECK-SCFG-220");
            require(config.scalar() == 0, "CHECK-SCFG-230");
            require(config.batcherHash() == bytes32(uint256(uint160(_cfg.batchSenderAddress()))), "CHECK-SCFG-240");
            require(config.gasLimit() == uint64(_cfg.l2GenesisBlockGasLimit()), "CHECK-SCFG-250");
            require(config.baseFee() == _cfg.l2GenesisBlockBaseFeePerGas(), "CHECK-SCFG-260");
            require(config.unsafeBlockSigner() == _cfg.p2pSequencerAddress(), "CHECK-SCFG-270");
            // Check _config
            require(resourceConfig.maxResourceLimit == 20_000_000, "CHECK-SCFG-280");
            require(resourceConfig.elasticityMultiplier == 10, "CHECK-SCFG-290");
            require(resourceConfig.baseFeeMaxChangeDenominator == 8, "CHECK-SCFG-300");
            require(resourceConfig.systemTxMaxGas == 1_000_000, "CHECK-SCFG-310");
            require(resourceConfig.minimumBaseFee == 1 gwei, "CHECK-SCFG-320");
            require(resourceConfig.maximumBaseFee == type(uint128).max, "CHECK-SCFG-330");
        }
    }

    /// @notice Asserts that the L2OutputOracle is setup correctly
    function checkL2OutputOracle(Types.ContractSet memory _contracts, DeployConfig _cfg, bool _isProxy) internal view {
        IL2OutputOracle oracle = IL2OutputOracle(_isProxy ? _contracts.L2OutputOracle : _contracts.L2OutputOracleImpl);
        console.log(
            "Running chain assertions on the L2OutputOracle %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(oracle)
        );
        require(address(oracle) != address(0), "CHECK-L2OO-10");

        if (_isProxy) {
            require(oracle.SUBMISSION_INTERVAL() == _cfg.l2OutputOracleSubmissionInterval(), "CHECK-L2OO-20");
            require(oracle.L2_BLOCK_TIME() == _cfg.l2BlockTime(), "CHECK-L2OO-30");
            require(oracle.PROPOSER() == _cfg.l2OutputOracleProposer(), "CHECK-L2OO-40");
            require(oracle.CHALLENGER() == _cfg.l2OutputOracleChallenger(), "CHECK-L2OO-50");
            require(oracle.FINALIZATION_PERIOD_SECONDS() == _cfg.finalizationPeriodSeconds(), "CHECK-L2OO-60");
            require(oracle.startingBlockNumber() == _cfg.l2OutputOracleStartingBlockNumber(), "CHECK-L2OO-70");
        } else {
            require(oracle.SUBMISSION_INTERVAL() == _cfg.l2OutputOracleSubmissionInterval(), "CHECK-L2OO-90");
            require(oracle.L2_BLOCK_TIME() == _cfg.l2BlockTime(), "CHECK-L2OO-100");
            require(oracle.PROPOSER() == _cfg.l2OutputOracleProposer(), "CHECK-L2OO-110");
            require(oracle.CHALLENGER() == _cfg.l2OutputOracleChallenger(), "CHECK-L2OO-120");
            require(oracle.FINALIZATION_PERIOD_SECONDS() == _cfg.finalizationPeriodSeconds(), "CHECK-L2OO-130");
        }
    }

    /// @notice Asserts that the L1CrossDomainMessenger is setup correctly
    function checkL1CrossDomainMessenger(
        Types.ContractSet memory _contracts,
        DeployConfig _cfg,
        bool _isProxy
    )
        internal
        view
    {
        IL1CrossDomainMessenger messenger = IL1CrossDomainMessenger(
            _isProxy ? _contracts.L1CrossDomainMessenger : _contracts.L1CrossDomainMessengerImpl
        );
        console.log(
            "Running chain assertions on the L1CrossDomainMessenger %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(messenger)
        );
        require(address(messenger) != address(0), "CHECK-L1XDM-10");

        if (_isProxy) {
            require(address(messenger.PORTAL()) == _contracts.OptimismPortal, "CHECK-L1XDM-20");
        } else {
            require(address(messenger.PORTAL()) == _contracts.OptimismPortal, "CHECK-L1XDM-30");
            require(messenger.L1_MNT_ADDRESS() == _cfg.l1MantleToken(), "CHECK-L1XDM-40");
        }
    }

    /// @notice Asserts that the L1StandardBridge is setup correctly
    function checkL1StandardBridge(
        Types.ContractSet memory _contracts,
        DeployConfig _cfg,
        bool _isProxy
    )
        internal
        view
    {
        IL1StandardBridge bridge =
            IL1StandardBridge(payable(_isProxy ? _contracts.L1StandardBridge : _contracts.L1StandardBridgeImpl));
        console.log(
            "Running chain assertions on the L1StandardBridge %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(bridge)
        );
        require(address(bridge) != address(0), "CHECK-L1SB-10");

        if (_isProxy) {
            require(address(bridge.MESSENGER()) == _contracts.L1CrossDomainMessenger, "CHECK-L1SB-20");
            require(address(bridge.OTHER_BRIDGE()) == Predeploys.L2_STANDARD_BRIDGE, "CHECK-L1SB-30");
        } else {
            require(address(bridge.MESSENGER()) == _contracts.L1CrossDomainMessenger, "CHECK-L1SB-40");
            require(address(bridge.OTHER_BRIDGE()) == Predeploys.L2_STANDARD_BRIDGE, "CHECK-L1SB-50");
            require(address(bridge.L1_MNT_ADDRESS()) == _cfg.l1MantleToken(), "CHECK-L1SB-60");
        }
    }

    /// @notice Asserts that the OptimismMintableERC20Factory is setup correctly
    function checkOptimismMintableERC20Factory(
        Types.ContractSet memory _contracts,
        DeployConfig _cfg,
        bool _isProxy
    )
        internal
        view
    {
        IOptimismMintableERC20Factory factory = IOptimismMintableERC20Factory(
            _isProxy ? _contracts.OptimismMintableERC20Factory : _contracts.OptimismMintableERC20FactoryImpl
        );
        console.log(
            "Running chain assertions on the OptimismMintableERC20Factory %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(factory)
        );
        require(address(factory) != address(0), "CHECK-MERC20F-10");

        if (_isProxy) {
            require(factory.BRIDGE() == _contracts.L1StandardBridge, "CHECK-MERC20F-20");
        } else {
            require(factory.BRIDGE() == _contracts.L1StandardBridge, "CHECK-MERC20F-30");
        }
    }

    /// @notice Asserts that the L1ERC721Bridge is setup correctly
    function checkL1ERC721Bridge(Types.ContractSet memory _contracts, DeployConfig _cfg, bool _isProxy) internal view {
        IL1ERC721Bridge bridge = IL1ERC721Bridge(_isProxy ? _contracts.L1ERC721Bridge : _contracts.L1ERC721BridgeImpl);
        console.log(
            "Running chain assertions on the L1ERC721Bridge %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(bridge)
        );
        require(address(bridge) != address(0), "CHECK-L1ERC721B-10");

        // // Check that the contract is initialized
        // DeployUtils.assertInitialized({ _contractAddress: address(bridge), _isProxy: _isProxy, _slot: 0, _offset: 0
        // });

        if (_isProxy) {
            require(address(bridge.MESSENGER()) == _contracts.L1CrossDomainMessenger, "CHECK-L1ERC721B-20");
            require(bridge.OTHER_BRIDGE() == Predeploys.L2_ERC721_BRIDGE, "CHECK-L1ERC721B-30");
        } else {
            require(address(bridge.OTHER_BRIDGE()) == Predeploys.L2_ERC721_BRIDGE, "CHECK-L1ERC721B-40");
            require(address(bridge.MESSENGER()) == _contracts.L1CrossDomainMessenger, "CHECK-L1ERC721B-50");
        }
    }

    /// @notice Asserts the OptimismPortal is setup correctly
    function checkOptimismPortal(Types.ContractSet memory _contracts, DeployConfig _cfg, bool _isProxy) internal view {
        IOptimismPortal portal =
            IOptimismPortal(payable(_isProxy ? _contracts.OptimismPortal : _contracts.OptimismPortalImpl));
        console.log(
            "Running chain assertions on the OptimismPortal %s at %s",
            _isProxy ? "proxy" : "implementation",
            address(portal)
        );
        require(address(portal) != address(0), "CHECK-OP2-10");

        address guardian = _cfg.portalGuardian();
        if (guardian.code.length == 0) {
            console.log("Guardian has no code: %s", guardian);
        }

        if (_isProxy) {
            require(address(portal.L2_ORACLE()) == _contracts.L2OutputOracle, "CHECK-OP2-20");
            require(portal.GUARDIAN() == _cfg.portalGuardian(), "CHECK-OP2-30");
            require(address(portal.SYSTEM_CONFIG()) == _contracts.SystemConfig, "CHECK-OP2-40");
            require(portal.paused() == false, "CHECK-OP2-50");
            (uint128 prevBaseFee, uint64 prevBoughtGas, uint64 prevBlockNum) =
                IResourceMetering(address(portal)).params();
            require(prevBaseFee == 1 gwei, "CHECK-OP2-60");
            require(prevBoughtGas == 0, "CHECK-OP2-70");
            require(prevBlockNum != 0, "CHECK-OP2-80");
        } else {
            require(address(portal.L2_ORACLE()) == _contracts.L2OutputOracle, "CHECK-OP2-90");
            require(address(portal.GUARDIAN()) == _cfg.portalGuardian(), "CHECK-OP2-100");
            require(address(portal.SYSTEM_CONFIG()) == _contracts.SystemConfig, "CHECK-OP2-110");
            require(address(portal.L1_MNT_ADDRESS()) == _cfg.l1MantleToken(), "CHECK-OP2-120");
        }
    }
}
