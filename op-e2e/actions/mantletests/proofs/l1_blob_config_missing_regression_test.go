package proofs_test

import (
	"math/big"
	"testing"

	batcherFlags "github.com/ethereum-optimism/optimism/op-batcher/flags"
	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/mantletests/proofs/helpers"
	legacybindings "github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// Test_ProgramAction_BlobConfigMissingRegression is a regression test for the bug where
// MainnetChainConfig was missing BPO1/BPO2 BlobScheduleConfig entries, causing the blob fee
// calculation to use Prague's UpdateFraction (5007716) instead of BPO2's (11684671).
// The exponential nature of the blob fee formula means this produces fees 10,000x+ higher
// than correct values.
//
// This test compares two L1 configurations:
//   - "complete": L1 with BPO1/BPO2 blob configs properly set (correct behavior)
//   - "missing":  L1 without BPO1/BPO2 blob configs (buggy behavior, falls back to Prague)
//
// It verifies that when BPO1/BPO2 blob configs are missing, the calculated blob base fee
// diverges dramatically from the correct value, demonstrating the fee explosion bug.
func Test_ProgramAction_BlobConfigMissingRegression(gt *testing.T) {
	runTest := func(gt *testing.T, testCfg *helpers.TestCfg[any]) {
		t := actionsHelpers.NewDefaultTesting(gt)

		// Create test environment with all L1 forks activated and CORRECT blob config
		env := helpers.NewL2ProofEnv(t, testCfg, helpers.NewTestParams(),
			helpers.NewBatcherCfg(
				func(c *actionsHelpers.BatcherCfg) {
					c.DataAvailabilityType = batcherFlags.BlobsType
				},
			),
			func(dp *genesis.DeployConfig) {
				dp.L1CancunTimeOffset = ptr(hexutil.Uint64(0))
				dp.L1PragueTimeOffset = ptr(hexutil.Uint64(12))
				dp.L1OsakaTimeOffset = ptr(hexutil.Uint64(24))
				dp.L1BPO1TimeOffset = ptr(hexutil.Uint64(36))
				dp.L1BPO2TimeOffset = ptr(hexutil.Uint64(48))
				// Complete blob schedule config (correct)
				dp.L1BlobScheduleConfig = &params.BlobScheduleConfig{
					Cancun: params.DefaultCancunBlobConfig,
					Prague: params.DefaultPragueBlobConfig,
					Osaka:  params.DefaultOsakaBlobConfig,
					BPO1:   params.DefaultBPO1BlobConfig,
					BPO2:   params.DefaultBPO2BlobConfig,
				}
				// Set high excess blob gas so blob fees are non-trivial
				dp.L1GenesisBlockExcessBlobGas = ptr(hexutil.Uint64(1e8))
			},
		)

		miner, sequencer := env.Miner, env.Sequencer

		// Bind to L1Block contract on L2
		l1BlockContract, err := legacybindings.NewL1Block(predeploys.L1BlockAddr, env.Engine.EthClient())
		require.NoError(t, err)

		// Advance L1 past BPO2 activation
		cfg := env.Sd.L1Cfg.Config
		require.NotNil(t, cfg.BPO2Time, "BPO2Time must be configured")
		h := miner.L1Chain().CurrentHeader()
		for h.Time < *cfg.BPO2Time {
			h = miner.ActEmptyBlock(t).Header()
		}

		// Verify BPO2 is active on L1
		l1Header := miner.L1Chain().CurrentHeader()
		require.True(t, cfg.IsBPO2(cfg.LondonBlock, l1Header.Time), "BPO2 should be active")

		// Calculate correct blob fee using the complete config
		correctBlobFee := eip4844.CalcBlobFee(cfg, l1Header)
		require.True(t, correctBlobFee.Cmp(big.NewInt(1)) > 0,
			"blob fee should be non-trivial for this test to be meaningful")

		// Now create a BUGGY config that simulates MainnetChainConfig missing BPO1/BPO2
		// This is the exact bug: BlobScheduleConfig only has Cancun+Prague, no BPO1/BPO2
		buggyConfig := *cfg // shallow copy
		buggyConfig.BlobScheduleConfig = &params.BlobScheduleConfig{
			Cancun: params.DefaultCancunBlobConfig,
			Prague: params.DefaultPragueBlobConfig,
			// BPO1 and BPO2 deliberately omitted — this is the bug
		}

		// Calculate blob fee with buggy config — it will fall back to Prague's UpdateFraction
		buggyBlobFee := eip4844.CalcBlobFee(&buggyConfig, l1Header)

		// The buggy fee should be dramatically higher than the correct fee.
		// Prague UpdateFraction = 5007716, BPO2 UpdateFraction = 11684671
		// With exponential pricing, this causes fees to explode.
		t.Logf("Correct blob fee (with BPO2 config):  %s", correctBlobFee.String())
		t.Logf("Buggy blob fee (missing BPO2 config):  %s", buggyBlobFee.String())

		// Calculate the ratio — it should be enormous (10,000x+)
		if correctBlobFee.Sign() > 0 {
			ratio := new(big.Int).Div(buggyBlobFee, correctBlobFee)
			t.Logf("Fee ratio (buggy/correct): %sx", ratio.String())
			// The fee with wrong UpdateFraction should be at least 100x higher
			// (in practice it's typically 10,000x+ depending on excess blob gas)
			require.True(t, ratio.Cmp(big.NewInt(100)) > 0,
				"missing BPO2 blob config causes fee to be %sx higher (expected >100x), proving the config bug causes massive fee inflation", ratio.String())
		}

		// Now verify the L2 side: advance L2 to sync with L1 and check blob fee on L2
		sequencer.ActL1HeadSignal(t)
		sequencer.ActBuildToL1HeadUnsafe(t)
		l2Block := sequencer.L2Unsafe()

		bbfL2, err := l1BlockContract.BlobBaseFee(&bind.CallOpts{
			BlockHash: l2Block.Hash,
		})
		require.NoError(t, err)

		// L2's blob base fee should match the CORRECT calculation (since env uses correct config)
		require.True(t, bbfL2.Cmp(correctBlobFee) == 0,
			"L2 blob base fee should match correct L1 calculation: got %s, want %s", bbfL2.String(), correctBlobFee.String())

		// And it should NOT match the buggy calculation
		require.True(t, bbfL2.Cmp(buggyBlobFee) != 0,
			"L2 blob base fee must NOT match buggy calculation")
	}

	matrix := helpers.NewMatrix[any]()
	matrix.AddDefaultTestCases(nil, helpers.MantleLatestForkOnly, runTest)
	matrix.Run(gt)
}

// Test_BlobScheduleConfig_MainnetCompleteness verifies that all chain configs with BPO fork times
// also have the corresponding BlobScheduleConfig entries. This is a pure unit test that catches
// the exact configuration omission that caused the fee explosion bug.
func Test_BlobScheduleConfig_MainnetCompleteness(gt *testing.T) {
	// These are the configs that matter — any config used as an L1 config
	configs := []struct {
		name string
		cfg  *params.ChainConfig
	}{
		{"Holesky", params.HoleskyChainConfig},
		{"Sepolia", params.SepoliaChainConfig},
		// MainnetChainConfig is the one that was missing BPO1/BPO2 — this test catches that
		{"Mainnet", params.MainnetChainConfig},
	}

	for _, tc := range configs {
		gt.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg
			bsc := cfg.BlobScheduleConfig

			// If the config has a fork time set, the corresponding blob config must exist
			if cfg.CancunTime != nil {
				require.NotNil(t, bsc, "%s: BlobScheduleConfig must not be nil when CancunTime is set", tc.name)
				require.NotNil(t, bsc.Cancun, "%s: Cancun blob config must be set when CancunTime is set", tc.name)
			}
			if cfg.PragueTime != nil {
				require.NotNil(t, bsc.Prague, "%s: Prague blob config must be set when PragueTime is set", tc.name)
			}
			if cfg.OsakaTime != nil {
				require.NotNil(t, bsc.Osaka, "%s: Osaka blob config must be set when OsakaTime is set", tc.name)
			}
			if cfg.BPO1Time != nil {
				require.NotNil(t, bsc, "%s: BlobScheduleConfig must not be nil when BPO1Time is set", tc.name)
				require.NotNil(t, bsc.BPO1,
					"%s: BPO1 blob config must be set when BPO1Time is set. "+
						"Missing this causes blob fee to use Prague's UpdateFraction (%d) instead of BPO1's (%d), "+
						"resulting in exponentially higher fees",
					tc.name, params.DefaultPragueBlobConfig.UpdateFraction, params.DefaultBPO1BlobConfig.UpdateFraction)
			}
			if cfg.BPO2Time != nil {
				require.NotNil(t, bsc.BPO2,
					"%s: BPO2 blob config must be set when BPO2Time is set. "+
						"Missing this causes blob fee to use Prague's UpdateFraction (%d) instead of BPO2's (%d), "+
						"resulting in exponentially higher fees",
					tc.name, params.DefaultPragueBlobConfig.UpdateFraction, params.DefaultBPO2BlobConfig.UpdateFraction)
			}
		})
	}
}
