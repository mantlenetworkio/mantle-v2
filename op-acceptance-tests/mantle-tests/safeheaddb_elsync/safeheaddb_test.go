package safeheaddb_elsync

import (
	"os"
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

func TestTruncateDatabaseOnELResync(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewSingleChainMultiNode(t)

	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalSafe, 1, 30),
		sys.L2CLB.AdvancedFn(types.LocalSafe, 1, 30))

	sys.L2CLB.Matched(sys.L2CL, types.LocalSafe, 30)
	sys.L2CLB.VerifySafeHeadDatabaseMatches(sys.L2CL)

	// Stop the verifier node. Since the sysgo EL uses in-memory storage this also wipes its database.
	// With the EL reset to genesis, when the CL restarts it will use EL sync to resync the chain rather than
	// deriving it from L1.
	sys.L2ELB.Stop()
	sys.L2CLB.Stop()

	sys.L2CL.Advanced(types.LocalSafe, 3, 30)

	sys.L2ELB.Start()
	sys.L2CLB.Start()
	sys.L2ELB.PeerWith(sys.L2EL)

	sys.L2CLB.Matched(sys.L2CL, types.LocalSafe, 30)
	sys.L2CLB.Advanced(types.LocalSafe, 1, 30) // At least one safe head db update after resync

	sys.L2CLB.VerifySafeHeadDatabaseMatches(sys.L2CL)
}

// TestNotTruncateDatabaseOnRestartWithExistingDatabase verifies that restarting the CL (op-node)
// while the EL retains its chain data does NOT truncate the safe head DB on geth.
//
// On reth (SupportsPostFinalizationELSync=true), the CL always triggers EL sync even when the EL
// has existing data, so safe head DB truncation is expected. That scenario is covered by
// TestTruncateDatabaseOnCLRestartWithReth below.
func TestNotTruncateDatabaseOnRestartWithExistingDatabase(gt *testing.T) {
	t := devtest.SerialT(gt)

	if os.Getenv("DEVSTACK_L2EL_KIND") == "op-reth" {
		t.Skip("reth always triggers EL sync on CL restart (SupportsPostFinalizationELSync=true); " +
			"see TestTruncateDatabaseOnCLRestartWithReth for reth-specific coverage")
	}

	sys := presets.NewSingleChainMultiNode(t)

	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalSafe, 1, 30),
		sys.L2CLB.AdvancedFn(types.LocalSafe, 1, 30))
	sys.L2CLB.Matched(sys.L2CL, types.LocalSafe, 30)

	preRestartSafeBlock := sys.L2CLB.SafeL2BlockRef().Number
	sys.L2CLB.VerifySafeHeadDatabaseMatches(sys.L2CL, dsl.WithMinRequiredL2Block(preRestartSafeBlock))

	// Restart the verifier op-node, but not the EL so the existing chain data is not deleted.
	sys.L2CLB.Stop()

	sys.L2CL.Advanced(types.LocalSafe, 3, 30)

	sys.L2CLB.Start()

	sys.L2CLB.Matched(sys.L2CL, types.LocalSafe, 30)
	sys.L2CLB.Advanced(types.LocalSafe, 1, 30) // At least one safe head db update after resync

	sys.L2CLB.VerifySafeHeadDatabaseMatches(sys.L2CL, dsl.WithMinRequiredL2Block(preRestartSafeBlock))
}

// TestTruncateDatabaseOnCLRestartWithReth verifies that on reth, restarting only the CL while
// the EL retains chain data still triggers EL sync (because SupportsPostFinalizationELSync=true),
// truncates the safe head DB, and then correctly rebuilds it to match the sequencer.
func TestTruncateDatabaseOnCLRestartWithReth(gt *testing.T) {
	t := devtest.SerialT(gt)

	if os.Getenv("DEVSTACK_L2EL_KIND") != "op-reth" {
		t.Skip("this test covers reth-specific EL sync behavior (SupportsPostFinalizationELSync=true)")
	}

	sys := presets.NewSingleChainMultiNode(t)

	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalSafe, 1, 30),
		sys.L2CLB.AdvancedFn(types.LocalSafe, 1, 30))
	sys.L2CLB.Matched(sys.L2CL, types.LocalSafe, 30)
	sys.L2CLB.VerifySafeHeadDatabaseMatches(sys.L2CL)

	// Restart only the CL. The EL keeps its data, but reth's SupportsPostFinalizationELSync=true
	// causes op-node to enter EL sync anyway, truncating the safe head DB.
	sys.L2CLB.Stop()

	sys.L2CL.Advanced(types.LocalSafe, 3, 30)

	sys.L2CLB.Start()

	// Verify the system recovers: verifier catches up and safe head DB is rebuilt correctly.
	sys.L2CLB.Matched(sys.L2CL, types.LocalSafe, 30)
	sys.L2CLB.Advanced(types.LocalSafe, 1, 30)

	sys.L2CLB.VerifySafeHeadDatabaseMatches(sys.L2CL)
}
