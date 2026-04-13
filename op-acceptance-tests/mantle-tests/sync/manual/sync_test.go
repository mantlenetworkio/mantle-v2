package manual

import (
	"os"
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
	"github.com/ethereum/go-ethereum"
)

func TestVerifierManualSync(gt *testing.T) {
	t := devtest.SerialT(gt)

	// Disable ELP2P and Batcher
	sys := presets.NewSingleChainMultiNodeWithoutCheck(t)
	require := t.Require()
	logger := t.Logger()

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
		_, err := sys.L2ELB.Escape().EthClient().BlockRefByNumber(t.Ctx(), blockNum)
		require.Error(err, ethereum.NotFound)
		_, err = sys.L2ELB.Escape().EthClient().BlockRefByHash(t.Ctx(), block.Hash)
		require.Error(err, ethereum.NotFound)

		// Insert payload
		logger.Info("NewPayload", "target", blockNum)
		sys.L2ELB.NewPayload(sys.L2EL, blockNum).IsValid()

		// On geth: after NewPayload returns VALID, the block is stored as non-canonical and
		// immediately accessible by hash (but not yet by number). On reth, engine_newPayload
		// returns VALID before the block is committed to the HTTP RPC DB (async pipeline), so
		// neither hash nor number lookup is guaranteed to succeed yet. Skip on reth.
		if os.Getenv("DEVSTACK_L2EL_KIND") != "op-reth" {
			require.Equal(blockNum, sys.L2ELB.BlockRefByHash(block.Hash).Number)
			_, errNC := sys.L2ELB.Escape().EthClient().BlockRefByNumber(t.Ctx(), blockNum)
			require.Error(errNC, ethereum.NotFound)
		}

		// On reth, engine_newPayload returns VALID before the block is committed
		// to the DB, so checks between NewPayload and ForkchoiceUpdate may race.
		// Go straight to FCU.
		logger.Info("ForkchoiceUpdate", "target", blockNum)
		sys.L2ELB.ForkchoiceUpdate(sys.L2EL, blockNum, 0, 0, nil).WaitUntilValid(attempts)
		// Payload valid and canonicalized
		require.Equal(block.Hash, sys.L2ELB.BlockRefByNumber(blockNum).Hash)
		require.Equal(blockNum, sys.L2ELB.BlockRefByHash(block.Hash).Number)
	}

	// Check correctly synced by comparing with sequencer EL
	res := sys.L2ELB.BlockRefByLabel(eth.Unsafe)
	require.Equal(startBlockNum+delta, res.Number)
	require.Equal(sys.L2EL.BlockRefByNumber(startBlockNum+delta).Hash, res.Hash)
}
