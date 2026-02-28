package limb

import (
	"strings"
	"testing"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// In a pre-Arsia environment, eth_estimateTotalFee must be rejected.
func TestEstimateTotalFee_PreArsiaRejected(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	//if sys.L2Chain.IsMantleForkActive(forks.MantleArsia) {
	//	t.Skip("Arsia is active in current gate; pre-Arsia rejection semantics are not applicable")
	//}
	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleLimb), "Limb fork must be active for this test")
	if sys.L2Chain.IsMantleForkActive(forks.MantleArsia) {
		t.Skip("Arsia is active in current gate; pre-Arsia rejection semantics are not applicable")
	}

	alice := sys.FunderL2.NewFundedEOA(eth.OneEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)

	var result hexutil.Big
	err := sys.L2EL.Escape().EthClient().RPC().CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
		"from":  alice.Address(),
		"to":    bob.Address(),
		"value": "0x1",
	}, "latest")
	require.Error(err)
	lowerErr := strings.ToLower(err.Error())
	require.True(
		strings.Contains(lowerErr, "arsia") ||
			(strings.Contains(lowerErr, "not supported") && strings.Contains(lowerErr, "estimate")),
		"unexpected pre-Arsia error: %v", err,
	)
}
