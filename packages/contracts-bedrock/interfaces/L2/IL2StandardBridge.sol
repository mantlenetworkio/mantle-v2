// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IL2StandardBridge {
    event DepositFinalized(
        address indexed l1Token,
        address indexed l2Token,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );
    event WithdrawalInitiated(
        address indexed l1Token,
        address indexed l2Token,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    receive() external payable;

    function L1_MNT_ADDRESS() external view returns (address);
    function l1TokenBridge() external view returns (address);
    function version() external view returns (string memory);
    function withdraw(
        address _l2Token,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _extraData
    )
        external
        payable;
    function withdrawTo(
        address _l2Token,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _extraData
    )
        external
        payable;

    function __constructor__(address payable _otherBridge, address _l1mnt) external;
}
