package opnode

import (
	"github.com/ethereum/go-ethereum/core"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

func initMantleUpgradeConfig(rollupConfig *rollup.Config) {
	upgradeConfig := core.GetUpgradeConfigForMantle(rollupConfig.L2ChainID)
	if upgradeConfig == nil {
		rollupConfig.BaseFeeTime = nil
		return
	}
	rollupConfig.BaseFeeTime = upgradeConfig.BaseFeeTime
}
