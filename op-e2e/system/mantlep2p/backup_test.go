package mantlep2p

import (
	"context"
	"math/big"
	"slices"
	"testing"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/opnode"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	rollupNodeCfg "github.com/ethereum-optimism/optimism/op-node/config"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/endpoint"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

// TestSystemRPCAltSync tests that verifier can sync L2 blocks via RPC (backup sync)
// when P2P is disabled and batcher is disabled.
// This simulates the Mantle backup sync feature where verifier fetches blocks from sequencer RPC.
func TestSystemRPCAltSync(t *testing.T) {
	op_e2e.InitParallel(t)

	// Use Mantle Arsia config with Arsia activated at genesis
	arsiaTimeOffset := hexutil.Uint64(0)
	cfg := e2esys.MantleArsiaSystemConfigP2PGossip(t, &arsiaTimeOffset)
	// the default is nil, but this may change in the future.
	// This test must ensure the blocks are not synced via Gossip, but instead via the alt RPC based sync.
	cfg.P2PTopology = nil
	// Disable batcher, so there will not be any L1 data to sync from
	cfg.DisableBatcher = true

	var published, received []string
	seqTracer, verifTracer := new(opnode.FnTracer), new(opnode.FnTracer)
	// The sequencer still publishes the blocks to the tracer, even if they do not reach the network due to disabled P2P
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayloadEnvelope) {
		published = append(published, payload.ExecutionPayload.ID().String())
	}
	// Blocks are now received via the RPC based alt-sync method
	verifTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
		received = append(received, payload.ExecutionPayload.ID().String())
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["verifier"].Tracer = verifTracer

	// Configure verifier to use sequencer's RPC for L2Sync (backup sync)
	// This must be done via StartOption to access the sequencer's endpoint after it's started
	sys, err := cfg.StartMantle(t, e2esys.StartOption{
		Key:  "afterRollupNodeStart",
		Role: "sequencer",
		Action: func(sCfg *e2esys.SystemConfig, system *e2esys.System) {
			// Get sequencer's RPC endpoint and configure verifier to use it for backup sync
			seqEndpoint := system.NodeEndpoint("sequencer")
			rpcClient := endpoint.DialRPC(endpoint.PreferAnyRPC, seqEndpoint, func(v string) *rpc.Client {
				cl, err := rpc.Dial(v)
				require.NoError(t, err, "failed to dial sequencer RPC")
				return cl
			})
			// Use sCfg (not cfg) to modify the config that StartMantle uses internally
			sCfg.Nodes["verifier"].L2Sync = &rollupNodeCfg.PreparedL2SyncEndpoint{
				Client: client.NewBaseRPCClient(rpcClient),
				//TrustRPC: true,
			}
			// Enable SyncModeReqResp to trigger backup sync requests
			sCfg.Nodes["verifier"].Sync.SyncModeReqResp = true
		},
	})
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.NodeClient("sequencer")
	l2Verif := sys.NodeClient("verifier")

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit a TX to L2 sequencer node
	receiptSeq := helpers.SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *helpers.TxOpts) {
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
