package oracle

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"

	ometrics "github.com/ethereum-optimism/optimism/gas-oracle/metrics"
	"github.com/ethereum-optimism/optimism/gas-oracle/tokenratio"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
)

func wrapUpdateTokenRatio(l1Backend bind.ContractTransactor, l2Backend DeployContractBackend, tokenRatio *tokenratio.Client, cfg *Config, auth *Auth) (func() error, error) {
	if cfg.l2ChainID == nil {
		return nil, errNoChainID
	}

	// Create a new contract bindings in scope of the updateL2GasPriceFn
	// that is returned from this function
	contract, err := bindings.NewGasPriceOracle(cfg.gasPriceOracleAddress, l2Backend)
	if err != nil {
		return nil, err
	}

	// initialize some metrics
	// initialize fee scalar from contract
	feeScalar, err := contract.Scalar(&bind.CallOpts{
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}
	ometrics.GasOracleStats.FeeScalarGauge.Update(feeScalar.Int64())

	return func() error {
		lastTokenRatio, err := contract.TokenRatio(&bind.CallOpts{
			Context: context.Background(),
		})
		if err != nil {
			return err
		}
		l1BaseFee, err := contract.L1BaseFee(&bind.CallOpts{
			Context: context.Background(),
		})
		if err != nil {
			return err
		}
		feeScalar, err := contract.Scalar(&bind.CallOpts{
			Context: context.Background(),
		})
		if err != nil {
			return err
		}
		// Update fee scalar & l1 base fee metrics
		ometrics.GasOracleStats.FeeScalarGauge.Update(feeScalar.Int64())
		ometrics.GasOracleStats.L1BaseFeeGauge.Update(l1BaseFee.Int64())

		// NOTE this will return base multiple with coin ratio
		latestRatio := tokenRatio.TokenRatio() * cfg.tokenRatioScalar
		ometrics.GasOracleStats.TokenRatioGauge.Update(tokenRatio.TokenRatio())
		ometrics.GasOracleStats.TokenRatioWithScalarGauge.Update(latestRatio)
		if !isDifferenceSignificant(lastTokenRatio.Uint64(), uint64(latestRatio), cfg.tokenRatioSignificanceFactor) {
			log.Warn("non significant tokenRatio update", "former", lastTokenRatio, "current", latestRatio)
			return nil
		}

		opts := auth.Opts()
		// set L1BaseFee to base fee + tip cap, to cover rollup tip cap
		tx, err := contract.SetTokenRatio(opts, big.NewInt(int64(latestRatio)))
		if err != nil {
			return fmt.Errorf("cannot update tokenRatio: %w", err)
		}
		log.Info("updating tokenRatio", "tx.gasPrice", tx.GasPrice(), "tx.gasLimit", tx.Gas(),
			"tx.data", hexutil.Encode(tx.Data()), "tx.to", tx.To().Hex(), "tx.nonce", tx.Nonce())
		log.Info("TokenRatio transaction already sent", "hash", tx.Hash().Hex(), "tokenRatio", int64(latestRatio))
		ometrics.GasOracleStats.TokenRatioOnchainGauge.Update(latestRatio)

		if cfg.waitForReceipt {
			// Wait for the receipt
			receipt, err := waitForReceiptWithMaxRetries(l2Backend, tx, 30)
			if err != nil {
				return err
			}

			log.Info("TokenRatio transaction confirmed", "hash", tx.Hash().Hex(),
				"gasUsed", receipt.GasUsed, "blockNumber", receipt.BlockNumber)
		}
		return nil
	}, nil
}
