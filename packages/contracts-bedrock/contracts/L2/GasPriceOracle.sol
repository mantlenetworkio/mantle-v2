// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Semver } from "../universal/Semver.sol";
import { Predeploys } from "../libraries/Predeploys.sol";
import { L1Block } from "../L2/L1Block.sol";
import { Arithmetic } from "../libraries/Arithmetic.sol";
import { Constants } from "../libraries/Constants.sol";

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
    uint256 public operatorFeeConstant;
    uint256 public operatorFeeScalar;

    uint256[10] public _gap;

    bool public isLimb;

    /**
     * @custom:semver 1.1.0
     */
    constructor() Semver(1, 1, 0) { }

    /**
     * Events
     */
    event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);
    event OperatorFeeConstantUpdated(
        uint256 indexed previousOperatorFeeConstant, uint256 indexed newOperatorFeeConstant
    );
    event OperatorFeeScalarUpdated(uint256 indexed previousOperatorFeeScalar, uint256 indexed newOperatorFeeScalar);

    /**
     * Modifiers
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
     * @notice Allows the owner to modify the operator.
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

    function setLimb() external {
        require(
            msg.sender == Constants.DEPOSITOR_ACCOUNT,
            "GasPriceOracle: only the depositor account can set isLimb flag"
        );
        require(isLimb == false, "GasPriceOracle: IsLimb already set");
        isLimb = true;
    }

    function setOperatorFeeConstant(uint256 _operatorFeeConstant) external onlyOperator {
        uint256 previousOperatorFeeConstant = operatorFeeConstant;
        operatorFeeConstant = _operatorFeeConstant;
        emit OperatorFeeConstantUpdated(previousOperatorFeeConstant, operatorFeeConstant);
    }

    function setOperatorFeeScalar(uint256 _operatorFeeScalar) external onlyOperator {
        uint256 previousOperatorFeeScalar = operatorFeeScalar;
        operatorFeeScalar = _operatorFeeScalar;
        emit OperatorFeeScalarUpdated(previousOperatorFeeScalar, operatorFeeScalar);
    }

    /// @notice Computes the L1 portion of the fee based on the size of the rlp encoded input
    ///         transaction, the current L1 base fee, and the various dynamic parameters.
    /// @param _data Unsigned fully RLP-encoded transaction to get the L1 fee for.
    /// @return L1 fee that should be paid for the tx
    function getL1Fee(bytes memory _data) external view returns (uint256) {
        return _getL1FeeBedrock(_data);
    }

    function getOperatorFee(uint256 _gasUsed) public view returns (uint256) {
        if (!isLimb) {
            return 0;
        }
        return
            Arithmetic.saturatingAdd(Arithmetic.saturatingMul(_gasUsed, operatorFeeScalar) / 1e6, operatorFeeConstant);
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
        require(!isLimb, "GasPriceOracle: overhead() is deprecated");
        return _overhead();
    }

    /**
     * @notice Retrieves the current fee scalar.
     *
     * @return Current fee scalar.
     */
    function scalar() public view returns (uint256) {
        require(!isLimb, "GasPriceOracle: scalar() is deprecated");
        return _scalar();
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
     * @custom:legacy
     * @notice Retrieves the number of decimals used in the scalar.
     *
     * @return Number of decimals used in the scalar.
     */
    function decimals() public pure returns (uint256) {
        return DECIMALS;
    }

    /// @notice Computes the amount of L1 gas used for a transaction. Adds 68 bytes
    ///         of padding to account for the fact that the input does not have a signature.
    /// @param _data Unsigned fully RLP-encoded transaction to get the L1 gas for.
    /// @return Amount of L1 gas used to publish the transaction.
    function getL1GasUsed(bytes memory _data) public view returns (uint256) {
        return _getCalldataGas(_data) + _overhead();
    }

    /// @notice L1 gas estimation calculation.
    /// @param _data Unsigned fully RLP-encoded transaction to get the L1 gas for.
    /// @return Amount of L1 gas used to publish the transaction.
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

    function _overhead() internal view returns (uint256) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).l1FeeOverhead();
    }

    function _scalar() internal view returns (uint256) {
        return L1Block(Predeploys.L1_BLOCK_ATTRIBUTES).l1FeeScalar();
    }

    /// @notice Computation of the L1 portion of the fee for Bedrock.
    /// @param _data Unsigned fully RLP-encoded transaction to get the L1 fee for.
    /// @return L1 fee that should be paid for the tx
    function _getL1FeeBedrock(bytes memory _data) internal view returns (uint256) {
        uint256 l1GasUsed = _getCalldataGas(_data);
        uint256 fee = (l1GasUsed + _overhead()) * l1BaseFee() * _scalar();
        return fee / (10 ** DECIMALS);
    }
}
