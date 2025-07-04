package op_e2e

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/client"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	rollupNode "github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum-optimism/optimism/op-node/p2p"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/rollup/driver"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum-optimism/optimism/op-service/backoff"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oppprof "github.com/ethereum-optimism/optimism/op-service/pprof"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

func TestL2OutputSubmitter(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	cfg.NonFinalizedProposals = true // speed up the time till we see output proposals

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]

	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.Nil(t, err)
	rollupClient := sources.NewRollupClient(client.NewBaseRPCClient(rollupRPCClient))

	//  OutputOracle is already deployed
	l2OutputOracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
	require.Nil(t, err)

	initialOutputBlockNumber, err := l2OutputOracle.LatestBlockNumber(&bind.CallOpts{})
	require.Nil(t, err)

	// Wait until the second output submission from L2. The output submitter submits outputs from the
	// unsafe portion of the chain which gets reorged on startup. The sequencer has an out of date view
	// when it creates it's first block and uses and old L1 Origin. It then does not submit a batch
	// for that block and subsequently reorgs to match what the verifier derives when running the
	// reconcillation process.
	l2Verif := sys.Clients["verifier"]
	_, err = waitForBlock(big.NewInt(6), l2Verif, 10*time.Duration(cfg.DeployConfig.L2BlockTime)*time.Second)
	require.Nil(t, err)

	// Wait for batch submitter to update L2 output oracle.
	timeoutCh := time.After(15 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		l2ooBlockNumber, err := l2OutputOracle.LatestBlockNumber(&bind.CallOpts{})
		require.Nil(t, err)

		// Wait for the L2 output oracle to have been changed from the initial
		// timestamp set in the contract constructor.
		if l2ooBlockNumber.Cmp(initialOutputBlockNumber) > 0 {
			// Retrieve the l2 output committed at this updated timestamp.
			committedL2Output, err := l2OutputOracle.GetL2OutputAfter(&bind.CallOpts{}, l2ooBlockNumber)
			require.NotEqual(t, [32]byte{}, committedL2Output.OutputRoot, "Empty L2 Output")
			require.Nil(t, err)

			// Fetch the corresponding L2 block and assert the committed L2
			// output matches the block's state root.
			//
			// NOTE: This assertion will change once the L2 output format is
			// finalized.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			l2Output, err := rollupClient.OutputAtBlock(ctx, l2ooBlockNumber.Uint64())
			require.Nil(t, err)
			require.Equal(t, l2Output.OutputRoot[:], committedL2Output.OutputRoot[:])
			break
		}

		select {
		case <-timeoutCh:
			t.Fatalf("State root oracle not updated")
		case <-ticker.C:
		}
	}
}

// TestSystemE2E sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that L1 deposits are reflected on L2.
// All nodes are run in process (but are the full nodes, not mocked or stubbed).
func TestSystemE2E(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	log := testlog.Logger(t, log.LvlInfo)
	log.Info("genesis", "l2", sys.RollupConfig.Genesis.L2, "l1", sys.RollupConfig.Genesis.L1, "l2_time", sys.RollupConfig.Genesis.L2Time)

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// Transactor Account
	ethPrivKey := sys.cfg.Secrets.Alice

	//init BVMETH
	BVMETH, err := bindings.NewBVMETH(predeploys.BVM_ETHAddr, l2Seq)

	// Send Transaction & wait for success
	fromAddr := sys.cfg.Secrets.Addresses().Alice

	startBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	require.Nil(t, err)

	// Send deposit transaction
	opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, cfg.L1ChainIDBig())
	require.Nil(t, err)
	mintAmount := big.NewInt(1_000_000_000_000)
	opts.Value = mintAmount
	SendDepositTx(t, cfg, l1Client, l2Verif, opts, func(l2Opts *DepositTxOpts) {})

	// Confirm balance

	endBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	require.Nil(t, err)

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, mintAmount, diff, "Did not get expected balance change")

	// Submit TX to L2 sequencer node
	receipt := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.Value = big.NewInt(1_000_000_000)
		opts.Nonce = 1 // Already have deposit
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.VerifyOnClients(l2Verif)
	})

	// Verify blocks match after batch submission on verifiers and sequencers
	verifBlock, err := l2Verif.BlockByNumber(context.Background(), receipt.BlockNumber)
	require.Nil(t, err)
	seqBlock, err := l2Seq.BlockByNumber(context.Background(), receipt.BlockNumber)
	require.Nil(t, err)
	require.Equal(t, verifBlock.NumberU64(), seqBlock.NumberU64(), "Verifier and sequencer blocks not the same after including a batch tx")
	require.Equal(t, verifBlock.ParentHash(), seqBlock.ParentHash(), "Verifier and sequencer blocks parent hashes not the same after including a batch tx")
	require.Equal(t, verifBlock.Hash(), seqBlock.Hash(), "Verifier and sequencer blocks not the same after including a batch tx")

	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.Nil(t, err)
	rollupClient := sources.NewRollupClient(client.NewBaseRPCClient(rollupRPCClient))
	// basic check that sync status works
	seqStatus, err := rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.LessOrEqual(t, seqBlock.NumberU64(), seqStatus.UnsafeL2.Number)
	// basic check that version endpoint works
	seqVersion, err := rollupClient.Version(context.Background())
	require.Nil(t, err)
	require.NotEqual(t, "", seqVersion)
}

// TestConfirmationDepth runs the rollup with both sequencer and verifier not immediately processing the tip of the chain.
func TestConfirmationDepth(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	cfg.DeployConfig.SequencerWindowSize = 4
	cfg.DeployConfig.MaxSequencerDrift = 10 * cfg.DeployConfig.L1BlockTime
	seqConfDepth := uint64(2)
	verConfDepth := uint64(5)
	cfg.Nodes["sequencer"].Driver.SequencerConfDepth = seqConfDepth
	cfg.Nodes["sequencer"].Driver.VerifierConfDepth = 0
	cfg.Nodes["verifier"].Driver.VerifierConfDepth = verConfDepth

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	log := testlog.Logger(t, log.LvlInfo)
	log.Info("genesis", "l2", sys.RollupConfig.Genesis.L2, "l1", sys.RollupConfig.Genesis.L1, "l2_time", sys.RollupConfig.Genesis.L2Time)

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// Wait enough time for the sequencer to submit a block with distance from L1 head, submit it,
	// and for the slower verifier to read a full sequence window and cover confirmation depth for reading and some margin
	<-time.After(time.Duration((cfg.DeployConfig.SequencerWindowSize+verConfDepth+3)*cfg.DeployConfig.L1BlockTime) * time.Second)

	// within a second, get both L1 and L2 verifier and sequencer block heads
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l1Head, err := l1Client.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	l2SeqHead, err := l2Seq.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	l2VerHead, err := l2Verif.BlockByNumber(ctx, nil)
	require.NoError(t, err)

	seqInfo, err := derive.L1InfoDepositTxData(l2SeqHead.Transactions()[0].Data())
	require.NoError(t, err)
	require.LessOrEqual(t, seqInfo.Number+seqConfDepth, l1Head.NumberU64(), "the seq L2 head block should have an origin older than the L1 head block by at least the sequencer conf depth")

	verInfo, err := derive.L1InfoDepositTxData(l2VerHead.Transactions()[0].Data())
	require.NoError(t, err)
	require.LessOrEqual(t, verInfo.Number+verConfDepth, l1Head.NumberU64(), "the ver L2 head block should have an origin older than the L1 head block by at least the verifier conf depth")
}

// TestPendingGasLimit tests the configuration of the gas limit of the pending block,
// and if it does not conflict with the regular gas limit on the verifier or sequencer.
func TestPendingGasLimit(t *testing.T) {
	t.Skipf("skipping TestPendingGasLimit tests")
	return
	InitParallel(t)

	cfg := DefaultSystemConfig(t)

	// configure the L2 gas limit to be high, and the pending gas limits to be lower for resource saving.
	cfg.DeployConfig.L2GenesisBlockGasLimit = 30_000_000
	cfg.GethOptions["sequencer"] = []GethOption{
		func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
			ethCfg.Miner.GasCeil = 10_000_000
			return nil
		},
	}
	cfg.GethOptions["verifier"] = []GethOption{
		func(ethCfg *ethconfig.Config, nodeCfg *node.Config) error {
			ethCfg.Miner.GasCeil = 9_000_000
			return nil
		},
	}

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	log := testlog.Logger(t, log.LvlInfo)
	log.Info("genesis", "l2", sys.RollupConfig.Genesis.L2, "l1", sys.RollupConfig.Genesis.L1, "l2_time", sys.RollupConfig.Genesis.L2Time)

	l2Verif := sys.Clients["verifier"]
	l2Seq := sys.Clients["sequencer"]

	checkGasLimit := func(client *ethclient.Client, number *big.Int, expected uint64) *types.Header {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		header, err := client.HeaderByNumber(ctx, number)
		cancel()
		require.NoError(t, err)
		require.Equal(t, expected, header.GasLimit)
		return header
	}

	// check if the gaslimits are matching the expected values,
	// and that the verifier/sequencer can use their locally configured gas limit for the pending block.
	for {
		checkGasLimit(l2Seq, big.NewInt(-1), 10_000_000)
		checkGasLimit(l2Verif, big.NewInt(-1), 9_000_000)
		checkGasLimit(l2Seq, nil, 30_000_000)
		latestVerifHeader := checkGasLimit(l2Verif, nil, 30_000_000)

		// Stop once the verifier passes genesis:
		// this implies we checked a new block from the sequencer, on both sequencer and verifier nodes.
		if latestVerifHeader.Number.Uint64() > 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// TestFinalize tests if L2 finalizes after sufficient time after L1 finalizes
func TestFinalize(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]

	// as configured in the extra geth lifecycle in testing setup
	const finalizedDistance = 8
	// Wait enough time for L1 to finalize and L2 to confirm its data in finalized L1 blocks
	time.Sleep(time.Duration((finalizedDistance+6)*cfg.DeployConfig.L1BlockTime) * time.Second)

	// fetch the finalizes head of geth
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l2Finalized, err := l2Seq.BlockByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	require.NoError(t, err)

	require.NotZerof(t, l2Finalized.NumberU64(), "must have finalized L2 block")
}

func TestMintOnRevertedDeposit(t *testing.T) {
	t.Skipf("skipping TestPendingGasLimit tests")
	return

	InitParallel(t)
	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]
	l2Verif := sys.Clients["verifier"]
	l2Seq := sys.Clients["sequencer"]
	l1Node := sys.Nodes["l1"]

	//create BVM_ETH
	BVMETH, err := bindings.NewBVMETH(predeploys.BVM_ETHAddr, l2Seq)

	// create signer
	ks := l1Node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, ks.Accounts()[0], cfg.L1ChainIDBig())
	require.Nil(t, err)
	fromAddr := opts.From

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	startBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	startMNTBalance, err := l2Verif.BalanceAt(ctx, fromAddr, nil)
	cancel()
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	startNonce, err := l2Verif.NonceAt(ctx, fromAddr, nil)
	require.NoError(t, err)
	cancel()

	toAddr := common.Address{0xff, 0xff}
	mintAmount := big.NewInt(9_000_000)
	opts.Value = mintAmount
	SendDepositTx(t, cfg, l1Client, l2Verif, opts, func(l2Opts *DepositTxOpts) {
		l2Opts.ToAddr = toAddr
		// trigger a revert by transferring more than we have available
		l2Opts.ETHValue = big.NewInt(0)
		//new(big.Int).Add(big.NewInt(10000), startBalance)
		l2Opts.MNTValue = new(big.Int).Mul(common.Big2, startMNTBalance)
		l2Opts.ExpectedStatus = types.ReceiptStatusFailed
	})

	// Confirm balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	endBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	cancel()
	require.Nil(t, err)
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	toAddrBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, toAddr)
	require.NoError(t, err)
	cancel()

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, mintAmount, diff, "Did not get expected balance change")
	require.Equal(t, common.Big0.Int64(), toAddrBalance.Int64(), "The recipient account balance should be zero")

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	endNonce, err := l2Verif.NonceAt(ctx, fromAddr, nil)
	require.NoError(t, err)
	cancel()
	require.Equal(t, startNonce+1, endNonce, "Nonce of deposit sender should increment on L2, even if the deposit fails")
}

func TestMissingBatchE2E(t *testing.T) {
	InitParallel(t)
	// Note this test zeroes the balance of the batch-submitter to make the batches unable to go into L1.
	// The test logs may look scary, but this is expected:
	// 'batcher unable to publish transaction    role=batcher   err="insufficient funds for gas * price + value"'

	cfg := DefaultSystemConfig(t)
	// small sequence window size so the test does not take as long
	cfg.DeployConfig.SequencerWindowSize = 4

	// Specifically set batch submitter balance to stop batches from being included
	cfg.Premine[cfg.Secrets.Addresses().Batcher] = big.NewInt(0)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]
	seqRollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.Nil(t, err)
	seqRollupClient := sources.NewRollupClient(client.NewBaseRPCClient(seqRollupRPCClient))

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit TX to L2 sequencer node
	receipt := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.Value = big.NewInt(1_000_000_000)
	})

	// Wait until the block it was first included in shows up in the safe chain on the verifier
	_, err = waitForBlock(receipt.BlockNumber, l2Verif, time.Duration((sys.RollupConfig.SeqWindowSize+4)*cfg.DeployConfig.L1BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for block on verifier")

	// Assert that the transaction is not found on the verifier
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = l2Verif.TransactionReceipt(ctx, receipt.TxHash)
	require.Equal(t, ethereum.NotFound, err, "Found transaction in verifier when it should not have been included")

	// Wait a short time for the L2 reorg to occur on the sequencer as well.
	err = waitForSafeHead(ctx, receipt.BlockNumber.Uint64(), seqRollupClient)
	require.Nil(t, err, "timeout waiting for L2 reorg on sequencer safe head")

	// Assert that the reconciliation process did an L2 reorg on the sequencer to remove the invalid block
	ctx2, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	block, err := l2Seq.BlockByNumber(ctx2, receipt.BlockNumber)
	require.Nil(t, err, "Get block from sequencer")
	require.NotEqual(t, block.Hash(), receipt.BlockHash, "L2 Sequencer did not reorg out transaction on it's safe chain")
}

func L1InfoFromState(ctx context.Context, contract *bindings.L1Block, l2Number *big.Int) (derive.L1BlockInfo, error) {
	var err error
	var out derive.L1BlockInfo
	opts := bind.CallOpts{
		BlockNumber: l2Number,
		Context:     ctx,
	}

	out.Number, err = contract.Number(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get number: %w", err)
	}

	out.Time, err = contract.Timestamp(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get timestamp: %w", err)
	}

	out.BaseFee, err = contract.Basefee(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get timestamp: %w", err)
	}

	blockHashBytes, err := contract.Hash(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get block hash: %w", err)
	}
	out.BlockHash = common.BytesToHash(blockHashBytes[:])

	out.SequenceNumber, err = contract.SequenceNumber(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get sequence number: %w", err)
	}

	overhead, err := contract.L1FeeOverhead(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get l1 fee overhead: %w", err)
	}
	out.L1FeeOverhead = eth.Bytes32(common.BigToHash(overhead))

	scalar, err := contract.L1FeeScalar(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get l1 fee scalar: %w", err)
	}
	out.L1FeeScalar = eth.Bytes32(common.BigToHash(scalar))

	batcherHash, err := contract.BatcherHash(&opts)
	if err != nil {
		return derive.L1BlockInfo{}, fmt.Errorf("failed to get batch sender: %w", err)
	}
	out.BatcherAddr = common.BytesToAddress(batcherHash[:])

	return out, nil
}

// TestSystemMockP2P sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that
// the nodes can sync L2 blocks before they are confirmed on L1.
func TestSystemMockP2P(t *testing.T) {
	t.Skip("flaky in CI") // TODO(CLI-3859): Re-enable this test.
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	// Disable batcher, so we don't sync from L1 & set a large sequence window so we only have unsafe blocks
	cfg.DisableBatcher = true
	cfg.DeployConfig.SequencerWindowSize = 100_000
	cfg.DeployConfig.MaxSequencerDrift = 100_000
	// disable at the start, so we don't miss any gossiped blocks.
	cfg.Nodes["sequencer"].Driver.SequencerStopped = true

	// connect the nodes
	cfg.P2PTopology = map[string][]string{
		"verifier": {"sequencer"},
	}

	var published, received []common.Hash
	seqTracer, verifTracer := new(FnTracer), new(FnTracer)
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayloadEnvelope) {
		published = append(published, payload.ExecutionPayload.BlockHash)
	}
	verifTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
		received = append(received, payload.ExecutionPayload.BlockHash)
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["verifier"].Tracer = verifTracer

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	// Enable the sequencer now that everyone is ready to receive payloads.
	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["sequencer"].HTTPEndpoint())
	require.Nil(t, err)

	verifierPeerID := sys.RollupNodes["verifier"].P2P().Host().ID()
	check := func() bool {
		sequencerBlocksTopicPeers := sys.RollupNodes["sequencer"].P2P().GossipOut().AllBlockTopicsPeers()
		return slices.Contains(sequencerBlocksTopicPeers, verifierPeerID)
	}

	// poll to see if the verifier node is connected & meshed on gossip.
	// Without this verifier, we shouldn't start sending blocks around, or we'll miss them and fail the test.
	backOffStrategy := backoff.Exponential()
	for i := 0; i < 10; i++ {
		if check() {
			break
		}
		time.Sleep(backOffStrategy.Duration(i))
	}
	require.True(t, check(), "verifier must be meshed with sequencer for gossip test to proceed")

	require.NoError(t, rollupRPCClient.Call(nil, "admin_startSequencer", sys.L2GenesisCfg.ToBlock().Hash()))

	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit TX to L2 sequencer node
	receiptSeq := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.Value = big.NewInt(1_000_000_000)

		// Wait until the block it was first included in shows up in the safe chain on the verifier
		opts.VerifyOnClients(l2Verif)
	})

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received))
	require.ElementsMatch(t, received, published[:len(received)])

	// Verify that the tx was received via p2p
	require.Contains(t, received, receiptSeq.BlockHash)
}

// TestSystemRPCAltSync sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that
// the nodes can sync L2 blocks before they are confirmed on L1.
//
// Test steps:
// 1. Spin up the nodes (P2P is disabled on the verifier)
// 2. Send a transaction to the sequencer.
// 3. Wait for the TX to be mined on the sequencer chain.
// 5. Wait for the verifier to detect a gap in the payload queue vs. the unsafe head
// 6. Wait for the RPC sync method to grab the block from the sequencer over RPC and insert it into the verifier's unsafe chain.
// 7. Wait for the verifier to sync the unsafe chain into the safe chain.
// 8. Verify that the TX is included in the verifier's safe chain.
func TestSystemRPCAltSync(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	// the default is nil, but this may change in the future.
	// This test must ensure the blocks are not synced via Gossip, but instead via the alt RPC based sync.
	cfg.P2PTopology = nil
	// Disable batcher, so there will not be any L1 data to sync from
	cfg.DisableBatcher = true

	var published, received []string
	seqTracer, verifTracer := new(FnTracer), new(FnTracer)
	// The sequencer still publishes the blocks to the tracer, even if they do not reach the network due to disabled P2P
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayloadEnvelope) {
		published = append(published, payload.ID().String())
	}
	// Blocks are now received via the RPC based alt-sync method
	verifTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
		received = append(received, payload.ID().String())
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["verifier"].Tracer = verifTracer

	sys, err := cfg.Start(SystemConfigOption{
		key:  "afterRollupNodeStart",
		role: "sequencer",
		action: func(sCfg *SystemConfig, system *System) {
			rpc := system.Nodes["sequencer"].Attach() // never errors
			cfg.Nodes["verifier"].L2Sync = &rollupNode.PreparedL2SyncEndpoint{
				Client: client.NewBaseRPCClient(rpc),
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

	// Verify that the tx was received via RPC sync (P2P is disabled)
	require.Contains(t, received, eth.BlockID{Hash: receiptSeq.BlockHash, Number: receiptSeq.BlockNumber.Uint64()}.String())

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received))
	require.ElementsMatch(t, received, published[:len(received)])
}

func TestSystemP2PAltSync(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)

	// remove default verifier node
	delete(cfg.Nodes, "verifier")
	// Add more verifier nodes
	cfg.Nodes["alice"] = &rollupNode.Config{
		Driver: driver.Config{
			VerifierConfDepth:  0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Nodes["bob"] = &rollupNode.Config{
		Driver: driver.Config{
			VerifierConfDepth:  0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Loggers["alice"] = testlog.Logger(t, log.LvlInfo).New("role", "alice")
	cfg.Loggers["bob"] = testlog.Logger(t, log.LvlInfo).New("role", "bob")

	// connect the nodes
	cfg.P2PTopology = map[string][]string{
		"sequencer": {"alice", "bob"},
		"alice":     {"sequencer", "bob"},
		"bob":       {"alice", "sequencer"},
	}
	// Enable the P2P req-resp based sync
	cfg.P2PReqRespSync = true

	// Disable batcher, so there will not be any L1 data to sync from
	cfg.DisableBatcher = true

	var published []string
	seqTracer := new(FnTracer)
	// The sequencer still publishes the blocks to the tracer, even if they do not reach the network due to disabled P2P
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayloadEnvelope) {
		published = append(published, payload.ID().String())
	}
	// Blocks are now received via the RPC based alt-sync method
	cfg.Nodes["sequencer"].Tracer = seqTracer

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit a TX to L2 sequencer node
	receiptSeq := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.Value = big.NewInt(1_000_000_000)
	})

	// Gossip is able to respond to IWANT messages for the duration of heartbeat_time * message_window = 0.5 * 12 = 6
	// Wait till we pass that, and then we'll have missed some blocks that cannot be retrieved in any way from gossip
	time.Sleep(time.Second * 10)

	// set up our syncer node, connect it to alice/bob
	cfg.Loggers["syncer"] = testlog.Logger(t, log.LvlInfo).New("role", "syncer")
	snapLog := log.NewLogger(log.DiscardHandler())

	// Create a peer, and hook up alice and bob
	h, err := sys.Mocknet.GenPeer()
	require.NoError(t, err)
	_, err = sys.Mocknet.LinkPeers(sys.RollupNodes["alice"].P2P().Host().ID(), h.ID())
	require.NoError(t, err)
	_, err = sys.Mocknet.LinkPeers(sys.RollupNodes["bob"].P2P().Host().ID(), h.ID())
	require.NoError(t, err)

	// Configure the new rollup node that'll be syncing
	var syncedPayloads []string
	syncNodeCfg := &rollupNode.Config{
		L2Sync:    &rollupNode.PreparedL2SyncEndpoint{Client: nil},
		Driver:    driver.Config{VerifierConfDepth: 0},
		Rollup:    *sys.RollupConfig,
		P2PSigner: nil,
		RPC: rollupNode.RPCConfig{
			ListenAddr:  "127.0.0.1",
			ListenPort:  0,
			EnableAdmin: true,
		},
		P2P:                 &p2p.Prepared{HostP2P: h, EnableReqRespSync: true},
		Metrics:             rollupNode.MetricsConfig{Enabled: false}, // no metrics server
		Pprof:               oppprof.CLIConfig{},
		L1EpochPollInterval: time.Second * 10,
		Tracer: &FnTracer{
			OnUnsafeL2PayloadFn: func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
				syncedPayloads = append(syncedPayloads, payload.ID().String())
			},
		},
	}
	configureL1(syncNodeCfg, sys.Nodes["l1"])
	syncerL2Engine, _, err := initL2Geth("syncer", big.NewInt(int64(cfg.DeployConfig.L2ChainID)), sys.L2GenesisCfg, cfg.JWTFilePath)
	require.NoError(t, err)
	require.NoError(t, syncerL2Engine.Start())

	configureL2(syncNodeCfg, syncerL2Engine, cfg.JWTSecret)

	syncerNode, err := rollupNode.New(context.Background(), syncNodeCfg, cfg.Loggers["syncer"], snapLog, "", metrics.NewMetrics(""))
	require.NoError(t, err)
	err = syncerNode.Start(context.Background())
	require.NoError(t, err)

	// connect alice and bob to our new syncer node
	_, err = sys.Mocknet.ConnectPeers(sys.RollupNodes["alice"].P2P().Host().ID(), syncerNode.P2P().Host().ID())
	require.NoError(t, err)
	_, err = sys.Mocknet.ConnectPeers(sys.RollupNodes["bob"].P2P().Host().ID(), syncerNode.P2P().Host().ID())
	require.NoError(t, err)

	rpc := syncerL2Engine.Attach()
	l2Verif := ethclient.NewClient(rpc)

	// It may take a while to sync, but eventually we should see the sequenced data show up
	receiptVerif, err := waitForTransaction(receiptSeq.TxHash, l2Verif, 100*time.Duration(sys.RollupConfig.BlockTime)*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on verifier")

	require.Equal(t, receiptSeq, receiptVerif)

	// Verify that the tx was received via P2P sync
	require.Contains(t, syncedPayloads, eth.BlockID{Hash: receiptVerif.BlockHash, Number: receiptVerif.BlockNumber.Uint64()}.String())

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(syncedPayloads))
	require.ElementsMatch(t, syncedPayloads, published[:len(syncedPayloads)])
}

// TestSystemDenseTopology sets up a dense p2p topology with 3 verifier nodes and 1 sequencer node.
func TestSystemDenseTopology(t *testing.T) {
	t.Skip("Skipping dense topology test to avoid flakiness. @refcell address in p2p scoring pr.")

	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	// slow down L1 blocks so we can see the L2 blocks arrive well before the L1 blocks do.
	// Keep the seq window small so the L2 chain is started quick
	cfg.DeployConfig.L1BlockTime = 10

	// Append additional nodes to the system to construct a dense p2p network
	cfg.Nodes["verifier2"] = &rollupNode.Config{
		Driver: driver.Config{
			VerifierConfDepth:  0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Nodes["verifier3"] = &rollupNode.Config{
		Driver: driver.Config{
			VerifierConfDepth:  0,
			SequencerConfDepth: 0,
			SequencerEnabled:   false,
		},
		L1EpochPollInterval: time.Second * 4,
	}
	cfg.Loggers["verifier2"] = testlog.Logger(t, log.LvlInfo).New("role", "verifier")
	cfg.Loggers["verifier3"] = testlog.Logger(t, log.LvlInfo).New("role", "verifier")

	// connect the nodes
	cfg.P2PTopology = map[string][]string{
		"verifier":  {"sequencer", "verifier2", "verifier3"},
		"verifier2": {"sequencer", "verifier", "verifier3"},
		"verifier3": {"sequencer", "verifier", "verifier2"},
	}

	// Set peer scoring for each node, but without banning
	for _, node := range cfg.Nodes {
		params, err := p2p.GetPeerScoreParams("light", 2)
		require.NoError(t, err)
		node.P2P = &p2p.Config{
			PeerScoring:    params,
			BanningEnabled: false,
		}
	}

	var published, received1, received2, received3 []common.Hash
	seqTracer, verifTracer, verifTracer2, verifTracer3 := new(FnTracer), new(FnTracer), new(FnTracer), new(FnTracer)
	seqTracer.OnPublishL2PayloadFn = func(ctx context.Context, payload *eth.ExecutionPayloadEnvelope) {
		published = append(published, payload.ExecutionPayload.BlockHash)
	}
	verifTracer.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
		received1 = append(received1, payload.ExecutionPayload.BlockHash)
	}
	verifTracer2.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
		received2 = append(received2, payload.ExecutionPayload.BlockHash)
	}
	verifTracer3.OnUnsafeL2PayloadFn = func(ctx context.Context, from peer.ID, payload *eth.ExecutionPayloadEnvelope) {
		received3 = append(received3, payload.ExecutionPayload.BlockHash)
	}
	cfg.Nodes["sequencer"].Tracer = seqTracer
	cfg.Nodes["verifier"].Tracer = verifTracer
	cfg.Nodes["verifier2"].Tracer = verifTracer2
	cfg.Nodes["verifier3"].Tracer = verifTracer3

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]
	l2Verif2 := sys.Clients["verifier2"]
	l2Verif3 := sys.Clients["verifier3"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice

	// Submit TX to L2 sequencer node
	receiptSeq := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.Value = big.NewInt(1_000_000_000)

		// Wait until the block it was first included in shows up in the safe chain on the verifiers
		opts.VerifyOnClients(l2Verif, l2Verif2, l2Verif3)
	})

	// Verify that everything that was received was published
	require.GreaterOrEqual(t, len(published), len(received1))
	require.GreaterOrEqual(t, len(published), len(received2))
	require.GreaterOrEqual(t, len(published), len(received3))
	require.ElementsMatch(t, published, received1[:len(published)])
	require.ElementsMatch(t, published, received2[:len(published)])
	require.ElementsMatch(t, published, received3[:len(published)])

	// Verify that the tx was received via p2p
	require.Contains(t, received1, receiptSeq.BlockHash)
	require.Contains(t, received2, receiptSeq.BlockHash)
	require.Contains(t, received3, receiptSeq.BlockHash)
}

func TestL1InfoContract(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	endVerifBlockNumber := big.NewInt(4)
	endSeqBlockNumber := big.NewInt(6)
	endVerifBlock, err := waitForBlock(endVerifBlockNumber, l2Verif, time.Minute)
	require.Nil(t, err)
	endSeqBlock, err := waitForBlock(endSeqBlockNumber, l2Seq, time.Minute)
	require.Nil(t, err)

	seqL1Info, err := bindings.NewL1Block(cfg.L1InfoPredeployAddress, l2Seq)
	require.Nil(t, err)

	verifL1Info, err := bindings.NewL1Block(cfg.L1InfoPredeployAddress, l2Verif)
	require.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fillInfoLists := func(start *types.Block, contract *bindings.L1Block, client *ethclient.Client) ([]derive.L1BlockInfo, []derive.L1BlockInfo) {
		var txList, stateList []derive.L1BlockInfo
		for b := start; ; {
			var infoFromTx derive.L1BlockInfo
			require.NoError(t, infoFromTx.UnmarshalBinary(b.Transactions()[0].Data()))
			txList = append(txList, infoFromTx)

			infoFromState, err := L1InfoFromState(ctx, contract, b.Number())
			require.Nil(t, err)
			stateList = append(stateList, infoFromState)

			// Genesis L2 block contains no L1 Deposit TX
			if b.NumberU64() == 1 {
				return txList, stateList
			}
			b, err = client.BlockByHash(ctx, b.ParentHash())
			require.Nil(t, err)
		}
	}

	l1InfosFromSequencerTransactions, l1InfosFromSequencerState := fillInfoLists(endSeqBlock, seqL1Info, l2Seq)
	l1InfosFromVerifierTransactions, l1InfosFromVerifierState := fillInfoLists(endVerifBlock, verifL1Info, l2Verif)

	l1blocks := make(map[common.Hash]derive.L1BlockInfo)
	maxL1Hash := l1InfosFromSequencerTransactions[0].BlockHash
	for h := maxL1Hash; ; {
		b, err := l1Client.BlockByHash(ctx, h)
		require.Nil(t, err)

		l1blocks[h] = derive.L1BlockInfo{
			Number:         b.NumberU64(),
			Time:           b.Time(),
			BaseFee:        b.BaseFee(),
			BlockHash:      h,
			SequenceNumber: 0, // ignored, will be overwritten
			BatcherAddr:    sys.RollupConfig.Genesis.SystemConfig.BatcherAddr,
			L1FeeOverhead:  sys.RollupConfig.Genesis.SystemConfig.Overhead,
			L1FeeScalar:    sys.RollupConfig.Genesis.SystemConfig.Scalar,
		}

		h = b.ParentHash()
		if b.NumberU64() == 0 {
			break
		}
	}

	checkInfoList := func(name string, list []derive.L1BlockInfo) {
		for _, info := range list {
			if expected, ok := l1blocks[info.BlockHash]; ok {
				expected.SequenceNumber = info.SequenceNumber // the seq nr is not part of the L1 info we know in advance, so we ignore it.
				require.Equal(t, expected, info)
			} else {
				t.Fatalf("Did not find block hash for L1 Info: %v in test %s", info, name)
			}
		}
	}

	checkInfoList("On sequencer with tx", l1InfosFromSequencerTransactions)
	checkInfoList("On sequencer with state", l1InfosFromSequencerState)
	checkInfoList("On verifier with tx", l1InfosFromVerifierTransactions)
	checkInfoList("On verifier with state", l1InfosFromVerifierState)

}

// calcGasFees determines the actual cost of the transaction given a specific basefee
// This does not include the L1 data fee charged from L2 transactions.
func calcGasFees(gasUsed uint64, gasTipCap *big.Int, gasFeeCap *big.Int, baseFee *big.Int) *big.Int {
	x := new(big.Int).Add(gasTipCap, baseFee)
	// If tip + basefee > gas fee cap, clamp it to the gas fee cap
	if x.Cmp(gasFeeCap) > 0 {
		x = gasFeeCap
	}
	return x.Mul(x, new(big.Int).SetUint64(gasUsed))
}

// calcL1GasUsed returns the gas used to include the transaction data in
// the calldata on L1
func calcL1GasUsed(data []byte, overhead *big.Int) *big.Int {
	var zeroes, ones uint64
	for _, byt := range data {
		if byt == 0 {
			zeroes++
		} else {
			ones++
		}
	}

	zeroesGas := zeroes * 4     // params.TxDataZeroGas
	onesGas := (ones + 68) * 16 // params.TxDataNonZeroGasEIP2028
	l1Gas := new(big.Int).SetUint64(zeroesGas + onesGas)
	return new(big.Int).Add(l1Gas, overhead)
}

// TestWithdrawals checks that a deposit and then withdrawal execution succeeds. It verifies the
// balance changes on L1 and L2 and has to include gas fees in the balance checks.
// It does not check that the withdrawal can be executed prior to the end of the finality period.
func TestWithdrawals(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	cfg.DeployConfig.FinalizationPeriodSeconds = 2 // 2s finalization period

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l1Client := sys.Clients["l1"]
	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice
	fromAddr := crypto.PubkeyToAddress(ethPrivKey.PublicKey)

	// Create L1 signer
	opts, err := bind.NewKeyedTransactorWithChainID(ethPrivKey, cfg.L1ChainIDBig())
	require.Nil(t, err)

	// Start L2 balance

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	BVMETH, err := bindings.NewBVMETH(predeploys.BVM_ETHAddr, l2Seq)
	startBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	require.Nil(t, err)

	// Send deposit tx
	mintAmount := big.NewInt(1_000_000_000_000)
	opts.Value = mintAmount
	SendDepositTx(t, cfg, l1Client, l2Verif, opts, func(l2Opts *DepositTxOpts) {
		l2Opts.ETHValue = common.Big0
		l2Opts.MNTValue = common.Big0
	})

	// Confirm L2 balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	require.Nil(t, err)

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, mintAmount, diff, "Did not get expected balance change after mint")

	// Start L2 balance for withdrawal
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err = BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	require.Nil(t, err)

	withdrawAmount := big.NewInt(500_000_000_000)
	tx, receipt := SendWithdrawal(t, cfg, l2Seq, ethPrivKey, func(opts *WithdrawalTxOpts) {
		opts.ETHValue = withdrawAmount
		opts.VerifyOnClients(l2Verif)
	})

	// Verify L2 balance after withdrawal
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err := l2Verif.HeaderByNumber(ctx, receipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err = BVMETH.BalanceOf(&bind.CallOpts{}, fromAddr)
	require.Nil(t, err)

	// Take fee into account
	diff = new(big.Int).Sub(startBalance, endBalance)
	//TODO add fee calculation into MNT withdrawal
	//fees := calcGasFees(receipt.GasUsed, tx.GasTipCap(), tx.GasFeeCap(), header.BaseFee)
	//fees = fees.Add(fees, receipt.L1Fee)
	//diff = diff.Sub(diff, fees)
	require.Equal(t, withdrawAmount, diff)

	// Take start balance on L1
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err = l1Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	proveReceipt, finalizeReceipt := ProveAndFinalizeWithdrawal(t, cfg, l1Client, sys.Nodes["verifier"], ethPrivKey, receipt)

	// Verify balance after withdrawal
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err = l1Client.HeaderByNumber(ctx, finalizeReceipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err = l1Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Ensure that withdrawal - gas fees are added to the L1 balance
	// Fun fact, the fee is greater than the withdrawal amount
	// NOTE: The gas fees include *both* the ProveWithdrawalTransaction and FinalizeWithdrawalTransaction transactions.
	diff = new(big.Int).Sub(endBalance, startBalance)
	fees := calcGasFees(proveReceipt.GasUsed+finalizeReceipt.GasUsed, tx.GasTipCap(), tx.GasFeeCap(), header.BaseFee)
	withdrawAmount = withdrawAmount.Sub(withdrawAmount, fees)
	require.Equal(t, withdrawAmount, diff)
}

// TestFees checks that L1/L2 fees are handled.
func TestFees(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	// TODO: after we have the system config contract and new op-geth L1 cost utils,
	// we can pull in l1 costs into every e2e test and account for it in assertions easily etc.
	cfg.DeployConfig.GasPriceOracleOverhead = 2100
	cfg.DeployConfig.GasPriceOracleScalar = 1000_000

	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// Transactor Account
	ethPrivKey := cfg.Secrets.Alice
	fromAddr := crypto.PubkeyToAddress(ethPrivKey.PublicKey)

	// Find gaspriceoracle contract
	gpoContract, err := bindings.NewGasPriceOracle(predeploys.GasPriceOracleAddr, l2Seq)
	require.Nil(t, err)

	overhead, err := gpoContract.Overhead(&bind.CallOpts{})
	require.Nil(t, err, "reading gpo overhead")
	decimals, err := gpoContract.Decimals(&bind.CallOpts{})
	require.Nil(t, err, "reading gpo decimals")
	scalar, err := gpoContract.Scalar(&bind.CallOpts{})
	require.Nil(t, err, "reading gpo scalar")

	require.Equal(t, overhead.Uint64(), uint64(2100), "wrong gpo overhead")
	require.Equal(t, decimals.Uint64(), uint64(6), "wrong gpo decimals")
	require.Equal(t, scalar.Uint64(), uint64(1_000_000), "wrong gpo scalar")

	// BaseFee Recipient
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	baseFeeRecipientStartBalance, err := l2Seq.BalanceAt(ctx, predeploys.BaseFeeVaultAddr, nil)
	require.Nil(t, err)

	// L1Fee Recipient
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l1FeeRecipientStartBalance, err := l2Seq.BalanceAt(ctx, predeploys.L1FeeVaultAddr, nil)
	require.Nil(t, err)

	// Simple transfer from signer to random account
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err := l2Verif.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	transferAmount := big.NewInt(1_000_000_000)
	gasTip := big.NewInt(10)
	receipt := SendL2Tx(t, cfg, l2Seq, ethPrivKey, func(opts *TxOpts) {
		opts.ToAddr = &common.Address{0xff, 0xff}
		opts.Value = transferAmount
		opts.GasTipCap = gasTip
		opts.Gas = 21000
		opts.GasFeeCap = big.NewInt(200)
		opts.VerifyOnClients(l2Verif)
	})

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	header, err := l2Seq.HeaderByNumber(ctx, receipt.BlockNumber)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	coinbaseStartBalance, err := l2Seq.BalanceAt(ctx, header.Coinbase, safeAddBig(header.Number, big.NewInt(-1)))
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	coinbaseEndBalance, err := l2Seq.BalanceAt(ctx, header.Coinbase, header.Number)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := l2Seq.BalanceAt(ctx, fromAddr, header.Number)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	baseFeeRecipientEndBalance, err := l2Seq.BalanceAt(ctx, predeploys.BaseFeeVaultAddr, header.Number)
	require.Nil(t, err)

	l1Header, err := sys.Clients["l1"].HeaderByNumber(ctx, nil)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	l1FeeRecipientEndBalance, err := l2Seq.BalanceAt(ctx, predeploys.L1FeeVaultAddr, header.Number)
	require.Nil(t, err)

	// Diff fee recipient + coinbase balances
	baseFeeRecipientDiff := new(big.Int).Sub(baseFeeRecipientEndBalance, baseFeeRecipientStartBalance)
	l1FeeRecipientDiff := new(big.Int).Sub(l1FeeRecipientEndBalance, l1FeeRecipientStartBalance)
	coinbaseDiff := new(big.Int).Sub(coinbaseEndBalance, coinbaseStartBalance)

	// Tally L2 Fee
	l2Fee := gasTip.Mul(gasTip, new(big.Int).SetUint64(receipt.GasUsed))
	require.Equal(t, l2Fee, coinbaseDiff, "l2 fee mismatch")

	// Tally BaseFee
	baseFee := new(big.Int).Mul(header.BaseFee, new(big.Int).SetUint64(receipt.GasUsed))
	require.Equal(t, baseFee, baseFeeRecipientDiff, "base fee fee mismatch")

	// Tally L1 Fee
	tx, _, err := l2Seq.TransactionByHash(ctx, receipt.TxHash)
	require.NoError(t, err, "Should be able to get transaction")
	bytes, err := tx.MarshalBinary()
	require.Nil(t, err)
	l1GasUsed := calcL1GasUsed(bytes, overhead)
	divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)
	l1Fee := new(big.Int).Mul(l1GasUsed, l1Header.BaseFee)
	l1Fee = l1Fee.Mul(l1Fee, scalar)
	l1Fee = l1Fee.Div(l1Fee, divisor)

	require.Equal(t, l1Fee, l1FeeRecipientDiff, "l1 fee mismatch")

	// Tally L1 fee against GasPriceOracle
	gpoL1Fee, err := gpoContract.GetL1Fee(&bind.CallOpts{}, bytes)
	require.Nil(t, err)
	require.Equal(t, l1Fee, gpoL1Fee, "l1 fee mismatch")

	// Calculate total fee
	baseFeeRecipientDiff.Add(baseFeeRecipientDiff, coinbaseDiff)
	totalFee := new(big.Int).Add(baseFeeRecipientDiff, l1FeeRecipientDiff)
	balanceDiff := new(big.Int).Sub(startBalance, endBalance)
	balanceDiff.Sub(balanceDiff, transferAmount)
	require.Equal(t, balanceDiff, totalFee, "balances should add up")
}

func TestStopStartSequencer(t *testing.T) {
	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	l2Seq := sys.Clients["sequencer"]
	rollupNode := sys.RollupNodes["sequencer"]

	nodeRPC, err := rpc.DialContext(context.Background(), rollupNode.HTTPEndpoint())
	require.Nil(t, err, "Error dialing node")

	blockBefore := latestBlock(t, l2Seq)
	time.Sleep(time.Duration(cfg.DeployConfig.L2BlockTime+1) * time.Second)
	blockAfter := latestBlock(t, l2Seq)
	require.Greaterf(t, blockAfter, blockBefore, "Chain did not advance")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	blockHash := common.Hash{}
	err = nodeRPC.CallContext(ctx, &blockHash, "admin_stopSequencer")
	require.Nil(t, err, "Error stopping sequencer")

	blockBefore = latestBlock(t, l2Seq)
	time.Sleep(time.Duration(cfg.DeployConfig.L2BlockTime+1) * time.Second)
	blockAfter = latestBlock(t, l2Seq)
	require.Equal(t, blockAfter, blockBefore, "Chain advanced after stopping sequencer")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = nodeRPC.CallContext(ctx, nil, "admin_startSequencer", blockHash)
	require.Nil(t, err, "Error starting sequencer")

	blockBefore = latestBlock(t, l2Seq)
	time.Sleep(time.Duration(cfg.DeployConfig.L2BlockTime+1) * time.Second)
	blockAfter = latestBlock(t, l2Seq)
	require.Greater(t, blockAfter, blockBefore, "Chain did not advance after starting sequencer")
}

func TestStopStartBatcher(t *testing.T) {
	//TODO skip this test cuz of 502 Gate Error
	t.Skipf("skipping TestStopStartBatcher tests")
	return

	InitParallel(t)

	cfg := DefaultSystemConfig(t)
	sys, err := cfg.Start()
	require.Nil(t, err, "Error starting up system")
	defer sys.Close()

	rollupRPCClient, err := rpc.DialContext(context.Background(), sys.RollupNodes["verifier"].HTTPEndpoint())
	require.Nil(t, err)
	rollupClient := sources.NewRollupClient(client.NewBaseRPCClient(rollupRPCClient))

	l2Seq := sys.Clients["sequencer"]
	l2Verif := sys.Clients["verifier"]

	// retrieve the initial sync status
	seqStatus, err := rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)

	nonce := uint64(0)
	sendTx := func() *types.Receipt {
		// Submit TX to L2 sequencer node
		receipt := SendL2Tx(t, cfg, l2Seq, cfg.Secrets.Alice, func(opts *TxOpts) {
			opts.ToAddr = &common.Address{0xff, 0xff}
			opts.Value = big.NewInt(1_000_000_000)
			opts.Nonce = nonce
		})
		nonce++
		return receipt
	}
	// send a transaction
	receipt := sendTx()

	// wait until the block the tx was first included in shows up in the safe chain on the verifier
	safeBlockInclusionDuration := time.Duration(3*cfg.DeployConfig.L1BlockTime) * time.Second
	_, err = waitForBlock(receipt.BlockNumber, l2Verif, safeBlockInclusionDuration)
	require.Nil(t, err, "Waiting for block on verifier")

	// ensure the safe chain advances
	newSeqStatus, err := rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain did not advance")

	// stop the batch submission
	err = sys.BatchSubmitter.Stop(context.Background())
	require.Nil(t, err)

	// wait for any old safe blocks being submitted / derived
	time.Sleep(safeBlockInclusionDuration)

	// get the initial sync status
	seqStatus, err = rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)

	// send another tx
	sendTx()
	time.Sleep(safeBlockInclusionDuration)

	// ensure that the safe chain does not advance while the batcher is stopped
	newSeqStatus, err = rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.Equal(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain advanced while batcher was stopped")

	// start the batch submission
	err = sys.BatchSubmitter.Start()
	require.Nil(t, err)
	time.Sleep(safeBlockInclusionDuration)

	// send a third tx
	receipt = sendTx()

	// wait until the block the tx was first included in shows up in the safe chain on the verifier
	_, err = waitForBlock(receipt.BlockNumber, l2Verif, safeBlockInclusionDuration)
	require.Nil(t, err, "Waiting for block on verifier")

	// ensure that the safe chain advances after restarting the batcher
	newSeqStatus, err = rollupClient.SyncStatus(context.Background())
	require.Nil(t, err)
	require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain did not advance after batcher was restarted")
}

func safeAddBig(a *big.Int, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

func latestBlock(t *testing.T, client *ethclient.Client) uint64 {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	blockAfter, err := client.BlockNumber(ctx)
	require.Nil(t, err, "Error getting latest block")
	return blockAfter
}
