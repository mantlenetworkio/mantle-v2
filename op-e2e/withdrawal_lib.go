package op_e2e

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func ProveAndFinalizeWithdrawalForSingleTx(t *testing.T, l1Client *ethclient.Client, ethPrivKey *ecdsa.PrivateKey, l2WithdrawalReceipt *types.Receipt) (*types.Receipt, *types.Receipt) {
	params, proveReceipt := ProveWithdrawalForSingleTx(t, l1Client, ethPrivKey, l2WithdrawalReceipt)
	finalizeReceipt := FinalizeWithdrawalForSingleTx(t, l1Client, ethPrivKey, l2WithdrawalReceipt, params)
	return proveReceipt, finalizeReceipt
}

func ProveWithdrawalForSingleTx(t *testing.T, l1Client *ethclient.Client, ethPrivKey *ecdsa.PrivateKey, l2WithdrawalReceipt *types.Receipt) (withdrawals.ProvenWithdrawalParameters, *types.Receipt) {
	// Get l2BlockNumber for proof generation
	ctx, cancel := context.WithTimeout(context.Background(), 40*15*time.Second)
	defer cancel()

	blockNumber, err := withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevOptimismPortalAddr, l2WithdrawalReceipt.BlockNumber)
	require.Nil(t, err)

	rpcClient, err := rpc.Dial("http://127.0.0.1:9545")
	require.Nil(t, err)
	proofCl := gethclient.New(rpcClient)
	receiptCl := ethclient.NewClient(rpcClient)

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Get the latest header
	header, err := receiptCl.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	require.Nil(t, err)

	// Now create withdrawal
	oracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
	require.Nil(t, err)

	params, err := withdrawals.ProveWithdrawalParameters(context.Background(), proofCl, receiptCl, l2WithdrawalReceipt.TxHash, header, oracle)
	require.Nil(t, err)

	portal, err := bindings.NewOptimismPortal(predeploys.DevOptimismPortalAddr, l1Client)
	require.Nil(t, err)

	l1ChainId, err := l1Client.NetworkID(ctx)
	require.NoError(t, err)

	opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, l1ChainId)
	require.Nil(t, err)

	// Prove withdrawal
	tx, err := portal.ProveWithdrawalTransaction(
		opts,
		bindings.TypesWithdrawalTransaction{
			Nonce:    params.Nonce,
			Sender:   params.Sender,
			Target:   params.Target,
			MntValue: params.MNTValue,
			EthValue: params.ETHValue,
			GasLimit: params.GasLimit,
			Data:     params.Data,
		},
		params.L2OutputIndex,
		params.OutputRootProof,
		params.WithdrawalProof,
	)
	require.Nil(t, err)

	// Ensure that our withdrawal was proved successfully
	proveReceipt, err := waitForTransaction(tx.Hash(), l1Client, 3*15*time.Second)
	require.Nil(t, err, "prove withdrawal")
	require.Equal(t, types.ReceiptStatusSuccessful, proveReceipt.Status)
	return params, proveReceipt
}

func FinalizeWithdrawalForSingleTx(t *testing.T, l1Client *ethclient.Client, privKey *ecdsa.PrivateKey, withdrawalReceipt *types.Receipt, params withdrawals.ProvenWithdrawalParameters) *types.Receipt {
	// Wait for finalization and then create the Finalized Withdrawal Transaction
	ctx, cancel := context.WithTimeout(context.Background(), 30*15*time.Second)
	defer cancel()
	_, err := withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevOptimismPortalAddr, withdrawalReceipt.BlockNumber)
	require.Nil(t, err)

	l1ChainId, err := l1Client.NetworkID(ctx)
	require.NoError(t, err)

	opts, err := bind.NewKeyedTransactorWithChainID(privKey, l1ChainId)
	require.Nil(t, err)
	portal, err := bindings.NewOptimismPortal(predeploys.DevOptimismPortalAddr, l1Client)
	require.Nil(t, err)
	// Finalize withdrawal
	tx, err := portal.FinalizeWithdrawalTransaction(
		opts,
		bindings.TypesWithdrawalTransaction{
			Nonce:    params.Nonce,
			Sender:   params.Sender,
			Target:   params.Target,
			MntValue: params.MNTValue,
			EthValue: params.ETHValue,
			GasLimit: params.GasLimit,
			Data:     params.Data,
		},
	)
	require.Nil(t, err)

	// Ensure that our withdrawal was finalized successfully
	finalizeReceipt, err := waitForTransaction(tx.Hash(), l1Client, 3*15*time.Second)
	require.Nil(t, err, "finalize withdrawal")
	require.Equal(t, types.ReceiptStatusSuccessful, finalizeReceipt.Status)
	return finalizeReceipt
}
