// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";
import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "./EigenLayrDelegationStorage.sol";
import "../permissions/Pausable.sol";
import "../interfaces/IRegistryPermission.sol";

/**
 * @title The primary delegation contract for EigenLayr.
 * @author Layr Labs, Inc.
 * @notice  This is the contract for delegation in EigenLayr. The main functionalities of this contract are
 * - enabling anyone to register as an operator in EigenLayr
 * - allowing new operators to provide a another eoa address, which may mediate their interactions with stakers who delegate to them
 * - enabling any staker to delegate its stake to the operator of its choice
 * - enabling a staker to undelegate its assets from an operator (performed as part of the withdrawal process, initiated through the InvestmentManager)
 */
contract EigenLayrDelegation is Initializable, OwnableUpgradeable, EigenLayrDelegationStorage, Pausable {
    uint8 internal constant PAUSED_NEW_DELEGATION = 0;

    /// @notice contract used for manage operator register permission
    IRegistryPermission public immutable permissionManager;

    /// @notice Simple permission for functions that are only callable by the InvestmentManager contract.
    modifier onlyInvestmentManager() {
        require(msg.sender == address(investmentManager), "onlyInvestmentManager");
        _;
    }

    // INITIALIZING FUNCTIONS
    constructor(IInvestmentManager _investmentManager,  IRegistryPermission _permissionManager)
        EigenLayrDelegationStorage(_investmentManager)
    {
        permissionManager = _permissionManager;
        _disableInitializers();
    }

    function initialize(IPauserRegistry _pauserRegistry, address initialOwner)
        external
        initializer
    {
        _initializePauser(_pauserRegistry, UNPAUSE_ALL);
        DOMAIN_SEPARATOR = keccak256(abi.encode(DOMAIN_TYPEHASH, bytes("EigenLayr"), block.chainid, address(this)));
        _transferOwnership(initialOwner);
    }

    // EXTERNAL FUNCTIONS
    /**
     * @notice This will be called by an operator to register itself as an operator that stakers can choose to delegate to.
     * @param  rewardReceiveAddress another EOA address for receive from mantle network
     */
    function registerAsOperator(address rewardReceiveAddress) external {
        require(
            permissionManager.getOperatorRegisterPermission(msg.sender) == true,
            "EigenLayrDelegation.registerAsOperator: Operator does not permission to register as operator"
        );
        require(
            operatorReceiverRewardAddress[msg.sender] == address(0),
            "EigenLayrDelegation.registerAsOperator: operator has already registered"
        );
        // store the address of the delegation contract that the operator is providing.
        operatorReceiverRewardAddress[msg.sender] = rewardReceiveAddress;
        _delegate(msg.sender, msg.sender);
    }

    /**
     *  @notice This will be called by a staker to delegate its assets to some operator.
     *  @param operator is the operator to whom staker (msg.sender) is delegating its assets
     */
    function delegateTo(address operator) external {
        require(
            permissionManager.getDelegatorPermission(msg.sender) == true,
            "InvestmentManager.depositIntoStrategy: delegator has not permission exec delegate to"
        );
        _delegate(msg.sender, operator);
    }

    /**
     * @notice Delegates from `staker` to `operator`.
     * @dev requires that r, vs are a valid ECSDA signature from `staker` indicating their intention for this action
     */
    function delegateToBySignature(address staker, address operator, uint256 expiry, bytes32 r, bytes32 vs)
        external
    {
        require(
            permissionManager.getDelegatorPermission(msg.sender) == true,
            "InvestmentManager.depositIntoStrategy: delegator has not permission exec delegate to"
        );
        require(expiry == 0 || expiry >= block.timestamp, "delegation signature expired");
        // calculate struct hash, then increment `staker`'s nonce
        bytes32 structHash = keccak256(abi.encode(DELEGATION_TYPEHASH, staker, operator, nonces[staker]++, expiry));
        bytes32 digestHash = keccak256(abi.encodePacked("\x19\x01", DOMAIN_SEPARATOR, structHash));
        //check validity of signature

        address recoveredAddress = ECDSA.recover(digestHash, r, vs);

        require(recoveredAddress == staker, "EigenLayrDelegation.delegateToBySignature: sig not from staker");
        _delegate(staker, operator);
    }

    /**
     * @notice Undelegates `staker` from the operator who they are delegated to.
     * @notice Callable only by the InvestmentManager
     * @dev Should only ever be called in the event that the `staker` has no active deposits in EigenLayer.
     */
    function undelegate(address staker) external onlyInvestmentManager {
        require(!isOperator(staker), "EigenLayrDelegation.undelegate: operators cannot undelegate from themselves");
        delegatedTo[staker] = address(0);
    }

    /// @notice returns the eoa address of the `operator`, which may mediate their interactions with stakers who delegate to them.
    function getOperatorRewardAddress(address operator) external view returns (address){
        return operatorReceiverRewardAddress[operator];
    }

    /**
     * @notice Increases the `staker`'s delegated shares in `strategy` by `shares, typically called when the staker has further deposits into EigenLayr
     * @dev Callable only by the InvestmentManager
     */
    function increaseDelegatedShares(address staker, IInvestmentStrategy strategy, uint256 shares)
        external
        onlyInvestmentManager
    {
        //if the staker is delegated to an operator
        if (isDelegated(staker)) {
            address operator = delegatedTo[staker];

            // add strategy shares to delegate's shares
            operatorShares[operator][strategy] += shares;
        }
    }

    /**
     * @notice Decreases the `staker`'s delegated shares in each entry of `strategies` by its respective `shares[i]`, typically called when the staker withdraws from EigenLayr
     * @dev Callable only by the InvestmentManager
     */
    function decreaseDelegatedShares(
        address staker,
        IInvestmentStrategy[] calldata strategies,
        uint256[] calldata shares
    )
        external
        onlyInvestmentManager
    {
        if (isDelegated(staker)) {
            address operator = delegatedTo[staker];
            // subtract strategy shares from delegate's shares
            uint256 stratsLength = strategies.length;
            for (uint256 i = 0; i < stratsLength;) {
                operatorShares[operator][strategies[i]] -= shares[i];
                unchecked {
                    ++i;
                }
            }
        }
    }

    /**
     * @notice Internal function implementing the delegation *from* `staker` *to* `operator`.
     * @param staker The address to delegate *from* -- this address is delegating control of its own assets.
     * @param operator The address to delegate *to* -- this address is being given power to place the `staker`'s assets at risk on services
     * @dev Ensures that the operator has registered as a delegate (`address(dt) != address(0)`), verifies that `staker` is not already
     * delegated, and records the new delegation.
     */
    function _delegate(address staker, address operator) internal onlyWhenNotPaused(PAUSED_NEW_DELEGATION) {
        address rewardAddress = operatorReceiverRewardAddress[operator];
        require(
            rewardAddress != address(0), "EigenLayrDelegation._delegate: operator has not yet registered as a delegate"
        );

        require(isNotDelegated(staker), "EigenLayrDelegation._delegate: staker has existing delegation");

        // record delegation relation between the staker and operator
        delegatedTo[staker] = operator;

        // retrieve list of strategies and their shares from investment manager
        (IInvestmentStrategy[] memory strategies, uint256[] memory shares) = investmentManager.getDeposits(staker);

        // add strategy shares to delegate's shares
        uint256 stratsLength = strategies.length;
        for (uint256 i = 0; i < stratsLength;) {
            // update the share amounts for each of the operator's strategies
            operatorShares[operator][strategies[i]] += shares[i];
            unchecked {
                ++i;
            }
        }
    }

    // VIEW FUNCTIONS
    /// @notice Returns 'true' if `staker` *is* actively delegated, and 'false' otherwise.
    function isDelegated(address staker) public view returns (bool) {
        return (delegatedTo[staker] != address(0));
    }

    /// @notice Returns 'true' if `staker` is *not* actively delegated, and 'false' otherwise.
    function isNotDelegated(address staker) public view returns (bool) {
        return (delegatedTo[staker] == address(0));
    }

    /// @notice Returns if an operator can be delegated to, i.e. it has called `registerAsOperator`.
    function isOperator(address operator) public view returns (bool) {
        return operatorReceiverRewardAddress[operator] != address(0);
    }
}
