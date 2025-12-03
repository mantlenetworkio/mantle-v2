package pipeline

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path"
	"runtime"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/broadcaster"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/env"
	op_service "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func GenerateMantleL2Genesis(pEnv *Env, intent *state.Intent, bundle ArtifactsBundle, st *state.State, chainID common.Hash) error {
	lgr := pEnv.Logger.New("stage", "generate-mantle-l2-genesis")

	lgr.Info("generating Mantle L2 genesis", "id", chainID.Hex())

	thisIntent, err := intent.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain intent: %w", err)
	}

	thisChainState, err := st.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	if !shouldGenerateL2Genesis(thisChainState) {
		lgr.Info("L2 genesis generation not needed")
		return nil
	}

	lgr.Info("generating Mantle L2 genesis", "id", chainID.Hex())

	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		pEnv.Logger,
		pEnv.Deployer,
		bundle.L2,
	)
	if err != nil {
		return fmt.Errorf("failed to create L2 script host: %w", err)
	}

	script, err := opcm.NewMantleL2GenesisScript(host)
	if err != nil {
		return fmt.Errorf("failed to create L2Genesis script: %w", err)
	}

	overrides, schedule, err := calculateL2GenesisOverrides(intent, thisIntent)
	if err != nil {
		return fmt.Errorf("failed to calculate L2 genesis overrides: %w", err)
	}

	if err := script.Run(opcm.MantleL2GenesisInput{
		L1ChainID:                   new(big.Int).SetUint64(intent.L1ChainID),
		L2ChainID:                   chainID.Big(),
		L1CrossDomainMessengerProxy: thisChainState.L1CrossDomainMessengerProxy,
		L1StandardBridgeProxy:       thisChainState.L1StandardBridgeProxy,
		L1ERC721BridgeProxy:         thisChainState.L1Erc721BridgeProxy,
		L1MNTAddress:                thisIntent.L1MNT,
		OpChainProxyAdminOwner:      thisIntent.Roles.L2ProxyAdminOwner,
		BaseFeeVaultRecipient:       thisIntent.BaseFeeVaultRecipient,
		L1FeeVaultRecipient:         thisIntent.L1FeeVaultRecipient,
		SequencerFeeVaultRecipient:  thisIntent.SequencerFeeVaultRecipient,
		OperatorFeeVaultRecipient:   thisIntent.OperatorFeeVaultRecipient,
		GasPriceOracleOwner:         thisIntent.GasPriceOracleOwner,
		MantleFork:                  big.NewInt(schedule.SolidityMantleForkNumber(1)),
		FundDevAccounts:             overrides.FundDevAccounts,
	}); err != nil {
		return fmt.Errorf("failed to call L2Genesis script: %w", err)
	}

	host.Wipe(pEnv.Deployer)

	dump, err := host.StateDump()
	if err != nil {
		return fmt.Errorf("failed to dump state: %w", err)
	}

	thisChainState.Allocs = &state.GzipData[foundry.ForgeAllocs]{
		Data: dump,
	}
	return nil
}

func DefaultMantleL2GenesisStates(logger log.Logger, deployer common.Address, config *genesis.DeployConfig) (*foundry.ForgeAllocs, error) {
	lgr := logger.New("stage", "generate-l2-genesis")

	lgr.Info("generating L2 genesis", "id", config.L2ChainID)

	// Get default forge artifacts
	artifactsFS, err := defaultForgeArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to get default forge artifacts: %w", err)
	}

	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		logger,
		deployer,
		artifactsFS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create L2 script host: %w", err)
	}

	// Create L2Genesis script
	script, err := opcm.NewMantleL2GenesisScript(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create L2Genesis script: %w", err)
	}

	if err := script.Run(opcm.MantleL2GenesisInput{
		L1ChainID:                   new(big.Int).SetUint64(config.L1ChainID),
		L2ChainID:                   new(big.Int).SetUint64(config.L2ChainID),
		L1CrossDomainMessengerProxy: config.L1CrossDomainMessengerProxy,
		L1StandardBridgeProxy:       config.L1StandardBridgeProxy,
		L1ERC721BridgeProxy:         config.L1ERC721BridgeProxy,
		L1MNTAddress:                config.L1MantleToken,
		OpChainProxyAdminOwner:      config.ProxyAdminOwner,
		BaseFeeVaultRecipient:       config.BaseFeeVaultRecipient,
		L1FeeVaultRecipient:         config.L1FeeVaultRecipient,
		SequencerFeeVaultRecipient:  config.SequencerFeeVaultRecipient,
		OperatorFeeVaultRecipient:   config.OperatorFeeVaultRecipient,
		GasPriceOracleOwner:         config.GasPriceOracleOwner,
		MantleFork:                  big.NewInt(config.SolidityMantleForkNumber(1)),
		FundDevAccounts:             config.FundDevAccounts,
	}); err != nil {
		return nil, fmt.Errorf("failed to call L2Genesis script: %w", err)
	}

	host.Wipe(deployer)

	dump, err := host.StateDump()
	if err != nil {
		return nil, fmt.Errorf("failed to dump state: %w", err)
	}

	return dump, nil
}

func defaultForgeArtifacts() (foundry.StatDirFs, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get caller filename")
	}
	monorepoDir, err := op_service.FindMonorepoRoot(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to find monorepo root: %w", err)
	}
	artifactsDir := path.Join(monorepoDir, "packages", "contracts-bedrock", "forge-artifacts")
	artifactsURL, err := url.Parse(fmt.Sprintf("file://%s", artifactsDir))
	if err != nil {
		return nil, fmt.Errorf("failed to parse artifacts URL: %w", err)
	}
	loc := &artifacts.Locator{
		URL: artifactsURL,
	}
	cacheDir := path.Join(monorepoDir, "packages", "contracts-bedrock", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	artifactsFS, err := artifacts.Download(context.Background(), loc, ioutil.NoopProgressor(), cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to download artifacts: %w", err)
	}
	return artifactsFS, nil
}
