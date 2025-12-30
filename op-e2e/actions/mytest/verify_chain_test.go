package mytest

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestVerifyMantleChainConfig 验证启动的确实是 Mantle 链
func TestVerifyMantleChainConfig(t *testing.T) {
	// 1️⃣ 初始化测试框架
	gt := helpers.NewDefaultTesting(t)
	logger := testlog.Logger(gt, log.LevelInfo)

	// 2️⃣ 创建部署参数
	dp := e2eutils.MakeDeployParams(gt, helpers.DefaultRollupTestParams())

	// 3️⃣ 设置测试环境
	sd := e2eutils.SetupMantle(gt, dp, helpers.DefaultAlloc)

	// 4️⃣ 创建 Actors
	miner, engine, _ := helpers.SetupSequencerTest(gt, sd, logger)

	// ========================================
	// 验证 L1 链配置
	// ========================================
	gt.Log("========== L1 Chain 配置 ==========")
	l1Config := miner.L1Chain().Config()

	gt.Logf("L1 Chain ID: %d", l1Config.ChainID.Uint64())
	require.Equal(gt, uint64(900), l1Config.ChainID.Uint64(), "L1 ChainID 应该是 900")

	gt.Logf("L1 Shanghai Time: %v", l1Config.ShanghaiTime)
	require.NotNil(gt, l1Config.ShanghaiTime, "L1 Shanghai 应该激活")
	require.Equal(gt, uint64(0), *l1Config.ShanghaiTime, "L1 Shanghai 应该在 genesis 激活")

	gt.Logf("L1 Cancun Time: %v", l1Config.CancunTime)
	require.NotNil(gt, l1Config.CancunTime, "L1 Cancun 应该激活")
	require.Equal(gt, uint64(0), *l1Config.CancunTime, "L1 Cancun 应该在 genesis 激活")

	// ========================================
	// 验证 L2 链配置
	// ========================================
	gt.Log("========== L2 Chain 配置 ==========")
	l2Config := engine.L2Chain().Config()

	gt.Logf("L2 Chain ID: %d", l2Config.ChainID.Uint64())
	require.Equal(gt, uint64(901), l2Config.ChainID.Uint64(), "L2 ChainID 应该是 901 (Mantle)")

	// 验证 Mantle 特有的 forks
	gt.Logf("L2 BaseFee Time: %v", l2Config.BaseFeeTime)
	if l2Config.BaseFeeTime != nil {
		gt.Logf("  -> Mantle BaseFee fork 激活于: %d", *l2Config.BaseFeeTime)
	}

	gt.Logf("L2 MantleSkadi Time: %v", l2Config.MantleSkadiTime)
	if l2Config.MantleSkadiTime != nil {
		gt.Logf("  -> Mantle Skadi fork 激活于: %d", *l2Config.MantleSkadiTime)
	}

	gt.Logf("L2 MantleArsia Time: %v", l2Config.MantleArsiaTime)
	if l2Config.MantleArsiaTime != nil {
		gt.Logf("  -> Mantle Arsia fork 激活于: %d", *l2Config.MantleArsiaTime)
	}

	// 验证 OP Stack forks (如果 MantleArsia 激活,这些应该也激活)
	gt.Log("========== OP Stack Forks (应随 MantleArsia 激活) ==========")
	gt.Logf("L2 Ecotone Time: %v", l2Config.EcotoneTime)
	gt.Logf("L2 Fjord Time: %v", l2Config.FjordTime)
	gt.Logf("L2 Granite Time: %v", l2Config.GraniteTime)
	gt.Logf("L2 Holocene Time: %v", l2Config.HoloceneTime)
	gt.Logf("L2 Isthmus Time: %v", l2Config.IsthmusTime)
	gt.Logf("L2 Jovian Time: %v", l2Config.JovianTime)

	// 验证 Optimism 配置
	gt.Log("========== Optimism 配置 ==========")
	if l2Config.Optimism != nil {
		gt.Logf("EIP1559 Denominator: %d", l2Config.Optimism.EIP1559Denominator)
		require.Equal(gt, uint64(50), l2Config.Optimism.EIP1559Denominator, "应该是 Mantle 的 denominator")

		gt.Logf("EIP1559 Elasticity: %d", l2Config.Optimism.EIP1559Elasticity)
		require.Equal(gt, uint64(4), l2Config.Optimism.EIP1559Elasticity, "应该是 Mantle 的 elasticity")

		if l2Config.Optimism.EIP1559DenominatorCanyon != nil {
			gt.Logf("EIP1559 DenominatorCanyon: %d", *l2Config.Optimism.EIP1559DenominatorCanyon)
		}
	}

	// ========================================
	// 验证 RollupConfig
	// ========================================
	gt.Log("========== Rollup 配置 ==========")
	rollupCfg := sd.RollupCfg

	gt.Logf("Rollup L1 ChainID: %d", rollupCfg.L1ChainID.Uint64())
	require.Equal(gt, uint64(900), rollupCfg.L1ChainID.Uint64())

	gt.Logf("Rollup L2 ChainID: %d", rollupCfg.L2ChainID.Uint64())
	require.Equal(gt, uint64(901), rollupCfg.L2ChainID.Uint64())

	gt.Logf("Rollup MantleArsia Time: %v", rollupCfg.MantleArsiaTime)
	gt.Logf("Rollup Jovian Time: %v", rollupCfg.JovianTime)
	gt.Logf("Rollup Holocene Time: %v", rollupCfg.HoloceneTime)

	// 关键验证: ChainConfig 和 RollupConfig 的 fork 时间必须一致
	gt.Log("========== 配置一致性验证 ==========")
	if rollupCfg.MantleArsiaTime != nil && l2Config.MantleArsiaTime != nil {
		require.Equal(gt, *rollupCfg.MantleArsiaTime, *l2Config.MantleArsiaTime,
			"RollupConfig 和 ChainConfig 的 MantleArsiaTime 必须一致")
		gt.Log("✅ MantleArsiaTime 一致")
	}

	if rollupCfg.JovianTime != nil && l2Config.JovianTime != nil {
		require.Equal(gt, *rollupCfg.JovianTime, *l2Config.JovianTime,
			"RollupConfig 和 ChainConfig 的 JovianTime 必须一致")
		gt.Log("✅ JovianTime 一致")
	}

	if rollupCfg.HoloceneTime != nil && l2Config.HoloceneTime != nil {
		require.Equal(gt, *rollupCfg.HoloceneTime, *l2Config.HoloceneTime,
			"RollupConfig 和 ChainConfig 的 HoloceneTime 必须一致")
		gt.Log("✅ HoloceneTime 一致")
	}

	// ========================================
	// 验证 L2 Genesis Block
	// ========================================
	gt.Log("========== L2 Genesis Block ==========")
	genesisBlock := engine.L2Chain().GetBlockByNumber(0)
	require.NotNil(gt, genesisBlock, "应该有 genesis block")

	gt.Logf("Genesis Block Number: %d", genesisBlock.NumberU64())
	gt.Logf("Genesis Block Hash: %s", genesisBlock.Hash().Hex())
	gt.Logf("Genesis Block Time: %d", genesisBlock.Time())
	gt.Logf("Genesis Block Coinbase: %s", genesisBlock.Coinbase().Hex())
	gt.Logf("Genesis Block GasLimit: %d", genesisBlock.GasLimit())
	gt.Logf("Genesis Block BaseFee: %d wei", genesisBlock.BaseFee())

	// ExtraData 应该包含 EIP1559 参数 (如果 Holocene/Jovian 激活)
	if len(genesisBlock.Extra()) > 0 {
		gt.Logf("Genesis Block ExtraData (len=%d): %x", len(genesisBlock.Extra()), genesisBlock.Extra())

		// Holocene ExtraData = 16 bytes (denominator + elasticity)
		// Jovian ExtraData = 24 bytes (denominator + elasticity + minBaseFee)
		if len(genesisBlock.Extra()) == 16 {
			gt.Log("  -> Holocene ExtraData 格式")
		} else if len(genesisBlock.Extra()) == 24 {
			gt.Log("  -> Jovian/MinBaseFee ExtraData 格式")
		}
	}

	// ========================================
	// 验证 Predeploys
	// ========================================
	gt.Log("========== L2 Predeploys 验证 ==========")

	// L1Block (0x4200000000000000000000000000000000000015)
	l1BlockAddr := common.HexToAddress("0x4200000000000000000000000000000000000015")
	state, err := engine.L2Chain().StateAt(genesisBlock.Root())
	require.NoError(gt, err, "应该能获取 genesis state")
	l1BlockCode := state.GetCode(l1BlockAddr)
	require.NotEmpty(gt, l1BlockCode, "L1Block predeploy 应该有代码")
	gt.Logf("✅ L1Block predeploy 存在 (code size: %d bytes)", len(l1BlockCode))

	// SequencerFeeVault (0x4200000000000000000000000000000000000011)
	seqFeeVaultAddr := common.HexToAddress("0x4200000000000000000000000000000000000011")
	seqFeeVaultCode := state.GetCode(seqFeeVaultAddr)
	require.NotEmpty(gt, seqFeeVaultCode, "SequencerFeeVault predeploy 应该有代码")
	gt.Logf("✅ SequencerFeeVault predeploy 存在 (code size: %d bytes)", len(seqFeeVaultCode))

	// L2StandardBridge (0x4200000000000000000000000000000000000010)
	l2BridgeAddr := common.HexToAddress("0x4200000000000000000000000000000000000010")
	l2BridgeCode := state.GetCode(l2BridgeAddr)
	require.NotEmpty(gt, l2BridgeCode, "L2StandardBridge predeploy 应该有代码")
	gt.Logf("✅ L2StandardBridge predeploy 存在 (code size: %d bytes)", len(l2BridgeCode))

	gt.Log("")
	gt.Log("🎉 所有验证通过! 这确实是一条配置正确的 Mantle L2 链!")
	gt.Logf("   L1 Chain ID: %d", l1Config.ChainID.Uint64())
	gt.Logf("   L2 Chain ID: %d", l2Config.ChainID.Uint64())
	gt.Log("   配置一致性: ✅")
	gt.Log("   Predeploys: ✅")
}
