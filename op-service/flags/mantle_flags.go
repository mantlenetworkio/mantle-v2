package flags

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	opservice "github.com/ethereum-optimism/optimism/op-service"
)

// OverridableMantleForks lists all mantle forks that can be overridden via CLI flags or env vars.
// It's all mainline forks from Skadi onwards.
var OverridableMantleForks = forks.MantleForksFrom(forks.MantleSkadi)

func MantleOverrideName(f forks.MantleForkName) string {
	name, _ := strings.CutPrefix(strings.ToLower(string(f)), "mantle")
	return "override." + name
}

func MantleOverrideEnvVar(envPrefix string, fork forks.MantleForkName) []string {
	name, _ := strings.CutPrefix(strings.ToLower(string(fork)), "mantle")
	return opservice.PrefixEnvVar(envPrefix, "OVERRIDE_"+strings.ToUpper(name))
}

func CLIMantleOverrideFlags(envPrefix string, category string) []cli.Flag {
	var flags []cli.Flag
	for _, fork := range OverridableMantleForks {
		flags = append(flags,
			&cli.Uint64Flag{
				Name:     MantleOverrideName(fork),
				Usage:    fmt.Sprintf("Manually specify the %s fork timestamp, overriding the bundled setting", fork),
				EnvVars:  MantleOverrideEnvVar(envPrefix, fork),
				Category: category,
			})
	}
	return flags
}
