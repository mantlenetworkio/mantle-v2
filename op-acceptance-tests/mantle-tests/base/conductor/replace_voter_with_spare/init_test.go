package replace_voter_with_spare

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithCompatibleTypes(compat.SysGo),
		presets.WithMantleMinimalWithSpareConductor(),
	)
}
