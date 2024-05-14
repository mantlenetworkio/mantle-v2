package retriever

import (
	"github.com/Layr-Labs/datalayr/common/contracts"
	"github.com/Layr-Labs/datalayr/common/graphView"
	"github.com/Layr-Labs/datalayr/common/logging"

	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
)

type Retriever struct {
	*Config

	Logger *logging.Logger

	ChainClient *contracts.DataLayrChainClient
	KzgGroup    *kzgRs.KzgEncoderGroup
	GraphClient *graphView.GraphClient
}

func NewRetriever(config *Config, logger *logging.Logger) (*Retriever, error) {
	chainLogger := logger.Sublogger("Chain")
	chainClient, err := contracts.NewDataLayrChainClient(config.ChainClientConfig, config.DlsmAddress, chainLogger)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create chain client")
		return nil, err
	}

	graphLogger := logger.Sublogger("Graph")
	graphClient := graphView.NewGraphClient(config.GraphProvider, graphLogger)

	kzgEncoderGroup, err := kzgRs.NewKzgEncoderGroup(&config.KzgConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create encoder group")
		return nil, err
	}

	ret := &Retriever{
		Config:      config,
		Logger:      logger.Sublogger("Retriever"),
		ChainClient: chainClient,
		GraphClient: graphClient,
		KzgGroup:    kzgEncoderGroup,
	}

	return ret, nil
}
