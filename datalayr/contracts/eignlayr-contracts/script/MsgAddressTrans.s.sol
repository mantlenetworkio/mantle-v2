pragma solidity ^0.8.0;

import "forge-std/console.sol";

contract MsgAddressTrans {
    uint160 constant offset = uint160(0x1111000000000000000000000000000000001111);

    function applyL1ToL2Alias(address l1Address) public pure returns (address l2Address) {
        unchecked {
            l2Address = address(uint160(l1Address) + offset);
        }
    }

    function undoL1ToL2Alias(address l2Address) public pure returns (address l1Address) {
        unchecked {
            l1Address = address(uint160(l2Address) - offset);
        }
    }

    function run() external {
        address l1CmsgTrans = applyL1ToL2Alias(address(0xbE59bda17Fa8786072AfFAa082C17d6dCD99dbA0));
        console.log(l1CmsgTrans);
        address originAddress = undoL1ToL2Alias(l1CmsgTrans);
        console.log(originAddress);
    }
}
