package config

import "testing"

// Mantle e2e init() uses initMantleAllocType only. initAllocType is the vanilla OP Stack alloc
// generator kept for upstream parity.
func TestInitAllocType_keptForUpstreamParity(t *testing.T) {
	t.Helper()
	var f func(string, AllocType) = initAllocType
	_ = f
}
