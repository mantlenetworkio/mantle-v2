package pipeline

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/standard"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
)

func DeployMantleOPChain(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "deploy-mantle-opchain")

	if !shouldDeployMantleOPChain(st, chainID) {
		lgr.Info("mantle opchain deployment not needed")
		return nil
	}

	thisIntent, err := intent.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain intent: %w", err)
	}

	thisState, err := st.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	lgr.Info("deploying Mantle OP chain", "id", chainID.Hex())

	deployParams, err := jsonutil.MergeJSON(
		MantleDeployParamsSet{
			BasefeeScalar:                           standard.BasefeeScalar,
			BlobbasefeeScalar:                       standard.BlobBaseFeeScalar,
			L2OutputOracleSubmissionInterval:        big.NewInt(SubmissionInterval),
			L2OutputOracleL2BlockTime:               big.NewInt(L2BlockTime),
			L2OutputOracleStartingBlockNumber:       big.NewInt(StartingBlockNumber),
			L2OutputOracleStartingTimestamp:         big.NewInt(StartingTimestamp),
			L2OutputOracleFinalizationPeriodSeconds: big.NewInt(FinalizationPeriodSeconds),
			OptimismPortalGuardian:                  common.HexToAddress(MantlePortalGuardian),
			OptimismPortalPaused:                    Paused,
		},
		intent.GlobalDeployOverrides,
	)
	if err != nil {
		return fmt.Errorf("error merging deploy params from overrides: %w", err)
	}

	dci := opcm.MantleDeployOPChainInput{
		ProxyAdmin:                        thisState.OpChainContracts.OpChainProxyAdminImpl,
		OptimismPortalImpl:                thisState.OptimismPortalImpl,
		OptimismPortalProxy:               thisState.OptimismPortalProxy,
		SystemConfigImpl:                  thisState.SystemConfigImpl,
		SystemConfigProxy:                 thisState.SystemConfigProxy,
		L1CrossDomainMessengerImpl:        thisState.L1CrossDomainMessengerImpl,
		L1CrossDomainMessengerProxy:       thisState.L1CrossDomainMessengerProxy,
		L1ERC721BridgeImpl:                thisState.L1Erc721BridgeImpl,
		L1ERC721BridgeProxy:               thisState.L1Erc721BridgeProxy,
		L1StandardBridgeImpl:              thisState.L1StandardBridgeImpl,
		L1StandardBridgeProxy:             thisState.L1StandardBridgeProxy,
		OptimismMintableERC20FactoryImpl:  thisState.OptimismMintableErc20FactoryImpl,
		OptimismMintableERC20FactoryProxy: thisState.OptimismMintableErc20FactoryProxy,
		L2OutputOracleImpl:                thisState.L2OutputOracleImpl,
		L2OutputOracleProxy:               thisState.L2OutputOracleProxy,
		FinalSystemOwner:                  thisIntent.Roles.L1ProxyAdminOwner,
		BasefeeScalar:                     deployParams.BasefeeScalar,
		BlobbasefeeScalar:                 deployParams.BlobbasefeeScalar,
		BatchSenderAddress:                thisIntent.Roles.Batcher,
		L2GenesisBlockGasLimit:            thisIntent.GasLimit,
		L2GenesisBlockBaseFeePerGas:       big.NewInt(int64(thisIntent.MinBaseFee)),
		P2pSequencerAddress:               thisIntent.Roles.UnsafeBlockSigner,
		L2OutputOracleStartingBlockNumber: deployParams.L2OutputOracleStartingBlockNumber,
		L2OutputOracleStartingTimestamp:   deployParams.L2OutputOracleStartingTimestamp,
	}

	err = env.MantleScripts.DeployOPChain.Run(dci)
	if err != nil {
		return fmt.Errorf("error deploying Mantle OP chain: %w", err)
	}

	thisState.MantleImplContracts.AllSetUp = true

	// New an empty instance of superchain deployment to avoid nil pointer dereference
	st.SuperchainDeployment = &addresses.SuperchainContracts{}

	return nil
}

func shouldDeployMantleOPChain(st *state.State, chainID common.Hash) bool {
	for _, chain := range st.Chains {
		if chain.ID == chainID {
			return !chain.MantleImplContracts.AllSetUp
		}
	}
	return true
}
