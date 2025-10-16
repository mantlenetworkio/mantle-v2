// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";

// Libraries
import { Chains } from "scripts/libraries/Chains.sol";
import { Types } from "scripts/libraries/Types.sol";

// Interfaces
import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
import { IOptimismPortal } from "interfaces/L1/IOptimismPortal.sol";
import { ISystemConfig } from "interfaces/L1/ISystemConfig.sol";
import { IL1CrossDomainMessenger } from "interfaces/L1/IL1CrossDomainMessenger.sol";
import { IL1ERC721Bridge } from "interfaces/L1/IL1ERC721Bridge.sol";
import { IL1StandardBridge } from "interfaces/L1/IL1StandardBridge.sol";
import { IOptimismMintableERC20Factory } from "interfaces/universal/IOptimismMintableERC20Factory.sol";
import { IL2OutputOracle } from "interfaces/L1/IL2OutputOracle.sol";
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";
import { Solarray } from "scripts/libraries/Solarray.sol";

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
    }

    struct Output {
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

    function runWithBytes(bytes memory _input, address _deployer) public returns (bytes memory) {
        Input memory input = abi.decode(_input, (Input));
        Output memory output = run(input, _deployer);
        return abi.encode(output);
    }

    function run(Input memory _input, address _deployer) public returns (Output memory output_) {
        assertValidInput(_input);

        // Deploy the implementations.
        vm.startBroadcast(_deployer);

        deploySystemConfigImpl(_input, output_);
        deployL1CrossDomainMessengerImpl(_input, output_);
        deployL1ERC721BridgeImpl(_input, output_);
        deployL1StandardBridgeImpl(_input, output_);
        deployOptimismMintableERC20FactoryImpl(_input, output_);
        deployOptimismPortalImpl(_input, output_);
        deployL2OutputOracleImpl(_input, output_);
        vm.stopBroadcast();

        assertValidOutput(_input, output_);
    }

    // -------- Deployment Steps --------

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
                _name: "OptimismPortal",
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

    function assertValidInput(Input memory _input) private pure {
        // TODO:
        // if deployer != controller, check portal guardian
        require(address(_input.l1mnt) != address(0), "DeployImplementations: l1mnt not set");

        require(_input.l2OutputOracle_l2BlockTime > 0, "DeployImplementations: l2BlockTime not set");
        require(
            _input.l2OutputOracle_submissionInterval > _input.l2OutputOracle_l2BlockTime,
            "DeployImplementations: submissionInterval must be greater than the l2BlockTime"
        );
    }

    function assertValidOutput(Input memory _input, Output memory _output) private {
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
    }
}
