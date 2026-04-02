package l2backend

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/reth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/services"
)

// l2ClientOnce ensures the L2 client type is parsed at most once per process,
// preventing unpredictable behavior from env var mutations during test runs.
var (
	l2ClientOnce sync.Once
	l2ClientType string
)

// L2Client returns the configured L2 client type.
// Parsed once from OP_E2E_L2_CLIENT env var; defaults to "geth".
func L2Client() string {
	l2ClientOnce.Do(func() {
		l2ClientType = os.Getenv("OP_E2E_L2_CLIENT")
		if l2ClientType == "" {
			l2ClientType = "geth"
		}
	})
	return l2ClientType
}

// IsGeth returns true when the active L2 client is geth.
func IsGeth() bool { return L2Client() == "geth" }

// IsReth returns true when the active L2 client is reth.
func IsReth() bool { return L2Client() == "reth" }

// RethBinPath returns the op-reth binary path from OP_E2E_L2_BIN.
// If reth is selected but the binary path is not configured, calls t.Skip,
// allowing CI to gracefully skip reth tests when the binary is unavailable.
func RethBinPath(tb testing.TB) string {
	p := os.Getenv("OP_E2E_L2_BIN")
	if p == "" {
		tb.Skip("op-reth binary not configured: set OP_E2E_L2_BIN=/path/to/op-reth")
	}
	return p
}

// ReadyTimeout returns the configured RPC readiness timeout.
// Parsed from OP_E2E_L2_READY_TIMEOUT env var (Go duration string); defaults to 30s.
func ReadyTimeout(tb testing.TB) time.Duration {
	if s := os.Getenv("OP_E2E_L2_READY_TIMEOUT"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			tb.Logf("warning: invalid OP_E2E_L2_READY_TIMEOUT %q, using default 30s: %v", s, err)
			return 30 * time.Second
		}
		return d
	}
	return 30 * time.Second
}

// InitL2Backend creates an L2 EthInstance using the configured client (geth or reth).
// Returns an unstarted instance — callers must call Start() before use.
//
// This two-phase lifecycle (Init → Start) preserves the ability to configure
// the instance between creation and startup, consistent with geth's existing
// InitL2() → Start() pattern.
func InitL2Backend(
	tb testing.TB,
	name string,
	genesis *core.Genesis,
	jwtPath string,
	gethOpts []geth.GethOption,
	rethOpts []reth.RethOption,
) (services.EthInstance, error) {
	switch L2Client() {
	case "geth":
		return geth.InitL2(name, genesis, jwtPath, gethOpts...)
	case "reth":
		return reth.InitReth(tb, name, genesis, jwtPath, RethBinPath(tb), rethOpts...)
	default:
		return nil, fmt.Errorf("unsupported OP_E2E_L2_CLIENT value: %q (supported: geth, reth)", L2Client())
	}
}
