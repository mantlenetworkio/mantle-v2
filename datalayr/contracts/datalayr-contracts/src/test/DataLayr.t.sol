// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "./DataLayrTestHelper.t.sol";
import "forge-std/Test.sol";

contract DataLayrTests is DSTest, DataLayrTestHelper {

    //checks that it is possible to init a data store
    function testInitDataStore() public returns (bytes32) {
        uint256 numSigners = 15;

        //register all the operators
        _registerNumSigners(numSigners);

        //change the current timestamp to be in the future 100 seconds and init
        return _testInitDataStore(block.timestamp + 100, address(this), header).metadata.headerHash;
    }

    function testLoopInitDataStore() public {
        uint256 g = gasleft();
        uint256 numSigners = 15;

        for (uint256 i = 0; i < 20; i++) {
            if(i==0){
                _registerNumSigners(numSigners);
            }
            _testInitDataStore(block.timestamp + 100, address(this), header).metadata.headerHash;
        }
        emit log_named_uint("gas", g - gasleft());
    }

    //verifies that it is possible to confirm a data store
    //checks that the store is marked as committed
    function testConfirmDataStore() public {
        _testConfirmDataStoreSelfOperators(15);
    }

    function testLoopConfirmDataStore() public {
        _testConfirmDataStoreSelfOperators(15);
        uint256 g = gasleft();
        for (uint256 i = 1; i < 3; i++) {
            _testConfirmDataStoreWithoutRegister(block.timestamp, i, 15);
        }
        emit log_named_uint("gas", g - gasleft());
    }

    function testCodingRatio() public {

        uint256 numSigners = 15;
        //register all the operators
        _registerNumSigners(numSigners);

        /// @notice this header has numSys set to 9.  Thus coding ration = 9/15, which is greater than the set adversary threshold in DataLayrServiceManager.
        bytes memory header = hex"0e75f28b7a90f89995e522d0cd3a340345e60e249099d4cd96daef320a3abfc31df7f4c8f6f8bc5dc1de03f56202933ec2cc40acad1199f40c7b42aefd45bfb10000000800000009000000020000014000000000000000000000000000000000000000002b4982b07d4e522c2a94b3e7c5ab68bfeecc33c5fa355bc968491c62c12cf93f0cd04099c3d9742620bf0898cf3843116efc02e6f7d408ba443aa472f950e4f3";

        uint256 initTimestamp = block.timestamp + 100;
        address confirmer  = address(this);

        _testInitDataStoreExpectRevert(initTimestamp, confirmer, header, bytes("DataLayrServiceManager.initDataStore: Coding ratio is too high"));
    }

    function testZeroTotalBytes() public {
        uint256 initTimestamp = block.timestamp + 100;
        address confirmer  = address(this);

        _testInitDataStoreExpectRevert(initTimestamp, confirmer, header, bytes("DataLayrServiceManager.initDataStore: totalBytes < MIN_STORE_SIZE"));

    }

    function testTotalOperatorIndex(uint32 wrongTotalOperatorsIndex) external {
        uint256 numSigners = 15;
        //register all the operators
        _registerNumSigners(numSigners);
        uint256 initTimestamp = block.timestamp + 100;

        cheats.assume(wrongTotalOperatorsIndex > uint32(dlReg.getLengthOfTotalOperatorsHistory()));

        // weth is set as the paymentToken of dlsm, so we must approve dlsm to transfer weth
        weth.transfer(storer, 1e11);
        cheats.startPrank(storer);

        uint32 blockNumber = uint32(block.number);

        require(initTimestamp >= block.timestamp, "_testInitDataStore: warping back in time!");
        cheats.warp(initTimestamp);

        cheats.expectRevert();
        dlsm.initDataStore(
            storer,
            address(this),
            durationToInit,
            blockNumber,
            wrongTotalOperatorsIndex,
            header
        );
    }


     //testing inclusion of nonsigners in DLN quorum, ensuring that nonsigner inclusion proof is working correctly.
    function testInadequateQuorumStake(uint256 ethAmount, uint256 eigenAmount) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e10);

        {
            // address operator = getOperatorAddress(0);
            _testInitiateDelegation(0, eigenAmount, ethAmount);
            _testRegisterBLSPubKey(0);
            _testRegisterOperatorWithDataLayr(0, 3, testSocket);
        }

        NonSignerPK memory nonSignerPK1;
        NonSignerPK memory nonSignerPK2;
        RegistrantAPKG2 memory registrantApkG2;
        RegistrantAPKG1 memory registrantApkG1;
        SignerAggSig memory signerAggSig;

        uint32 numberOfNonSigners = 2;

        (signerAggSig.sigma0,  signerAggSig.sigma1) = getNonSignerAggSig(numberOfNonSigners);
        (nonSignerPK1.x, nonSignerPK1.y) = getNonSignerPK(0, numberOfNonSigners);
        (nonSignerPK2.x, nonSignerPK2.y) = getNonSignerPK(1, numberOfNonSigners);
        (registrantApkG2.apk0, registrantApkG2.apk1, registrantApkG2.apk2, registrantApkG2.apk3) = getAggPubKeyG2WithoutNonSigners(numberOfNonSigners);
        (registrantApkG1.apk0, registrantApkG1.apk1) = getAggregatePublicKeyG1(numberOfNonSigners);

        {

            _testRegisterSigners(15, false);
        }
        bytes memory data;
        IDataLayrServiceManager.DataStoreSearchData memory searchData = _testInitDataStore(1000000001, address(this), header);
        // multiple scoped blocks helps fix 'stack too deep' errors
        {
            uint256 initTime = 1000000001;
            data = _getTwoNonSignerCallData(
                keccak256(
                    abi.encodePacked(
                        searchData.metadata.globalDataStoreId,
                        searchData.metadata.headerHash,
                        searchData.duration,
                        initTime,
                        uint32(0)
                    )
                ),
                2, //number of nonSigners
                registrantApkG1,
                registrantApkG2,
                signerAggSig,
                nonSignerPK1,
                nonSignerPK2,
                searchData.metadata.referenceBlockNumber,
                dlsm.taskNumber() - 1 //dataStoreID
            );
        }

        cheats.expectRevert(bytes("DataLayrServiceManager.confirmDataStore: signatories do not own at least threshold percentage of both quorums"));
        dlsm.confirmDataStore(data, searchData);

    }

    //testing inclusion of nonsigners in DLN quorum, ensuring that nonsigner inclusion proof is working correctly.
    function testForNonSigners(uint256 ethAmount, uint256 eigenAmount) public {
        cheats.assume(ethAmount > 0 && ethAmount < 1e18);
        cheats.assume(eigenAmount > 0 && eigenAmount < 1e10);

        // address operator = getOperatorAddress(0);
        uint8 operatorType = 3;
        _testInitiateDelegation(0, eigenAmount, ethAmount);
        _testRegisterBLSPubKey(0);
        _testRegisterOperatorWithDataLayr(0, operatorType, testSocket);

        NonSignerPK memory nonSignerPK;
        RegistrantAPKG2 memory registrantApkG2;
        RegistrantAPKG1 memory registrantApkG1;
        SignerAggSig memory signerAggSig;
        uint32 numberOfNonSigners = 1;

        (nonSignerPK.x, nonSignerPK.y) = getNonSignerPK(numberOfNonSigners-1, numberOfNonSigners);
        (signerAggSig.sigma0,  signerAggSig.sigma1) = getNonSignerAggSig(numberOfNonSigners);
        //the non signer is the 15th operator with stake Index 14
        (registrantApkG2.apk0, registrantApkG2.apk1, registrantApkG2.apk2, registrantApkG2.apk3) = getAggPubKeyG2WithoutNonSigners(numberOfNonSigners);
        //in BLSSignatureChecker we only is G1 PK to subtract NonSignerPK's from, so we pass in the full signer set aggPK
        (registrantApkG1.apk0, registrantApkG1.apk1) = getAggregatePublicKeyG1(numberOfNonSigners);

        uint32 numberOfSigners = 15;
        _testRegisterSigners(numberOfSigners, false);

        // scoped block helps fix 'stack too deep' errors
        {
            uint256 initTime = 1000000001;
            IDataLayrServiceManager.DataStoreSearchData memory searchData = _testInitDataStore(initTime, address(this), header);
            uint32 dataStoreId = dlsm.taskNumber() - 1;

            bytes memory data = _getOneNonSignerCallData(
                keccak256(
                    abi.encodePacked(
                        searchData.metadata.globalDataStoreId,
                        searchData.metadata.headerHash,
                        searchData.duration,
                        initTime,
                        uint32(0)
                    )
                ),
                numberOfNonSigners,
                registrantApkG2,
                registrantApkG1,
                signerAggSig,
                nonSignerPK,
                searchData.metadata.referenceBlockNumber,
                dataStoreId
            );

            uint256 gasbefore = gasleft();

            dlsm.confirmDataStore(data, searchData);

            emit log_named_uint("gas cost", gasbefore - gasleft());
        }
    }


    function _registerNumSigners(uint256 numSigners) internal {
        for (uint256 i = 0; i < numSigners; ++i) {
            _testRegisterAdditionalSelfOperator(i);
        }
    }

    // function testZeroPadding() public {
    //     uint256 metaPolyCommitX;
    //     uint256 metaPolyCommitY;
    //     uint256 paddingProofX;
    //     uint256 paddingProofY;
    //     uint256 paddingQuotientPolyCommitX;
    //     uint256 paddingQuotientPolyCommitY;




    //     (metaPolyCommitX, metaPolyCommitY, paddingProofX, paddingProofY, paddingQuotientPolyCommitX, paddingQuotientPolyCommitY) =  zeroPaddingProofData();

    //     BN254.G1Point memory metaPolyCommit;
    //     BN254.G1Point memory paddingProof;
    //     BN254.G1Point memory paddingQuotientPolyCommit;
    //     metaPolyCommit = BN254.G1Point({X: metaPolyCommitX, Y: metaPolyCommitY});
    //     paddingProof = BN254.G1Point({X: paddingProofX, Y: paddingProofY});
    //     paddingQuotientPolyCommit = BN254.G1Point({X: paddingQuotientPolyCommitX, Y: paddingQuotientPolyCommitY});



    //     assertEq(dlldcImplementation.verifyZeroPaddingProof(metaPolyCommit, paddingProof, paddingQuotientPolyCommit), true);

    // }
}
