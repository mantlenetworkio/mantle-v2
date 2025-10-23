// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "./CommonTest.t.sol";
import { L1Block } from "src/L2/L1Block.sol";

contract L1BlockTest is CommonTest {
    L1Block lb;
    address depositor;
    bytes32 immutable NON_ZERO_HASH = keccak256(abi.encode(1));

    function setUp() public virtual override {
        super.setUp();
        lb = new L1Block();
        depositor = lb.DEPOSITOR_ACCOUNT();
        vm.prank(depositor);
        lb.setL1BlockValues({
            _number: uint64(1),
            _timestamp: uint64(2),
            _basefee: 3,
            _hash: NON_ZERO_HASH,
            _sequenceNumber: uint64(4),
            _batcherHash: bytes32(0),
            _l1FeeOverhead: 2,
            _l1FeeScalar: 3
        });
    }

    function testFuzz_updatesValues_succeeds(
        uint64 n,
        uint64 t,
        uint256 b,
        bytes32 h,
        uint64 s,
        bytes32 bt,
        uint256 fo,
        uint256 fs
    )
        external
    {
        vm.prank(depositor);
        lb.setL1BlockValues(n, t, b, h, s, bt, fo, fs);
        assertEq(lb.number(), n);
        assertEq(lb.timestamp(), t);
        assertEq(lb.basefee(), b);
        assertEq(lb.hash(), h);
        assertEq(lb.sequenceNumber(), s);
        assertEq(lb.batcherHash(), bt);
        assertEq(lb.l1FeeOverhead(), fo);
        assertEq(lb.l1FeeScalar(), fs);
    }

    function test_number_succeeds() external {
        assertEq(lb.number(), uint64(1));
    }

    function test_timestamp_succeeds() external {
        assertEq(lb.timestamp(), uint64(2));
    }

    function test_basefee_succeeds() external {
        assertEq(lb.basefee(), 3);
    }

    function test_hash_succeeds() external {
        assertEq(lb.hash(), NON_ZERO_HASH);
    }

    function test_sequenceNumber_succeeds() external {
        assertEq(lb.sequenceNumber(), uint64(4));
    }

    function test_updateValues_succeeds() external {
        vm.prank(depositor);
        lb.setL1BlockValues({
            _number: type(uint64).max,
            _timestamp: type(uint64).max,
            _basefee: type(uint256).max,
            _hash: keccak256(abi.encode(1)),
            _sequenceNumber: type(uint64).max,
            _batcherHash: bytes32(type(uint256).max),
            _l1FeeOverhead: type(uint256).max,
            _l1FeeScalar: type(uint256).max
        });
    }

    function test_setL1BlockValues_notDepositor_reverts() external {
        vm.expectRevert("L1Block: only the depositor account can set L1 block values");
        lb.setL1BlockValues(1, 2, 3, bytes32(0), 4, bytes32(0), 5, 6);
    }

    function test_depositorAccount_succeeds() external view {
        assertEq(lb.DEPOSITOR_ACCOUNT(), 0xDeaDDEaDDeAdDeAdDEAdDEaddeAddEAdDEAd0001);
    }
}

contract L1Block_ArsiaTest is CommonTest {
    L1Block lb;
    address depositor;

    function setUp() public virtual override {
        super.setUp();
        lb = new L1Block();
        depositor = lb.DEPOSITOR_ACCOUNT();
    }

    function test_setL1BlockValuesArsia_succeeds() external {
        // Prepare calldata for Arsia format
        // According to assembly code comment, the packing is:
        // Bytes 4-7: baseFeeScalar (uint32)
        // Bytes 8-11: blobBaseFeeScalar (uint32)
        // Bytes 12-19: sequenceNumber (uint64)
        // Bytes 20-27: timestamp (uint64)
        // Bytes 28-35: number (uint64)
        // Bytes 36-67: basefee (uint256)
        // Bytes 68-99: blobBaseFee (uint256)
        // Bytes 100-131: hash (bytes32)
        // Bytes 132-163: batcherHash (bytes32)
        // Bytes 164-167: operatorFeeScalar (uint32)
        // Bytes 168-175: operatorFeeConstant (uint64)

        bytes memory data = abi.encodePacked(
            uint32(1000), // baseFeeScalar
            uint32(2000), // blobBaseFeeScalar
            uint64(5), // sequenceNumber
            uint64(12345), // timestamp
            uint64(67890), // number
            uint256(1 gwei), // basefee
            uint256(2 gwei), // blobBaseFee
            keccak256("test"), // hash
            keccak256("batcher"), // batcherHash
            uint32(500000), // operatorFeeScalar
            uint64(1000) // operatorFeeConstant
        );

        vm.prank(depositor);
        (bool success,) = address(lb).call(abi.encodePacked(lb.setL1BlockValuesArsia.selector, data));
        require(success, "Call failed");

        // Verify values by reading from storage
        // sequenceNumber is in slot 3 (standalone)
        bytes32 slot3 = vm.load(address(lb), bytes32(uint256(3)));
        uint64 sequenceNumberRead = uint64(uint256(slot3));
        assertEq(sequenceNumberRead, 5, "sequenceNumber mismatch");

        // baseFeeScalar and blobBaseFeeScalar are packed in slot 7
        bytes32 slot7 = vm.load(address(lb), bytes32(uint256(7)));
        uint32 baseFeeScalarRead = uint32(uint256(slot7));
        uint32 blobBaseFeeScalarRead = uint32(uint256(slot7) >> 32);
        assertEq(baseFeeScalarRead, 1000, "baseFeeScalar mismatch");
        assertEq(blobBaseFeeScalarRead, 2000, "blobBaseFeeScalar mismatch");

        // Verify other values
        assertEq(lb.timestamp(), 12345);
        assertEq(lb.number(), 67890);
        assertEq(lb.basefee(), 1 gwei);
        assertEq(lb.blobBaseFee(), 2 gwei);
        assertEq(lb.hash(), keccak256("test"));
        assertEq(lb.batcherHash(), keccak256("batcher"));

        // operatorFeeScalar and operatorFeeConstant are packed in slot 9
        bytes32 slot9 = vm.load(address(lb), bytes32(uint256(9)));
        uint64 operatorFeeConstantRead = uint64(uint256(slot9));
        uint32 operatorFeeScalarRead = uint32(uint256(slot9) >> 64);

        assertEq(operatorFeeScalarRead, 500000, "operatorFeeScalar mismatch");
        assertEq(operatorFeeConstantRead, 1000, "operatorFeeConstant mismatch");
    }

    function testFuzz_setL1BlockValuesArsia_succeeds(
        uint32 baseFeeScalar_,
        uint32 blobBaseFeeScalar_,
        uint64 sequenceNumber_,
        uint64 timestamp_,
        uint64 number_,
        uint256 basefee_,
        uint256 blobBaseFee_,
        bytes32 hash_,
        bytes32 batcherHash_,
        uint32 operatorFeeScalar_,
        uint64 operatorFeeConstant_
    )
        external
    {
        bytes memory data = abi.encodePacked(
            baseFeeScalar_,
            blobBaseFeeScalar_,
            sequenceNumber_,
            timestamp_,
            number_,
            basefee_,
            blobBaseFee_,
            hash_,
            batcherHash_,
            operatorFeeScalar_,
            operatorFeeConstant_
        );

        vm.prank(depositor);
        (bool success,) = address(lb).call(abi.encodePacked(lb.setL1BlockValuesArsia.selector, data));
        require(success, "Call failed");

        assertEq(lb.baseFeeScalar(), baseFeeScalar_);
        assertEq(lb.blobBaseFeeScalar(), blobBaseFeeScalar_);
        assertEq(lb.sequenceNumber(), sequenceNumber_);
        assertEq(lb.timestamp(), timestamp_);
        assertEq(lb.number(), number_);
        assertEq(lb.basefee(), basefee_);
        assertEq(lb.blobBaseFee(), blobBaseFee_);
        assertEq(lb.hash(), hash_);
        assertEq(lb.batcherHash(), batcherHash_);
        assertEq(lb.operatorFeeScalar(), operatorFeeScalar_);
        assertEq(lb.operatorFeeConstant(), operatorFeeConstant_);
    }

    function test_setL1BlockValuesArsia_notDepositor_reverts() external {
        bytes memory data = abi.encodePacked(
            uint32(1000),
            uint32(2000),
            uint64(5),
            uint64(12345),
            uint64(67890),
            uint256(1 gwei),
            uint256(2 gwei),
            keccak256("test"),
            keccak256("batcher"),
            uint32(500000),
            uint64(1000)
        );

        // Should revert with NotDepositor() error
        (bool success, bytes memory returnData) =
            address(lb).call(abi.encodePacked(lb.setL1BlockValuesArsia.selector, data));

        assertFalse(success, "Call should have reverted");
        // Check that the error selector is 0x3cc50b45 (NotDepositor())
        assertEq(bytes4(returnData), bytes4(hex"3cc50b45"));
    }

    function test_setL1BlockValuesArsia_maxValues_succeeds() external {
        bytes memory data = abi.encodePacked(
            type(uint32).max, // baseFeeScalar
            type(uint32).max, // blobBaseFeeScalar
            type(uint64).max, // sequenceNumber
            type(uint64).max, // timestamp
            type(uint64).max, // number
            type(uint256).max, // basefee
            type(uint256).max, // blobBaseFee
            bytes32(type(uint256).max), // hash
            bytes32(type(uint256).max), // batcherHash
            type(uint32).max, // operatorFeeScalar
            type(uint64).max // operatorFeeConstant
        );

        vm.prank(depositor);
        (bool success,) = address(lb).call(abi.encodePacked(lb.setL1BlockValuesArsia.selector, data));
        require(success, "Call failed");

        assertEq(lb.baseFeeScalar(), type(uint32).max);
        assertEq(lb.blobBaseFeeScalar(), type(uint32).max);
        assertEq(lb.operatorFeeScalar(), type(uint32).max);
        assertEq(lb.operatorFeeConstant(), type(uint64).max);
    }

    function test_setL1BlockValuesArsia_zeroValues_succeeds() external {
        bytes memory data = abi.encodePacked(
            uint32(0),
            uint32(0),
            uint64(0),
            uint64(0),
            uint64(0),
            uint256(0),
            uint256(0),
            bytes32(0),
            bytes32(0),
            uint32(0),
            uint64(0)
        );

        vm.prank(depositor);
        (bool success,) = address(lb).call(abi.encodePacked(lb.setL1BlockValuesArsia.selector, data));
        require(success, "Call failed");

        assertEq(lb.baseFeeScalar(), 0);
        assertEq(lb.blobBaseFeeScalar(), 0);
        assertEq(lb.operatorFeeScalar(), 0);
        assertEq(lb.operatorFeeConstant(), 0);
    }
}
