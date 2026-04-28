// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IFeeVault {
    event Withdrawal(uint256 value, address to, address from);

    receive() external payable;

    function MIN_WITHDRAWAL_AMOUNT() external view returns (uint256);
    function RECIPIENT() external view returns (address);
    function totalProcessed() external view returns (uint256);
    function withdraw() external;
    function version() external view returns (string memory);

    function __constructor__(address _recipient) external;
}
