package elfaultinjector

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/compat"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

// TestMain registers the sysgo-only "Mantle multinode + Engine API
// fault-injection proxy" preset for every test in this package.
//
// The preset wires an elfaultinjector.Proxy in front of every L2 EL's
// Engine API. The proxies start in pure pass-through mode; tests must
// explicitly call EngineFaultInjectors[id].Activate(rule) to begin
// synthesizing INVALID PayloadStatusV1 responses for matching
// engine_newPayloadV{3,4} requests.
func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithMantleFaultyMultiNode(),
		presets.WithCompatibleTypes(compat.SysGo),
	)
}
