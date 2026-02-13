package helpers

import (
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// ApplyLimbTimeOffset adjusts fork configuration to activate Limb fork at the specified time.
// It ensures that Arsia (which comes after Limb) won't accidentally activate at the same time or before Limb.
// When activating only Limb, Arsia will be delayed to a much later time (Limb + 10000 seconds).
func ApplyLimbTimeOffset(dp *e2eutils.DeployParams, limbTimeOffset *hexutil.Uint64) {
	dp.DeployConfig.L2GenesisMantleLimbTimeOffset = limbTimeOffset

	// Ensure Arsia doesn't activate at the same time as or before Limb
	if limbTimeOffset == nil {
		// If Limb is disabled, also disable Arsia
		dp.DeployConfig.L2GenesisMantleArsiaTimeOffset = nil
	} else {
		// If Arsia is not set or would activate at/before Limb, delay it significantly
		if dp.DeployConfig.L2GenesisMantleArsiaTimeOffset == nil ||
			*dp.DeployConfig.L2GenesisMantleArsiaTimeOffset <= *limbTimeOffset {
			// Delay Arsia to Limb + 10000 seconds (far in the future for test purposes)
			delayedArsia := hexutil.Uint64(*limbTimeOffset + 10000)
			dp.DeployConfig.L2GenesisMantleArsiaTimeOffset = &delayedArsia
		}
	}
}

// ApplyArsiaTimeOffset adjusts fork configuration to activate Arsia fork at the specified time.
// Arsia is a Mantle-specific fork that encompasses multiple OP Stack forks:
// Canyon + Delta + Ecotone + Fjord + Granite + Holocene + Isthmus + Jovian
//
// This function ensures all constituent OP Stack forks are activated at the Arsia time,
// and that earlier Mantle forks (Limb and predecessors) won't accidentally activate after Arsia.
func ApplyArsiaTimeOffset(dp *e2eutils.DeployParams, arsiaTimeOffset *hexutil.Uint64) {
	dp.DeployConfig.L2GenesisMantleArsiaTimeOffset = arsiaTimeOffset

	// If Arsia is being activated, set all constituent OP Stack forks to activate at the same time
	if arsiaTimeOffset != nil {
		// Canyon
		if dp.DeployConfig.L2GenesisCanyonTimeOffset == nil {
			dp.DeployConfig.L2GenesisCanyonTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisCanyonTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisCanyonTimeOffset = arsiaTimeOffset
		}

		// Delta
		if dp.DeployConfig.L2GenesisDeltaTimeOffset == nil {
			dp.DeployConfig.L2GenesisDeltaTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisDeltaTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisDeltaTimeOffset = arsiaTimeOffset
		}

		// Ecotone
		if dp.DeployConfig.L2GenesisEcotoneTimeOffset == nil {
			dp.DeployConfig.L2GenesisEcotoneTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisEcotoneTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisEcotoneTimeOffset = arsiaTimeOffset
		}

		// Fjord
		if dp.DeployConfig.L2GenesisFjordTimeOffset == nil {
			dp.DeployConfig.L2GenesisFjordTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisFjordTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisFjordTimeOffset = arsiaTimeOffset
		}

		// Granite
		if dp.DeployConfig.L2GenesisGraniteTimeOffset == nil {
			dp.DeployConfig.L2GenesisGraniteTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisGraniteTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisGraniteTimeOffset = arsiaTimeOffset
		}

		// Holocene
		if dp.DeployConfig.L2GenesisHoloceneTimeOffset == nil {
			dp.DeployConfig.L2GenesisHoloceneTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisHoloceneTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisHoloceneTimeOffset = arsiaTimeOffset
		}

		// Isthmus
		if dp.DeployConfig.L2GenesisIsthmusTimeOffset == nil {
			dp.DeployConfig.L2GenesisIsthmusTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisIsthmusTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisIsthmusTimeOffset = arsiaTimeOffset
		}

		// Jovian
		if dp.DeployConfig.L2GenesisJovianTimeOffset == nil {
			dp.DeployConfig.L2GenesisJovianTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisJovianTimeOffset != *arsiaTimeOffset {
			dp.DeployConfig.L2GenesisJovianTimeOffset = arsiaTimeOffset
		}
	}

	// Configure Limb to activate with Arsia (Arsia depends on Limb)
	if arsiaTimeOffset != nil {
		// Ensure Limb is also activated (at or before Arsia)
		if dp.DeployConfig.L2GenesisMantleLimbTimeOffset == nil {
			// If Limb is not set, activate it at the same time as Arsia
			dp.DeployConfig.L2GenesisMantleLimbTimeOffset = arsiaTimeOffset
		} else if *dp.DeployConfig.L2GenesisMantleLimbTimeOffset > *arsiaTimeOffset {
			// Limb must activate before or at the same time as Arsia
			dp.DeployConfig.L2GenesisMantleLimbTimeOffset = arsiaTimeOffset
		}
	}
}

// ApplyLimbToArsiaUpgrade configures a deployment to start with Limb active and upgrade to Arsia later.
// This is useful for testing the Limb → Arsia upgrade path.
//
// limbTimeOffset: Time offset for Limb activation (typically 0 for genesis activation)
// arsiaTimeOffset: Time offset for Arsia activation (must be > limbTimeOffset)
func ApplyLimbToArsiaUpgrade(dp *e2eutils.DeployParams, limbTimeOffset, arsiaTimeOffset *hexutil.Uint64) {
	// First apply Limb configuration
	ApplyLimbTimeOffset(dp, limbTimeOffset)

	// Then apply Arsia configuration
	// Note: ApplyArsiaTimeOffset will handle ensuring Limb doesn't come after Arsia
	ApplyArsiaTimeOffset(dp, arsiaTimeOffset)
}
