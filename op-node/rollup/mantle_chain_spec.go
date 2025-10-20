package rollup

type MantleForkName string

const (
	MantleEverest MantleForkName = "mantle_everest"
	MantleEuboea  MantleForkName = "mantle_euboea"
	MantleSkadi   MantleForkName = "mantle_skadi"
	MantleLimb    MantleForkName = "mantle_limb"
	MantleArsia   MantleForkName = "mantle_arsia"
	MantleNone    MantleForkName = ""
)

var AllMantleForks = []MantleForkName{
	MantleEverest,
	MantleEuboea,
	MantleSkadi,
	MantleLimb,
	MantleArsia,
	// ADD NEW FORKS HERE!
}

func ForkToMantleFork(fork ForkName) MantleForkName {
	switch fork {
	case Bedrock, Regolith:
		return MantleEverest
	case Canyon, Delta, Ecotone, Fjord, Granite, Holocene, Isthmus, Jovian:
		return MantleArsia
	default:
		return MantleNone
	}
}
