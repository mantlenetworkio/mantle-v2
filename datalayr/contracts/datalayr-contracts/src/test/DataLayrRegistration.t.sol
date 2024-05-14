// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@eigenlayer/contracts/libraries/BytesLib.sol";

import "./DataLayrTestHelper.t.sol";

contract DataLayrRegistrationTests is DataLayrTestHelper {
    using BytesLib for bytes;

    /// @notice This test ensures that the optimistic flow for BLS registration
    ///         works as intended by checking storage updates in the registration contracts.
    function testBLSRegistration(
        uint8 operatorIndex,
        uint256 ethAmount,
        uint256 eigenAmount
    ) fuzzedOperatorIndex(operatorIndex) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);
        uint8 operatorType = 3;

        (
            uint256 amountEthStaked,
            uint256 amountEigenStaked
        ) = _testInitiateDelegation(
                operatorIndex,
                eigenAmount,
                ethAmount
            );

        _testRegisterBLSPubKey(operatorIndex);


        bytes32 hashofPk = BN254.hashG1Point(getOperatorPubkeyG1(operatorIndex));
        require(pubkeyCompendium.operatorToPubkeyHash(getOperatorAddress(operatorIndex)) == hashofPk, "hash not stored correctly");
        require(pubkeyCompendium.pubkeyHashToOperator(hashofPk) == getOperatorAddress(operatorIndex), "hash not stored correctly");

        {
            uint96 ethStakedBefore = dlReg.getTotalStakeFromIndex(dlReg.getLengthOfTotalStakeHistory()-1).firstQuorumStake;
            uint96 eigenStakedBefore = dlReg.getTotalStakeFromIndex(dlReg.getLengthOfTotalStakeHistory()-1).secondQuorumStake;

            _testRegisterOperatorWithDataLayr(
                operatorIndex,
                operatorType,
                testSocket
            );

            uint256 numOperators = dlReg.numOperators();
            require(dlReg.operatorList(numOperators-1) == getOperatorAddress(operatorIndex), "operatorList not updated");


            uint96 ethStakedAfter = dlReg.getTotalStakeFromIndex(dlReg.getLengthOfTotalStakeHistory()-1).firstQuorumStake;
            uint96 eigenStakedAfter = dlReg.getTotalStakeFromIndex(dlReg.getLengthOfTotalStakeHistory()-1).secondQuorumStake;


            require(ethStakedAfter - ethStakedBefore == amountEthStaked, "eth quorum staked value not updated correctly");
            require(eigenStakedAfter - eigenStakedBefore == amountEigenStaked, "eigen quorum staked value not updated correctly");
        }
    }

    /// @notice Tests that registering the same public key twice reverts appropriately.
    function testRegisterPublicKeyTwice(uint8 operatorIndex) fuzzedOperatorIndex(operatorIndex) public {
        cheats.startPrank(getOperatorAddress(operatorIndex));
        //try to register the same pubkey twice
        (uint256 s, BN254.G1Point memory rPoint) = getOperatorSchnorrSignature(operatorIndex);
        pubkeyCompendium.registerBLSPublicKey(s, rPoint, getOperatorPubkeyG1(operatorIndex), getOperatorPubkeyG2(operatorIndex));
        cheats.expectRevert(
            "BLSPublicKeyCompendium.registerBLSPublicKey: operator already registered pubkey"
        );
        (s, rPoint) = getOperatorSchnorrSignature(operatorIndex);
        pubkeyCompendium.registerBLSPublicKey(s, rPoint, getOperatorPubkeyG1(operatorIndex), getOperatorPubkeyG2(operatorIndex));
    }

    /// @notice Tests that re-registering while an msg.sender is actively registered reverts.
    function testRegisterWhileAlreadyActive(
        uint8 operatorIndex,
        uint256 ethAmount,
        uint256 eigenAmount
    ) fuzzedOperatorIndex(operatorIndex) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);

        uint8 operatorType = 3;
        _testInitiateDelegation(
            operatorIndex,
            eigenAmount,
            ethAmount
        );
        _testRegisterBLSPubKey(
            operatorIndex
        );
        _testRegisterOperatorWithDataLayr(
            operatorIndex,
            operatorType,
            testSocket
        );
        cheats.startPrank(getOperatorAddress(operatorIndex));

        //try to register after already registered
        cheats.expectRevert(
            "RegistryBase._registrationStakeEvaluation: Operator is already registered"
        );
        dlReg.registerOperator(
            3,
            getOperatorPubkeyG1(operatorIndex),
            testSocket
        );
        cheats.stopPrank();
    }

    /// @notice Test that when operator tries to register with DataLayr with a public key
    ///         that they haven't registered in the BLSPublicKeyCompendium, it fails
    function testOperatorDoesNotOwnPublicKey(
        uint8 operatorIndex,
        uint256 ethAmount,
        uint256 eigenAmount
    ) fuzzedOperatorIndex(operatorIndex) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);

        uint8 operatorType = 3;
        _testInitiateDelegation(
            operatorIndex,
            eigenAmount,
            ethAmount
        );
        //registering the operator without having registered their BLS public key
        cheats.expectRevert(bytes("BLSRegistry._registerOperator: operator does not own pubkey"));

        _testRegisterOperatorWithDataLayr(
            operatorIndex,
            operatorType,
            testSocket
        );
    }
    /// @notice Tests that registering without having delegated in any quorum reverts
    function testRegisterForDataLayrWithNeitherQuorum(
        uint8 operatorIndex,
        uint256 ethAmount,
        uint256 eigenAmount
    ) fuzzedOperatorIndex(operatorIndex) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);
        uint8 noQuorumOperatorType = 0;

        _testInitiateDelegation(
            operatorIndex,
            eigenAmount,
            ethAmount
        );
        _testRegisterBLSPubKey(
            operatorIndex
        );
        cheats.expectRevert(bytes("RegistryBase._registrationStakeEvaluation: Must register as at least one type of validator"));
        _testRegisterOperatorWithDataLayr(
            operatorIndex,
            noQuorumOperatorType,
            testSocket
        );
    }
    /// @notice Tests that registering without adequate quorum stake reverts
    function testRegisterWithoutEnoughQuorumStake(
        uint8 operatorIndex
    ) fuzzedOperatorIndex(operatorIndex) public {
        address operator = getOperatorAddress(operatorIndex);
        cheats.startPrank(operator);
        delegation.registerAsOperator(operator);
        cheats.stopPrank();

        _testRegisterBLSPubKey(
            operatorIndex
        );

        uint8 operatorType = 1;
        cheats.expectRevert(bytes("RegistryBase._registrationStakeEvaluation: Must register as at least one type of validator"));
        _testRegisterOperatorWithDataLayr(operatorIndex, operatorType, testSocket);

        operatorType = 2;
        cheats.expectRevert(bytes("RegistryBase._registrationStakeEvaluation: Must register as at least one type of validator"));
        _testRegisterOperatorWithDataLayr(operatorIndex, operatorType, testSocket);

        operatorType = 3;
        cheats.expectRevert(bytes("RegistryBase._registrationStakeEvaluation: Must register as at least one type of validator"));
        _testRegisterOperatorWithDataLayr(operatorIndex, operatorType, testSocket);
    }

    /// @notice Tests for registering without having opted into slashing.
    function testRegisterWithoutSlashingOptIn(
        uint8 operatorIndex,
        uint256 ethAmount,
        uint256 eigenAmount
     ) fuzzedOperatorIndex(operatorIndex) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e18);

        uint8 operatorType = 3;

        _testInitiateDelegation(
            operatorIndex,
            eigenAmount,
            ethAmount
        );

        cheats.startPrank(getOperatorAddress(operatorIndex));
        (uint256 s, BN254.G1Point memory rPoint) = getOperatorSchnorrSignature(operatorIndex);
        pubkeyCompendium.registerBLSPublicKey(s, rPoint, getOperatorPubkeyG1(operatorIndex), getOperatorPubkeyG2(operatorIndex));
        cheats.stopPrank();

        cheats.expectRevert(bytes("RegistryBase._addRegistrant: operator must be opted into slashing by the serviceManager"));
        _testRegisterOperatorWithDataLayr(
            operatorIndex,
            operatorType,
            testSocket
        );
     }
}
