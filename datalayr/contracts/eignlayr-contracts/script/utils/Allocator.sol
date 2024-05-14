// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract Allocator {

    function allocate(IERC20 token, address[] memory recipients, uint256 amount) public {
        token.transferFrom(msg.sender, address(this), recipients.length * amount);
        for (uint i = 0; i < recipients.length; i++) {
            token.transfer(recipients[i], amount);
        }
    }

    function allocateArray(IERC20 token, address[] memory recipients, uint256[] memory amounts, uint256 totalAmount) public {
        token.transferFrom(msg.sender, address(this), totalAmount);
        for (uint i = 0; i < recipients.length; i++) {
            token.transfer(recipients[i], amounts[i]);
        }
    }
}
