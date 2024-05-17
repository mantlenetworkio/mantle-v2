package eigenda

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

type DaUpgradeChainConfig struct {
	ChainID              *big.Int `json:"chainId"`                 // chainId identifies the current chain and is used for replay protection
	EigenDaUpgradeHeight *big.Int `json:"Eigen_da_upgrade_height"` // Upgrade Da from MantleDA to EigenDA
}

// OP Stack chain config
var (
	OptimismGoerliChainId = big.NewInt(420)
	// March 17, 2023 @ 7:00:00 pm UTC
	OptimismGoerliRegolithTime = uint64(1679079600)
	BaseGoerliChainId          = big.NewInt(84531)
	// April 27, 2023 @ 5:00:00 pm UTC
	BaseGoerliRegolithTime = uint64(1682614800)

	// Mantle chain_id
	MantleMainnetChainId = big.NewInt(5000)
	MantleSepoliaChainId = big.NewInt(5003)
	MantleLocalChainId   = big.NewInt(17)
)

var (
	MantleMainnetUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              MantleMainnetChainId,
		EigenDaUpgradeHeight: big.NewInt(0),
	}

	MantleSepoliaUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              MantleSepoliaChainId,
		EigenDaUpgradeHeight: big.NewInt(0),
	}
	MantleLocalUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              MantleLocalChainId,
		EigenDaUpgradeHeight: big.NewInt(0),
	}
	MantleDefaultUpgradeConfig = DaUpgradeChainConfig{
		EigenDaUpgradeHeight: big.NewInt(0),
	}
)

func GetDaUpgradeConfigForMantle(chainID *big.Int) *DaUpgradeChainConfig {
	if chainID == nil {
		return nil
	}
	switch chainID.Int64() {
	case params.MantleMainnetChainId.Int64():
		return &MantleMainnetUpgradeConfig
	case params.MantleSepoliaChainId.Int64():
		return &MantleSepoliaUpgradeConfig
	case params.MantleLocalChainId.Int64():
		return &MantleLocalUpgradeConfig
	default:
		return &MantleDefaultUpgradeConfig
	}
}
