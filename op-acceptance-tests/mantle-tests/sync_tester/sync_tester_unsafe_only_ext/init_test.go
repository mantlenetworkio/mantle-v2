package sync_tester_unsafe_only_ext

import (
	"fmt"
	"os"
	"testing"

	synctester "github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/sync_tester"
	bss "github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

func TestMain(m *testing.M) {
	config, err := synctester.GetNetworkPresetFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load network preset from env: %v\n", err)
		os.Exit(1)
	}
	if config.RollupConfig == nil {
		fmt.Fprintf(os.Stderr, "rollup config is required for unsafe-only test\n")
		os.Exit(1)
	}
	genesisL2Number := config.RollupConfig.Genesis.L2.Number

	presets.DoMain(m,
		presets.WithExternalEL(config),
		// CL connected to sync tester EL is verifier
		presets.WithExecutionLayerSyncOnVerifiers(),
		// Make sync tester EL mock EL Sync
		presets.WithELSyncActive(),
		// Only rely on EL sync for unsafe gap filling
		presets.WithReqRespSyncDisabled(),
		presets.WithNoDiscovery(),
		presets.WithCompatibleTypes(compat.SysGo),
		presets.WithUnsafeOnly(),
		stack.MakeCommon(sysgo.WithBatcherOption(func(id stack.L2BatcherID, cfg *bss.CLIConfig) {
			// For stopping derivation, not to advance safe heads
			cfg.Stopped = true
		})),
		// Sync tester EL at genesis
		presets.WithSyncTesterELInitialState(eth.FCUState{
			Latest:    genesisL2Number,
			Safe:      genesisL2Number,
			Finalized: genesisL2Number,
		}),
	)
}
