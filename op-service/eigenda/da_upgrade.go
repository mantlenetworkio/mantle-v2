package eigenda

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

type DaUpgradeChainConfig struct {
	ChainID              *big.Int `json:"chainId"`                 // chainId identifies the current chain and is used for replay protection
	EigenDaUpgradeHeight *big.Int `json:"Eigen_da_upgrade_height"` // Upgrade Da from MantleDA to EigenDA
}

var (
	MantleMainnetUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              params.MantleMainnetChainId,
		EigenDaUpgradeHeight: big.NewInt(0),
	}

	MantleSepoliaUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              params.MantleSepoliaChainId,
		EigenDaUpgradeHeight: big.NewInt(0),
	}
	MantleLocalUpgradeConfig = DaUpgradeChainConfig{
		ChainID:              params.MantleLocalChainId,
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
