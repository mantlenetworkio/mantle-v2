package synctester

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum-optimism/optimism/devnet-sdk/descriptors"
	"github.com/ethereum-optimism/optimism/devnet-sdk/shell/env"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

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
