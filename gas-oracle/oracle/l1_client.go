package oracle

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/gas-oracle/tokenratio"
	"github.com/ethereum/go-ethereum/ethclient"
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
