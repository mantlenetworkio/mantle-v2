package oracle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

// DeployContractBackend represents the union of the
// DeployBackend and the ContractBackend
type DeployContractBackend interface {
	bind.DeployBackend
	bind.ContractBackend
}

// Only update the gas price when it must be changed by at least
// a paramaterizable amount. If the param is greater than the result
// of 1 - (min/max) where min and max are the gas prices then do not
// update the gas price
func isDifferenceSignificant(a, b uint64, c float64) bool {
	max := max(a, b)
	min := min(a, b)
	factor := 1 - (float64(min) / float64(max))
	return c <= factor
}

// Wait for the receipt by polling the backend
func waitForReceiptWithMaxRetries(backend DeployContractBackend, tx *types.Transaction, maxRetries int) (*types.Receipt, error) {
	t := time.NewTicker(300 * time.Millisecond)
	receipt := new(types.Receipt)
	var err error
	for i := 0; i < maxRetries; i++ {
		select {
		case <-t.C:
			receipt, err = backend.TransactionReceipt(context.Background(), tx.Hash())
			if errors.Is(err, ethereum.NotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if receipt != nil {
				t.Stop()
				return receipt, nil
			}
		}
	}
	return nil, fmt.Errorf("transaction receipt not found after %d retries", maxRetries)
}

func max(a, b uint64) uint64 {
	if a >= b {
		return a
	}
	return b
}

func min(a, b uint64) uint64 {
	if a >= b {
		return b
	}
	return a
}
