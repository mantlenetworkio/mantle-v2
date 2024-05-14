// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@eigenlayer/contracts/middleware/BLSPublicKeyCompendium.sol";
import "@eigenlayer/contracts/middleware/BLSRegistry.sol";
import "@eigenlayer/test/EigenLayrDeployer.t.sol";

import "../contracts/core/DataLayrServiceManager.sol";
import "../contracts/core/DataLayrChallengeUtils.sol";
import "../contracts/core/DataLayrChallenge.sol";

import "../contracts/libraries/DataStoreUtils.sol";

import "./utils/Signatures.sol";


contract DataLayrDeployer is EigenLayrDeployer, SignatureUtils {
    using BytesLib for bytes;

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

    bytes[] registrationData;

    address storer = address(420);

    address public dataLayrReputedMultisig = address(this);
    address public dataLayrTeamMultisig = address(this);

    string testSocket = "255.255.255.255";

    struct NonSignerPK {
        uint256 x;
        uint256 y;
        uint32 stakeIndex;
    }

    struct RegistrantAPKG2 {
        uint256 apk0;
        uint256 apk1;
        uint256 apk2;
        uint256 apk3;
    }

    struct RegistrantAPKG1 {
        uint256 apk0;
        uint256 apk1;
    }

    struct SignerAggSig{
        uint256 sigma0;
        uint256 sigma1;
    }

    modifier fuzzedAddress(address addr) virtual override {
        cheats.assume(addr != address(0));
        cheats.assume(addr != address(eigenLayrProxyAdmin));
        cheats.assume(addr != address(investmentManager));
        cheats.assume(addr != address(dataLayrProxyAdmin));
        cheats.assume(addr != dlsm.owner());
        _;
    }

    modifier fuzzedOperatorIndex(uint8 operatorIndex) {
        cheats.assume(operatorIndex < getNumOperators());
        _;
    }

    //performs basic deployment before each test
    function setUp() public virtual override {
        // do the EigenLayer deployment
        _deployEigenLayrContracts();
        // deploy all the DataLayr contracts
        _deployDataLayrContracts();
        _setUpTestingParameters();
    }

    // deploy all the DataLayr contracts. Relies on many EL contracts having already been deployed.
    function _deployDataLayrContracts() internal {
        // deploy proxy admin for ability to upgrade proxy contracts
        dataLayrProxyAdmin = new ProxyAdmin();

        // deploy pauser registry
        dataLayrPauserReg = new PauserRegistry(pauser, unpauser);

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
            weth,
            dlldc,
            // TODO: fix this
            DataLayrBombVerifier(address(0)),
            rgPermission
        );
        pubkeyCompendiumImplementation = new BLSPublicKeyCompendium(rgPermissionImplementation);
        dlldcImplementation = new DataLayrChallenge(dlsm, dlReg, challengeUtils);
        {
            uint8 _NUMBER_OF_QUORUMS = 2;
            dlRegImplementation = new BLSRegistry(
                investmentManager,
                dlsm,
                _NUMBER_OF_QUORUMS,
                pubkeyCompendium,
                rgPermissionImplementation,
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
            VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[] memory ethStratsAndMultipliers =
                new VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[](1);
            ethStratsAndMultipliers[0].strategy = wethStrat;
            ethStratsAndMultipliers[0].multiplier = multiplier;
            VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[] memory eigenStratsAndMultipliers =
                new VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[](1);
            eigenStratsAndMultipliers[0].strategy = eigenStrat;
            eigenStratsAndMultipliers[0].multiplier = multiplier;

            dataLayrProxyAdmin.upgradeAndCall(
                TransparentUpgradeableProxy(payable(address(dlReg))),
                address(dlRegImplementation),
                abi.encodeWithSelector(BLSRegistry.initialize.selector, _quorumBips, msg.sender, ethStratsAndMultipliers, eigenStratsAndMultipliers)
            );
        }
    }

    function _setUpTestingParameters() internal {
        registrationData.push(
            hex"075dcd2e66658b1f4f61aa809f001bb79324b91089af99b9a78e27284e8c73130d884d46e54bf17137028ddc3fd38d5b89686b7c433099b28149f9c8f771c8431f5bda9b7d94f525e0f9b667127df9fa884e9917453db7fe3119820b994b5e5d2428c354c0019c338afd3994e186d7d443ec1d8abab2e2d1e19bac019ee295f202a45cfe62ffb797ab25355a7f54788277f7fd9fda544ac6a7e38623d75fdd001074a61258b73d4773971a8073f04a6dd072409bea915d4ece0583c65f09fbfe"
        );
        registrationData.push(
            hex"2669082021fd1033646a940aabe3f459e7b7a808d959c392af45c91b3fe064960bce92bfb1a54bc1af73b41a1edb13bd9e5006471c5d4708f77ea530f1045b7a0914646c43c0b404345c7864daa76091996c36227ac5b2ad5a7468ab49ebaf7b13357d53c87adfee0aa3b2c7dbca5d00660c4c5ed1acbeebb4c9202101dab4f00953b9e7b44ec5991070966ed70c1cd37b03b06797059b6828b0a2abc1d5210c134a2cc96c98c4ed34e2c7399695d25c0c2dfce27e0885ad13b979eb1c465b99"
        );
        registrationData.push(
            hex"142b758de8ad4c74e8167d71b3667cf75e982f006480ecafdde2a403748e7d1b2dd77f6eac473a31fddba53321584cd0aa296f14d14f098093937a5b93dd61c90cc3e0a7657c894d178a7ff41ae51b5ccc4c697684c599015b003aceeb2fec641863a130465043a63a1acf5494ee76895779044613264c5f65a106834b6615902def894e6c296e5b789398128a3b8f05054314ee82739e8e51cea9e4432a000d028d664abd661c75fe7ed0506c347f3b94d782d82e2259c7ecb39c9796922b04"
        );
        registrationData.push(
            hex"2af2ac3833ce14949c9ef3fbccf620e3a13c9df686687634f9546a76ec5899f7219bfb0cf2f2817525cd89082302218c3cf83b3beae6c4fbe25ae4a790e948d307d64b5418c89567b5956590d6232c4ed95afd9d06d5a13b1f9c0c306a9260fe04783304a0c560710cb4f1bdc8096e7a67e39be589513dc644845b2e66fc19dd24fddaf89dd8e1f6ed4d5d8750fed28b4159442ea7edd367c9335bb07a3a00ea00bbc408f2a7336e2ac8694db6df7603708293aac6ee702cdfc0eefb32c37b27"
        );
        registrationData.push(
            hex"2c63a558d2384cf3f387db39c48c3b72595ef13adbc3ca7689bc90bae7e4ab060620e82d1bb6c52977529ece1fe1d31b0521492a06c661e06363b3be8306acd10746c80e9dacb5731c65232cf5fb5a2450e4f2e44d44fbc9d6cbf19dd30db776226488c51bfacbf7704d12065eb3ad1b9a707a4f61d41effdcb2ced3e01c42691f7631be59f69c691c082e7d192e4c4bbedab7c296ff6fc879e6f5511f3fc9a316f8e0f3a57a58ee42165206ec70f94ee1e80a41907f3ec36fb8cdeaaa08ca52"
        );
        registrationData.push(
            hex"0076c0c034a6916e712bb41ed97530c4475c78c89f916137511d03ee94b670691a904a8de426166c9a7e6e3e36260973db56b218336dc89c68e2710026abe9e61612d3f5da47c52b552d66322623d688f5046baa625e4f66556cafed25c61980017458bcef061aafd36e998f0f5958439f175df8ffd3a286bc4986eafdb6d47015e186650610a8d2d336913f53adff244280748c91ffc37d21179f2051deef662ce36aca626ad16812b5a8ffe3bb8c258154b7e962a90e72bd4732f21f808645"
        );
        registrationData.push(
            hex"1fb489ea26c1b85899bad2104702946ef256a7e59f26080bfddce2a64e94e3991947cd387f975963abc04838968f3eb128263b73c57c6820107395eca138fd98100bcb4ba69885f5020187520c35df6ff5b991b01bab7b83ad63c23af7e03b0c1efe7165964b7e66443b25b76fe6717739760afae192948aed7ae74f81564255264d1fac1a8f1c5d6f2d8e7e38ebdfc59a512c7281b5abfb727aa883a688f4381a970b882e097f1c1c754c9fd8ebc503a30488ffe821ac98bf79062f9b1d81c5"
        );
        registrationData.push(
            hex"0e7fc7b5bca43de3fab4acc5a7a014bb9bb5aff171cb26ae31bffe2bc529db0f1269f9809e4069bddf06aaf88187192e241fb817a6c8bbb5aff3836a0520e6b61aca04d4cc4f83d755ac2e9e083197afee1ea77d42e9429fa4b3fb64276f78001e7951e39e5de9c4c89e41fc0fbcf8f59438e85a60d1ac40293ab862f1b4c3bd0e225ae617a66cfc67ae42283156ff19878b9857cce60a2ae322075579cc8ed207d30ecd2feac39c5e2a7cacf6fe38c78a41b1b97313060b41a41b499477148c"
        );
        registrationData.push(
            hex"18b1f796356a80ea2cc1c0e23a3e7331a97a417473cf83a5f6942ecf9a84cc351a187ceef1a2436db814c6d5a83b16b6dd48f69b23d07f7e3544cf9f3a4edd8a031b8f1c6711edff8267eb49c6a9ecd2de39eaea18621db1f601186b6c8b56ee1a7bb20411a152aaac50010240dad6f82a7dc818fe6565db4132350d69eaeec62a47a927850ea2e09f6d0757d3f3201000eb58c24a9fa0160076433be84960ef031aeb05ae95495541e544f3a8345331f016ed542d05b64ca5076112faeb9b1a"
        );
        registrationData.push(
            hex"15ba1ac04f35335cfd1c9c1fcaac012871e3543bb7876b38be193e3f07592aab0323619b00d87f3c03d4bab25c91b8bc4b7aa96818930f2b4684ae8f6e92464b30298b441eaadcfb3b86e0b3f0e41250060dbb89e34c2d67acef7ed9a2590db42108f4f14af5ff87b2b9b7d766c4be119b790f34c9b3b1a62d16f6a95935d2e00463223946956732c65085bd6b2f3651944757099d6f643c0370fac983c27f1e0dc2b54a54fde7495e81d43c6346549cf824fb45ecd18f77d4537e8fba7e7e0c"
        );
        registrationData.push(
            hex"0822ccd871333690ea42c6e7fa1b594c785d8296fc8bacc8a10ddda8f3378ebb0d68db879257ec3f74d4fc1cffd17a9f1b6db08b7c421753dbce0751d6d7d23a07873fdb87a38f72a537da1cc20b48d1186594430718e15ec5e195ab3c65f8102f6a351c01b3cfc217c9ab936382a53b9a350851ecbaf43e6a0f086bf8ec395409fe90efaeae3703fbddaf8f331451d3dd3d138fce006af813b579d8c67313d71353b0fd3e02d50c77889d7095d09eb4874a7425604f20c3d7b619bf5efe3274"
        );
        registrationData.push(
            hex"1a6962a7170cf4ea2ad4bd0bf9a95c4e6bf96e9302e345b9d12bfbf6fb86dc911733c8198257dc9003ba0163d217b48fdb14e6ce91691242064ae21d821980481ad1e21ac4adc2eebad1e279e490b307aafafcf43a3e63decb19f7dac7d5a26c1fa208243839cf96ee3218652239dc06119770cebe08776c1bd92af9626f04d025cc5bff6c03978aab365592207f4e24fe1cce9eece22e86c84535ce3b0851732fb29f8709e77f2c38ee09f4eb3143fa17eb2381785485fa7990ee0b161367e7"
        );
        registrationData.push(
            hex"13185695a1abc17847ce6a90edc65eb04c0ebd218156f122ef689674e82ebb331ad5be86a500c6b0b490cbe70610356448aa2b06442f364b138fe7cd0df5efa9294cbc1ccb8c6afdbc05938f368521351328222ac99388e7a26c4f9d51ad1024042a5a5286bbcc22f94e95555be8a193731c2c265b64aa25fde8a047202a6d9501b635713a31a9322e81ad50f9331775856e610bfdb5546aadeb681143dc015023b5d07f7004ad42a5a2c74fd1c87991326b7575a75e73a347a7c59741d21db5"
        );
        registrationData.push(
            hex"1db8d40c46e9992c0e020568b3f1c02fa4aef44c5db1610325093280218f2ab014c3ab56f0d82ad9ff275fae94e51a17c613302e5aa2f2de7001ae181727f8d4053c3d457ad36273361e3b35d02cea6c93879a55f0d086a77e58dc0d5805c6b428fc018be860797143a2b0296ed35113addbf3c0e8aaf6ea93c0acb3db78bae1216edaa7fff2998dfd2adee5620745512c2faca1f547b996892eef199fe8bfd515696133c1920636012e494103e3c592283583296d73924bbacba7d299ca0e7d"
        );
        registrationData.push(
            hex"16bb52aa5a1e51cf22ac1926d02e95fdeb411ad48b567337d4c4d5138e84bd5516a6e1e18fb4cd148bd6b7abd46a5d6c54444c11ba5a208b6a8230e86cc8f80828427fd024e29e9a31945cd91433fde23fc9656a44424794a9dfdcafa9275baa06d5b28737bc0a5c21279b3c5309e35287cd72deb204abf6d6c91a0e0b38d0a41ae35db861ea707fc72c6b7756a6139e8cccf15392e59297c21af365de013b4312caa1e05d5aac7c5513fff386248f1955298f11e0e165ed9a20c9beefe2f8a0"
        );
    }

    function calculateFee(uint256 totalBytes, uint256 feePerBytePerTime, uint256 duration)
        internal
        view
        returns (uint256)
    {
        return dlsm.calculateFee(totalBytes, feePerBytePerTime, uint32(duration));
    }

    function _testSig() internal view {
        uint256[12] memory input;
        //1d9b51a4ffb5b3f402748854ea5bbb8025324782062324e99bedcdc2cec4102f
        //000000000004
        //00000918
        //00000007
        //00000000
        //00000003
        //0d8c5e0a5954cbbc30123d0990c7643b1e8b43278457d3a89de59cfc620ac48a
        //068a2ec2615a4064fd820f759d6030475fed69925655aae8a463e72b53f697e9
        //014d5b9af4f3e72635652fe695fdb3c46ee3e5142820b228bf9564fdef30bd92
        //0238c50db7b36820321b2e25700486c18e5750dea646d266870ec1be812456fa
        //1e041e0df4821a4b7668999e4381cca9c015916f033512ca0829179c639f285c
        //1a2ebe9095bed1d16f938c00d283c3a08462c7dc168a590ffa8ce192e05996ab

        (input[0], input[1]) = BN254.hashToG1(0x1d9b51a4ffb5b3f402748854ea5bbb8025324782062324e99bedcdc2cec4102f);
        input[3] = uint256(0x0d8c5e0a5954cbbc30123d0990c7643b1e8b43278457d3a89de59cfc620ac48a);
        input[2] = uint256(0x068a2ec2615a4064fd820f759d6030475fed69925655aae8a463e72b53f697e9);
        input[5] = uint256(0x014d5b9af4f3e72635652fe695fdb3c46ee3e5142820b228bf9564fdef30bd92);
        input[4] = uint256(0x0238c50db7b36820321b2e25700486c18e5750dea646d266870ec1be812456fa);
        input[6] = uint256(0x1e041e0df4821a4b7668999e4381cca9c015916f033512ca0829179c639f285c);
        input[7] = uint256(0x1a2ebe9095bed1d16f938c00d283c3a08462c7dc168a590ffa8ce192e05996ab);
        // insert negated coordinates of the generator for G2
        input[8] = BN254.nG2x1;
        input[9] = BN254.nG2x0;
        input[10] = BN254.nG2y1;
        input[11] = BN254.nG2y0;

        assembly {
            // check the pairing; if incorrect, revert
            if iszero(
                // staticcall address 8 (ecPairing precompile), forward all gas, send 384 bytes (0x180 in hex) = 12 (32-byte) inputs.
                // store the return data in input[11] (352 bytes / '0x160' in hex), and copy only 32 bytes of return data (since precompile returns boolean)
                staticcall(not(0), 0x08, input, 0x180, add(input, 0x160), 0x20)
            ) { revert(0, 0) }
        }

        // check that the provided signature is correct
        require(input[11] == 1, "BLSSignatureChecker.checkSignatures: Pairing unsuccessful");
    }

    function _testBLSPairing() internal {
            uint256[12] memory input;

            uint256 sigmaX = 18033935401377046968253993369420882761639101147199761382164100964672839397476;
            uint256 sigmaY = 1296611607075364961854999662642612779184492063389140410860059877500726169961;

            bytes32 msgHash = 0x536ea2113b06bc65d2d6310b51424f268f1b3155e1fe82cbc90d9b8712d14a0a;
            (uint256 msgHashX, uint256 msgHashY) = BN254.hashToG1(msgHash);

            emit log_named_uint("msgHashX", msgHashX);
            emit log_named_uint("msgHashY", msgHashY);

            input[0] = sigmaX;
            input[1] = sigmaY;
            input[2] = BN254.nG2x1;
            input[3] = BN254.nG2x0;
            input[4] = BN254.nG2y1;
            input[5] = BN254.nG2y0;

            input[6] = msgHashX;
            input[7] = msgHashY;
            // insert negated coordinates of the generator for G2
            input[8] = 2548741418739206695596229529236657819733103689248810431091319058064536250278;
            input[9] = 17890127137359027111482509378509337249586291091685072336190236845225812702820;
            input[10] = 12498134380415317036640719391312524222291167329168408451224344109201613968031;
            input[11] = 18577908915005185161399472001797886901908616360139528062172259974922524099491;

            assembly {
                // check the pairing; if incorrect, revert
                if iszero(
                    staticcall(sub(gas(), 2000), 8, input, 0x180, input, 0x20)
                ) {
                    revert(0, 0)
                }
            }

            require(
                input[0] == 1,
                "BLSSignatureChecker.checkSignatures: Pairing unsuccessful"
            );
        }

    function _testVKPairing() internal view {
        uint256[12] memory input;

        uint256 pkg1X = 11746114415387181186350609321861313487282937637157292915572974055983718048797;
        uint256 pkg1Y = 6199836912972052411871307285755230980030751238632264470990041456311661808876;


        input[0] = pkg1X;
        input[1] = pkg1Y;
        input[2] = BN254.nG2x1;
        input[3] = BN254.nG2x0;
        input[4] = BN254.nG2y1;
        input[5] = BN254.nG2y0;

        input[6] = 1;
        input[7] = 2;
        // insert negated coordinates of the generator for G2
        input[8] = 2548741418739206695596229529236657819733103689248810431091319058064536250278;
        input[9] = 17890127137359027111482509378509337249586291091685072336190236845225812702820;
        input[10] = 12498134380415317036640719391312524222291167329168408451224344109201613968031;
        input[11] = 18577908915005185161399472001797886901908616360139528062172259974922524099491;

        assembly {
            // check the pairing; if incorrect, revert
            if iszero(
                staticcall(sub(gas(), 2000), 8, input, 0x180, input, 0x20)
            ) { revert(0, 0) }
        }

        require(input[0] == 1, "BLSSignatureChecker.checkSignatures: Pairing unsuccessful");
    }
}
