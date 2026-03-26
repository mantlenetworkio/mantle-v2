// op-acceptance-tests/tests/custom_gas_token/helpers.go
package custom_gas_token

import (
	"context"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lmittmann/w3"
)

var (
	// L2 predeploy: L1Block (address is stable across OP Stack chains)
	l1BlockAddr = common.HexToAddress("0x4200000000000000000000000000000000000015")
)

type minimalLike interface {
	L2ELNode() *dsl.L2ELNode
}

// isCGTEnabled checks if CGT mode is enabled without skipping the test.
// Returns true if CGT is enabled, false if native ETH mode, and false if the check fails.
func isCGTEnabled(t devtest.T, sys minimalLike) bool {
	l2 := sys.L2ELNode().Escape().L2EthClient()
	isCustomGasTokenFunc := w3.MustNewFunc("isCustomGasToken()", "bool")

	ctx, cancel := context.WithTimeout(t.Ctx(), 20*time.Second)
	defer cancel()

	data, _ := isCustomGasTokenFunc.EncodeArgs()
	out, err := l2.Call(ctx, ethereum.CallMsg{To: &l1BlockAddr, Data: data}, rpc.LatestBlockNumber)
	if err != nil {
		return false
	}

	var isCustom bool
	if err := isCustomGasTokenFunc.DecodeReturns(out, &isCustom); err != nil {
		return false
	}

	return isCustom
}

// SkipIfCGT probes L2 L1Block for CGT mode. If CGT is enabled, the test is skipped.
// This is useful for tests that should only run with native ETH.
func SkipIfCGT(t devtest.T, sys minimalLike) {
	if isCGTEnabled(t, sys) {
		t.Skip("Test skipped: CGT is enabled (test requires native ETH)")
	}
}
