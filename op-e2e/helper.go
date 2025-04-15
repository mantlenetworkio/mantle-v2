package op_e2e

import (
	"flag"
	"os"
	"testing"
)

var enableParallelTesting bool = true

// Init testing to enable test flags
var _ = func() bool {
	testing.Init()
	return true
}()

var verboseGethNodes bool

func init() {
	flag.BoolVar(&verboseGethNodes, "gethlogs", true, "Enable logs on geth nodes")
	flag.Parse()
	if os.Getenv("OP_E2E_DISABLE_PARALLEL") == "true" {
		enableParallelTesting = false
	}
}

func InitParallel(t *testing.T) {
	t.Helper()
	if enableParallelTesting {
		t.Parallel()
	}
}
