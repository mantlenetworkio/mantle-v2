package withdrawal

import (
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"testing"
)

func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithCompatibleTypes(compat.SysGo),
		presets.WithMantleMinimal(),
		presets.WithTimeTravel(),
		presets.WithFinalizationPeriodSeconds(1),
	)
}
