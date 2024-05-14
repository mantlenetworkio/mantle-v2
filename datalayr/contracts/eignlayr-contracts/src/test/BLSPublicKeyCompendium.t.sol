// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "./EigenLayrTestHelper.t.sol";

contract BLSPublicKeyCompendiumTests is EigenLayrTestHelper {

    function testRegisterBLSPublicKey(address operator, uint256 s, BN254.G1Point memory rPoint, BN254.G1Point memory pubkeyG1, BN254.G2Point memory pubkeyG2) public {
        
        cheats.expectRevert(bytes("BLSPublicKeyCompendium.registerBLSPublicKey: Operator does not permission to register bls public key"));
        bLSPC.registerBLSPublicKey(s, rPoint, pubkeyG1, pubkeyG2);

        // Input parameters are too complex and not suitable for fuzz testing
        // _testAddOperatorRegisterPermission(operator, permission);

        // cheats.prank(operator);
        // bLSPC.registerBLSPublicKey(s, rPoint, pubkeyG1, pubkeyG2);
    }
}