// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Interfaces
// import { ISemver } from "interfaces/universal/ISemver.sol";
import { Semver } from "src/universal/Semver.sol";
import { L1Block } from "src/L2/L1Block.sol";

// Libraries
import { Predeploys } from "src/libraries/Predeploys.sol";
import { Constants } from "src/libraries/Constants.sol";
import { Arithmetic } from "src/libraries/Arithmetic.sol";
import { LibZip } from "@solady/utils/LibZip.sol";

/**
 * @custom:proxied
 * @custom:predeploy 0x420000000000000000000000000000000000000F
 * @title GasPriceOracle
 * @notice This contract maintains the variables responsible for computing the L1 portion of the
 *         total fee charged on L2. Before Bedrock, this contract held variables in state that were
 *         read during the state transition function to compute the L1 portion of the transaction
 *         fee. After Bedrock, this contract now simply proxies the L1Block contract, which has
 *         the values used to compute the L1 portion of the fee in its state.
 *
 *         The contract exposes an API that is useful for knowing how large the L1 portion of the
 *         transaction fee will be. The following events were deprecated with Bedrock:
 *         - event OverheadUpdated(uint256 overhead);
 *         - event ScalarUpdated(uint256 scalar);
 *         - event DecimalsUpdated(uint256 decimals);
 */
contract GasPriceOracle is Semver {
    /**
     * @notice Number of decimals used in the scalar.
     */
    uint256 public constant DECIMALS = 6;
    uint256 public tokenRatio;
    address public owner;
    address public operator;

    /**
     * @notice This is the intercept value for the linear regression used to estimate the final size of the
     *         compressed transaction.
     */
    int32 private constant COST_INTERCEPT = -42_585_600;

    /**
     * @notice This is the coefficient value for the linear regression used to estimate the final size of the
     *         compressed transaction.
     */
    uint32 private constant COST_FASTLZ_COEF = 836_500;

    /**
     * @notice This is the minimum bound for the fastlz to brotli size estimation. Any estimations below this
     *         are set to this value.
     */
    uint256 private constant MIN_TRANSACTION_SIZE = 100;

    /**
     * @notice Indicates whether the network uses Arsia gas calculation.
     */
    bool public isArsia;

    /**
     *
     * Events *
     *
     */
    event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);

    /**
     *
     * Modifiers *
     *
     */
    modifier onlyOwner() {
        require(owner == msg.sender, "Caller is not the owner");
        _;
    }

    modifier onlyOperator() {
        require(operator == msg.sender, "Caller is not the operator");
        _;
    }

    /**
     * @custom:semver 1.1.0
     */
    constructor() Semver(1, 1, 0) { }

    /**
     * Allows the owner to modify the operator.
     * @param _operator New operator
     */
    // slither-disable-next-line external-function
    function setOperator(address _operator) external onlyOwner {
        address previousOperator = operator;
        operator = _operator;
        emit OperatorUpdated(previousOperator, operator);
    }

    /**
     * @dev Transfers ownership of the contract to a new account (`_owner`).
     * Can only be called by the current owner.
     */
    function transferOwnership(address _owner) external onlyOwner {
        require(_owner != address(0), "new owner is the zero address");
        address previousOwner = owner;
        owner = _owner;
        emit OwnershipTransferred(previousOwner, owner);
    }

    /**
     * Allows the operator to modify the token ratio.
     * @param _tokenRatio New tokenRatio
     */
    // slither-disable-next-line external-function
    function setTokenRatio(uint256 _tokenRatio) external onlyOperator {
        uint256 previousTokenRatio = tokenRatio;
        tokenRatio = _tokenRatio;
        emit TokenRatioUpdated(previousTokenRatio, tokenRatio);
    }

    /**
     * @notice Set chain to be Arsia chain (callable by depositor account)
     */
    function setArsia() external {
        require(
            msg.sender == Constants.DEPOSITOR_ACCOUNT, "GasPriceOracle: only the depositor account can set isArsia flag"
        );
        require(isArsia == false, "GasPriceOracle: Arsia already active");
        isArsia = true;
    }

    /**
     * @notice Computes the L1 portion of the fee based on the size of the rlp encoded input
     *         transaction, the current L1 base fee, and the various dynamic parameters.
     *
     * @param _data Unsigned fully RLP-encoded transaction to get the L1 fee for.
     *
     * @return L1 fee that should be paid for the tx
     */
    function getL1Fee(bytes memory _data) external view returns (uint256) {
        if (isArsia) {
            return _getL1FeeArsia(_data);
        }
        return _getL1FeeBedrock(_data);
    }

    /**
     * @notice returns an upper bound for the L1 fee for a given transaction size.
     *         It is provided for callers who wish to estimate L1 transaction costs in the
     *         write path, and is much more gas efficient than `getL1Fee`.
     *         It assumes the worst case of fastlz upper-bound which covers %99.99 txs.
     *
     * @param _unsignedTxSize Unsigned fully RLP-encoded transaction size to get the L1 fee for.
     *
     * @return L1 estimated upper-bound fee that should be paid for the tx
     */
    function getL1FeeUpperBound(uint256 _unsignedTxSize) external view returns (uint256) {
        require(isArsia, "GasPriceOracle: getL1FeeUpperBound only supports Arsia");

        // Add 68 to the size to account for unsigned tx:
        uint256 txSize = _unsignedTxSize + 68;
        // txSize / 255 + 16 is the practical fastlz upper-bound covers %99.99 txs.
        uint256 flzUpperBound = txSize + txSize / 255 + 16;

        return _arsiaL1Cost(flzUpperBound);
    }

    /**
     * @notice Retrieves the current gas price (base fee).
     *
     * @return Current L2 gas price (base fee).
     */
    function gasPrice() public view returns (uint256) {
        return block.basefee;
    }

    /**
     * @notice Retrieves the current base fee.
     *
     * @return Current L2 base fee.
     */
    function baseFee() public view returns (uint256) {
        return block.basefee;
    }

    /**
     * @notice Retrieves the current fee overhead.
     *
     * @return Current fee overhead.
     */
    function overhead() public view returns (uint256) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).l1FeeOverhead();
    }

    /**
     * @notice Retrieves the current fee scalar.
     *
     * @return Current fee scalar.
     */
    function scalar() public view returns (uint256) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).l1FeeScalar();
    }

    /**
     * @notice Retrieves the latest known L1 base fee.
     *
     * @return Latest known L1 base fee.
     */
    function l1BaseFee() public view returns (uint256) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).basefee();
    }

    /**
     * @notice Retrieves the current blob base fee.
     *
     * @return Current blob base fee.
     */
    function blobBaseFee() public view returns (uint256) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).blobBaseFee();
    }

    /**
     * @notice Retrieves the current base fee scalar.
     *
     * @return Current base fee scalar.
     */
    function baseFeeScalar() public view returns (uint32) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).baseFeeScalar();
    }

    /**
     * @notice Retrieves the current blob base fee scalar.
     *
     * @return Current blob base fee scalar.
     */
    function blobBaseFeeScalar() public view returns (uint32) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).blobBaseFeeScalar();
    }

    /**
     * @notice Retrieves the operator fee scalar.
     *
     * @return Operator fee scalar.
     */
    function operatorFeeScalar() public view returns (uint32) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).operatorFeeScalar();
    }

    /**
     * @notice Retrieves the operator fee constant.
     *
     * @return Operator fee constant.
     */
    function operatorFeeConstant() public view returns (uint64) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).operatorFeeConstant();
    }

    /**
     * @notice Computes the amount of L1 gas used for a transaction. Adds 68 bytes
     *         of padding to account for the fact that the input does not have a signature.
     *
     * @param _data Unsigned fully RLP-encoded transaction to get the L1 gas for.
     *
     * @return Amount of L1 gas used to publish the transaction.
     *
     * @custom:deprecated This method does not accurately estimate the gas used for a transaction.
     *                    If you are calculating fees use getL1Fee or getL1FeeUpperBound.
     */
    function getL1GasUsed(bytes memory _data) public view returns (uint256) {
        if (isArsia) {
            // Add 68 to the size to account for unsigned tx
            // Assume the compressed data is mostly non-zero, and would pay 16 gas per calldata byte
            // Divide by 1e6 due to the scaling factor of the linear regression
            return _arsiaLinearRegression(LibZip.flzCompress(_data).length + 68) * 16 / 1e6;
        }
        uint256 l1GasUsed = _getCalldataGas(_data);
        return l1GasUsed + overhead();
    }

    function getOperatorFee(uint256 _gasUsed) public view returns (uint256) {
        if (!isArsia) {
            return 0;
        }
        return Arithmetic.saturatingAdd(
            Arithmetic.saturatingMul(_gasUsed, operatorFeeScalar()) / 1e6, operatorFeeConstant()
        );
    }

    /**
     * @custom:legacy
     * @notice Retrieves the number of decimals used in the scalar.
     *
     * @return Number of decimals used in the scalar.
     */
    function decimals() public pure returns (uint256) {
        return DECIMALS;
    }

    /**
     * @notice Computes the L1 portion of the fee for Bedrock.
     *
     * @param _data Unsigned fully RLP-encoded transaction to get the L1 fee for.
     *
     * @return L1 fee that should be paid for the tx
     */
    function _getL1FeeBedrock(bytes memory _data) internal view returns (uint256) {
        uint256 l1GasUsed = getL1GasUsed(_data);
        uint256 l1Fee = l1GasUsed * l1BaseFee();
        uint256 divisor = 10 ** DECIMALS;
        uint256 unscaled = l1Fee * scalar();
        uint256 scaled = unscaled / divisor;
        return scaled;
    }

    /**
     * @notice Computes the L1 portion of the fee for Arsia.
     *
     * @param _data Unsigned fully RLP-encoded transaction to get the L1 fee for.
     *
     * @return L1 fee that should be paid for the tx
     */
    function _getL1FeeArsia(bytes memory _data) internal view returns (uint256) {
        return _arsiaL1Cost(LibZip.flzCompress(_data).length + 68);
    }

    /**
     * @notice Computes the amount of L1 gas used for a transaction. Adds 68 bytes
     *         of padding to account for the fact that the input does not have a signature.
     *
     * @param _data Unsigned fully RLP-encoded transaction to get the L1 gas for.
     *
     * @return Amount of L1 gas used to publish the transaction.
     */
    function _getCalldataGas(bytes memory _data) internal pure returns (uint256) {
        uint256 total = 0;
        uint256 length = _data.length;
        for (uint256 i = 0; i < length; i++) {
            if (_data[i] == 0) {
                total += 4;
            } else {
                total += 16;
            }
        }
        return total + (68 * 16);
    }

    /**
     * @notice Arsia L1 cost based on the compressed and original tx size.
     *
     * @param _fastLzSize estimated compressed tx size.
     *
     * @return Arsia L1 fee that should be paid for the tx
     */
    function _arsiaL1Cost(uint256 _fastLzSize) internal view returns (uint256) {
        // Apply the linear regression to estimate the Brotli 10 size
        uint256 estimatedSize = _arsiaLinearRegression(_fastLzSize);
        uint256 feeScaled = baseFeeScalar() * 16 * l1BaseFee() + blobBaseFeeScalar() * blobBaseFee();
        return estimatedSize * feeScaled / (10 ** (DECIMALS * 2));
    }

    /**
     * @notice Takes the fastLz size compression and returns the estimated Brotli
     *
     * @param _fastLzSize fastlz compressed tx size.
     *
     * @return Number of bytes in the compressed transaction
     */
    function _arsiaLinearRegression(uint256 _fastLzSize) internal pure returns (uint256) {
        int256 estimatedSize = COST_INTERCEPT + int256(COST_FASTLZ_COEF * _fastLzSize);
        if (estimatedSize < int256(MIN_TRANSACTION_SIZE) * 1e6) {
            estimatedSize = int256(MIN_TRANSACTION_SIZE) * 1e6;
        }
        return uint256(estimatedSize);
    }
}
