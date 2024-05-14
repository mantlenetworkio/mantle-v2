// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "./EigenLayerParser.sol";

contract DepositAndDelegate is Script, DSTest, EigenLayerParser {

    //performs basic deployment before each test
    function run() external {
        parseEigenLayerParams();

        uint256 mantleAmount;

        address dlnAddr;

        //get the corresponding dln
        //is there an easier way to do this?
        for (uint256 i = 0; i < numStaker; i++) {
            address stakerAddr =
                stdJson.readAddress(configJson, string.concat(".staker[", string.concat(vm.toString(i), "].address")));

            mantleAmount = vm.parseUint(stdJson.readString(configJson, string.concat(".staker[", string.concat(vm.toString(i), "].stake"))));

            if (stakerAddr == msg.sender) {
                dlnAddr =
                    stdJson.readAddress(configJson, string.concat(".dln[", string.concat(vm.toString(i), "].address")));
                    break;
            }
        }

        emit log("mantleAmount");
        emit log_uint(mantleAmount);

        vm.startBroadcast(msg.sender);
//        mantle.approve(address(investmentManager), mantleAmount);
//        investmentManager.depositIntoStrategy(mantleSencodStrat, mantle, mantleAmount);
        mantle.approve(address(investmentManager), mantleAmount);
        investmentManager.depositIntoStrategy(mantleFirstStrat, mantle, mantleAmount);
        delegation.delegateTo(dlnAddr);
        vm.stopBroadcast();
    }
}
