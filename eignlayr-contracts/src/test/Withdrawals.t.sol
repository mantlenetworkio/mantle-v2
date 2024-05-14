// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/utils/math/Math.sol";

import "../contracts/libraries/BytesLib.sol";

import "./mocks/MiddlewareRegistryMock.sol";
import "./mocks/ServiceManagerMock.sol";
import "./Delegation.t.sol";

contract WithdrawalTests is DelegationTests {

    // packed info used to help handle stack-too-deep errors
    struct DataForTestWithdrawal {
        IInvestmentStrategy[] delegatorStrategies;
        uint256[] delegatorShares;
        IInvestmentManager.WithdrawerAndNonce withdrawerAndNonce;
    }

    MiddlewareRegistryMock public generalReg1;
    ServiceManagerMock public generalServiceManager1;

    MiddlewareRegistryMock public generalReg2;
    ServiceManagerMock public generalServiceManager2;

    function initializeGeneralMiddlewares() public {
        generalServiceManager1 = new ServiceManagerMock(investmentManager);

        generalReg1 = new MiddlewareRegistryMock(
             generalServiceManager1,
             investmentManager
        );

        generalServiceManager2 = new ServiceManagerMock(investmentManager);

        generalReg2 = new MiddlewareRegistryMock(
             generalServiceManager2,
             investmentManager
        );
    }

    //This function helps with stack too deep issues with "testWithdrawal" test
    function testWithdrawalWrapper(
            address operator,
            address depositor,
            address withdrawer,
            uint256 ethAmount,
            uint256 eigenAmount,
            bool withdrawAsTokens,
            bool RANDAO
        )
            public
            fuzzedAddress(operator)
            fuzzedAddress(depositor)
            fuzzedAddress(withdrawer)
        {
            cheats.assume(depositor != operator);
            cheats.assume(ethAmount <= 1e18);
            cheats.assume(eigenAmount <= 1e18);
            cheats.assume(ethAmount > 0);
            cheats.assume(eigenAmount > 0);

            initializeGeneralMiddlewares();

            if(RANDAO){
                _testWithdrawalAndDeregistration(operator, depositor, withdrawer, ethAmount, eigenAmount, withdrawAsTokens);
            }
            else{
                _testWithdrawalWithStakeUpdate(operator, depositor, withdrawer, ethAmount, eigenAmount, withdrawAsTokens);
            }

        }

    /// @notice test staker's ability to undelegate/withdraw from an operator.
    /// @param operator is the operator being delegated to.
    /// @param depositor is the staker delegating stake to the operator.
    function _testWithdrawalAndDeregistration(
            address operator,
            address depositor,
            address withdrawer,
            uint256 ethAmount,
            uint256 eigenAmount,
            bool withdrawAsTokens
        )
            internal
        {

        testDelegation(operator, depositor, ethAmount, eigenAmount);

        cheats.startPrank(operator);
        cheats.stopPrank();

        generalReg1.registerOperator(operator, uint32(block.timestamp) + 3 days);

        address delegatedTo = delegation.delegatedTo(depositor);

        // packed data structure to deal with stack-too-deep issues
        DataForTestWithdrawal memory dataForTestWithdrawal;

        // scoped block to deal with stack-too-deep issues
        {
            //delegator-specific information
            (IInvestmentStrategy[] memory delegatorStrategies, uint256[] memory delegatorShares) =
                investmentManager.getDeposits(depositor);
            dataForTestWithdrawal.delegatorStrategies = delegatorStrategies;
            dataForTestWithdrawal.delegatorShares = delegatorShares;

            IInvestmentManager.WithdrawerAndNonce memory withdrawerAndNonce =
                IInvestmentManager.WithdrawerAndNonce({
                    withdrawer: withdrawer,
                    // harcoded nonce value
                    nonce: 0
                }
            );
            dataForTestWithdrawal.withdrawerAndNonce = withdrawerAndNonce;
        }

        uint256[] memory strategyIndexes = new uint256[](2);
        IERC20[] memory tokensArray = new IERC20[](2);
        {
            // hardcoded values
            strategyIndexes[0] = 0;
            strategyIndexes[1] = 0;
            tokensArray[0] = weth;
            tokensArray[1] = eigenToken;
        }

        cheats.warp(uint32(block.timestamp) + 1 days);
        cheats.roll(uint32(block.timestamp) + 1 days);

        uint32 queuedWithdrawalBlock = uint32(block.number);

        //now withdrawal block time is before deregistration
        cheats.warp(uint32(block.timestamp) + 2 days);
        cheats.roll(uint32(block.timestamp) + 2 days);

        generalReg1.deregisterOperator(operator);
    }

    /// @notice test staker's ability to undelegate/withdraw from an operator.
    /// @param operator is the operator being delegated to.
    /// @param depositor is the staker delegating stake to the operator.
    function _testWithdrawalWithStakeUpdate(
            address operator,
            address depositor,
            address withdrawer,
            uint256 ethAmount,
            uint256 eigenAmount,
            bool withdrawAsTokens
        )
            public
        {
        testDelegation(operator, depositor, ethAmount, eigenAmount);

        cheats.startPrank(operator);
        cheats.stopPrank();

        // emit log_named_uint("Linked list element 1", uint256(uint160(address(generalServiceManager1))));
        // emit log_named_uint("Linked list element 2", uint256(uint160(address(generalServiceManager2))));
        // emit log("________________________________________________________________");
        generalReg1.registerOperator(operator, uint32(block.timestamp) + 5 days);
        // emit log_named_uint("Middleware 1 Update Block", uint32(block.number));

        cheats.warp(uint32(block.timestamp) + 1 days);
        cheats.roll(uint32(block.number) + 1);

        generalReg2.registerOperator(operator, uint32(block.timestamp) + 5 days);
        // emit log_named_uint("Middleware 2 Update Block", uint32(block.number));

        address delegatedTo = delegation.delegatedTo(depositor);

        // packed data structure to deal with stack-too-deep issues
        DataForTestWithdrawal memory dataForTestWithdrawal;

        // scoped block to deal with stack-too-deep issues
        {
            //delegator-specific information
            (IInvestmentStrategy[] memory delegatorStrategies, uint256[] memory delegatorShares) =
                investmentManager.getDeposits(depositor);
            dataForTestWithdrawal.delegatorStrategies = delegatorStrategies;
            dataForTestWithdrawal.delegatorShares = delegatorShares;

            IInvestmentManager.WithdrawerAndNonce memory withdrawerAndNonce =
                IInvestmentManager.WithdrawerAndNonce({
                    withdrawer: withdrawer,
                    // harcoded nonce value
                    nonce: 0
                }
            );
            dataForTestWithdrawal.withdrawerAndNonce = withdrawerAndNonce;
        }

        uint256[] memory strategyIndexes = new uint256[](2);
        IERC20[] memory tokensArray = new IERC20[](2);
        {
            // hardcoded values
            strategyIndexes[0] = 0;
            strategyIndexes[1] = 0;
            tokensArray[0] = weth;
            tokensArray[1] = eigenToken;
        }

        cheats.warp(uint32(block.timestamp) + 1 days);
        cheats.roll(uint32(block.number) + 1);
    }
    // @notice This function tests to ensure that a delegator can re-delegate to an operator after undelegating.
    // @param operator is the operator being delegated to.
    // @param staker is the staker delegating stake to the operator.
    function testRedelegateAfterWithdrawal(
            address operator,
            address depositor,
            address withdrawer,
            uint256 ethAmount,
            uint256 eigenAmount,
            bool withdrawAsShares
        )
            public
            fuzzedAddress(operator)
            fuzzedAddress(depositor)
            fuzzedAddress(withdrawer)
        {
        cheats.assume(depositor != operator);
        //this function performs delegation and subsequent withdrawal
        testWithdrawalWrapper(operator, depositor, withdrawer, ethAmount, eigenAmount, withdrawAsShares, true);
        //warps past fraudproof time interval
        cheats.warp(block.timestamp + 7 days + 1);
        testDelegation(operator, depositor, ethAmount, eigenAmount);
    }
}
