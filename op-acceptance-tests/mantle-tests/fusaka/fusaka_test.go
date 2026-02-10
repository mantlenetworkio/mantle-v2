package fusaka

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/tests/interop/loadtest"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txinclude"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func TestSafeHeadAdvancesAfterOsaka(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	l1Config := sys.L1Network.Escape().ChainConfig()
	t.Log("Waiting for Osaka to activate")
	t.Require().NotNil(l1Config.OsakaTime)
	sys.L1EL.WaitForTime(*l1Config.OsakaTime)
	t.Log("Osaka activated")

	l2BlockTime := time.Duration(sys.L2Chain.Escape().RollupConfig().BlockTime) * time.Second
	for {
		l2SafeRef := sys.L2EL.BlockRefByLabel(eth.Safe)
		if l1Config.IsOsaka(new(big.Int).SetUint64(l2SafeRef.Number), l2SafeRef.Time) {
			return
		}
		t.Log("L2 safe head predates Osaka activation on L1, waiting for it to advance...")
		select {
		case <-time.After(l2BlockTime):
		case <-t.Ctx().Done():
			t.Require().Fail("Never found a safe L2 block after Osaka activated on L1")
		}
	}
}

func TestBlobBaseFeeIsCorrectAfterBPOFork(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	t.Log("Waiting for BPO1 to activate")
	if sys.L1Network.Escape().ChainConfig().BPO1Time == nil {
		t.Skip("BPO1 time not configured on L1; skipping BPO1-specific validation")
	}
	sys.L1EL.WaitForTime(*sys.L1Network.Escape().ChainConfig().BPO1Time)
	t.Log("BPO1 activated")

	spamBlobs(t, sys) // Raise the blob base fee to make blob parameter changes visible.

	t.Log("Waiting for non trivial BPO1 block")
	l2UnsafeHash, l1BlobBaseFee := waitForNonTrivialBPO1Block(t, sys)
	t.Log("Non-trivial BPO1 block found")
	l2Info, l2Txs, err := sys.L2EL.Escape().EthClient().InfoAndTxsByHash(t.Ctx(), l2UnsafeHash)
	t.Require().NoError(err)

	t.Log("Checking the L1 blob base fee in the system deposit tx")
	blockInfo, err := derive.L1BlockInfoFromBytes(sys.L2Chain.Escape().RollupConfig(), l2Info.Time(), l2Txs[0].Data())
	t.Require().NoError(err)
	l2BlobBaseFee := blockInfo.BlobBaseFee
	t.Require().Equal(l1BlobBaseFee, l2BlobBaseFee)

	t.Log("Checking the L1 blob base fee in the L1Block contract")
	l1Block := bindings.NewL1Block(bindings.WithClient(sys.L2EL.Escape().EthClient()), bindings.WithTo(predeploys.L1BlockAddr))
	l2BlobBaseFee, err = contractio.Read(l1Block.BlobBaseFee(), t.Ctx(), func(tx *txplan.PlannedTx) {
		tx.AgainstBlock.Set(l2Info)
	})
	t.Require().NoError(err)
	t.Require().Equal(l1BlobBaseFee, l2BlobBaseFee)
}

// TestBlobBaseFeePlumbing checks that the L1 blob base fee is correctly surfaced in L2 system data.
// It does not require BPO1 time to be configured and only validates the L1->L2 plumbing.
func TestBlobBaseFeePlumbing(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)

	spamBlobs(t, sys)

	t.Log("Waiting for a block with blob base fee")
	l2UnsafeHash, l1BlobBaseFee := waitForL1BlobBaseFee(t, sys)
	t.Log("Found block with blob base fee")

	l2Info, l2Txs, err := sys.L2EL.Escape().EthClient().InfoAndTxsByHash(t.Ctx(), l2UnsafeHash)
	t.Require().NoError(err)

	t.Log("Checking the L1 blob base fee in the system deposit tx")
	blockInfo, err := derive.L1BlockInfoFromBytes(sys.L2Chain.Escape().RollupConfig(), l2Info.Time(), l2Txs[0].Data())
	t.Require().NoError(err)
	l2BlobBaseFee := blockInfo.BlobBaseFee
	t.Require().Equal(l1BlobBaseFee, l2BlobBaseFee)

	t.Log("Checking the L1 blob base fee in the L1Block contract")
	l1Block := bindings.NewL1Block(bindings.WithClient(sys.L2EL.Escape().EthClient()), bindings.WithTo(predeploys.L1BlockAddr))
	l2BlobBaseFee, err = contractio.Read(l1Block.BlobBaseFee(), t.Ctx(), func(tx *txplan.PlannedTx) {
		tx.AgainstBlock.Set(l2Info)
	})
	t.Require().NoError(err)
	t.Require().Equal(l1BlobBaseFee, l2BlobBaseFee)
}

// waitForNonTrivialBPO1Block will return an L1 blob base fee that can only be calculated using the
// correct BPO1 parameters (i.e., the Osaka parameters result in a different value). It also
// returns an L2 block hash from the same epoch.
func waitForNonTrivialBPO1Block(t devtest.T, sys *presets.MantleMinimal) (common.Hash, *big.Int) {
	l1ChainConfig := sys.L1Network.Escape().ChainConfig()
	l1BlockTime := estimateL1BlockTimeSafe(t, sys)
	for {
		l2UnsafeRef := sys.L2CL.SyncStatus().UnsafeL2

		l1Info, _, err := sys.L1EL.EthClient().InfoAndTxsByHash(t.Ctx(), l2UnsafeRef.L1Origin.Hash)
		if errors.Is(err, ethereum.NotFound) { // Possible reorg, try again.
			continue
		}
		t.Require().NoError(err)

		// Calculate expected blob base fee with old Osaka parameters.
		osakaBlobBaseFee := eip4844.CalcBlobFee(l1ChainConfig, &types.Header{
			Time:          *l1ChainConfig.OsakaTime,
			ExcessBlobGas: l1Info.ExcessBlobGas(),
		})

		// Calculate expected blob base fee with new BPO1 parameters.
		bpo1BlobBaseFee := eip4844.CalcBlobFee(l1ChainConfig, &types.Header{
			Time:          l1Info.Time(),
			ExcessBlobGas: l1Info.ExcessBlobGas(),
		})

		if bpo1BlobBaseFee.Cmp(osakaBlobBaseFee) != 0 {
			return l2UnsafeRef.Hash, bpo1BlobBaseFee
		}

		select {
		case <-t.Ctx().Done():
			t.Require().Fail("context canceled before finding a block with a divergent base fee")
		case <-time.After(l1BlockTime):
		}
	}
}

func waitForL1BlobBaseFee(t devtest.T, sys *presets.MantleMinimal) (common.Hash, *big.Int) {
	l1ChainConfig := sys.L1Network.Escape().ChainConfig()
	l1BlockTime := estimateL1BlockTimeSafe(t, sys)

	for {
		l2UnsafeRef := sys.L2CL.SyncStatus().UnsafeL2
		l1Info, _, err := sys.L1EL.EthClient().InfoAndTxsByHash(t.Ctx(), l2UnsafeRef.L1Origin.Hash)
		if errors.Is(err, ethereum.NotFound) {
			continue
		}
		t.Require().NoError(err)

		l1BlobBaseFee := l1Info.BlobBaseFee(l1ChainConfig)
		if l1BlobBaseFee != nil {
			return l2UnsafeRef.Hash, l1BlobBaseFee
		}

		select {
		case <-t.Ctx().Done():
			t.Require().Fail("context canceled before finding a block with blob base fee")
		case <-time.After(l1BlockTime):
		}
	}
}

func spamBlobs(t devtest.T, sys *presets.MantleMinimal) {
	l1BlockTime := estimateL1BlockTimeSafe(t, sys)
	l1ChainConfig := sys.L1Network.Escape().ChainConfig()

	eoa := sys.FunderL1.NewFundedEOA(eth.OneTenthEther)
	signer := txinclude.NewPkSigner(eoa.Key().Priv(), sys.L1Network.ChainID().ToBig())
	l1ETHClient := sys.L1EL.EthClient()
	syncEOA := loadtest.NewSyncEOA(txinclude.NewPersistent(signer, struct {
		*txinclude.Monitor
		*txinclude.Resubmitter
	}{
		txinclude.NewMonitor(l1ETHClient, l1BlockTime),
		txinclude.NewResubmitter(l1ETHClient, l1BlockTime),
	}), eoa.Plan())

	var blob eth.Blob
	_, err := rand.Read(blob[:])
	t.Require().NoError(err)
	// get the field-elements into a valid range
	for i := range 4096 {
		blob[32*i] &= 0b0011_1111
	}

	const maxBlobTxsPerAccountInMempool = 16 // Private policy param in geth.
	spammer := loadtest.SpammerFunc(func(t devtest.T) error {
		_, err := syncEOA.Include(t, txplan.WithBlobs([]*eth.Blob{&blob}, l1ChainConfig), txplan.WithTo(&common.Address{}))
		return err
	})
	txsPerSlot := calcBlobTxsPerSlot(t, l1ChainConfig, maxBlobTxsPerAccountInMempool)
	schedule := loadtest.NewConstant(l1BlockTime, loadtest.WithBaseRPS(uint64(txsPerSlot)))

	ctx, cancel := context.WithCancel(t.Ctx())
	var wg sync.WaitGroup
	t.Cleanup(func() {
		cancel()
		wg.Wait()
	})
	wg.Add(1)
	go func() {
		defer wg.Done()
		schedule.Run(t.WithCtx(ctx), spammer)
	}()
}

func calcBlobTxsPerSlot(t devtest.T, l1ChainConfig *params.ChainConfig, maxBlobTxsPerAccountInMempool int) int {
	if l1ChainConfig.BlobScheduleConfig == nil || l1ChainConfig.BlobScheduleConfig.BPO1 == nil {
		t.Log("BPO1 blob schedule not configured; using minimal blob tx rate")
		return 1
	}
	return min(l1ChainConfig.BlobScheduleConfig.BPO1.Max*3/4, maxBlobTxsPerAccountInMempool)
}

func estimateL1BlockTimeSafe(t devtest.T, sys *presets.MantleMinimal) time.Duration {
	client := sys.L1EL.EthClient()
	for {
		latest, err := client.BlockRefByLabel(t.Ctx(), eth.Unsafe)
		t.Require().NoError(err)
		if latest.Number > 1 {
			lowerNum := latest.Number - 1
			if latest.Number > 1000 {
				lowerNum = latest.Number - 1000
			}
			if lowerNum == 0 {
				lowerNum = 1
			}
			lowerBlock, err := client.BlockRefByNumber(t.Ctx(), lowerNum)
			t.Require().NoError(err)
			deltaTime := latest.Time - lowerBlock.Time
			deltaNum := latest.Number - lowerBlock.Number
			if deltaNum == 0 {
				return 2 * time.Second
			}
			return time.Duration(deltaTime) * time.Second / time.Duration(deltaNum)
		}
		select {
		case <-time.After(time.Millisecond * 500):
		case <-t.Ctx().Done():
			t.Require().Fail("context was canceled before L1 block time could be estimated")
		}
	}
}
