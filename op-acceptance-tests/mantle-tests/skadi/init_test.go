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
		// The Arsia version of the SystemConfig contract does not support high gas limit parameters in pre-Arsia.
		presets.WithCompatibleTypes(compat.Persistent),
	)
}
