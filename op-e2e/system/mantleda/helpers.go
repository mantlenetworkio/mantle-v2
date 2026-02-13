package mantleda

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources"
)

// sendTx sends a transaction with random calldata of the specified size.
// In Mantle Arsia fork, DA footprint gas consumption must be considered.
func sendTx(t *testing.T, senderKey *ecdsa.PrivateKey, nonce uint64, size int, chainID *big.Int, cl *ethclient.Client) common.Hash {
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	require.NoError(t, err, "failed to generate random data")

	// In Arsia fork, DA footprint requires more gas
	// Formula: base gas + calldata gas + DA footprint gas
	// DA footprint ≈ size * 400 (Arsia scalar)
	// For safety, we use a larger gas limit
	gasLimit := uint64(21_000 + len(randomBytes)*16 + len(randomBytes)*500)

	tx := types.MustSignNewTx(senderKey, types.LatestSignerForChainID(chainID), &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &common.Address{0xff, 0xff},
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       gasLimit,
		Data:      randomBytes,
	})
	err = cl.SendTransaction(context.Background(), tx)
	require.NoError(t, err, "failed to send transaction")
	return tx.Hash()
}

// waitForReceipt waits for a transaction receipt.
func waitForReceipt(t *testing.T, hash common.Hash, cl *ethclient.Client) *types.Receipt {
	receipt, err := wait.ForReceiptOK(context.Background(), cl, hash)
	require.NoError(t, err, "failed to wait for receipt")
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "transaction should succeed")
	return receipt
}

// waitForSafeBlock waits for the specified block to become safe on the rollup node.
// This is more reliable than waiting for verifier head, especially when batcher just started.
// Note: wait.ForSafeBlock has an internal 60s timeout, but when batcher just started,
// it may need more time to submit the first batch to L1. We use a custom implementation
// with longer timeout to handle this case.
func waitForSafeBlock(t *testing.T, blockNumber *big.Int, rc *sources.RollupClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Custom implementation with longer timeout
	_, err := wait.AndGet(ctx, time.Second, func() (*eth.SyncStatus, error) {
		return rc.SyncStatus(ctx)
	}, func(syncStatus *eth.SyncStatus) bool {
		return syncStatus.SafeL2.Number >= blockNumber.Uint64()
	})
	require.NoError(t, err, "failed to wait for safe block")
	require.NoError(t, wait.ForProcessingFullBatch(context.Background(), rc), "failed to wait for batch processing")
}

// sendAndWaitForReceipt sends a transaction and waits for its receipt.
func sendAndWaitForReceipt(t *testing.T, senderKey *ecdsa.PrivateKey, nonce uint64, size int, chainID *big.Int, cl *ethclient.Client) *types.Receipt {
	hash := sendTx(t, senderKey, nonce, size, chainID, cl)
	return waitForReceipt(t, hash, cl)
}
