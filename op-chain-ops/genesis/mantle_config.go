package genesis

import (
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-chain-ops/foundry"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	op_service "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
func BuildMantleGenesis(config *DeployConfig, dump *foundry.ForgeAllocs, l1StartBlock *eth.BlockRef) (*core.Genesis, error) {
	genesis, err := BuildL2Genesis(config, dump, l1StartBlock)
	if err != nil {
		return nil, err
	}

	fillInMantleForksIntoGenesis(config, genesis, l1StartBlock.Time)

	// align the Ethereum forks with the Mantle forks
	alignEthWithMantle(genesis)

	if genesis.Config.IsMantleSkadi(genesis.Timestamp) {
		genesis.BlobGasUsed = u64ptr(0)
		genesis.ExcessBlobGas = u64ptr(0)

		genesis.Alloc[params.HistoryStorageAddress] = types.Account{Nonce: 1, Code: params.HistoryStorageCode, Balance: common.Big0}
	}

	if genesis.Config.IsHolocene(genesis.Timestamp) && genesis.Config.Optimism != nil {
		denom := uint64(genesis.Config.Optimism.EIP1559Denominator)
		elasticity := uint64(genesis.Config.Optimism.EIP1559Elasticity)
		genesis.ExtraData = eip1559.EncodeHoloceneExtraData(denom, elasticity)
	}
	if genesis.Config.IsMinBaseFee(genesis.Timestamp) && genesis.Config.Optimism != nil {
		denom := uint64(genesis.Config.Optimism.EIP1559Denominator)
		elasticity := uint64(genesis.Config.Optimism.EIP1559Elasticity)
		genesis.ExtraData = eip1559.EncodeMinBaseFeeExtraData(denom, elasticity, 0)
	}

	if config.GasPriceOracleTokenRatio != 0 {
		tokenRatioSlot := common.BigToHash(big.NewInt(0))
		gpoAccount := genesis.Alloc[predeploys.GasPriceOracleAddr]
		if gpoAccount.Storage == nil {
			gpoAccount.Storage = map[common.Hash]common.Hash{}
		}
		gpoAccount.Storage[tokenRatioSlot] = common.BigToHash(new(big.Int).SetUint64(config.GasPriceOracleTokenRatio))
		genesis.Alloc[predeploys.GasPriceOracleAddr] = gpoAccount
	}

	// ExtraData is already set by NewL2Genesis based on config values
	// Do not override it here with hardcoded MinBaseFeeExtraData

	return genesis, nil
}

// Avoid modifying BuildL2Genesis function directly
func fillInMantleForksIntoGenesis(config *DeployConfig, genesis *core.Genesis, l1StartBlockTimestamp uint64) {
	chainConfig := genesis.Config
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

// alignEthWithMantle aligns the Ethereum forks with the Mantle forks.
// Optimism forks are aligned as well to avoid potential errors.
func alignEthWithMantle(genesis *core.Genesis) {
	chainConfig := genesis.Config

	// Skadi
	chainConfig.ShanghaiTime = chainConfig.MantleSkadiTime
	chainConfig.CancunTime = chainConfig.MantleSkadiTime
	chainConfig.PragueTime = chainConfig.MantleSkadiTime

	// Limb
	chainConfig.OsakaTime = chainConfig.MantleLimbTime

	// Arsia
	chainConfig.CanyonTime = chainConfig.MantleArsiaTime
	chainConfig.EcotoneTime = chainConfig.MantleArsiaTime
	chainConfig.FjordTime = chainConfig.MantleArsiaTime
	chainConfig.GraniteTime = chainConfig.MantleArsiaTime
	chainConfig.HoloceneTime = chainConfig.MantleArsiaTime
	chainConfig.IsthmusTime = chainConfig.MantleArsiaTime
	chainConfig.JovianTime = chainConfig.MantleArsiaTime
}

/////////////////////////////////////////////////////////////
// rollup config
/////////////////////////////////////////////////////////////

// MantleRollupConfig converts a DeployConfig to a rollup.Config. If Ecotone is active at genesis, the
// Overhead value is considered a noop.
func (d *DeployConfig) MantleRollupConfig(l1StartBlock *eth.BlockRef, l2GenesisBlockHash common.Hash, l2GenesisBlockNumber uint64) (*rollup.Config, error) {
	rollupConfig, err := d.RollupConfig(l1StartBlock, l2GenesisBlockHash, l2GenesisBlockNumber)
	if err != nil {
		return nil, err
	}

	fillInMantleForksIntoRollupConfig(d, rollupConfig, l1StartBlock.Time)

	if err := rollupConfig.AlignOpWithMantle(); err != nil {
		return nil, err
	}

	if d.L2GenesisMantleArsiaTimeOffset == nil {
		rollupConfig.Genesis.SystemConfig.MarshalPreHolocene = true
	}

	return rollupConfig, nil
}

// Avoid modifying RollupConfig function directly
func fillInMantleForksIntoRollupConfig(config *DeployConfig, rollupConfig *rollup.Config, l1StartTime uint64) {
	rollupConfig.MantleBaseFeeTime = config.MantleBaseFeeTime(l1StartTime)
	rollupConfig.MantleEverestTime = config.MantleEverestTime(l1StartTime)
	rollupConfig.MantleEuboeaTime = config.MantleEuboeaTime(l1StartTime)
	rollupConfig.MantleSkadiTime = config.MantleSkadiTime(l1StartTime)
	rollupConfig.MantleLimbTime = config.MantleLimbTime(l1StartTime)
	rollupConfig.MantleArsiaTime = config.MantleArsiaTime(l1StartTime)
}

/////////////////////////////////////////////////////////////
// Mantle fork activation helpers
/////////////////////////////////////////////////////////////

// ActivateMantleForkAtOffset activates the given Mantle fork at the given offset.
// This method follows the same pattern as ActivateForkAtOffset:
// - Activates all previous forks (dependencies) at genesis (offset 0)
// - Activates the target fork at the specified offset
// - Deactivates all later forks (sets them to nil)
func (d *UpgradeScheduleDeployConfig) ActivateMantleForkAtOffset(fork rollup.MantleForkName, offset uint64) {
	if !forks.IsValidMantleFork(fork) {
		panic(fmt.Sprintf("invalid mantle fork: %s", fork))
	}

	ts := new(uint64) // 0 for previous forks
	for i, f := range forks.AllMantleForks {
		if f == fork {
			d.SetMantleForkTimeOffset(fork, &offset) // Set target fork
			ts = nil                                 // Later forks will be set to nil
		} else {
			d.SetMantleForkTimeOffset(forks.AllMantleForks[i], ts)
		}
	}

	// Special handling for Arsia: activate OP Stack forks
	if fork == forks.MantleArsia {
		d.L2GenesisCanyonTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisDeltaTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisEcotoneTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisFjordTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisGraniteTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisHoloceneTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisIsthmusTimeOffset = (*hexutil.Uint64)(&offset)
		d.L2GenesisJovianTimeOffset = (*hexutil.Uint64)(&offset)
	} else {
		// Non-Arsia forks: clear all OP Stack forks
		d.L2GenesisCanyonTimeOffset = nil
		d.L2GenesisDeltaTimeOffset = nil
		d.L2GenesisEcotoneTimeOffset = nil
		d.L2GenesisFjordTimeOffset = nil
		d.L2GenesisGraniteTimeOffset = nil
		d.L2GenesisHoloceneTimeOffset = nil
		d.L2GenesisIsthmusTimeOffset = nil
		d.L2GenesisJovianTimeOffset = nil
	}
}

// SetMantleForkTimeOffset sets the time offset for a Mantle fork
func (d *UpgradeScheduleDeployConfig) SetMantleForkTimeOffset(fork rollup.MantleForkName, offset *uint64) {
	switch fork {
	case forks.MantleBaseFee:
		d.L2GenesisMantleBaseFeeTimeOffset = (*hexutil.Uint64)(offset)
	case forks.MantleEverest:
		d.L2GenesisMantleEverestTimeOffset = (*hexutil.Uint64)(offset)
	case forks.MantleEuboea:
		d.L2GenesisMantleEuboeaTimeOffset = (*hexutil.Uint64)(offset)
	case forks.MantleSkadi:
		d.L2GenesisMantleSkadiTimeOffset = (*hexutil.Uint64)(offset)
	case forks.MantleLimb:
		d.L2GenesisMantleLimbTimeOffset = (*hexutil.Uint64)(offset)
	case forks.MantleArsia:
		d.L2GenesisMantleArsiaTimeOffset = (*hexutil.Uint64)(offset)
	default:
		panic(fmt.Sprintf("unsupported mantle fork: %s", fork))
	}
}

// ActivateMantleForkAtGenesis activates the given Mantle fork at genesis.
func (d *UpgradeScheduleDeployConfig) ActivateMantleForkAtGenesis(fork rollup.MantleForkName) {
	d.ActivateMantleForkAtOffset(fork, 0)
}
