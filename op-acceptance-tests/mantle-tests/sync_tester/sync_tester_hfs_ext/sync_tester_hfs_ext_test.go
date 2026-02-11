package sync_tester_hfs_ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/log"

	synctester "github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/sync_tester"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/stack/match"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

func TestSyncTesterHFS_Arsia_CLSync(gt *testing.T) {
	hfsExt(gt, forks.MantleArsia, sync.CLSync)
}

func TestSyncTesterHFS_Arsia_ELSync(gt *testing.T) {
	hfsExt(gt, forks.MantleArsia, sync.ELSync)
}

// forkBlockFromRollupConfig calculates the L2 block number at which a fork activates,
// using the rollup config's genesis info and block time.
func forkBlockFromRollupConfig(rollupCfg *rollup.Config, upgradeName forks.MantleForkName) (uint64, error) {
	activationTime := rollupCfg.MantleActivationTime(upgradeName)
	if activationTime == nil {
		return 0, fmt.Errorf("fork %s not activated in rollup config", upgradeName)
	}
	genesisTime := rollupCfg.Genesis.L2Time
	genesisBlock := rollupCfg.Genesis.L2.Number
	blockTime := rollupCfg.BlockTime
	if blockTime == 0 {
		return 0, fmt.Errorf("rollup config has zero block time")
	}
	if *activationTime <= genesisTime {
		return genesisBlock, nil // Fork activated at or before genesis
	}
	return genesisBlock + (*activationTime-genesisTime)/blockTime, nil
}

// setupOrchestrator initializes and configures the orchestrator for the test
func setupOrchestrator(gt *testing.T, t devtest.T, config stack.ExtNetworkConfig, blk uint64, l2CLSyncMode sync.Mode) *sysgo.Orchestrator {
	l := t.Logger()

	// Setup orchestrator
	logger := testlog.Logger(gt, log.LevelInfo)
	onFail := func(now bool) {
		if now {
			gt.FailNow()
		} else {
			gt.Fail()
		}
	}
	onSkipNow := func() {
		gt.SkipNow()
	}
	p := devtest.NewP(context.Background(), logger, onFail, onSkipNow)
	gt.Cleanup(p.Close)

	// Runtime configuration values
	l.Info("Runtime configuration values for TestSyncTesterHFS")
	l.Info("L2_NETWORK_NAME", "value", config.L2NetworkName)
	l.Info("L1_CHAIN_ID", "value", config.L1ChainID)
	l.Info("L2_EL_ENDPOINT", "value", config.L2ELEndpoint)
	l.Info("L1_CL_BEACON_ENDPOINT", "value", config.L1CLBeaconEndpoint)
	l.Info("L1_EL_ENDPOINT", "value", config.L1ELEndpoint)
	l.Info("L2_CL_SYNCMODE", "value", l2CLSyncMode)
	l.Info("Config has RollupConfig", "value", config.RollupConfig != nil)
	l.Info("Config has L2ChainConfig", "value", config.L2ChainConfig != nil)
	l.Info("Config has L1ChainConfig", "value", config.L1ChainConfig != nil)

	opt := presets.WithExternalEL(config)
	if l2CLSyncMode == sync.ELSync {
		t.Require().NotNil(config.RollupConfig, "rollup config is required for EL sync mode")
		opt = stack.Combine(opt,
			presets.WithExecutionLayerSyncOnVerifiers(),
			presets.WithELSyncActive(),
			presets.WithSyncTesterELInitialState(eth.FCUState{
				Latest: blk,
				Safe:   0,
				// Need to set finalized to genesis to unskip EL Sync
				Finalized: config.RollupConfig.Genesis.L2.Number,
			}),
		)
	} else {
		opt = stack.Combine(opt,
			presets.WithSyncTesterELInitialState(eth.FCUState{
				Latest:    blk,
				Safe:      blk,
				Finalized: blk,
			}),
		)
	}

	var orch stack.Orchestrator = sysgo.NewOrchestrator(p, stack.SystemHook(opt))
	stack.ApplyOptionLifecycle(opt, orch)

	return orch.(*sysgo.Orchestrator)
}

func hfsExt(gt *testing.T, upgradeName forks.MantleForkName, l2CLSyncMode sync.Mode) {
	t := devtest.ParallelT(gt)
	l := t.Logger()
	require := t.Require()

	config, err := synctester.GetNetworkPresetFromEnv()
	require.NoError(err, "failed to load network config from env")
	require.NotNil(config.RollupConfig, "rollup config is required for HFS tests")

	// Calculate the fork activation block from the rollup config
	forkBlock, err := forkBlockFromRollupConfig(config.RollupConfig, upgradeName)
	require.NoError(err, "failed to calculate fork block for %s", upgradeName)

	// Skip if the fork activated at or near genesis — can't sync across a boundary that doesn't exist
	if forkBlock < 5 {
		t.Skipf("fork %s activated at block %d (at/near genesis), skipping HFS boundary test", upgradeName, forkBlock)
	}

	// Start syncing from 5 blocks before the fork
	blk := forkBlock - 5
	blocksToSync := uint64(10)
	targetBlock := blk + blocksToSync

	l.Info("HFS test parameters",
		"upgrade", upgradeName,
		"fork_block", forkBlock,
		"initial_block", blk,
		"target_block", targetBlock,
		"sync_mode", l2CLSyncMode,
	)

	// Initialize orchestrator
	orch := setupOrchestrator(gt, t, config, blk, l2CLSyncMode)
	system := shim.NewSystem(t)
	orch.Hydrate(system)

	l2 := system.L2Network(match.L2ChainA)
	verifierCL := l2.L2CLNode(match.FirstL2CL)
	syncTester := l2.SyncTester(match.FirstSyncTester)

	sys := &struct {
		L2CL         *dsl.L2CLNode
		L2ELReadOnly *dsl.L2ELNode
		L2EL         *dsl.L2ELNode
		SyncTester   *dsl.SyncTester
		L2           *dsl.L2Network
	}{
		L2CL:         dsl.NewL2CLNode(verifierCL, orch.ControlPlane()),
		L2ELReadOnly: dsl.NewL2ELNode(l2.L2ELNode(match.FirstL2EL), orch.ControlPlane()),
		L2EL:         dsl.NewL2ELNode(l2.L2ELNode(match.SecondL2EL), orch.ControlPlane()),
		SyncTester:   dsl.NewSyncTester(syncTester),
		L2:           dsl.NewL2Network(l2, orch.ControlPlane()),
	}

	ft := sys.L2.Escape().RollupConfig().MantleActivationTime(upgradeName)
	var l2CLSyncStatus *eth.SyncStatus
	attempts := 1000
	if l2CLSyncMode == sync.ELSync {
		// After EL Sync is finished, the FCU state will advance to target immediately so less attempts
		attempts = 5
		// Signal L2CL for finishing EL Sync
		// Must send consecutive three payloads due to default EL Sync policy
		for i := 2; i >= 0; i-- {
			sys.L2CL.SignalTarget(sys.L2ELReadOnly, targetBlock-uint64(i))
		}
	} else {
		l2CLSyncStatus := sys.L2CL.WaitForNonZeroUnsafeTime(t.Ctx())
		require.Less(l2CLSyncStatus.UnsafeL2.Time, *ft, "L2CL unsafe time should be less than fork timestamp before upgrade")
	}

	sys.L2CL.Reached(types.LocalUnsafe, targetBlock, attempts)
	l.Info("L2CL unsafe reached", "targetBlock", targetBlock, "upgrade_name", upgradeName)
	sys.L2CL.Reached(types.LocalSafe, targetBlock, attempts)
	l.Info("L2CL safe reached", "targetBlock", targetBlock, "upgrade_name", upgradeName)

	l2CLSyncStatus = sys.L2CL.SyncStatus()
	require.NotNil(l2CLSyncStatus, "L2CL should have sync status")
	require.Greater(l2CLSyncStatus.UnsafeL2.Time, *ft, "L2CL unsafe time should be greater than fork timestamp after upgrade")

	unsafeL2Ref := l2CLSyncStatus.UnsafeL2
	ref := sys.L2EL.BlockRefByNumber(unsafeL2Ref.Number)
	require.Equal(unsafeL2Ref.Hash, ref.Hash, "L2EL should be on the same block as L2CL")

	stSessions := sys.SyncTester.ListSessions()
	require.Equal(len(stSessions), 1, "expect exactly one session")

	stSession := sys.SyncTester.GetSession(stSessions[0])
	require.GreaterOrEqualf(stSession.CurrentState.Latest, stSession.InitialState.Latest+blocksToSync, "SyncTester session CurrentState.Latest only advanced %d", stSession.CurrentState.Latest-stSession.InitialState.Latest)
	require.GreaterOrEqualf(stSession.CurrentState.Safe, stSession.InitialState.Safe+blocksToSync, "SyncTester session CurrentState.Safe only advanced %d", stSession.CurrentState.Safe-stSession.InitialState.Safe)

	l.Info("SyncTester HFS Ext test completed successfully", "l2cl_chain_id", sys.L2CL.ID().ChainID(), "l2cl_sync_status", l2CLSyncStatus, "upgrade_name", upgradeName)
}
