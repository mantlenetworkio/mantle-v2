package dln

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"

	"github.com/Layr-Labs/datalayr/common/contracts"
	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/graphView"
	"github.com/Layr-Labs/datalayr/common/logging"
)

type Dln struct {
	config          *Config
	logger          *logging.Logger
	bls             *bls.BlsKeyPair
	metrics         *Metrics
	store           *Store
	kzgEncoderGroup *rs.KzgEncoderGroup

	GraphClient *graphView.GraphClient
	ChainClient *contracts.DataLayrChainClient
}

// NewDln creates a new DLN with the provided config
func NewDln(config *Config, logger *logging.Logger) (*Dln, error) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Make sure config folder exists
	err := os.MkdirAll(config.DbPath, os.ModePerm)
	if err != nil {
		logger.Error().Err(err).
			Msgf("Could not create db directory: %v", config.DbPath)
		return nil, err
	}

	chainLogger := logger.Sublogger("Chain")
	chainClient, err := contracts.NewDataLayrChainClient(config.ChainClientConfig, config.DlsmAddress, chainLogger)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create chain client")
		return nil, err
	}
	config.Address = chainClient.AccountAddress.Hex()

	// Create the graph client
	graphLogger := logger.Sublogger("Graph")
	gc := graphView.NewGraphClient(config.GraphProvider, graphLogger)

	// Setup metrics
	metricsLogger := logger.Sublogger("Metrics")
	metrics := NewMetrics(metricsLogger)

	// Generate BLS keys
	bls, err := bls.BlsKeysFromString(config.PrivateBls)
	if err != nil {
		logger.Error().Err(err).Msg("Could not generate bls key pair")
		return nil, err
	}

	// Create new store
	storeLogger := logger.Sublogger("Store")
	store, err := NewStore(config.DbPath+"/frame", storeLogger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create new store")
		return nil, err
	}

	// Create new KzgEncoderGroup
	kzgEncoderGroup, err := rs.NewKzgEncoderGroup(&config.KzgConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create KzgEncoderGroup")
		return nil, err
	}

	return &Dln{
		config:          config,
		logger:          logger.Sublogger("Dln"),
		bls:             bls,
		metrics:         metrics,
		store:           store,
		kzgEncoderGroup: kzgEncoderGroup,
		GraphClient:     gc,
		ChainClient:     chainClient,
	}, nil
}

// Starts the DLNs ongoing process by registering / updating its socket
// if needed and starts the GRPC server.
func (d *Dln) Start(ctx context.Context) error {
	d.logger.Trace().Msg("Entering Start function...")
	defer d.logger.Trace().Msg("Exiting Start function...")

	if d.config.EnableMetrics {
		httpSocket := fmt.Sprintf(":%s", d.config.MetricsPort)
		d.metrics.Start(httpSocket)
		d.logger.Info().Msgf("Enabled metrics with socket: %v", httpSocket)
	}

	err := d.GraphClient.WaitForSubgraph()
	if err != nil {
		d.logger.Error().Err(err).Msg("Failed waiting for subgraph")
		return err
	}

	socket := fmt.Sprintf("%s:%s", d.config.Hostname, d.config.GrpcPort)
	operator, err := d.GraphClient.QueryOperator(d.config.Address)
	if err == nil {
		// the operator is registered
		d.logger.Info().Msg("Found registered operator")
		if string(operator.Socket) != socket {
			d.logger.Warn().Err(err).Msg("Config vs chain socket mismatch")

			// If the socket from the chain is different from the one one in the config, update it
			err = d.ChainClient.UpdateSocket(ctx, socket)
			if err != nil {
				d.logger.Error().Err(err).Msg("Update Socket Failed")
				return err
			}
			d.logger.Info().Msg("Update Socket Success")
		}
	} else {
		d.logger.Warn().Err(err).Msg("Couldn't find operator")
		operatorType := uint8(3) // TODO: magic number
		d.logger.Info().Msgf("Starting self registration.")
		// DLN is not registered, attempt registration on chain
		feeParams, err := d.ChainClient.GetFeeParams()
		if err != nil {
			d.logger.Error().Err(err).Msg("Failed to get fee params")
			return err
		}

		err = d.ChainClient.Register(ctx, d.bls, operatorType, socket, feeParams.DlnStake)
		if err != nil {
			d.logger.Error().Err(err).Msg("Self registration submission failed")
			return err
		}

		d.logger.Info().Msg("Self Registration success")
	}

	// Emit metric for registered operator
	d.metrics.Registered.Set(1)

	return nil
}
