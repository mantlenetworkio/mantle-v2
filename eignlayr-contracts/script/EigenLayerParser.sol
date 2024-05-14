// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "../src/contracts/interfaces/IEigenLayrDelegation.sol";
import "../src/contracts/core/EigenLayrDelegation.sol";

import "../src/contracts/core/InvestmentManager.sol";
import "../src/contracts/strategies/InvestmentStrategyBase.sol";

import "forge-std/Test.sol";
import "forge-std/Script.sol";
import "forge-std/StdJson.sol";

import "../src/contracts/interfaces/IRegistryPermission.sol";


contract EigenLayerParser is Script, DSTest {
    Vm cheats = Vm(HEVM_ADDRESS);

    uint256 numDis;
    uint256 numDln;
    uint256 numStaker;
    uint256 numCha;

    uint256 public constant eigenTotalSupply = 1000e18;
    EigenLayrDelegation public delegation;
    InvestmentManager public investmentManager;
    InvestmentStrategyBase public mantleFirstStrat;
    IERC20 public mantle;
    InvestmentStrategyBase public mantleSencodStrat;
    IRegistryPermission public rgPermission;

    string internal configJson;
    string internal addressJson;

    function parseEigenLayerParams() internal {
        configJson = vm.readFile("data/participants.json");
        numDis = stdJson.readUint(configJson, ".numDis");
        numDln = stdJson.readUint(configJson, ".numDln");
        numStaker = stdJson.readUint(configJson, ".numStaker");

        addressJson = vm.readFile("data/addresses.json");
        delegation = EigenLayrDelegation(stdJson.readAddress(addressJson, ".delegation"));
        investmentManager = InvestmentManager(stdJson.readAddress(addressJson, ".investmentManager"));
        mantle = IERC20(stdJson.readAddress(addressJson, ".mantle"));
        mantleFirstStrat = InvestmentStrategyBase(stdJson.readAddress(addressJson, ".mantleFirstStrat"));
        mantleSencodStrat = InvestmentStrategyBase(stdJson.readAddress(addressJson, ".mantleSencodStrat"));
        rgPermission = IRegistryPermission(stdJson.readAddress(addressJson, ".rgPermission"));
    }
}
