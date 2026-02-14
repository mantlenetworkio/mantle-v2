package sync_tester_ext_el

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// Configuration defaults for mantle-sepolia
const (
	DefaultNetworkPreset = "mantle-sepolia"

	// Tailscale networking endpoints
	DefaultL2ELEndpointTailscale       = ""
	DefaultL1CLBeaconEndpointTailscale = ""
	DefaultL1ELEndpointTailscale       = ""
)

var (
	// Network presets for different networks against which we test op-node syncing
	networkPresets = map[string]stack.ExtNetworkConfig{
		"mantle-sepolia": {
			L2NetworkName:      "mantle-sepolia",
			L1ChainID:          eth.ChainIDFromUInt64(11155111),
			L2ELEndpoint:       "",
			L1CLBeaconEndpoint: "",
			L1ELEndpoint:       "",
		},
		"mantle-mainnet": {
			L2NetworkName:      "mantle-mainnet",
			L1ChainID:          eth.ChainIDFromUInt64(1),
			L2ELEndpoint:       "",
			L1CLBeaconEndpoint: "",
			L1ELEndpoint:       "",
		},
	}
)

func GetNetworkPreset(name string) (stack.ExtNetworkConfig, error) {
	var config stack.ExtNetworkConfig
	if name == "" {
		config = networkPresets[DefaultNetworkPreset]
	} else {
		var ok bool
		config, ok = networkPresets[name]
		if !ok {
			return stack.ExtNetworkConfig{}, fmt.Errorf("NETWORK_PRESET %s not found", name)
		}
	}
	// Override configuration with Tailscale endpoints if Tailscale networking is enabled
	if os.Getenv("TAILSCALE_NETWORKING") == "true" {
		config.L2ELEndpoint = getEnvOrDefault("L2_EL_ENDPOINT_TAILSCALE", DefaultL2ELEndpointTailscale)
		config.L1CLBeaconEndpoint = getEnvOrDefault("L1_CL_BEACON_ENDPOINT_TAILSCALE", DefaultL1CLBeaconEndpointTailscale)
		config.L1ELEndpoint = getEnvOrDefault("L1_EL_ENDPOINT_TAILSCALE", DefaultL1ELEndpointTailscale)
	}
	return config, nil
}

// getEnvOrDefault returns the environment variable value or the default if not set
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}
