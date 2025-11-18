package rollup

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum/go-ethereum/params"
)

// IsMantleBaseFee returns true if the MantleBaseFee hardfork is active at or past the given timestamp.
func (c *Config) IsMantleBaseFee(timestamp uint64) bool {
	return c.IsMantleForkActive(forks.MantleBaseFee, timestamp)
}

// IsMantleEverest returns true if the MantleEverest hardfork is active at or past the given timestamp.
func (c *Config) IsMantleEverest(timestamp uint64) bool {
	return c.IsMantleForkActive(forks.MantleEverest, timestamp)
}

// IsMantleEuboea returns true if the MantleEuboea hardfork is active at or past the given timestamp.
func (c *Config) IsMantleEuboea(timestamp uint64) bool {
	return c.IsMantleForkActive(forks.MantleEuboea, timestamp)
}

// IsMantleSkadi returns true if the MantleSkadi hardfork is active at or past the given timestamp.
func (c *Config) IsMantleSkadi(timestamp uint64) bool {
	return c.IsMantleForkActive(forks.MantleSkadi, timestamp)
}

// IsMantleLimb returns true if the MantleLimb hardfork is active at or past the given timestamp.
func (c *Config) IsMantleLimb(timestamp uint64) bool {
	return c.IsMantleForkActive(forks.MantleLimb, timestamp)
}

// IsMantleArsia returns true if the MantleArsia hardfork is active at or past the given timestamp.
func (c *Config) IsMantleArsia(timestamp uint64) bool {
	return c.IsMantleForkActive(forks.MantleArsia, timestamp)
}

// IsMantleBaseFeeActivationBlock returns whether the specified block is the first block subject to the
// MantleBaseFee upgrade.
func (c *Config) IsMantleBaseFeeActivationBlock(l2BlockTime uint64) bool {
	return c.IsMantleBaseFee(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsMantleBaseFee(l2BlockTime-c.BlockTime)
}

// IsMantleEverestActivationBlock returns whether the specified block is the first block subject to the
// MantleEverest upgrade.
func (c *Config) IsMantleEverestActivationBlock(l2BlockTime uint64) bool {
	return c.IsMantleEverest(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsMantleEverest(l2BlockTime-c.BlockTime)
}

// IsMantleEuboeaActivationBlock returns whether the specified block is the first block subject to the
// MantleEuboea upgrade.
func (c *Config) IsMantleEuboeaActivationBlock(l2BlockTime uint64) bool {
	return c.IsMantleEuboea(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsMantleEuboea(l2BlockTime-c.BlockTime)
}

// IsMantleSkadiActivationBlock returns whether the specified block is the first block subject to the
// MantleSkadi upgrade.
func (c *Config) IsMantleSkadiActivationBlock(l2BlockTime uint64) bool {
	return c.IsMantleSkadi(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsMantleSkadi(l2BlockTime-c.BlockTime)
}

// IsMantleLimbActivationBlock returns whether the specified block is the first block subject to the
// MantleLimb upgrade.
func (c *Config) IsMantleLimbActivationBlock(l2BlockTime uint64) bool {
	return c.IsMantleLimb(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsMantleLimb(l2BlockTime-c.BlockTime)
}

// IsMantleArsiaActivationBlock returns whether the specified block is the first block subject to the
// MantleArsia upgrade.
func (c *Config) IsMantleArsiaActivationBlock(l2BlockTime uint64) bool {
	return c.IsMantleArsia(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsMantleArsia(l2BlockTime-c.BlockTime)
}

func (c *Config) MantleActivationTime(fork MantleForkName) *uint64 {
	switch fork {
	case forks.MantleArsia:
		return c.MantleArsiaTime
	case forks.MantleLimb:
		return c.MantleLimbTime
	case forks.MantleSkadi:
		return c.MantleSkadiTime
	case forks.MantleEuboea:
		return c.MantleEuboeaTime
	case forks.MantleEverest:
		return c.MantleEverestTime
	case forks.MantleBaseFee:
		return c.MantleBaseFeeTime
	case forks.MantleNoSupport:
		return nil
	default:
		panic(fmt.Sprintf("unknown fork: %v", fork))
	}
}

func (c *Config) SetMantleActivationTime(fork MantleForkName, timestamp *uint64) {
	switch fork {
	case forks.MantleArsia:
		c.MantleArsiaTime = timestamp
	case forks.MantleLimb:
		c.MantleLimbTime = timestamp
	case forks.MantleSkadi:
		c.MantleSkadiTime = timestamp
	case forks.MantleEuboea:
		c.MantleEuboeaTime = timestamp
	case forks.MantleEverest:
		c.MantleEverestTime = timestamp
	case forks.MantleBaseFee:
		c.MantleBaseFeeTime = timestamp
	default:
		panic(fmt.Sprintf("unknown mantle fork: %v", fork))
	}
}

type MantleForkName = forks.MantleForkName

var scheduleableMantleForks = forks.MantleForksFrom(forks.MantleBaseFee)

func (c *Config) MantleActivateAtGenesis(hardfork MantleForkName) {
	c.MantleActivateAt(hardfork, 0)
}

func (c *Config) MantleActivateAt(fork MantleForkName, timestamp uint64) {
	if !forks.IsValidMantleFork(fork) {
		panic(fmt.Sprintf("invalid mantle fork: %s", fork))
	}
	ts := new(uint64)
	for i, f := range scheduleableMantleForks {
		if f == fork {
			c.SetMantleActivationTime(fork, &timestamp)
			ts = nil
		} else {
			c.SetMantleActivationTime(scheduleableMantleForks[i], ts)
		}
	}
}

func (c *Config) ApplyMantleOverrides() error {
	// Since we get the upgrade config from op-geth, configuration in rollup json will not be effective.
	// Which also means that we cannot customize the devnet upgrades through deploy config. The only way
	// to make it is to override the upgrade time in op-geth code.
	upgradeConfig := params.GetUpgradeConfigForMantle(c.L2ChainID)
	if upgradeConfig == nil {
		c.MantleBaseFeeTime = nil
	} else {
		c.MantleBaseFeeTime = upgradeConfig.BaseFeeTime
		c.MantleEverestTime = upgradeConfig.MantleEverestTime
		// No consensus&execution update for Euboea, just use the same as Everest
		c.MantleEuboeaTime = upgradeConfig.MantleEverestTime
		c.MantleSkadiTime = upgradeConfig.MantleSkadiTime
		c.MantleLimbTime = upgradeConfig.MantleLimbTime
		c.MantleArsiaTime = upgradeConfig.MantleArsiaTime

		// Map Optimism forks to Mantle forks
		c.CanyonTime = c.MantleArsiaTime
		c.DeltaTime = c.MantleArsiaTime
		c.EcotoneTime = c.MantleArsiaTime
		c.FjordTime = c.MantleArsiaTime
		c.GraniteTime = c.MantleArsiaTime
		c.HoloceneTime = c.MantleArsiaTime
		c.IsthmusTime = c.MantleArsiaTime
		c.JovianTime = c.MantleArsiaTime
	}

	if c.ChainOpConfig == nil {
		c.ChainOpConfig = &params.OptimismConfig{
			EIP1559Elasticity:  4,
			EIP1559Denominator: 50,
		}
	}
	// Mantle don't have a historical change of the denominator, so we use the same as the denominator
	c.ChainOpConfig.EIP1559DenominatorCanyon = &c.ChainOpConfig.EIP1559Denominator

	return c.CheckMantleForks()
}

func (cfg *Config) CheckMantleForks() error {
	if err := checkMantleFork(cfg.MantleBaseFeeTime, cfg.MantleEverestTime, forks.MantleBaseFee, forks.MantleEverest); err != nil {
		return err
	}
	if err := checkMantleFork(cfg.MantleEverestTime, cfg.MantleEuboeaTime, forks.MantleEverest, forks.MantleEuboea); err != nil {
		return err
	}
	if err := checkMantleFork(cfg.MantleEuboeaTime, cfg.MantleSkadiTime, forks.MantleEuboea, forks.MantleSkadi); err != nil {
		return err
	}
	if err := checkMantleFork(cfg.MantleSkadiTime, cfg.MantleLimbTime, forks.MantleSkadi, forks.MantleLimb); err != nil {
		return err
	}
	if err := checkMantleFork(cfg.MantleLimbTime, cfg.MantleArsiaTime, forks.MantleLimb, forks.MantleArsia); err != nil {
		return err
	}

	return nil
}

// checkFork checks that fork A is before or at the same time as fork B
func checkMantleFork(a, b *uint64, aName, bName MantleForkName) error {
	if a == nil && b == nil {
		return nil
	}
	if a == nil && b != nil {
		return fmt.Errorf("mantle fork %s set (to %d), but prior fork %s missing", bName, *b, aName)
	}
	if a != nil && b == nil {
		return nil
	}
	if *a > *b {
		return fmt.Errorf("mantle fork %s set to %d, but prior fork %s has higher offset %d", bName, *b, aName, *a)
	}
	return nil
}
