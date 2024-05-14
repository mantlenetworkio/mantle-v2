// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/utils/math/Math.sol";

import "../test/EigenLayrTestHelper.t.sol";

import "../contracts/libraries/BytesLib.sol";

import "./mocks/MiddlewareRegistryMock.sol";
import "./mocks/MiddlewareVoteWeigherMock.sol";
import "./mocks/ServiceManagerMock.sol";
import "./mocks/ServiceManagerMock.sol";

contract DelegationTests is EigenLayrTestHelper {
    using BytesLib for bytes;
    using Math for uint256;

    uint256 public PRIVATE_KEY = 420;

    uint32 serveUntil = 100;

    ServiceManagerMock public serviceManager;
    MiddlewareVoteWeigherMock public voteWeigher;
    MiddlewareVoteWeigherMock public voteWeigherImplementation;

    modifier fuzzedAmounts(uint256 ethAmount, uint256 eigenAmount){
        cheats.assume(ethAmount >= 0 && ethAmount <= 1e18);
        cheats.assume(eigenAmount >= 0 && eigenAmount <= 1e18);
        _;
    }

    function setUp() public virtual override {
        EigenLayrDeployer.setUp();

        initializeMiddlewares();
    }

    function initializeMiddlewares() public {
        serviceManager = new ServiceManagerMock(investmentManager);

        voteWeigher = MiddlewareVoteWeigherMock(
            address(new TransparentUpgradeableProxy(address(emptyContract), address(eigenLayrProxyAdmin), ""))
        );

        voteWeigherImplementation = new MiddlewareVoteWeigherMock(delegation, investmentManager, serviceManager);

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

            eigenLayrProxyAdmin.upgradeAndCall(
                TransparentUpgradeableProxy(payable(address(voteWeigher))),
                address(voteWeigherImplementation),
                abi.encodeWithSelector(MiddlewareVoteWeigherMock.initialize.selector, _quorumBips, ethStratsAndMultipliers, eigenStratsAndMultipliers)
            );
        }
    }

    // /// @notice testing if an operator can register to themselves.
    // function testSelfOperatorRegister() public {
    //     _testRegisterAdditionalOperator(0, serveUntil);
    // }

    // /// @notice testing if an operator can delegate to themselves.
    // /// @param sender is the address of the operator.
    // function testSelfOperatorDelegate(address sender) public {
    //     cheats.assume(sender != address(0));
    //     cheats.assume(sender != address(eigenLayrProxyAdmin));
    //     _testRegisterAsOperator(sender, sender);
    // }

    // function testTwoSelfOperatorsRegister() public {
    //     _testRegisterAdditionalOperator(0, serveUntil);
    //     _testRegisterAdditionalOperator(1, serveUntil);
    // }

    /// @notice registers a fixed address as a delegate, delegates to it from a second address,
    ///         and checks that the delegate's voteWeights increase properly
    /// @param operator is the operator being delegated to.
    /// @param staker is the staker delegating stake to the operator.
    function testDelegation(address operator, address staker, uint256 ethAmount, uint256 eigenAmount)
        public
        fuzzedAddress(operator)
        fuzzedAddress(staker)
        fuzzedAmounts(ethAmount, eigenAmount)
    {
        cheats.assume(staker != operator);

        _testDelegation(operator, staker, ethAmount, eigenAmount, voteWeigher);
    }

    // /// @notice tests delegation to EigenLayr via an ECDSA signatures - meta transactions are the future bby
    // /// @param operator is the operator being delegated to.
    // function testDelegateToBySignature(address operator, uint256 ethAmount, uint256 eigenAmount)
    //     public
    //     fuzzedAddress(operator)
    // {
    //     cheats.assume(ethAmount >= 0 && ethAmount <= 1e18);
    //     cheats.assume(eigenAmount >= 0 && eigenAmount <= 1e18);
    //     if (!delegation.isOperator(operator)) {
    //         _testRegisterAsOperator(operator, operator);
    //     }
    //     address staker = cheats.addr(PRIVATE_KEY);
    //     cheats.assume(staker != operator);

    //     //making additional deposits to the investment strategies
    //     assertTrue(delegation.isNotDelegated(staker) == true, "testDelegation: staker is not delegate");
    //     _testDepositWeth(staker, ethAmount);
    //     _testDepositEigen(staker, eigenAmount);


    //     uint256 nonceBefore = delegation.nonces(staker);

    //     bytes32 structHash = keccak256(abi.encode(delegation.DELEGATION_TYPEHASH(), staker, operator, nonceBefore, 0));
    //     bytes32 digestHash = keccak256(abi.encodePacked("\x19\x01", delegation.DOMAIN_SEPARATOR(), structHash));


    //     (uint8 v, bytes32 r, bytes32 s) = cheats.sign(PRIVATE_KEY, digestHash);

    //     bytes32 vs = getVSfromVandS(v, s);

    //     delegation.delegateToBySignature(staker, operator, 0, r, vs);
    //     assertTrue(delegation.isDelegated(staker) == true, "testDelegation: staker is not delegate");
    //     assertTrue(nonceBefore + 1 == delegation.nonces(staker), "nonce not incremented correctly");
    //     assertTrue(delegation.delegatedTo(staker) == operator, "staker delegated to wrong operator");
    // }

    // /// @notice tests delegation to EigenLayr via an ECDSA signatures with invalid signature
    // /// @param operator is the operator being delegated to.
    // function testDelegateToByInvalidSignature(
    //     address operator,
    //     uint256 ethAmount,
    //     uint256 eigenAmount,
    //     uint8 v,
    //     bytes32 r,
    //     bytes32 s
    // )
    //     public
    //     fuzzedAddress(operator)
    //     fuzzedAmounts(ethAmount, eigenAmount)
    // {
    //     if (!delegation.isOperator(operator)) {
    //         _testRegisterAsOperator(operator, operator);
    //     }
    //     address staker = cheats.addr(PRIVATE_KEY);
    //     cheats.assume(staker != operator);

    //     //making additional deposits to the investment strategies
    //     assertTrue(delegation.isNotDelegated(staker) == true, "testDelegation: staker is not delegate");
    //     _testDepositWeth(staker, ethAmount);
    //     _testDepositEigen(staker, eigenAmount);

    //     bytes32 vs = getVSfromVandS(v, s);

    //     cheats.expectRevert();
    //     delegation.delegateToBySignature(staker, operator, 0, r, vs);
    // }

    // /// @notice registers a fixed address as a delegate, delegates to it from a second address,
    // /// and checks that the delegate's voteWeights increase properly
    // /// @param operator is the operator being delegated to.
    // /// @param staker is the staker delegating stake to the operator.
    // function testDelegationMultipleStrategies(uint8 numStratsToAdd, address operator, address staker)
    //     public
    //     fuzzedAddress(operator)
    //     fuzzedAddress(staker)
    // {
    //     cheats.assume(staker != operator);

    //     cheats.assume(numStratsToAdd > 0 && numStratsToAdd <= 20);
    //     uint96 operatorEthWeightBefore = voteWeigher.weightOfOperator(operator, 0);
    //     uint96 operatorEigenWeightBefore = voteWeigher.weightOfOperator(operator, 1);
    //     _testRegisterAsOperator(operator, operator);
    //     _testDepositStrategies(staker, 1e18, numStratsToAdd);

    //     // add strategies to voteWeigher
    //     uint96 multiplier = 1e18;
    //     for (uint16 i = 0; i < numStratsToAdd; ++i) {
    //         VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[] memory ethStratsAndMultipliers =
    //         new VoteWeigherBaseStorage.StrategyAndWeightingMultiplier[](
    //                 1
    //             );
    //         ethStratsAndMultipliers[0].strategy = strategies[i];
    //         ethStratsAndMultipliers[0].multiplier = multiplier;
    //         cheats.startPrank(voteWeigher.serviceManager().owner());
    //         voteWeigher.addStrategiesConsideredAndMultipliers(0, ethStratsAndMultipliers);
    //         cheats.stopPrank();
    //     }

    //     _testDepositEigen(staker, 1e18);
    //     _testDelegateToOperator(staker, operator);
    //     uint96 operatorEthWeightAfter = voteWeigher.weightOfOperator(operator, 0);
    //     uint96 operatorEigenWeightAfter = voteWeigher.weightOfOperator(operator, 1);
    //     assertTrue(
    //         operatorEthWeightAfter > operatorEthWeightBefore, "testDelegation: operatorEthWeight did not increase!"
    //     );
    //     assertTrue(
    //         operatorEigenWeightAfter > operatorEigenWeightBefore, "testDelegation: operatorEthWeight did not increase!"
    //     );
    // }

    // /// @notice This function tests to ensure that a delegation contract
    // ///         cannot be intitialized multiple times
    // function testCannotInitMultipleTimesDelegation() public cannotReinit {
    //     //delegation has already been initialized in the Deployer test contract
    //     delegation.initialize(eigenLayrPauserReg, address(this));
    // }

    /// @notice This function tests to ensure that a you can't register as a delegate multiple times
    /// @param operator is the operator being delegated to.
    function testRegisterAsOperatorMultipleTimes(address operator) public fuzzedAddress(operator) {
        _testRegisterAsOperator(operator, operator);
        cheats.expectRevert(bytes("EigenLayrDelegation.registerAsOperator: operator has already registered"));
        _testRegisterAsOperator(operator, operator);
    }

    function testRegisterAsOperatorWithPermisssion(address operator) public fuzzedAddress(operator) {
        
        _testAddOperatorRegisterPermission(operator, permission);
        _testRegisterAsOperator(operator, operator);
    }

    // /// @notice This function tests to ensure that a staker cannot delegate to an unregistered operator
    // /// @param delegate is the unregistered operator
    // function testDelegationToUnregisteredDelegate(address delegate) public fuzzedAddress(delegate) {
    //     //deposit into 1 strategy for getOperatorAddress(1), who is delegating to the unregistered operator
    //     _testDepositStrategies(getOperatorAddress(1), 1e18, 1);
    //     _testDepositEigen(getOperatorAddress(1), 1e18);

    //     cheats.expectRevert(bytes("EigenLayrDelegation._delegate: operator has not yet registered as a delegate"));
    //     cheats.startPrank(getOperatorAddress(1));
    //     delegation.delegateTo(delegate);
    //     cheats.stopPrank();
    // }

    // function _testRegisterAdditionalOperator(uint256 index, uint32 _serveUntil) internal {
    //     address sender = getOperatorAddress(index);

    //     //register as both ETH and EIGEN operator
    //     uint256 wethToDeposit = 1e18;
    //     uint256 eigenToDeposit = 1e10;
    //     _testDepositWeth(sender, wethToDeposit);
    //     _testDepositEigen(sender, eigenToDeposit);
    //     _testRegisterAsOperator(sender, sender);

    //     cheats.startPrank(sender);

    //     //whitelist the serviceManager to slash the operator
    //     slasher.optIntoSlashing(address(serviceManager));

    //     voteWeigher.registerOperator(sender, _serveUntil);

    //     cheats.stopPrank();
    // }
}
