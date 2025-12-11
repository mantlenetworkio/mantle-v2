package jovian

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

func TestMain(m *testing.M) {
	presets.DoMain(m, presets.WithMantleMinimal(), presets.WithMantleArsiaAtGenesis())
}
