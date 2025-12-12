package pipeline

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/standard"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
)

const (
	SubmissionInterval        = 10
	L2BlockTime               = 2
	StartingBlockNumber       = 0
	StartingTimestamp         = 0
	FinalizationPeriodSeconds = 2
	MantlePortalGuardian      = "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
	Paused                    = true
)

type MantleDeployParamsSet struct {
	BasefeeScalar                           uint32
	BlobbasefeeScalar                       uint32
	L2OutputOracleSubmissionInterval        *big.Int
	L2OutputOracleL2BlockTime               *big.Int
	L2OutputOracleStartingBlockNumber       *big.Int
	L2OutputOracleStartingTimestamp         *big.Int
	L2OutputOracleFinalizationPeriodSeconds *big.Int
	OptimismPortalGuardian                  common.Address
	OptimismPortalPaused                    bool
}

func DeployMantleImplementations(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "deploy-mantle-implementations")

	if !shouldDeployMantleImplementations(intent, st, chainID) {
		lgr.Info("mantle implementations deployment not needed")
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

	lgr.Info("deploying Mantle implementations")

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

	dii := opcm.MantleDeployImplementationsInput{
		SystemConfigOwner:                       thisIntent.Roles.SystemConfigOwner,
		SystemConfigBatcherHash:                 common.BytesToHash(thisIntent.Roles.Batcher.Bytes()),
		SystemConfigGasLimit:                    thisIntent.GasLimit,
		SystemConfigBaseFee:                     big.NewInt(int64(thisIntent.MinBaseFee)),
		SystemConfigUnsafeBlockSigner:           thisIntent.Roles.UnsafeBlockSigner,
		SystemConfigConfig:                      opcm.DecodeResourceConfig(),
		SystemConfigBasefeeScalar:               deployParams.BasefeeScalar,
		SystemConfigBlobbasefeeScalar:           deployParams.BlobbasefeeScalar,
		OptimismPortal:                          thisState.OptimismPortalProxy,
		L1MNT:                                   thisIntent.L1MNT,
		L1CrossDomainMessenger:                  thisState.L1CrossDomainMessengerProxy,
		L2OutputOracle:                          thisState.L2OutputOracleProxy,
		SystemConfig:                            thisState.SystemConfigProxy,
		L1StandardBridge:                        thisState.L1StandardBridgeProxy,
		L1ERC721BridgeOtherBridge:               common.HexToAddress(predeploys.L2ERC721Bridge),
		L2OutputOracleSubmissionInterval:        deployParams.L2OutputOracleSubmissionInterval,
		L2OutputOracleL2BlockTime:               deployParams.L2OutputOracleL2BlockTime,
		L2OutputOracleStartingBlockNumber:       deployParams.L2OutputOracleStartingBlockNumber,
		L2OutputOracleStartingTimestamp:         deployParams.L2OutputOracleStartingTimestamp,
		L2OutputOracleProposer:                  thisIntent.Roles.Proposer,
		L2OutputOracleChallenger:                thisIntent.Roles.Challenger,
		L2OutputOracleFinalizationPeriodSeconds: deployParams.L2OutputOracleFinalizationPeriodSeconds,
		OptimismPortalGuardian:                  deployParams.OptimismPortalGuardian,
		OptimismPortalPaused:                    deployParams.OptimismPortalPaused,
	}

	dio, err := env.MantleScripts.DeployImplementations.Run(dii)
	if err != nil {
		return fmt.Errorf("error deploying Mantle implementations: %w", err)
	}

	thisState.OptimismPortalImpl = dio.OptimismPortalImpl
	thisState.SystemConfigImpl = dio.SystemConfigImpl
	thisState.L1CrossDomainMessengerImpl = dio.L1CrossDomainMessengerImpl
	thisState.L1Erc721BridgeImpl = dio.L1ERC721BridgeImpl
	thisState.L1StandardBridgeImpl = dio.L1StandardBridgeImpl
	thisState.OptimismMintableErc20FactoryImpl = dio.OptimismMintableERC20FactoryImpl
	thisState.L2OutputOracleImpl = dio.L2OutputOracleImpl

	// New an empty instance of implementations deployment to avoid nil pointer dereference
	st.ImplementationsDeployment = &addresses.ImplementationsContracts{}

	return nil
}

func shouldDeployMantleImplementations(intent *state.Intent, st *state.State, chainID common.Hash) bool {
	for _, chain := range st.Chains {
		if chain.ID == chainID {
			return chain.SystemConfigImpl == common.Address{}
		}
	}
	return true
}
