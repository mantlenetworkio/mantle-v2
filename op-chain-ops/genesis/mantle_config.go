package genesis

import (
	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	op_service "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
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
		L2GenesisRegolithTimeOffset:                op_service.U64UtilPtr(0),
		L2GenesisMantleBaseFeeTimeOffset:           op_service.U64UtilPtr(0),
		L2GenesisMantleBVMETHMintUpgradeTimeOffset: op_service.U64UtilPtr(0),
		L2GenesisMantleMetaTxV2UpgradeTimeOffset:   op_service.U64UtilPtr(0),
		L2GenesisMantleMetaTxV3UpgradeTimeOffset:   op_service.U64UtilPtr(0),
		L2GenesisMantleProxyOwnerUpgradeTimeOffset: op_service.U64UtilPtr(0),
		L2GenesisMantleEverestTimeOffset:           op_service.U64UtilPtr(0),
		L2GenesisMantleEuboeaTimeOffset:            op_service.U64UtilPtr(0),
		L2GenesisMantleSkadiTimeOffset:             op_service.U64UtilPtr(0),
		L2GenesisMantleLimbTimeOffset:              op_service.U64UtilPtr(0),
	}
}

/////////////////////////////////////////////////////////////
// genesis
/////////////////////////////////////////////////////////////

// BuildMantleGenesis will build the mantle genesis block.
func BuildMantleGenesis(config *DeployConfig, dump *foundry.ForgeAllocs, l1StartBlock *eth.BlockRef, overrides *params.MantleUpgradeChainConfig) (*core.Genesis, error) {
	genesis, err := BuildL2Genesis(config, dump, l1StartBlock)
	if err != nil {
		return nil, err
	}

	// Apply mantle geth overrides
	applyMantleGethOverrides(config, genesis, l1StartBlock.Time, overrides)

	if genesis.Config.IsMantleSkadi(genesis.Timestamp) {
		genesis.BlobGasUsed = u64ptr(0)
		genesis.ExcessBlobGas = u64ptr(0)

		genesis.Alloc[params.HistoryStorageAddress] = types.Account{Nonce: 1, Code: params.HistoryStorageCode, Balance: common.Big0}
	}

	if genesis.Config.IsMantleArsia(genesis.Timestamp) {
		genesis.ExtraData = MinBaseFeeExtraData
	}

	return genesis, nil
}

// applyMantleGethOverrides applies the mantle geth overrides to the genesis config.
// Ref: https://github.com/mantlenetworkio/op-geth/blob/13f718f59d4d523ea4edf4c5a0174423946e97db/core/genesis.go#L329-L352
// Key differences:
// - allow a non hard coded mantle upgradeconfig, to support custom mantle upgrade schedules
func applyMantleGethOverrides(config *DeployConfig, genesis *core.Genesis, l1StartBlockTimestamp uint64, overrides *params.MantleUpgradeChainConfig) {
	chainConfig := genesis.Config

	if overrides != nil {
		chainConfig.BaseFeeTime = overrides.BaseFeeTime
		chainConfig.BVMETHMintUpgradeTime = overrides.BVMETHMintUpgradeTime
		chainConfig.MetaTxV2UpgradeTime = overrides.MetaTxV2UpgradeTime
		chainConfig.MetaTxV3UpgradeTime = overrides.MetaTxV3UpgradeTime
		chainConfig.ProxyOwnerUpgradeTime = overrides.ProxyOwnerUpgradeTime
		chainConfig.MantleEverestTime = overrides.MantleEverestTime
		chainConfig.MantleSkadiTime = overrides.MantleSkadiTime
		chainConfig.MantleLimbTime = overrides.MantleLimbTime
		chainConfig.MantleArsiaTime = overrides.MantleArsiaTime
	} else {
		chainConfig.BaseFeeTime = config.MantleBaseFeeTime(l1StartBlockTimestamp)
		chainConfig.BVMETHMintUpgradeTime = config.MantleBVMETHMintUpgradeTime(l1StartBlockTimestamp)
		chainConfig.MetaTxV2UpgradeTime = config.MantleMetaTxV2UpgradeTime(l1StartBlockTimestamp)
		chainConfig.MetaTxV3UpgradeTime = config.MantleMetaTxV3UpgradeTime(l1StartBlockTimestamp)
		chainConfig.ProxyOwnerUpgradeTime = config.MantleProxyOwnerUpgradeTime(l1StartBlockTimestamp)
		chainConfig.MantleEverestTime = config.MantleEverestTime(l1StartBlockTimestamp)
		chainConfig.MantleSkadiTime = config.MantleSkadiTime(l1StartBlockTimestamp)
		chainConfig.MantleLimbTime = config.MantleLimbTime(l1StartBlockTimestamp)
		chainConfig.MantleArsiaTime = config.MantleArsiaTime(l1StartBlockTimestamp)
	}

	chainConfig.ShanghaiTime = chainConfig.MantleSkadiTime
	chainConfig.CancunTime = chainConfig.MantleSkadiTime
	chainConfig.PragueTime = chainConfig.MantleSkadiTime

	chainConfig.OsakaTime = chainConfig.MantleLimbTime

	chainConfig.CanyonTime = chainConfig.MantleArsiaTime
	chainConfig.EcotoneTime = chainConfig.MantleArsiaTime
	chainConfig.FjordTime = chainConfig.MantleArsiaTime
	chainConfig.GraniteTime = chainConfig.MantleArsiaTime
	chainConfig.HoloceneTime = chainConfig.MantleArsiaTime
	chainConfig.IsthmusTime = chainConfig.MantleArsiaTime
	chainConfig.JovianTime = chainConfig.MantleArsiaTime

	if chainConfig.MantleArsiaTime != nil {
		chainConfig.Optimism = &params.OptimismConfig{
			EIP1559Elasticity:  4,
			EIP1559Denominator: 50,
		}
	}
}

/////////////////////////////////////////////////////////////
// rollup config
/////////////////////////////////////////////////////////////

// MantleRollupConfig converts a DeployConfig to a rollup.Config. If Ecotone is active at genesis, the
// Overhead value is considered a noop.
func (d *DeployConfig) MantleRollupConfig(l1StartBlock *eth.BlockRef, l2GenesisBlockHash common.Hash, l2GenesisBlockNumber uint64, overrides *params.MantleUpgradeChainConfig) (*rollup.Config, error) {
	rollupConfig, err := d.RollupConfig(l1StartBlock, l2GenesisBlockHash, l2GenesisBlockNumber)
	if err != nil {
		return nil, err
	}

	if err := rollupConfig.ApplyMantleOverrides(overrides); err != nil {
		return nil, err
	}

	// setup initial 1559 params in rollup system config
	if d.L2GenesisMantleArsiaTimeOffset == nil {
		rollupConfig.Genesis.SystemConfig.MarshalPreHolocene = true
	}
	if d.L2GenesisMantleArsiaTimeOffset != nil && *d.L2GenesisMantleArsiaTimeOffset == 0 {
		rollupConfig.Genesis.SystemConfig.EIP1559Params = eth.Bytes8(eip1559.EncodeHolocene1559Params(d.EIP1559Denominator, d.EIP1559Elasticity))
	}

	return rollupConfig, nil
}
