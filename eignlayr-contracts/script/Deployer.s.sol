// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/presets/ERC20PresetFixedSupply.sol";
import "@openzeppelin/contracts/proxy/transparent/ProxyAdmin.sol";
import "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/proxy/beacon/IBeacon.sol";
import "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";

import "../src/contracts/interfaces/IEigenLayrDelegation.sol";
import "../src/contracts/core/EigenLayrDelegation.sol";
import "../src/contracts/core/SlashRecorder.sol";

import "../src/contracts/core/InvestmentManager.sol";
import "../src/contracts/strategies/InvestmentStrategyBase.sol";

import "../src/contracts/permissions/PauserRegistry.sol";
import "../src/contracts/middleware/RegistryPermission.sol";

import "../src/contracts/libraries/BytesLib.sol";

import "../src/test/mocks/EmptyContract.sol";

import "forge-std/Test.sol";

import "forge-std/Script.sol";
import "forge-std/StdJson.sol";

import "../src/contracts/libraries/BytesLib.sol";

// # To load the variables in the .env file
// source .env

// # To deploy and verify our contract
// forge script script/Deployer.s.sol:EigenLayrDeployer --rpc-url $RPC_URL  --private-key $PRIVATE_KEY --broadcast -vvvv
contract EigenLayrDeployer is Script, DSTest {
    //,
    // Signers,
    // SignatureUtils

    using BytesLib for bytes;

    Vm cheats = Vm(HEVM_ADDRESS);

    uint256 public constant DURATION_SCALE = 1 hours;

    // EigenLayer contracts
    ProxyAdmin public mantleLayrProxyAdmin;
    PauserRegistry public mantleLayrPauserReg;
    EigenLayrDelegation public delegation;
    SlashRecorder public slashRecorder;
    InvestmentManager public investmentManager;

    // DataLayr contracts
    ProxyAdmin public dataLayrProxyAdmin;
    PauserRegistry public dataLayrPauserReg;
    RegistryPermission public rgPermission;

    // testing/mock contracts
    IERC20 public mantleToken;

    InvestmentStrategyBase public mantleFirstStrat;
    InvestmentStrategyBase public mantleSencodStrat;
    InvestmentStrategyBase public baseStrategyImplementation;

    EmptyContract public emptyContract;

    uint256 nonce = 69;
    uint32 PARTIAL_WITHDRAWAL_FRAUD_PROOF_PERIOD = 7 days / 12 seconds;
    uint256 REQUIRED_BALANCE_WEI = 31.4 ether;
    uint64 MAX_PARTIAL_WTIHDRAWAL_AMOUNT_GWEI = 1 ether / 1e9;

    bytes[] registrationData;

    //strategy indexes for undelegation (see commitUndelegation function)
    uint256[] public strategyIndexes;

    address storer = address(420);
    address registrant = address(0x4206904396bF2f8b173350ADdEc5007A52664293); //sk: e88d9d864d5d731226020c5d2f02b62a4ce2a4534a39c225d32d3db795f83319

    //from testing seed phrase
    // bytes32 priv_key_0 =
    //     0x1234567812345678123456781234567812345678123456781234567812345678;
    // address acct_0 = cheats.addr(uint256(priv_key_0));

    // bytes32 priv_key_1 =
    //     0x1234567812345678123456781234567812345698123456781234567812348976;
    // address acct_1 = cheats.addr(uint256(priv_key_1));

    uint256 public constant mantleTotalSupply = 1000e18;

    uint256 public gasLimit = 750000;

    function run() external {
        vm.startBroadcast();

        emit log_address(address(this));
        address pauser = msg.sender;
        address unpauser = msg.sender;
        address mantleLayrReputedMultisig = msg.sender;


        // deploy proxy admin for ability to upgrade proxy contracts
        mantleLayrProxyAdmin = new ProxyAdmin();

        //deploy pauser registry
        mantleLayrPauserReg = new PauserRegistry(pauser, unpauser);

        /**
         * First, deploy upgradeable proxy contracts that **will point** to the implementations. Since the implementation contracts are
         * not yet deployed, we give these proxies an empty contract as the initial implementation, to act as if they have no code.
         */
        emptyContract = new EmptyContract();
        delegation = EigenLayrDelegation(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(mantleLayrProxyAdmin), ""))
        );
        slashRecorder = SlashRecorder(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(mantleLayrProxyAdmin), ""))
        );
        investmentManager = InvestmentManager(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(mantleLayrProxyAdmin), ""))
        );
        rgPermission = RegistryPermission(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(mantleLayrProxyAdmin), ""))
        );

        // Second, deploy the *implementation* contracts, using the *proxy contracts* as inputs
        EigenLayrDelegation delegationImplementation = new EigenLayrDelegation(investmentManager, rgPermission);
        SlashRecorder slashRecorderImplementation = new SlashRecorder();
        InvestmentManager investmentManagerImplementation = new InvestmentManager(delegation, rgPermission);
        RegistryPermission rgPermissionImplementation = new RegistryPermission();

        // Third, upgrade the proxy contracts to use the correct implementation contracts and initialize them.
        mantleLayrProxyAdmin.upgradeAndCall(
            TransparentUpgradeableProxy(payable(address(delegation))),
            address(delegationImplementation),
            abi.encodeWithSelector(EigenLayrDelegation.initialize.selector, mantleLayrPauserReg, mantleLayrReputedMultisig)
        );
        mantleLayrProxyAdmin.upgradeAndCall(
            TransparentUpgradeableProxy(payable(address(slashRecorder))),
            address(slashRecorderImplementation),
            abi.encodeWithSelector(EigenLayrDelegation.initialize.selector, msg.sender, mantleLayrReputedMultisig)
        );
        mantleLayrProxyAdmin.upgradeAndCall(
            TransparentUpgradeableProxy(payable(address(investmentManager))),
            address(investmentManagerImplementation),
            abi.encodeWithSelector(InvestmentManager.initialize.selector, mantleLayrPauserReg, mantleLayrReputedMultisig)
        );
        mantleLayrProxyAdmin.upgradeAndCall(
            TransparentUpgradeableProxy(payable(address(rgPermission))),
            address(rgPermissionImplementation),
            abi.encodeWithSelector(RegistryPermission.initialize.selector, msg.sender, mantleLayrReputedMultisig)
        );

        string memory mantleAddrStr = vm.envString("L1_MANTLE_ADDRESS");
        if(bytes(mantleAddrStr).length == 42) { // "0x...." string is 42 char long
            address mantleAddr = vm.envAddress("L1_MANTLE_ADDRESS");
            mantleToken = IERC20(mantleAddr);
        }

        // deploy InvestmentStrategyBase contract implementation, then create upgradeable proxy that points to implementation and initialize it
        baseStrategyImplementation = new InvestmentStrategyBase(investmentManager);
        mantleFirstStrat = InvestmentStrategyBase(
            address(
                new TransparentUpgradeableProxy(
                    address(baseStrategyImplementation),
                    address(mantleLayrProxyAdmin),
                    abi.encodeWithSelector(InvestmentStrategyBase.initialize.selector, mantleToken, mantleLayrPauserReg, mantleLayrReputedMultisig)
                )
            )
        );
        investmentManager.setInvestmentStrategy(mantleFirstStrat);

        // deploy upgradeable proxy that points to InvestmentStrategyBase implementation and initialize it
        mantleSencodStrat = InvestmentStrategyBase(
            address(
                new TransparentUpgradeableProxy(
                    address(baseStrategyImplementation),
                    address(mantleLayrProxyAdmin),
                    abi.encodeWithSelector(InvestmentStrategyBase.initialize.selector, mantleToken, mantleLayrPauserReg, mantleLayrReputedMultisig)
                )
            )
        );
        investmentManager.setInvestmentStrategy(mantleSencodStrat);

        vm.writeFile("data/investmentManager.addr", vm.toString(address(investmentManager)));
        vm.writeFile("data/delegation.addr", vm.toString(address(delegation)));
        vm.writeFile("data/slashRecorder.addr", vm.toString(address(slashRecorder)));
        vm.writeFile("data/mantle.addr", vm.toString(address(mantleToken)));
        vm.writeFile("data/mantleFirstStrat.addr", vm.toString(address(mantleFirstStrat)));
        vm.writeFile("data/mantleSencodStrat.addr", vm.toString(address(mantleSencodStrat)));
        vm.writeFile("data/rgPermission.addr", vm.toString(address(rgPermission)));

        vm.stopBroadcast();
    }
}
