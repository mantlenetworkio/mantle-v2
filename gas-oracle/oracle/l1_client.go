package oracle

import (
	"context"
	"fmt"
	"math/big"

	ometrics "github.com/ethereum-optimism/optimism/gas-oracle/metrics"
	"github.com/ethereum-optimism/optimism/gas-oracle/tokenratio"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

type TokenRatioClient struct {
	l1Client   *ethclient.Client
	l2Client   *ethclient.Client
	tokenRatio *tokenratio.Client
}

func NewTokenRatioClient(ethereumHttpUrl, layerTwoHttpUrl, tokenRatioCexURL, tokenRatioDexURL string, tokenRatioUpdateFrequencySecond uint64) (*TokenRatioClient, error) {
	l1Client, err := ethclient.Dial(ethereumHttpUrl)
	if err != nil {
		return nil, err
	}
	l2Client, err := ethclient.Dial(layerTwoHttpUrl)
	if err != nil {
		return nil, err
	}
	tokenRatio := tokenratio.NewClient(tokenRatioCexURL, tokenRatioDexURL,
		tokenRatioUpdateFrequencySecond)
	if tokenRatio == nil {
		return nil, fmt.Errorf("invalid token price client")
	}
	return &TokenRatioClient{
		l1Client:   l1Client,
		l2Client:   l2Client,
		tokenRatio: tokenRatio,
	}, nil
}

func (c *TokenRatioClient) TokenRatio(ctx context.Context, number *big.Int) (float64, error) {
	ratio := c.tokenRatio.TokenRatio()
	log.Info("show base fee context", "ratio", ratio)
	ometrics.GasOracleStats.TokenRatioGauge.Update(ratio)
	return ratio, nil
}
