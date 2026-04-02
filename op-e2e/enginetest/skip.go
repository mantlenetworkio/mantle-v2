package enginetest

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/l2backend"
)

// SkipIfNotGeth skips the current test if the active L2 client is not geth.
// Use this for tests that exercise geth-specific features (e.g. eth_sendRawTransactionConditional)
// that are not available in reth or other alternative clients.
func SkipIfNotGeth(t testing.TB) {
	t.Helper()
	if !l2backend.IsGeth() {
		t.Skip("this test requires geth-specific features; skipping for client: " + l2backend.L2Client())
	}
}
