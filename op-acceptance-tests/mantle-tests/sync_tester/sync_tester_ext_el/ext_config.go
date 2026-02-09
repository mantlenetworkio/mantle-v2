package sync_tester_ext_el

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum-optimism/optimism/devnet-sdk/descriptors"
	"github.com/ethereum-optimism/optimism/devnet-sdk/shell/env"
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

// GetNetworkPresetFromEnv constructs an ExtNetworkConfig directly from the devnet
// environment descriptor (loaded from the URL in the DEVNET_ENV_URL env var).
// This extracts rollup config, chain config, and service endpoints from the env JSON,
// bypassing the superchain registry entirely. This is the preferred approach for networks
// not in the superchain registry (e.g. Mantle).
func GetNetworkPresetFromEnv() (stack.ExtNetworkConfig, error) {
	devnetURL := os.Getenv(env.EnvURLVar)
	if devnetURL == "" {
		return stack.ExtNetworkConfig{}, fmt.Errorf("no devnet URL specified (set %s)", env.EnvURLVar)
	}
	devnetEnv, err := env.LoadDevnetFromURL(devnetURL)
	if err != nil {
		return stack.ExtNetworkConfig{}, fmt.Errorf("failed to load devnet environment: %w", err)
	}
	if devnetEnv.Env == nil {
		return stack.ExtNetworkConfig{}, fmt.Errorf("devnet environment is nil")
	}
	if len(devnetEnv.Env.L2) == 0 {
		return stack.ExtNetworkConfig{}, fmt.Errorf("devnet environment has no L2 chains")
	}
	if devnetEnv.Env.L1 == nil {
		return stack.ExtNetworkConfig{}, fmt.Errorf("devnet environment has no L1 chain")
	}

	l2 := devnetEnv.Env.L2[0]
	l1 := devnetEnv.Env.L1

	config := stack.ExtNetworkConfig{
		L2NetworkName: l2.Name,
		L1ChainConfig: l1.Config,
		RollupConfig:  l2.RollupConfig,
		L2ChainConfig: l2.Config,
	}

	if l1.Config != nil && l1.Config.ChainID != nil {
		config.L1ChainID = eth.ChainIDFromBig(l1.Config.ChainID)
	}
	if l2.Config != nil && l2.Config.ChainID != nil {
		config.L2ChainID = eth.ChainIDFromBig(l2.Config.ChainID)
	}

	// Extract L1 EL endpoint from the first L1 node's "el" service
	if len(l1.Nodes) > 0 {
		if elService, ok := l1.Nodes[0].Services["el"]; ok {
			config.L1ELEndpoint = resolveServiceEndpoint(elService, "rpc")
		}
		// Extract L1 CL beacon endpoint from the first L1 node's "cl" service
		if clService, ok := l1.Nodes[0].Services["cl"]; ok {
			config.L1CLBeaconEndpoint = resolveServiceEndpoint(clService, "http")
		}
	}

	// Extract L2 EL endpoint from the first L2 node's "el" service
	if len(l2.Nodes) > 0 {
		if elService, ok := l2.Nodes[0].Services["el"]; ok {
			config.L2ELEndpoint = resolveServiceEndpoint(elService, "rpc")
		}
	}

	return config, nil
}

// resolveServiceEndpoint builds an endpoint URL from a service descriptor's endpoint info.
func resolveServiceEndpoint(service *descriptors.Service, protocol string) string {
	endpoint, ok := service.Endpoints[protocol]
	if !ok || endpoint == nil {
		return ""
	}
	scheme := endpoint.Scheme
	if scheme == "" {
		scheme = "http"
	}
	host := endpoint.Host
	path := ""
	if strings.Contains(host, "/") {
		parts := strings.SplitN(host, "/", 2)
		host = parts[0]
		path = "/" + parts[1]
	}
	if endpoint.Port != 0 {
		return fmt.Sprintf("%s://%s:%d%s", scheme, host, endpoint.Port, path)
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

// getEnvOrDefault returns the environment variable value or the default if not set
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}
