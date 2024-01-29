package genesis

import "testing"

func TestPostCheckLegacyMNT(t *testing.T) {
	l2ChainID := uint64(5001)
	legacySlot := LegacyMNTCheckSlots[l2ChainID]
	if legacySlot == nil {
		legacySlot = LegacyMNTCheckSlots[0]
	}
	t.Log(legacySlot)
}
