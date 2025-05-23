// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// Scripts
import { Vm } from "forge-std/Vm.sol";
import { console2 as console } from "forge-std/console2.sol";
// import { Artifacts } from "scripts/Artifacts.s.sol";

// // Libraries
import { LibString } from "@solady/utils/LibString.sol";
import { Bytes } from "src/libraries/Bytes.sol";
// import { Constants } from "src/libraries/Constants.sol";
// import { Blueprint } from "src/libraries/Blueprint.sol";

// // Interfaces
// import { IProxy } from "interfaces/universal/IProxy.sol";
// import { IAddressManager } from "interfaces/legacy/IAddressManager.sol";
// import { IL1ChugSplashProxy, IStaticL1ChugSplashProxy } from "interfaces/legacy/IL1ChugSplashProxy.sol";
// import { IResolvedDelegateProxy } from "interfaces/legacy/IResolvedDelegateProxy.sol";
// import { IReinitializableBase } from "interfaces/universal/IReinitializableBase.sol";

library DeployUtils {
    Vm internal constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    bytes32 internal constant DEFAULT_SALT = keccak256("op-stack-contract-impls-salt-v0");

    /// @notice Deploys a contract with the given name and arguments via CREATE.
    /// @param _name Name of the contract to deploy.
    /// @param _args ABI-encoded constructor arguments.
    /// @return addr_ Address of the deployed contract.
    function create1(string memory _name, bytes memory _args) internal returns (address payable addr_) {
        bytes memory bytecode = abi.encodePacked(vm.getCode(_name), _args);
        assembly {
            addr_ := create(0, add(bytecode, 0x20), mload(bytecode))
        }
        assertValidContractAddress(addr_);
    }

    // /// @notice Deploys a contract with the given name and arguments via CREATE and saves the result.
    // /// @param _save Artifacts contract.
    // /// @param _name Name of the contract to deploy.
    // /// @param _nick Nickname to save the address to.
    // /// @param _args ABI-encoded constructor arguments.
    // /// @return addr_ Address of the deployed contract.
    // function create1AndSave(
    //     Artifacts _save,
    //     string memory _name,
    //     string memory _nick,
    //     bytes memory _args
    // )
    //     internal
    //     returns (address payable addr_)
    // {
    //     console.log("Deploying %s", _nick);
    //     addr_ = create1(_name, _args);
    //     _save.save(_nick, addr_);
    //     console.log("%s deployed at %s", _nick, addr_);
    // }

    // /// @notice Deploys a contract with the given name and arguments via CREATE and saves the result.
    // /// @param _save Artifacts contract.
    // /// @param _name Name of the contract to deploy.
    // /// @param _args ABI-encoded constructor arguments.
    // /// @return addr_ Address of the deployed contract.
    // function create1AndSave(
    //     Artifacts _save,
    //     string memory _name,
    //     bytes memory _args
    // )
    //     internal
    //     returns (address payable addr_)
    // {
    //     return create1AndSave(_save, _name, _name, _args);
    // }

    // /// @notice Deploys a contract with the given name and arguments via CREATE2.
    // /// @param _name Name of the contract to deploy.
    // /// @param _args ABI-encoded constructor arguments.
    // /// @param _salt Salt for the CREATE2 operation.
    // /// @return addr_ Address of the deployed contract.
    // function create2(string memory _name, bytes memory _args, bytes32 _salt) internal returns (address payable) {
    //     bytes memory initCode = abi.encodePacked(vm.getCode(_name), _args);
    //     address preComputedAddress = vm.computeCreate2Address(_salt, keccak256(initCode));
    //     require(preComputedAddress.code.length == 0, "DeployUtils: contract already deployed");
    //     return create2asm(initCode, _salt);
    // }

    // function create2asm(bytes memory _initCode, bytes32 _salt) private returns (address payable addr_) {
    //     assembly {
    //         addr_ := create2(0, add(_initCode, 0x20), mload(_initCode), _salt)
    //         if iszero(addr_) {
    //             let size := returndatasize()
    //             returndatacopy(0, 0, size)
    //             revert(0, size)
    //         }
    //     }
    //     assertValidContractAddress(addr_);
    // }

    // /// @notice Deploys a contract with the given name and arguments via CREATE2 and saves the result.
    // /// @param _save Artifacts contract.
    // /// @param _name Name of the contract to deploy.
    // /// @param _nick Nickname to save the address to.
    // /// @param _args ABI-encoded constructor arguments.
    // /// @param _salt Salt for the CREATE2 operation.
    // /// @return addr_ Address of the deployed contract.
    // function create2AndSave(
    //     Artifacts _save,
    //     string memory _name,
    //     string memory _nick,
    //     bytes memory _args,
    //     bytes32 _salt
    // )
    //     internal
    //     returns (address payable addr_)
    // {
    //     console.log("Deploying %s", _nick);
    //     addr_ = create2(_name, _args, _salt);
    //     _save.save(_nick, addr_);
    //     console.log("%s deployed at %s", _nick, addr_);
    // }

    // /// @notice Deploys a contract with the given name and arguments via CREATE2 and saves the result.
    // /// @param _save Artifacts contract.
    // /// @param _name Name of the contract to deploy.
    // /// @param _args ABI-encoded constructor arguments.
    // /// @param _salt Salt for the CREATE2 operation.
    // /// @return addr_ Address of the deployed contract.
    // function create2AndSave(
    //     Artifacts _save,
    //     string memory _name,
    //     bytes memory _args,
    //     bytes32 _salt
    // )
    //     internal
    //     returns (address payable addr_)
    // {
    //     return create2AndSave(_save, _name, _name, _args, _salt);
    // }

    // /// @notice Deploys a contract with the given name using CREATE2. If the contract is already deployed, this
    // method
    // /// does nothing.
    // /// @param _name Name of the contract to deploy.
    // /// @param _args ABI-encoded constructor arguments.
    // function createDeterministic(
    //     string memory _name,
    //     bytes memory _args,
    //     bytes32 _salt
    // )
    //     internal
    //     returns (address payable addr_)
    // {
    //     bytes memory initCode = abi.encodePacked(vm.getCode(_name), _args);
    //     address preComputedAddress = vm.computeCreate2Address(_salt, keccak256(initCode));
    //     if (preComputedAddress.code.length > 0) {
    //         addr_ = payable(preComputedAddress);
    //     } else {
    //         vm.broadcast(msg.sender);
    //         addr_ = create2asm(initCode, _salt);
    //     }
    // }

    // /// @notice Deploys a blueprint contract with the given name using CREATE2. If the contract is already deployed,
    // /// this method does nothing.
    // /// @param _rawBytecode Raw bytecode of the contract the blueprint will deploy.
    // function createDeterministicBlueprint(
    //     bytes memory _rawBytecode,
    //     bytes32 _salt
    // )
    //     internal
    //     returns (address newContract1_, address newContract2_)
    // {
    //     uint32 maxSize = Blueprint.maxInitCodeSize();
    //     if (_rawBytecode.length <= maxSize) {
    //         bytes memory bpBytecode = Blueprint.blueprintDeployerBytecode(_rawBytecode);
    //         newContract1_ = vm.computeCreate2Address(_salt, keccak256(bpBytecode));
    //         if (newContract1_.code.length == 0) {
    //             (address deployedContract) = Blueprint.deploySmallBytecode(bpBytecode, _salt);
    //             require(deployedContract == newContract1_, "DeployUtils: unexpected blueprint address");
    //         }
    //         newContract2_ = address(0);
    //     } else {
    //         bytes memory part1Slice = Bytes.slice(_rawBytecode, 0, maxSize);
    //         bytes memory part2Slice = Bytes.slice(_rawBytecode, maxSize, _rawBytecode.length - maxSize);
    //         bytes memory bp1Bytecode = Blueprint.blueprintDeployerBytecode(part1Slice);
    //         bytes memory bp2Bytecode = Blueprint.blueprintDeployerBytecode(part2Slice);
    //         newContract1_ = vm.computeCreate2Address(_salt, keccak256(bp1Bytecode));
    //         if (newContract1_.code.length == 0) {
    //             address deployedContract = Blueprint.deploySmallBytecode(bp1Bytecode, _salt);
    //             require(deployedContract == newContract1_, "DeployUtils: unexpected part 1 blueprint address");
    //         }
    //         newContract2_ = vm.computeCreate2Address(_salt, keccak256(bp2Bytecode));
    //         if (newContract2_.code.length == 0) {
    //             address deployedContract = Blueprint.deploySmallBytecode(bp2Bytecode, _salt);
    //             require(deployedContract == newContract2_, "DeployUtils: unexpected part 2 blueprint address");
    //         }
    //     }
    // }

    // /// @notice Takes a sender and an identifier and returns a deterministic address based on the
    // ///         two. The result is used to etch the input and output contracts to a deterministic
    // ///         address based on those two values, where the identifier represents the input or
    // ///         output contract, such as `optimism.DeploySuperchainInput` or
    // ///         `optimism.DeployOPChainOutput`.
    // ///         Example: `toIOAddress(msg.sender, "optimism.DeploySuperchainInput")`
    // /// @param _sender Address of the sender.
    // /// @param _identifier Additional identifier.
    // /// @return Deterministic address.
    // function toIOAddress(address _sender, string memory _identifier) internal pure returns (address) {
    //     return address(uint160(uint256(keccak256(abi.encode(_sender, _identifier)))));
    // }

    /// @notice Strips the first 4 bytes of `_data` and returns the remaining bytes
    ///         If `_data` is not greater than 4 bytes, it returns empty bytes type.
    /// @param _data constructor arguments prefixed with a psuedo-constructor function signature
    /// @return encodedData_ constructor arguments without the psuedo-constructor function signature prefix
    function encodeConstructor(bytes memory _data) internal pure returns (bytes memory encodedData_) {
        require(_data.length >= 4, "DeployUtils: encodeConstructor takes in _data of length >= 4");
        encodedData_ = Bytes.slice(_data, 4);
    }

    /// @notice Asserts that the given address is a valid contract address.
    /// @param _who Address to check.
    function assertValidContractAddress(address _who) internal view {
        // Foundry will set returned address to address(1) whenever a contract creation fails
        // inside of a test. If this is the case then let Foundry handle the error itself and don't
        // trigger a revert (which would probably break a test).
        if (_who == address(1)) return;
        require(_who != address(0), "DeployUtils: zero address");
        require(_who.code.length > 0, string.concat("DeployUtils: no code at ", LibString.toHexStringChecksummed(_who)));
    }

    // /// @notice Asserts that the given proxy has an implementation set.
    // /// @param _proxy Proxy to check.
    // function assertERC1967ImplementationSet(address _proxy) internal returns (address implementation_) {
    //     // We prank as the zero address due to the Proxy's `proxyCallIfNotAdmin` modifier.
    //     // Pranking inside this function also means it can no longer be considered `view`.
    //     vm.prank(address(0));
    //     implementation_ = IProxy(payable(_proxy)).implementation();
    //     assertValidContractAddress(implementation_);
    // }

    // /// @notice Asserts that the given L1ChugSplashProxy has an implementation set.
    // /// @param _proxy L1ChugSplashProxy to check.
    // function assertL1ChugSplashImplementationSet(address _proxy) internal returns (address implementation_) {
    //     vm.prank(address(0));
    //     implementation_ = IStaticL1ChugSplashProxy(_proxy).getImplementation();
    //     assertValidContractAddress(implementation_);
    // }

    // /// @notice Asserts that the given ResolvedDelegateProxy has an implementation set.
    // /// @param _implementationName Name of the implementation contract.
    // /// @param _addressManager AddressManager contract.
    // function assertResolvedDelegateProxyImplementationSet(
    //     string memory _implementationName,
    //     IAddressManager _addressManager
    // )
    //     internal
    //     view
    //     returns (address implementation_)
    // {
    //     implementation_ = _addressManager.getAddress(_implementationName);
    //     assertValidContractAddress(implementation_);
    // }

    // /// @notice Builds an ERC1967 Proxy with a dummy implementation.
    // /// @param _proxyImplName Name of the implementation contract.
    // function buildERC1967ProxyWithImpl(string memory _proxyImplName) internal returns (IProxy genericProxy_) {
    //     genericProxy_ = IProxy(
    //         create1({
    //             _name: "Proxy",
    //             _args: DeployUtils.encodeConstructor(abi.encodeCall(IProxy.__constructor__, (address(0))))
    //         })
    //     );
    //     address implementation = address(vm.addr(uint256(keccak256(abi.encodePacked(_proxyImplName)))));
    //     vm.etch(address(implementation), hex"01");
    //     vm.prank(address(0));
    //     genericProxy_.upgradeTo(address(implementation));
    //     vm.etch(address(genericProxy_), address(genericProxy_).code);
    // }

    // /// @notice Builds an L1ChugSplashProxy with a dummy implementation.
    // /// @param _proxyImplName Name of the implementation contract.
    // function buildL1ChugSplashProxyWithImpl(string memory _proxyImplName)
    //     internal
    //     returns (IL1ChugSplashProxy proxy_)
    // {
    //     proxy_ = IL1ChugSplashProxy(
    //         create1({
    //             _name: "L1ChugSplashProxy",
    //             _args: DeployUtils.encodeConstructor(abi.encodeCall(IL1ChugSplashProxy.__constructor__,
    // (address(0))))
    //         })
    //     );
    //     address implementation = address(vm.addr(uint256(keccak256(abi.encodePacked(_proxyImplName)))));
    //     vm.etch(address(implementation), hex"01");
    //     vm.prank(address(0));
    //     proxy_.setStorage(Constants.PROXY_IMPLEMENTATION_ADDRESS, bytes32(uint256(uint160(implementation))));
    // }

    // /// @notice Builds a ResolvedDelegateProxy with a dummy implementation.
    // /// @param _addressManager AddressManager contract.
    // /// @param _proxyImplName Name of the implementation contract.
    // function buildResolvedDelegateProxyWithImpl(
    //     IAddressManager _addressManager,
    //     string memory _proxyImplName
    // )
    //     internal
    //     returns (IResolvedDelegateProxy proxy_)
    // {
    //     proxy_ = IResolvedDelegateProxy(
    //         create1({
    //             _name: "ResolvedDelegateProxy",
    //             _args: DeployUtils.encodeConstructor(
    //                 abi.encodeCall(IResolvedDelegateProxy.__constructor__, (_addressManager, _proxyImplName))
    //             )
    //         })
    //     );
    //     address implementation = address(vm.addr(uint256(keccak256(abi.encodePacked(_proxyImplName)))));
    //     vm.etch(address(implementation), hex"01");
    //     _addressManager.setAddress(_proxyImplName, implementation);
    // }

    // /// @notice Builds an AddressManager contract.
    // function buildAddressManager() internal returns (IAddressManager addressManager_) {
    //     addressManager_ = IAddressManager(
    //         create1({
    //             _name: "AddressManager",
    //             _args: DeployUtils.encodeConstructor(abi.encodeCall(IAddressManager.__constructor__, ()))
    //         })
    //     );
    // }

    // /// @notice Asserts that the given list of addresses does not contain duplicates.
    // /// @param _addrs Addresses to check.
    // function assertUniqueAddresses(address[] memory _addrs) internal pure {
    //     // All addresses should be unique.
    //     for (uint256 i = 0; i < _addrs.length; i++) {
    //         for (uint256 j = i + 1; j < _addrs.length; j++) {
    //             require(
    //                 _addrs[i] != _addrs[j],
    //                 string.concat(
    //                     "DeployUtils: check failed, duplicates at ", LibString.toString(i), ",",
    // LibString.toString(j)
    //                 )
    //             );
    //         }
    //     }
    // }

    // /// @notice Asserts that the given addresses are valid contract addresses.
    // /// @param _addrs Addresses to check.
    // function assertValidContractAddresses(address[] memory _addrs) internal view {
    //     // Assert that all addresses are non-zero and have code.
    //     // We use LibString to avoid the need for adding cheatcodes to this contract.
    //     for (uint256 i = 0; i < _addrs.length; i++) {
    //         address who = _addrs[i];
    //         assertValidContractAddress(who);
    //     }

    //     // All addresses should be unique.
    //     assertUniqueAddresses(_addrs);
    // }

    // /// @dev Asserts that for a given contract the value of a storage slot at an offset is 1 (if a proxy contract) or
    // ///      type(uint8).max (if an implementation contract).
    // ///      A call to `initialize` will set proxies to 1 and a call to _disableInitializers will set implementations
    // to
    // ///      type(uint8).max.
    // function assertInitialized(address _contractAddress, bool _isProxy, uint256 _slot, uint256 _offset) internal view
    // {
    //     bytes32 slotVal = vm.load(_contractAddress, bytes32(_slot));
    //     uint8 val = uint8((uint256(slotVal) >> (_offset * 8)) & 0xFF);
    //     if (_isProxy) {
    //         // Using a try/catch here to check if the contract has an initVersion() defined.
    //         // EIP-150 safe because we require that we have at least 200k gas before the call which
    //         // is more than enough to avoid running out of gas when 63/64 of the gas is provided to
    //         // the initVersion() call (which simply reads an immutable variable). Since this is
    //         // only ever triggered as part of a script, we can safely assume we'll have the gas.
    //         require(gasleft() > 200_000, "DeployUtils: insufficient gas for initVersion() call");

    //         // eip150-safe
    //         try IReinitializableBase(_contractAddress).initVersion() returns (uint8 initVersion_) {
    //             require(val == initVersion_, "DeployUtils: storage value is incorrect at the given slot and offset");
    //         } catch {
    //             require(val == 1, "DeployUtils: storage value is not set at the given slot and offset");
    //         }
    //     } else {
    //         require(val == type(uint8).max, "DeployUtils: storage value is not 0xff at the given slot and offset");
    //     }
    // }

    /// @notice Etches a contract, labels it, and allows cheatcodes for it.
    /// @param _etchTo Address of the contract to etch.
    /// @param _cname The contract name (also used to label the contract).
    /// @param _artifactPath The path to the artifact to etch.
    function etchLabelAndAllowCheatcodes(address _etchTo, string memory _cname, string memory _artifactPath) internal {
        vm.etch(_etchTo, vm.getDeployedCode(_artifactPath));
        vm.label(_etchTo, _cname);
        vm.allowCheatcodes(_etchTo);
    }

    /// @notice Etches a contract, labels it, and allows cheatcodes for it.
    /// @param _etchTo Address of the contract to etch.
    /// @param _cname The contract name (also used to label the contract). MUST be the name of both the file and the
    ///               contract.
    function etchLabelAndAllowCheatcodes(address _etchTo, string memory _cname) internal {
        string memory artifactPath = string.concat(_cname, ".s.sol:", _cname);
        etchLabelAndAllowCheatcodes(_etchTo, _cname, artifactPath);
    }
}
