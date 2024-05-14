package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	FallbackGasTipCap = big.NewInt(15000000000)
)

// EstimateGasPriceAndLimitAndSendTx sends and returns an otherwise identical txn
// to the one provided but with updated gas prices sampled from the existing network
// conditions and an accurate gasLimit
//
// Note: tx must be a to a contract, not an EOA
//
// Slightly modified from: https://github.com/ethereum-optimism/optimism/blob/ec266098641820c50c39c31048aa4e953bece464/batch-submitter/drivers/sequencer/driver.go#L314
func (c *Client) EstimateGasPriceAndLimitAndSendTx(
	ctx context.Context,
	tx *types.Transaction,
	tag string,
) (*common.Hash, error) {
	log := c.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering EstimateGasPriceAndLimitAndSendTx function...")
	defer log.Trace().Msg("Exiting EstimateGasPriceAndLimitAndSendTx function...")

	gasTipCap, err := c.ChainClient.SuggestGasTipCap(ctx)
	if err != nil {
		// If the transaction failed because the backend does not support
		// eth_maxPriorityFeePerGas, fallback to using the default constant.
		// Currently Alchemy is the only backend provider that exposes this
		// method, so in the event their API is unreachable we can fallback to a
		// degraded mode of operation. This also applies to our test
		// environments, as hardhat doesn't support the query either.
		log.Info().Msg("eth_maxPriorityFeePerGas is unsupported by current backend, using fallback gasTipCap")
		gasTipCap = FallbackGasTipCap
	}

	header, err := c.ChainClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	gasFeeCap := new(big.Int).Add(header.BaseFee, gasTipCap)

	// The estimated gas limits performed by RawTransact fail semi-regularly
	// with out of gas exceptions. To remedy this we extract the internal calls
	// to perform gas price/gas limit estimation here and add a buffer to
	// account for any network variability.
	gasLimit, err := c.ChainClient.EstimateGas(ctx, ethereum.CallMsg{
		From:      c.AccountAddress,
		To:        tx.To(),
		GasPrice:  nil,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     nil,
		Data:      tx.Data(),
	})
	if err != nil {
		return nil, err
	}

	opts, err := bind.NewKeyedTransactorWithChainID(c.privateKey, tx.ChainId())
	if err != nil {
		log.Error().Err(err).Msg("Cannot create transactOpts")
		return nil, err
	}
	opts.Context = ctx
	opts.Nonce = new(big.Int).SetUint64(tx.Nonce())
	opts.GasTipCap = gasTipCap
	opts.GasFeeCap = gasFeeCap
	opts.GasLimit = addGasBuffer(gasLimit)

	contract := c.Contracts[*tx.To()]
	// if the contract has not been cached
	if contract == nil {
		// create a dummy bound contract tied to the `to` address of the transaction
		contract = bind.NewBoundContract(*tx.To(), abi.ABI{}, c.ChainClient, c.ChainClient, c.ChainClient)
		// cache the contract for later use
		c.Contracts[*tx.To()] = contract
	}

	tx, err = contract.RawTransact(opts, tx.Data())

	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("Error sending %s tx", tag))
		return nil, err
	}

	err = c.EnsureTransactionEvaled(
		tx,
		tag,
	)
	txHash := tx.Hash()
	log.Trace().Str("TxHash", txHash.Hex()).Msg(fmt.Sprintf("%s complete", tag))

	return &txHash, err
}

func (c *Client) EnsureTransactionEvaled(tx *types.Transaction, tag string) error {
	receipt, err := bind.WaitMined(context.Background(), c.ChainClient, tx)
	if err != nil {
		c.Logger.Error().Err(err).Str("tag", tag).Msg("Error waiting for transaction to mine")
		return err
	}
	if receipt.Status != 1 {
		c.Logger.Error().Err(err).Str("tag", tag).Str("txHash", tx.Hash().Hex()).Uint64("status", receipt.Status).Uint64("GasUsed", receipt.GasUsed).Msg("[ChainIO] Transaction Failed")
		return ErrTransactionFailed
	}
	return nil
}

func addGasBuffer(gasLimit uint64) uint64 {
	return 6 * gasLimit / 5 // add 20% buffer to gas limit
}
