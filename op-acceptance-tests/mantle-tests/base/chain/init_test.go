package chain

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-devstack/presets"
)

func TestMain(m *testing.M) {
	presets.DoMain(m,
		presets.WithMantleMinimal(),
		//stack.MakeCommon(sysgo.WithBatcherOption(func(_ stack.L2BatcherID, cfg *bss.CLIConfig) {
		//	// Brotli requires Fjord; use zlib before Fjord activates.
		//	cfg.CompressionAlgo = derive.Zlib
		//})),
	)
}
