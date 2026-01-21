package env

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/ethereum-optimism/optimism/devnet-sdk/descriptors"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum/go-ethereum/log"
)

const FeatureMantle = "mantle"

// fixupMantleDevnetConfig applies Mantle-specific configuration fixups.
// It checks for the "mantle" feature flag and fetches rollup configs from L2 nodes
// to ensure Mantle fork times are properly populated.
// Op-acceptor is an external tool that build with optimism code.
// Op-acceptor strips the Mantle fork times from the rollup config because it read and write the rollup config
// to a temp file and does not provide the original devnet environment to test runner. Along with reading, our
// added fields to the rollup config are stripped.
func fixupMantleDevnetConfig(config *descriptors.DevnetEnvironment) error {
	// Check if "mantle" feature is enabled
	if !slices.Contains(config.Features, FeatureMantle) {
		return nil
	}

	// Fetch rollup config from L2 nodes to get the authoritative config
	// This ensures we have all fork times (including Mantle-specific ones) that may have
	// been stripped by intermediate JSON processing (e.g., by op-acceptor)
	for _, l2Chain := range config.L2 {
		rollupCfg, err := fetchRollupConfigFromL2Node(l2Chain)
		if err != nil {
			// Log warning but don't fail - the existing config may still work
			log.Warn("failed to fetch rollup config from L2 node, using existing config",
				"chain", l2Chain.Name, "error", err)
			continue
		}
		if rollupCfg != nil {
			l2Chain.RollupConfig = rollupCfg
		}
	}

	return nil
}

// fetchRollupConfigFromL2Node attempts to fetch the rollup config from an L2 node's RPC endpoint.
// It looks for a node with a "cl" (consensus layer / op-node) service and queries its optimism_rollupConfig RPC method.
func fetchRollupConfigFromL2Node(l2Chain *descriptors.L2Chain) (*rollup.Config, error) {
	// Find a node with a CL (op-node) service
	var rpcURL string
	for _, node := range l2Chain.Nodes {
		if clService, ok := node.Services["cl"]; ok {
			if rpcEndpoint, ok := clService.Endpoints["rpc"]; ok {
				scheme := rpcEndpoint.Scheme
				if scheme == "" {
					scheme = "http"
				}
				rpcURL = fmt.Sprintf("%s://%s:%d", scheme, rpcEndpoint.Host, rpcEndpoint.Port)
				break
			}
		}
	}

	if rpcURL == "" {
		return nil, fmt.Errorf("no CL node RPC endpoint found for L2 chain %s", l2Chain.Name)
	}

	// Create RPC client with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lgr := log.New()
	rpcClient, err := client.NewRPC(ctx, lgr, rpcURL, client.WithDialAttempts(1))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to L2 node at %s: %w", rpcURL, err)
	}
	defer rpcClient.Close()

	// Create rollup client and fetch config
	rollupClient := sources.NewRollupClient(rpcClient)
	rollupCfg, err := rollupClient.RollupConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rollup config from %s: %w", rpcURL, err)
	}

	r, _ := json.Marshal(rollupCfg)
	log.Info("rollupCfg", "rollupCfg", string(r))

	return rollupCfg, nil
}
