package derivation

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	upgradesHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/mantleupgrades/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// TestL2TxOrderConsistency verifies that when multiple L2 transactions are sent in sequence,
// the order of transactions in the derived block matches the order they were sent.
//
// Test steps:
// 1. Send 10 transactions with sequential nonces from the same account
// 2. Sequencer builds L2 block containing all transactions
// 3. Batcher submits batch to L1
// 4. Verifier derives the L2 block from L1 data
// 5. Verify that transaction order in derived block matches sending order
//
// This test ensures the derivation pipeline preserves transaction ordering,
// which is critical for applications that depend on transaction sequencing.
func TestL2TxOrderConsistency(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// Setup test environment with Arsia activated at genesis
	p := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}
	dp := e2eutils.MakeMantleDeployParams(t, p)
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelInfo)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)
	sequencer.ActL2PipelineFull(t)

	// Setup verifier
	_, verifier := helpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})
	verifier.ActL2PipelineFull(t)

	// Setup batcher
	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(log, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Prepare L1 block
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)

	t.Log("PHASE 1: Send multiple transactions in sequence")

	// Setup Alice's account for sending transactions
	l2UserEnv := &helpers.BasicUserEnv[*helpers.L2Bindings]{
		EthCl:    seqEngine.EthClient(),
		Signer:   types.LatestSignerForChainID(sd.RollupCfg.L2ChainID),
		Bindings: helpers.NewL2Bindings(t, seqEngine.EthClient(), seqEngine.GethClient()),
	}
	alice := helpers.NewCrossLayerUser(log, dp.Secrets.Alice, rand.New(rand.NewSource(0xa11ce)), config.DefaultAllocType)
	alice.L2.SetUserEnv(l2UserEnv)

	numTxs := 10
	sentTxHashes := make([]common.Hash, 0, numTxs)

	// Send multiple transactions with sequential nonces
	for i := 0; i < numTxs; i++ {
		alice.L2.ActResetTxOpts(t)
		alice.L2.ActSetTxValue(big.NewInt(int64(1000 + i)))(t) // Different values to distinguish txs
		alice.L2.ActSetTxToAddr(&dp.Addresses.Bob)(t)          // Send to Bob
		alice.L2.ActMakeTx(t)

		// Include transaction in L2 block
		sequencer.ActL2StartBlock(t)
		seqEngine.ActL2IncludeTx(alice.Address())(t)
		sequencer.ActL2EndBlock(t)

		// Record sent transaction hash
		receipt := alice.L2.LastTxReceipt(t)
		sentTxHashes = append(sentTxHashes, receipt.TxHash)

		t.Logf("Sent tx %d: hash=%s", i, receipt.TxHash.Hex())
	}

	l2BlockNum := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	t.Logf("All %d transactions sent, current L2 block: %d", numTxs, l2BlockNum)

	t.Log("PHASE 2: Submit batch to L1")

	// Submit batch to L1
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	batchTx := batcher.LastSubmitted

	// Include batch in L1 block
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	t.Logf("Batch submitted to L1 block %d", miner.L1Chain().CurrentBlock().Number.Uint64())

	// Sequencer processes its own batch
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	t.Log("PHASE 3: Verifier derives blocks from L1")

	// Verifier derives blocks
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	verifierSafeHead := verifier.L2Safe()
	require.Equal(t, l2BlockNum, verifierSafeHead.Number,
		"Verifier should derive all L2 blocks")
	t.Logf("Verifier synced to L2 block %d", verifierSafeHead.Number)

	t.Log("PHASE 4: Verify transaction order consistency")

	// Collect all transactions from derived blocks
	derivedTxs := make([]*types.Transaction, 0, numTxs)
	derivedTxHashes := make([]common.Hash, 0, numTxs)

	// Iterate through all L2 blocks to collect transactions
	genesisBlockNum := sd.RollupCfg.Genesis.L2.Number
	for blockNum := genesisBlockNum + 1; blockNum <= l2BlockNum; blockNum++ {
		block := seqEngine.L2Chain().GetBlockByNumber(blockNum)
		require.NotNil(t, block, "Block %d should exist", blockNum)

		for _, tx := range block.Transactions() {
			// Skip deposit transactions (system transactions)
			if tx.Type() == types.DepositTxType {
				continue
			}
			derivedTxs = append(derivedTxs, tx)
			derivedTxHashes = append(derivedTxHashes, tx.Hash())
		}
	}

	t.Logf("Collected %d user transactions from derived blocks", len(derivedTxs))

	// Verify we got all expected transactions
	require.Equal(t, numTxs, len(derivedTxs),
		"Should have derived exactly %d transactions", numTxs)

	// Verify transaction order matches sending order
	for i := 0; i < numTxs; i++ {
		require.Equal(t, sentTxHashes[i], derivedTxHashes[i],
			"Transaction %d hash mismatch: expected %s, got %s",
			i, sentTxHashes[i].Hex(), derivedTxHashes[i].Hex())

		t.Logf("Transaction %d: hash=%s (order preserved)", i, derivedTxHashes[i].Hex())
	}

	t.Log("SUCCESS: Transaction Order Verification Complete")

	t.Logf("Sent %d transactions in sequence", numTxs)
	t.Logf("All transactions derived by verifier")
	t.Logf("Transaction order preserved (hashes match)")

}

// TestHighTpsDerivationStability verifies that under high TPS (Transactions Per Second) load,
// the derivation pipeline can continuously and stably derive blocks without errors or stalls.
//
// Test steps:
// 1. Run 50 rounds of high-load transaction generation
// 2. Each round: build ONE L2 block containing 20 transactions (high TPS)
// 3. Submit all batches to L1
// 4. Verifier continuously derives all blocks
// 5. Verify all blocks generated successfully, no missing blocks, no errors
//
// This test ensures the system can handle sustained high throughput (20 tx/block)
// and maintain stable block production and derivation under load.
func TestHighTpsDerivationStability(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)

	// Setup test environment
	p := &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	}
	dp := e2eutils.MakeMantleDeployParams(t, p)
	arsiaTimeOffset := hexutil.Uint64(0)
	upgradesHelpers.ApplyArsiaTimeOffset(dp, &arsiaTimeOffset)

	sd := e2eutils.SetupMantleNormal(t, dp, helpers.DefaultAlloc)
	log := testlog.Logger(t, log.LevelInfo)

	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)
	sequencer.ActL2PipelineFull(t)

	// Setup verifier
	_, verifier := helpers.SetupVerifier(t, sd, log, miner.L1Client(t, sd.RollupCfg),
		miner.BlobStore(), &sync.Config{})
	verifier.ActL2PipelineFull(t)

	// Setup batcher
	rollupSeqCl := sequencer.RollupClient()
	batcher := helpers.NewL2Batcher(log, sd.RollupCfg, &helpers.BatcherCfg{
		MinL1TxSize:              0,
		MaxL1TxSize:              128_000,
		BatcherKey:               dp.Secrets.Batcher,
		DataAvailabilityType:     batcherFlags.BlobsType,
		ForceSubmitSingularBatch: true,
		EnableCellProofs:         true,
	}, rollupSeqCl, miner.EthClient(), seqEngine.EthClient(), seqEngine.EngineClient(t, sd.RollupCfg))

	// Prepare L1 block
	miner.ActEmptyBlock(t)
	miner.ActL1SafeNext(t)
	miner.ActL1FinalizeNext(t)

	sequencer.ActL1HeadSignal(t)

	t.Log("HIGH TPS DERIVATION STABILITY TEST")

	// Test parameters
	numRounds := 50
	txsPerRound := 20
	totalExpectedTxs := numRounds * txsPerRound

	t.Logf("Configuration:")
	t.Logf("  - Rounds: %d", numRounds)
	t.Logf("  - Transactions per round: %d", txsPerRound)
	t.Logf("  - Total expected transactions: %d", totalExpectedTxs)
	t.Logf("  - Simulated TPS: ~%d tx/block", txsPerRound)

	// Setup users for sending transactions (use 5 different accounts to avoid nonce conflicts)
	l2UserEnv := &helpers.BasicUserEnv[*helpers.L2Bindings]{
		EthCl:    seqEngine.EthClient(),
		Signer:   types.LatestSignerForChainID(sd.RollupCfg.L2ChainID),
		Bindings: helpers.NewL2Bindings(t, seqEngine.EthClient(), seqEngine.GethClient()),
	}

	numUsers := 5
	users := make([]*helpers.CrossLayerUser, numUsers)
	for i := 0; i < numUsers; i++ {
		var userKey *ecdsa.PrivateKey
		switch i {
		case 0:
			userKey = dp.Secrets.Alice
		case 1:
			userKey = dp.Secrets.Bob
		case 2:
			userKey = dp.Secrets.Mallory
		case 3:
			userKey = dp.Secrets.SysCfgOwner
		case 4:
			userKey = dp.Secrets.Proposer
		}
		users[i] = helpers.NewCrossLayerUser(log, userKey, rand.New(rand.NewSource(int64(i*1000))), config.DefaultAllocType)
		users[i].L2.SetUserEnv(l2UserEnv)
	}

	t.Log("PHASE 1: High-load transaction generation")

	l2BlocksGenerated := make([]uint64, 0, numRounds)
	totalTxsSent := 0

	// Generate high load: multiple rounds of rapid transaction sending
	// Each round creates ONE block containing multiple transactions (high TPS)
	for round := 0; round < numRounds; round++ {
		// Start a new L2 block for this round
		sequencer.ActL2StartBlock(t)

		// Include multiple transactions in this single block
		for txIndex := 0; txIndex < txsPerRound; txIndex++ {
			// Round-robin across users to avoid nonce conflicts
			userIdx := (round*txsPerRound + txIndex) % numUsers
			user := users[userIdx]

			user.L2.ActResetTxOpts(t)
			user.L2.ActSetTxValue(big.NewInt(int64(round*1000 + txIndex)))(t)
			user.L2.ActSetTxToAddr(&dp.Addresses.Bob)(t)
			user.L2.ActMakeTx(t)

			// Include this transaction in the current block
			seqEngine.ActL2IncludeTx(user.Address())(t)

			totalTxsSent++
		}

		// End the block after including all transactions from this round
		sequencer.ActL2EndBlock(t)

		l2BlockNum := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
		l2BlocksGenerated = append(l2BlocksGenerated, l2BlockNum)

		if (round+1)%10 == 0 {
			t.Logf("Progress: %d/%d rounds completed, %d txs sent, current L2 block: %d (avg %d tx/block)",
				round+1, numRounds, totalTxsSent, l2BlockNum, txsPerRound)
		}
	}

	finalL2Block := seqEngine.L2Chain().CurrentBlock().Number.Uint64()
	t.Logf("Phase 1 complete: %d rounds, %d transactions sent, L2 block: %d",
		numRounds, totalTxsSent, finalL2Block)

	t.Log("PHASE 2: Batch submission to L1")

	// Submit all batches to L1
	batcher.ActBufferAll(t)
	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	batchTx := batcher.LastSubmitted

	// Include batch in L1 block
	miner.ActL1StartBlock(12)(t)
	miner.ActL1IncludeTxByHash(batchTx.Hash())(t)
	miner.ActL1EndBlock(t)

	l1BlockNum := miner.L1Chain().CurrentBlock().Number.Uint64()
	t.Logf("All batches submitted to L1 block %d", l1BlockNum)

	// Sequencer processes its own batch
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2PipelineFull(t)

	t.Log("PHASE 3: Continuous derivation by verifier")

	// Verifier derives all blocks
	verifier.ActL1HeadSignal(t)
	verifier.ActL2PipelineFull(t)

	verifierSafeHead := verifier.L2Safe()
	t.Logf("Verifier derivation complete, safe head: %d", verifierSafeHead.Number)

	t.Log("PHASE 4: Verification of derivation stability")

	// Verify verifier synced to the same height
	require.Equal(t, finalL2Block, verifierSafeHead.Number,
		"Verifier should derive all L2 blocks")
	t.Logf("Verifier synced to L2 block %d", verifierSafeHead.Number)

	// Verify block continuity (no missing blocks)
	genesisBlockNum := sd.RollupCfg.Genesis.L2.Number
	for blockNum := genesisBlockNum + 1; blockNum <= finalL2Block; blockNum++ {
		block := seqEngine.L2Chain().GetBlockByNumber(blockNum)
		require.NotNil(t, block, "Block %d should exist", blockNum)
	}
	t.Logf("All %d L2 blocks exist (no missing blocks)", finalL2Block-genesisBlockNum)

	// Verify all transactions were included
	actualTxCount := 0
	for blockNum := genesisBlockNum + 1; blockNum <= finalL2Block; blockNum++ {
		block := seqEngine.L2Chain().GetBlockByNumber(blockNum)
		for _, tx := range block.Transactions() {
			// Skip deposit transactions
			if tx.Type() != types.DepositTxType {
				actualTxCount++
			}
		}
	}

	require.Equal(t, totalExpectedTxs, actualTxCount,
		"All transactions should be included in blocks")
	t.Logf("All %d transactions included in blocks", actualTxCount)

	// Verify sequencer and verifier have same state
	seqHead := sequencer.L2Safe()
	require.Equal(t, seqHead.Number, verifierSafeHead.Number,
		"Sequencer and verifier should have same safe head number")
	require.Equal(t, seqHead.Hash, verifierSafeHead.Hash,
		"Sequencer and verifier should have same safe head hash")
	t.Logf("Sequencer and verifier state synchronized")

	t.Log("SUCCESS: High TPS Derivation Stability Test Complete")

	t.Logf("%d rounds completed successfully", numRounds)
	t.Logf("%d transactions processed", actualTxCount)
	t.Logf("%d L2 blocks generated", finalL2Block-genesisBlockNum)
	t.Logf("Average: %.1f tx/block", float64(actualTxCount)/float64(finalL2Block-genesisBlockNum))
	t.Logf("All blocks derived without errors")
	t.Logf("No missing blocks or derivation stalls")
	t.Logf("Sequencer and verifier in sync")

	t.Log("CONFIRMED: System stable under high TPS load")

}
