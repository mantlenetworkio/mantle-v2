package opcm

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-chain-ops/script"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// MantleDeployProxiesInput matches the Input struct from DeployProxies.s.sol
type MantleDeployProxiesInput struct{}

// MantleDeployProxiesOutput matches the Output struct from DeployProxies.s.sol
type MantleDeployProxiesOutput struct {
	AddressManager                    common.Address
	ProxyAdmin                        common.Address
	L1StandardBridgeProxy             common.Address
	L2OutputOracleProxy               common.Address
	L1CrossDomainMessengerProxy       common.Address
	OptimismPortalProxy               common.Address
	OptimismMintableERC20FactoryProxy common.Address
	L1ERC721BridgeProxy               common.Address
	SystemConfigProxy                 common.Address
}

type MantleDeployProxiesScript script.ForgeScript

func NewMantleDeployProxiesScript(host *script.Host) (MantleDeployProxiesScript, error) {
	return script.NewForgeScriptFromFile(host, "DeployProxies.s.sol", "DeployProxies")
}

func RunMantleDeployProxiesScript(s MantleDeployProxiesScript) (MantleDeployProxiesOutput, error) {
	packed, err := s.ABI().Pack("run")
	if err != nil {
		return MantleDeployProxiesOutput{}, err
	}
	result, err := s.Call(packed)
	if err != nil {
		return MantleDeployProxiesOutput{}, err
	}

	// We then decode the raw output to an anonymous struct
	unpacked, err := s.ABI().Unpack("run", result)
	if err != nil {
		return MantleDeployProxiesOutput{}, err
	}

	// And finally we convert the anonymous struct into our typed output
	return *abi.ConvertType(unpacked[0], new(MantleDeployProxiesOutput)).(*MantleDeployProxiesOutput), nil
}

type ResourceConfig struct {
	MaxResourceLimit            uint32
	ElasticityMultiplier        uint8
	BaseFeeMaxChangeDenominator uint8
	MinimumBaseFee              uint32
	SystemTxMaxGas              uint32
	MaximumBaseFee              *big.Int
}

func DecodeResourceConfig() ResourceConfig {
	return ResourceConfig{
		MaxResourceLimit:            genesis.MaxResourceLimit,
		ElasticityMultiplier:        genesis.ElasticityMultiplier,
		BaseFeeMaxChangeDenominator: genesis.BaseFeeMaxChangeDenominator,
		MinimumBaseFee:              genesis.MinimumBaseFee,
		SystemTxMaxGas:              genesis.SystemTxMaxGas,
		MaximumBaseFee:              genesis.MaximumBaseFee,
	}
}

// MantleDeployImplementationsInput matches the Input struct from DeployImplementations.s.sol
type MantleDeployImplementationsInput struct {
	SystemConfigOwner                       common.Address
	SystemConfigBatcherHash                 common.Hash
	SystemConfigGasLimit                    uint64
	SystemConfigBaseFee                     *big.Int
	SystemConfigUnsafeBlockSigner           common.Address
	SystemConfigConfig                      ResourceConfig
	SystemConfigBasefeeScalar               uint32
	SystemConfigBlobbasefeeScalar           uint32
	OptimismPortal                          common.Address
	L1MNT                                   common.Address
	L1CrossDomainMessenger                  common.Address
	L2OutputOracle                          common.Address
	SystemConfig                            common.Address
	L1StandardBridge                        common.Address
	L1ERC721BridgeOtherBridge               common.Address
	L2OutputOracleSubmissionInterval        *big.Int
	L2OutputOracleL2BlockTime               *big.Int
	L2OutputOracleStartingBlockNumber       *big.Int
	L2OutputOracleStartingTimestamp         *big.Int
	L2OutputOracleProposer                  common.Address
	L2OutputOracleChallenger                common.Address
	L2OutputOracleFinalizationPeriodSeconds *big.Int
	OptimismPortalGuardian                  common.Address
	OptimismPortalPaused                    bool
}

// MantleDeployImplementationsOutput matches the Output struct from DeployImplementations.s.sol
type MantleDeployImplementationsOutput struct {
	OptimismPortalImpl               common.Address
	SystemConfigImpl                 common.Address
	L1CrossDomainMessengerImpl       common.Address
	L1ERC721BridgeImpl               common.Address
	L1StandardBridgeImpl             common.Address
	OptimismMintableERC20FactoryImpl common.Address
	L2OutputOracleImpl               common.Address
}

type MantleDeployImplementationsScript script.DeployScriptWithOutput[MantleDeployImplementationsInput, MantleDeployImplementationsOutput]

func NewMantleDeployImplementationsScript(host *script.Host) (MantleDeployImplementationsScript, error) {
	return script.NewDeployScriptWithOutputFromFile[MantleDeployImplementationsInput, MantleDeployImplementationsOutput](host, "DeployImplementations.s.sol", "DeployImplementations")
}

// MantleDeployOPChainInput matches the Input struct from DeployOPChain.s.sol
type MantleDeployOPChainInput struct {
	ProxyAdmin                        common.Address
	OptimismPortalImpl                common.Address
	OptimismPortalProxy               common.Address
	SystemConfigImpl                  common.Address
	SystemConfigProxy                 common.Address
	L1CrossDomainMessengerImpl        common.Address
	L1CrossDomainMessengerProxy       common.Address
	L1ERC721BridgeImpl                common.Address
	L1ERC721BridgeProxy               common.Address
	L1StandardBridgeImpl              common.Address
	L1StandardBridgeProxy             common.Address
	OptimismMintableERC20FactoryImpl  common.Address
	OptimismMintableERC20FactoryProxy common.Address
	L2OutputOracleImpl                common.Address
	L2OutputOracleProxy               common.Address
	FinalSystemOwner                  common.Address
	BasefeeScalar                     uint32
	BlobbasefeeScalar                 uint32
	BatchSenderAddress                common.Address
	L2GenesisBlockGasLimit            uint64
	L2GenesisBlockBaseFeePerGas       *big.Int
	P2pSequencerAddress               common.Address
	L2OutputOracleStartingBlockNumber *big.Int
	L2OutputOracleStartingTimestamp   *big.Int
}

type MantleDeployOPChainScript script.DeployScriptWithoutOutput[MantleDeployOPChainInput]

func NewMantleDeployOPChainScript(host *script.Host) (MantleDeployOPChainScript, error) {
	return script.NewDeployScriptWithoutOutputFromFile[MantleDeployOPChainInput](host, "DeployOPChain.s.sol", "DeployOPChain")
}

type MantleL2GenesisInput struct {
	L1ChainID                   *big.Int
	L2ChainID                   *big.Int
	L1CrossDomainMessengerProxy common.Address
	L1StandardBridgeProxy       common.Address
	L1ERC721BridgeProxy         common.Address
	L1MNTAddress                common.Address
	OpChainProxyAdminOwner      common.Address
	SequencerFeeVaultRecipient  common.Address
	BaseFeeVaultRecipient       common.Address
	L1FeeVaultRecipient         common.Address
	OperatorFeeVaultRecipient   common.Address
	GasPriceOracleOwner         common.Address
	MantleFork                  *big.Int
	FundDevAccounts             bool
}

type MantleL2GenesisScript script.DeployScriptWithoutOutput[MantleL2GenesisInput]

func NewMantleL2GenesisScript(host *script.Host) (MantleL2GenesisScript, error) {
	return script.NewDeployScriptWithoutOutputFromFile[MantleL2GenesisInput](host, "L2Genesis.s.sol", "L2Genesis")
}

// MantleScripts contains all the deployment scripts for ease of passing them around
type MantleScripts struct {
	DeployProxies         MantleDeployProxiesScript
	DeployImplementations MantleDeployImplementationsScript
	DeployOPChain         MantleDeployOPChainScript
}

// NewMantleScripts collects all the deployment scripts, raising exceptions if any of them
// are not found or if the Go types don't match the ABI
func NewMantleScripts(host *script.Host) (*MantleScripts, error) {
	deployProxies, err := NewMantleDeployProxiesScript(host)
	if err != nil {
		return nil, fmt.Errorf("failed to load DeployProxies script: %w", err)
	}

	deployImplementations, err := NewMantleDeployImplementationsScript(host)
	if err != nil {
		return nil, fmt.Errorf("failed to load DeployImplementations script: %w", err)
	}

	deployOPChain, err := NewMantleDeployOPChainScript(host)
	if err != nil {
		return nil, fmt.Errorf("failed to load DeployOPChain script: %w", err)
	}

	return &MantleScripts{
		DeployProxies:         deployProxies,
		DeployImplementations: deployImplementations,
		DeployOPChain:         deployOPChain,
	}, nil
}
