package genesis

import (
	op_service "github.com/ethereum-optimism/optimism/op-service"
)

func (d *UpgradeScheduleDeployConfig) MantleBaseFeeTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleBaseFeeTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleBVMETHMintUpgradeTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleBVMETHMintUpgradeTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleMetaTxV2UpgradeTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleMetaTxV2UpgradeTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleMetaTxV3UpgradeTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleMetaTxV3UpgradeTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleProxyOwnerUpgradeTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleProxyOwnerUpgradeTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleEverestTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleEverestTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleEuboeaTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleEuboeaTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleSkadiTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleSkadiTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleLimbTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleLimbTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) MantleArsiaTime(genesisTime uint64) *uint64 {
	return offsetToUpgradeTime(d.L2GenesisMantleArsiaTimeOffset, genesisTime)
}

func (d *UpgradeScheduleDeployConfig) mantleForks() []Fork {
	return []Fork{
		{L2GenesisTimeOffset: d.L2GenesisMantleBaseFeeTimeOffset, Name: "mantle_base_fee"},
		{L2GenesisTimeOffset: d.L2GenesisMantleBVMETHMintUpgradeTimeOffset, Name: "mantle_bvm_eth_mint_upgrade"},
		{L2GenesisTimeOffset: d.L2GenesisMantleMetaTxV2UpgradeTimeOffset, Name: "mantle_meta_tx_v2_upgrade"},
		{L2GenesisTimeOffset: d.L2GenesisMantleMetaTxV3UpgradeTimeOffset, Name: "mantle_meta_tx_v3_upgrade"},
		{L2GenesisTimeOffset: d.L2GenesisMantleProxyOwnerUpgradeTimeOffset, Name: "mantle_proxy_owner_upgrade"},
		{L2GenesisTimeOffset: d.L2GenesisMantleEverestTimeOffset, Name: "mantle_everest"},
		{L2GenesisTimeOffset: d.L2GenesisMantleEuboeaTimeOffset, Name: "mantle_euboea"},
		{L2GenesisTimeOffset: d.L2GenesisMantleSkadiTimeOffset, Name: "mantle_skadi"},
		{L2GenesisTimeOffset: d.L2GenesisMantleLimbTimeOffset, Name: "mantle_limb"},
		{L2GenesisTimeOffset: d.L2GenesisMantleArsiaTimeOffset, Name: "mantle_arsia"},
	}
}

func (d *UpgradeScheduleDeployConfig) SolidityMantleForkNumber(genesisTime uint64) int64 {
	forks := d.mantleForks()
	for i := len(forks) - 1; i >= 4; i-- {
		if forkTime := offsetToUpgradeTime(forks[i].L2GenesisTimeOffset, genesisTime); forkTime != nil && *forkTime == 0 {
			// Subtract 4 since Solidity does not have the first five forks and have a "none" fork type
			return int64(i - 4)
		}
	}
	panic("should never reach here")
}

func DefaultMantleHardforkSchedule() *UpgradeScheduleDeployConfig {
	return &UpgradeScheduleDeployConfig{
		L2GenesisMantleBaseFeeTimeOffset:           op_service.U64UtilPtr(0),
		L2GenesisMantleBVMETHMintUpgradeTimeOffset: op_service.U64UtilPtr(0),
		L2GenesisMantleMetaTxV2UpgradeTimeOffset:   op_service.U64UtilPtr(0),
		L2GenesisMantleMetaTxV3UpgradeTimeOffset:   op_service.U64UtilPtr(0),
		L2GenesisMantleProxyOwnerUpgradeTimeOffset: op_service.U64UtilPtr(0),
		L2GenesisMantleEverestTimeOffset:           op_service.U64UtilPtr(0),
		L2GenesisMantleEuboeaTimeOffset:            op_service.U64UtilPtr(0),
		L2GenesisMantleSkadiTimeOffset:             op_service.U64UtilPtr(0),
		L2GenesisMantleLimbTimeOffset:              op_service.U64UtilPtr(0),
		L2GenesisMantleArsiaTimeOffset:             op_service.U64UtilPtr(0),
		L2GenesisRegolithTimeOffset:                op_service.U64UtilPtr(0),
	}
}
