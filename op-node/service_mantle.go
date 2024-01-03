package opnode

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

func initMantleUpgradeConfig(rollupConfig *rollup.Config) {
	switch rollupConfig.L2ChainID {
	case params.MantleSepoliaChainId:
		rollupConfig.BaseFeeTime = core.MantleSepoliaUpgradeConfig.BaseFeeTime
	default:
		rollupConfig.BaseFeeTime = nil
	}
}
