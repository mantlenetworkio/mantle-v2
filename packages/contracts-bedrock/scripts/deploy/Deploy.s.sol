// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Testing
import { VmSafe } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";

// Scripts
import { Deployer } from "scripts/deploy/Deployer.sol";
import { Chains } from "scripts/libraries/Chains.sol";
import { Config } from "scripts/libraries/Config.sol";
import { StateDiff } from "scripts/libraries/StateDiff.sol";
import { ChainAssertions } from "scripts/deploy/ChainAssertions.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { DeployImplementations } from "scripts/deploy/DeployImplementations.s.sol";
import { DeployProxies } from "scripts/deploy/DeployProxy.s.sol";

// Libraries
import { Types } from "scripts/libraries/Types.sol";
import { Constants } from "src/libraries/Constants.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";

// Interfaces
import { IOwnable } from "interfaces/universal/IOwnable.sol";
import { IProxy } from "interfaces/universal/IProxy.sol";
import { IProxyAdmin } from "interfaces/universal/IProxyAdmin.sol";
import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";
import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";

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

    /// @notice Modifier that wraps a function with statediff recording.
    ///         The returned AccountAccess[] array is then written to
    ///         the `snapshots/state-diff/<name>.json` output file.
    modifier stateDiff() {
        vm.startStateDiffRecording();
        _;
        VmSafe.AccountAccess[] memory accesses = vm.stopAndReturnStateDiff();
        console.log(
            "Writing %d state diff account accesses to snapshots/state-diff/%s.json",
            accesses.length,
            vm.toString(block.chainid)
        );
        string memory json = StateDiff.encodeAccountAccesses(accesses);
        string memory statediffPath =
            string.concat(vm.projectRoot(), "/snapshots/state-diff/", vm.toString(block.chainid), ".json");
        vm.writeJson({ json: json, path: statediffPath });
    }

    ////////////////////////////////////////////////////////////////
    //                        Accessors                           //
    ////////////////////////////////////////////////////////////////

    /// @notice The create2 salt used for deployment of the contract implementations.
    ///         Using this helps to reduce config across networks as the implementation
    ///         addresses will be the same across networks when deployed with create2.
    function _implSalt() internal view returns (bytes32) {
        return keccak256(bytes(Config.implSalt()));
    }

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
            L1ERC721Bridge: artifacts.getAddress("L1ERC721BridgeProxy"),
            L1CrossDomainMessengerImpl: address(0),
            L1StandardBridgeImpl: address(0),
            L2OutputOracleImpl: address(0),
            OptimismMintableERC20FactoryImpl: address(0),
            OptimismPortalImpl: address(0),
            SystemConfigImpl: address(0),
            L1ERC721BridgeImpl: address(0)
        });
    }

    ////////////////////////////////////////////////////////////////
    //                    SetUp and Run                           //
    ////////////////////////////////////////////////////////////////

    /// @notice Deploy all of the L1 contracts necessary for an Op Chain.
    function run() public {
        console.log("Deploying a fresh OP Stack");
        _run();
    }

    /// @notice Deploy all L1 contracts and write the state diff to a file.
    ///         Used to generate kontrol tests.
    function runWithStateDiff() public stateDiff {
        _run();
    }

    /// @notice Internal function containing the deploy logic.
    function _run() internal virtual {
        console.log("start of L1 Deploy!");

        deployProxiesAndAddressManager();

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

        DeployImplementations.Output memory dio = di.run(
            DeployImplementations.Input({
                systemConfig_owner: cfg.finalSystemOwner(),
                systemConfig_basefeeScalar: uint32(cfg.basefeeScalar()),
                systemConfig_blobbasefeeScalar: uint32(cfg.blobbasefeeScalar()),
                systemConfig_batcherHash: bytes32(uint256(uint160(cfg.batchSenderAddress()))),
                systemConfig_gasLimit: uint64(cfg.l2GenesisBlockGasLimit()),
                systemConfig_baseFee: cfg.l2GenesisBlockBaseFeePerGas(),
                systemConfig_unsafeBlockSigner: cfg.p2pSequencerAddress(),
                systemConfig_config: defaultResourceConfig(),
                optimismPortal: IOptimismPortal(payable(artifacts.mustGetAddress("OptimismPortalProxy"))),
                l1mnt: cfg.l1MantleToken(),
                l1CrossDomainMessenger: IL1CrossDomainMessenger(artifacts.mustGetAddress("L1CrossDomainMessengerProxy")),
                l2OutputOracle: IL2OutputOracle(artifacts.mustGetAddress("L2OutputOracleProxy")),
                systemConfig: ISystemConfig(artifacts.mustGetAddress("SystemConfigProxy")),
                l1StandardBridge: IL1StandardBridge(artifacts.mustGetAddress("L1StandardBridgeProxy")),
                l1ERC721Bridge_otherBridge: Predeploys.L2_ERC721_BRIDGE,
                l2OutputOracle_submissionInterval: cfg.l2OutputOracleSubmissionInterval(),
                l2OutputOracle_l2BlockTime: cfg.l2BlockTime(),
                l2OutputOracle_startingBlockNumber: 0,
                l2OutputOracle_startingTimestamp: 0,
                l2OutputOracle_proposer: cfg.l2OutputOracleProposer(),
                l2OutputOracle_challenger: cfg.l2OutputOracleChallenger(),
                l2OutputOracle_finalizationPeriodSeconds: cfg.finalizationPeriodSeconds(),
                optimismPortal_guardian: cfg.portalGuardian(),
                optimismPortal_paused: true
            }),
            msg.sender
        );

        // Save the implementation addresses which are needed outside of this function or script.
        // When called in a fork test, this will overwrite the existing implementations.
        artifacts.save("L1CrossDomainMessenger", address(dio.l1CrossDomainMessengerImpl));
        artifacts.save("L1StandardBridge", address(dio.l1StandardBridgeImpl));
        artifacts.save("L2OutputOracle", address(dio.l2OutputOracleImpl));
        artifacts.save("OptimismMintableERC20Factory", address(dio.optimismMintableERC20FactoryImpl));
        artifacts.save("OptimismPortal", address(dio.optimismPortalImpl));
        artifacts.save("SystemConfig", address(dio.systemConfigImpl));
        artifacts.save("L1ERC721Bridge", address(dio.l1ERC721BridgeImpl));

        // Get a contract set
        Types.ContractSet memory proxies = getDeployOutput();

        ChainAssertions.checkL1CrossDomainMessenger({ _contracts: proxies, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkL1StandardBridge({ _contracts: proxies, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkL1ERC721Bridge({ _contracts: proxies, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkOptimismPortal({ _contracts: proxies, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkL2OutputOracle({ _contracts: proxies, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkOptimismMintableERC20Factory({ _contracts: proxies, _cfg: cfg, _isProxy: false });
        ChainAssertions.checkSystemConfig({ _contracts: proxies, _cfg: cfg, _isProxy: false });
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

        initializeProxies();

        address proxyAdmin = artifacts.mustGetAddress("ProxyAdmin");
        address finalSystemOwner = cfg.finalSystemOwner();
        vm.broadcast(msg.sender);
        transferOwnership(proxyAdmin, finalSystemOwner);

        // Store code in the Final system owner address so that it can be used for prank delegatecalls
        // Store "fe" opcode so that accidental calls to this address revert
        vm.etch(finalSystemOwner, hex"fe");

        ChainAssertions.postDeployAssertions({ _prox: getDeployOutput(), _cfg: cfg });
    }

    ////////////////////////////////////////////////////////////////
    //                Proxy Initialization Functions               //
    ////////////////////////////////////////////////////////////////

    function initializeSystemConfig() public {
        console.log("Initializing SystemConfig");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("SystemConfigProxy");
        address impl = artifacts.mustGetAddress("SystemConfig");

        bytes memory data = abi.encodeCall(
            ISystemConfig.initialize,
            (
                cfg.finalSystemOwner(),
                uint32(cfg.basefeeScalar()), // basefeeScalar
                uint32(cfg.blobbasefeeScalar()), // blobbasefeeScalar
                bytes32(uint256(uint160(cfg.batchSenderAddress()))),
                uint64(cfg.l2GenesisBlockGasLimit()),
                cfg.l2GenesisBlockBaseFeePerGas(),
                cfg.p2pSequencerAddress(),
                defaultResourceConfig()
            )
        );
        vm.broadcast(msg.sender);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    function initializeL1StandardBridge() public {
        console.log("Initializing L1StandardBridge");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L1StandardBridgeProxy");
        address impl = artifacts.mustGetAddress("L1StandardBridge");

        vm.broadcast(msg.sender);
        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeL1ERC721Bridge() public {
        console.log("Initializing L1ERC721Bridge");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L1ERC721BridgeProxy");
        address impl = artifacts.mustGetAddress("L1ERC721Bridge");

        vm.broadcast(msg.sender);
        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeOptimismMintableERC20Factory() public {
        console.log("Initializing OptimismMintableERC20Factory");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("OptimismMintableERC20FactoryProxy");
        address impl = artifacts.mustGetAddress("OptimismMintableERC20Factory");

        vm.broadcast(msg.sender);
        proxyAdmin.upgrade(payable(address(proxy)), impl);
    }

    function initializeL1CrossDomainMessenger() public {
        console.log("Initializing L1CrossDomainMessenger");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L1CrossDomainMessengerProxy");
        address impl = artifacts.mustGetAddress("L1CrossDomainMessenger");

        bytes memory data = abi.encodeCall(IL1CrossDomainMessenger.initialize, ());
        vm.broadcast(msg.sender);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    function initializeL2OutputOracle() public {
        console.log("Initializing L2OutputOracle");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("L2OutputOracleProxy");
        address impl = artifacts.mustGetAddress("L2OutputOracle");

        bytes memory data = abi.encodeCall(
            IL2OutputOracle.initialize, (cfg.l2OutputOracleStartingBlockNumber(), cfg.l2OutputOracleStartingTimestamp())
        );
        vm.broadcast(msg.sender);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    function initializeOptimismPortal() public {
        console.log("Initializing OptimismPortal");
        IProxyAdmin proxyAdmin = IProxyAdmin(artifacts.mustGetAddress("ProxyAdmin"));
        address proxy = artifacts.mustGetAddress("OptimismPortalProxy");
        address impl = artifacts.mustGetAddress("OptimismPortal");

        bytes memory data = abi.encodeCall(IOptimismPortal.initialize, (false));
        vm.broadcast(msg.sender);
        proxyAdmin.upgradeAndCall(payable(address(proxy)), impl, data);
    }

    /// @notice Returns the default resource config. We encourage using interface instead of the original contract.
    function defaultResourceConfig() public view returns (IResourceMetering.ResourceConfig memory) {
        return abi.decode(abi.encode(Constants.DEFAULT_RESOURCE_CONFIG()), (IResourceMetering.ResourceConfig));
    }

    function transferOwnership(address _contract, address _newOwner) public {
        IOwnable(_contract).transferOwnership(_newOwner);
    }

    function getDeployOutput() public view returns (Types.ContractSet memory) {
        return Types.ContractSet({
            ProxyAdmin: artifacts.mustGetAddress("ProxyAdmin"),
            AddressManager: artifacts.mustGetAddress("AddressManager"),
            L1CrossDomainMessenger: artifacts.mustGetAddress("L1CrossDomainMessengerProxy"),
            L1StandardBridge: artifacts.mustGetAddress("L1StandardBridgeProxy"),
            L2OutputOracle: artifacts.mustGetAddress("L2OutputOracleProxy"),
            OptimismMintableERC20Factory: artifacts.mustGetAddress("OptimismMintableERC20FactoryProxy"),
            OptimismPortal: artifacts.mustGetAddress("OptimismPortalProxy"),
            SystemConfig: artifacts.mustGetAddress("SystemConfigProxy"),
            L1ERC721Bridge: artifacts.mustGetAddress("L1ERC721BridgeProxy"),
            L1CrossDomainMessengerImpl: artifacts.mustGetAddress("L1CrossDomainMessenger"),
            L1StandardBridgeImpl: artifacts.mustGetAddress("L1StandardBridge"),
            L2OutputOracleImpl: artifacts.mustGetAddress("L2OutputOracle"),
            OptimismMintableERC20FactoryImpl: artifacts.mustGetAddress("OptimismMintableERC20Factory"),
            OptimismPortalImpl: artifacts.mustGetAddress("OptimismPortal"),
            SystemConfigImpl: artifacts.mustGetAddress("SystemConfig"),
            L1ERC721BridgeImpl: artifacts.mustGetAddress("L1ERC721Bridge")
        });
    }
}
