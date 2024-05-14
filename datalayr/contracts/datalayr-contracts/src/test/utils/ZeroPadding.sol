// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "forge-std/Test.sol";
import "forge-std/Script.sol";
import "forge-std/StdJson.sol";
import "@openzeppelin/contracts/utils/Strings.sol";
import "./Signatures.sol";


contract ZeroPaddingUtils is Test,  SignatureUtils{

    string internal zeroPaddingJson;

    constructor() {
        zeroPaddingJson = vm.readFile("./src/test/test-data/zeroPadding.json");
    }


    function zeroPaddingProofData()
        internal 
        returns (uint256 metaPolyCommitX, uint256 metaPolyCommitY, uint256 paddingProofX, uint256 paddingProofY, uint256 paddingQuotientPolyCommitX, uint256 paddingQuotientPolyCommitY)
    {
        // get commitment of meta polynomial to G1
        metaPolyCommitX = getUintFromJson(zeroPaddingJson, "metaPolyCommit.X");
        metaPolyCommitY = getUintFromJson(zeroPaddingJson, "metaPolyCommit.Y");

        // get zero padding proof 
        paddingProofX =  getUintFromJson(zeroPaddingJson, "paddingproof.X");
        paddingProofY =  getUintFromJson(zeroPaddingJson, "paddingproof.Y");


        // get commitment of zero padding quotient poly to G1
        paddingQuotientPolyCommitX = getUintFromJson(zeroPaddingJson, "paddingQuotientPolyCommit.X");
        paddingQuotientPolyCommitY = getUintFromJson(zeroPaddingJson, "paddingQuotientPolyCommit.Y");
        
        return (metaPolyCommitX, metaPolyCommitY, paddingProofX, paddingProofY, paddingQuotientPolyCommitX, paddingQuotientPolyCommitY);
    }



}