package proofs

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type activationBlockTestCfg struct {
	fork          rollup.MantleForkName
	numUpgradeTxs int
}

// TestUpgradeBlockTxOmission tests that the sequencer omits user transactions in activation blocks
// and that batches that contain user transactions in an activation block are dropped.
func TestActivationBlockTxOmission(gt *testing.T) {
	matrix := helpers.NewMatrix[activationBlockTestCfg]()

	matrix.AddDefaultTestCasesWithName(
		"MantleArsia",
		activationBlockTestCfg{fork: rollup.MantleForkName("MantleArsia"), numUpgradeTxs: 7},
		helpers.MantleArsiaOnly(),
		testActivationBlockTxOmission,
	)
	// New forks should be added here in the future.

	matrix.Run(gt)
}

func testActivationBlockTxOmission(gt *testing.T, testCfg *helpers.TestCfg[activationBlockTestCfg]) {
	tcfg := testCfg.Custom
	t := actionsHelpers.NewDefaultTesting(gt)
	offset := uint64(48)
	testSetup := func(dc *genesis.DeployConfig) {
		dc.L1GenesisBlockTimestamp = hexutil.Uint64(uint64(time.Now().Unix()))
		dc.L1PragueTimeOffset = ptr(hexutil.Uint64(0))
		// activate fork after a few blocks
		dc.ActivateMantleForkAtOffset(tcfg.fork, offset)
	}
	env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(), helpers.NewBatcherCfg(), testSetup)

	engine := env.Engine
	sequencer := env.Sequencer
	miner := env.Miner
	rollupCfg := env.Sd.RollupCfg
	//blockTime := rollupCfg.BlockTime

	miner.ActEmptyBlock(t)

	sequencer.ActL1HeadSignal(t)
	for i := 0; i < int(offset)-1; i++ {
		sequencer.ActL2EmptyBlock(t)
	}

	// Manually create a transaction to avoid gas estimation issues in activation block
	env.Alice.L2.ActResetTxOpts(t)
	nonce := env.Alice.L2.PendingNonce(t)

	// Get latest header for gas price info
	latestHeader, err := engine.EthClient().HeaderByNumber(t.Ctx(), nil)
	require.NoError(t, err)
	gasTipCap := big.NewInt(2 * params.GWei)
	gasFeeCap := new(big.Int).Add(gasTipCap, new(big.Int).Mul(latestHeader.BaseFee, big.NewInt(2)))

	toAddr := env.Alice.Address()
	tx := types.MustSignNewTx(env.Alice.L2.Secret(), env.Alice.L2.Signer(), &types.DynamicFeeTx{
		ChainID:   env.Alice.L2.Signer().ChainID(),
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       50_000, // Manually set gas limit to avoid estimation
		To:        &toAddr,
		Value:     big.NewInt(0),
		Data:      []byte{},
	})

	sequencer.ActL2StartBlock(t)
	// we assert later that the sequencer actually omits this tx in the activation block
	engine.ActL2IncludeTx(env.Alice.Address())
	sequencer.ActL2EndBlock(t)

	actHeader := engine.L2Chain().CurrentHeader()
	// For Mantle forks, use the Mantle-specific activation check
	// because OP Stack forks (Canyon, Delta, etc.) are mapped to MantleArsia
	//isMantleArsiaActivation := rollupCfg.IsMantleArsiaActivationBlock(actHeader.Time)
	var isMantleForkActivation bool
	switch tcfg.fork {
	case rollup.MantleForkName("MantleArsia"):
		isMantleForkActivation = rollupCfg.IsMantleArsiaActivationBlock(actHeader.Time)
	case rollup.MantleForkName("MantleLimb"):
		isMantleForkActivation = rollupCfg.IsMantleLimbActivationBlock(actHeader.Time)
	// Add new forks here in the future.
	default:
		t.Fatalf("unsupported fork: %s", tcfg.fork)
	}
	require.True(t, isMantleForkActivation, "this block should be the MantleArsia activation block")
	actBlock := engine.L2Chain().GetBlockByHash(actHeader.Hash())
	require.Len(t, actBlock.Transactions(), tcfg.numUpgradeTxs+1, "activation block contains unexpected txs")

	batcher := env.Batcher
	t.Logf("=== Batcher Submit Time ===")
	t.Logf("Real time.Now(): %d", time.Now().Unix())
	t.Logf("Genesis L2Time: %d", env.Sequencer.RollupCfg.Genesis.L2Time)
	t.Logf("MantleArsiaTime: %d", *env.Sequencer.RollupCfg.MantleArsiaTime)
	t.Logf("EcotoneTime: %d", *env.Sequencer.RollupCfg.EcotoneTime)
	t.Logf("IsMantleArsia(time.Now()): %v", env.Sequencer.RollupCfg.IsMantleArsia(uint64(time.Now().Unix())))
	for i := 0; i < int(offset)-1; i++ {
		batcher.ActL2BatchBuffer(t)
	}
	batcher.ActL2BatchBuffer(t,
		actionsHelpers.WithBlockModifier(func(block *types.Block) *types.Block {
			// inject user tx into activation batch
			return block.WithBody(types.Body{Transactions: append(block.Transactions(), tx)})
		}))

	batcher.ActL2ChannelClose(t)
	batcher.ActL2BatchSubmitMantle(t)
	env.Miner.ActL1StartBlock(12)(t)
	env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
	env.Miner.ActL1EndBlock(t)

	env.Sequencer.ActL1HeadSignal(t)
	env.Sequencer.ActL2PipelineFull(t)

	recs := env.Logs.FindLogs(testlog.NewMessageFilter("dropping batch with user transactions in fork activation block"))
	require.Len(t, recs, 1)

	l2SafeHead := engine.L2Chain().CurrentSafeBlock()
	preActHeader := engine.L2Chain().GetHeaderByHash(actHeader.ParentHash)
	require.Equal(t, eth.HeaderBlockID(preActHeader), eth.HeaderBlockID(l2SafeHead), "derivation only reaches pre-upgrade block")

	//env.RunFaultProofProgramFromGenesis(t, l2SafeHead.Number.Uint64(), testCfg.CheckResult, testCfg.InputParams...)
}
