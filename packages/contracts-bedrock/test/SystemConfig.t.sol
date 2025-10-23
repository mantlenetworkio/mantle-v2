// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { CommonTest } from "./CommonTest.t.sol";
import { SystemConfig } from "src/L1/SystemConfig.sol";
import { ResourceMetering } from "src/L1/ResourceMetering.sol";
import { Constants } from "src/libraries/Constants.sol";

contract SystemConfig_Init is CommonTest {
    SystemConfig sysConf;

    function setUp() public virtual override {
        super.setUp();

        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 8,
            minimumBaseFee: 1 gwei,
            systemTxMaxGas: 1_000_000,
            maximumBaseFee: type(uint128).max
        });

        sysConf = new SystemConfig({
            _owner: alice,
            _basefeeScalar: 0,
            _blobbasefeeScalar: 0,
            _batcherHash: bytes32(hex"abcd"),
            _gasLimit: 30_000_000,
            _baseFee: 1_000_000_000,
            _unsafeBlockSigner: address(1),
            _config: config
        });
    }
}

contract SystemConfig_Initialize_TestFail is SystemConfig_Init {
    function test_initialize_lowGasLimit_reverts() external {
        uint64 minimumGasLimit = sysConf.minimumGasLimit();

        ResourceMetering.ResourceConfig memory cfg = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 8,
            minimumBaseFee: 1 gwei,
            systemTxMaxGas: 1_000_000,
            maximumBaseFee: type(uint128).max
        });

        vm.expectRevert("SystemConfig: gas limit too low");
        new SystemConfig({
            _owner: alice,
            _basefeeScalar: 0,
            _blobbasefeeScalar: 0,
            _batcherHash: bytes32(hex""),
            _gasLimit: minimumGasLimit - 1,
            _baseFee: 1_000_000_000,
            _unsafeBlockSigner: address(1),
            _config: cfg
        });
    }
}

contract SystemConfig_Setters_TestFail is SystemConfig_Init {
    function test_setBatcherHash_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setBatcherHash(bytes32(hex""));
    }

    function test_setGasConfig_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setGasConfig(0, 0);
    }

    function test_setGasConfigArsia_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setGasConfigArsia(1000, 2000);
    }

    function test_setOperatorFeeScalars_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setOperatorFeeScalars(500000, 1000);
    }

    function test_setGasLimit_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setGasLimit(0);
    }

    function test_setUnsafeBlockSigner_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setUnsafeBlockSigner(address(0x20));
    }

    function test_setBaseFee_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setBaseFee(1 gwei);
    }

    function test_setEIP1559Params_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setEIP1559Params(8, 10);
    }

    function test_setMinBaseFee_notOwner_reverts() external {
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setMinBaseFee(1 gwei);
    }

    function test_setResourceConfig_notOwner_reverts() external {
        ResourceMetering.ResourceConfig memory config = Constants.DEFAULT_RESOURCE_CONFIG();
        vm.expectRevert("Ownable: caller is not the owner");
        sysConf.setResourceConfig(config);
    }

    function test_setResourceConfig_badMinMax_reverts() external {
        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 8,
            systemTxMaxGas: 1_000_000,
            minimumBaseFee: 2 gwei,
            maximumBaseFee: 1 gwei
        });
        vm.prank(sysConf.owner());
        vm.expectRevert("SystemConfig: min base fee must be less than max base");
        sysConf.setResourceConfig(config);
    }

    function test_setResourceConfig_zeroDenominator_reverts() external {
        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 0,
            systemTxMaxGas: 1_000_000,
            minimumBaseFee: 1 gwei,
            maximumBaseFee: 2 gwei
        });
        vm.prank(sysConf.owner());
        vm.expectRevert("SystemConfig: denominator must be larger than 1");
        sysConf.setResourceConfig(config);
    }

    function test_setResourceConfig_lowGasLimit_reverts() external {
        uint64 gasLimit = sysConf.gasLimit();

        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: uint32(gasLimit),
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 8,
            systemTxMaxGas: uint32(gasLimit),
            minimumBaseFee: 1 gwei,
            maximumBaseFee: 2 gwei
        });
        vm.prank(sysConf.owner());
        vm.expectRevert("SystemConfig: gas limit too low");
        sysConf.setResourceConfig(config);
    }

    function test_setResourceConfig_badPrecision_reverts() external {
        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 11,
            baseFeeMaxChangeDenominator: 8,
            systemTxMaxGas: 1_000_000,
            minimumBaseFee: 1 gwei,
            maximumBaseFee: 2 gwei
        });
        vm.prank(sysConf.owner());
        vm.expectRevert("SystemConfig: precision loss with target resource limit");
        sysConf.setResourceConfig(config);
    }
}

contract SystemConfig_Setters_Test is SystemConfig_Init {
    event ConfigUpdate(uint256 indexed version, SystemConfig.UpdateType indexed updateType, bytes data);

    function testFuzz_setBatcherHash_succeeds(bytes32 newBatcherHash) external {
        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.BATCHER, abi.encode(newBatcherHash));

        vm.prank(sysConf.owner());
        sysConf.setBatcherHash(newBatcherHash);
        assertEq(sysConf.batcherHash(), newBatcherHash);
    }

    function testFuzz_setGasConfig_succeeds(uint256 newOverhead, uint256 newScalar) external {
        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.FEE_SCALARS, abi.encode(newOverhead, newScalar));

        vm.prank(sysConf.owner());
        sysConf.setGasConfig(newOverhead, newScalar);
        assertEq(sysConf.overhead(), newOverhead);
        assertEq(sysConf.scalar(), newScalar);
    }

    function testFuzz_setGasLimit_succeeds(uint64 newGasLimit) external {
        uint64 minimumGasLimit = sysConf.minimumGasLimit();
        newGasLimit = uint64(bound(uint256(newGasLimit), uint256(minimumGasLimit), uint256(type(uint64).max)));

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.GAS_LIMIT, abi.encode(newGasLimit));

        vm.prank(sysConf.owner());
        sysConf.setGasLimit(newGasLimit);
        assertEq(sysConf.gasLimit(), newGasLimit);
    }

    function testFuzz_setUnsafeBlockSigner_succeeds(address newUnsafeSigner) external {
        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.UNSAFE_BLOCK_SIGNER, abi.encode(newUnsafeSigner));

        vm.prank(sysConf.owner());
        sysConf.setUnsafeBlockSigner(newUnsafeSigner);
        assertEq(sysConf.unsafeBlockSigner(), newUnsafeSigner);
    }

    function testFuzz_setBaseFee_succeeds(uint256 newBaseFee) external {
        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.BASE_FEE, abi.encode(newBaseFee));

        vm.prank(sysConf.owner());
        sysConf.setBaseFee(newBaseFee);
        assertEq(sysConf.baseFee(), newBaseFee);
    }
}

contract SystemConfig_ArsiaSetters_Test is SystemConfig_Init {
    event ConfigUpdate(uint256 indexed version, SystemConfig.UpdateType indexed updateType, bytes data);

    function test_setGasConfigArsia_succeeds() external {
        uint32 newBasefeeScalar = 1000;
        uint32 newBlobbasefeeScalar = 2000;

        // Calculate expected packed scalar value
        uint256 expectedScalar = (uint256(0x01) << 248) | (uint256(newBlobbasefeeScalar) << 32) | newBasefeeScalar;

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.FEE_SCALARS, abi.encode(sysConf.overhead(), expectedScalar));

        vm.prank(sysConf.owner());
        sysConf.setGasConfigArsia(newBasefeeScalar, newBlobbasefeeScalar);

        assertEq(sysConf.basefeeScalar(), newBasefeeScalar);
        assertEq(sysConf.blobbasefeeScalar(), newBlobbasefeeScalar);
        assertEq(sysConf.scalar(), expectedScalar);
    }

    function testFuzz_setGasConfigArsia_succeeds(uint32 newBasefeeScalar, uint32 newBlobbasefeeScalar) external {
        // Calculate expected packed scalar value
        uint256 expectedScalar = (uint256(0x01) << 248) | (uint256(newBlobbasefeeScalar) << 32) | newBasefeeScalar;

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.FEE_SCALARS, abi.encode(sysConf.overhead(), expectedScalar));

        vm.prank(sysConf.owner());
        sysConf.setGasConfigArsia(newBasefeeScalar, newBlobbasefeeScalar);

        assertEq(sysConf.basefeeScalar(), newBasefeeScalar);
        assertEq(sysConf.blobbasefeeScalar(), newBlobbasefeeScalar);
        assertEq(sysConf.scalar(), expectedScalar);
    }

    function test_setOperatorFeeScalars_succeeds() external {
        uint32 newOperatorFeeScalar = 500000; // 0.5 in 1e6
        uint64 newOperatorFeeConstant = 1000;

        // Calculate expected packed value
        uint256 expectedPacked = (uint256(newOperatorFeeScalar) << 64) | newOperatorFeeConstant;

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.OPERATOR_FEE_PARAMS, abi.encode(expectedPacked));

        vm.prank(sysConf.owner());
        sysConf.setOperatorFeeScalars(newOperatorFeeScalar, newOperatorFeeConstant);

        assertEq(sysConf.operatorFeeScalar(), newOperatorFeeScalar);
        assertEq(sysConf.operatorFeeConstant(), newOperatorFeeConstant);
    }

    function testFuzz_setOperatorFeeScalars_succeeds(
        uint32 newOperatorFeeScalar,
        uint64 newOperatorFeeConstant
    )
        external
    {
        // Calculate expected packed value
        uint256 expectedPacked = (uint256(newOperatorFeeScalar) << 64) | newOperatorFeeConstant;

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.OPERATOR_FEE_PARAMS, abi.encode(expectedPacked));

        vm.prank(sysConf.owner());
        sysConf.setOperatorFeeScalars(newOperatorFeeScalar, newOperatorFeeConstant);

        assertEq(sysConf.operatorFeeScalar(), newOperatorFeeScalar);
        assertEq(sysConf.operatorFeeConstant(), newOperatorFeeConstant);
    }

    function test_gasConfigArsia_scalarPacking_succeeds() external {
        // Test edge cases for scalar packing
        uint32 basefeeScalar = type(uint32).max;
        uint32 blobbasefeeScalar = type(uint32).max;

        vm.prank(sysConf.owner());
        sysConf.setGasConfigArsia(basefeeScalar, blobbasefeeScalar);

        // Verify unpacking
        uint256 packedScalar = sysConf.scalar();
        uint8 version = uint8(packedScalar >> 248);
        uint32 unpackedBlobbasefee = uint32(packedScalar >> 32);
        uint32 unpackedBasefee = uint32(packedScalar);

        assertEq(version, 1, "Version should be 1");
        assertEq(unpackedBlobbasefee, blobbasefeeScalar, "Blobbasefee scalar mismatch");
        assertEq(unpackedBasefee, basefeeScalar, "Basefee scalar mismatch");
    }

    function test_operatorFeeScalars_zeroes_succeeds() external {
        vm.prank(sysConf.owner());
        sysConf.setOperatorFeeScalars(0, 0);

        assertEq(sysConf.operatorFeeScalar(), 0);
        assertEq(sysConf.operatorFeeConstant(), 0);
    }

    function test_operatorFeeScalars_maxValues_succeeds() external {
        uint32 maxScalar = type(uint32).max;
        uint64 maxConstant = type(uint64).max;

        vm.prank(sysConf.owner());
        sysConf.setOperatorFeeScalars(maxScalar, maxConstant);

        assertEq(sysConf.operatorFeeScalar(), maxScalar);
        assertEq(sysConf.operatorFeeConstant(), maxConstant);
    }
}

contract SystemConfig_EIP1559Params_Test is SystemConfig_Init {
    event ConfigUpdate(uint256 indexed version, SystemConfig.UpdateType indexed updateType, bytes data);

    function test_setEIP1559Params_succeeds() external {
        uint32 newDenominator = 50;
        uint32 newElasticity = 10;

        uint256 expectedPacked = (uint256(newDenominator) << 32) | uint64(newElasticity);

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.EIP_1559_PARAMS, abi.encode(expectedPacked));

        vm.prank(sysConf.owner());
        sysConf.setEIP1559Params(newDenominator, newElasticity);

        assertEq(sysConf.eip1559Denominator(), newDenominator);
        assertEq(sysConf.eip1559Elasticity(), newElasticity);
    }

    function testFuzz_setEIP1559Params_succeeds(uint32 denominator, uint32 elasticity) external {
        vm.assume(denominator >= 1);
        vm.assume(elasticity >= 1);

        uint256 expectedPacked = (uint256(denominator) << 32) | uint64(elasticity);

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.EIP_1559_PARAMS, abi.encode(expectedPacked));

        vm.prank(sysConf.owner());
        sysConf.setEIP1559Params(denominator, elasticity);

        assertEq(sysConf.eip1559Denominator(), denominator);
        assertEq(sysConf.eip1559Elasticity(), elasticity);
    }

    function test_setEIP1559Params_zeroDenominator_reverts() external {
        vm.prank(sysConf.owner());
        vm.expectRevert("SystemConfig: denominator must be >= 1");
        sysConf.setEIP1559Params(0, 10);
    }

    function test_setEIP1559Params_zeroElasticity_reverts() external {
        vm.prank(sysConf.owner());
        vm.expectRevert("SystemConfig: elasticity must be >= 1");
        sysConf.setEIP1559Params(8, 0);
    }

    function test_setMinBaseFee_succeeds() external {
        uint64 newMinBaseFee = 1 gwei;

        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.MIN_BASE_FEE, abi.encode(newMinBaseFee));

        vm.prank(sysConf.owner());
        sysConf.setMinBaseFee(newMinBaseFee);

        assertEq(sysConf.minBaseFee(), newMinBaseFee);
    }

    function testFuzz_setMinBaseFee_succeeds(uint64 newMinBaseFee) external {
        vm.expectEmit(true, true, true, true);
        emit ConfigUpdate(0, SystemConfig.UpdateType.MIN_BASE_FEE, abi.encode(newMinBaseFee));

        vm.prank(sysConf.owner());
        sysConf.setMinBaseFee(newMinBaseFee);

        assertEq(sysConf.minBaseFee(), newMinBaseFee);
    }

    function test_setMinBaseFee_zero_succeeds() external {
        vm.prank(sysConf.owner());
        sysConf.setMinBaseFee(0);

        assertEq(sysConf.minBaseFee(), 0);
    }
}

contract SystemConfig_Getters_Test is SystemConfig_Init {
    function test_resourceConfig_succeeds() external view {
        ResourceMetering.ResourceConfig memory config = sysConf.resourceConfig();
        assertEq(config.maxResourceLimit, 20_000_000);
        assertEq(config.elasticityMultiplier, 10);
        assertEq(config.baseFeeMaxChangeDenominator, 8);
        assertEq(config.minimumBaseFee, 1 gwei);
        assertEq(config.systemTxMaxGas, 1_000_000);
        assertEq(config.maximumBaseFee, type(uint128).max);
    }

    function test_initialValues_succeeds() external view {
        assertEq(sysConf.owner(), alice);
        assertEq(sysConf.basefeeScalar(), 0);
        assertEq(sysConf.blobbasefeeScalar(), 0);
        assertEq(sysConf.batcherHash(), bytes32(hex"abcd"));
        assertEq(sysConf.gasLimit(), 30_000_000);
        assertEq(sysConf.baseFee(), 1_000_000_000);
        assertEq(sysConf.unsafeBlockSigner(), address(1));
        assertEq(sysConf.operatorFeeScalar(), 0);
        assertEq(sysConf.operatorFeeConstant(), 0);
    }

    function test_minimumGasLimit_succeeds() external view {
        uint64 expected = 20_000_000 + 1_000_000; // maxResourceLimit + systemTxMaxGas
        assertEq(sysConf.minimumGasLimit(), expected);
    }
}

contract SystemConfig_Initialize_WithScalars_Test is CommonTest {
    event ConfigUpdate(uint256 indexed version, SystemConfig.UpdateType indexed updateType, bytes data);

    function test_initialize_withNonZeroScalars_succeeds() external {
        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 8,
            minimumBaseFee: 1 gwei,
            systemTxMaxGas: 1_000_000,
            maximumBaseFee: type(uint128).max
        });

        uint32 basefeeScalar = 1000;
        uint32 blobbasefeeScalar = 2000;

        SystemConfig newConfig = new SystemConfig({
            _owner: alice,
            _basefeeScalar: basefeeScalar,
            _blobbasefeeScalar: blobbasefeeScalar,
            _batcherHash: bytes32(hex"abcd"),
            _gasLimit: 30_000_000,
            _baseFee: 1_000_000_000,
            _unsafeBlockSigner: address(1),
            _config: config
        });

        assertEq(newConfig.basefeeScalar(), basefeeScalar);
        assertEq(newConfig.blobbasefeeScalar(), blobbasefeeScalar);
    }

    function testFuzz_initialize_withScalars_succeeds(uint32 basefeeScalar, uint32 blobbasefeeScalar) external {
        ResourceMetering.ResourceConfig memory config = ResourceMetering.ResourceConfig({
            maxResourceLimit: 20_000_000,
            elasticityMultiplier: 10,
            baseFeeMaxChangeDenominator: 8,
            minimumBaseFee: 1 gwei,
            systemTxMaxGas: 1_000_000,
            maximumBaseFee: type(uint128).max
        });

        SystemConfig newConfig = new SystemConfig({
            _owner: alice,
            _basefeeScalar: basefeeScalar,
            _blobbasefeeScalar: blobbasefeeScalar,
            _batcherHash: bytes32(hex"abcd"),
            _gasLimit: 30_000_000,
            _baseFee: 1_000_000_000,
            _unsafeBlockSigner: address(1),
            _config: config
        });

        assertEq(newConfig.basefeeScalar(), basefeeScalar);
        assertEq(newConfig.blobbasefeeScalar(), blobbasefeeScalar);
        assertEq(newConfig.operatorFeeScalar(), 0);
        assertEq(newConfig.operatorFeeConstant(), 0);
    }
}
