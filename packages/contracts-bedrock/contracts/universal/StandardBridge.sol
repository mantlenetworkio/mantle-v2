// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { ERC165Checker } from "@openzeppelin/contracts/utils/introspection/ERC165Checker.sol";
import { Address } from "@openzeppelin/contracts/utils/Address.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { SafeCall } from "../libraries/SafeCall.sol";
import { IOptimismMintableERC20, ILegacyMintableERC20 } from "./IOptimismMintableERC20.sol";
import { CrossDomainMessenger } from "./CrossDomainMessenger.sol";
import { OptimismMintableERC20 } from "./OptimismMintableERC20.sol";

/**
 * @custom:upgradeable
 * @title StandardBridge
 * @notice StandardBridge is a base contract for the L1 and L2 standard ERC20 bridges. It handles
 *         the core bridging logic, including escrowing tokens that are native to the local chain
 *         and minting/burning tokens that are native to the remote chain.
 */
abstract contract StandardBridge {
    using SafeERC20 for IERC20;

    /**
     * @notice The L2 gas limit set when eth is depoisited using the receive() function.
     */
    uint32 internal constant RECEIVE_DEFAULT_GAS_LIMIT = 200_000;


    /**
     * @custom:legacy
     * @custom:spacer messenger
     * @notice Spacer for backwards compatibility.
     */
    address private spacer_0_0_20;

    /**
     * @custom:legacy
     * @custom:spacer l2TokenBridge
     * @notice Spacer for backwards compatibility.
     */
    address private spacer_1_0_20;

    /**
     * @notice Mapping that stores deposits for a given pair of local and remote tokens.
     */
    mapping(address => mapping(address => uint256)) public deposits;

    /**
     * @notice Reserve extra slots (to a total of 50) in the storage layout for future upgrades.
     *         A gap size of 47 was chosen here, so that the first slot used in a child contract
     *         would be a multiple of 50.
     */
    uint256[47] private __gap;

    /**
     * @notice Emitted when an ETH bridge is initiated to the other chain.
     *
     * @param from      Address of the sender.
     * @param to        Address of the receiver.
     * @param amount    Amount of ETH sent.
     * @param extraData Extra data sent with the transaction.
     */
    event ETHBridgeInitiated(
        address indexed from,
        address indexed to,
        uint256 amount,
        bytes extraData
    );

    /**
     * @notice Emitted when an ETH bridge is finalized on this chain.
     *
     * @param from      Address of the sender.
     * @param to        Address of the receiver.
     * @param amount    Amount of ETH sent.
     * @param extraData Extra data sent with the transaction.
     */
    event ETHBridgeFinalized(
        address indexed from,
        address indexed to,
        uint256 amount,
        bytes extraData
    );

    /**
     * @notice Emitted when an ERC20 bridge is initiated to the other chain.
     *
     * @param localToken  Address of the ERC20 on this chain.
     * @param remoteToken Address of the ERC20 on the remote chain.
     * @param from        Address of the sender.
     * @param to          Address of the receiver.
     * @param amount      Amount of the ERC20 sent.
     * @param extraData   Extra data sent with the transaction.
     */
    event ERC20BridgeInitiated(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    /**
     * @notice Emitted when an ERC20 bridge is finalized on this chain.
     *
     * @param localToken  Address of the ERC20 on this chain.
     * @param remoteToken Address of the ERC20 on the remote chain.
     * @param from        Address of the sender.
     * @param to          Address of the receiver.
     * @param amount      Amount of the ERC20 sent.
     * @param extraData   Extra data sent with the transaction.
     */
    event ERC20BridgeFinalized(
        address indexed localToken,
        address indexed remoteToken,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    /**
     * @notice Only allow EOAs to call the functions. Note that this is not safe against contracts
     *         calling code within their constructors, but also doesn't really matter since we're
     *         just trying to prevent users accidentally depositing with smart contract wallets.
     */
    modifier onlyEOA() {
        require(
            !Address.isContract(msg.sender),
            "StandardBridge: function can only be called from an EOA"
        );
        _;
    }



    /**
     * @notice Allows EOAs to bridge ETH by sending directly to the bridge.
     *         Must be implemented by contracts that inherit.
     */
    receive() external payable virtual;


    /**
     * @notice Sends ETH to the sender's address on the other chain.
     *
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function bridgeETH(uint32 _minGasLimit, bytes calldata _extraData) virtual public payable onlyEOA {
        _initiateBridgeETH(msg.sender, msg.sender, msg.value, _minGasLimit, _extraData);
    }

    /**
     * @notice Sends ETH to a receiver's address on the other chain. Note that if ETH is sent to a
     *         smart contract and the call fails, the ETH will be temporarily locked in the
     *         StandardBridge on the other chain until the call is replayed. If the call cannot be
     *         replayed with any amount of gas (call always reverts), then the ETH will be
     *         permanently locked in the StandardBridge on the other chain. ETH will also
     *         be locked if the receiver is the other bridge, because finalizeBridgeETH will revert
     *         in that case.
     *
     * @param _to          Address of the receiver.
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function bridgeETHTo(
        address _to,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) public payable virtual{
        _initiateBridgeETH(msg.sender, _to, msg.value, _minGasLimit, _extraData);
    }

    /**
     * @notice Sends ERC20 tokens to the sender's address on the other chain. Note that if the
     *         ERC20 token on the other chain does not recognize the local token as the correct
     *         pair token, the ERC20 bridge will fail and the tokens will be returned to sender on
     *         this chain.
     *
     * @param _localToken  Address of the ERC20 on this chain.
     * @param _remoteToken Address of the corresponding token on the remote chain.
     * @param _amount      Amount of local tokens to deposit.
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function bridgeERC20(
        address _localToken,
        address _remoteToken,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) public virtual onlyEOA {
        _initiateBridgeERC20(
            _localToken,
            _remoteToken,
            msg.sender,
            msg.sender,
            _amount,
            _minGasLimit,
            _extraData
        );
    }

    /**
     * @notice Sends ERC20 tokens to a receiver's address on the other chain. Note that if the
     *         ERC20 token on the other chain does not recognize the local token as the correct
     *         pair token, the ERC20 bridge will fail and the tokens will be returned to sender on
     *         this chain.
     *
     * @param _localToken  Address of the ERC20 on this chain.
     * @param _remoteToken Address of the corresponding token on the remote chain.
     * @param _to          Address of the receiver.
     * @param _amount      Amount of local tokens to deposit.
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function bridgeERC20To(
        address _localToken,
        address _remoteToken,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) public virtual {
        _initiateBridgeERC20(
            _localToken,
            _remoteToken,
            msg.sender,
            _to,
            _amount,
            _minGasLimit,
            _extraData
        );
    }

    /**
     * @notice Finalizes an ETH bridge on this chain. Can only be triggered by the other
     *         StandardBridge contract on the remote chain.
     *
     * @param _from      Address of the sender.
     * @param _to        Address of the receiver.
     * @param _amount    Amount of ETH being bridged.
     * @param _extraData Extra data to be sent with the transaction. Note that the recipient will
     *                   not be triggered with this data, but it will be emitted and can be used
     *                   to identify the transaction.
     */
    function finalizeBridgeETH(
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _extraData
    ) public payable virtual;


    /**
     * @notice Finalizes an ERC20 bridge on this chain. Can only be triggered by the other
     *         StandardBridge contract on the remote chain.
     *
     * @param _localToken  Address of the ERC20 on this chain.
     * @param _remoteToken Address of the corresponding token on the remote chain.
     * @param _from        Address of the sender.
     * @param _to          Address of the receiver.
     * @param _amount      Amount of the ERC20 being bridged.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function finalizeBridgeERC20(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _extraData
    ) public virtual;

    /**
     * @notice Initiates a bridge of ETH through the CrossDomainMessenger.
     *
     * @param _from        Address of the sender.
     * @param _to          Address of the receiver.
     * @param _amount      Amount of ETH being bridged.
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function _initiateBridgeETH(
        address _from,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes memory _extraData
    ) internal virtual;


    /**
     * @notice Sends ERC20 tokens to a receiver's address on the other chain.
     *
     * @param _localToken  Address of the ERC20 on this chain.
     * @param _remoteToken Address of the corresponding token on the remote chain.
     * @param _to          Address of the receiver.
     * @param _amount      Amount of local tokens to deposit.
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function _initiateBridgeERC20(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes memory _extraData
    ) internal virtual ;

    /**
     * @notice Checks if a given address is an OptimismMintableERC20. Not perfect, but good enough.
     *         Just the way we like it.
     *
     * @param _token Address of the token to check.
     *
     * @return True if the token is an OptimismMintableERC20.
     */
    function _isOptimismMintableERC20(address _token) internal view returns (bool) {
        return
            ERC165Checker.supportsInterface(_token, type(ILegacyMintableERC20).interfaceId) ||
            ERC165Checker.supportsInterface(_token, type(IOptimismMintableERC20).interfaceId);
    }

    /**
     * @notice Checks if the "other token" is the correct pair token for the OptimismMintableERC20.
     *         Calls can be saved in the future by combining this logic with
     *         `_isOptimismMintableERC20`.
     *
     * @param _mintableToken OptimismMintableERC20 to check against.
     * @param _otherToken    Pair token to check.
     *
     * @return True if the other token is the correct pair token for the OptimismMintableERC20.
     */
    function _isCorrectTokenPair(address _mintableToken, address _otherToken)
    internal
    view
    returns (bool)
    {
        if (
            ERC165Checker.supportsInterface(_mintableToken, type(ILegacyMintableERC20).interfaceId)
        ) {
            return _otherToken == ILegacyMintableERC20(_mintableToken).l1Token();
        } else {
            return _otherToken == IOptimismMintableERC20(_mintableToken).remoteToken();
        }
    }

    /** @notice Emits the ETHBridgeInitiated event and if necessary the appropriate legacy event
     *          when an ETH bridge is finalized on this chain.
     *
     * @param _from      Address of the sender.
     * @param _to        Address of the receiver.
     * @param _amount    Amount of ETH sent.
     * @param _extraData Extra data sent with the transaction.
     */
    function _emitETHBridgeInitiated(
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal virtual {
        emit ETHBridgeInitiated(_from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the ETHBridgeFinalized and if necessary the appropriate legacy event when an
     *         ETH bridge is finalized on this chain.
     *
     * @param _from      Address of the sender.
     * @param _to        Address of the receiver.
     * @param _amount    Amount of ETH sent.
     * @param _extraData Extra data sent with the transaction.
     */
    function _emitETHBridgeFinalized(
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal virtual {
        emit ETHBridgeFinalized(_from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the ERC20BridgeInitiated event and if necessary the appropriate legacy
     *         event when an ERC20 bridge is initiated to the other chain.
     *
     * @param _localToken  Address of the ERC20 on this chain.
     * @param _remoteToken Address of the ERC20 on the remote chain.
     * @param _from        Address of the sender.
     * @param _to          Address of the receiver.
     * @param _amount      Amount of the ERC20 sent.
     * @param _extraData   Extra data sent with the transaction.
     */
    function _emitERC20BridgeInitiated(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal virtual {
        emit ERC20BridgeInitiated(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the ERC20BridgeFinalized event and if necessary the appropriate legacy
     *         event when an ERC20 bridge is initiated to the other chain.
     *
     * @param _localToken  Address of the ERC20 on this chain.
     * @param _remoteToken Address of the ERC20 on the remote chain.
     * @param _from        Address of the sender.
     * @param _to          Address of the receiver.
     * @param _amount      Amount of the ERC20 sent.
     * @param _extraData   Extra data sent with the transaction.
     */
    function _emitERC20BridgeFinalized(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal virtual {
        emit ERC20BridgeFinalized(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }
}
