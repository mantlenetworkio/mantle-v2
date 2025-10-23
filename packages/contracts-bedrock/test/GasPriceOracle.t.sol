// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "./CommonTest.t.sol";
import { GasPriceOracle } from "src/L2/GasPriceOracle.sol";
import { L1Block } from "src/L2/L1Block.sol";
import { Predeploys } from "src/libraries/Predeploys.sol";

contract GasPriceOracle_Test is CommonTest {
    event OverheadUpdated(uint256);
    event ScalarUpdated(uint256);
    event DecimalsUpdated(uint256);

    GasPriceOracle gasOracle;
    L1Block l1Block;
    address depositor;

    // set the initial L1 context values
    uint64 constant number = 10;
    uint64 constant timestamp = 11;
    uint256 constant basefee = 100;
    bytes32 constant hash = bytes32(uint256(64));
    uint64 constant sequenceNumber = 0;
    bytes32 constant batcherHash = bytes32(uint256(777));
    uint256 constant l1FeeOverhead = 310;
    uint256 constant l1FeeScalar = 10;

    function setUp() public virtual override {
        super.setUp();
        // place the L1Block contract at the predeploy address
        vm.etch(Predeploys.L1_BLOCK_ATTRIBUTES, address(new L1Block()).code);

        l1Block = L1Block(Predeploys.L1_BLOCK_ATTRIBUTES);
        depositor = l1Block.DEPOSITOR_ACCOUNT();

        // We are not setting the gas oracle at its predeploy
        // address for simplicity purposes. Nothing in this test
        // requires it to be at a particular address
        gasOracle = new GasPriceOracle();

        vm.prank(depositor);
        l1Block.setL1BlockValues({
            _number: number,
            _timestamp: timestamp,
            _basefee: basefee,
            _hash: hash,
            _sequenceNumber: sequenceNumber,
            _batcherHash: batcherHash,
            _l1FeeOverhead: l1FeeOverhead,
            _l1FeeScalar: l1FeeScalar
        });
    }

    function test_l1BaseFee_succeeds() external {
        assertEq(gasOracle.l1BaseFee(), basefee);
    }

    function test_getL1GasUsed_succeeds() external {
        assertEq(gasOracle.getL1GasUsed("dead"), 1462);
    }

    function test_getL1Fee_succeeds() external {
        assertEq(gasOracle.getL1Fee("dead"), 1);
    }

    function test_gasPrice_succeeds() external {
        vm.fee(100);
        uint256 gasPrice = gasOracle.gasPrice();
        assertEq(gasPrice, 100);
    }

    function test_baseFee_succeeds() external {
        vm.fee(64);
        uint256 gasPrice = gasOracle.baseFee();
        assertEq(gasPrice, 64);
    }

    function test_scalar_succeeds() external {
        assertEq(gasOracle.scalar(), l1FeeScalar);
    }

    function test_overhead_succeeds() external {
        assertEq(gasOracle.overhead(), l1FeeOverhead);
    }

    function test_decimals_succeeds() external {
        assertEq(gasOracle.decimals(), 6);
        assertEq(gasOracle.DECIMALS(), 6);
    }

    // Removed in bedrock
    function test_setGasPrice_doesNotExist_reverts() external {
        (bool success, bytes memory returndata) =
            address(gasOracle).call(abi.encodeWithSignature("setGasPrice(uint256)", 1));

        assertEq(success, false);
        assertEq(returndata, hex"");
    }

    // Removed in bedrock
    function test_setL1BaseFee_doesNotExist_reverts() external {
        (bool success, bytes memory returndata) =
            address(gasOracle).call(abi.encodeWithSignature("setL1BaseFee(uint256)", 1));

        assertEq(success, false);
        assertEq(returndata, hex"");
    }
}

contract GasPriceOracle_Ownership_Test is CommonTest {
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);
    event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio);

    GasPriceOracle gasOracle;
    L1Block l1Block;
    address depositor;
    address owner;
    address operator;

    function setUp() public virtual override {
        super.setUp();
        vm.etch(Predeploys.L1_BLOCK_ATTRIBUTES, address(new L1Block()).code);

        l1Block = L1Block(Predeploys.L1_BLOCK_ATTRIBUTES);
        depositor = l1Block.DEPOSITOR_ACCOUNT();

        gasOracle = new GasPriceOracle();
        owner = address(this);
        operator = alice;

        // Set owner first, then operator
        vm.startPrank(address(0)); // GasPriceOracle doesn't have an initial owner
        vm.stopPrank();

        // Manually set the owner by transferring from zero address
        vm.store(address(gasOracle), bytes32(uint256(1)), bytes32(uint256(uint160(owner)))); // owner is in slot 1
        vm.store(address(gasOracle), bytes32(uint256(2)), bytes32(uint256(uint160(operator)))); // operator is in slot 2
    }

    function test_transferOwnership_succeeds() external {
        address newOwner = bob;

        vm.expectEmit(true, true, false, false);
        emit OwnershipTransferred(owner, newOwner);

        gasOracle.transferOwnership(newOwner);

        assertEq(gasOracle.owner(), newOwner);
    }

    function test_transferOwnership_notOwner_reverts() external {
        vm.prank(bob);
        vm.expectRevert("Caller is not the owner");
        gasOracle.transferOwnership(bob);
    }

    function test_transferOwnership_zeroAddress_reverts() external {
        vm.expectRevert("new owner is the zero address");
        gasOracle.transferOwnership(address(0));
    }

    function test_setOperator_succeeds() external {
        address newOperator = bob;

        vm.expectEmit(true, true, false, false);
        emit OperatorUpdated(operator, newOperator);

        gasOracle.setOperator(newOperator);

        assertEq(gasOracle.operator(), newOperator);
    }

    function test_setOperator_notOwner_reverts() external {
        vm.prank(bob);
        vm.expectRevert("Caller is not the owner");
        gasOracle.setOperator(bob);
    }

    function test_setTokenRatio_succeeds() external {
        uint256 newRatio = 12345;

        vm.expectEmit(true, true, false, false);
        emit TokenRatioUpdated(0, newRatio);

        vm.prank(operator);
        gasOracle.setTokenRatio(newRatio);

        assertEq(gasOracle.tokenRatio(), newRatio);
    }

    function test_setTokenRatio_notOperator_reverts() external {
        vm.prank(bob);
        vm.expectRevert("Caller is not the operator");
        gasOracle.setTokenRatio(12345);
    }

    function testFuzz_setTokenRatio_succeeds(uint256 newRatio) external {
        vm.prank(operator);
        gasOracle.setTokenRatio(newRatio);
        assertEq(gasOracle.tokenRatio(), newRatio);
    }
}

contract GasPriceOracle_Arsia_Test is CommonTest {
    GasPriceOracle gasOracle;
    L1Block l1Block;
    address depositor;

    uint32 constant baseFeeScalar = 1000;
    uint32 constant blobBaseFeeScalar = 2000;
    uint256 constant l1BaseFee = 1 gwei;
    uint256 constant l1BlobBaseFee = 2 gwei;
    uint32 constant operatorFeeScalar_val = 500000; // 0.5 in 1e6
    uint64 constant operatorFeeConstant_val = 1000;

    function setUp() public virtual override {
        super.setUp();
        vm.etch(Predeploys.L1_BLOCK_ATTRIBUTES, address(new L1Block()).code);

        l1Block = L1Block(Predeploys.L1_BLOCK_ATTRIBUTES);
        depositor = l1Block.DEPOSITOR_ACCOUNT();

        gasOracle = new GasPriceOracle();

        // Initialize with Arsia values
        bytes memory data = abi.encodePacked(
            baseFeeScalar,
            blobBaseFeeScalar,
            uint64(5), // sequenceNumber
            uint64(12345), // timestamp
            uint64(67890), // number
            l1BaseFee,
            l1BlobBaseFee,
            keccak256("test"), // hash
            keccak256("batcher"), // batcherHash
            operatorFeeScalar_val,
            operatorFeeConstant_val
        );

        vm.prank(depositor);
        (bool success,) = address(l1Block).call(abi.encodePacked(l1Block.setL1BlockValuesArsia.selector, data));
        require(success, "L1Block setup failed");

        // Set owner and enable Arsia
        vm.store(address(gasOracle), bytes32(uint256(1)), bytes32(uint256(uint160(address(this))))); // Set owner
        gasOracle.setArsia();
    }

    function test_setArsia_succeeds() external {
        // Create a new oracle that hasn't been set to Arsia yet
        GasPriceOracle newOracle = new GasPriceOracle();
        vm.store(address(newOracle), bytes32(uint256(1)), bytes32(uint256(uint160(address(this)))));

        assertEq(newOracle.isArsia(), false);

        newOracle.setArsia();

        assertEq(newOracle.isArsia(), true);
    }

    function test_setArsia_notOwner_reverts() external {
        GasPriceOracle newOracle = new GasPriceOracle();
        vm.store(address(newOracle), bytes32(uint256(1)), bytes32(uint256(uint160(alice))));

        vm.prank(bob);
        vm.expectRevert("Caller is not the owner");
        newOracle.setArsia();
    }

    function test_setArsia_alreadyActive_reverts() external {
        vm.expectRevert("GasPriceOracle: Arsia already active");
        gasOracle.setArsia();
    }

    function test_baseFeeScalar_succeeds() external view {
        assertEq(gasOracle.baseFeeScalar(), baseFeeScalar);
    }

    function test_blobBaseFeeScalar_succeeds() external view {
        assertEq(gasOracle.blobBaseFeeScalar(), blobBaseFeeScalar);
    }

    function test_blobBaseFee_succeeds() external view {
        assertEq(gasOracle.blobBaseFee(), l1BlobBaseFee);
    }

    function test_operatorFeeScalar_succeeds() external view {
        assertEq(gasOracle.operatorFeeScalar(), operatorFeeScalar_val);
    }

    function test_operatorFeeConstant_succeeds() external view {
        assertEq(gasOracle.operatorFeeConstant(), operatorFeeConstant_val);
    }

    function test_getL1Fee_arsia_succeeds() external view {
        bytes memory data = hex"dead";
        uint256 fee = gasOracle.getL1Fee(data);

        // Verify fee is non-zero
        assertGt(fee, 0);

        // Calculate expected fee manually
        // hex"dead" = 2 non-zero bytes (0xde, 0xad) + 68 bytes padding
        uint256 l1GasUsed = (2 * 16 + 68 * 16); // 2 non-zero bytes + 68 padding
        uint256 scaledBaseFee = baseFeeScalar * 16 * l1BaseFee;
        uint256 scaledBlobBaseFee = blobBaseFeeScalar * l1BlobBaseFee;
        uint256 expectedFee = l1GasUsed * (scaledBaseFee + scaledBlobBaseFee) / (16 * 10 ** 6);

        assertEq(fee, expectedFee);
    }

    function testFuzz_getL1Fee_arsia_succeeds(bytes memory data) external view {
        // Bound data length to avoid gas limit issues
        vm.assume(data.length > 0 && data.length < 10000);

        uint256 fee = gasOracle.getL1Fee(data);
        assertGt(fee, 0);
    }

    function test_getOperatorFee_succeeds() external view {
        uint256 gasUsed = 1_000_000;
        uint256 fee = gasOracle.getOperatorFee(gasUsed);

        // Expected: (1_000_000 * 500000) / 1e6 + 1000 = 500_000 + 1000 = 501_000
        uint256 expected = (gasUsed * operatorFeeScalar_val) / 1e6 + operatorFeeConstant_val;
        assertEq(fee, expected);
    }

    function test_getOperatorFee_zero_succeeds() external view {
        uint256 fee = gasOracle.getOperatorFee(0);
        assertEq(fee, operatorFeeConstant_val);
    }

    function testFuzz_getOperatorFee_succeeds(uint256 gasUsed) external view {
        // Bound to avoid overflow
        vm.assume(gasUsed < type(uint256).max / operatorFeeScalar_val);

        uint256 fee = gasOracle.getOperatorFee(gasUsed);
        uint256 expected = (gasUsed * operatorFeeScalar_val) / 1e6 + operatorFeeConstant_val;
        assertEq(fee, expected);
    }

    function test_getOperatorFee_notArsia_returnsZero() external {
        // Create a new oracle without Arsia enabled
        GasPriceOracle newOracle = new GasPriceOracle();
        uint256 fee = newOracle.getOperatorFee(1_000_000);
        assertEq(fee, 0);
    }
}

contract GasPriceOracle_CalldataGas_Test is CommonTest {
    GasPriceOracle gasOracle;
    L1Block l1Block;
    address depositor;

    function setUp() public virtual override {
        super.setUp();
        vm.etch(Predeploys.L1_BLOCK_ATTRIBUTES, address(new L1Block()).code);

        l1Block = L1Block(Predeploys.L1_BLOCK_ATTRIBUTES);
        depositor = l1Block.DEPOSITOR_ACCOUNT();

        gasOracle = new GasPriceOracle();

        vm.prank(depositor);
        l1Block.setL1BlockValues({
            _number: 10,
            _timestamp: 11,
            _basefee: 100,
            _hash: bytes32(uint256(64)),
            _sequenceNumber: 0,
            _batcherHash: bytes32(uint256(777)),
            _l1FeeOverhead: 310,
            _l1FeeScalar: 10
        });
    }

    function test_getL1GasUsed_empty_succeeds() external view {
        bytes memory empty = "";
        // 68 bytes of padding * 16 = 1088
        assertEq(gasOracle.getL1GasUsed(empty), 1088 + 310);
    }

    function test_getL1GasUsed_allZeros_succeeds() external view {
        bytes memory zeros = hex"0000";
        // 2 zero bytes * 4 + 68 * 16 = 8 + 1088 = 1096
        assertEq(gasOracle.getL1GasUsed(zeros), 1096 + 310);
    }

    function test_getL1GasUsed_allNonZeros_succeeds() external view {
        bytes memory nonZeros = hex"ffff";
        // 2 non-zero bytes * 16 + 68 * 16 = 32 + 1088 = 1120
        assertEq(gasOracle.getL1GasUsed(nonZeros), 1120 + 310);
    }

    function test_getL1GasUsed_mixed_succeeds() external view {
        bytes memory mixed = hex"00ff";
        // 1 zero * 4 + 1 non-zero * 16 + 68 * 16 = 4 + 16 + 1088 = 1108
        assertEq(gasOracle.getL1GasUsed(mixed), 1108 + 310);
    }

    function testFuzz_getL1GasUsed_succeeds(bytes memory data) external view {
        vm.assume(data.length < 10000);
        uint256 gasUsed = gasOracle.getL1GasUsed(data);

        // Verify gas used is at least the overhead + padding
        assertGe(gasUsed, 310 + 1088);
    }
}

contract GasPriceOracle_FeeBedrock_Test is CommonTest {
    GasPriceOracle gasOracle;
    L1Block l1Block;
    address depositor;

    uint256 constant l1BaseFee = 1 gwei;
    uint256 constant l1FeeOverhead = 188;
    uint256 constant l1FeeScalar = 684000;

    function setUp() public virtual override {
        super.setUp();
        vm.etch(Predeploys.L1_BLOCK_ATTRIBUTES, address(new L1Block()).code);

        l1Block = L1Block(Predeploys.L1_BLOCK_ATTRIBUTES);
        depositor = l1Block.DEPOSITOR_ACCOUNT();

        gasOracle = new GasPriceOracle();

        vm.prank(depositor);
        l1Block.setL1BlockValues({
            _number: 10,
            _timestamp: 11,
            _basefee: l1BaseFee,
            _hash: bytes32(uint256(64)),
            _sequenceNumber: 0,
            _batcherHash: bytes32(uint256(777)),
            _l1FeeOverhead: l1FeeOverhead,
            _l1FeeScalar: l1FeeScalar
        });
    }

    function test_getL1Fee_bedrock_succeeds() external view {
        bytes memory data = hex"dead";

        uint256 fee = gasOracle.getL1Fee(data);

        // Calculate expected fee
        uint256 l1GasUsed = gasOracle.getL1GasUsed(data);
        uint256 l1Fee = l1GasUsed * l1BaseFee;
        uint256 unscaled = l1Fee * l1FeeScalar;
        uint256 expected = unscaled / (10 ** 6);

        assertEq(fee, expected);
    }

    function testFuzz_getL1Fee_bedrock_succeeds(bytes memory data) external view {
        vm.assume(data.length > 0 && data.length < 10000);

        uint256 fee = gasOracle.getL1Fee(data);
        assertGt(fee, 0);
    }
}
