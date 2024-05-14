package chain

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	RpcUrl             string
	privateKey         *ecdsa.PrivateKey
	AccountAddress     common.Address
	ChainClient        *ethclient.Client
	NoSendTransactOpts *bind.TransactOpts
	Contracts          map[common.Address]*bind.BoundContract
	Logger             *logging.Logger
}

func NewClient(config ClientConfig, logger *logging.Logger) (*Client, error) {
	chainClient, err := ethclient.Dial(config.RpcUrl)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error. Cannot connect to provider")
		return nil, err
	}
	var accountAddress common.Address
	var privateKey *ecdsa.PrivateKey

	if len(config.PrivateKeyString) != 0 {
		privateKey, err = crypto.HexToECDSA(config.PrivateKeyString)
		if err != nil {
			logger.Error().Err(err).Msg("Invalid key. ")
			return nil, err
		}
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

		if !ok {
			logger.Error().Msg("Cannot get publicKeyECDSA")
			return nil, ErrCannotGetECDSAPubKey
		}
		accountAddress = crypto.PubkeyToAddress(*publicKeyECDSA)
	}

	chainIdBigInt := new(big.Int).SetUint64(config.ChainId)

	c := &Client{
		RpcUrl:         config.RpcUrl,
		privateKey:     privateKey,
		AccountAddress: accountAddress,
		ChainClient:    chainClient,
		Contracts:      make(map[common.Address]*bind.BoundContract),
		Logger:         logger,
	}

	// generate and memoize NoSendTransactOpts
	opts, err := bind.NewKeyedTransactorWithChainID(c.privateKey, chainIdBigInt)
	if err != nil {
		logger.Error().Err(err).Msg("Cannot create NoSendTransactOpts")
		return nil, err
	}

	opts.NoSend = true

	c.NoSendTransactOpts = opts

	return c, err
}

func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	return c.ChainClient.BlockNumber(ctx)
}
