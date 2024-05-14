// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@eigenlayer/contracts/interfaces/IServiceManager.sol";
import "@eigenlayer/contracts/interfaces/IQuorumRegistry.sol";
import "@eigenlayer/contracts/interfaces/IInvestmentManager.sol";

import "forge-std/Test.sol";

contract EigenDARegistryMock is IQuorumRegistry, DSTest{
    IServiceManager public immutable serviceManager;
    IInvestmentManager public immutable investmentManager;


    constructor(
        IServiceManager _serviceManager,
        IInvestmentManager _investmentManager
    ){
        serviceManager = _serviceManager;
        investmentManager = _investmentManager;
    }

    function registerOperator(
        address operator,
        uint32 serveUntil
    ) public {
//        require(slasher.canSlash(operator, address(serviceManager)), "Not opted into slashing");
//        serviceManager.recordFirstStakeUpdate(operator, serveUntil);
    }

    function deregisterOperator(
        address operator,
        uint256 startIndex
    ) public {
//        uint32 latestTime = serviceManager.latestTime();
//        serviceManager.recordLastStakeUpdateAndRevokeSlashingAbility(operator, latestTime);
    }

    function propagateStakeUpdate(address operator, uint32 blockNumber, uint256 prevElement) external {
//        uint32 serveUntil = serviceManager.latestTime();
//        serviceManager.recordStakeUpdate(operator, blockNumber, serveUntil, prevElement);
    }

     function isActiveOperator(address operator) external pure returns (bool) {
        if (operator != address(0)){
            return true;
        } else {
            return false;
        }
     }

    function getLengthOfTotalStakeHistory() external view returns (uint256){}

    function getTotalStakeFromIndex(uint256 index) external view returns (OperatorStake memory){}

    /// @notice Returns the stored pubkeyHash for the specified `operator`.
    function getOperatorPubkeyHash(address operator) external view returns (bytes32){}

    /// @notice Returns task number from when `operator` has been registered.
    function getFromTaskNumberForOperator(address operator) external view returns (uint32){}

    function getStakeFromPubkeyHashAndIndex(bytes32 pubkeyHash, uint256 index) external view returns (OperatorStake memory){}

    function getOperatorIndex(address operator, uint32 blockNumber, uint32 index) external view returns (uint32){}

    function getTotalOperators(uint32 blockNumber, uint32 index) external view returns (uint32){}

    function numOperators() external view returns (uint32){}

    function operatorStakes(address operator) external view returns (uint96, uint96){}

    function checkOperatorActiveAtBlockNumber(
        address,
        uint256,
        uint256
        ) external pure returns (bool) {
        return true;
    }

    function checkOperatorInactiveAtBlockNumber(
        address,
        uint256,
        uint256
        ) external pure returns (bool) {
            return true;
        }

    /// @notice Returns the stake amounts from the latest entry in `totalStakeHistory`.
    function totalStake() external view returns (uint96, uint96){}
}
