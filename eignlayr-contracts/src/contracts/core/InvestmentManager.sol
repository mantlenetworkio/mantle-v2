// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import "@openzeppelin-upgrades/contracts/security/ReentrancyGuardUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "../permissions/Pausable.sol";
import "./InvestmentManagerStorage.sol";
import "../interfaces/IServiceManager.sol";
import "../interfaces/IRegistryPermission.sol";

// import "forge-std/Test.sol";

/**
 * @title The primary entry- and exit-point for funds into and out of EigenLayr.
 * @author Layr Labs, Inc.
 * @notice This contract is for managing investments in different strategies. The main
 * functionalities are:
 * - adding and removing investment strategies that any delegator can invest into
 * - enabling deposit of assets into specified investment strategy(s)
 * - enabling removal of assets from specified investment strategy(s)
 * - recording deposit of ETH into settlement layer
 * - recording deposit of Eigen for securing EigenLayr
 */
contract InvestmentManager is
    Initializable,
    OwnableUpgradeable,
    ReentrancyGuardUpgradeable,
    InvestmentManagerStorage,
    Pausable
    // ,Test
{
    using SafeERC20 for IERC20;

    uint256 constant GWEI_TO_WEI = 1e9;

    uint8 internal constant PAUSED_DEPOSITS = 0;
    uint8 internal constant PAUSED_WITHDRAWALS = 1;

    /// @notice contract used for manage operator delegator permission
    IRegistryPermission public immutable permissionManager;

    /// @notice Emitted when a staker withdrawal
    event DelegatorWithdrawal(
        address indexed withdrawer,
        IInvestmentStrategy[] strategies,
        IERC20[] tokens,
        uint256[] shares
    );

    modifier onlyStakeAndDelegate() {
        require(
            permissionManager.getDelegatorPermission(msg.sender) == true, "InvestmentManager.depositIntoStrategy: staker has not permission to do this action"
        );
        _;
    }

    /**
     * @param _delegation The delegation contract of EigenLayr.
     */
    constructor(IEigenLayrDelegation _delegation, IRegistryPermission _permissionManager)
        InvestmentManagerStorage(_delegation)
    {
        permissionManager = _permissionManager;
        _disableInitializers();
    }

    // EXTERNAL FUNCTIONS

    /**
     * @notice Initializes the investment manager contract. Sets the `pauserRegistry` (currently **not** modifiable after being set),
     * and transfers contract ownership to the specified `initialOwner`.
     * @param _pauserRegistry Used for access control of pausing.
     * @param initialOwner Ownership of this contract is transferred to this address.
     */
    function initialize(IPauserRegistry _pauserRegistry, address initialOwner)
        external
        initializer
    {
        //TODO: abstract this logic into an inherited contract for Delegation and Investment manager and have a conversation about meta transactions in general
        DOMAIN_SEPARATOR = keccak256(abi.encode(DOMAIN_TYPEHASH, bytes("EigenLayr"), block.chainid, address(this)));
        _initializePauser(_pauserRegistry, UNPAUSE_ALL);
        _transferOwnership(initialOwner);
    }

    /**
     * @notice Deposits `amount` of `token` into the specified `strategy`, with the resultant shares credited to `depositor`
     * @param strategy is the specified strategy where investment is to be made,
     * @param token is the denomination in which the investment is to be made,
     * @param amount is the amount of token to be invested in the strategy by the depositor
     * @dev The `msg.sender` must have previously approved this contract to transfer at least `amount` of `token` on their behalf.
     * @dev Cannot be called by an address that is 'frozen' (this function will revert if the `msg.sender` is frozen).
     */
    function depositIntoStrategy(IInvestmentStrategy strategy, IERC20 token, uint256 amount)
        external
        onlyWhenNotPaused(PAUSED_DEPOSITS)
        nonReentrant
        onlyStakeAndDelegate
        returns (uint256 shares)
    {
        shares = _depositIntoStrategy(msg.sender, strategy, token, amount);
    }

    /**
     * @notice Used for investing an asset into the specified strategy with the resultant shared created to `staker`,
     * who must sign off on the action
     * @param strategy is the specified strategy where investment is to be made,
     * @param token is the denomination in which the investment is to be made,
     * @param amount is the amount of token to be invested in the strategy by the depositor
     * @param staker the staker that the assets will be deposited on behalf of
     * @param expiry the timestamp at which the signature expires
     * @param r and @param vs are the elements of the ECDSA signature
     * @dev The `msg.sender` must have previously approved this contract to transfer at least `amount` of `token` on their behalf.
     * @dev A signature is required for this function to eliminate the possibility of griefing attacks, specifically those
     * targetting stakers who may be attempting to undelegate.
     * @dev Cannot be called on behalf of a staker that is 'frozen' (this function will revert if the `staker` is frozen).
     */
    function depositIntoStrategyOnBehalfOf(
        IInvestmentStrategy strategy,
        IERC20 token,
        uint256 amount,
        address staker,
        uint256 expiry,
        bytes32 r,
        bytes32 vs
    )
        external
        onlyWhenNotPaused(PAUSED_DEPOSITS)
        nonReentrant
        onlyStakeAndDelegate
        returns (uint256 shares)
    {
        require(
            expiry == 0 || expiry >= block.timestamp,
            "InvestmentManager.depositIntoStrategyOnBehalfOf: delegation signature expired"
        );
        // calculate struct hash, then increment `staker`'s nonce
        bytes32 structHash = keccak256(abi.encode(DEPOSIT_TYPEHASH, strategy, token, amount, nonces[staker]++, expiry));
        bytes32 digestHash = keccak256(abi.encodePacked("\x19\x01", DOMAIN_SEPARATOR, structHash));
        // check validity of signature
        address recoveredAddress = ECDSA.recover(digestHash, r, vs);
        require(recoveredAddress == staker, "InvestmentManager.depositIntoStrategyOnBehalfOf: sig not from staker");

        shares = _depositIntoStrategy(staker, strategy, token, amount);
    }

    /**
    * @notice Called by a delegator withdraw theirs stake token
     */
    function delegatorWithdraw(
        uint256[] calldata strategyIndexes,
        IInvestmentStrategy[] calldata strategies,
        IERC20[] calldata tokens,
        uint256[] calldata shares,
        bool undelegateIfPossible
    )
        external
        nonReentrant
        onlyStakeAndDelegate
        returns (bool)
    {
        require(!paused(PAUSED_WITHDRAWALS), "Pausable: index is paused");

        require(
            delegatorWithdrawWhiteList[msg.sender] == true,
            "InvestmentManager.delegatorWithdraw: you has not permission to withdraw token"
        );

        delegation.decreaseDelegatedShares(msg.sender, strategies, shares);

        uint256 strategyIndexIndex;

        uint256 strategiesLength = strategies.length;

        for (uint256 i = 0; i < strategiesLength;) {
            if (_removeShares(msg.sender, strategyIndexes[strategyIndexIndex], strategies[i], shares[i])) {
                unchecked {
                    ++strategyIndexIndex;
                }
            }
            strategies[i].withdraw(msg.sender, tokens[i], shares[i]);
            unchecked {
                ++i;
            }
        }

        if (undelegateIfPossible && investorStrats[msg.sender].length == 0) {
            _undelegate(msg.sender);
        }

        emit DelegatorWithdrawal(msg.sender, strategies, tokens, shares);

        return true;
    }

    /**
     * @notice Called by a staker to undelegate entirely from EigenLayer. The staker must first withdraw all of their existing deposits
     * (through use of the `queueWithdrawal` function), or else otherwise have never deposited in EigenLayer prior to delegating.
     */
    function undelegate() external onlyStakeAndDelegate {
        _undelegate(msg.sender);
    }


    /**
     * @notice This function adds `shares` for a given `strategy` to the `depositor` and runs through the necessary update logic.
     * @dev In particular, this function calls `delegation.increaseDelegatedShares(depositor, strategy, shares)` to ensure that all
     * delegated shares are tracked, increases the stored share amount in `investorStratShares[depositor][strategy]`, and adds `strategy`
     * to the `depositor`'s list of strategies, if it is not in the list already.
     */
    function _addShares(address depositor, IInvestmentStrategy strategy, uint256 shares) internal {
        // sanity check on `shares` input
        require(shares != 0, "InvestmentManager._addShares: shares should not be zero!");
        require(strategyStorage[strategy], "InvestmentManager._addShares: do not support this strategy!");

        // if they dont have existing shares of this strategy, add it to their strats
        if (investorStratShares[depositor][strategy] == 0) {
            require(
                investorStrats[depositor].length < MAX_INVESTOR_STRATS_LENGTH,
                "InvestmentManager._addShares: deposit would exceed MAX_INVESTOR_STRATS_LENGTH"
            );
            investorStrats[depositor].push(strategy);
        }

        // add the returned shares to their existing shares for this strategy
        investorStratShares[depositor][strategy] += shares;

        // if applicable, increase delegated shares accordingly
        delegation.increaseDelegatedShares(depositor, strategy, shares);
    }

    /**
     * @notice Internal function in which `amount` of ERC20 `token` is transferred from `msg.sender` to the InvestmentStrategy-type contract
     * `strategy`, with the resulting shares credited to `depositor`.
     * @return shares The amount of *new* shares in `strategy` that have been credited to the `depositor`.
     */
    function _depositIntoStrategy(address depositor, IInvestmentStrategy strategy, IERC20 token, uint256 amount)
        internal
        returns (uint256 shares)
    {
        // transfer tokens from the sender to the strategy
        token.safeTransferFrom(msg.sender, address(strategy), amount);

        // deposit the assets into the specified strategy and get the equivalent amount of shares in that strategy
        shares = strategy.deposit(token, amount);

        // add the returned shares to the depositor's existing shares for this strategy
        _addShares(depositor, strategy, shares);

        return shares;
    }

    /**
     * @notice Decreases the shares that `depositor` holds in `strategy` by `shareAmount`.
     * @dev If the amount of shares represents all of the depositor`s shares in said strategy,
     * then the strategy is removed from investorStrats[depositor] and 'true' is returned. Otherwise 'false' is returned.
     */
    function _removeShares(address depositor, uint256 strategyIndex, IInvestmentStrategy strategy, uint256 shareAmount)
        internal
        returns (bool)
    {
        // sanity check on `shareAmount` input
        require(shareAmount != 0, "InvestmentManager._removeShares: shareAmount should not be zero!");

        //check that the user has sufficient shares
        uint256 userShares = investorStratShares[depositor][strategy];

        require(shareAmount <= userShares, "InvestmentManager._removeShares: shareAmount too high");
        //unchecked arithmetic since we just checked this above
        unchecked {
            userShares = userShares - shareAmount;
        }

        // subtract the shares from the depositor's existing shares for this strategy
        investorStratShares[depositor][strategy] = userShares;
        // if no existing shares, remove is from this investors strats

        if (userShares == 0) {
            // remove the strategy from the depositor's dynamic array of strategies
            _removeStrategyFromInvestorStrats(depositor, strategyIndex, strategy);

            // return true in the event that the strategy was removed from investorStrats[depositor]
            return true;
        }
        // return false in the event that the strategy was *not* removed from investorStrats[depositor]
        return false;
    }

    /**
     * @notice Removes `strategy` from `depositor`'s dynamic array of strategies, i.e. from `investorStrats[depositor]`
     * @dev the provided `strategyIndex` input is optimistically used to find the strategy quickly in the list. If the specified
     * index is incorrect, then we revert to a brute-force search.
     */
    function _removeStrategyFromInvestorStrats(address depositor, uint256 strategyIndex, IInvestmentStrategy strategy) internal {
        // if the strategy matches with the strategy index provided
        if (investorStrats[depositor][strategyIndex] == strategy) {
            // replace the strategy with the last strategy in the list
            investorStrats[depositor][strategyIndex] =
                investorStrats[depositor][investorStrats[depositor].length - 1];
        } else {
            //loop through all of the strategies, find the right one, then replace
            uint256 stratsLength = investorStrats[depositor].length;

            for (uint256 j = 0; j < stratsLength;) {
                if (investorStrats[depositor][j] == strategy) {
                    //replace the strategy with the last strategy in the list
                    investorStrats[depositor][j] = investorStrats[depositor][investorStrats[depositor].length - 1];
                    break;
                }
                unchecked {
                    ++j;
                }
            }
        }

        // pop off the last entry in the list of strategies
        investorStrats[depositor].pop();
    }

    /**
     * @notice If the `depositor` has no existing shares, then they can `undelegate` themselves.
     * This allows people a "hard reset" in their relationship with EigenLayer after withdrawing all of their stake.
     */
    function _undelegate(address depositor) internal {
        require(investorStrats[depositor].length == 0, "InvestmentManager._undelegate: depositor has active deposits");
        delegation.undelegate(depositor);
    }

    // VIEW FUNCTIONS
    /**
     * @notice Get all details on the depositor's investments and corresponding shares
     * @return (depositor's strategies, shares in these strategies)
     */
    function getDeposits(address depositor) external view returns (IInvestmentStrategy[] memory, uint256[] memory) {
        uint256 strategiesLength = investorStrats[depositor].length;
        uint256[] memory shares = new uint256[](strategiesLength);

        for (uint256 i = 0; i < strategiesLength;) {
            shares[i] = investorStratShares[depositor][investorStrats[depositor][i]];
            unchecked {
                ++i;
            }
        }
        return (investorStrats[depositor], shares);
    }

    /// @notice Simple getter function that returns `investorStrats[staker].length`.
    function investorStratsLength(address staker) external view returns (uint256) {
        return investorStrats[staker].length;
    }

    function setDelegatorCanWithdraw(address withdrawer) external onlyOwner {
        delegatorWithdrawWhiteList[withdrawer] = true;
    }

    function setInvestmentStrategy(IInvestmentStrategy _strategy) external onlyOwner {
        strategyStorage[_strategy] = true;
    }
}
