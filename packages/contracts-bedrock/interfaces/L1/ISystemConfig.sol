// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { IResourceMetering } from "interfaces/L1/IResourceMetering.sol";
/// @title ISystemConfig
/// @notice Interface for the SystemConfig contract

interface ISystemConfig {
    /// @notice Constructor function with the same parameters as SystemConfig
    /// @param _owner             Initial owner of the contract
    /// @param _basefeeScalar     Initial basefee scalar value
    /// @param _blobbasefeeScalar Initial blobbasefee scalar value
    /// @param _batcherHash       Initial batcher hash
    /// @param _gasLimit          Initial gas limit
    /// @param _baseFee           Initial base fee
    /// @param _unsafeBlockSigner Initial unsafe block signer address
    /// @param _config            Initial resource config
    function __constructor__(
        address _owner,
        uint32 _basefeeScalar,
        uint32 _blobbasefeeScalar,
        bytes32 _batcherHash,
        uint64 _gasLimit,
        uint256 _baseFee,
        address _unsafeBlockSigner,
        IResourceMetering.ResourceConfig memory _config
    )
        external;

    function initialize(
        address _owner,
        uint32 _basefeeScalar,
        uint32 _blobbasefeeScalar,
        bytes32 _batcherHash,
        uint64 _gasLimit,
        uint256 _baseFee,
        address _unsafeBlockSigner,
        IResourceMetering.ResourceConfig memory _config
    )
        external;

    function VERSION() external view returns (uint256);
    function UNSAFE_BLOCK_SIGNER_SLOT() external view returns (bytes32);
    function owner() external view returns (address);
    function overhead() external view returns (uint256);
    function scalar() external view returns (uint256);
    function batcherHash() external view returns (bytes32);
    function gasLimit() external view returns (uint64);
    function baseFee() external view returns (uint256);
    function eip1559Denominator() external view returns (uint32);
    function eip1559Elasticity() external view returns (uint32);
    function minBaseFee() external view returns (uint64);
    function basefeeScalar() external view returns (uint32);
    function blobbasefeeScalar() external view returns (uint32);
    function operatorFeeScalar() external view returns (uint32);
    function operatorFeeConstant() external view returns (uint64);
    function unsafeBlockSigner() external view returns (address);
    function resourceConfig() external view returns (IResourceMetering.ResourceConfig memory);
    function minimumGasLimit() external view returns (uint64);
    function setUnsafeBlockSigner(address _unsafeBlockSigner) external;
    function setBatcherHash(bytes32 _batcherHash) external;
    function setGasConfig(uint256 _overhead, uint256 _scalar) external;
    function setGasLimit(uint64 _gasLimit) external;
    function setBaseFee(uint256 _baseFee) external;
    function setEIP1559Params(uint32 _denominator, uint32 _elasticity) external;
    function setMinBaseFee(uint64 _minBaseFee) external;
    function setGasConfigArsia(uint32 _basefeeScalar, uint32 _blobbasefeeScalar) external;
    function setOperatorFeeScalars(uint32 _operatorFeeScalar, uint64 _operatorFeeConstant) external;
    function setResourceConfig(IResourceMetering.ResourceConfig memory _config) external;
}
