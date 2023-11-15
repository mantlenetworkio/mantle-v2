package oracle

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/gas-oracle/bindings"
	ometrics "github.com/ethereum-optimism/optimism/gas-oracle/metrics"
	"github.com/ethereum-optimism/optimism/gas-oracle/tokenratio"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	//kms "cloud.google.com/go/kms/apiv1"
	//"google.golang.org/api/option"
)

func wrapUpdateTokenRatio(l1Backend bind.ContractTransactor, l2Backend DeployContractBackend, tokenRatio *tokenratio.Client, cfg *Config) (func() error, error) {
	if cfg.l2ChainID == nil {
		return nil, errNoChainID
	}

	var opts *bind.TransactOpts
	var err error
	if !cfg.EnableHsm {
		if cfg.privateKey == nil {
			return nil, errNoPrivateKey
		}
		if cfg.l2ChainID == nil {
			return nil, errNoChainID
		}

		opts, err = bind.NewKeyedTransactorWithChainID(cfg.privateKey, cfg.l2ChainID)
		if err != nil {
			return nil, err
		}
	} else {
		//seqBytes, err := hex.DecodeString(cfg.HsmCreden)
		//apikey := option.WithCredentialsJSON(seqBytes)
		//client, err := kms.NewKeyManagementClient(context.Background(), apikey)
		//if err != nil {
		//	log.Crit("gasoracle", "create signer error", err.Error())
		//}
		//mk := &bsscore.ManagedKey{
		//	KeyName:      cfg.HsmAPIName,
		//	EthereumAddr: common.HexToAddress(cfg.HsmAddress),
		//	Gclient:      client,
		//}
		//opts, err = mk.NewEthereumTransactorrWithChainID(context.Background(), cfg.l2ChainID)
		//if err != nil {
		//	log.Crit("gasoracle", "create signer error", err.Error())
		//	return nil, err
		//}
	}
	// Once https://github.com/ethereum/go-ethereum/pull/23062 is released
	// then we can remove setting the context here
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	// Don't send the transaction using the `contract` so that we can inspect
	// it beforehand
	opts.NoSend = true

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
		latestRatio := tokenRatio.TokenRatio()
		if !isDifferenceSignificant(lastTokenRatio.Uint64(), uint64(latestRatio), cfg.tokenRatioSignificanceFactor) {
			log.Warn("non significant token ratio update", "former", lastTokenRatio, "current", latestRatio)
			return nil
		}

		// Use the configured gas price if it is set,
		// otherwise use gas estimation
		if cfg.gasPrice != nil {
			opts.GasPrice = cfg.gasPrice
		} else {
			gasPrice, err := l2Backend.SuggestGasPrice(opts.Context)
			if err != nil {
				return err
			}
			opts.GasPrice = gasPrice
		}
		// set L1BaseFee to base fee + tip cap, to cover rollup tip cap
		tx, err := contract.SetTokenRatio(opts, big.NewInt(int64(latestRatio)))
		if err != nil {
			return err
		}
		log.Info("updating token ratio", "tx.gasPrice", tx.GasPrice(), "tx.gasLimit", tx.Gas(),
			"tx.data", hexutil.Encode(tx.Data()), "tx.to", tx.To().Hex(), "tx.nonce", tx.Nonce())
		if err := l2Backend.SendTransaction(context.Background(), tx); err != nil {
			return fmt.Errorf("cannot update token ratio: %w", err)
		}
		log.Info("Token ratio transaction already sent", "hash", tx.Hash().Hex(), "token ratio", int64(latestRatio))
		ometrics.GasOracleStats.TokenRatioGauge.Update(latestRatio)

		if cfg.waitForReceipt {
			// Wait for the receipt
			receipt, err := waitForReceipt(l2Backend, tx)
			if err != nil {
				return err
			}

			log.Info("Token ratio transaction confirmed", "hash", tx.Hash().Hex(),
				"gas-used", receipt.GasUsed, "blocknumber", receipt.BlockNumber)
		}
		return nil
	}, nil
}
