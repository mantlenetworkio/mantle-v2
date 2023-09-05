// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { BaseL2CrossDomainMessenger } from "./BaseL2CrossDomainMessenger.sol";
import { Address } from "@openzeppelin/contracts/utils/Address.sol";
import {ERC721Bridge} from  "../universal/ERC721Bridge.sol";

/**
 * @title ERC721Bridge
 * @notice ERC721Bridge is a base contract for the L1 and L2 ERC721 bridges.
 */
abstract contract BaseL2ERC721Bridge is ERC721Bridge{
    /**
     * @notice Messenger contract on this domain.
     */
    BaseL2CrossDomainMessenger public immutable MESSENGER;

    /**
     * @notice Address of the bridge on the other network.
     */
    address public immutable OTHER_BRIDGE;



    /**
     * @notice Ensures that the caller is a cross-chain message from the other bridge.
     */
    modifier onlyOtherBridge() {
        require(
            msg.sender == address(MESSENGER) && MESSENGER.xDomainMessageSender() == OTHER_BRIDGE,
            "ERC721Bridge: function can only be called from the other bridge"
        );
        _;
    }

    /**
     * @param _messenger   Address of the CrossDomainMessenger on this network.
     * @param _otherBridge Address of the ERC721 bridge on the other network.
     */
    constructor(address _messenger, address _otherBridge) {
        require(_messenger != address(0), "ERC721Bridge: messenger cannot be address(0)");
        require(_otherBridge != address(0), "ERC721Bridge: other bridge cannot be address(0)");

        MESSENGER = BaseL2CrossDomainMessenger(_messenger);
        OTHER_BRIDGE = _otherBridge;
    }

    /**
     * @custom:legacy
     * @notice Legacy getter for messenger contract.
     *
     * @return Messenger contract on this domain.
     */
    function messenger() external view returns (BaseL2CrossDomainMessenger) {
        return MESSENGER;
    }

    /**
     * @custom:legacy
     * @notice Legacy getter for other bridge address.
     *
     * @return Address of the bridge on the other network.
     */
    function otherBridge() external view returns (address) {
        return OTHER_BRIDGE;
    }

    /**
     * @notice Initiates a bridge of an NFT to the caller's account on the other chain. Note that
     *         this function can only be called by EOAs. Smart contract wallets should use the
     *         `bridgeERC721To` function after ensuring that the recipient address on the remote
     *         chain exists. Also note that the current owner of the token on this chain must
     *         approve this contract to operate the NFT before it can be bridged.
     *         **WARNING**: Do not bridge an ERC721 that was originally deployed on Optimism. This
     *         bridge only supports ERC721s originally deployed on Ethereum. Users will need to
     *         wait for the one-week challenge period to elapse before their Optimism-native NFT
     *         can be refunded on L2.
     *
     * @param _localToken  Address of the ERC721 on this domain.
     * @param _remoteToken Address of the ERC721 on the remote domain.
     * @param _tokenId     Token ID to bridge.
     * @param _minGasLimit Minimum gas limit for the bridge message on the other domain.
     * @param _extraData   Optional data to forward to the other chain. Data supplied here will not
     *                     be used to execute any code on the other chain and is only emitted as
     *                     extra data for the convenience of off-chain tooling.
     */
    function bridgeERC721(
        address _localToken,
        address _remoteToken,
        uint256 _tokenId,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) external override {
        // Modifier requiring sender to be EOA. This prevents against a user error that would occur
        // if the sender is a smart contract wallet that has a different address on the remote chain
        // (or doesn't have an address on the remote chain at all). The user would fail to receive
        // the NFT if they use this function because it sends the NFT to the same address as the
        // caller. This check could be bypassed by a malicious contract via initcode, but it takes
        // care of the user error we want to avoid.
        require(!Address.isContract(msg.sender), "ERC721Bridge: account is not externally owned");

        _initiateBridgeERC721(
            _localToken,
            _remoteToken,
            msg.sender,
            msg.sender,
            _tokenId,
            _minGasLimit,
            _extraData
        );
    }

    /**
     * @notice Initiates a bridge of an NFT to some recipient's account on the other chain. Note
     *         that the current owner of the token on this chain must approve this contract to
     *         operate the NFT before it can be bridged.
     *         **WARNING**: Do not bridge an ERC721 that was originally deployed on Optimism. This
     *         bridge only supports ERC721s originally deployed on Ethereum. Users will need to
     *         wait for the one-week challenge period to elapse before their Optimism-native NFT
     *         can be refunded on L2.
     *
     * @param _localToken  Address of the ERC721 on this domain.
     * @param _remoteToken Address of the ERC721 on the remote domain.
     * @param _to          Address to receive the token on the other domain.
     * @param _tokenId     Token ID to bridge.
     * @param _minGasLimit Minimum gas limit for the bridge message on the other domain.
     * @param _extraData   Optional data to forward to the other chain. Data supplied here will not
     *                     be used to execute any code on the other chain and is only emitted as
     *                     extra data for the convenience of off-chain tooling.
     */
    function bridgeERC721To(
        address _localToken,
        address _remoteToken,
        address _to,
        uint256 _tokenId,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) external override {
        require(_to != address(0), "ERC721Bridge: nft recipient cannot be address(0)");

        _initiateBridgeERC721(
            _localToken,
            _remoteToken,
            msg.sender,
            _to,
            _tokenId,
            _minGasLimit,
            _extraData
        );
    }


}
