// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "forge-std/Test.sol";
import "forge-std/Script.sol";
import "forge-std/StdJson.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

contract SignatureUtils is Test {
    //numSigners => array of signatures for 5 datastores
    mapping(uint256 => uint256[]) signatures;
    

    string internal signatureJson;

    function signaturePrefix(uint256 numSigners) public returns(string memory) {
        return string.concat(".signatures[", string.concat(vm.toString(numSigners), "]."));
    }

    //returns aggPK.X0, aggPK.X1, aggPK.Y0, aggPK.Y1
    function getAggregatePublicKeyG2(uint32 numNonSigners)
        internal
        returns (uint256 aggPKX0, uint256 aggPKX1, uint256 aggPKY0, uint256 aggPKY1)
    {
        setSignatureFile(numNonSigners);
        aggPKX0 = getUintFromJson(signatureJson, "aggregateSignature.AggPubkeyG2.X.A0");
        aggPKX1 = getUintFromJson(signatureJson, "aggregateSignature.AggPubkeyG2.X.A1");
        aggPKY0 = getUintFromJson(signatureJson, "aggregateSignature.AggPubkeyG2.Y.A0");
        aggPKY1 = getUintFromJson(signatureJson, "aggregateSignature.AggPubkeyG2.Y.A1");

        return (aggPKX0, aggPKX1, aggPKY0, aggPKY1);
    }

    function getAggPubKeyG2WithoutNonSigners(uint32 numNonSigners)
        internal
        returns (uint256 aggPKX0, uint256 aggPKX1, uint256 aggPKY0, uint256 aggPKY1)
    {
        setSignatureFile(numNonSigners);
        aggPKX0 = getAggPubKeyG2WithoutNonSignersFromJson(signatureJson, "AggPubkeyG2WithoutNonSigners.X.A0");
        aggPKX1 = getAggPubKeyG2WithoutNonSignersFromJson(signatureJson, "AggPubkeyG2WithoutNonSigners.X.A1");
        aggPKY0 = getAggPubKeyG2WithoutNonSignersFromJson(signatureJson, "AggPubkeyG2WithoutNonSigners.Y.A0");
        aggPKY1 = getAggPubKeyG2WithoutNonSignersFromJson(signatureJson, "AggPubkeyG2WithoutNonSigners.Y.A1");

        return (aggPKX0, aggPKX1, aggPKY0, aggPKY1);
    }

    //returns aggPK.X, aggPK.Y
    function getAggregatePublicKeyG1(uint32 numNonSigners)
        internal 
        returns (uint256 aggPKX, uint256 aggPKY)
    {
        setSignatureFile(numNonSigners);
        aggPKX = getUintFromJson(signatureJson, "aggregateSignature.AggPubkeyG1.X");
        aggPKY = getUintFromJson(signatureJson, "aggregateSignature.AggPubkeyG1.Y");

        return (aggPKX, aggPKY);
    }

    //get the aggregate signature of all 15 signers
    function getAggSignature(uint256 index, uint32 numNonSigners) internal returns (uint256 sigX, uint256 sigY) {
        setSignatureFile(numNonSigners);
        if (index == 0){
            sigX = getUintFromJson(signatureJson, "aggregateSignature.Signature.X");
            sigY = getUintFromJson(signatureJson, "aggregateSignature.Signature.Y");
        }
        //if index is > 0 that means there are multiple datastores in the test.
        //In this case only the signature changes, so we pull it from a helper function
        else{
            (sigX, sigY) = getSignatureOnAdditionalDataStore(index);
        }

        return (sigX, sigY);
    }

    function getNonSignerPK(uint32 pkIndex, uint32 numNonSigners) internal returns (uint256 PKX, uint256 PKY) {
        setSignatureFile(numNonSigners);
        PKX = getNonSignerPKFromJson(signatureJson, pkIndex, "PubkeyG1.X");
        PKY = getNonSignerPKFromJson(signatureJson, pkIndex, "PubkeyG1.Y");
        return(PKX, PKY);
    }

    function getNonSignerAggSig(uint32 numNonSigners) internal returns (uint256 sigmaX, uint256 sigmaY) {
        setSignatureFile(numNonSigners);
        sigmaX = getNonSignerAggSigFromJson(signatureJson, "AggSignature.X");
        sigmaY = getNonSignerAggSigFromJson(signatureJson, "AggSignature.Y");

        return(sigmaX, sigmaY);
    }
    //this is for testConfirmDataStoreLoop.  The signature changes when you increment the datastoreID.  That tests
    // Increments it twice more, so we have two additional signatures for 15 signers.
    function getSignatureOnAdditionalDataStore(uint256 index) internal pure returns(uint256 sigX, uint256 sigY){
        if(index == 1){
            sigX = 8733421643731740631536151352850909620731477919778472879183197990316971409408;
            sigY = 13039511527876118055099735862345750567119937347959409814678461768562329210287;
        }
        if(index == 2){
            sigX = 6538597068907069739387395712285884705875484283135662276126649661004687333241;
            sigY = 16337857511690675761965203377410131750898415908969868647246684783625084131030;
        }
        return(sigX, sigY);
        
    }

    function getUintFromJson(string memory json, string memory key) internal returns(uint256){
        string memory word =  stdJson.readString(json, key);
        return convertStringToUint(word);
    }

    function getNonSignerPKFromJson(string memory json, uint256 pubkeyIndex, string memory key) internal returns(uint256){
        string memory temp = string.concat(vm.toString(pubkeyIndex), "].");
        string memory pubKeyEntry = string.concat("nonSignersData.NonSigners[", temp);
        string memory word =  stdJson.readString(json, string.concat(pubKeyEntry, key));

        return convertStringToUint(word);
    }

    function getNonSignerAggSigFromJson(string memory json, string memory key) internal returns(uint256){
        string memory word =  stdJson.readString(json, string.concat("nonSignersData.", key));
        return convertStringToUint(word);
    }

    function getAggPubKeyG2WithoutNonSignersFromJson(string memory json, string memory key) internal returns(uint256){
        string memory word =  stdJson.readString(json, string.concat("nonSignersData.", key));
        return convertStringToUint(word);
    }

    function convertStringToUint(string memory s) public pure returns (uint) {
        bytes memory b = bytes(s);
        uint result = 0;
        for (uint i = 0; i < b.length; i++) {
            if (uint256(uint8(b[i])) >= 48 && uint256(uint8(b[i])) <= 57) {
                result = result * 10 + (uint256(uint8(b[i])) - 48); 
            }
        }
        return result;
    }

    function setSignatureFile(uint32 numNonSigners) internal {
        signatureJson = vm.readFile(string.concat("./src/test/test-data/signatures_", string.concat(Strings.toString(numNonSigners) , "_nonsigner.json")));
    }
}