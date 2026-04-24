package manual

import (
	"os"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
	"github.com/ethereum/go-ethereum"
)

func TestVerifierManualSync(gt *testing.T) {
	t := devtest.SerialT(gt)

	// On reth, engine_newPayload returns VALID before the block is committed to the
	// HTTP RPC DB (async pipeline), so immediate hash/number lookups may fail.
	// See TestVerifierManualSync_Reth for reth-specific coverage.
	if os.Getenv("DEVSTACK_L2EL_KIND") == "op-reth" {
		t.Skip("reth defers block DB commit after NewPayload VALID (async pipeline); " +
			"see TestVerifierManualSync_Reth for reth-specific coverage")
	}

	// Disable ELP2P and Batcher
	sys := presets.NewSingleChainMultiNodeWithoutCheck(t)
	require := t.Require()
	logger := t.Logger()

	delta := uint64(7)
	sys.L2CL.Advanced(types.LocalUnsafe, delta, 30)

	// Disable Derivation
	sys.L2CLB.Stop()

	startBlockNum := sys.L2ELB.BlockRefByLabel(eth.Unsafe).Number

	// Manual Block insertion using engine APIs
	for i := uint64(1); i <= delta; i++ {
		blockNum := startBlockNum + i
		block := sys.L2EL.BlockRefByNumber(blockNum)
		// Validator does not have canonical nor noncanonical block for blockNum
		_, err := sys.L2ELB.Escape().EthClient().BlockRefByNumber(t.Ctx(), blockNum)
		require.Error(err, ethereum.NotFound)
		_, err = sys.L2ELB.Escape().EthClient().BlockRefByHash(t.Ctx(), block.Hash)
		require.Error(err, ethereum.NotFound)

		// Insert payload
		logger.Info("NewPayload", "target", blockNum)
		sys.L2ELB.NewPayload(sys.L2EL, blockNum).IsValid()
		// Payload valid but not canonicalized. Cannot fetch block by number
		_, err = sys.L2ELB.Escape().EthClient().BlockRefByNumber(t.Ctx(), blockNum)
		require.Error(err, ethereum.NotFound)
		// Now fetchable by hash
		require.Equal(blockNum, sys.L2ELB.BlockRefByHash(block.Hash).Number)

		// FCU
		logger.Info("ForkchoiceUpdate", "target", blockNum)
		sys.L2ELB.ForkchoiceUpdate(sys.L2EL, blockNum, 0, 0, nil).IsValid()
		// Payload valid and canonicalized
		require.Equal(block.Hash, sys.L2ELB.BlockRefByNumber(blockNum).Hash)
		require.Equal(blockNum, sys.L2ELB.BlockRefByHash(block.Hash).Number)
	}

	// Check correctly synced by comparing with sequencer EL
	res := sys.L2ELB.BlockRefByLabel(eth.Unsafe)
	require.Equal(startBlockNum+delta, res.Number)
	require.Equal(sys.L2EL.BlockRefByNumber(startBlockNum+delta).Hash, res.Hash)
}

// TestVerifierManualSync_Reth is the reth-specific companion to TestVerifierManualSync.
//
// On reth, engine_newPayload returns VALID before the block is committed to the HTTP
// RPC DB (async pipeline). This test verifies that:
//   - After NewPayload returns VALID, the block eventually becomes accessible by hash.
//   - The block is not accessible by number until ForkchoiceUpdate canonicalizes it.
//   - After ForkchoiceUpdate, the block is fully canonical (accessible by both hash and number).
func TestVerifierManualSync_Reth(gt *testing.T) {
	t := devtest.SerialT(gt)

	if os.Getenv("DEVSTACK_L2EL_KIND") != "op-reth" {
		t.Skip("this test covers reth-specific async DB commit behavior after NewPayload VALID")
	}

	// Disable ELP2P and Batcher
	sys := presets.NewSingleChainMultiNodeWithoutCheck(t)
	require := t.Require()
	logger := t.Logger()
	ctx := t.Ctx()

	delta := uint64(7)
	attempts := 10
	sys.L2CL.Advanced(types.LocalUnsafe, delta, 30)

	// Disable Derivation
	sys.L2CLB.Stop()

	startBlockNum := sys.L2ELB.BlockRefByLabel(eth.Unsafe).Number

	// Manual Block insertion using engine APIs
	for i := uint64(1); i <= delta; i++ {
		blockNum := startBlockNum + i
		block := sys.L2EL.BlockRefByNumber(blockNum)
		// Validator does not have canonical nor noncanonical block for blockNum
		_, err := sys.L2ELB.Escape().EthClient().BlockRefByNumber(ctx, blockNum)
		require.Error(err, ethereum.NotFound)
		_, err = sys.L2ELB.Escape().EthClient().BlockRefByHash(ctx, block.Hash)
		require.Error(err, ethereum.NotFound)

		// Insert payload
		logger.Info("NewPayload", "target", blockNum)
		sys.L2ELB.NewPayload(sys.L2EL, blockNum).IsValid()

		// On reth, NewPayload VALID only validates the block in the engine tree.
		// The block is NOT committed to the HTTP RPC DB until ForkchoiceUpdate
		// promotes it — unlike geth which stores it as non-canonical immediately.
		// Verify this: block must NOT be accessible by hash or number yet.
		_, errHash := sys.L2ELB.Escape().EthClient().BlockRefByHash(ctx, block.Hash)
		require.Error(errHash, "reth should not expose block by hash before FCU")
		_, errNum := sys.L2ELB.Escape().EthClient().BlockRefByNumber(ctx, blockNum)
		require.Error(errNum, "reth should not expose block by number before FCU")

		// ForkchoiceUpdate to canonicalize and commit to RPC DB.
		logger.Info("ForkchoiceUpdate", "target", blockNum)
		sys.L2ELB.ForkchoiceUpdate(sys.L2EL, blockNum, 0, 0, nil).WaitUntilValid(attempts)

		// After FCU, block is committed and accessible by both hash and number.
		require.Eventually(func() bool {
			ref, hashErr := sys.L2ELB.Escape().EthClient().BlockRefByHash(ctx, block.Hash)
			return hashErr == nil && ref.Number == blockNum
		}, 10*time.Second, 200*time.Millisecond,
			"block should be accessible by hash after ForkchoiceUpdate")
		require.Equal(block.Hash, sys.L2ELB.BlockRefByNumber(blockNum).Hash)
	}

	// Check correctly synced by comparing with sequencer EL
	res := sys.L2ELB.BlockRefByLabel(eth.Unsafe)
	require.Equal(startBlockNum+delta, res.Number)
	require.Equal(sys.L2EL.BlockRefByNumber(startBlockNum+delta).Hash, res.Hash)
}
