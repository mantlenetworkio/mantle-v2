// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { OptimismPortal } from "../L1/OptimismPortal.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title PortalSender
 * @notice The PortalSender is a simple intermediate contract that will transfer the balance of the
 *         L1StandardBridge to the OptimismPortal during the Bedrock migration.
 */
contract PortalSender {
    /**
     * @notice Address of the OptimismPortal contract.
     */
    OptimismPortal public immutable PORTAL;

    /**
     * @param _portal Address of the OptimismPortal contract.
     */
    constructor(OptimismPortal _portal) {
        PORTAL = _portal;
    }

    /**
     * @notice Sends balance of this contract to the OptimismPortal.
                on the Mantle Mainnet, this function will donate ETH and MNT
     */
    function donate() external {
        uint256 totalAmount = IERC20(PORTAL.L1_MNT_ADDRESS()).balanceOf(address(this));
        bool succ = IERC20(PORTAL.L1_MNT_ADDRESS()).transfer(address(PORTAL),totalAmount);
        require(succ,"donate mnt failed");
        PORTAL.donateETH{ value: address(this).balance }();
    }
}
