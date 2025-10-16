// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library Types {
    /// @notice Represents a set of L1 contracts. Used to represent a set of proxies.
    /// This is not an exhaustive list of all contracts on L1, but rather a subset.
    struct ContractSet {
        address ProxyAdmin;
        address AddressManager;
        address L1CrossDomainMessenger;
        address L1StandardBridge;
        address L2OutputOracle;
        address OptimismMintableERC20Factory;
        address OptimismPortal;
        address SystemConfig;
        address L1ERC721Bridge;
        address L1CrossDomainMessengerImpl;
        address L1StandardBridgeImpl;
        address L2OutputOracleImpl;
        address OptimismMintableERC20FactoryImpl;
        address OptimismPortalImpl;
        address SystemConfigImpl;
        address L1ERC721BridgeImpl;
    }
}
