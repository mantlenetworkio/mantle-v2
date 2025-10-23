package rollup

type MantleForkName string

const (
	MantleBaseFee MantleForkName = "mantle_base_fee"
	MantleEverest MantleForkName = "mantle_everest"
	MantleEuboea  MantleForkName = "mantle_euboea"
	MantleSkadi   MantleForkName = "mantle_skadi"
	MantleLimb    MantleForkName = "mantle_limb"
	MantleArsia   MantleForkName = "mantle_arsia"
	MantleNone    MantleForkName = ""
	// This fork never existed, it is used to indicate that a optimism fork is not active at a given timestamp.
	MantleNoSupport MantleForkName = "mantle_no_support"
)

var AllMantleForks = []MantleForkName{
	MantleNoSupport,
	MantleBaseFee,
	MantleEverest,
	MantleEuboea,
	MantleSkadi,
	MantleLimb,
	MantleArsia,
	// ADD NEW FORKS HERE!
}

func ForkToMantleFork(fork ForkName) MantleForkName {
	switch fork {
	case Interop:
		return MantleNoSupport
	case Canyon, Delta, Ecotone, Fjord, Granite, Holocene, Isthmus, Jovian:
		return MantleArsia
	case Bedrock, Regolith:
		return MantleSkadi
	default:
		return MantleNone
	}
}
