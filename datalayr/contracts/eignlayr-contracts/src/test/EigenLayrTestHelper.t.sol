// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "../contracts/libraries/BytesLib.sol";
import "../test/EigenLayrDeployer.t.sol";


contract EigenLayrTestHelper is EigenLayrDeployer {
    using BytesLib for bytes;

    uint8 durationToInit = 2;
    uint256 public SECP256K1N_MODULUS = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141;
    uint256 public SECP256K1N_MODULUS_HALF = 0x7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0;

    uint256[] sharesBefore;
    uint256[] balanceBefore;
    uint256[] priorTotalShares;
    uint256[] strategyTokenBalance;

    event AddOperatorRegisterPermission(address operator, bool status);

    function _testInitiateDelegation(
        uint8 operatorIndex,
        uint256 amountEigenToDeposit,
        uint256 amountEthToDeposit
    )
        public returns (uint256 amountEthStaked, uint256 amountEigenStaked)
    {

        address operator = getOperatorAddress(operatorIndex);

        //setting up operator's delegation terms
        _testRegisterAsOperator(operator, operator);

        for (uint256 i; i < stakers.length; i++) {
            //initialize weth, eigen and eth balances for staker
            eigenToken.transfer(stakers[i], amountEigenToDeposit);
            weth.transfer(stakers[i], amountEthToDeposit);

            //deposit staker's eigen and weth into investment manager
            _testDepositEigen(stakers[i], amountEigenToDeposit);
            _testDepositWeth(stakers[i], amountEthToDeposit);

            //delegate the staker's deposits to operator
            uint256 operatorEigenSharesBefore = delegation.operatorShares(operator, eigenStrat);
            uint256 operatorWETHSharesBefore = delegation.operatorShares(operator, wethStrat);
            _testDelegateToOperator(stakers[i], operator);
            //verify that `increaseOperatorShares` worked
            assertTrue(
                delegation.operatorShares(operator, eigenStrat) - operatorEigenSharesBefore == amountEigenToDeposit
            );
            assertTrue(delegation.operatorShares(operator, wethStrat) - operatorWETHSharesBefore == amountEthToDeposit);

        }
        amountEthStaked += delegation.operatorShares(operator, wethStrat);
        amountEigenStaked += delegation.operatorShares(operator, eigenStrat);

        return (amountEthStaked, amountEigenStaked);
    }

    // simply tries to register 'sender' as an operator, setting their 'DelegationTerms' contract in EigenLayrDelegation to 'dt'
    // verifies that the storage of EigenLayrDelegation contract is updated appropriately
    function _testRegisterAsOperator(address sender, address rewardAddress) internal {

        cheats.expectRevert(bytes("EigenLayrDelegation.registerAsOperator: Operator does not permission to register as operator"));
        cheats.prank(address(0));
        delegation.registerAsOperator(rewardAddress);

        (IInvestmentStrategy[] memory delegateStrategies, uint256[] memory delegateShares) =
            investmentManager.getDeposits(sender);

        uint256 numStrats = delegateShares.length;
        uint256[] memory inititalSharesInStrats = new uint256[](numStrats);
        for (uint256 i = 0; i < numStrats; ++i) {
            inititalSharesInStrats[i] = delegation.operatorShares(rewardAddress, delegateStrategies[i]);
        }

        cheats.startPrank(sender);

        cheats.expectRevert(bytes("EigenLayrDelegation._delegate: operator has not yet registered as a delegate"));
        delegation.registerAsOperator(address(0));

        delegation.registerAsOperator(rewardAddress);

        cheats.expectRevert(bytes("EigenLayrDelegation.registerAsOperator: operator has already registered"));
        delegation.registerAsOperator(rewardAddress);


        assertTrue(delegation.isOperator(sender), "testRegisterAsOperator: sender is not a operator");

        assertTrue(
            sender == rewardAddress, "_testRegisterAsOperator: delegationTerms not set appropriately"
        );
        assertTrue(
            delegation.operatorReceiverRewardAddress(sender) == rewardAddress, 
            "operatorReceiverRewardAddress: delegationTerms not set appropriately"
        );

        assertTrue(delegation.isDelegated(sender), "_testRegisterAsOperator: sender not marked as actively delegated");
        cheats.stopPrank();


        assertTrue(
            delegation.delegatedTo(sender) == rewardAddress,
            "_testDelegateToOperator: delegated address not set appropriately"
        );
        assertTrue(
            delegation.isDelegated(sender),
            "_testDelegateToOperator: delegated status not set appropriately"
        );

        (delegateStrategies, delegateShares) = investmentManager.getDeposits(sender);
        numStrats = delegateShares.length;
        for (uint256 i = 0; i < numStrats; ++i) {
            uint256 operatorSharesBefore = inititalSharesInStrats[i];
            uint256 operatorSharesAfter = delegation.operatorShares(rewardAddress, delegateStrategies[i]);
            assertTrue(
                operatorSharesAfter == (operatorSharesBefore + delegateShares[i]),
                "_testDelegateToOperator: delegatedShares not increased correctly"
            );
        }
    }

    function _testAddOperatorRegisterPermission(address operator, address permission) internal {

        cheats.expectRevert(bytes("Only the permission person can do this action"));
        rgPermission.addOperatorRegisterPermission(operator);

        vm.expectEmit(true, true, false, true, address(rgPermission));
        emit AddOperatorRegisterPermission(operator, true);
        cheats.startPrank(permission);
        rgPermission.addOperatorRegisterPermission(operator);

        assertTrue(rgPermission.operatorRegisterPermission(operator), "operatorRegisterPermission: operator haven't permission");
        assertTrue(rgPermission.getOperatorRegisterPermission(operator), "getOperatorRegisterPermission: operator haven't permission");

        cheats.stopPrank();
    }

    /**
     * @notice Deposits `amountToDeposit` of WETH from address `sender` into `wethStrat`.
     * @param sender The address to spoof calls from using `cheats.startPrank(sender)`
     * @param amountToDeposit Amount of WETH that is first *transferred from this contract to `sender`* and then deposited by `sender` into `stratToDepositTo`
     */
    function _testDepositWeth(address sender, uint256 amountToDeposit) internal returns (uint256 amountDeposited) {
        cheats.assume(amountToDeposit <= wethInitialSupply);
        amountDeposited = _testDepositToStrategy(sender, amountToDeposit, weth, wethStrat);
    }

    /**
     * @notice Deposits `amountToDeposit` of EIGEN from address `sender` into `eigenStrat`.
     * @param sender The address to spoof calls from using `cheats.startPrank(sender)`
     * @param amountToDeposit Amount of EIGEN that is first *transferred from this contract to `sender`* and then deposited by `sender` into `stratToDepositTo`
     */
    function _testDepositEigen(address sender, uint256 amountToDeposit) internal returns (uint256 amountDeposited) {
        cheats.assume(amountToDeposit <= eigenTotalSupply);
        amountDeposited = _testDepositToStrategy(sender, amountToDeposit, eigenToken, eigenStrat);
    }

    /**
     * @notice Deposits `amountToDeposit` of `underlyingToken` from address `sender` into `stratToDepositTo`.
     * *If*  `sender` has zero shares prior to deposit, *then* checks that `stratToDepositTo` is correctly added to their `investorStrats` array.
     *
     * @param sender The address to spoof calls from using `cheats.startPrank(sender)`
     * @param amountToDeposit Amount of WETH that is first *transferred from this contract to `sender`* and then deposited by `sender` into `stratToDepositTo`
     */
    function _testDepositToStrategy(
        address sender,
        uint256 amountToDeposit,
        IERC20 underlyingToken,
        IInvestmentStrategy stratToDepositTo
    )
        internal
        returns (uint256 amountDeposited)
    {
        // deposits will revert when amountToDeposit is 0
        cheats.assume(amountToDeposit > 0);

        uint256 operatorSharesBefore = investmentManager.investorStratShares(sender, stratToDepositTo);
        // assumes this contract already has the underlying token!
        uint256 contractBalance = underlyingToken.balanceOf(address(this));
        // logging and error for misusing this function (see assumption above)
        if (amountToDeposit > contractBalance) {
            emit log("amountToDeposit > contractBalance");
            emit log_named_uint("amountToDeposit is", amountToDeposit);
            emit log_named_uint("while contractBalance is", contractBalance);
            revert("_testDepositToStrategy failure");
        } else {
            underlyingToken.transfer(sender, amountToDeposit);
            cheats.startPrank(sender);
            underlyingToken.approve(address(investmentManager), type(uint256).max);
            investmentManager.depositIntoStrategy(stratToDepositTo, underlyingToken, amountToDeposit);
            amountDeposited = amountToDeposit;

            //check if depositor has never used this strat, that it is added correctly to investorStrats array.
            if (operatorSharesBefore == 0) {
                // check that strategy is appropriately added to dynamic array of all of sender's strategies
                assertTrue(
                    investmentManager.investorStrats(sender, investmentManager.investorStratsLength(sender) - 1)
                        == stratToDepositTo,
                    "_depositToStrategy: investorStrats array updated incorrectly"
                );
            }

            //in this case, since shares never grow, the shares should just match the deposited amount
            assertEq(
                investmentManager.investorStratShares(sender, stratToDepositTo) - operatorSharesBefore,
                amountDeposited,
                "_depositToStrategy: shares should match deposit"
            );
        }
        cheats.stopPrank();
    }

    // tries to delegate from 'staker' to 'operator'
    // verifies that:
    //                  staker has at least some shares
    //                  delegatedShares update correctly for 'operator'
    //                  delegated status is updated correctly for 'staker'
    function _testDelegateToOperator(address staker, address operator) internal {
        //staker-specific information
        (IInvestmentStrategy[] memory delegateStrategies, uint256[] memory delegateShares) =
            investmentManager.getDeposits(staker);

        uint256 numStrats = delegateShares.length;
        assertTrue(numStrats != 0, "_testDelegateToOperator: delegating from address with no investments");
        uint256[] memory inititalSharesInStrats = new uint256[](numStrats);
        for (uint256 i = 0; i < numStrats; ++i) {
            inititalSharesInStrats[i] = delegation.operatorShares(operator, delegateStrategies[i]);
        }

        cheats.startPrank(staker);
        delegation.delegateTo(operator);
        cheats.stopPrank();

        assertTrue(
            delegation.delegatedTo(staker) == operator,
            "_testDelegateToOperator: delegated address not set appropriately"
        );
        assertTrue(
            delegation.isDelegated(staker),
            "_testDelegateToOperator: delegated status not set appropriately"
        );

        for (uint256 i = 0; i < numStrats; ++i) {
            uint256 operatorSharesBefore = inititalSharesInStrats[i];
            uint256 operatorSharesAfter = delegation.operatorShares(operator, delegateStrategies[i]);
            assertTrue(
                operatorSharesAfter == (operatorSharesBefore + delegateShares[i]),
                "_testDelegateToOperator: delegatedShares not increased correctly"
            );
        }
    }

    /// deploys 'numStratsToAdd' strategies contracts and initializes them to treat `underlyingToken` as their underlying token
    /// and then deposits 'amountToDeposit' to each of them from 'sender'
    function _testDepositStrategies(address sender, uint256 amountToDeposit, uint8 numStratsToAdd) internal {
        // hard-coded input
        IERC20 underlyingToken = weth;

        cheats.assume(numStratsToAdd > 0 && numStratsToAdd <= 20);
        IInvestmentStrategy[] memory stratsToDepositTo = new IInvestmentStrategy[](
                numStratsToAdd
            );
        for (uint8 i = 0; i < numStratsToAdd; ++i) {
            stratsToDepositTo[i] = InvestmentStrategyBase(
                address(
                    new TransparentUpgradeableProxy(
                        address(baseStrategyImplementation),
                        address(eigenLayrProxyAdmin),
                    abi.encodeWithSelector(InvestmentStrategyBase.initialize.selector, underlyingToken, eigenLayrPauserReg)
                    )
                )
            );
            _testDepositToStrategy(sender, amountToDeposit, weth, InvestmentStrategyBase(address(stratsToDepositTo[i])));
        }
        for (uint8 i = 0; i < numStratsToAdd; ++i) {
            // check that strategy is appropriately added to dynamic array of all of sender's strategies
            assertTrue(
                investmentManager.investorStrats(sender, i) == stratsToDepositTo[i],
                "investorStrats array updated incorrectly"
            );

            // TODO: perhaps remove this is we can. seems brittle if we don't track the number of strategies somewhere
            //store strategy in mapping of strategies
            strategies[i] = IInvestmentStrategy(address(stratsToDepositTo[i]));
        }
    }

    /**
    * combines V and S into VS - if S is greater than SECP256K1N_MODULUS_HALF, then we
    * get the modulus, so that the leading bit of s is always 0.  Then we set the leading
    * bit to be either 0 or 1 based on the value of v, which is either 27 or 28
    */
    function getVSfromVandS(uint8 v, bytes32 s) internal view returns(bytes32){
        if (uint256(s) > SECP256K1N_MODULUS_HALF) {
            s = bytes32(SECP256K1N_MODULUS - uint256(s));
        }

        bytes32 vs = s;
        if(v == 28){
            vs = bytes32(uint256(s) ^ (1 << 255));
        }

        return vs;
    }

    /// @notice registers a fixed address as an operator, delegates to it from a second address,
    ///         and checks that the operator's voteWeights increase properly
    /// @param operator is the operator being delegated to.
    /// @param staker is the staker delegating stake to the operator.
    /// @param voteWeigher is the VoteWeigher-type contract to consult for stake weight changes
    function _testDelegation(address operator, address staker, uint256 ethAmount, uint256 eigenAmount, IVoteWeigher voteWeigher)
        internal
    {
        if (!delegation.isOperator(operator)) {
            _testRegisterAsOperator(operator, operator);
        }

        uint256[3] memory amountsBefore;
        amountsBefore[0] = voteWeigher.weightOfOperator(operator, 0);
        amountsBefore[1] = voteWeigher.weightOfOperator(operator, 1);
        amountsBefore[2] = delegation.operatorShares(operator, wethStrat);

        //making additional deposits to the investment strategies
        assertTrue(delegation.isNotDelegated(staker) == true, "testDelegation: staker is not delegate");
        _testDepositWeth(staker, ethAmount);
        _testDepositEigen(staker, eigenAmount);
        _testDelegateToOperator(staker, operator);
        assertTrue(delegation.isDelegated(staker) == true, "testDelegation: staker is not delegate");

        (IInvestmentStrategy[] memory updatedStrategies, uint256[] memory updatedShares) =
            investmentManager.getDeposits(staker);

        {
            uint256 stakerEthWeight = investmentManager.investorStratShares(staker, updatedStrategies[0]);
            uint256 stakerEigenWeight = investmentManager.investorStratShares(staker, updatedStrategies[1]);

            uint256 operatorEthWeightAfter = voteWeigher.weightOfOperator(operator, 0);
            uint256 operatorEigenWeightAfter = voteWeigher.weightOfOperator(operator, 1);

            assertTrue(
                operatorEthWeightAfter - amountsBefore[0] == stakerEthWeight,
                "testDelegation: operatorEthWeight did not increment by the right amount"
            );
            assertTrue(
                operatorEigenWeightAfter - amountsBefore[1] == stakerEigenWeight,
                "Eigen weights did not increment by the right amount"
            );
        }
        {
            IInvestmentStrategy _strat = wethStrat;
            // IInvestmentStrategy _strat = investmentManager.investorStrats(staker, 0);
            assertTrue(address(_strat) != address(0), "investorStrats not updated correctly");

            assertTrue(
                delegation.operatorShares(operator, _strat) - updatedShares[0] == amountsBefore[2],
                "ETH operatorShares not updated correctly"
            );
        }
    }
}

