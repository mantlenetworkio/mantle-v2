// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

/**
 * @title Burn
 * @notice Utilities for burning stuff.
 */
library Burn {
    /**
     * Burns a given amount of MNT.
     *
     * @param _amount Amount of MNT to burn.
     */
    function mnt(uint256 _amount) internal {
        new Burner{ value: _amount }();
    }

    /**
     * Burns a given amount of gas.
     *
     * @param _amount Amount of gas to consume.
     */
    function gas(uint256 _amount) internal view {
        uint256 i = 0;
        uint256 initialGas = gasleft();
        while (initialGas - gasleft() < _amount) {
            ++i;
        }
    }
}

/**
 * @title Burner
 * @notice Burner self-destructs on creation and sends all MNT to itself, removing all MNT given to
 *         the contract from the circulating supply. Self-destructing is the only way to remove MNT
 *         from the circulating supply.
 */
contract Burner {
    constructor() payable {
        selfdestruct(payable(address(this)));
    }
}
