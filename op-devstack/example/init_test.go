package example

import (
	"testing"
)

// TestMain creates the test-setups against the shared backend
func TestMain(m *testing.M) {
	// Skipping example test due to mismatched deploy script between op and mantle.
	// presets.DoMain(m, presets.WithSimpleInterop(),
	// 	// Logging can be adjusted with filters globally
	// 	presets.WithPkgLogFilter(
	// 		logfilter.DefaultShow( // Random configuration
	// 			stack.KindSelector(stack.L2ProposerKind).Mute(),
	// 			stack.KindSelector(stack.L2BatcherKind).And(logfilter.Level(log.LevelError)).Show(),
	// 			stack.KindSelector(stack.L2CLNodeKind).Mute(),
	// 		),
	// 		// E.g. allow test interactions through while keeping background resource logs quiet
	// 	),
	// 	presets.WithTestLogFilter(logfilter.DefaultMute(logfilter.Level(log.LevelInfo).Show())),
	// )
}
