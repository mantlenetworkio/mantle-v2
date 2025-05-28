// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { VmSafe } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";
// import { EIP1967Helper } from "test/mocks/EIP1967Helper.sol";

// Scripts
import { Deployer } from "scripts/deploy/Deployer.sol";
import { Chains } from "scripts/libraries/Chains.sol";
import { Config } from "scripts/libraries/Config.sol";
// import { StateDiff } from "scripts/libraries/StateDiff.sol";
import { ChainAssertions } from "scripts/deploy/ChainAssertions.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { DeployImplementations } from "scripts/deploy/DeployImplementations.s.sol";
import { DeployProxies } from "scripts/deploy/DeployProxies.s.sol";
// import { DeployAltDA } from "scripts/deploy/DeployAltDA.s.sol";
// import { StandardConstants } from "scripts/deploy/StandardConstants.sol";

// // Libraries
import { Types } from "scripts/libraries/Types.sol";
import { Constants } from "src/libraries/Constants.sol";
// import { Duration } from "src/dispute/lib/LibUDT.sol";
// import { GameType, Claim, GameTypes, Proposal, Hash } from "src/dispute/lib/Types.sol";

// Interfaces
// import { IOPContractsManager } from "interfaces/L1/IOPContractsManager.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";
import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { ISystemDictator } from "interfaces/legacy/ISystemDictator.sol";
import { IOwnable } from "interfaces/universal/IOwnable.sol";

/// @title Deploy
/// @notice Script used to deploy a bedrock system. The entire system is deployed within the `run` function.
///         To add a new contract to the system, add a public function that deploys that individual contract.
///         Then add a call to that function inside of `run`. Be sure to call the `save` function after each
///         deployment so that hardhat-deploy style artifacts can be generated using a call to `sync()`.
///         This contract must not have constructor logic because it is set into state using `etch`.
contract Deploy is Deployer {
    using stdJson for string;

    ////////////////////////////////////////////////////////////////
    //                        Modifiers                           //
    ////////////////////////////////////////////////////////////////

    /// @notice Modifier that wraps a function in broadcasting.
    modifier broadcast() {
        vm.startBroadcast(msg.sender);
        _;
        vm.stopBroadcast();
    }

    /// @notice Modifier that will only allow a function to be called on devnet.
    modifier onlyDevnet() {
        uint256 chainid = block.chainid;
        if (chainid == Chains.LocalDevnet || chainid == Chains.GethDevnet) {
            _;
        }
    }

    // /// @notice Modifier that wraps a function with statediff recording.
    // ///         The returned AccountAccess[] array is then written to
    // ///         the `snapshots/state-diff/<name>.json` output file.
    // modifier stateDiff() {
    //     vm.startStateDiffRecording();
    //     _;
    //     VmSafe.AccountAccess[] memory accesses = vm.stopAndReturnStateDiff();
    //     console.log(
    //         "Writing %d state diff account accesses to snapshots/state-diff/%s.json",
    //         accesses.length,
    //         vm.toString(block.chainid)
    //     );
    //     string memory json = StateDiff.encodeAccountAccesses(accesses);
    //     string memory statediffPath =
    //         string.concat(vm.projectRoot(), "/snapshots/state-diff/", vm.toString(block.chainid), ".json");
    //     vm.writeJson({ json: json, path: statediffPath });
    // }

    ////////////////////////////////////////////////////////////////
    //                        Accessors                           //
    ////////////////////////////////////////////////////////////////

    // /// @notice The create2 salt used for deployment of the contract implementations.
    // ///         Using this helps to reduce config across networks as the implementation
    // ///         addresses will be the same across networks when deployed with create2.
    // function _implSalt() internal view returns (bytes32) {
    //     return keccak256(bytes(Config.implSalt()));
    // }

    /// @notice Returns the proxy addresses, not reverting if any are unset.
    function _proxies() internal view returns (Types.ContractSet memory proxies_) {
        proxies_ = Types.ContractSet({
            ProxyAdmin: artifacts.getAddress("ProxyAdmin"),
            AddressManager: artifacts.getAddress("AddressManager"),
            L1CrossDomainMessenger: artifacts.getAddress("L1CrossDomainMessengerProxy"),
            L1StandardBridge: artifacts.getAddress("L1StandardBridgeProxy"),
            L2OutputOracle: artifacts.getAddress("L2OutputOracleProxy"),
            OptimismMintableERC20Factory: artifacts.getAddress("OptimismMintableERC20FactoryProxy"),
            OptimismPortal: artifacts.getAddress("OptimismPortalProxy"),
            SystemConfig: artifacts.getAddress("SystemConfigProxy"),
            L1ERC721Bridge: artifacts.getAddress("L1ERC721BridgeProxy")
        });
    }

    ////////////////////////////////////////////////////////////////
    //                    SetUp and Run                           //
    ////////////////////////////////////////////////////////////////

    /// @notice Deploy all of the L1 contracts necessary for a full Superchain with a single Op Chain.
    function run() public {
        console.log("Deploying a fresh OP Stack without SuperchainConfig");
        _run();
    }

    // /// @notice Deploy all L1 contracts and write the state diff to a file.
    // ///         Used to generate kontrol tests.
    // function runWithStateDiff() public stateDiff {
    //     _run({ _needsSuperchain: true });
    // }

    /// @notice Internal function containing the deploy logic.
    function _run() internal virtual {
        console.log("start of L1 Deploy!");

        deployImplementations();

        // Deploy Current OPChain Contracts
        deployOpChain();

        console.log("set up op chain!");
    }

    ////////////////////////////////////////////////////////////////
    //           High Level Deployment Functions                  //
    ////////////////////////////////////////////////////////////////

    /// @notice Deploy all of the implementations
    function deployImplementations() public {
        console.log("Deploying implementations");

        DeployImplementations di = new DeployImplementations();
        DeployImplementations.Input memory dii = DeployImplementations.Input({
            systemConfig_owner: cfg.finalSystemOwner(),
            systemConfig_overhead: cfg.gasPriceOracleOverhead(),
            systemConfig_scalar: cfg.gasPriceOracleScalar(),
            systemConfig_batcherHash: bytes32(uint256(uint160(cfg.batchSenderAddress()))),
            systemConfig_gasLimit: uint64(cfg.l2GenesisBlockGasLimit()),
            systemConfig_baseFee: cfg.l2GenesisBlockBaseFeePerGas(),
            systemConfig_unsafeBlockSigner: cfg.p2pSequencerAddress(),
            systemConfig_config: defaultResourceConfig(),
            optimismPortal: IOptimismPortal(payable(address(0))),
            l1mnt: cfg.l1MantleToken(),
            l1CrossDomainMessenger: IL1CrossDomainMessenger(address(0)),
            l2OutputOracle: IL2OutputOracle(address(0)),
            systemConfig: ISystemConfig(address(0)),
            l1StandardBridge: IL1StandardBridge(address(0)),
            l1ERC721Bridge_otherBridge: address(0),
            l2OutputOracle_submissionInterval: cfg.l2OutputOracleSubmissionInterval(),
            l2OutputOracle_l2BlockTime: cfg.l2BlockTime(),
            l2OutputOracle_startingBlockNumber: 0,
            l2OutputOracle_startingTimestamp: 0,
            l2OutputOracle_proposer: cfg.l2OutputOracleProposer(),
            l2OutputOracle_challenger: cfg.l2OutputOracleChallenger(),
            l2OutputOracle_finalizationPeriodSeconds: cfg.finalizationPeriodSeconds(),
            optimismPortal_guardian: cfg.portalGuardian(),
            optimismPortal_paused: true,
            l1ContractsRelease: "dev",
            upgradeController: cfg.finalSystemOwner()
        });

        DeployImplementations.Output memory dio = di.run(dii);

        // Save the implementation addresses which are needed outside of this function or script.
        // When called in a fork test, this will overwrite the existing implementations.
        artifacts.save("L1CrossDomainMessenger", address(dio.l1CrossDomainMessengerImpl));
        artifacts.save("L1StandardBridge", address(dio.l1StandardBridgeImpl));
        artifacts.save("L2OutputOracle", address(dio.l2OutputOracleImpl));
        artifacts.save("OptimismMintableERC20Factory", address(dio.optimismMintableERC20FactoryImpl));
        artifacts.save("OptimismPortal", address(dio.optimismPortalImpl));
        artifacts.save("SystemConfig", address(dio.systemConfigImpl));
        artifacts.save("L1ERC721Bridge", address(dio.l1ERC721BridgeImpl));
        // artifacts.save("OPContractsManager", address(dio.opcm));

        // Get a contract set from the implementation addresses which were just deployed.
        Types.ContractSet memory impls = Types.ContractSet({
            ProxyAdmin: address(0),
            AddressManager: address(0),
            L1CrossDomainMessenger: address(dio.l1CrossDomainMessengerImpl),
            L1StandardBridge: address(dio.l1StandardBridgeImpl),
            L2OutputOracle: address(dio.l2OutputOracleImpl),
            OptimismMintableERC20Factory: address(dio.optimismMintableERC20FactoryImpl),
            OptimismPortal: address(dio.optimismPortalImpl),
            SystemConfig: address(dio.systemConfigImpl),
            L1ERC721Bridge: address(dio.l1ERC721BridgeImpl)
        });

        ChainAssertions.checkL1CrossDomainMessenger({ _contracts: impls, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkL1StandardBridge({ _contracts: impls, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkL1ERC721Bridge({ _contracts: impls, _isProxy: false });
        ChainAssertions.checkOptimismPortal({ _contracts: impls, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkL2OutputOracle({ _contracts: impls, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkOptimismMintableERC20Factory({ _contracts: impls, _isProxy: false });
        ChainAssertions.checkSystemConfig({ _contracts: impls, _cfg: cfg, _isProxy: false });

        // ChainAssertions.checkOPContractsManager({
        //     _impls: impls,
        //     _proxies: _proxies(),
        //     _opcm: IOPContractsManager(address(dio.opcm)),
        //     _mips: IMIPS(address(dio.mipsSingleton)),
        //     _superchainProxyAdmin: superchainProxyAdmin
        // });
    }

    /// @notice Deploy all of the proxies, ProxyAdmin and AddressManager, for legacy usage. Will be removed once we have
    /// a bespoke OPCM
    function deployProxiesAndAddressManager() public {
        console.log("Deploying proxies and address manager");

        DeployProxies dp = new DeployProxies();
        DeployProxies.Output memory dpo = dp.run(msg.sender);

        // Save all deploy outputs
        artifacts.save("AddressManager", address(dpo.addressManager));
        artifacts.save("ProxyAdmin", address(dpo.proxyAdmin));
        artifacts.save("L1StandardBridgeProxy", address(dpo.l1StandardBridgeProxy));
        artifacts.save("L2OutputOracleProxy", address(dpo.l2OutputOracleProxy));
        artifacts.save("L1CrossDomainMessengerProxy", address(dpo.l1CrossDomainMessengerProxy));
        artifacts.save("OptimismPortalProxy", address(dpo.optimismPortalProxy));
        artifacts.save("OptimismMintableERC20FactoryProxy", address(dpo.optimismMintableERC20FactoryProxy));
        artifacts.save("L1ERC721BridgeProxy", address(dpo.l1ERC721BridgeProxy));
        artifacts.save("SystemConfigProxy", address(dpo.systemConfigProxy));
    }

    function initializeProxies() public {
        console.log("Initializing proxies");

        initializeSystemConfig();
        initializeL1StandardBridge();
        initializeL1ERC721Bridge();
        initializeOptimismMintableERC20Factory();
        initializeL1CrossDomainMessenger();
        initializeL2OutputOracle();
        initializeOptimismPortal();
    }

    /// @notice Deploy all of the OP Chain specific contracts
    function deployOpChain() public {
        console.log("Deploying OP Chain");

        // // Ensure that the requisite contracts are deployed
        // IOPContractsManager opcm = IOPContractsManager(artifacts.mustGetAddress("OPContractsManager"));

        // IOPContractsManager.DeployInput memory deployInput = getDeployInput();
        // IOPContractsManager.DeployOutput memory deployOutput = opcm.deploy(deployInput);

        // TODO: might change to bespoke opcm in future
        deployProxiesAndAddressManager();

        initializeProxies();

        transferOwnership(artifacts.mustGetAddress("ProxyAdmin"), cfg.finalSystemOwner());

        // Store code in the Final system owner address so that it can be used for prank delegatecalls
        // Store "fe" opcode so that accidental calls to this address revert
        vm.etch(cfg.finalSystemOwner(), hex"fe");

        ChainAssertions.postDeployAssertions({ _prox: _proxies(), _cfg: cfg });
    }

    ////////////////////////////////////////////////////////////////
    //                Proxy Initialization Functions               //
    ////////////////////////////////////////////////////////////////

    function initializeSystemConfig() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("SystemConfigProxy");
        address impl = artifacts.mustGetAddress("SystemConfig");

        proxyAdmin.upgradeAndCall(
            payable(address(proxy)),
            impl,
            abi.encodeCall(
                ISystemConfig.initialize,
                (
                    cfg.finalSystemOwner(),
                    cfg.gasPriceOracleOverhead(),
                    cfg.gasPriceOracleScalar(),
                    bytes32(uint256(uint160(cfg.batchSenderAddress()))),
                    uint64(cfg.l2GenesisBlockGasLimit()),
                    cfg.l2GenesisBlockBaseFeePerGas(),
                    cfg.p2pSequencerAddress(),
                    defaultResourceConfig()
                )
            )
        );
    }

    function initializeL1StandardBridge() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L1StandardBridgeProxy");
        address impl = artifacts.mustGetAddress("L1StandardBridge");

        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeL1ERC721Bridge() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L1ERC721BridgeProxy");
        address impl = artifacts.mustGetAddress("L1ERC721Bridge");

        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeOptimismMintableERC20Factory() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("OptimismMintableERC20FactoryProxy");
        address impl = artifacts.mustGetAddress("OptimismMintableERC20Factory");

        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeL1CrossDomainMessenger() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L1CrossDomainMessengerProxy");
        address impl = artifacts.mustGetAddress("L1CrossDomainMessenger");

        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, abi.encodeCall(IL1CrossDomainMessenger.initialize, ()));
    }

    function initializeL2OutputOracle() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L2OutputOracleProxy");
        address impl = artifacts.mustGetAddress("L2OutputOracle");

        proxyAdmin.upgradeAndCall(
            payable(address(proxy)),
            impl,
            abi.encodeCall(
                IL2OutputOracle.initialize,
                (cfg.l2OutputOracleStartingBlockNumber(), cfg.l2OutputOracleStartingTimestamp())
            )
        );
    }

    function initializeOptimismPortal() public {
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("OptimismPortalProxy");
        address impl = artifacts.mustGetAddress("OptimismPortal");

        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, abi.encodeCall(IOptimismPortal.initialize, (false)));
    }

    ////////////////////////////////////////////////////////////////
    //                Proxy Deployment Functions                  //
    ////////////////////////////////////////////////////////////////

    // /// @notice Deploys an ERC1967Proxy contract with a specified owner.
    // /// @param _name The name of the proxy contract to be deployed.
    // /// @param _proxyOwner The address of the owner of the proxy contract.
    // /// @return addr_ The address of the deployed proxy contract.
    // function deployERC1967ProxyWithOwner(
    //     string memory _name,
    //     address _proxyOwner
    // )
    //     public
    //     broadcast
    //     returns (address addr_)
    // {
    //     IProxy proxy = IProxy(
    //         DeployUtils.create2AndSave({
    //             _save: artifacts,
    //             _salt: keccak256(abi.encode(_implSalt(), _name)),
    //             _name: "Proxy",
    //             _nick: _name,
    //             _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (_proxyOwner)))
    //         })
    //     );
    //     require(EIP1967Helper.getAdmin(address(proxy)) == _proxyOwner, "Deploy: EIP1967Proxy admin not set");
    //     addr_ = address(proxy);
    // }

    /// @notice Deploys an ERC1967Proxy contract with a specified owner using create1.
    /// @param _name The name of the proxy contract to be deployed.
    /// @param _proxyOwner The address of the owner of the proxy contract.
    /// @return addr_ The address of the deployed proxy contract.
    function deployERC1967ProxyWithOwnerCreate1(
        string memory _name,
        address _proxyOwner
    )
        public
        broadcast
        returns (address addr_)
    {
        IProxy proxy = IProxy(
            DeployUtils.create1AndSave({
                _save: artifacts,
                _name: "Proxy",
                _nick: _name,
                _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (_proxyOwner)))
            })
        );
        require(proxy.admin() == _proxyOwner, "Deploy: EIP1967Proxy admin not set");
        addr_ = address(proxy);
    }

    function initSystemDictator() public {
        console.log("Initializing SystemDictator");
        address impl = DeployUtils.create1AndSave({
            _save: artifacts,
            _name: "SystemDictator",
            _nick: "SystemDictatorImpl",
            _args: DeployUtils.encodeConstructor(abi.encodeCall(ISystemDictator.__constructor__, ()))
        });
        IProxy proxy = IProxy(artifacts.mustGetAddress("SystemDictatorProxy"));

        ISystemDictator.DeployConfig memory deployConfig = ISystemDictator.DeployConfig({
            globalConfig: ISystemDictator.GlobalConfig({
                proxyAdmin: artifacts.mustGetAddress("ProxyAdmin"),
                controller: cfg.controller(),
                finalOwner: cfg.finalSystemOwner(),
                addressManager: artifacts.mustGetAddress("AddressManager")
            }),
            proxyAddressConfig: ISystemDictator.ProxyAddressConfig({
                l2OutputOracleProxy: artifacts.mustGetAddress("L2OutputOracleProxy"),
                optimismPortalProxy: artifacts.mustGetAddress("OptimismPortalProxy"),
                l1CrossDomainMessengerProxy: artifacts.mustGetAddress("L1CrossDomainMessengerProxy"),
                l1StandardBridgeProxy: artifacts.mustGetAddress("L1StandardBridgeProxy"),
                optimismMintableERC20FactoryProxy: artifacts.mustGetAddress("OptimismMintableERC20FactoryProxy"),
                l1ERC721BridgeProxy: artifacts.mustGetAddress("L1ERC721BridgeProxy"),
                systemConfigProxy: artifacts.mustGetAddress("SystemConfigProxy")
            }),
            implementationAddressConfig: ISystemDictator.ImplementationAddressConfig({
                l2OutputOracleImpl: artifacts.mustGetAddress("L2OutputOracle"),
                optimismPortalImpl: artifacts.mustGetAddress("OptimismPortal"),
                l1CrossDomainMessengerImpl: artifacts.mustGetAddress("L1CrossDomainMessenger"),
                l1StandardBridgeImpl: artifacts.mustGetAddress("L1StandardBridge"),
                optimismMintableERC20FactoryImpl: artifacts.mustGetAddress("OptimismMintableERC20Factory"),
                l1ERC721BridgeImpl: artifacts.mustGetAddress("L1ERC721Bridge"),
                portalSenderImpl: address(0),
                systemConfigImpl: artifacts.mustGetAddress("SystemConfig")
            }),
            systemConfigConfig: ISystemDictator.SystemConfigConfig({
                owner: cfg.finalSystemOwner(),
                overhead: cfg.gasPriceOracleOverhead(),
                scalar: cfg.gasPriceOracleScalar(),
                batcherHash: bytes32(uint256(uint160(cfg.batchSenderAddress()))),
                gasLimit: uint64(cfg.l2GenesisBlockGasLimit()),
                baseFee: cfg.l2GenesisBlockBaseFeePerGas(),
                unsafeBlockSigner: cfg.p2pSequencerAddress(),
                resourceConfig: defaultResourceConfig()
            })
        });
        proxy.upgradeToAndCall(impl, abi.encodeCall(ISystemDictator.initialize, (deployConfig)));
    }

    // /// @notice Get the DeployInput struct to use for testing
    // function getDeployInput() public view returns (IOPContractsManager.DeployInput memory) {
    //     string memory saltMixer = "salt mixer";
    //     return IOPContractsManager.DeployInput({
    //         roles: IOPContractsManager.Roles({
    //             opChainProxyAdminOwner: cfg.finalSystemOwner(),
    //             systemConfigOwner: cfg.finalSystemOwner(),
    //             batcher: cfg.batchSenderAddress(),
    //             unsafeBlockSigner: cfg.p2pSequencerAddress(),
    //             proposer: cfg.l2OutputOracleProposer(),
    //             challenger: cfg.l2OutputOracleChallenger()
    //         }),
    //         basefeeScalar: cfg.basefeeScalar(),
    //         blobBasefeeScalar: cfg.blobbasefeeScalar(),
    //         l2ChainId: cfg.l2ChainID(),
    //         startingAnchorRoot: abi.encode(
    //             Proposal({ root: Hash.wrap(cfg.faultGameGenesisOutputRoot()), l2SequenceNumber:
    // cfg.faultGameGenesisBlock() })
    //         ),
    //         saltMixer: saltMixer,
    //         gasLimit: uint64(cfg.l2GenesisBlockGasLimit()),
    //         disputeGameType: GameTypes.PERMISSIONED_CANNON,
    //         disputeAbsolutePrestate: Claim.wrap(bytes32(cfg.faultGameAbsolutePrestate())),
    //         disputeMaxGameDepth: cfg.faultGameMaxDepth(),
    //         disputeSplitDepth: cfg.faultGameSplitDepth(),
    //         disputeClockExtension: Duration.wrap(uint64(cfg.faultGameClockExtension())),
    //         disputeMaxClockDuration: Duration.wrap(uint64(cfg.faultGameMaxClockDuration()))
    //     });
    // }

    /// @notice Returns the default resource config. We encourage using interface instead of the original contract.
    function defaultResourceConfig() public view returns (IResourceMetering.ResourceConfig memory) {
        return abi.decode(abi.encode(Constants.DEFAULT_RESOURCE_CONFIG()), (IResourceMetering.ResourceConfig));
    }

    function transferOwnership(address _contract, address _newOwner) public {
        IOwnable(_contract).transferOwnership(_newOwner);
    }
}
