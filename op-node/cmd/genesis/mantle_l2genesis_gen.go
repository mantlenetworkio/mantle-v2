package genesis

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path"
	"runtime"

	"github.com/ethereum-optimism/optimism/op-service/ioutil"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-chain-ops/script"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/broadcaster"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/env"
	op_service "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

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
	MantleFork                  *big.Int
	FundDevAccounts             bool
}

type MantleL2GenesisScript script.DeployScriptWithoutOutput[MantleL2GenesisInput]

func NewMantleL2GenesisScript(host *script.Host) (MantleL2GenesisScript, error) {
	return script.NewDeployScriptWithoutOutputFromFile[MantleL2GenesisInput](host, "L2Genesis.s.sol", "L2Genesis")
}

func GenerateL2Genesis(logger log.Logger, deployer common.Address, config *genesis.DeployConfig) (*foundry.ForgeAllocs, error) {
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
	script, err := NewMantleL2GenesisScript(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create L2Genesis script: %w", err)
	}

	if err := script.Run(MantleL2GenesisInput{
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
