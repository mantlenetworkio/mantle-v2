// SPDX-License-Verifier: UNLICENSED
pragma solidity ^0.8.9;

import "@eigenlayer/contracts/libraries/BytesLib.sol";

import "./DataLayrTestHelper.t.sol";

contract DataLayrDeregistrationTests is DataLayrTestHelper {
    using BytesLib for bytes;

    /**
    *   @notice Tests that optimistically deregistering works as intended, i.e., 
    *   the aggregate public key is updated correctly, and all storage is correctly updated.
    */
    function testBLSDeregistration(
        uint8 operatorIndex,
        uint256 ethAmount, 
        uint256 eigenAmount
    ) public fuzzedOperatorIndex(operatorIndex) {
        //TODO: probably a stronger test would be to register a few operators and then ensure that apk is updated correctly
        (uint256 x, uint256 y) = dlReg.apk();
        BN254.G1Point memory apk = BN254.G1Point({X: x, Y: y});
        bytes32 prevAPKHash = BN254.hashG1Point(apk);

        _BLSRegistration(operatorIndex, ethAmount, eigenAmount);

        BN254.G1Point memory operatorPubkey = getOperatorPubkeyG1(operatorIndex);

        bytes32 pubkeyHash = BN254.hashG1Point(operatorPubkey);    

        _testDeregisterOperatorWithDataLayr(operatorIndex, uint8(dlReg.numOperators()-1), getOperatorPubkeyG1(operatorIndex));

        (,uint32 nextUpdateBlockNumber,uint96 firstQuorumStake, uint96 secondQuorumStake) = dlReg.pubkeyHashToStakeHistory(pubkeyHash, dlReg.getStakeHistoryLength(pubkeyHash)-1);
        require( nextUpdateBlockNumber == 0, "Stake history not updated correctly");
        require( firstQuorumStake == 0, "Stake history not updated correctly");
        require( secondQuorumStake == 0, "Stake history not updated correctly");

        bytes32 currAPKHash = dlReg.apkHashes(dlReg.getApkUpdatesLength()-1);
        require(currAPKHash == prevAPKHash, "aggregate public key has not been updated correctly following deregistration");
    }

    /**
    *   @notice Tests that deregistering with an incorrect public key
    *           reverts.
    */
    function testMismatchedPubkeyHashAndProvidedPubkeyHash(
        uint8 operatorIndex,
        uint256 ethAmount, 
        uint256 eigenAmount,
        BN254.G1Point memory pkToRemove
    ) public fuzzedOperatorIndex(operatorIndex) {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);
        cheats.assume(pkToRemove.X != getOperatorPubkeyG1(operatorIndex).X);

    
        _BLSRegistration(operatorIndex, ethAmount, eigenAmount);
        uint8 operatorListIndex = uint8(dlReg.numOperators()-1);
        cheats.expectRevert(bytes("BLSRegistry._deregisterOperator: pubkey input does not match stored pubkeyHash"));
        _testDeregisterOperatorWithDataLayr(operatorIndex, operatorListIndex, pkToRemove);
    }

    /**
    *   @notice Tests that deregistering an operator who has already 
    *           been deregistered/was never registered reverts
    */
    function testDeregisteringAlreadyDeregisteredOperator(
        uint8 operatorIndex,
        uint256 ethAmount, 
        uint256 eigenAmount
    ) public fuzzedOperatorIndex(operatorIndex) {

        _BLSRegistration(operatorIndex, ethAmount, eigenAmount);


        uint8 operatorListIndex = uint8(dlReg.numOperators());                  
        _testDeregisterOperatorWithDataLayr(operatorIndex, operatorListIndex-1, getOperatorPubkeyG1(operatorIndex));
        
        cheats.expectRevert(bytes("RegistryBase._deregistrationCheck: Operator is not registered"));
        _testDeregisterOperatorWithDataLayr(operatorIndex, operatorListIndex-1, getOperatorPubkeyG1(operatorIndex));

    }


    /// @notice Helper function that performs registration 
    function _BLSRegistration(
        uint8 operatorIndex,
        uint256 ethAmount, 
        uint256 eigenAmount
    ) internal fuzzedOperatorIndex(operatorIndex) {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);
        
        uint8 operatorType = 3;
        _testInitiateDelegation(
            operatorIndex,
            eigenAmount,
            ethAmount
        );
        _testRegisterBLSPubKey(operatorIndex);
        _testRegisterOperatorWithDataLayr(
            operatorIndex,
            operatorType,
            testSocket
        );
    }

}