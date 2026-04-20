// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

interface IL2ERC721Bridge {
    function finalizeBridgeERC721(
        address _localToken,
        address _remoteToken,
        address _from,
        address _to,
        uint256 _tokenId,
        bytes calldata _extraData
    )
        external;
    function version() external view returns (string memory);

    function __constructor__(address _messenger, address _otherBridge) external;
}
