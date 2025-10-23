// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { Semver } from "../universal/Semver.sol";
import { ResourceMetering } from "./ResourceMetering.sol";

/**
 * @title SystemConfig
 * @notice The SystemConfig contract is used to manage configuration of an Optimism network. All
 *         configuration is stored on L1 and picked up by L2 as part of the derviation of the L2
 *         chain.
 */
contract SystemConfig is OwnableUpgradeable, Semver {
    /**
     * @notice Enum representing different types of updates.
     *
     * @custom:value BATCHER              Represents an update to the batcher hash.
     * @custom:value FEE_SCALARS          Represents an update to l1 data fee scalars.
     * @custom:value GAS_LIMIT            Represents an update to gas limit on L2.
     * @custom:value UNSAFE_BLOCK_SIGNER  Represents an update to the signer key for unsafe
     *                                    block distrubution.
     * @custom:value BASE_FEE             Represents an update to L2 base fee.
     * @custom:value EIP_1559_PARAMS      Represents an update to EIP-1559 parameters.
     * @custom:value OPERATOR_FEE_PARAMS  Represents an update to operator fee parameters.
     * @custom:value MIN_BASE_FEE         Represents an update to the minimum base fee.
     */
    enum UpdateType {
        BATCHER, // Batcher submitter address
        FEE_SCALARS, // L1 base fee and blob fee scalars
        GAS_LIMIT, // L2 gas limit
        UNSAFE_BLOCK_SIGNER, // L2 sequencer signer
        BASE_FEE, // L2 base fee
        EIP_1559_PARAMS, // EIP-1559 parameters
        OPERATOR_FEE_PARAMS, // Operator fee scalar and constant
        MIN_BASE_FEE // Minimum base fee

    }

    /**
     * @notice Version identifier, used for upgrades.
     */
    uint256 public constant VERSION = 0;

    /**
     * @notice Storage slot that the unsafe block signer is stored at. Storing it at this
     *         deterministic storage slot allows for decoupling the storage layout from the way
     *         that `solc` lays out storage. The `op-node` uses a storage proof to fetch this value.
     */
    bytes32 public constant UNSAFE_BLOCK_SIGNER_SLOT = keccak256("systemconfig.unsafeblocksigner");

    /**
     * @notice Fixed L2 gas overhead. Used as part of the L2 fee calculation.
     */
    uint256 public overhead;

    /**
     * @notice Dynamic L2 gas overhead. Used as part of the L2 fee calculation.
     */
    uint256 public scalar;

    /**
     * @notice Identifier for the batcher. For version 1 of this configuration, this is represented
     *         as an address left-padded with zeros to 32 bytes.
     */
    bytes32 public batcherHash;

    /**
     * @notice L2 block gas limit.
     */
    uint64 public gasLimit;

    /**
     * @notice The configuration for the deposit fee market. Used by the OptimismPortal
     *         to meter the cost of buying L2 gas on L1. Set as internal and wrapped with a getter
     *         so that the struct is returned instead of a tuple.
     */
    ResourceMetering.ResourceConfig internal _resourceConfig;

    /**
     * @notice L2 block base fee.
     */
    uint256 public baseFee;

    /**
     * @notice Basefee scalar value. Part of the L2 fee calculation.
     */
    uint32 public basefeeScalar;

    /**
     * @notice Blobbasefee scalar value. Part of the L2 fee calculation.
     */
    uint32 public blobbasefeeScalar;

    /**
     * @notice The EIP-1559 base fee max change denominator.
     */
    uint32 public eip1559Denominator;

    /**
     * @notice The EIP-1559 elasticity multiplier.
     */
    uint32 public eip1559Elasticity;

    /**
     * @notice The operator fee scalar.
     */
    uint32 public operatorFeeScalar;

    /**
     * @notice The operator fee constant.
     */
    uint64 public operatorFeeConstant;

    /**
     * @notice The minimum base fee, in wei.
     */
    uint64 public minBaseFee;

    /**
     * @notice Emitted when configuration is updated
     *
     * @param version    SystemConfig version.
     * @param updateType Type of update.
     * @param data       Encoded update data.
     */
    event ConfigUpdate(uint256 indexed version, UpdateType indexed updateType, bytes data);

    /**
     * @custom:semver 1.4.0
     *
     * @param _owner             Initial owner of the contract.
     * @param _basefeeScalar     Initial basefee scalar value.
     * @param _blobbasefeeScalar Initial blobbasefee scalar value.
     * @param _batcherHash       Initial batcher hash.
     * @param _gasLimit          Initial gas limit.
     * @param _unsafeBlockSigner Initial unsafe block signer address.
     * @param _config            Initial resource config.
     */
    constructor(
        address _owner,
        uint32 _basefeeScalar,
        uint32 _blobbasefeeScalar,
        bytes32 _batcherHash,
        uint64 _gasLimit,
        uint256 _baseFee,
        address _unsafeBlockSigner,
        ResourceMetering.ResourceConfig memory _config
    )
        Semver(1, 4, 0)
    {
        initialize({
            _owner: _owner,
            _basefeeScalar: _basefeeScalar,
            _blobbasefeeScalar: _blobbasefeeScalar,
            _batcherHash: _batcherHash,
            _gasLimit: _gasLimit,
            _baseFee: _baseFee,
            _unsafeBlockSigner: _unsafeBlockSigner,
            _config: _config
        });
    }

    /**
     * @notice Initializer. The resource config must be set before the
     *         require check.
     *
     * @param _owner             Initial owner of the contract.
     * @param _basefeeScalar     Initial basefee scalar value.
     * @param _blobbasefeeScalar Initial blobbasefee scalar value.
     * @param _batcherHash       Initial batcher hash.
     * @param _gasLimit          Initial gas limit.
     * @param _unsafeBlockSigner Initial unsafe block signer address.
     * @param _config            Initial ResourceConfig.
     */
    function initialize(
        address _owner,
        uint32 _basefeeScalar,
        uint32 _blobbasefeeScalar,
        bytes32 _batcherHash,
        uint64 _gasLimit,
        uint256 _baseFee,
        address _unsafeBlockSigner,
        ResourceMetering.ResourceConfig memory _config
    )
        public
        initializer
    {
        __Ownable_init();
        transferOwnership(_owner);
        batcherHash = _batcherHash;
        gasLimit = _gasLimit;
        baseFee = _baseFee;
        basefeeScalar = _basefeeScalar;
        blobbasefeeScalar = _blobbasefeeScalar;
        _setUnsafeBlockSigner(_unsafeBlockSigner);
        _setResourceConfig(_config);
        require(_gasLimit >= minimumGasLimit(), "SystemConfig: gas limit too low");
    }

    /**
     * @notice Returns the minimum L2 gas limit that can be safely set for the system to
     *         operate. The L2 gas limit must be larger than or equal to the amount of
     *         gas that is allocated for deposits per block plus the amount of gas that
     *         is allocated for the system transaction.
     *         This function is used to determine if changes to parameters are safe.
     *
     * @return uint64
     */
    function minimumGasLimit() public view returns (uint64) {
        return uint64(_resourceConfig.maxResourceLimit) + uint64(_resourceConfig.systemTxMaxGas);
    }

    /**
     * @notice High level getter for the unsafe block signer address. Unsafe blocks can be
     *         propagated across the p2p network if they are signed by the key corresponding to
     *         this address.
     *
     * @return Address of the unsafe block signer.
     */
    // solhint-disable-next-line ordering
    function unsafeBlockSigner() external view returns (address) {
        address addr;
        bytes32 slot = UNSAFE_BLOCK_SIGNER_SLOT;
        assembly {
            addr := sload(slot)
        }
        return addr;
    }

    /**
     * @notice Updates the unsafe block signer address.
     *
     * @param _unsafeBlockSigner New unsafe block signer address.
     */
    function setUnsafeBlockSigner(address _unsafeBlockSigner) external onlyOwner {
        _setUnsafeBlockSigner(_unsafeBlockSigner);

        bytes memory data = abi.encode(_unsafeBlockSigner);
        emit ConfigUpdate(VERSION, UpdateType.UNSAFE_BLOCK_SIGNER, data);
    }

    /**
     * @notice Updates the batcher hash.
     *
     * @param _batcherHash New batcher hash.
     */
    function setBatcherHash(bytes32 _batcherHash) external onlyOwner {
        batcherHash = _batcherHash;

        bytes memory data = abi.encode(_batcherHash);
        emit ConfigUpdate(VERSION, UpdateType.BATCHER, data);
    }

    /**
     * @notice Updates gas config.
     *         Deprecated in favor of setGasConfigArsia since the Arsia upgrade.
     * @param _overhead New overhead value.
     * @param _scalar   New scalar value.
     */
    function setGasConfig(uint256 _overhead, uint256 _scalar) external onlyOwner {
        overhead = _overhead;
        scalar = _scalar;

        bytes memory data = abi.encode(_overhead, _scalar);
        emit ConfigUpdate(VERSION, UpdateType.FEE_SCALARS, data);
    }

    /**
     * @notice Updates the L2 gas limit.
     *
     * @param _gasLimit New gas limit.
     */
    function setGasLimit(uint64 _gasLimit) external onlyOwner {
        require(_gasLimit >= minimumGasLimit(), "SystemConfig: gas limit too low");
        gasLimit = _gasLimit;

        bytes memory data = abi.encode(_gasLimit);
        emit ConfigUpdate(VERSION, UpdateType.GAS_LIMIT, data);
    }

    /**
     * @notice Updates the L2 base fee.
     *
     * @param _baseFee New base fee.
     */
    function setBaseFee(uint256 _baseFee) external onlyOwner {
        baseFee = _baseFee;

        bytes memory data = abi.encode(_baseFee);
        emit ConfigUpdate(VERSION, UpdateType.BASE_FEE, data);
    }

    /// @notice Updates the EIP-1559 parameters of the chain. Can only be called by the owner.
    /// @param _denominator EIP-1559 base fee max change denominator.
    /// @param _elasticity  EIP-1559 elasticity multiplier.
    function setEIP1559Params(uint32 _denominator, uint32 _elasticity) external onlyOwner {
        _setEIP1559Params(_denominator, _elasticity);
    }

    /// @notice Internal function for updating the EIP-1559 parameters.
    function _setEIP1559Params(uint32 _denominator, uint32 _elasticity) internal {
        // require the parameters have sane values:
        require(_denominator >= 1, "SystemConfig: denominator must be >= 1");
        require(_elasticity >= 1, "SystemConfig: elasticity must be >= 1");
        eip1559Denominator = _denominator;
        eip1559Elasticity = _elasticity;

        bytes memory data = abi.encode(uint256(_denominator) << 32 | uint64(_elasticity));
        emit ConfigUpdate(VERSION, UpdateType.EIP_1559_PARAMS, data);
    }

    /// @notice Updates the minimum base fee. Can only be called by the owner.
    ///         Setting this value to 0 is equivalent to disabling the min base fee feature
    /// @param _minBaseFee New minimum base fee.
    function setMinBaseFee(uint64 _minBaseFee) external onlyOwner {
        _setMinBaseFee(_minBaseFee);
    }

    /// @notice Internal function for updating the minimum base fee.
    function _setMinBaseFee(uint64 _minBaseFee) internal {
        minBaseFee = _minBaseFee;
        bytes memory data = abi.encode(_minBaseFee);
        emit ConfigUpdate(VERSION, UpdateType.MIN_BASE_FEE, data);
    }

    /**
     * @notice Updates gas config for Arsia.
     *
     * @param _basefeeScalar     New basefeeScalar value.
     * @param _blobbasefeeScalar New blobbasefeeScalar value.
     */
    function setGasConfigArsia(uint32 _basefeeScalar, uint32 _blobbasefeeScalar) external onlyOwner {
        basefeeScalar = _basefeeScalar;
        blobbasefeeScalar = _blobbasefeeScalar;

        // Update the legacy scalar field for compatibility
        scalar = (uint256(0x01) << 248) | (uint256(_blobbasefeeScalar) << 32) | _basefeeScalar;

        bytes memory data = abi.encode(overhead, scalar);
        emit ConfigUpdate(VERSION, UpdateType.FEE_SCALARS, data);
    }

    /**
     * @notice Updates the operator fee parameters.
     *
     * @param _operatorFeeScalar   New operator fee scalar.
     * @param _operatorFeeConstant New operator fee constant.
     */
    function setOperatorFeeScalars(uint32 _operatorFeeScalar, uint64 _operatorFeeConstant) external onlyOwner {
        operatorFeeScalar = _operatorFeeScalar;
        operatorFeeConstant = _operatorFeeConstant;

        bytes memory data = abi.encode((uint256(_operatorFeeScalar) << 64) | _operatorFeeConstant);
        emit ConfigUpdate(VERSION, UpdateType.OPERATOR_FEE_PARAMS, data);
    }

    /**
     * @notice Low level setter for the unsafe block signer address. This function exists to
     *         deduplicate code around storing the unsafeBlockSigner address in storage.
     *
     * @param _unsafeBlockSigner New unsafeBlockSigner value.
     */
    function _setUnsafeBlockSigner(address _unsafeBlockSigner) internal {
        bytes32 slot = UNSAFE_BLOCK_SIGNER_SLOT;
        assembly {
            sstore(slot, _unsafeBlockSigner)
        }
    }

    /**
     * @notice A getter for the resource config. Ensures that the struct is
     *         returned instead of a tuple.
     *
     * @return ResourceConfig
     */
    function resourceConfig() external view returns (ResourceMetering.ResourceConfig memory) {
        return _resourceConfig;
    }

    /**
     * @notice An external setter for the resource config. In the future, this
     *         method may emit an event that the `op-node` picks up for when the
     *         resource config is changed.
     *
     * @param _config The new resource config values.
     */
    function setResourceConfig(ResourceMetering.ResourceConfig memory _config) external onlyOwner {
        _setResourceConfig(_config);
    }

    /**
     * @notice An internal setter for the resource config. Ensures that the
     *         config is sane before storing it by checking for invariants.
     *
     * @param _config The new resource config.
     */
    function _setResourceConfig(ResourceMetering.ResourceConfig memory _config) internal {
        // Min base fee must be less than or equal to max base fee.
        require(
            _config.minimumBaseFee <= _config.maximumBaseFee, "SystemConfig: min base fee must be less than max base"
        );
        // Base fee change denominator must be greater than 1.
        require(_config.baseFeeMaxChangeDenominator > 1, "SystemConfig: denominator must be larger than 1");
        // Max resource limit plus system tx gas must be less than or equal to the L2 gas limit.
        // The gas limit must be increased before these values can be increased.
        require(_config.maxResourceLimit + _config.systemTxMaxGas <= gasLimit, "SystemConfig: gas limit too low");
        // Elasticity multiplier must be greater than 0.
        require(_config.elasticityMultiplier > 0, "SystemConfig: elasticity multiplier cannot be 0");
        // No precision loss when computing target resource limit.
        require(
            ((_config.maxResourceLimit / _config.elasticityMultiplier) * _config.elasticityMultiplier)
                == _config.maxResourceLimit,
            "SystemConfig: precision loss with target resource limit"
        );

        _resourceConfig = _config;
    }
}
