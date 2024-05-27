package eigenda

import (
	"math/big"
)

type DaUpgradeChainConfig struct {
	ChainID              *big.Int `json:"chainId"`                 // chainId identifies the current chain and is used for replay protection
	EigenDaUpgradeHeight *big.Int `json:"Eigen_da_upgrade_height"` // Upgrade Da from MantleDA to EigenDA
}

// chain config
var (
	// Mantle chain_id
	MantleMainnetChainId   = big.NewInt(5000)
	MantleSepoliaChainId   = big.NewInt(5003)
	MantleSepoliaQAChainId = big.NewInt(5003003)
	MantleLocalChainId     = big.NewInt(17)
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
	MantleSepoliaQAUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              MantleSepoliaQAChainId,
		EigenDaUpgradeHeight: big.NewInt(3300000),
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
	case MantleMainnetChainId.Int64():
		return &MantleMainnetUpgradeConfig
	case MantleSepoliaChainId.Int64():
		return &MantleSepoliaUpgradeConfig
	case MantleSepoliaQAChainId.Int64():
		return &MantleSepoliaQAUpgradeConfig
	case MantleLocalChainId.Int64():
		return &MantleLocalUpgradeConfig
	default:
		return &MantleDefaultUpgradeConfig
	}
}
