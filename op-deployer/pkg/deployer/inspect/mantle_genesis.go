package inspect

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
)

func MantleGenesisAndRollup(globalState *state.State, chainID common.Hash) (*core.Genesis, *rollup.Config, error) {
	if globalState.AppliedIntent == nil {
		return nil, nil, fmt.Errorf("chain state is not applied - run op-deployer apply")
	}

	chainIntent, err := globalState.AppliedIntent.Chain(chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get applied chain intent: %w", err)
	}

	chainState, err := globalState.Chain(chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chain ID %s: %w", chainID.String(), err)
	}

	l2Allocs := chainState.Allocs.Data
	config, err := state.CombineMantleDeployConfig(
		globalState.AppliedIntent,
		chainIntent,
		globalState,
		chainState,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to combine L2 init config: %w", err)
	}

	// Normally we should use params.GetUpgradeConfigForMantle(new(big.Int).SetUint64(config.L2ChainID)) to get the upgrade config,
	// but that upgrade config is hard coded in geth repo. In order to make an in-memory env configurable, we use nil here.
	l2GenesisBuilt, err := genesis.BuildMantleGenesis(&config, l2Allocs, chainState.StartBlock.ToBlockRef(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build L2 genesis: %w", err)
	}
	l2GenesisBlock := l2GenesisBuilt.ToBlock()

	// the same for rollup config, we use nil for the mantle upgrade config.
	rollupConfig, err := config.MantleRollupConfig(
		chainState.StartBlock.ToBlockRef(),
		l2GenesisBlock.Hash(),
		l2GenesisBlock.Number().Uint64(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build rollup config: %w", err)
	}

	if err := rollupConfig.Check(); err != nil {
		return nil, nil, fmt.Errorf("generated rollup config does not pass validation: %w", err)
	}

	return l2GenesisBuilt, rollupConfig, nil
}
