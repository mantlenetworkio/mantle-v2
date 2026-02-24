package conductor

import (
	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithCompatibleTypes(compat.SysGo),
		presets.WithMantleMinimalWithConductors(),
	)
}
