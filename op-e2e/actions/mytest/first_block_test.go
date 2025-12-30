package mytest

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// TestLimbChainProduceBlock 验证 LIMB 版本链出块
func TestLimbChainProduceBlock(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	log := testlog.Logger(t, log.LevelInfo)

	// 创建 LIMB 版本部署参数
	dp := e2eutils.MakeDeployParamsLimb(t, &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	})

	// 创建 LIMB 版本测试环境
	sd := e2eutils.SetupMantleLimb(t, dp, helpers.DefaultAlloc)

	// 验证 LIMB 已激活，Arsia 未激活
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.True(t, sd.RollupCfg.IsMantleLimb(genesisTime), "LIMB fork 应该已激活")
	require.False(t, sd.RollupCfg.IsMantleArsia(genesisTime), "Arsia fork 不应该激活")

	// 创建 Actors
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)
	miner.ActL1SetFeeRecipient(common.Address{'A'})
	sequencer.ActL2PipelineFull(t)

	// 生成 L1 区块
	miner.ActEmptyBlock(t)

	// 生成 L2 区块
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2StartBlock(t)
	sequencer.ActL2EndBlock(t)

	// 验证区块已生成
	status := sequencer.SyncStatus()
	require.Equal(t, uint64(1), status.UnsafeL2.Number, "应该生成第 1 个 L2 区块")

	// 验证 ExtraData 为空 (LIMB 版本无 Holocene)
	l2Block := seqEngine.L2Chain().GetBlockByNumber(status.UnsafeL2.Number)
	require.NotNil(t, l2Block, "L2 区块不应为空")
	extraData := l2Block.Extra()
	require.Equal(t, 0, len(extraData), "LIMB 版本 ExtraData 应该为 0 字节")

	t.Log("✅ LIMB 链成功出块")
}

// TestArsiaChainProduceBlock 验证 ARSIA 版本链出块
func TestArsiaChainProduceBlock(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	log := testlog.Logger(t, log.LevelInfo)

	// 创建 ARSIA 版本部署参数
	dp := e2eutils.MakeDeployParams(t, &e2eutils.TestParams{
		MaxSequencerDrift:   40,
		SequencerWindowSize: 120,
		ChannelTimeout:      120,
		L1BlockTime:         12,
		AllocType:           config.DefaultAllocType,
	})

	// 创建 ARSIA 版本测试环境
	sd := e2eutils.SetupMantle(t, dp, helpers.DefaultAlloc)

	// 验证 Arsia 已激活
	genesisTime := sd.RollupCfg.Genesis.L2Time
	require.True(t, sd.RollupCfg.IsMantleArsia(genesisTime), "Arsia fork 应该已激活")

	// 创建 Actors
	miner, seqEngine, sequencer := helpers.SetupSequencerTest(t, sd, log)
	miner.ActL1SetFeeRecipient(common.Address{'A'})
	sequencer.ActL2PipelineFull(t)

	// 生成 L1 区块
	miner.ActEmptyBlock(t)

	// 生成 L2 区块
	sequencer.ActL1HeadSignal(t)
	sequencer.ActL2StartBlock(t)
	sequencer.ActL2EndBlock(t)

	// 验证区块已生成
	status := sequencer.SyncStatus()
	require.Equal(t, uint64(1), status.UnsafeL2.Number, "应该生成第 1 个 L2 区块")

	// 验证 ExtraData 非空 (ARSIA 版本有 Holocene)
	currentTime := sequencer.L2Unsafe().Time
	if sd.RollupCfg.IsHolocene(currentTime) {
		l2Block := seqEngine.L2Chain().GetBlockByNumber(status.UnsafeL2.Number)
		require.NotNil(t, l2Block, "L2 区块不应为空")
		extraData := l2Block.Extra()
		require.GreaterOrEqual(t, len(extraData), 9, "ARSIA 版本 ExtraData 至少 9 字节")
	}

	t.Log("✅ ARSIA 链成功出块")
}
