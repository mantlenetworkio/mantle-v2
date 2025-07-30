package oracle

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	ometrics "github.com/ethereum-optimism/optimism/gas-oracle/metrics"
	"github.com/ethereum-optimism/optimism/gas-oracle/tokenratio"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

var (
	// errInvalidSigningKey represents the error when the signing key used
	// is not the Owner of the contract and therefore cannot update the gasprice
	errInvalidSigningKey = errors.New("invalid signing key")
	// errNoChainID represents the error when the chain id is not provided
	// and it cannot be remotely fetched
	errNoChainID = errors.New("no chain id provided")
	// errNoPrivateKey represents the error when the private key is not provided to
	// the application
	errNoPrivateKey = errors.New("no private key provided")
	// errWrongChainID represents the error when the configured chain id is not
	// correct
	errWrongChainID = errors.New("wrong chain id provided")
	// errNoBaseFee represents the error when the base fee is not found on the
	// block. This means that the block being queried is pre eip1559
	errNoBaseFee = errors.New("base fee not found on block")
)

const (
	DefaultOperatorFeeUpdateInterval     = 300  // 5 minutes
	DefaultOperatorFeeSignificanceFactor = 0.05 // 5% threshold
)

// GasPriceOracle manages a hot key that can update the L2 Gas Price
type GasPriceOracle struct {
	l1ChainID  *big.Int
	l2ChainID  *big.Int
	ctx        context.Context
	stop       chan struct{}
	contract   *bindings.GasPriceOracle
	l2Backend  DeployContractBackend
	l1Backend  bind.ContractTransactor
	tokenRatio *tokenratio.Client
	config     *Config
	auth       *Auth

	// Operator fee constant update
	lastOperatorFeeConstant *big.Int                // Cache for the last operator fee constant to avoid contract calls
	lastOperatorFeeScalar   *big.Int                // Cache for the last operator fee scalar to avoid contract calls
	operatorFeeCalculator   *OperatorFeeCalculator  // Operator fee calculator
	explorerClient          ExplorerClientInterface // Explorer client for fetching tx count
}

// Start runs the GasPriceOracle
func (g *GasPriceOracle) Start() error {
	if g.config.l1ChainID == nil {
		return fmt.Errorf("layer-one: %w", errNoChainID)
	}
	if g.config.l2ChainID == nil {
		return fmt.Errorf("layer-two: %w", errNoChainID)
	}
	var address common.Address
	if !g.config.EnableHsm {
		if g.config.privateKey == nil {
			return errNoPrivateKey
		}
		address = crypto.PubkeyToAddress(g.config.privateKey.PublicKey)
	} else {
		address = common.HexToAddress(g.config.HsmAddress)
	}

	log.Info("Starting Gas Price Oracle", "l1-chain-id", g.l1ChainID,
		"l2-chain-id", g.l2ChainID, "address", address.Hex())

	go g.TokenRatioLoop()

	// Start operator fee update loop if configured
	if g.config.OperatorFeeUpdateEnabled {
		go g.OperatorFeeLoop()
	}

	return nil
}

func (g *GasPriceOracle) Stop() {
	close(g.stop)
}

func (g *GasPriceOracle) Wait() {
	<-g.stop
}

// ensure makes sure that the configured private key is the operator
// of the `GasPriceOracle`. If it is not the operator, then it will
// not be able to make updates to the token ratio.
func (g *GasPriceOracle) ensure() error {
	operator, err := g.contract.Operator(&bind.CallOpts{
		Context: g.ctx,
	})
	if err != nil {
		return err
	}
	var address common.Address
	if g.config.EnableHsm {
		address = common.HexToAddress(g.config.HsmAddress)
	} else {
		address = crypto.PubkeyToAddress(g.config.privateKey.PublicKey)
	}
	if address != operator {
		log.Error("Signing key does not match contract operator", "signer", address.Hex(), "operator", operator.Hex())
		return errInvalidSigningKey
	}
	return nil
}

func (g *GasPriceOracle) TokenRatioLoop() {
	timer := time.NewTicker(time.Duration(g.config.tokenRatioEpochLengthSeconds) * time.Second)
	defer timer.Stop()

	updateTokenRatio, err := wrapUpdateTokenRatio(g.l1Backend, g.l2Backend, g.tokenRatio, g.config, g.auth)
	if err != nil {
		panic(err)
	}
	for {
		select {
		case <-timer.C:
			if err := updateTokenRatio(); err != nil {
				log.Error("cannot update tokenRatio", "message", err)
			}
		case <-g.ctx.Done():
			g.Stop()
		}
	}
}

func (g *GasPriceOracle) OperatorFeeLoop() {
	updateInterval := g.config.OperatorFeeUpdateInterval
	if updateInterval == 0 {
		updateInterval = DefaultOperatorFeeUpdateInterval
		log.Info("Operator fee update interval is not set, setting to default", "interval", updateInterval)
	}

	timer := time.NewTicker(time.Duration(updateInterval) * time.Second)
	defer timer.Stop()

	log.Info("Starting operator fee update loop",
		"update_interval_seconds", updateInterval)

	for {
		select {
		case <-timer.C:
			var wg sync.WaitGroup

			// Update operator fee constant
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := g.updateOperatorFeeConstant(); err != nil {
					log.Error("Failed to update operator fee constant", "error", err)
				} else {
					log.Debug("Successfully updated operator fee constant")
				}
			}()

			// Update operator fee scalar
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := g.updateOperatorFeeScalar(); err != nil {
					log.Error("Failed to update operator fee scalar", "error", err)
				} else {
					log.Debug("Successfully updated operator fee scalar")
				}
			}()

			wg.Wait()
		case <-g.ctx.Done():
			log.Info("Stopping operator fee update loop")
			return
		}
	}
}

// NewGasPriceOracle creates a new GasPriceOracle based on a Config
func NewGasPriceOracle(cfg *Config) (*GasPriceOracle, error) {
	tokenRatioClient, err := NewTokenRatioClient(cfg.ethereumHttpUrl, cfg.layerTwoHttpUrl, cfg.tokenRatioCexURL, cfg.tokenRatioDexURL,
		cfg.tokenRatioUpdateFrequencySecond)
	if err != nil {
		return nil, err
	}

	// Ensure that we can actually connect to both backends
	log.Info("Connecting to layer two")
	if err := ensureConnection(tokenRatioClient.l2Client); err != nil {
		log.Error("Unable to connect to layer two")
		return nil, err
	}
	log.Info("Connecting to layer one")
	if err := ensureConnection(tokenRatioClient.l1Client); err != nil {
		log.Error("Unable to connect to layer one")
		return nil, err
	}

	address := cfg.gasPriceOracleAddress
	contract, err := bindings.NewGasPriceOracle(address, tokenRatioClient.l2Client)
	if err != nil {
		return nil, err
	}

	// Fetch the current gas price to use as the current price
	currentPrice, err := contract.GasPrice(&bind.CallOpts{
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}

	// Create a gas pricer for the gas price updater
	log.Info("Creating GasPricer", "currentPrice", currentPrice)

	l2ChainID, err := tokenRatioClient.l2Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	l1ChainID, err := tokenRatioClient.l1Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	if cfg.l2ChainID != nil {
		if cfg.l2ChainID.Cmp(l2ChainID) != 0 {
			return nil, fmt.Errorf("%w: L2: configured with %d and got %d",
				errWrongChainID, cfg.l2ChainID, l2ChainID)
		}
	} else {
		cfg.l2ChainID = l2ChainID
	}

	if cfg.l1ChainID != nil {
		if cfg.l1ChainID.Cmp(l1ChainID) != 0 {
			return nil, fmt.Errorf("%w: L1: configured with %d and got %d",
				errWrongChainID, cfg.l1ChainID, l1ChainID)
		}
	} else {
		cfg.l1ChainID = l1ChainID
	}

	if !cfg.EnableHsm && cfg.privateKey == nil {
		return nil, errNoPrivateKey
	}

	// Create a transaction signer
	auth, err := NewAuth(cfg, tokenRatioClient.l2Client)
	if err != nil {
		return nil, err
	}

	// Fetch the current operator fee constant and create a calculator if enabled
	var currentOperatorFeeConstant *big.Int
	var currentOperatorFeeScalar *big.Int
	var operatorFeeCalculator *OperatorFeeCalculator
	var explorerClient ExplorerClientInterface
	if cfg.OperatorFeeUpdateEnabled {
		log.Info("Operator fee update is enabled")

		currentOperatorFeeConstant, currentOperatorFeeScalar, err = fetchCurrentOperatorFee(contract)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch current operator fee: %w", err)
		}

		ometrics.GasOracleStats.OperatorFeeConstantGauge.Update(currentOperatorFeeConstant.Int64())
		ometrics.GasOracleStats.OperatorFeeScalarGauge.Update(currentOperatorFeeScalar.Int64())

		log.Info("Initialized operator fee constant cache", "value", currentOperatorFeeConstant.String())
		log.Info("Initialized operator fee scalar cache", "value", currentOperatorFeeScalar.String())

		operatorFeeCalculator = NewOperatorFeeCalculator(cfg.IntrinsicSp1GasPerTx, cfg.IntrinsicSp1GasPerBlock, cfg.Sp1PricePerBGasInDollars, cfg.Sp1GasScalar, cfg.OperatorFeeMarkupPercentage)

		if cfg.BlockscoutExplorerURL != "" {
			explorerClient = NewBlockscoutClient(cfg.BlockscoutExplorerURL)
		} else {
			if cfg.EtherscanAPIKey == "" {
				return nil, fmt.Errorf("etherscan api key is not set")
			}
			explorerClient = NewEtherscanClient(cfg.EtherscanExplorerURL, cfg.EtherscanAPIKey)
		}
	}

	gpo := GasPriceOracle{
		l2ChainID:               l2ChainID,
		l1ChainID:               l1ChainID,
		ctx:                     context.Background(),
		stop:                    make(chan struct{}),
		contract:                contract,
		config:                  cfg,
		l2Backend:               tokenRatioClient.l2Client,
		l1Backend:               tokenRatioClient.l1Client,
		tokenRatio:              tokenRatioClient.tokenRatio,
		auth:                    auth,
		lastOperatorFeeConstant: currentOperatorFeeConstant,
		lastOperatorFeeScalar:   currentOperatorFeeScalar,
		operatorFeeCalculator:   operatorFeeCalculator,
		explorerClient:          explorerClient,
	}

	if err := gpo.ensure(); err != nil {
		return nil, err
	}

	return &gpo, nil
}

// updateOperatorFeeConstant calculate and update operator fee constant
func (g *GasPriceOracle) updateOperatorFeeConstant() error {
	// Step 1: Get current ETH price from token ratio client
	ethPrice := g.tokenRatio.EthPrice()

	// Step 2: Fetch transaction count from the explorer client
	// DailyTxCountFromUser is the transaction count from the explorer client minus the daily block count
	txCount, err := g.explorerClient.DailyTxCountFromUser(g.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction count: %w", err)
	}

	// Step 3: Calculate new operator fee constant using the calculator function
	newConstant, err := g.operatorFeeCalculator.CalOperatorFeeConstant(txCount, ethPrice)
	if err != nil {
		return fmt.Errorf("failed to calculate operator fee constant: %w", err)
	}

	// Step 4: Get cached operator fee constant
	currentConstant := g.lastOperatorFeeConstant

	log.Debug("Getting cached operator fee constant", "cached_value", currentConstant.String())

	// Step 5: Only update if the value has changed by more than the significance factor
	significanceFactor := g.config.OperatorFeeSignificanceFactor
	if significanceFactor <= 0 {
		significanceFactor = DefaultOperatorFeeSignificanceFactor
	}
	if isDifferenceSignificant(currentConstant.Uint64(), newConstant.Uint64(), significanceFactor) {
		log.Info("Updating operator fee constant - change exceeds threshold",
			"current", currentConstant.String(),
			"new", newConstant.String())

		return g.setOperatorFeeConstant(newConstant)
	} else {
		log.Debug("Operator fee constant unchanged or change is below threshold, skipping update",
			"current_value", currentConstant.String())
	}

	return nil
}

// updateOperatorFeeConstantOnContract updates the operator fee constant on the smart contract
func (g *GasPriceOracle) setOperatorFeeConstant(newConstant *big.Int) error {
	// Send transaction to update operator fee constant
	tx, err := g.contract.SetOperatorFeeConstant(g.auth.Opts(), newConstant)
	if err != nil {
		return fmt.Errorf("failed to update operator fee constant: %w", err)
	}

	log.Info("Operator fee constant update transaction sent",
		"tx_hash", tx.Hash().Hex())

	// Wait for receipt if configured
	if g.config.waitForReceipt {
		// Wait for the receipt
		receipt, err := waitForReceiptWithMaxRetries(g.l2Backend, tx, 30)
		if err != nil {
			return err
		}
		log.Info("Operator fee constant update transaction confirmed",
			"tx_hash", tx.Hash().Hex(),
			"block_number", receipt.BlockNumber)

		// Update the cache with the new value
		g.lastOperatorFeeConstant = newConstant
		ometrics.GasOracleStats.OperatorFeeConstantGauge.Update(newConstant.Int64())
	}

	return nil
}

// updateOperatorFeeScalar calculate and update operator fee scalar
func (g *GasPriceOracle) updateOperatorFeeScalar() error {
	// Get current ETH price from token ratio client
	ethPrice := g.tokenRatio.EthPrice()

	// Calculate new operator fee scalar based on ETH price
	newScalar, err := g.operatorFeeCalculator.CalOperatorFeeScalar(ethPrice)
	if err != nil {
		return fmt.Errorf("failed to calculate operator fee scalar: %w", err)
	}

	// Get cached operator fee scalar
	currentScalar := g.lastOperatorFeeScalar

	// Only update if the value has changed by more than the significance factor
	significanceFactor := g.config.OperatorFeeSignificanceFactor
	if significanceFactor <= 0 {
		significanceFactor = DefaultOperatorFeeSignificanceFactor
	}
	if isDifferenceSignificant(currentScalar.Uint64(), newScalar.Uint64(), significanceFactor) {
		log.Info("Updating operator fee scalar - change exceeds threshold",
			"current", currentScalar.String(),
			"new", newScalar.String())

		return g.setOperatorFeeScalar(newScalar)
	} else {
		log.Debug("Operator fee scalar unchanged or change is below threshold, skipping update",
			"current_value", currentScalar.String())
	}

	return nil
}

// setOperatorFeeScalar updates the operator fee scalar on the smart contract
func (g *GasPriceOracle) setOperatorFeeScalar(newScalar *big.Int) error {
	// Send transaction to update operator fee scalar
	tx, err := g.contract.SetOperatorFeeScalar(g.auth.Opts(), newScalar)
	if err != nil {
		return fmt.Errorf("failed to update operator fee scalar: %w", err)
	}

	log.Info("Operator fee scalar update transaction sent",
		"tx_hash", tx.Hash().Hex())

	// Wait for receipt if configured
	if g.config.waitForReceipt {
		// Wait for the receipt
		receipt, err := waitForReceiptWithMaxRetries(g.l2Backend, tx, 30)
		if err != nil {
			return err
		}
		log.Info("Operator fee scalar update transaction confirmed",
			"tx_hash", tx.Hash().Hex(),
			"block_number", receipt.BlockNumber)

		// Update the cache with the new value
		g.lastOperatorFeeScalar = newScalar
		ometrics.GasOracleStats.OperatorFeeScalarGauge.Update(newScalar.Int64())
	}

	return nil
}

// fetchCurrentOperatorFee fetches the current operator fee constant and scalar from the contract
func fetchCurrentOperatorFee(contract *bindings.GasPriceOracle) (*big.Int, *big.Int, error) {
	currentOperatorFeeConstant, err := contract.OperatorFeeConstant(&bind.CallOpts{
		Context: context.Background(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch current operator fee constant: %w", err)
	}

	currentOperatorFeeScalar, err := contract.OperatorFeeScalar(&bind.CallOpts{
		Context: context.Background(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch current operator fee scalar: %w", err)
	}

	return currentOperatorFeeConstant, currentOperatorFeeScalar, nil
}

// Ensure that we can actually connect
func ensureConnection(client *ethclient.Client) error {
	t := time.NewTicker(1 * time.Second)
	retries := 0
	defer t.Stop()
	for ; true; <-t.C {
		_, err := client.ChainID(context.Background())
		if err == nil {
			break
		} else {
			retries += 1
			if retries > 90 {
				return err
			}
		}
	}
	return nil
}
