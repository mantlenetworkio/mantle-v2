package skadi

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithMantleMinimal(),
		presets.WithMantleSkadiAtGenesis(),
		// Restrict to persistent (sysext) environments. Under sysgo, this preset's
		// GasLimit (2^50) exceeds SystemConfig.MAX_GAS_LIMIT (500_000_000, added by
		// PR #330), so SystemConfig.initialize() reverts; TestMain then gracefully
		// skips. compat.Persistent makes the intent explicit instead of relying on
		// that runtime skip.
		presets.WithCompatibleTypes(compat.Persistent),
	)
}
