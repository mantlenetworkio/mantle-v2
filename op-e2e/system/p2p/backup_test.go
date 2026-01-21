package p2p

import (
	"context"
	"math/big"
	"slices"
	"testing"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestSystemRPCAltSync(t *testing.T) {
	op_e2e.InitParallel(t)

	cfg := DefaultSystemConfig(t)
	// the default is nil, but this may change in the future.
	// This test must ensure the blocks are not synced via Gossip, but instead via the alt RPC based sync.
	cfg.P2PTopology = nil
	// Disable batcher, so there will not be any L1 data to sync from
	cfg.DisableBatcher = true

	var published, received []string
	seqTracer, verifTracer := new(FnTracer), new(FnTracer)
	// The sequencer still publishes the blocks to the tracer, even if they do not reach the network due to disabled P2P
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayload) {
		published = append(published, payload.ID().String())
	}
	// Blocks are now received via the RPC based alt-sync method
	verifTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayload) {
		received = append(received, payload.ID().String())
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["verifier"].Tracer = verifTracer

	sys, err := cfg.Start(t, SystemConfigOption{
		key:  "afterRollupNodeStart",
		role: "sequencer",
		action: func(sCfg *SystemConfig, system *System) {
			cfg.Nodes["verifier"].L2Sync = &rollupNode.PreparedL2SyncEndpoint{
				Client: client.NewBaseRPCClient(system.RawClients["sequencer"]),
			}
		},
	})
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit a TX to L2 sequencer node
	receiptSeq := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.Value = big.NewInt(1_000_000_000)

		// Wait for alt RPC sync to pick up the blocks on the sequencer chain
		opts.VerifyOnClients(l2Verif)
	})

	// Sometimes we get duplicate blocks on the sequencer which makes this test flaky
	published = slices.Compact(published)
	received = slices.Compact(received)

	// Verify that the tx was received via RPC sync (P2P is disabled)
	require.Contains(t, received, eth.BlockID{Hash: receiptSeq.BlockHash, Number: receiptSeq.BlockNumber.Uint64()}.String())

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received))
	require.ElementsMatch(t, received, published[:len(received)])
}
