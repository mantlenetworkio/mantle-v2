package eth

import (
	"sync"

	"github.com/ethereum/go-ethereum/params"
)

// sepoliaL1ChainConfig returns Sepolia fork timings and metadata with mainnet's
// blob schedule (Cancun/Prague only). Sepolia's built-in config includes extra
// fork stages (Osaka, BPO1, …) that mainnet does not yet expose in BlobScheduleConfig.
var sepoliaL1ChainConfig = sync.OnceValue(func() *params.ChainConfig {
	cfg := *params.SepoliaChainConfig
	cfg.BlobScheduleConfig = params.MainnetChainConfig.BlobScheduleConfig
	return &cfg
})

// L1ChainConfigByChainID returns the chain config for the given chain ID,
// if it is in the set of known chain IDs (Mainnet, Sepolia, Holesky, Hoodi).
// If the chain ID is not known, it returns nil.
func L1ChainConfigByChainID(chainID ChainID) *params.ChainConfig {
	switch chainID {
	case ChainIDFromBig(params.MainnetChainConfig.ChainID):
		return params.MainnetChainConfig
	case ChainIDFromBig(params.SepoliaChainConfig.ChainID):
		return sepoliaL1ChainConfig()
	case ChainIDFromBig(params.HoleskyChainConfig.ChainID):
		return params.HoleskyChainConfig
	case ChainIDFromBig(params.HoodiChainConfig.ChainID):
		return params.HoodiChainConfig
	default:
		return nil
	}
}
