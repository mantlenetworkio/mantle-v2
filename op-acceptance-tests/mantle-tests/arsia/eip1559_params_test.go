package arsia

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum-optimism/optimism/op-service/txplan"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type eip1559ParamsEnv struct {
	l1Client     *dsl.L1ELNode
	l2Network    *dsl.L2Network
	l2EL         *dsl.L2ELNode
	systemConfig eip1559ParamsSystemConfig
}

type eip1559ParamsSystemConfig struct {
	SetEIP1559Params   func(denominator uint32, elasticity uint32) bindings.TypedCall[any] `sol:"setEIP1559Params"`
	Eip1559Denominator func() bindings.TypedCall[uint32]                                   `sol:"eip1559Denominator"`
	Eip1559Elasticity  func() bindings.TypedCall[uint32]                                   `sol:"eip1559Elasticity"`
}

func newEIP1559Params(t devtest.T, l2Network *dsl.L2Network, l1EL *dsl.L1ELNode, l2EL *dsl.L2ELNode) *eip1559ParamsEnv {
	systemConfig := bindings.NewBindings[eip1559ParamsSystemConfig](
		bindings.WithClient(l1EL.EthClient()),
		bindings.WithTo(l2Network.Escape().Deployment().SystemConfigProxyAddr()),
		bindings.WithTest(t))

	return &eip1559ParamsEnv{
		l1Client:     l1EL,
		l2Network:    l2Network,
		l2EL:         l2EL,
		systemConfig: systemConfig,
	}
}

func (ep *eip1559ParamsEnv) checkCompatibility(t devtest.T) {
	_, err := contractio.Read(ep.systemConfig.Eip1559Denominator(), t.Ctx())
	if err != nil {
		t.Fail()
	}
}

func (ep *eip1559ParamsEnv) getSystemConfigOwner(t devtest.T) *dsl.EOA {
	priv := ep.l2Network.Escape().Keys().Secret(devkeys.SystemConfigOwner.Key(ep.l2Network.ChainID().ToBig()))
	return dsl.NewKey(t, priv).User(ep.l1Client)
}

func (ep *eip1559ParamsEnv) ownerPlan(t devtest.T) txplan.Option {
	owner := ep.getSystemConfigOwner(t)
	t.Log("system config owner during test", owner.Address())

	elClient := ep.l1Client.Escape().EthClient()
	return txplan.Combine(
		owner.Plan(),
		txplan.WithRetryInclusion(elClient, 10, retry.Exponential()),
	)
}

func (ep *eip1559ParamsEnv) setEIP1559ParamsOnL1(t devtest.T, denominator, elasticity uint32) {
	receipt, err := contractio.Write(ep.systemConfig.SetEIP1559Params(denominator, elasticity), t.Ctx(), ep.ownerPlan(t))
	t.Require().NoError(err, "SetEIP1559Params transaction failed")

	t.Log("tx hash", "tx hash", receipt.TxHash)
	t.Logf("Set EIP-1559 params on L1: denominator=%d, elasticity=%d", denominator, elasticity)
}

func (ep *eip1559ParamsEnv) readL1EIP1559Params(t devtest.T) (denominator, elasticity uint32) {
	denom, err := contractio.Read(ep.systemConfig.Eip1559Denominator(), t.Ctx())
	t.Require().NoError(err, "reading denominator from L1")

	elast, err := contractio.Read(ep.systemConfig.Eip1559Elasticity(), t.Ctx())
	t.Require().NoError(err, "reading elasticity from L1")

	return denom, elast
}

// waitForEIP1559ParamsOnL2 waits until the L2 block header extradata encodes
// the expected denominator and elasticity.
func (ep *eip1559ParamsEnv) waitForEIP1559ParamsOnL2(t devtest.T, expectedDenom, expectedElasticity uint32) {
	client := ep.l2EL.Escape().L2EthClient()

	t.Require().Eventually(func() bool {
		info, err := client.InfoByLabel(t.Ctx(), "latest")
		if err != nil {
			return false
		}

		headerRLP, err := info.HeaderRLP()
		if err != nil {
			return false
		}

		var header types.Header
		if err := rlp.DecodeBytes(headerRLP, &header); err != nil {
			return false
		}

		// Extradata must be at least 9 bytes: version(1) + denominator(4) + elasticity(4)
		if len(header.Extra) < 9 {
			return false
		}

		gotDenom := binary.BigEndian.Uint32(header.Extra[1:5])
		gotElasticity := binary.BigEndian.Uint32(header.Extra[5:9])
		return gotDenom == expectedDenom && gotElasticity == expectedElasticity
	}, 3*time.Minute, 5*time.Second, "L2 EIP-1559 params in block header did not sync within timeout")
}

// readL2EIP1559Params returns the denominator and elasticity currently encoded
// in the latest L2 block header extradata.
func (ep *eip1559ParamsEnv) readL2EIP1559Params(t devtest.T) (denominator, elasticity uint32) {
	client := ep.l2EL.Escape().L2EthClient()

	info, err := client.InfoByLabel(t.Ctx(), "latest")
	t.Require().NoError(err)

	headerRLP, err := info.HeaderRLP()
	t.Require().NoError(err)

	var header types.Header
	t.Require().NoError(rlp.DecodeBytes(headerRLP, &header))
	t.Require().True(len(header.Extra) >= 9, "block header extradata too short: %d bytes", len(header.Extra))

	return binary.BigEndian.Uint32(header.Extra[1:5]), binary.BigEndian.Uint32(header.Extra[5:9])
}

// TestEIP1559Params sets different EIP-1559 denominator and elasticity values
// on L1 via SystemConfig and verifies they propagate to L2 block headers.
func TestEIP1559Params(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia), "Arsia fork must be active for this test")

	ep := newEIP1559Params(t, sys.L2Chain, sys.L1EL, sys.L2EL)
	ep.checkCompatibility(t)

	systemOwner := ep.getSystemConfigOwner(t)
	sys.FunderL1.FundAtLeast(systemOwner, eth.OneHundredthEther)

	// Determine expected default EIP-1559 params from the rollup config.
	// When ChainOpConfig (the Optimism field) is set, use its values;
	// otherwise fall back to the hardcoded defaults (denominator=8, elasticity=2).
	var expectedDefaultDenom, expectedDefaultElasticity uint32
	rollupCfg := sys.L2Chain.Escape().RollupConfig()
	if rollupCfg.ChainOpConfig != nil {
		expectedDefaultDenom = uint32(rollupCfg.ChainOpConfig.EIP1559Denominator)
		expectedDefaultElasticity = uint32(rollupCfg.ChainOpConfig.EIP1559Elasticity)
	} else {
		expectedDefaultDenom = 8
		expectedDefaultElasticity = 2
	}
	t.Logf("Expected default EIP-1559 params from rollup config: denominator=%d, elasticity=%d",
		expectedDefaultDenom, expectedDefaultElasticity)

	// Before any changes: L1 SystemConfig should have both values at zero (initial state),
	// while L2 block headers should carry the defaults from the rollup config.
	origDenom, origElasticity := ep.readL1EIP1559Params(t)
	t.Logf("Initial L1 SystemConfig EIP-1559 params: denominator=%d, elasticity=%d", origDenom, origElasticity)
	require.Equal(uint32(0), origDenom, "initial L1 denominator should be zero")
	require.Equal(uint32(0), origElasticity, "initial L1 elasticity should be zero")

	l2Denom, l2Elast := ep.readL2EIP1559Params(t)
	t.Logf("Initial L2 block header EIP-1559 params: denominator=%d, elasticity=%d", l2Denom, l2Elast)
	require.Equal(expectedDefaultDenom, l2Denom, "L2 denominator should match rollup config default")
	require.Equal(expectedDefaultElasticity, l2Elast, "L2 elasticity should match rollup config default")

	testCases := []struct {
		name        string
		denominator uint32
		elasticity  uint32
	}{
		{"LargeDenominator", 50, 6},
		{"SmallDenominator", 2, 2},
		{"DefaultLike", 8, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t devtest.T) {
			ep.setEIP1559ParamsOnL1(t, tc.denominator, tc.elasticity)
			ep.waitForEIP1559ParamsOnL2(t, tc.denominator, tc.elasticity)

			t.Log("Test completed successfully:",
				"testCase", tc.name,
				"denominator", tc.denominator,
				"elasticity", tc.elasticity)
		})
	}
}
