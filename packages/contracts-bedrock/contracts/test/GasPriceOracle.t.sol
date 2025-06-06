// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "./CommonTest.t.sol";
import { GasPriceOracle } from "../L2/GasPriceOracle.sol";
import { L1Block } from "../L2/L1Block.sol";
import { Predeploys } from "../libraries/Predeploys.sol";

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

    uint256 constant operatorFeeConstant = 0;
    uint256 constant operatorFeeScalar = 5490_000000;

    struct TestData {
        bytes data;
        uint256 gasUsed;
        uint256 l1Fee;
    }

    TestData[] public caseBedrock;
    TestData[] public caseSkadi;

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
        vm.store(address(gasOracle), bytes32(uint256(1)), bytes32(uint256(uint160(address(this)))));
        gasOracle.setOperator(address(this));
        gasOracle.setOperatorFeeConstant(operatorFeeConstant);
        gasOracle.setOperatorFeeScalar(operatorFeeScalar);
        _prepareTestData();

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
        for (uint256 i = 0; i < caseBedrock.length; i++) {
            assertEq(gasOracle.getL1GasUsed(caseBedrock[i].data), caseBedrock[i].gasUsed);
        }
        _setIsSkadi();
        for (uint256 i = 0; i < caseSkadi.length; i++) {
            assertEq(gasOracle.getL1GasUsed(caseSkadi[i].data), caseSkadi[i].gasUsed);
        }
    }

    function test_getL1Fee_succeeds() external {
        for (uint256 i = 0; i < caseBedrock.length; i++) {
            assertEq(gasOracle.getL1Fee(caseBedrock[i].data), caseBedrock[i].l1Fee);
        }
        _setIsSkadi();
        for (uint256 i = 0; i < caseSkadi.length; i++) {
            assertEq(gasOracle.getL1Fee(caseSkadi[i].data), caseSkadi[i].l1Fee);
        }
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

    function test_setIsSkadi_succeeds() external {
        gasOracle.setIsSkadi();
        assertEq(gasOracle.isSkadi(), true);
    }

    function test_setIsSkadi_reverts() external {
        gasOracle.setIsSkadi();
        vm.expectRevert("isSkadi already set");
        gasOracle.setIsSkadi();
    }

    function test_getOverhead_reverts() external {
        _setIsSkadi();
        vm.expectRevert("GasPriceOracle: overhead() is deprecated");
        gasOracle.overhead();
    }

    function test_getScalar_reverts() external {
        _setIsSkadi();
        vm.expectRevert("GasPriceOracle: scalar() is deprecated");
        gasOracle.scalar();
    }

    function test_setOperatorFeeConstant_succeeds() external {
        gasOracle.setOperatorFeeConstant(100);
        assertEq(gasOracle.operatorFeeConstant(), 100);
        _setIsSkadi();
        gasOracle.setOperatorFeeConstant(200);
        assertEq(gasOracle.operatorFeeConstant(), 200);
    }

    function test_setOperatorFeeScalar_succeeds() external {
        gasOracle.setOperatorFeeScalar(100);
        assertEq(gasOracle.operatorFeeScalar(), 100);
        _setIsSkadi();
        gasOracle.setOperatorFeeScalar(200);
        assertEq(gasOracle.operatorFeeScalar(), 200);
    }

    function _setIsSkadi() internal {
        gasOracle.setIsSkadi();
    }

    function _prepareTestData() internal {
        caseBedrock.push(
            TestData({
                data: hex"0001020304",
                gasUsed: 4 + 16 * 4 + 68 * 16 + l1FeeOverhead,
                l1Fee: (4 + 16 * 4 + 68 * 16 + l1FeeOverhead) * basefee * l1FeeScalar / 1e6
            })
        );

        caseSkadi.push(
            TestData({
                data: hex"0001020304",
                gasUsed: 4 + 16 * 4 + 68 * 16 + l1FeeOverhead,
                l1Fee: (4 + 16 * 4 + 68 * 16 + l1FeeOverhead) * basefee * l1FeeScalar / 1e6
            })
        );
    }
}
