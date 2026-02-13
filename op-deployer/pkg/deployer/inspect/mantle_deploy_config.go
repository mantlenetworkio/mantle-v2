package inspect

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum/go-ethereum/common"
)

func MantleDeployConfig(globalState *state.State, chainID common.Hash) (*genesis.DeployConfig, error) {
	chainState, err := globalState.Chain(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chain state: %w", err)
	}

	intent := globalState.AppliedIntent
	if intent == nil {
		return nil, fmt.Errorf("can only run this command following a full apply")
	}
	chainIntent, err := intent.Chain(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chain intent: %w", err)
	}

	config, err := state.CombineMantleDeployConfig(intent, chainIntent, globalState, chainState)
	if err != nil {
		return nil, fmt.Errorf("failed to generate deploy config: %w", err)
	}

	return &config, nil
}
