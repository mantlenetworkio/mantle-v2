// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Bytes } from "src/libraries/Bytes.sol";

/// @notice Methods for working with ERC-5202 blueprint contracts.
/// https://eips.ethereum.org/EIPS/eip-5202
library Blueprint {
    /// @notice The structure of a blueprint contract per ERC-5202.
    struct Preamble {
        uint8 ercVersion;
        bytes preambleData;
        bytes initcode;
    }

    /// @notice Thrown when converting a bytes array to a uint256 and the bytes array is too long.
    error BytesArrayTooLong();

    /// @notice Throw when contract deployment fails.
    error DeploymentFailed();

    /// @notice Thrown when parsing a blueprint preamble and the resulting initcode is empty.
    error EmptyInitcode();

    /// @notice Thrown when call to the identity precompile fails.
    error IdentityPrecompileCallFailed();

    /// @notice Thrown when parsing a blueprint preamble and the bytecode does not contain the expected prefix bytes.
    error NotABlueprint();

    /// @notice Thrown when parsing a blueprint preamble and the reserved bits are set.
    error ReservedBitsSet();

    /// @notice Thrown when parsing a blueprint preamble and the preamble data is not empty.
    /// We do not use the preamble data, so it's expected to be empty.
    error UnexpectedPreambleData(bytes data);

    /// @notice Thrown during deployment if the ERC version is not supported.
    error UnsupportedERCVersion(uint8 version);

    /// @notice Takes the desired initcode for a blueprint as a parameter, and returns EVM code
    /// which will deploy a corresponding blueprint contract (with no data section). Based on the
    /// reference implementation in https://eips.ethereum.org/EIPS/eip-5202.
    function blueprintDeployerBytecode(bytes memory _initcode) internal pure returns (bytes memory) {
        // Check that the initcode is not empty.
        if (_initcode.length == 0) revert EmptyInitcode();

        bytes memory blueprintPreamble = hex"FE7100"; // ERC-5202 preamble.
        bytes memory blueprintBytecode = bytes.concat(blueprintPreamble, _initcode);

        // The length of the deployed code in bytes.
        bytes2 lenBytes = bytes2(uint16(blueprintBytecode.length));

        // Copy <blueprintBytecode> to memory and `RETURN` it per EVM creation semantics.
        // PUSH2 <len> RETURNDATASIZE DUP2 PUSH1 10 RETURNDATASIZE CODECOPY RETURN
        bytes memory deployBytecode = bytes.concat(hex"61", lenBytes, hex"3d81600a3d39f3");

        return bytes.concat(deployBytecode, blueprintBytecode);
    }

    /// @notice Given bytecode as a sequence of bytes, parse the blueprint preamble and deconstruct
    /// the bytecode into the ERC version, preamble data and initcode. Reverts if the bytecode is
    /// not a valid blueprint contract according to ERC-5202.
    function parseBlueprintPreamble(bytes memory _bytecode) internal view returns (Preamble memory) {
        if (_bytecode.length < 2 || _bytecode[0] != 0xFE || _bytecode[1] != 0x71) {
            revert NotABlueprint();
        }

        uint8 ercVersion = uint8(_bytecode[2] & 0xFC) >> 2;
        uint8 nLengthBytes = uint8(_bytecode[2] & 0x03);
        if (nLengthBytes == 0x03) revert ReservedBitsSet();

        uint256 dataLength = 0;
        if (nLengthBytes > 0) {
            bytes memory lengthBytes = new bytes(nLengthBytes);
            for (uint256 i = 0; i < nLengthBytes; i++) {
                lengthBytes[i] = _bytecode[3 + i];
            }
            dataLength = bytesToUint(lengthBytes);
        }

        bytes memory preambleData = new bytes(dataLength);
        if (nLengthBytes != 0) {
            uint256 dataStart = 3 + nLengthBytes;
            // This loop is very small, so not worth using the identity precompile like we do with initcode below.
            for (uint256 i = 0; i < dataLength; i++) {
                preambleData[i] = _bytecode[dataStart + i];
            }
        }

        // Parsing the initcode byte-by-byte is too costly for long initcode, so we perform a staticcall
        // to the identity precompile at address(0x04) to copy the initcode.
        uint256 initcodeStart = 3 + nLengthBytes + dataLength;
        uint256 initcodeLength = _bytecode.length - initcodeStart;
        if (initcodeLength == 0) revert EmptyInitcode();

        bytes memory initcode = new bytes(initcodeLength);
        bool success;
        assembly ("memory-safe") {
            // Calculate the memory address of the input data (initcode) within _bytecode.
            // - add(_bytecode, 32): Moves past the length field to the start of _bytecode's data.
            // - add(..., initcodeStart): Adds the offset to reach the initcode within _bytecode.
            let inputData := add(add(_bytecode, 32), initcodeStart)

            // Calculate the memory address for the output data in initcode.
            let outputData := add(initcode, 32)

            // Perform the staticcall to the identity precompile.
            success := staticcall(gas(), 0x04, inputData, initcodeLength, outputData, initcodeLength)
        }

        if (!success) revert IdentityPrecompileCallFailed();
        return Preamble(ercVersion, preambleData, initcode);
    }

    /// @notice Parses the code at the given `_target` as a blueprint and deploys the resulting initcode.
    /// This version of `deployFrom` is used when the initcode requires no constructor arguments.
    function deployFrom(address _target, bytes32 _salt) internal returns (address) {
        return deployFrom(_target, _salt, new bytes(0));
    }

    /// @notice Parses the code at the given `_target` as a blueprint and deploys the resulting initcode
    /// with the given `_data` appended, i.e. `_data` is the ABI-encoded constructor arguments.
    function deployFrom(address _target, bytes32 _salt, bytes memory _data) internal returns (address newContract_) {
        Preamble memory preamble = parseBlueprintPreamble(address(_target).code);
        if (preamble.ercVersion != 0) revert UnsupportedERCVersion(preamble.ercVersion);
        if (preamble.preambleData.length != 0) revert UnexpectedPreambleData(preamble.preambleData);

        bytes memory initcode = bytes.concat(preamble.initcode, _data);
        assembly ("memory-safe") {
            newContract_ := create2(0, add(initcode, 0x20), mload(initcode), _salt)
        }
        if (newContract_ == address(0)) revert DeploymentFailed();
    }

    /// @notice Parses the code at two target addresses as individual blueprints, concatentates them and then deploys
    /// the resulting initcode with the given `_data` appended, i.e. `_data` is the ABI-encoded constructor arguments.
    function deployFrom(
        address _target1,
        address _target2,
        bytes32 _salt,
        bytes memory _data
    )
        internal
        returns (address newContract_)
    {
        Preamble memory preamble1 = parseBlueprintPreamble(address(_target1).code);
        if (preamble1.ercVersion != 0) revert UnsupportedERCVersion(preamble1.ercVersion);
        if (preamble1.preambleData.length != 0) revert UnexpectedPreambleData(preamble1.preambleData);

        Preamble memory preamble2 = parseBlueprintPreamble(address(_target2).code);
        if (preamble2.ercVersion != 0) revert UnsupportedERCVersion(preamble2.ercVersion);
        if (preamble2.preambleData.length != 0) revert UnexpectedPreambleData(preamble2.preambleData);

        bytes memory initcode = bytes.concat(preamble1.initcode, preamble2.initcode, _data);
        assembly ("memory-safe") {
            newContract_ := create2(0, add(initcode, 0x20), mload(initcode), _salt)
        }
        if (newContract_ == address(0)) revert DeploymentFailed();
    }

    /// @notice Deploys a blueprint contract with the given `_rawBytecode` and `_salt`. If the blueprint is too large to
    /// fit in a single deployment, it is split across two addresses. It is the responsibility of the caller to handle
    /// large contracts by checking if the second return value is not address(0).
    function create(
        bytes memory _rawBytecode,
        bytes32 _salt
    )
        internal
        returns (address newContract1_, address newContract2_)
    {
        if (_rawBytecode.length <= maxInitCodeSize()) {
            newContract1_ = deploySmallBytecode(blueprintDeployerBytecode(_rawBytecode), _salt);
            return (newContract1_, address(0));
        }

        (newContract1_, newContract2_) = deployBigBytecode(_rawBytecode, _salt);
    }

    /// @notice Deploys a blueprint contract that can fit in a single address.
    function deploySmallBytecode(bytes memory _bytecode, bytes32 _salt) internal returns (address newContract_) {
        assembly ("memory-safe") {
            newContract_ := create2(0, add(_bytecode, 0x20), mload(_bytecode), _salt)
        }
        require(newContract_ != address(0), "Blueprint: create2 failed");
    }

    /// @notice Deploys a two blueprint contracts, splitting the bytecode across both of them.
    function deployBigBytecode(
        bytes memory _bytecode,
        bytes32 _salt
    )
        internal
        returns (address newContract1_, address newContract2_)
    {
        uint32 maxSize = maxInitCodeSize();
        bytes memory part1Slice = Bytes.slice(_bytecode, 0, maxSize);
        bytes memory part1 = blueprintDeployerBytecode(part1Slice);
        bytes memory part2Slice = Bytes.slice(_bytecode, maxSize, _bytecode.length - maxSize);
        bytes memory part2 = blueprintDeployerBytecode(part2Slice);

        newContract1_ = deploySmallBytecode(part1, _salt);
        newContract2_ = deploySmallBytecode(part2, _salt);
    }

    /// @notice Convert a bytes array to a uint256.
    function bytesToUint(bytes memory _b) internal pure returns (uint256) {
        if (_b.length > 32) revert BytesArrayTooLong();
        uint256 number;
        for (uint256 i = 0; i < _b.length; i++) {
            number = number + uint256(uint8(_b[i])) * (2 ** (8 * (_b.length - (i + 1))));
        }
        return number;
    }

    /// @notice Returns the maximum init code size for each blueprint. The preamble needs 3 bytes.
    function maxInitCodeSize() internal pure returns (uint32) {
        // Technically this should be 24576 - 3 (max minus preamble size) but we use 23500 here to
        // resolve a discrepancy that can occur between development and production builds.
        return 23500;
    }
}
