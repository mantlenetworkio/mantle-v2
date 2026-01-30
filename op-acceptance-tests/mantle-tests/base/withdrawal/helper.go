package withdrawal

import (
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
)

var bvmETHAddr = common.HexToAddress("0xdEAddEaDdeadDEadDEADDEAddEADDEAddead1111")

func waitForL1TimeAfter(t devtest.T, sys *presets.MantleMinimal, delay time.Duration) {
	head := sys.L1EL.BlockRefByLabel(eth.Unsafe)
	target := head.Time + uint64(delay.Seconds())
	t.Require().Eventually(func() bool {
		head = sys.L1EL.BlockRefByLabel(eth.Unsafe)
		return head.Time >= target
	}, sys.L1EL.TransactionTimeout(), time.Second, "L1 time did not advance past finalization window")
}
