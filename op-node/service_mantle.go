package opnode

import (
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

func initMantleUpgradeConfig(rollupConfig *rollup.Config) {
	upgradeConfig := params.GetUpgradeConfigForMantle(rollupConfig.L2ChainID)
	if upgradeConfig == nil {
		rollupConfig.BaseFeeTime = nil
		return
	}
	rollupConfig.BaseFeeTime = upgradeConfig.BaseFeeTime
	rollupConfig.MantleSkadiTime = upgradeConfig.MantleSkadiTime
}
