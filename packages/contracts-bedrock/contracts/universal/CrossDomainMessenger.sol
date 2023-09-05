// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { SafeCall } from "../libraries/SafeCall.sol";
import { Hashing } from "../libraries/Hashing.sol";
import { Encoding } from "../libraries/Encoding.sol";
import { Constants } from "../libraries/Constants.sol";

/**
 * @custom:legacy
 * @title CrossDomainMessengerLegacySpacer0
 * @notice Contract only exists to add a spacer to the CrossDomainMessenger where the
 *         libAddressManager variable used to exist. Must be the first contract in the inheritance
 *         tree of the CrossDomainMessenger.
 */
contract CrossDomainMessengerLegacySpacer0 {
    /**
     * @custom:legacy
     * @custom:spacer libAddressManager
     * @notice Spacer for backwards compatibility.
     */
    address private spacer_0_0_20;
}

/**
 * @custom:legacy
 * @title CrossDomainMessengerLegacySpacer1
 * @notice Contract only exists to add a spacer to the CrossDomainMessenger where the
 *         PausableUpgradable and OwnableUpgradeable variables used to exist. Must be
 *         the third contract in the inheritance tree of the CrossDomainMessenger.
 */
contract CrossDomainMessengerLegacySpacer1 {
    /**
     * @custom:legacy
     * @custom:spacer ContextUpgradable's __gap
     * @notice Spacer for backwards compatibility. Comes from OpenZeppelin
     *         ContextUpgradable.
     *
     */
    uint256[50] private spacer_1_0_1600;

    /**
     * @custom:legacy
     * @custom:spacer OwnableUpgradeable's _owner
     * @notice Spacer for backwards compatibility.
     *         Come from OpenZeppelin OwnableUpgradeable.
     */
    address private spacer_51_0_20;

    /**
     * @custom:legacy
     * @custom:spacer OwnableUpgradeable's __gap
     * @notice Spacer for backwards compatibility. Comes from OpenZeppelin
     *         OwnableUpgradeable.
     */
    uint256[49] private spacer_52_0_1568;

    /**
     * @custom:legacy
     * @custom:spacer PausableUpgradable's _paused
     * @notice Spacer for backwards compatibility. Comes from OpenZeppelin
     *         PausableUpgradable.
     */
    bool private spacer_101_0_1;

    /**
     * @custom:legacy
     * @custom:spacer PausableUpgradable's __gap
     * @notice Spacer for backwards compatibility. Comes from OpenZeppelin
     *         PausableUpgradable.
     */
    uint256[49] private spacer_102_0_1568;

    /**
     * @custom:legacy
     * @custom:spacer ReentrancyGuardUpgradeable's `_status` field.
     * @notice Spacer for backwards compatibility.
     */
    uint256 private spacer_151_0_32;

    /**
     * @custom:legacy
     * @custom:spacer ReentrancyGuardUpgradeable's __gap
     * @notice Spacer for backwards compatibility.
     */
    uint256[49] private spacer_152_0_1568;

    /**
     * @custom:legacy
     * @custom:spacer blockedMessages
     * @notice Spacer for backwards compatibility.
     */
    mapping(bytes32 => bool) private spacer_201_0_32;

    /**
     * @custom:legacy
     * @custom:spacer relayedMessages
     * @notice Spacer for backwards compatibility.
     */
    mapping(bytes32 => bool) private spacer_202_0_32;
}

/**
 * @custom:upgradeable
 * @title CrossDomainMessenger
 * @notice CrossDomainMessenger is a base contract that provides the core logic for the L1 and L2
 *         cross-chain messenger contracts. It's designed to be a universal interface that only
 *         needs to be extended slightly to provide low-level message passing functionality on each
 *         chain it's deployed on. Currently only designed for message passing between two paired
 *         chains and does not support one-to-many interactions.
 *
 *         Any changes to this contract MUST result in a semver bump for contracts that inherit it.
 */
abstract contract CrossDomainMessenger is
    CrossDomainMessengerLegacySpacer0,
    Initializable,
    CrossDomainMessengerLegacySpacer1
{
    /**
     * @notice Current message version identifier.
     */
    uint16 public constant MESSAGE_VERSION = 1;

    /**
     * @notice Constant overhead added to the base gas for a message.
     */
    uint64 public constant RELAY_CONSTANT_OVERHEAD = 200_000;

    /**
     * @notice Numerator for dynamic overhead added to the base gas for a message.
     */
    uint64 public constant MIN_GAS_DYNAMIC_OVERHEAD_NUMERATOR = 64;

    /**
     * @notice Denominator for dynamic overhead added to the base gas for a message.
     */
    uint64 public constant MIN_GAS_DYNAMIC_OVERHEAD_DENOMINATOR = 63;

    /**
     * @notice Extra gas added to base gas for each byte of calldata in a message.
     */
    uint64 public constant MIN_GAS_CALLDATA_OVERHEAD = 16;

    /**
     * @notice Gas reserved for performing the external call in `relayMessage`.
     */
    uint64 public constant RELAY_CALL_OVERHEAD = 40_000;

    /**
     * @notice Gas reserved for finalizing the execution of `relayMessage` after the safe call.
     */
    uint64 public constant RELAY_RESERVED_GAS = 40_000;

    /**
     * @notice Gas reserved for the execution between the `hasMinGas` check and the external
     *         call in `relayMessage`.
     */
    uint64 public constant RELAY_GAS_CHECK_BUFFER = 5_000;

    /**
     * @notice Address of the paired CrossDomainMessenger contract on the other chain.
     */
    address public immutable OTHER_MESSENGER;

    /**
     * @notice Mapping of message hashes to boolean receipt values. Note that a message will only
     *         be present in this mapping if it has successfully been relayed on this chain, and
     *         can therefore not be relayed again.
     */
    mapping(bytes32 => bool) public successfulMessages;

    /**
     * @notice Address of the sender of the currently executing message on the other chain. If the
     *         value of this variable is the default value (0x00000000...dead) then no message is
     *         currently being executed. Use the xDomainMessageSender getter which will throw an
     *         error if this is the case.
     */
    address internal xDomainMsgSender;

    /**
     * @notice Nonce for the next message to be sent, without the message version applied. Use the
     *         messageNonce getter which will insert the message version into the nonce to give you
     *         the actual nonce to be used for the message.
     */
    uint240 internal msgNonce;

    /**
     * @notice Mapping of message hashes to a boolean if and only if the message has failed to be
     *         executed at least once. A message will not be present in this mapping if it
     *         successfully executed on the first attempt.
     */
    mapping(bytes32 => bool) public failedMessages;

    /**
     * @notice Reserve extra slots in the storage layout for future upgrades.
     *         A gap size of 41 was chosen here, so that the first slot used in a child contract
     *         would be a multiple of 50.
     */
    uint256[42] private __gap;

    /**
     * @notice Emitted whenever a message is sent to the other chain.
     *
     * @param target       Address of the recipient of the message.
     * @param sender       Address of the sender of the message.
     * @param message      Message to trigger the recipient address with.
     * @param messageNonce Unique nonce attached to the message.
     * @param gasLimit     Minimum gas limit that the message can be executed with.
     */
    event SentMessage(
        address indexed target,
        address sender,
        bytes message,
        uint256 messageNonce,
        uint256 gasLimit
    );

    /**
     * @notice Additional event data to emit, required as of Bedrock. Cannot be merged with the
     *         SentMessage event without breaking the ABI of this contract, this is good enough.
     *
     * @param sender Address of the sender of the message.
     * @param value  ETH value sent along with the message to the recipient.
     */
    event SentMessageExtension1(address indexed sender, uint256 value);

    /**
     * @notice Emitted whenever a message is successfully relayed on this chain.
     *
     * @param msgHash Hash of the message that was relayed.
     */
    event RelayedMessage(bytes32 indexed msgHash);

    /**
     * @notice Emitted whenever a message fails to be relayed on this chain.
     *
     * @param msgHash Hash of the message that failed to be relayed.
     */
    event FailedRelayedMessage(bytes32 indexed msgHash);

    /**
     * @param _otherMessenger Address of the messenger on the paired chain.
     */
    constructor(address _otherMessenger) {
        OTHER_MESSENGER = _otherMessenger;
    }

    /**
     * @notice Sends a message to some target address on the other chain. Note that if the call
     *         always reverts, then the message will be unrelayable, and any ETH sent will be
     *         permanently locked. The same will occur if the target on the other chain is
     *         considered unsafe (see the _isUnsafeTarget() function).
     *
     * @param _target      Target contract or wallet address.
     * @param _message     Message to trigger the target address with.
     * @param _minGasLimit Minimum gas limit that the message can be executed with.
     */
    function sendMessage(
        address _target,
        bytes calldata _message,
        uint32 _minGasLimit
    ) external payable virtual ;

    /**
     * @notice Relays a message that was sent by the other CrossDomainMessenger contract. Can only
     *         be executed via cross-chain call from the other messenger OR if the message was
     *         already received once and is currently being replayed.
     *
     * @param _nonce       Nonce of the message being relayed.
     * @param _sender      Address of the user who sent the message.
     * @param _target      Address that the message is targeted at.
     * @param _value       ETH value to send with the message.
     * @param _minGasLimit Minimum amount of gas that the message can be executed with.
     * @param _message     Message to send to the target.
     */
    function relayMessage(
        uint256 _nonce,
        address _sender,
        address _target,
        uint256 _value,
        uint256 _minGasLimit,
        bytes calldata _message
    ) external payable virtual;

    /**
     * @notice Retrieves the address of the contract or wallet that initiated the currently
     *         executing message on the other chain. Will throw an error if there is no message
     *         currently being executed. Allows the recipient of a call to see who triggered it.
     *
     * @return Address of the sender of the currently executing message on the other chain.
     */
    function xDomainMessageSender() external view virtual returns (address);


    /**
     * @notice Retrieves the next message nonce. Message version will be added to the upper two
     *         bytes of the message nonce. Message version allows us to treat messages as having
     *         different structures.
     *
     * @return Nonce of the next message to be sent, with added message version.
     */
    function messageNonce() public view virtual returns (uint256);

    /**
     * @notice Computes the amount of gas required to guarantee that a given message will be
     *         received on the other chain without running out of gas. Guaranteeing that a message
     *         will not run out of gas is important because this ensures that a message can always
     *         be replayed on the other chain if it fails to execute completely.
     *
     * @param _message     Message to compute the amount of required gas for.
     * @param _minGasLimit Minimum desired gas limit when message goes to target.
     *
     * @return Amount of gas required to guarantee message receipt.
     */
    function baseGas(bytes calldata _message, uint32 _minGasLimit) public pure virtual returns (uint64);
    /**
     * @notice Intializer.
     */
    // solhint-disable-next-line func-name-mixedcase
    function __CrossDomainMessenger_init() internal virtual ;

    /**
     * @notice Sends a low-level message to the other messenger. Needs to be implemented by child
     *         contracts because the logic for this depends on the network where the messenger is
     *         being deployed.
     *
     * @param _to       Recipient of the message on the other chain.
     * @param _gasLimit Minimum gas limit the message can be executed with.
     * @param _value    Amount of ETH to send with the message.
     * @param _data     Message data.
     */
    function _sendMessage(
        address _to,
        uint64 _gasLimit,
        uint256 _value,
        bytes memory _data
    ) internal virtual;

    /**
     * @notice Checks whether the message is coming from the other messenger. Implemented by child
     *         contracts because the logic for this depends on the network where the messenger is
     *         being deployed.
     *
     * @return Whether the message is coming from the other messenger.
     */
    function _isOtherMessenger() internal view virtual returns (bool);

    /**
     * @notice Checks whether a given call target is a system address that could cause the
     *         messenger to peform an unsafe action. This is NOT a mechanism for blocking user
     *         addresses. This is ONLY used to prevent the execution of messages to specific
     *         system addresses that could cause security issues, e.g., having the
     *         CrossDomainMessenger send messages to itself.
     *
     * @param _target Address of the contract to check.
     *
     * @return Whether or not the address is an unsafe system address.
     */
    function _isUnsafeTarget(address _target) internal view virtual returns (bool);
}
