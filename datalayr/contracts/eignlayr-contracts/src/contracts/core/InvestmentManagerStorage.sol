// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC1155/IERC1155.sol";
import "../interfaces/IInvestmentManager.sol";
import "../interfaces/IInvestmentStrategy.sol";
import "../interfaces/IEigenLayrDelegation.sol";

/**
 * @title Storage variables for the `InvestmentManager` contract.
 * @author Layr Labs, Inc.
 * @notice This storage contract is separate from the logic to simplify the upgrade process.
 */
abstract contract InvestmentManagerStorage is IInvestmentManager {
    /// @notice The EIP-712 typehash for the contract's domain
    bytes32 public constant DOMAIN_TYPEHASH =
        keccak256("EIP712Domain(string name,uint256 chainId,address verifyingContract)");
    /// @notice The EIP-712 typehash for the deposit struct used by the contract
    bytes32 public constant DEPOSIT_TYPEHASH =
        keccak256("Deposit(address strategy,address token,uint256 amount,uint256 nonce,uint256 expiry)");
    /// @notice EIP-712 Domain separator
    bytes32 public DOMAIN_SEPARATOR;
    // staker => number of signed deposit nonce (used in depositIntoStrategyOnBehalfOf)
    mapping(address => uint256) public nonces;

    // maximum length of dynamic arrays in `investorStrats` mapping, for sanity's sake
    uint8 internal constant MAX_INVESTOR_STRATS_LENGTH = 32;

    // system contracts
    IEigenLayrDelegation public immutable delegation;

    // staker => InvestmentStrategy => number of shares which they currently hold
    mapping(address => mapping(IInvestmentStrategy => uint256)) public investorStratShares;
    // staker => array of strategies in which they have nonzero shares
    mapping(address => IInvestmentStrategy[]) public investorStrats;

    mapping(IInvestmentStrategy => bool) public strategyStorage;

    // staker => can withdraw from investmentManager contracts
    mapping(address => bool) public delegatorWithdrawWhiteList;

    constructor(IEigenLayrDelegation _delegation) {
        delegation = _delegation;
    }
}
