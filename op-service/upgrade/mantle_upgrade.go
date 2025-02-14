package upgrade

import (
	"math/big"
)

type UpgradeChainConfig struct {
	ChainID              *big.Int `json:"chainId"`                 // chainId identifies the current chain and is used for replay protection
	EigenDaUpgradeHeight *big.Int `json:"Eigen_da_upgrade_height"` // Upgrade Da from MantleDA to EigenDA
}

// chain config
var (
	// Mantle chain_id
	MantleMainnetChainId    = big.NewInt(5000)
	MantleSepoliaChainId    = big.NewInt(5003)
	MantleSepoliaQAChainId  = big.NewInt(5003003)
	MantleHoleskyQA1ChainId = big.NewInt(5004001)
	MantleLocalChainId      = big.NewInt(17)
)

var (
	MantleMainnetUpgradeConfig = UpgradeChainConfig{
		ChainID:              MantleMainnetChainId,
		EigenDaUpgradeHeight: nil,
	}
	MantleSepoliaUpgradeConfig = UpgradeChainConfig{
		ChainID:              MantleSepoliaChainId,
		EigenDaUpgradeHeight: big.NewInt(9525400),
	}
	MantleSepoliaQAUpgradeConfig = UpgradeChainConfig{
		ChainID:              MantleSepoliaQAChainId,
		EigenDaUpgradeHeight: big.NewInt(3274000),
	}
	MantleHoleskyQA1UpgradeConfig = UpgradeChainConfig{
		ChainID:              MantleHoleskyQA1ChainId,
		EigenDaUpgradeHeight: big.NewInt(2374100),
	}
	MantleLocalUpgradeConfig = UpgradeChainConfig{
		ChainID:              MantleLocalChainId,
		EigenDaUpgradeHeight: big.NewInt(0),
	}
	MantleDefaultUpgradeConfig = UpgradeChainConfig{
		EigenDaUpgradeHeight: big.NewInt(0),
	}
)

func GetUpgradeConfigForMantle(chainID *big.Int) *UpgradeChainConfig {
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
	case MantleHoleskyQA1ChainId.Int64():
		return &MantleHoleskyQA1UpgradeConfig
	case MantleLocalChainId.Int64():
		return &MantleLocalUpgradeConfig
	default:
		return &MantleDefaultUpgradeConfig
	}
}

func (cfg *UpgradeChainConfig) IsEqualEigenDaUpgradeBlock(l2Block *big.Int) bool {
	if cfg != nil && cfg.EigenDaUpgradeHeight != nil && cfg.EigenDaUpgradeHeight.Cmp(l2Block) == 0 {
		return true
	}
	return false
}

func (cfg *UpgradeChainConfig) IsUseEigenDa(l2Block *big.Int) bool {
	if cfg != nil && cfg.EigenDaUpgradeHeight != nil && cfg.EigenDaUpgradeHeight.Cmp(l2Block) <= 0 {
		return true
	}
	return false
}
