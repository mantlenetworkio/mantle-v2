// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

import { Script } from "forge-std/Script.sol";
import { console2 as console } from "forge-std/console2.sol";
import { stdJson } from "forge-std/StdJson.sol";
import { Process } from "scripts/libraries/Process.sol";
import { Config, Fork, ForkUtils } from "scripts/libraries/Config.sol";

/// @title DeployConfig
/// @notice Represents the configuration required to deploy the system. It is expected
///         to read the file from JSON. A future improvement would be to have fallback
///         values if they are not defined in the JSON themselves.
contract DeployConfig is Script {
    using stdJson for string;
    using ForkUtils for Fork;

    /// @notice Represents an unset offset value, as opposed to 0, which denotes no-offset.
    uint256 constant NULL_OFFSET = type(uint256).max;

    string internal _json;

    address public finalSystemOwner;
    address public portalGuardian;
    address public controller;
    uint256 public l1ChainID;
    uint256 public l2ChainID;
    uint256 public l2BlockTime;
    uint256 public maxSequencerDrift;
    uint256 public sequencerWindowSize;
    uint256 public channelTimeout;
    address public p2pSequencerAddress;
    address public batchInboxAddress;
    address public batchSenderAddress;
    uint256 public l2OutputOracleSubmissionInterval;
    int256 internal _l2OutputOracleStartingTimestamp;
    uint256 public l2OutputOracleStartingBlockNumber;
    address public l2OutputOracleProposer;
    address public l2OutputOracleChallenger;
    uint256 public l2GenesisBlockGasLimit;
    uint256 public l1BlockTime;
    address public l1MantleToken;
    address public cliqueSignerAddress;
    address public baseFeeVaultRecipient;
    address public l1FeeVaultRecipient;
    address public sequencerFeeVaultRecipient;
    address public proxyAdminOwner;
    uint256 public finalizationPeriodSeconds;
    uint256 public numDeployConfirmations;
    bool public fundDevAccounts;
    uint256 public l2GenesisBlockBaseFeePerGas;
    address public gasPriceOracleOwner;
    uint256 public gasPriceOracleTokenRatio;
    uint256 public gasPriceOracleOverhead;
    uint256 public gasPriceOracleScalar;
    string public governanceTokenSymbol;
    string public governanceTokenName;
    address public governanceTokenOwner;
    uint256 public eip1559Denominator;
    uint256 public eip1559Elasticity;
    uint256 public l1GenesisBlockTimestamp;
    // string public l1StartingBlockTag;
    uint256 public l2GenesisRegolithTimeOffset;

    function read(string memory _path) public {
        console.log("DeployConfig: reading file %s", _path);
        try vm.readFile(_path) returns (string memory data_) {
            _json = data_;
        } catch {
            require(false, string.concat("DeployConfig: cannot find deploy config file at ", _path));
        }

        finalSystemOwner = stdJson.readAddress(_json, "$.finalSystemOwner");
        portalGuardian = stdJson.readAddress(_json, "$.portalGuardian");
        controller = stdJson.readAddress(_json, "$.controller");
        l1ChainID = stdJson.readUint(_json, "$.l1ChainID");
        l2ChainID = stdJson.readUint(_json, "$.l2ChainID");
        l2BlockTime = stdJson.readUint(_json, "$.l2BlockTime");
        maxSequencerDrift = stdJson.readUint(_json, "$.maxSequencerDrift");
        sequencerWindowSize = stdJson.readUint(_json, "$.sequencerWindowSize");
        channelTimeout = stdJson.readUint(_json, "$.channelTimeout");
        p2pSequencerAddress = stdJson.readAddress(_json, "$.p2pSequencerAddress");
        batchInboxAddress = stdJson.readAddress(_json, "$.batchInboxAddress");
        batchSenderAddress = stdJson.readAddress(_json, "$.batchSenderAddress");
        l2OutputOracleSubmissionInterval = stdJson.readUint(_json, "$.l2OutputOracleSubmissionInterval");
        _l2OutputOracleStartingTimestamp = stdJson.readInt(_json, "$.l2OutputOracleStartingTimestamp");
        l2OutputOracleStartingBlockNumber = stdJson.readUint(_json, "$.l2OutputOracleStartingBlockNumber");
        l2OutputOracleProposer = stdJson.readAddress(_json, "$.l2OutputOracleProposer");
        l2OutputOracleChallenger = stdJson.readAddress(_json, "$.l2OutputOracleChallenger");
        l2GenesisBlockGasLimit = stdJson.readUint(_json, "$.l2GenesisBlockGasLimit");
        l1BlockTime = stdJson.readUint(_json, "$.l1BlockTime");
        l1MantleToken = stdJson.readAddress(_json, "$.l1MantleToken");
        cliqueSignerAddress = stdJson.readAddress(_json, "$.cliqueSignerAddress");
        baseFeeVaultRecipient = stdJson.readAddress(_json, "$.baseFeeVaultRecipient");
        l1FeeVaultRecipient = stdJson.readAddress(_json, "$.l1FeeVaultRecipient");
        sequencerFeeVaultRecipient = stdJson.readAddress(_json, "$.sequencerFeeVaultRecipient");
        proxyAdminOwner = stdJson.readAddress(_json, "$.proxyAdminOwner");
        finalizationPeriodSeconds = stdJson.readUint(_json, "$.finalizationPeriodSeconds");
        numDeployConfirmations = stdJson.readUint(_json, "$.numDeployConfirmations");
        fundDevAccounts = _readOr(_json, "$.fundDevAccounts", false);
        l2GenesisBlockBaseFeePerGas = stdJson.readUint(_json, "$.l2GenesisBlockBaseFeePerGas");
        gasPriceOracleOwner = stdJson.readAddress(_json, "$.gasPriceOracleOwner");
        gasPriceOracleTokenRatio = stdJson.readUint(_json, "$.gasPriceOracleTokenRatio");
        gasPriceOracleOverhead = stdJson.readUint(_json, "$.gasPriceOracleOverhead");
        gasPriceOracleScalar = stdJson.readUint(_json, "$.gasPriceOracleScalar");
        governanceTokenSymbol = stdJson.readString(_json, "$.governanceTokenSymbol");
        governanceTokenName = stdJson.readString(_json, "$.governanceTokenName");
        governanceTokenOwner = stdJson.readAddress(_json, "$.governanceTokenOwner");
        eip1559Denominator = stdJson.readUint(_json, "$.eip1559Denominator");
        eip1559Elasticity = stdJson.readUint(_json, "$.eip1559Elasticity");
        l1GenesisBlockTimestamp = stdJson.readUint(_json, "$.l1GenesisBlockTimestamp");
        // l1StartingBlockTag = stdJson.readString(_json, "$.l1StartingBlockTag");
        l2GenesisRegolithTimeOffset = stdJson.readUint(_json, "$.l2GenesisRegolithTimeOffset");
    }

    // function fork() public view returns (Fork fork_) {
    //     // let env var take precedence
    //     fork_ = Config.fork();
    //     if (fork_ == Fork.NONE) {
    //         // Will revert if no deploy config can be found either.
    //         fork_ = latestGenesisFork();
    //         console.log("DeployConfig: using deploy config fork: %s", fork_.toString());
    //     } else {
    //         console.log("DeployConfig: using env var fork: %s", fork_.toString());
    //     }
    // }

    function l1StartingBlockTag() public returns (bytes32) {
        try vm.parseJsonBytes32(_json, "$.l1StartingBlockTag") returns (bytes32 tag_) {
            return tag_;
        } catch {
            try vm.parseJsonString(_json, "$.l1StartingBlockTag") returns (string memory tag_) {
                return _getBlockByTag(tag_);
            } catch {
                try vm.parseJsonUint(_json, "$.l1StartingBlockTag") returns (uint256 tag_) {
                    return _getBlockByTag(vm.toString(tag_));
                } catch { }
            }
        }
        revert(
            "DeployConfig: l1StartingBlockTag must be a bytes32, string or uint256 or cannot fetch l1StartingBlockTag"
        );
    }

    function l2OutputOracleStartingTimestamp() public returns (uint256) {
        if (_l2OutputOracleStartingTimestamp < 0) {
            bytes32 tag = l1StartingBlockTag();
            string memory cmd = string.concat("cast block ", vm.toString(tag), " --json | jq .timestamp");
            string memory res = Process.bash(cmd);
            return stdJson.readUint(res, "");
        }
        return uint256(_l2OutputOracleStartingTimestamp);
    }

    /// @notice Allow the `fundDevAccounts` config to be overridden.
    function setFundDevAccounts(bool _fundDevAccounts) public {
        fundDevAccounts = _fundDevAccounts;
    }

    // function latestGenesisFork() internal view returns (Fork) {
    //     if (l2GenesisHoloceneTimeOffset == 0) {
    //         return Fork.HOLOCENE;
    //     } else if (l2GenesisGraniteTimeOffset == 0) {
    //         return Fork.GRANITE;
    //     } else if (l2GenesisFjordTimeOffset == 0) {
    //         return Fork.FJORD;
    //     } else if (l2GenesisEcotoneTimeOffset == 0) {
    //         return Fork.ECOTONE;
    //     } else if (l2GenesisDeltaTimeOffset == 0) {
    //         return Fork.DELTA;
    //     }
    //     revert("DeployConfig: no supported fork active at genesis");
    // }

    function _getBlockByTag(string memory _tag) internal returns (bytes32) {
        string memory cmd = string.concat("cast block ", _tag, " --json | jq -r .hash");
        bytes memory res = bytes(Process.bash(cmd));
        return abi.decode(res, (bytes32));
    }

    function _readOr(string memory _jsonInp, string memory _key, bool _defaultValue) internal view returns (bool) {
        return _jsonInp.readBoolOr(_key, _defaultValue);
    }

    function _readOr(
        string memory _jsonInp,
        string memory _key,
        uint256 _defaultValue
    )
        internal
        view
        returns (uint256)
    {
        return (vm.keyExistsJson(_jsonInp, _key) && !_isNull(_json, _key)) ? _jsonInp.readUint(_key) : _defaultValue;
    }

    function _readOr(
        string memory _jsonInp,
        string memory _key,
        address _defaultValue
    )
        internal
        view
        returns (address)
    {
        return _jsonInp.readAddressOr(_key, _defaultValue);
    }

    function _isNull(string memory _jsonInp, string memory _key) internal pure returns (bool) {
        string memory value = _jsonInp.readString(_key);
        return (keccak256(bytes(value)) == keccak256(bytes("null")));
    }

    function _readOr(
        string memory _jsonInp,
        string memory _key,
        string memory _defaultValue
    )
        internal
        view
        returns (string memory)
    {
        return _jsonInp.readStringOr(_key, _defaultValue);
    }
}
