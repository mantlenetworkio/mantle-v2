// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "./EigenLayerParser.sol";

// forge script script/BecomeOperator.s.sol:BecomeOperator --rpc-url http://127.0.0.1  --private-key 6a6494edf0c00b3d0117f1635ad32a6005587cb6e9e808874da622f7b8925697 --broadcast -vvvv
contract BecomeOperator is Script, DSTest, EigenLayerParser {
    //performs basic deployment before each test
    function run() external {
        parseEigenLayerParams();
        emit log_address(msg.sender);
        vm.broadcast(msg.sender);
        delegation.registerAsOperator(msg.sender);
    }
}
