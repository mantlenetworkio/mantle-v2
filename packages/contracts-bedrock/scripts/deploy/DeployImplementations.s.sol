// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";

// Libraries
import { LibString } from "@solady/utils/LibString.sol";

// Interfaces
import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
// import {
//     IOPContractsManager,
//     IOPContractsManagerGameTypeAdder,
//     IOPContractsManagerDeployer,
//     IOPContractsManagerUpgrader,
//     IOPContractsManagerContractsContainer,
//     IOPContractsManagerInteropMigrator
// } from "interfaces/L1/IOPContractsManager.sol";
import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";
import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";

// import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { Solarray } from "scripts/libraries/Solarray.sol";

import { Predeploys } from "src/libraries/Predeploys.sol";

contract DeployImplementations is Script {
    struct Input {
        address systemConfig_owner;
        uint256 systemConfig_overhead;
        uint256 systemConfig_scalar;
        bytes32 systemConfig_batcherHash;
        uint64 systemConfig_gasLimit;
        uint256 systemConfig_baseFee;
        address systemConfig_unsafeBlockSigner;
        IResourceMetering.ResourceConfig systemConfig_config;
        IOptimismPortal optimismPortal;
        address l1mnt;
        IL1CrossDomainMessenger l1CrossDomainMessenger;
        IL2OutputOracle l2OutputOracle;
        ISystemConfig systemConfig;
        IL1StandardBridge l1StandardBridge;
        address l1ERC721Bridge_otherBridge;
        uint256 l2OutputOracle_submissionInterval;
        uint256 l2OutputOracle_l2BlockTime;
        uint256 l2OutputOracle_startingBlockNumber;
        uint256 l2OutputOracle_startingTimestamp;
        address l2OutputOracle_proposer;
        address l2OutputOracle_challenger;
        uint256 l2OutputOracle_finalizationPeriodSeconds;
        address optimismPortal_guardian;
        bool optimismPortal_paused;
        // This is used in opcm to signal which version of the L1 smart contracts is deployed.
        // It takes the format of `op-contracts/v*.*.*`.
        string l1ContractsRelease;
        address upgradeController;
    }

    struct Output {
        // IOPContractsManager opcm;
        // IOPContractsManagerContractsContainer opcmContractsContainer;
        // IOPContractsManagerGameTypeAdder opcmGameTypeAdder;
        // IOPContractsManagerDeployer opcmDeployer;
        // IOPContractsManagerUpgrader opcmUpgrader;
        // IOPContractsManagerInteropMigrator opcmInteropMigrator;
        IOptimismPortal optimismPortalImpl;
        ISystemConfig systemConfigImpl;
        IL1CrossDomainMessenger l1CrossDomainMessengerImpl;
        IL1ERC721Bridge l1ERC721BridgeImpl;
        IL1StandardBridge l1StandardBridgeImpl;
        IOptimismMintableERC20Factory optimismMintableERC20FactoryImpl;
        IL2OutputOracle l2OutputOracleImpl;
    }

    bytes32 internal _salt = DeployUtils.DEFAULT_SALT;

    // -------- Core Deployment Methods --------

    function run(Input memory _input) public returns (Output memory output_) {
        assertValidInput(_input);

        // Deploy the implementations.
        deploySystemConfigImpl(_input, output_);
        deployL1CrossDomainMessengerImpl(_input, output_);
        deployL1ERC721BridgeImpl(_input, output_);
        deployL1StandardBridgeImpl(_input, output_);
        deployOptimismMintableERC20FactoryImpl(_input, output_);
        deployOptimismPortalImpl(_input, output_);
        deployL2OutputOracleImpl(_input, output_);

        // Deploy the OP Contracts Manager with the new implementations set.
        // deployOPContractsManager(_input, output_);

        assertValidOutput(_input, output_);
    }

    // -------- Deployment Steps --------

    // --- OP Contracts Manager ---

    // function createOPCMContract(
    //     Input memory _input,
    //     Output memory _output,
    //     IOPContractsManager.Blueprints memory _blueprints,
    //     string memory _l1ContractsRelease
    // )
    //     private
    //     returns (IOPContractsManager opcm_)
    // {
    //     IOPContractsManager.Implementations memory implementations = IOPContractsManager.Implementations({
    //         superchainConfigImpl: address(_output.superchainConfigImpl),
    //         protocolVersionsImpl: address(_output.protocolVersionsImpl),
    //         l1ERC721BridgeImpl: address(_output.l1ERC721BridgeImpl),
    //         optimismPortalImpl: address(_output.optimismPortalImpl),
    //         ethLockboxImpl: address(_output.ethLockboxImpl),
    //         systemConfigImpl: address(_output.systemConfigImpl),
    //         optimismMintableERC20FactoryImpl: address(_output.optimismMintableERC20FactoryImpl),
    //         l1CrossDomainMessengerImpl: address(_output.l1CrossDomainMessengerImpl),
    //         l1StandardBridgeImpl: address(_output.l1StandardBridgeImpl),
    //         disputeGameFactoryImpl: address(_output.disputeGameFactoryImpl),
    //         anchorStateRegistryImpl: address(_output.anchorStateRegistryImpl),
    //         delayedWETHImpl: address(_output.delayedWETHImpl),
    //         mipsImpl: address(_output.mipsSingleton)
    //     });

    //     deployOPCMBPImplsContainer(_output, _blueprints, implementations);
    //     deployOPCMGameTypeAdder(_output);
    //     deployOPCMDeployer(_input, _output);
    //     deployOPCMUpgrader(_output);
    //     deployOPCMInteropMigrator(_output);

    //     // Semgrep rule will fail because the arguments are encoded inside of a separate function.
    //     opcm_ = IOPContractsManager(
    //         // nosemgrep: sol-safety-deployutils-args
    //         DeployUtils.createDeterministic({
    //             _name: "OPContractsManager",
    //             _args: encodeOPCMConstructor(_l1ContractsRelease, _input, _output),
    //             _salt: _salt
    //         })
    //     );

    //     vm.label(address(opcm_), "OPContractsManager");
    //     _output.opcm = opcm_;
    // }

    // /// @notice Encodes the constructor of the OPContractsManager contract. Used to avoid stack too
    // ///         deep errors inside of the createOPCMContract function.
    // /// @param _l1ContractsRelease The release of the L1 contracts.
    // /// @param _input The deployment input parameters.
    // /// @param _output The deployment output parameters.
    // /// @return encoded_ The encoded constructor.
    // function encodeOPCMConstructor(
    //     string memory _l1ContractsRelease,
    //     Input memory _input,
    //     Output memory _output
    // )
    //     private
    //     pure
    //     returns (bytes memory encoded_)
    // {
    //     encoded_ = DeployUtils.encodeConstructor(
    //         abi.encodeCall(
    //             IOPContractsManager.__constructor__,
    //             (
    //                 _output.opcmGameTypeAdder,
    //                 _output.opcmDeployer,
    //                 _output.opcmUpgrader,
    //                 _output.opcmInteropMigrator,
    //                 _input.superchainConfigProxy,
    //                 _input.protocolVersionsProxy,
    //                 _input.superchainProxyAdmin,
    //                 _l1ContractsRelease,
    //                 _input.upgradeController
    //             )
    //         )
    //     );
    // }

    // function deployOPContractsManager(Input memory _input, Output memory _output) private {
    //     string memory l1ContractsRelease = _input.l1ContractsRelease;

    //     // First we deploy the blueprints for the singletons deployed by OPCM.
    //     // forgefmt: disable-start
    //     IOPContractsManager.Blueprints memory blueprints;
    //     vm.startBroadcast(msg.sender);
    //     address checkAddress;
    //     (blueprints.addressManager, checkAddress) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("AddressManager"), _salt);
    //     require(checkAddress == address(0), "OPCM-10");
    //     (blueprints.proxy, checkAddress) = DeployUtils.createDeterministicBlueprint(vm.getCode("Proxy"), _salt);
    //     require(checkAddress == address(0), "OPCM-20");
    //     (blueprints.proxyAdmin, checkAddress) = DeployUtils.createDeterministicBlueprint(vm.getCode("ProxyAdmin"),
    // _salt);
    //     require(checkAddress == address(0), "OPCM-30");
    //     (blueprints.l1ChugSplashProxy, checkAddress) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("L1ChugSplashProxy"), _salt);
    //     require(checkAddress == address(0), "OPCM-40");
    //     (blueprints.resolvedDelegateProxy, checkAddress) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("ResolvedDelegateProxy"), _salt);
    //     require(checkAddress == address(0), "OPCM-50");
    //     // The max initcode/runtimecode size is 48KB/24KB.
    //     // But for Blueprint, the initcode is stored as runtime code, that's why it's necessary to split into 2
    // parts.
    //     (blueprints.permissionedDisputeGame1, blueprints.permissionedDisputeGame2) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("PermissionedDisputeGame"), _salt);
    //     (blueprints.permissionlessDisputeGame1, blueprints.permissionlessDisputeGame2) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("FaultDisputeGame"), _salt);
    //     (blueprints.superPermissionedDisputeGame1, blueprints.superPermissionedDisputeGame2) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("SuperPermissionedDisputeGame"), _salt);
    //     (blueprints.superPermissionlessDisputeGame1, blueprints.superPermissionlessDisputeGame2) =
    // DeployUtils.createDeterministicBlueprint(vm.getCode("SuperFaultDisputeGame"), _salt);
    //     // forgefmt: disable-end
    //     vm.stopBroadcast();

    //     IOPContractsManager opcm = createOPCMContract(_input, _output, blueprints, l1ContractsRelease);

    //     vm.label(address(opcm), "OPContractsManager");
    //     _output.opcm = opcm;
    // }

    // --- Core Contracts ---

    function deploySystemConfigImpl(Input memory _input, Output memory _output) private {
        ISystemConfig impl = ISystemConfig(
            DeployUtils.create1({
                _name: "SystemConfig",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        ISystemConfig.__constructor__,
                        (
                            _input.systemConfig_owner,
                            _input.systemConfig_overhead,
                            _input.systemConfig_scalar,
                            _input.systemConfig_batcherHash,
                            _input.systemConfig_gasLimit,
                            _input.systemConfig_baseFee,
                            _input.systemConfig_unsafeBlockSigner,
                            _input.systemConfig_config
                        )
                    )
                )
            })
        );
        vm.label(address(impl), "SystemConfigImpl");
        _output.systemConfigImpl = impl;
    }

    function deployL1CrossDomainMessengerImpl(Input memory _input, Output memory _output) private {
        IL1CrossDomainMessenger impl = IL1CrossDomainMessenger(
            DeployUtils.create1({
                _name: "L1CrossDomainMessenger",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(IL1CrossDomainMessenger.__constructor__, (_input.optimismPortal, _input.l1mnt))
                )
            })
        );
        vm.label(address(impl), "L1CrossDomainMessengerImpl");
        _output.l1CrossDomainMessengerImpl = impl;
    }

    function deployL1ERC721BridgeImpl(Input memory _input, Output memory _output) private {
        IL1ERC721Bridge impl = IL1ERC721Bridge(
            DeployUtils.create1({
                _name: "L1ERC721Bridge",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        IL1ERC721Bridge.__constructor__,
                        (address(_input.l1CrossDomainMessenger), _input.l1ERC721Bridge_otherBridge)
                    )
                )
            })
        );
        vm.label(address(impl), "L1ERC721BridgeImpl");
        _output.l1ERC721BridgeImpl = impl;
    }

    function deployL1StandardBridgeImpl(Input memory _input, Output memory _output) private {
        IL1StandardBridge impl = IL1StandardBridge(
            DeployUtils.create1({
                _name: "L1StandardBridge",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        IL1StandardBridge.__constructor__, (payable(address(_input.l1CrossDomainMessenger)), _input.l1mnt)
                    )
                )
            })
        );
        vm.label(address(impl), "L1StandardBridgeImpl");
        _output.l1StandardBridgeImpl = impl;
    }

    function deployOptimismMintableERC20FactoryImpl(Input memory _input, Output memory _output) private {
        IOptimismMintableERC20Factory impl = IOptimismMintableERC20Factory(
            DeployUtils.create1({
                _name: "OptimismMintableERC20Factory",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(IOptimismMintableERC20Factory.__constructor__, (address(_input.l1StandardBridge)))
                )
            })
        );
        vm.label(address(impl), "OptimismMintableERC20FactoryImpl");
        _output.optimismMintableERC20FactoryImpl = impl;
    }

    function deployOptimismPortalImpl(Input memory _input, Output memory _output) private {
        IOptimismPortal impl = IOptimismPortal(
            DeployUtils.create1({
                _name: "OptimismPortal2",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        IOptimismPortal.__constructor__,
                        (
                            _input.l2OutputOracle,
                            _input.optimismPortal_guardian,
                            _input.optimismPortal_paused,
                            _input.systemConfig,
                            _input.l1mnt
                        )
                    )
                )
            })
        );
        vm.label(address(impl), "OptimismPortalImpl");
        _output.optimismPortalImpl = impl;
    }

    function deployL2OutputOracleImpl(Input memory _input, Output memory _output) private {
        IL2OutputOracle impl = IL2OutputOracle(
            DeployUtils.create1({
                _name: "L2OutputOracle",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        IL2OutputOracle.__constructor__,
                        (
                            _input.l2OutputOracle_submissionInterval,
                            _input.l2OutputOracle_l2BlockTime,
                            _input.l2OutputOracle_startingBlockNumber,
                            _input.l2OutputOracle_startingTimestamp,
                            _input.l2OutputOracle_proposer,
                            _input.l2OutputOracle_challenger,
                            _input.l2OutputOracle_finalizationPeriodSeconds
                        )
                    )
                )
            })
        );
        vm.label(address(impl), "L2OutputOracleImpl");
        _output.l2OutputOracleImpl = impl;
    }

    // function deployOPCMBPImplsContainer(
    //     Output memory _output,
    //     IOPContractsManager.Blueprints memory _blueprints,
    //     IOPContractsManager.Implementations memory _implementations
    // )
    //     private
    // {
    //     IOPContractsManagerContractsContainer impl = IOPContractsManagerContractsContainer(
    //         DeployUtils.createDeterministic({
    //             _name: "OPContractsManager.sol:OPContractsManagerContractsContainer",
    //             _args: DeployUtils.encodeConstructor(
    //                 abi.encodeCall(IOPContractsManagerContractsContainer.__constructor__, (_blueprints,
    // _implementations))
    //             ),
    //             _salt: _salt
    //         })
    //     );
    //     vm.label(address(impl), "OPContractsManagerBPImplsContainerImpl");
    //     _output.opcmContractsContainer = impl;
    // }

    // function deployOPCMGameTypeAdder(Output memory _output) private {
    //     IOPContractsManagerGameTypeAdder impl = IOPContractsManagerGameTypeAdder(
    //         DeployUtils.createDeterministic({
    //             _name: "OPContractsManager.sol:OPContractsManagerGameTypeAdder",
    //             _args: DeployUtils.encodeConstructor(
    //                 abi.encodeCall(IOPContractsManagerGameTypeAdder.__constructor__,
    // (_output.opcmContractsContainer))
    //             ),
    //             _salt: _salt
    //         })
    //     );
    //     vm.label(address(impl), "OPContractsManagerGameTypeAdderImpl");
    //     _output.opcmGameTypeAdder = impl;
    // }

    // function deployOPCMDeployer(Input memory, Output memory _output) private {
    //     IOPContractsManagerDeployer impl = IOPContractsManagerDeployer(
    //         DeployUtils.createDeterministic({
    //             _name: "OPContractsManager.sol:OPContractsManagerDeployer",
    //             _args: DeployUtils.encodeConstructor(
    //                 abi.encodeCall(IOPContractsManagerDeployer.__constructor__, (_output.opcmContractsContainer))
    //             ),
    //             _salt: _salt
    //         })
    //     );
    //     vm.label(address(impl), "OPContractsManagerDeployerImpl");
    //     _output.opcmDeployer = impl;
    // }

    // function deployOPCMUpgrader(Output memory _output) private {
    //     IOPContractsManagerUpgrader impl = IOPContractsManagerUpgrader(
    //         DeployUtils.createDeterministic({
    //             _name: "OPContractsManager.sol:OPContractsManagerUpgrader",
    //             _args: DeployUtils.encodeConstructor(
    //                 abi.encodeCall(IOPContractsManagerUpgrader.__constructor__, (_output.opcmContractsContainer))
    //             ),
    //             _salt: _salt
    //         })
    //     );
    //     vm.label(address(impl), "OPContractsManagerUpgraderImpl");
    //     _output.opcmUpgrader = impl;
    // }

    // function deployOPCMInteropMigrator(Output memory _output) private {
    //     IOPContractsManagerInteropMigrator impl = IOPContractsManagerInteropMigrator(
    //         DeployUtils.createDeterministic({
    //             _name: "OPContractsManager.sol:OPContractsManagerInteropMigrator",
    //             _args: DeployUtils.encodeConstructor(
    //                 abi.encodeCall(IOPContractsManagerInteropMigrator.__constructor__,
    // (_output.opcmContractsContainer))
    //             ),
    //             _salt: _salt
    //         })
    //     );
    //     vm.label(address(impl), "OPContractsManagerInteropMigratorImpl");
    //     _output.opcmInteropMigrator = impl;
    // }

    function assertValidInput(Input memory _input) private pure {
        // todo:
        // if deployer != controller, check portal guardian
        require(address(_input.l1mnt) != address(0), "DeployImplementations: l1mnt not set");

        require(_input.l2OutputOracle_l2BlockTime > 0, "DeployImplementations: l2BlockTime not set");
        require(
            _input.l2OutputOracle_submissionInterval > _input.l2OutputOracle_l2BlockTime,
            "DeployImplementations: submissionInterval must be greater than the l2BlockTime"
        );
        // require(!LibString.eq(_input.l1ContractsRelease, ""), "DeployImplementations: l1ContractsRelease not set");
        // require(address(_input.upgradeController) != address(0), "DeployImplementations: upgradeController not set");
    }

    function assertValidOutput(Input memory _input, Output memory _output) private view {
        address[] memory addrs = Solarray.addresses(
            // address(_output.opcm),
            address(_output.systemConfigImpl),
            address(_output.l1CrossDomainMessengerImpl),
            address(_output.l1ERC721BridgeImpl),
            address(_output.l1StandardBridgeImpl),
            address(_output.optimismMintableERC20FactoryImpl),
            address(_output.optimismPortalImpl),
            address(_output.l2OutputOracleImpl)
        );

        DeployUtils.assertValidContractAddresses(addrs);

        assertValidL1CrossDomainMessengerImpl(_input, _output);
        assertValidL1ERC721BridgeImpl(_input, _output);
        assertValidL1StandardBridgeImpl(_input, _output);
        // assertValidOpcm(_input, _output);
        assertValidOptimismMintableERC20FactoryImpl(_input, _output);
        assertValidOptimismPortalImpl(_input, _output);
        assertValidSystemConfigImpl(_input, _output);
        assertValidL2OutputOracleImpl(_input, _output);
    }

    // function assertValidOpcm(Input memory _input, Output memory _output) private view {
    //     IOPContractsManager impl = IOPContractsManager(address(_output.opcm));
    //     require(address(impl.superchainConfig()) == address(_input.superchainConfigProxy), "OPCMI-10");
    //     require(address(impl.protocolVersions()) == address(_input.protocolVersionsProxy), "OPCMI-20");
    //     require(impl.upgradeController() == _input.upgradeController, "OPCMI-30");
    // }

    function assertValidL2OutputOracleImpl(Input memory _input, Output memory _output) private view {
        IL2OutputOracle oracle = _output.l2OutputOracleImpl;

        require(oracle.SUBMISSION_INTERVAL() == _input.l2OutputOracle_submissionInterval, "L2OO-10");
        require(oracle.L2_BLOCK_TIME() == _input.l2OutputOracle_l2BlockTime, "L2OO-20");
        require(oracle.PROPOSER() == _input.l2OutputOracle_proposer, "L2OO-30");
        require(oracle.CHALLENGER() == _input.l2OutputOracle_challenger, "L2OO-40");
        require(oracle.FINALIZATION_PERIOD_SECONDS() == _input.l2OutputOracle_finalizationPeriodSeconds, "L2OO-50");
    }

    function assertValidOptimismPortalImpl(Input memory _input, Output memory _output) private view {
        IOptimismPortal portal = _output.optimismPortalImpl;

        require(address(portal.L2_ORACLE()) == address(_input.l2OutputOracle), "PORTAL-10");
        require(address(portal.GUARDIAN()) == _input.optimismPortal_guardian, "PORTAL-20");
        require(address(portal.SYSTEM_CONFIG()) == address(_input.systemConfig), "PORTAL-30");
        require(address(portal.L1_MNT_ADDRESS()) == address(_input.l1mnt), "PORTAL-40");
    }

    function assertValidSystemConfigImpl(Input memory _input, Output memory _output) private view {
        ISystemConfig systemConfig = _output.systemConfigImpl;

        require(systemConfig.owner() == _input.systemConfig_owner, "SYSCON-10");
        require(systemConfig.overhead() == _input.systemConfig_overhead, "SYSCON-20");
        require(systemConfig.scalar() == _input.systemConfig_scalar, "SYSCON-30");
        require(systemConfig.batcherHash() == _input.systemConfig_batcherHash, "SYSCON-40");
        require(systemConfig.unsafeBlockSigner() == _input.systemConfig_unsafeBlockSigner, "SYSCON-50");

        IResourceMetering.ResourceConfig memory resourceConfig = systemConfig.resourceConfig();
        require(resourceConfig.maxResourceLimit == 20_000_000, "SYSCON-60");
        require(resourceConfig.elasticityMultiplier == 10, "SYSCON-70");
        require(resourceConfig.baseFeeMaxChangeDenominator == 8, "SYSCON-80");
        require(resourceConfig.systemTxMaxGas == 1_000_000, "SYSCON-90");
        require(resourceConfig.minimumBaseFee == 1 gwei, "SYSCON-100");
        require(resourceConfig.maximumBaseFee == type(uint128).max, "SYSCON-110");
    }

    function assertValidL1CrossDomainMessengerImpl(Input memory _input, Output memory _output) private view {
        IL1CrossDomainMessenger messenger = _output.l1CrossDomainMessengerImpl;

        require(address(messenger.PORTAL()) == address(_input.optimismPortal), "L1xDM-10");
        require(address(messenger.L1_MNT_ADDRESS()) == address(_input.l1mnt), "L1xDM-20");
    }

    function assertValidL1ERC721BridgeImpl(Input memory _input, Output memory _output) private view {
        IL1ERC721Bridge bridge = _output.l1ERC721BridgeImpl;

        require(address(bridge.MESSENGER()) == address(_input.l1CrossDomainMessenger), "L721B-10");
        require(address(bridge.OTHER_BRIDGE()) == address(Predeploys.L2_ERC721_BRIDGE), "L721B-20");
    }

    function assertValidL1StandardBridgeImpl(Input memory _input, Output memory _output) private view {
        IL1StandardBridge bridge = _output.l1StandardBridgeImpl;

        require(address(bridge.MESSENGER()) == address(_input.l1CrossDomainMessenger), "L1SB-10");
        require(address(bridge.OTHER_BRIDGE()) == address(Predeploys.L2_STANDARD_BRIDGE), "L1SB-20");
        require(address(bridge.L1_MNT_ADDRESS()) == address(_input.l1mnt), "L1SB-30");
    }

    function assertValidOptimismMintableERC20FactoryImpl(Input memory _input, Output memory _output) private view {
        IOptimismMintableERC20Factory factory = _output.optimismMintableERC20FactoryImpl;

        require(address(factory.BRIDGE()) == address(_input.l1StandardBridge), "MERC20F-10");
    }
}
