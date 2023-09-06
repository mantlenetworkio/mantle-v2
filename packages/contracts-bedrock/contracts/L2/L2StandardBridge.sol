// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Predeploys } from "../libraries/Predeploys.sol";
import { StandardBridge } from "../universal/StandardBridge.sol";
import { Semver } from "../universal/Semver.sol";
import { SafeCall } from "../libraries/SafeCall.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { OptimismMintableERC20 } from "../universal/OptimismMintableERC20.sol";
import { BridgeConstants } from "../libraries/BridgeConstants.sol";

/**
 * @custom:proxied
 * @custom:predeploy 0x4200000000000000000000000000000000000010
 * @title L2StandardBridge
 * @notice The L2StandardBridge is responsible for transfering ETH and ERC20 tokens between L1 and
 *         L2. In the case that an ERC20 token is native to L2, it will be escrowed within this
 *         contract. If the ERC20 token is native to L1, it will be burnt.
 *         NOTE: this contract is not intended to support all variations of ERC20 tokens. Examples
 *         of some token types that may not be properly supported by this contract include, but are
 *         not limited to: tokens with transfer fees, rebasing tokens, and tokens with blocklists.
 */
contract L2StandardBridge is StandardBridge, Semver {
    address L1_MNT_ADDRESS;

    using SafeERC20 for IERC20;
    /**
     * @custom:legacy
     * @notice Emitted whenever a withdrawal from L2 to L1 is initiated.
     *
     * @param l1Token   Address of the token on L1.
     * @param l2Token   Address of the corresponding token on L2.
     * @param from      Address of the withdrawer.
     * @param to        Address of the recipient on L1.
     * @param amount    Amount of the ERC20 withdrawn.
     * @param extraData Extra data attached to the withdrawal.
     */
    event WithdrawalInitiated(
        address indexed l1Token,
        address indexed l2Token,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    /**
     * @custom:legacy
     * @notice Emitted whenever an ERC20 deposit is finalized.
     *
     * @param l1Token   Address of the token on L1.
     * @param l2Token   Address of the corresponding token on L2.
     * @param from      Address of the depositor.
     * @param to        Address of the recipient on L2.
     * @param amount    Amount of the ERC20 deposited.
     * @param extraData Extra data attached to the deposit.
     */
    event DepositFinalized(
        address indexed l1Token,
        address indexed l2Token,
        address indexed from,
        address to,
        uint256 amount,
        bytes extraData
    );

    /**
     * @custom:semver 1.1.0
     *
     * @param _otherBridge Address of the L1StandardBridge.
     */
    constructor(address payable _otherBridge)
        Semver(1, 1, 0)
        StandardBridge(payable(Predeploys.L2_CROSS_DOMAIN_MESSENGER), _otherBridge)
    {
        L1_MNT_ADDRESS = BridgeConstants.L1_MNT;
    }

    /**
     * @notice Allows EOAs to bridge ETH by sending directly to the bridge.
     */
    receive() external payable override onlyEOA {
        _initiateWithdrawal(
            Predeploys.LEGACY_ERC20_MNT,
            msg.sender,
            msg.sender,
            msg.value,
            RECEIVE_DEFAULT_GAS_LIMIT,
            bytes("")
        );
    }

    /**
     * @custom:legacy
     * @notice Initiates a withdrawal from L2 to L1.
     *         This function only works with OptimismMintableERC20 tokens or ether. Use the
     *         `bridgeERC20` function to bridge native L2 tokens to L1.
     *
     * @param _l2Token     Address of the L2 token to withdraw.
     * @param _amount      Amount of the L2 token to withdraw.
     * @param _minGasLimit Minimum gas limit to use for the transaction.
     * @param _extraData   Extra data attached to the withdrawal.
     */
    function withdraw(
        address _l2Token,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) external payable onlyEOA {
        _initiateWithdrawal(_l2Token, msg.sender, msg.sender, _amount, _minGasLimit, _extraData);
    }

    /**
     * @custom:legacy
     * @notice Initiates a withdrawal from L2 to L1 to a target account on L1.
     *         Note that if ETH is sent to a contract on L1 and the call fails, then that ETH will
     *         be locked in the L1StandardBridge. ETH may be recoverable if the call can be
     *         successfully replayed by increasing the amount of gas supplied to the call. If the
     *         call will fail for any amount of gas, then the ETH will be locked permanently.
     *         This function only works with OptimismMintableERC20 tokens or ether. Use the
     *         `bridgeERC20To` function to bridge native L2 tokens to L1.
     *
     * @param _l2Token     Address of the L2 token to withdraw.
     * @param _to          Recipient account on L1.
     * @param _amount      Amount of the L2 token to withdraw.
     * @param _minGasLimit Minimum gas limit to use for the transaction.
     * @param _extraData   Extra data attached to the withdrawal.
     */
    function withdrawTo(
        address _l2Token,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes calldata _extraData
    ) external payable {
        _initiateWithdrawal(_l2Token, msg.sender, _to, _amount, _minGasLimit, _extraData);
    }

    /**
     * @custom:legacy
     * @notice Finalizes a deposit from L1 to L2. To finalize a deposit of ether, use address(0)
     *         and the l1Token and the Legacy ERC20 ether predeploy address as the l2Token.
     *
     * @param _l1Token   Address of the L1 token to deposit.
     * @param _l2Token   Address of the corresponding L2 token.
     * @param _from      Address of the depositor.
     * @param _to        Address of the recipient.
     * @param _amount    Amount of the tokens being deposited.
     * @param _extraData Extra data attached to the deposit.
     */
    function finalizeDeposit(
        address _l1Token,
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        bytes calldata _extraData
    ) external payable {
        if (_l1Token == BridgeConstants.L1_MNT && _l2Token == Predeploys.LEGACY_ERC20_MNT) {
            finalizeBridgeMNT(_l2Token,_l1Token,_from, _to, _amount, _extraData);
        } else if (_l1Token == address(0) && _l2Token == Predeploys.BVM_ETH){
            finalizeBridgeETH(_l2Token,_l1Token,_from, _to, _amount, _extraData);
        } else{
            finalizeBridgeERC20(_l2Token, _l1Token, _from, _to, _amount, _extraData);
        }
    }

    /**
     * @custom:legacy
     * @notice Retrieves the access of the corresponding L1 bridge contract.
     *
     * @return Address of the corresponding L1 bridge contract.
     */
    function l1TokenBridge() external view returns (address) {
        return address(OTHER_BRIDGE);
    }

    /**
     * @custom:legacy
     * @notice Internal function to a withdrawal from L2 to L1 to a target account on L1.
     *
     * @param _l2Token     Address of the L2 token to withdraw.
     * @param _from        Address of the withdrawer.
     * @param _to          Recipient account on L1.
     * @param _amount      Amount of the L2 token to withdraw.
     * @param _minGasLimit Minimum gas limit to use for the transaction.
     * @param _extraData   Extra data attached to the withdrawal.
     */
    function _initiateWithdrawal(
        address _l2Token,
        address _from,
        address _to,
        uint256 _amount,
        uint32 _minGasLimit,
        bytes memory _extraData
    ) internal {
        if (_l2Token == Predeploys.BVM_ETH ) {
            _initiateBridgeETH(_l2Token,address(0), _from, _to, _amount, _minGasLimit, _extraData);
        } else if (_l2Token == address(0) ){
            require(msg.value==_amount,"L2StandardBridge :_amount must equal the MNT value.");
            _initiateBridgeMNT(address(0), BridgeConstants.L1_MNT,_from, _to, _amount, _minGasLimit, _extraData);
        } else {
            address l1Token = OptimismMintableERC20(_l2Token).l1Token();
            _initiateBridgeERC20(_l2Token, l1Token, _from, _to, _amount, _minGasLimit, _extraData);
        }
    }

    /**
     * @notice Emits the legacy WithdrawalInitiated event followed by the ETHBridgeInitiated event.
     *         This is necessary for backwards compatibility with the legacy bridge.
     *
     * @inheritdoc StandardBridge
     */
    function _emitETHBridgeInitiated(
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal override {
        emit WithdrawalInitiated(
            address(0),
            Predeploys.BVM_ETH,
            _from,
            _to,
            _amount,
            _extraData
        );
        super._emitETHBridgeInitiated(_from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the legacy DepositFinalized event followed by the ETHBridgeFinalized event.
     *         This is necessary for backwards compatibility with the legacy bridge.
     *
     * @inheritdoc StandardBridge
     */
    function _emitETHBridgeFinalized(
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal override {
        emit DepositFinalized(
            address(0),
            Predeploys.BVM_ETH,
            _from,
            _to,
            _amount,
            _extraData
        );
        super._emitETHBridgeFinalized(_from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the legacy WithdrawalInitiated event followed by the ERC20BridgeInitiated
     *         event. This is necessary for backwards compatibility with the legacy bridge.
     *
     * @inheritdoc StandardBridge
     */
    function _emitERC20BridgeInitiated(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal override {
        emit WithdrawalInitiated(_remoteToken, _localToken, _from, _to, _amount, _extraData);
        super._emitERC20BridgeInitiated(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the legacy DepositFinalized event followed by the ERC20BridgeFinalized event.
     *         This is necessary for backwards compatibility with the legacy bridge.
     *
     * @inheritdoc StandardBridge
     */
    function _emitERC20BridgeFinalized(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal override {
        emit DepositFinalized(_remoteToken, _localToken, _from, _to, _amount, _extraData);
        super._emitERC20BridgeFinalized(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }

    /**
     * @notice Emits the legacy WithdrawalInitiated event followed by the MNTBridgeInitiated
     *         event. This is necessary for backwards compatibility with the legacy bridge.
     *
     * @inheritdoc StandardBridge
     */
    function _emitMNTBridgeInitiated(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal override {
        emit WithdrawalInitiated(_remoteToken, _localToken, _from, _to, _amount, _extraData);
        super._emitMNTBridgeInitiated(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }

    /**
 * @notice Emits the legacy DepositFinalized event followed by the ERC20BridgeFinalized event.
     *         This is necessary for backwards compatibility with the legacy bridge.
     *
     * @inheritdoc StandardBridge
     */
    function _emitMNTBridgeFinalized(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _amount,
        bytes memory _extraData
    ) internal override {
        emit DepositFinalized(_remoteToken, _localToken, _from, _to, _amount, _extraData);
        super._emitMNTBridgeFinalized(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }

    /**
     * @notice Sends ETH to the sender's address on the other chain.
     *
     * @param _minGasLimit Minimum amount of gas that the bridge can be relayed with.
     * @param _extraData   Extra data to be sent with the transaction. Note that the recipient will
     *                     not be triggered with this data, but it will be emitted and can be used
     *                     to identify the transaction.
     */
    function bridgeETH(uint32 _minGasLimit, bytes calldata _extraData) public payable override onlyEOA {
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
    ) public payable override {
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
    ) public onlyEOA override {
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
    ) public override {
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
    ) public payable onlyOtherBridge override {
        require(msg.value == _amount, "StandardBridge: amount sent does not match amount required");
        require(_to != address(this), "StandardBridge: cannot send to self");
        require(_to != address(MESSENGER), "StandardBridge: cannot send to messenger");

        // Emit the correct events. By default this will be _amount, but child
        // contracts may override this function in order to emit legacy events as well.
        _emitETHBridgeFinalized(_from, _to, _amount, _extraData);

        bool success = SafeCall.call(_to, gasleft(), _amount, hex"");
        require(success, "StandardBridge: ETH transfer failed");
    }

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
    ) public onlyOtherBridge override {
        if (_isOptimismMintableERC20(_localToken)) {
            require(
                _isCorrectTokenPair(_localToken, _remoteToken),
                "StandardBridge: wrong remote token for Optimism Mintable ERC20 local token"
            );

            OptimismMintableERC20(_localToken).mint(_to, _amount);
        } else {
            deposits[_localToken][_remoteToken] = deposits[_localToken][_remoteToken] - _amount;
            IERC20(_localToken).safeTransfer(_to, _amount);
        }

        // Emit the correct events. By default this will be ERC20BridgeFinalized, but child
        // contracts may override this function in order to emit legacy events as well.
        _emitERC20BridgeFinalized(_localToken, _remoteToken, _from, _to, _amount, _extraData);
    }
}
