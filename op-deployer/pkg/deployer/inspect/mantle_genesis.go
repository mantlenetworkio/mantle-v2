package inspect

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
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
	config, err := state.CombineDeployConfig(
		globalState.AppliedIntent,
		chainIntent,
		globalState,
		chainState,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to combine L2 init config: %w", err)
	}

	l2GenesisBuilt, err := genesis.BuildL2Genesis(&config, l2Allocs, chainState.StartBlock.ToBlockRef())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build L2 genesis: %w", err)
	}
	l2GenesisBlock := l2GenesisBuilt.ToBlock()

	rollupConfig, err := config.RollupConfig(
		chainState.StartBlock.ToBlockRef(),
		l2GenesisBlock.Hash(),
		l2GenesisBlock.Number().Uint64(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build rollup config: %w", err)
	}

	// Mantle features
	// Apply Mantle overrides to the rollup config
	if err := rollupConfig.ApplyMantleOverrides(params.GetUpgradeConfigForMantle(new(big.Int).SetUint64(config.L2ChainID))); err != nil {
		return nil, nil, fmt.Errorf("failed to apply mantle overrides: %w", err)
	}
	// setup initial 1559 params in rollup system config
	if config.L2GenesisMantleArsiaTimeOffset == nil {
		rollupConfig.Genesis.SystemConfig.MarshalPreHolocene = true
	}
	if config.L2GenesisMantleArsiaTimeOffset != nil && *config.L2GenesisMantleArsiaTimeOffset == 0 {
		rollupConfig.Genesis.SystemConfig.EIP1559Params = eth.Bytes8(eip1559.EncodeHolocene1559Params(config.EIP1559Denominator, config.EIP1559Elasticity))
	}

	if err := rollupConfig.Check(); err != nil {
		return nil, nil, fmt.Errorf("generated rollup config does not pass validation: %w", err)
	}

	return l2GenesisBuilt, rollupConfig, nil
}
