// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

library BridgeConstants {

    address internal constant L1_MNT = 0x6900000000000000000000000000000000000020;
    address internal constant L1_CROSSDOMAIN_MESSENGER = 0x6900000000000000000000000000000000000002;
    address internal constant OPTIMISM_PORTAL = 0x6900000000000000000000000000000000000001;




    uint32 constant ETH_DEPOSIT_TX = 0;
    uint32 constant MNT_DEPOSIT_TX = 1;

    uint32 constant ETH_WITHDRAWAL_TX = 3;
    uint32 constant MNT_WITHDRAWAL_TX = 4;
    uint32 constant ERC20_TX = 5;

    uint32 constant ERC721_TX = 6;
    uint256 constant ERC721_AMOUNT = 1;
}
