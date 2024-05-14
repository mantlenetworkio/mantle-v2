// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "./EigenLayrTestHelper.t.sol";
import "../contracts/core/InvestmentManagerStorage.sol";

contract InvestmentTests is EigenLayrTestHelper {
    /**
     * @notice Verifies that it is possible to deposit WETH
     * @param amountToDeposit Fuzzed input for amount of WETH to deposit
     */
    function testWethDeposit(uint256 amountToDeposit) public returns (uint256 amountDeposited) {
        return _testDepositWeth(getOperatorAddress(0), amountToDeposit);
    }

// TODO: reimplement with queued withdrawals
/*
    ///@notice This test verifies that it is possible to withdraw WETH after depositing it
    ///@param amountToDeposit The amount of WETH to try depositing
    ///@param amountToWithdraw The amount of shares to try withdrawing
    function testWethWithdrawal(uint96 amountToDeposit, uint96 amountToWithdraw) public {
        // want to deposit at least 1 wei
        cheats.assume(amountToDeposit > 0);
        // want to withdraw at least 1 wei
        cheats.assume(amountToWithdraw > 0);
        // cannot withdraw more than we deposit
        cheats.assume(amountToWithdraw <= amountToDeposit);
        // hard-coded inputs
        address sender = getOperatorAddress(0);
        uint256 strategyIndex = 0;
        _testDepositToStrategy(sender, amountToDeposit, weth, wethStrat);
        _testWithdrawFromStrategy(sender, strategyIndex, amountToWithdraw, weth, wethStrat);
    }
*/
// TODO: reimplement with queued withdrawals
/*
    /**
     * @notice Verifies that a strategy gets removed from the dynamic array 'investorStrats' when the user no longer has any shares in the strategy
     * @param amountToDeposit Fuzzed input for the amount deposited into the strategy, prior to withdrawing all shares
     */
/*
    function testRemovalOfStrategyOnWithdrawal(uint96 amountToDeposit) public {
        // hard-coded inputs
        IInvestmentStrategy _strat = wethStrat;
        IERC20 underlyingToken = weth;
        address sender = getOperatorAddress(0);

        _testDepositToStrategy(sender, amountToDeposit, underlyingToken, _strat);
        uint256 investorStratsLengthBefore = investmentManager.investorStratsLength(sender);
        uint256 investorSharesBefore = investmentManager.investorStratShares(sender, _strat);
        _testWithdrawFromStrategy(sender, 0, investorSharesBefore, underlyingToken, _strat);
        uint256 investorSharesAfter = investmentManager.investorStratShares(sender, _strat);
        uint256 investorStratsLengthAfter = investmentManager.investorStratsLength(sender);
        assertEq(investorSharesAfter, 0, "testRemovalOfStrategyOnWithdrawal: did not remove all shares!");
        assertEq(
            investorStratsLengthBefore - investorStratsLengthAfter,
            1,
            "testRemovalOfStrategyOnWithdrawal: strategy not removed from dynamic array when it should be"
        );
    }
*/


    /// @notice deploys 'numStratsToAdd' strategies using '_testAddStrategy' and then deposits '1e18' to each of them from 'getOperatorAddress(0)'
    /// @param numStratsToAdd is the number of strategies being added and deposited into
    function testDepositStrategies(uint8 numStratsToAdd) public {
        _testDepositStrategies(getOperatorAddress(0), 1e18, numStratsToAdd);
    }

    /// @notice Verifies that it is possible to deposit eigen.
    /// @param eigenToDeposit is amount of eigen to deposit into the eigen strategy
    function testDepositEigen(uint96 eigenToDeposit) public {
        // sanity check for inputs; keeps fuzzed tests from failing
        cheats.assume(eigenToDeposit < eigenTotalSupply);
        _testDepositEigen(getOperatorAddress(0), eigenToDeposit);
    }

    /**
     * @notice Tries to deposit an unsupported token into an `InvestmentStrategyBase` contract by calling `investmentManager.depositIntoStrategy`.
     * Verifies that reversion occurs correctly.
     */
    function testDepositUnsupportedToken() public {
        IERC20 token = new ERC20PresetFixedSupply(
            "badToken",
            "BADTOKEN",
            100,
            address(this)
        );
        token.approve(address(investmentManager), type(uint256).max);
        cheats.expectRevert(bytes("InvestmentStrategyBase.deposit: Can only deposit underlyingToken"));
        investmentManager.depositIntoStrategy(wethStrat, token, 10);
    }

    /**
     * @notice Tries to deposit into an unsupported strategy by calling `investmentManager.depositIntoStrategy`.
     * Verifies that reversion occurs correctly.
     */
    function testDepositNonexistentStrategy(address nonexistentStrategy) public fuzzedAddress(nonexistentStrategy) {
        // assume that the fuzzed address is not already a contract!
        uint256 size;
        assembly {
            size := extcodesize(nonexistentStrategy)
        }
        cheats.assume(size == 0);
        // check against calls from precompile addresses -- was getting fuzzy failures from this
        cheats.assume(uint160(nonexistentStrategy) > 9);

        // harcoded input
        uint256 testDepositAmount = 10;

        IERC20 token = new ERC20PresetFixedSupply(
            "badToken",
            "BADTOKEN",
            100,
            address(this)
        );
        token.approve(address(investmentManager), type(uint256).max);
        cheats.expectRevert();
        investmentManager.depositIntoStrategy(IInvestmentStrategy(nonexistentStrategy), token, testDepositAmount);
    }

    /// @notice verify that trying to deposit an amount of zero will correctly revert
    function testRevertOnZeroDeposit() public {
        cheats.expectRevert(bytes("InvestmentManager._addShares: shares should not be zero!"));
        investmentManager.depositIntoStrategy(wethStrat, weth, 0);
        cheats.stopPrank();
    }
}
