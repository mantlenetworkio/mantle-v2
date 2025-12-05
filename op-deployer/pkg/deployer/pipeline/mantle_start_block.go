package pipeline

import (
	"errors"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum/go-ethereum/common"
)

func SetMantleStartBlockGenesisStrategy(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "set-start-block", "strategy", "genesis")
	lgr.Info("setting start block", "id", chainID.Hex())

	if st.L1DevGenesis == nil {
		return errors.New("must seal L1 genesis state before determining start-block")
	}
	thisChainState, err := st.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	// Mantle geth don't use genesis.stateHash. It's just a compatible field for optimism op-node.
	// So we need to manually set the state dump to get the start block hash.
	st.L1DevGenesis.Alloc = st.L1StateDump.Data.Accounts

	thisChainState.StartBlock = state.BlockRefJsonFromHeader(st.L1DevGenesis.ToBlock().Header())

	// Reset the state dump to avoid petential test failures
	st.L1DevGenesis.Alloc = nil

	return nil
}
