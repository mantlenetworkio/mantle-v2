package inspect

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/pipeline"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"

	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func L1CLI(cliCtx *cli.Context) error {
	cfg, err := readConfig(cliCtx)
	if err != nil {
		return err
	}

	globalState, err := pipeline.ReadState(cfg.Workdir)
	if err != nil {
		return fmt.Errorf("failed to read intent: %w", err)
	}

	l1Contracts, err := L1(globalState, cfg.ChainID)
	if err != nil {
		return fmt.Errorf("failed to generate l1Contracts: %w", err)
	}

	if err := jsonutil.WriteJSON(l1Contracts, ioutil.ToStdOutOrFileOrNoop(cfg.Outfile, 0o666)); err != nil {
		return fmt.Errorf("failed to write L1 contract addresses: %w", err)
	}

	return nil
}

func L1(globalState *state.State, chainID common.Hash) (*addresses.L1Contracts, error) {
	chainState, err := globalState.Chain(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain state for ID %s: %w", chainID.String(), err)
	}

	l1Contracts := addresses.L1Contracts{
		SuperchainContracts:      *globalState.SuperchainDeployment,
		ImplementationsContracts: *globalState.ImplementationsDeployment,
		OpChainContracts:         chainState.OpChainContracts,
	}

	return &l1Contracts, nil
}

func L1ForMantle(globalState *state.State, chainID common.Hash) (*addresses.L1Contracts, error) {
	chainState, err := globalState.Chain(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain state for ID %s: %w", chainID.String(), err)
	}

	l1Contracts := addresses.L1Contracts{
		SuperchainContracts:      *globalState.SuperchainDeployment,
		ImplementationsContracts: *globalState.ImplementationsDeployment,
		OpChainContracts:         chainState.OpChainContracts,
	}
	// Mantle special handling: If the implementation contracts in ImplementationsContracts are empty
	//then they are filled from MantleImplContracts (Mantle stores the implementation contracts in MantleImplContracts

	if chainState.MantleImplContracts.AllSetUp ||
		chainState.MantleImplContracts.L1CrossDomainMessengerImpl != (common.Address{}) {

		//Fill in the missing Mantle implementation contract in ImplementationsContracts
		if l1Contracts.L1CrossDomainMessengerImpl == (common.Address{}) {
			l1Contracts.L1CrossDomainMessengerImpl = chainState.MantleImplContracts.L1CrossDomainMessengerImpl
		}
		if l1Contracts.L1StandardBridgeImpl == (common.Address{}) {
			l1Contracts.L1StandardBridgeImpl = chainState.MantleImplContracts.L1StandardBridgeImpl
		}
		if l1Contracts.L1Erc721BridgeImpl == (common.Address{}) {
			l1Contracts.L1Erc721BridgeImpl = chainState.MantleImplContracts.L1Erc721BridgeImpl
		}
		if l1Contracts.OptimismMintableErc20FactoryImpl == (common.Address{}) {
			l1Contracts.OptimismMintableErc20FactoryImpl = chainState.MantleImplContracts.OptimismMintableErc20FactoryImpl
		}
		if l1Contracts.SystemConfigImpl == (common.Address{}) {
			l1Contracts.SystemConfigImpl = chainState.MantleImplContracts.SystemConfigImpl
		}
		if l1Contracts.OptimismPortalImpl == (common.Address{}) {
			l1Contracts.OptimismPortalImpl = chainState.MantleImplContracts.OptimismPortalImpl
		}
		if l1Contracts.L2OutputOracleImpl == (common.Address{}) {
			l1Contracts.L2OutputOracleImpl = chainState.MantleImplContracts.L2OutputOracleImpl
		}
	}

	return &l1Contracts, nil
}
