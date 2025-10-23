// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Semver } from "../universal/Semver.sol";
import { Predeploys } from "../libraries/Predeploys.sol";
import { L1Block } from "../L2/L1Block.sol";
import { Arithmetic } from "../libraries/Arithmetic.sol";

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
     * @notice Set chain to be Arsia chain (callable by owner)
     */
    function setArsia() external onlyOwner {
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
     * @notice Computes the amount of L1 gas used for a transaction. Adds the overhead which
     *         represents the per-transaction gas overhead of posting the transaction and state
     *         roots to L1. Adds 68 bytes of padding to account for the fact that the input does
     *         not have a signature.
     *
     * @param _data Unsigned fully RLP-encoded transaction to get the L1 gas for.
     *
     * @return Amount of L1 gas used to publish the transaction.
     */
    function getL1GasUsed(bytes memory _data) public view returns (uint256) {
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
        uint256 l1GasUsed = _getCalldataGas(_data);
        uint256 scaledBaseFee = baseFeeScalar() * 16 * l1BaseFee();
        uint256 scaledBlobBaseFee = blobBaseFeeScalar() * blobBaseFee();
        uint256 fee = l1GasUsed * (scaledBaseFee + scaledBlobBaseFee);
        return fee / (16 * 10 ** DECIMALS);
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
}
