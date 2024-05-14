// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin-upgrades/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin-upgrades/contracts/access/OwnableUpgradeable.sol";

import "./RegistryBase.sol";
import "../interfaces/IBLSPublicKeyCompendium.sol";
import "../interfaces/IBLSRegistry.sol";
import "../libraries/BN254.sol";
import "../interfaces/IRegistryPermission.sol";

/**
 * @title A Registry-type contract using aggregate BLS signatures.
 * @author Layr Labs, Inc.
 * @notice This contract is used for
 * - registering new operators
 * - committing to and finalizing de-registration as an operator
 * - updating the stakes of the operator
 */
contract BLSRegistry is Initializable, OwnableUpgradeable, RegistryBase, IBLSRegistry {
    using BytesLib for bytes;

    // Hash of the zero public key
    bytes32 internal constant ZERO_PK_HASH = hex"012893657d8eb2efad4de0a91bcd0e39ad9837745dec3ea923737ea803fc8e3d";

    /// @notice contract used for looking up operators' BLS public keys
    IBLSPublicKeyCompendium public immutable pubkeyCompendium;

    /// @notice contract used for manage operator register permission
    IRegistryPermission public immutable permissionManager;

    /**
     * @notice list of keccak256(apk_x, apk_y) of operators, and the block numbers at which the aggregate
     * pubkeys were updated. This occurs whenever a new operator registers or deregisters.
     */
    ApkUpdate[] internal _apkUpdates;

    /**
     * @dev Initialized value of APK is the point at infinity: (0, 0)
     * @notice used for storing current aggregate public key
     */
    BN254.G1Point public apk;


    /// @notice Address that has permission to deregister any operator
    address public forceDeregister;

    // EVENTS
    /**
     * @notice Emitted upon the registration of a new operator for the middleware
     * @param operator Address of the new operator
     * @param pkHash The keccak256 hash of the operator's public key
     * @param pk The operator's public key itself
     * @param apkHashIndex The index of the latest (i.e. the new) APK update
     * @param apkHash The keccak256 hash of the new Aggregate Public Key
     */
    event Registration(
        address indexed operator,
        bytes32 pkHash,
        BN254.G1Point pk,
        uint32 apkHashIndex,
        bytes32 apkHash,
        string socket
    );

    /// @notice when applied to a function, ensures that the function is only callable by the `feeSetter`.
    modifier onlyForceDeregister() {
        require(msg.sender == forceDeregister, "onlyForceDeregister can do this action");
        _;
    }

    constructor(
        IInvestmentManager _investmentManager,
        IServiceManager _serviceManager,
        uint8 _NUMBER_OF_QUORUMS,
        IBLSPublicKeyCompendium _pubkeyCompendium,
        IRegistryPermission _permissionManager,
        address _forceDeregister
    )
        RegistryBase(
            _investmentManager,
            _serviceManager,
            _NUMBER_OF_QUORUMS
        )
    {
        // set compendium
        pubkeyCompendium = _pubkeyCompendium;
        // set permission
        permissionManager = _permissionManager;
        // set forceDeregister
        forceDeregister = _forceDeregister;
    }

    /// @notice Initialize the APK, the payment split between quorums, and the quorum strategies + multipliers.
    function initialize(
        uint256[] memory _quorumBips,
        address initialOwner,
        StrategyAndWeightingMultiplier[] memory _firstQuorumStrategiesConsideredAndMultipliers,
        StrategyAndWeightingMultiplier[] memory _secondQuorumStrategiesConsideredAndMultipliers
    ) public virtual initializer {
        _transferOwnership(initialOwner);

        // process an apk update to get index and totalStake arrays to the same length
        _processApkUpdate(BN254.G1Point(0, 0));
        RegistryBase._initialize(
            _quorumBips,
            _firstQuorumStrategiesConsideredAndMultipliers,
            _secondQuorumStrategiesConsideredAndMultipliers
        );
    }

    /**
     * @notice called for registering as an operator
     * @param operatorType specifies whether the operator want to register as staker for one or both quorums
     * @param pk is the operator's G1 public key
     * @param socket is the socket address of the operator
     */
    function registerOperator(uint8 operatorType, BN254.G1Point memory pk, string calldata socket) external virtual {
        require(
            permissionManager.getOperatorRegisterPermission(msg.sender) == true,
            "BLSRegistry.registerOperator: Operator does not permission to register"
        );
        _registerOperator(msg.sender, operatorType, pk, socket);
    }

    /**
     * @param operator is the node who is registering to be a operator
     * @param operatorType specifies whether the operator want to register as staker for one or both quorums
     * @param pk is the operator's G1 public key
     * @param socket is the socket address of the operator
     */
    function _registerOperator(address operator, uint8 operatorType, BN254.G1Point memory pk, string calldata socket)
        internal
    {
        // validate the registration of `operator` and find their `OperatorStake`
        OperatorStake memory _operatorStake = _registrationStakeEvaluation(operator, operatorType);

        // getting pubkey hash
        bytes32 pubkeyHash = BN254.hashG1Point(pk);

        require(pubkeyHash != ZERO_PK_HASH, "BLSRegistry._registerOperator: Cannot register with 0x0 public key");

        require(
            pubkeyCompendium.pubkeyHashToOperator(pubkeyHash) == operator,
            "BLSRegistry._registerOperator: operator does not own pubkey"
        );

        // the new aggregate public key is the current one added to registering operator's public key
        BN254.G1Point memory newApk = BN254.plus(apk, pk);

        // record the APK update and get the hash of the new APK
        bytes32 newApkHash = _processApkUpdate(newApk);

        // add the operator to the list of registrants and do accounting
        _addRegistrant(operator, pubkeyHash, _operatorStake);

        emit Registration(operator, pubkeyHash, pk, uint32(_apkUpdates.length - 1), newApkHash, socket);
    }

    /**
     * @notice Used by an operator to de-register itself from providing service to the middleware.
     * @param pkToRemove is the sender's pubkey in affine coordinates
     * @param index is the sender's location in the dynamic array `operatorList`
     */
    function deregisterOperator(BN254.G1Point memory pkToRemove, uint32 index) external virtual returns (bool) {
        require(
            permissionManager.getOperatorDeregisterPermission(msg.sender) == true,
            "BLSRegistry.deregisterOperator: Operator should apply deregister permission first and then can deregister"
        );
        _deregisterOperator(msg.sender, pkToRemove, index);
        return true;
    }

    function forceDeregisterOperator(BN254.G1Point memory pkToRemove, address operator, uint32 index) external onlyForceDeregister returns (bool) {
        _deregisterOperator(operator, pkToRemove, index);
        return true;
    }

    /**
     * @notice Used to process de-registering an operator from providing service to the middleware.
     * @param operator The operator to be deregistered
     * @param pkToRemove is the sender's pubkey
     * @param index is the sender's location in the dynamic array `operatorList`
     */
    function _deregisterOperator(address operator, BN254.G1Point memory pkToRemove, uint32 index) internal {
        // verify that the `operator` is an active operator and that they've provided the correct `index`
        _deregistrationCheck(operator, index);


        /// @dev Fetch operator's stored pubkeyHash
        bytes32 pubkeyHash = registry[operator].pubkeyHash;
        /// @dev Verify that the stored pubkeyHash matches the 'pubkeyToRemoveAff' input
        require(
            pubkeyHash == BN254.hashG1Point(pkToRemove),
            "BLSRegistry._deregisterOperator: pubkey input does not match stored pubkeyHash"
        );

        // the new apk is the current one minus the sender's pubkey (apk = apk + (-pk))
        BN254.G1Point memory newApk = BN254.plus(apk, BN254.negate(pkToRemove));

        bytes32 newApkHash = BN254.hashG1Point(newApk);

        // Perform necessary updates for removing operator, including updating operator list and index histories
        _removeOperator(operator, pubkeyHash, pkToRemove, newApkHash, index);

        // update the aggregate public key of all registered operators and record this update in history
        _processApkUpdate(newApk);
    }

    /**
     * @notice Used for updating information on deposits of nodes.
     * @param operators are the nodes whose deposit information is getting updated
     * @param prevElements are the elements before this middleware in the operator's linked list within the slasher
     */
    function updateStakes(address[] calldata operators, uint256[] calldata prevElements) external {
        // copy total stake to memory
        OperatorStake memory _totalStake = totalStakeHistory[totalStakeHistory.length - 1];

        // placeholders reused inside of loop
        OperatorStake memory currentStakes;
        bytes32 pubkeyHash;
        uint256 operatorsLength = operators.length;
        // make sure lengths are consistent
        require(operatorsLength == prevElements.length, "BLSRegistry.updateStakes: prevElement is not the same length as operators");
        // iterating over all the tuples that are to be updated
        for (uint256 i = 0; i < operatorsLength;) {
            // get operator's pubkeyHash
            pubkeyHash = registry[operators[i]].pubkeyHash;
            // fetch operator's existing stakes
            currentStakes = pubkeyHashToStakeHistory[pubkeyHash][pubkeyHashToStakeHistory[pubkeyHash].length - 1];
            // decrease _totalStake by operator's existing stakes
            _totalStake.firstQuorumStake -= currentStakes.firstQuorumStake;
            _totalStake.secondQuorumStake -= currentStakes.secondQuorumStake;

            // update the stake for the i-th operator
            currentStakes = _updateOperatorStake(operators[i], pubkeyHash, currentStakes, prevElements[i]);

            // increase _totalStake by operator's updated stakes
            _totalStake.firstQuorumStake += currentStakes.firstQuorumStake;
            _totalStake.secondQuorumStake += currentStakes.secondQuorumStake;

            unchecked {
                ++i;
            }
        }

        // update storage of total stake
        _recordTotalStakeUpdate(_totalStake);
    }

    /**
     * @notice Updates the stored APK to `newApk`, calculates its hash, and pushes new entries to the `_apkUpdates` array
     * @param newApk The updated APK. This will be the `apk` after this function runs!
     */
    function _processApkUpdate(BN254.G1Point memory newApk) internal returns (bytes32) {
        // update stored aggregate public key
        // slither-disable-next-line costly-loop
        apk = newApk;

        // find the hash of aggregate pubkey
        bytes32 newApkHash = BN254.hashG1Point(newApk);

        // store the apk hash and the current block number in which the aggregated pubkey is being updated
        _apkUpdates.push(ApkUpdate({
            apkHash: newApkHash,
            blockNumber: uint32(block.number)
        }));

        return newApkHash;
    }

    /// @notice Used by Eigenlayr governance to adjust the address of the `forceDeregister`
    function setForceDeregister(address _forceDeregister) external onlyOwner {
        require(_forceDeregister != address(0), "BLSRegistry.setForceDeregister: forceDeregister address is the zero address");
        forceDeregister = _forceDeregister;
    }

    /**
     * @notice get hash of a historical aggregated public key corresponding to a given index;
     * called by checkSignatures in BLSSignatureChecker.sol.
     */
    function getCorrectApkHash(uint256 index, uint32 blockNumber) external view returns (bytes32) {
        // check that the `index`-th APK update occurred at or before `blockNumber`
        require(blockNumber >= _apkUpdates[index].blockNumber, "BLSRegistry.getCorrectApkHash: index too recent");

        // if not last update
        if (index != _apkUpdates.length - 1) {
            // check that there was not *another APK update* that occurred at or before `blockNumber`
            require(blockNumber < _apkUpdates[index + 1].blockNumber, "BLSRegistry.getCorrectApkHash: Not latest valid apk update");
        }

        return _apkUpdates[index].apkHash;
    }

    /// @notice returns the total number of APK updates that have ever occurred (including one for initializing the pubkey as the generator)
    function getApkUpdatesLength() external view returns (uint256) {
        return _apkUpdates.length;
    }

    /// @notice returns the `ApkUpdate` struct at `index` in the list of APK updates
    function apkUpdates(uint256 index) external view returns (ApkUpdate memory) {
        return _apkUpdates[index];
    }

    /// @notice returns the APK hash that resulted from the `index`th APK update
    function apkHashes(uint256 index) external view returns (bytes32) {
        return _apkUpdates[index].apkHash;
    }

    /// @notice returns the block number at which the `index`th APK update occurred
    function apkUpdateBlockNumbers(uint256 index) external view returns (uint32) {
        return _apkUpdates[index].blockNumber;
    }
}
