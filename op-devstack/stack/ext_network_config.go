package stack

import (
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type ExtNetworkConfig struct {
	L2NetworkName      string
	L1ChainID          eth.ChainID
	L2ChainID          eth.ChainID
	L2ELEndpoint       string
	L1CLBeaconEndpoint string
	L1ELEndpoint       string

	// L1ChainConfig is the chain config for the L1 network.
	// When provided, the eth.L1ChainConfigByChainID lookup is skipped.
	// Required for non-standard L1 chains (e.g. local devnets with chain ID 31337).
	L1ChainConfig *params.ChainConfig
	// RollupConfig is the rollup configuration for the L2 network.
	// When provided, the superchain registry lookup is skipped.
	RollupConfig *rollup.Config
	// L2ChainConfig is the chain config (genesis params) for the L2 network.
	// When provided, the superchain registry lookup is skipped.
	L2ChainConfig *params.ChainConfig
}
