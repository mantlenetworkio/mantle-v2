package opreth

// Engine-API tests for op-reth.
//
// These tests mirror the engine-API focused subset of opgeth/op_geth_test.go
// and run against an external op-reth subprocess.  Tests that require
// geth-specific features (txpool RPC, pending-block behaviour, etc.) are
// omitted here; run the opgeth suite for those.
//
// Scope note:
//   op-reth supports the OP Stack from the Skadi fork onwards.
//   Pre-Skadi fork tests (Regolith, Canyon, Ecotone, Fjord, Isthmus, …)
//   are not applicable to reth and are skipped here.  The corresponding
//   tests in opgeth remain the authoritative coverage for those forks.
//
// Required env vars:
//
//	OP_E2E_L2_BIN=/path/to/op-reth   – path to the op-reth binary
//
// Optional:
//
//	OP_E2E_L2_READY_TIMEOUT=60s      – how long to wait for reth to be ready

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// TestMissingGasLimit verifies that op-reth cannot build a block without a
// gas limit while OP Stack is active.
func TestMissingGasLimit(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opReth, err := NewRethEngine(t, ctx, &cfg)
	require.NoError(t, err)
	defer opReth.Close()

	attrs, err := opReth.CreatePayloadAttributes()
	require.NoError(t, err)
	// Remove the GasLimit from the otherwise valid attributes.
	attrs.GasLimit = nil

	res, err := opReth.StartBlockBuilding(ctx, attrs)
	require.Error(t, err)
	// reth returns standard JSON-RPC -32602 (Invalid params) for missing gasLimit;
	// geth returns the Engine API specific -38003 (InvalidPayloadAttributes).
	// Both are correct rejections — verify the call was rejected and returned no payload ID.
	var rpcErr rpc.Error
	require.ErrorAs(t, err, &rpcErr, "error should be an RPC error")
	code := rpcErr.ErrorCode()
	require.True(t,
		code == int(eth.InvalidPayloadAttributes) || code == -32602,
		"expected InvalidPayloadAttributes (-38003) or Invalid params (-32602), got: %d", code,
	)
	require.Nil(t, res)
}

// TestInvalidDepositInFCU runs an invalid deposit through a
// FCU/GetPayload/NewPayload/FCU set of calls and asserts the block is still
// built (deposits must never prevent block building).
func TestInvalidDepositInFCU(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opReth, err := NewRethEngine(t, ctx, &cfg)
	require.NoError(t, err)
	defer opReth.Close()

	// Create a deposit from a new account that will always fail (not enough funds).
	fromKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	fromAddr := crypto.PubkeyToAddress(fromKey.PublicKey)
	balance, err := opReth.L2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)
	require.Equal(t, 0, balance.Cmp(common.Big0))

	badDepositTx := types.NewTx(&types.DepositTx{
		From:                fromAddr,
		To:                  &fromAddr, // send to ourselves
		Value:               big.NewInt(params.Ether),
		Gas:                 25000,
		IsSystemTransaction: false,
	})

	// The invalid deposit should still be included in the block.
	_, err = opReth.AddL2Block(ctx, badDepositTx)
	require.NoError(t, err)

	// Deposit tx was included, but the account should still have no ETH.
	balance, err = opReth.L2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)
	require.Equal(t, 0, balance.Cmp(common.Big0))
}

// TestPreregolith is skipped for op-reth.
//
// op-reth supports the OP Stack from the Skadi fork onwards and does
// not implement pre-Regolith gas-accounting semantics.  The authoritative
// coverage for pre-Regolith behaviour lives in opgeth.
func TestPreregolith(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"pre-Regolith behaviour is covered by opgeth")
}

// TestRegolith is skipped for op-reth.
//
// op-reth supports the OP Stack from the Skadi fork onwards and does
// not implement Regolith-specific gas-accounting semantics in isolation.
// The authoritative coverage for Regolith behaviour lives in opgeth.
func TestRegolith(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"Regolith behaviour is covered by opgeth")
}

// TestGethOnlyPendingBlockIsLatest is skipped for op-reth.
//
// This test exercises geth-specific pending-block semantics.  op-reth's
// public RPC does not support "pending" block tag in the same way.
// Covered by opgeth.
func TestGethOnlyPendingBlockIsLatest(t *testing.T) {
	t.Skip("geth-specific: pending block semantics differ in reth; covered by opgeth")
}

// TestPreCanyon is skipped for op-reth.
//
// configureMantleForks sets SkadiTimeOffset to the fork time (nil or future),
// creating a pre-Skadi genesis.  op-reth always activates Skadi at genesis.
// Covered by opgeth.
func TestPreCanyon(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"pre-Canyon behaviour is covered by opgeth")
}

// TestCanyon is skipped for op-reth.
//
// CanyonSystemConfig does not schedule Skadi, leaving SkadiTime unset.
// op-reth always activates Skadi at genesis, so this canyon-only
// configuration is not reachable.  Covered by opgeth.
func TestCanyon(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"Canyon behaviour is covered by opgeth")
}

// TestPreEcotone is skipped for op-reth.
//
// configureMantleForks sets SkadiTimeOffset to the fork time (nil or future),
// creating a pre-Skadi genesis.  op-reth always activates Skadi at genesis.
// Covered by opgeth.
func TestPreEcotone(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"pre-Ecotone behaviour is covered by opgeth")
}

// TestEcotone is skipped for op-reth.
//
// EcotoneSystemConfig does not schedule Skadi, leaving SkadiTime unset.
// op-reth always activates Skadi at genesis, so this ecotone-only
// configuration is not reachable.  Covered by opgeth.
func TestEcotone(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"Ecotone behaviour is covered by opgeth")
}

// TestPreFjord is skipped for op-reth.
//
// FjordNotScheduled sets SkadiTime=nil; FjordNotYetActive sets it to a
// future time — both produce a pre-Skadi genesis.  op-reth always activates
// Skadi at genesis.  Covered by opgeth.
func TestPreFjord(t *testing.T) {
	t.Skip("pre-Skadi: op-reth only supports forks from Skadi onwards; " +
		"pre-Fjord behaviour is covered by opgeth")
}
