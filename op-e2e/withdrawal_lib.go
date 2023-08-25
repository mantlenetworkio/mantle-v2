package op_e2e

//
//import (
//	"context"
//	"crypto/ecdsa"
//	"math/big"
//	"testing"
//	"time"
//
//	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
//	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
//	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
//
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/ethclient"
//	"github.com/stretchr/testify/require"
//)
//
//func ProveAndFinalizeWithdrawal(t *testing.T, cfg SystemConfig, l1Client *ethclient.Client, ethPrivKey *ecdsa.PrivateKey, l2WithdrawalReceipt *types.Receipt) (*types.Receipt, *types.Receipt) {
//	params, proveReceipt := ProveWithdrawal(t, cfg, l1Client, ethPrivKey, l2WithdrawalReceipt)
//	finalizeReceipt := FinalizeWithdrawal(t, cfg, l1Client, ethPrivKey, l2WithdrawalReceipt, params)
//	return proveReceipt, finalizeReceipt
//}
//
//func ProveWithdrawal(t *testing.T, l1Client *ethclient.Client, ethPrivKey *ecdsa.PrivateKey, l2WithdrawalReceipt *types.Receipt) (withdrawals.ProvenWithdrawalParameters, *types.Receipt) {
//	// Get l2BlockNumber for proof generation
//
//	ctx, cancel := context.WithTimeout(context.Background(), 40*15*time.Second)
//	defer cancel()
//	blockNumber, err := withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevOptimismPortalAddr, l2WithdrawalReceipt.BlockNumber)
//	require.Nil(t, err)
//
//	l2Client, err := ethclient.Dial("http://127.0.0.1:8545")
//	require.Nil(t, err)
//
//	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	// Get the latest header
//	header, err := l2Client.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
//	require.Nil(t, err)
//
//	// Now create withdrawal
//	oracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
//	require.Nil(t, err)
//
//	params, err := withdrawals.ProveWithdrawalParameters(context.Background(), proofCl, receiptCl, l2WithdrawalReceipt.TxHash, header, oracle)
//	require.Nil(t, err)
//
//	portal, err := bindings.NewOptimismPortal(predeploys.DevOptimismPortalAddr, l1Client)
//	require.Nil(t, err)
//
//	opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, cfg.L1ChainIDBig())
//	require.Nil(t, err)
//
//	// Prove withdrawal
//	tx, err := portal.ProveWithdrawalTransaction(
//		opts,
//		bindings.TypesWithdrawalTransaction{
//			Nonce:    params.Nonce,
//			Sender:   params.Sender,
//			Target:   params.Target,
//			Value:    params.Value,
//			GasLimit: params.GasLimit,
//			Data:     params.Data,
//		},
//		params.L2OutputIndex,
//		params.OutputRootProof,
//		params.WithdrawalProof,
//	)
//	require.Nil(t, err)
//
//	// Ensure that our withdrawal was proved successfully
//	proveReceipt, err := waitForTransaction(tx.Hash(), l1Client, 3*15*time.Second)
//	require.Nil(t, err, "prove withdrawal")
//	require.Equal(t, types.ReceiptStatusSuccessful, proveReceipt.Status)
//	return params, proveReceipt
//}
//
//func FinalizeWithdrawal(t *testing.T, cfg SystemConfig, l1Client *ethclient.Client, privKey *ecdsa.PrivateKey, withdrawalReceipt *types.Receipt, params withdrawals.ProvenWithdrawalParameters) *types.Receipt {
//	// Wait for finalization and then create the Finalized Withdrawal Transaction
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
//	defer cancel()
//	_, err := withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevOptimismPortalAddr, withdrawalReceipt.BlockNumber)
//	require.Nil(t, err)
//
//	opts, err := bind.NewKeyedTransactorWithChainID(privKey, cfg.L1ChainIDBig())
//	require.Nil(t, err)
//	portal, err := bindings.NewOptimismPortal(predeploys.DevOptimismPortalAddr, l1Client)
//	require.Nil(t, err)
//	// Finalize withdrawal
//	tx, err := portal.FinalizeWithdrawalTransaction(
//		opts,
//		bindings.TypesWithdrawalTransaction{
//			Nonce:    params.Nonce,
//			Sender:   params.Sender,
//			Target:   params.Target,
//			Value:    params.Value,
//			GasLimit: params.GasLimit,
//			Data:     params.Data,
//		},
//	)
//	require.Nil(t, err)
//
//	// Ensure that our withdrawal was finalized successfully
//	finalizeReceipt, err := waitForTransaction(tx.Hash(), l1Client, 3*time.Duration(cfg.DeployConfig.L1BlockTime)*time.Second)
//	require.Nil(t, err, "finalize withdrawal")
//	require.Equal(t, types.ReceiptStatusSuccessful, finalizeReceipt.Status)
//	return finalizeReceipt
//}
