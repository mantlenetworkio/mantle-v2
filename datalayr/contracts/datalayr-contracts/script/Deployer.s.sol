// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/presets/ERC20PresetFixedSupply.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import "@openzeppelin/contracts/proxy/beacon/IBeacon.sol";
import "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";

import "@eigenlayer/contracts/interfaces/IEigenLayrDelegation.sol";
import "@eigenlayer/contracts/interfaces/IInvestmentManager.sol";
import "@eigenlayer/contracts/interfaces/IInvestmentStrategy.sol";

import "@eigenlayer/contracts/permissions/PauserRegistry.sol";

import "@eigenlayer/contracts/middleware/BLSPublicKeyCompendium.sol";
import "@eigenlayer/contracts/middleware/BLSRegistry.sol";
import "@eigenlayer/contracts/interfaces/IRegistryPermission.sol";

import "@eigenlayer/test/mocks/EmptyContract.sol";

import "../src/contracts/core/DataLayrServiceManager.sol";
import "../src/contracts/core/DataLayrChallengeUtils.sol";
import "../src/contracts/core/DataLayrChallenge.sol";
import "../src/contracts/libraries/DataStoreUtils.sol";

import "forge-std/Test.sol";

import "forge-std/Script.sol";
import "forge-std/StdJson.sol";


// # To load the variables in the .env file
// source .env

// # To deploy and verify our contract
// forge script script/Deployer.s.sol:DataLayrDeployer --rpc-url $RPC_URL  --private-key $PRIVATE_KEY --broadcast -vvvv

//TODO: encode data properly so that we initialize TransparentUpgradeableProxy contracts in their constructor rather than a separate call (if possible)
contract DataLayrDeployer is Script, Test {
    uint256 public constant DURATION_SCALE = 1 hours;

    // EigenLayer contracts
    IEigenLayrDelegation public delegation;
    IInvestmentManager public investmentManager;

    // DataLayr contracts
    ProxyAdmin public dataLayrProxyAdmin;
    PauserRegistry public dataLayrPauserReg;

    DataLayrChallengeUtils public challengeUtils;
    BLSPublicKeyCompendium public pubkeyCompendium;
    BLSRegistry public dlReg;
    DataLayrServiceManager public dlsm;
    DataLayrChallenge public dlldc;

    DataLayrChallengeUtils public challengeUtilsImplementation;
    BLSPublicKeyCompendium public pubkeyCompendiumImplementation;
    BLSRegistry public dlRegImplementation;
    DataLayrServiceManager public dlsmImplementation;
    DataLayrChallenge public dlldcImplementation;

    // testing/mock contracts
    IERC20 public mantleToken;
    IInvestmentStrategy public mantleFirstStrat;
    IInvestmentStrategy public mantleSencodStrat;
    IRegistryPermission public registryPermission;
    EmptyContract public emptyContract;

    uint256 public gasLimit = 750000;

    // deploy all the DataLayr contracts. Relies on many EL contracts having already been deployed.
    function run() external {
        address pauser = msg.sender;
        address unpauser = msg.sender;
        address dataLayrReputedMultisig = msg.sender;
        address dataLayrTeamMultisig = msg.sender;

        string memory deployConfigJson = vm.readFile("data/datalayr_deploy_config.json");
        mantleToken = IERC20(stdJson.readAddress(deployConfigJson, ".mantle"));
        investmentManager = IInvestmentManager(stdJson.readAddress(deployConfigJson, ".investmentManager"));
        delegation = IEigenLayrDelegation(stdJson.readAddress(deployConfigJson, ".delegation"));
        mantleFirstStrat = IInvestmentStrategy(stdJson.readAddress(deployConfigJson, ".mantleFirstStrat"));
        mantleSencodStrat = IInvestmentStrategy(stdJson.readAddress(deployConfigJson, ".mantleSencodStrat"));
        registryPermission = IRegistryPermission(stdJson.readAddress(deployConfigJson, ".rgPermission"));

        emit log_address(address(mantleToken));

        vm.startBroadcast();

        // deploy proxy admin for ability to upgrade proxy contracts
        dataLayrProxyAdmin = new ProxyAdmin();

        // deploy pauser registry
        dataLayrPauserReg = new PauserRegistry(pauser, unpauser);

        emptyContract = new EmptyContract();

        // hard-coded inputs
        uint256 feePerBytePerTime = 1;
        uint256 _paymentFraudproofCollateral = 1e16;

        /**
         * First, deploy upgradeable proxy contracts that **will point** to the implementations. Since the implementation contracts are
         * not yet deployed, we give these proxies an empty contract as the initial implementation, to act as if they have no code.
         */
        challengeUtils = DataLayrChallengeUtils(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(dataLayrProxyAdmin), ""))
        );
        dlsm = DataLayrServiceManager(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(dataLayrProxyAdmin), ""))
        );
        pubkeyCompendium = BLSPublicKeyCompendium(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(dataLayrProxyAdmin), ""))
        );
        dlldc = DataLayrChallenge(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(dataLayrProxyAdmin), ""))
        );
        dlReg = BLSRegistry(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(dataLayrProxyAdmin), ""))
        );

        // Second, deploy the *implementation* contracts, using the *proxy contracts* as inputs
        challengeUtilsImplementation = new DataLayrChallengeUtils();
        dlsmImplementation = new DataLayrServiceManager(
            dlReg,
            investmentManager,
            delegation,
            mantleToken,
            dlldc,
            DataLayrBombVerifier(address(0)),   // TODO: fix this
            registryPermission
        );

        pubkeyCompendiumImplementation = new BLSPublicKeyCompendium(registryPermission);

        dlldcImplementation = new DataLayrChallenge(dlsm, dlReg, challengeUtils);
        {
            uint8 _NUMBER_OF_QUORUMS = 2;
            dlRegImplementation = new BLSRegistry(
                investmentManager,
                dlsm,
                _NUMBER_OF_QUORUMS,
                pubkeyCompendium,
                registryPermission,
                msg.sender
            );
        }

        // Third, upgrade the proxy contracts to use the correct implementation contracts and initialize them.
        dataLayrProxyAdmin.upgrade(
            TransparentUpgradeableProxy(payable(address(challengeUtils))),
            address(challengeUtilsImplementation)
        );
        {
            uint16 quorumThresholdBasisPoints = 9000;
            uint16 adversaryThresholdBasisPoints = 4000;
            dataLayrProxyAdmin.upgradeAndCall(
                TransparentUpgradeableProxy(payable(address(dlsm))),
                address(dlsmImplementation),
                abi.encodeWithSelector(
                    DataLayrServiceManager.initialize.selector,
                    dataLayrPauserReg,
                    dataLayrReputedMultisig,
                    quorumThresholdBasisPoints,
                    adversaryThresholdBasisPoints,
                    feePerBytePerTime,
                    dataLayrTeamMultisig
                )
            );
        }
        dataLayrProxyAdmin.upgrade(
            TransparentUpgradeableProxy(payable(address(pubkeyCompendium))),
            address(pubkeyCompendiumImplementation)
        );
        dataLayrProxyAdmin.upgrade(
            TransparentUpgradeableProxy(payable(address(dlldc))),
            address(dlldcImplementation)
        );
        {
            uint96 multiplier = 1e18;
            uint8 _NUMBER_OF_QUORUMS = 2;
            uint256[] memory _quorumBips = new uint256[](_NUMBER_OF_QUORUMS);
            // split 60% ETH quorum, 40% EIGEN quorum
            _quorumBips[0] = 6000;
            _quorumBips[1] = 4000;
            VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[] memory mantleFirstStratsAndMultipliers =
                new VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[](1);
            mantleFirstStratsAndMultipliers[0].strategy = mantleFirstStrat;
            mantleFirstStratsAndMultipliers[0].multiplier = multiplier;
            VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[] memory mantleSencodStratsAndMultipliers =
                new VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[](1);
            mantleSencodStratsAndMultipliers[0].strategy = mantleSencodStrat;
            mantleSencodStratsAndMultipliers[0].multiplier = multiplier;

            dataLayrProxyAdmin.upgradeAndCall(
                TransparentUpgradeableProxy(payable(address(dlReg))),
                address(dlRegImplementation),
                abi.encodeWithSelector(BLSRegistry.initialize.selector, _quorumBips, dataLayrTeamMultisig, mantleFirstStratsAndMultipliers, mantleSencodStratsAndMultipliers)
            );
        }

        vm.writeFile("data/dlsm.addr", vm.toString(address(dlsm)));
        vm.writeFile("data/dlReg.addr", vm.toString(address(dlReg)));
        vm.writeFile("data/pubkeyCompendium.addr", vm.toString(address(pubkeyCompendium)));
        vm.writeFile("data/registryPermission.addr", vm.toString(address(registryPermission)));

        vm.stopBroadcast();
    }
}
