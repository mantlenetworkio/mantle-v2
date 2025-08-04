package oracle

import (
	"context"
	"encoding/hex"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/ethereum-optimism/optimism/op-service/hsm"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/api/option"
)

type Auth struct {
	client        *ethclient.Client
	auth          *bind.TransactOpts
	fixedGasPrice bool
}

func NewAuth(cfg *Config, client *ethclient.Client) (*Auth, error) {
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
		seqBytes, err := hex.DecodeString(cfg.HsmCreden)
		if err != nil {
			log.Warn("gasoracle/auth", "decode hsm creden fail", err.Error())
			return nil, err
		}
		apikey := option.WithCredentialsJSON(seqBytes)
		client, err := kms.NewKeyManagementClient(context.Background(), apikey)
		if err != nil {
			log.Warn("gasoracle/auth", "create signer error", err.Error())
			return nil, err
		}
		mk := &hsm.ManagedKey{
			KeyName:      cfg.HsmAPIName,
			EthereumAddr: common.HexToAddress(cfg.HsmAddress),
			Gclient:      client,
		}
		opts, err = mk.NewEthereumTransactorWithChainID(context.Background(), cfg.l2ChainID)
		if err != nil {
			log.Warn("gasoracle/auth", "create signer error", err.Error())
			return nil, err
		}
	}
	// Once https://github.com/ethereum/go-ethereum/pull/23062 is released
	// then we can remove setting the context here
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	fixedGasPrice := false

	// Use the configured gas price if it is set,
	// otherwise use gas estimation
	if cfg.gasPrice != nil {
		fixedGasPrice = true
		opts.GasPrice = cfg.gasPrice
	} else {
		gasPrice, err := client.SuggestGasPrice(opts.Context)
		if err != nil {
			return nil, err
		}
		opts.GasPrice = gasPrice
	}

	return &Auth{
		client:        client,
		auth:          opts,
		fixedGasPrice: fixedGasPrice,
	}, nil
}

func (a *Auth) Opts() *bind.TransactOpts {
	// Update gas price if needed
	if !a.fixedGasPrice {
		gasPrice, err := a.client.SuggestGasPrice(a.auth.Context)
		if err != nil {
			log.Error("gasoracle/auth", "update gas price error", err.Error())
		} else {
			a.auth.GasPrice = gasPrice
		}
	}

	return a.auth
}
