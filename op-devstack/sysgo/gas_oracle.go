package sysgo

import (
	"github.com/ethereum-optimism/optimism/gas-oracle/oracle"
	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/intentbuilder"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/crypto"
)

// GasOracleService wraps the gas-oracle service for the devstack.
type GasOracleService struct {
	id      stack.GasOracleID
	service *oracle.GasPriceOracle
	l1RPC   string
	l2RPC   string
}

// GasOracleOption is a function that modifies the gas oracle configuration.
type GasOracleOption func(cfg *oracle.Config)

// WithTokenRatioEpochLength sets the token ratio epoch length.
func WithTokenRatioEpochLength(seconds uint64) GasOracleOption {
	return func(cfg *oracle.Config) {
		cfg.TokenRatioEpochLengthSeconds = seconds
	}
}

// WithTokenRatioCexURL sets the CEX URL for token ratio.
func WithTokenRatioCexURL(url string) GasOracleOption {
	return func(cfg *oracle.Config) {
		cfg.TokenRatioCexURL = url
	}
}

// WithTokenRatioDexURL sets the DEX URL for token ratio.
func WithTokenRatioDexURL(url string) GasOracleOption {
	return func(cfg *oracle.Config) {
		cfg.TokenRatioDexURL = url
	}
}

// We temporarily use the challenger address as the gas price oracle owner.
// WithL2GasPriceOracleOwner should be used with WithFundedL2GasPriceOracleOwner to fund the challenger address.
// Since L2 configurator does not support setting the GasPriceOracleOwner, we set it in the deployer pipeline.
func WithL2GasPriceOracleOwner() DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		for _, l2 := range intent.Chains {
			cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting GasPriceOracleOwner", "owner", l2.Roles.Challenger.Hex())
			l2.GasPriceOracleOwner = l2.Roles.Challenger
		}
	}
}

// WithFundedL2GasPriceOracleOwner funds the challenger role for the L2 GasPriceOracle owner.
func WithFundedL2GasPriceOracleOwner() stack.Option[*Orchestrator] {
	return stack.BeforeDeploy(func(o *Orchestrator) {
		o.P().Require().NotNil(o.wb, "must have a world builder")
		l2s := o.wb.builder.L2s()
		for _, l2 := range l2s {
			addrFor := intentbuilder.RoleToAddrProvider(o.P(), o.keys, l2.ChainID())
			l2.WithPrefundedAccount(addrFor(devkeys.ChallengerRole), *eth.BillionEther.ToU256())
		}
	})
}

func WithGasOracleService(
	gasOracleID stack.GasOracleID,
	l1ELID stack.L1ELNodeID,
	l2ELID stack.L2ELNodeID,
	opts ...GasOracleOption,
) stack.Option[*Orchestrator] {
	return stack.AfterDeploy(func(orch *Orchestrator) {
		p := orch.P().WithCtx(stack.ContextWithID(orch.P().Ctx(), gasOracleID))

		require := p.Require()
		require.False(orch.gasOracles.Has(gasOracleID), "gas oracle must not already exist")

		l1EL, ok := orch.l1ELs.Get(l1ELID)
		require.True(ok, "L1 EL node must exist")

		l2EL, ok := orch.l2ELs.Get(l2ELID)
		require.True(ok, "L2 EL node must exist")

		// Get the operator key (challenger role is used as the GasPriceOracle operator)
		operatorKey, err := orch.keys.Secret(devkeys.ChallengerRole.Key(l2ELID.ChainID().ToBig()))
		require.NoError(err, "failed to get operator key")

		logger := p.Logger()
		logger.SetContext(p.Ctx())
		logger.Info("Starting GasOracle service",
			"l1RPC", l1EL.UserRPC(),
			"l2RPC", l2EL.UserRPC(),
			"operator", crypto.PubkeyToAddress(operatorKey.PublicKey).Hex())

		// Create the gas oracle config
		oracleCfg := &oracle.Config{
			EthereumHttpUrl:                 l1EL.UserRPC(),
			LayerTwoHttpUrl:                 l2EL.UserRPC(),
			PrivateKey:                      operatorKey,
			L1ChainID:                       l1ELID.ChainID().ToBig(),
			L2ChainID:                       l2ELID.ChainID().ToBig(),
			GasPriceOracleAddress:           predeploys.GasPriceOracleAddr,
			TokenRatioEpochLengthSeconds:    1,
			TokenRatioSignificanceFactor:    0.05,
			TokenRatioScalar:                1.0,
			TokenRatioCexURL:                "https://api.bybit.com",
			TokenRatioDexURL:                l2EL.UserRPC(), // fault url
			TokenRatioUpdateFrequencySecond: 1,
		}

		// Apply options
		for _, opt := range opts {
			opt(oracleCfg)
		}

		gpo, err := oracle.NewGasPriceOracle(oracleCfg)
		require.NoError(err, "failed to create gas oracle")

		err = gpo.Start()
		require.NoError(err, "failed to start gas oracle")

		p.Cleanup(func() {
			logger.Info("Stopping GasOracle service")
			gpo.Stop()
			logger.Info("Stopped GasOracle service")
		})

		g := &GasOracleService{
			id:      gasOracleID,
			service: gpo,
			l1RPC:   l1EL.UserRPC(),
			l2RPC:   l2EL.UserRPC(),
		}
		orch.gasOracles.Set(gasOracleID, g)
	})
}

// WithGasOracle adds a gas oracle service to the orchestrator.
// The gas oracle is responsible for updating the token ratio on the GasPriceOracle contract.
func WithGasOracle(
	l1ELID stack.L1ELNodeID,
	l2ELID stack.L2ELNodeID,
	opts ...GasOracleOption,
) stack.Option[*Orchestrator] {
	return stack.Combine(
		WithGasOracleService(stack.NewGasOracleID("main", l2ELID.ChainID()), l1ELID, l2ELID, opts...),
		WithFundedL2GasPriceOracleOwner(),
		WithDeployerPipelineOption(WithL2GasPriceOracleOwner()),
	)
}
