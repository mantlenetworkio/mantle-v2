package p2p

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-node/testlog"
)

func TestGuardGossipValidator(t *testing.T) {
	logger := testlog.Logger(t, log.LevelCrit)
	val := guardGossipValidator(logger, func(ctx context.Context, id peer.ID, message *pubsub.Message) pubsub.ValidationResult {
		if id == "mallory" {
			panic("mallory was here")
		}
		if id == "bob" {
			return pubsub.ValidationIgnore
		}
		return pubsub.ValidationAccept
	})
	// Test that panics from mallory are recovered and rejected,
	// and test that we can continue to ignore bob and accept alice.
	require.Equal(t, pubsub.ValidationAccept, val(context.Background(), "alice", nil))
	require.Equal(t, pubsub.ValidationReject, val(context.Background(), "mallory", nil))
	require.Equal(t, pubsub.ValidationIgnore, val(context.Background(), "bob", nil))
	require.Equal(t, pubsub.ValidationReject, val(context.Background(), "mallory", nil))
	require.Equal(t, pubsub.ValidationAccept, val(context.Background(), "alice", nil))
	require.Equal(t, pubsub.ValidationIgnore, val(context.Background(), "bob", nil))
}

func TestVerifyBlockSignature(t *testing.T) {
	logger := testlog.Logger(t, log.LevelCrit)
	cfg := &rollup.Config{
		L2ChainID: big.NewInt(100),
	}
	peerId := peer.ID("foo")
	secrets, err := e2eutils.DefaultMnemonicConfig.Secrets()
	require.NoError(t, err)
	msg := []byte("any msg")

	t.Run("Valid", func(t *testing.T) {
		runCfg := &testutils.MockRuntimeConfig{P2PSeqAddress: crypto.PubkeyToAddress(secrets.SequencerP2P.PublicKey)}
		signer := &PreparedSigner{Signer: NewLocalSigner(secrets.SequencerP2P)}
		sig, err := signer.Sign(context.Background(), SigningDomainBlocksV1, cfg.L2ChainID, msg)
		require.NoError(t, err)
		result := verifyBlockSignature(logger, cfg, runCfg, peerId, sig[:65], msg)
		require.Equal(t, pubsub.ValidationAccept, result)
	})

	t.Run("WrongSigner", func(t *testing.T) {
		runCfg := &testutils.MockRuntimeConfig{P2PSeqAddress: common.HexToAddress("0x1234")}
		signer := &PreparedSigner{Signer: NewLocalSigner(secrets.SequencerP2P)}
		sig, err := signer.Sign(context.Background(), SigningDomainBlocksV1, cfg.L2ChainID, msg)
		require.NoError(t, err)
		result := verifyBlockSignature(logger, cfg, runCfg, peerId, sig[:65], msg)
		require.Equal(t, pubsub.ValidationReject, result)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		runCfg := &testutils.MockRuntimeConfig{P2PSeqAddress: crypto.PubkeyToAddress(secrets.SequencerP2P.PublicKey)}
		sig := make([]byte, 65)
		result := verifyBlockSignature(logger, cfg, runCfg, peerId, sig, msg)
		require.Equal(t, pubsub.ValidationReject, result)
	})

	t.Run("NoSequencer", func(t *testing.T) {
		runCfg := &testutils.MockRuntimeConfig{}
		signer := &PreparedSigner{Signer: NewLocalSigner(secrets.SequencerP2P)}
		sig, err := signer.Sign(context.Background(), SigningDomainBlocksV1, cfg.L2ChainID, msg)
		require.NoError(t, err)
		result := verifyBlockSignature(logger, cfg, runCfg, peerId, sig[:65], msg)
		require.Equal(t, pubsub.ValidationIgnore, result)
	})
}
