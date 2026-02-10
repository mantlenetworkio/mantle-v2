package sysgo

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params/forks"
)

// Mantle variant of DefaultMinimalSystemWithSyncTester.
func DefaultMantleMinimalSystemWithSyncTester(dest *DefaultMinimalSystemWithSyncTesterIDs, fcu eth.FCUState) stack.Option[*Orchestrator] {
	l1ID := DefaultL1ID
	l2ID := DefaultL2AID
	ids := NewDefaultMinimalSystemWithSyncTesterIDs(l1ID, l2ID)

	opt := stack.Combine[*Orchestrator]()
	opt.Add(stack.BeforeDeploy(func(o *Orchestrator) {
		o.P().Logger().Info("Setting up")
	}))

	opt.Add(WithMnemonicKeys(devkeys.TestMnemonic))

	opt.Add(WithMantleDeployer(),
		WithDeployerOptions(
			WithLocalContractSources(),
			WithCommons(ids.L1.ChainID()),
			WithPrefundedL2(ids.L1.ChainID(), ids.L2.ChainID()),
		),
	)

	opt.Add(WithL1Nodes(ids.L1EL, ids.L1CL))

	opt.Add(WithL2ELNode(ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL, ids.L1CL, ids.L1EL, ids.L2EL, L2CLSequencer()))

	opt.Add(WithBatcher(ids.L2Batcher, ids.L1EL, ids.L2CL, ids.L2EL))
	opt.Add(WithProposer(ids.L2Proposer, ids.L1EL, &ids.L2CL, nil))

	opt.Add(WithFaucets([]stack.L1ELNodeID{ids.L1EL}, []stack.L2ELNodeID{ids.L2EL}))

	opt.Add(WithTestSequencer(ids.TestSequencer, ids.L1CL, ids.L2CL, ids.L1EL, ids.L2EL))

	opt.Add(WithL2Challenger(ids.L2Challenger, ids.L1EL, ids.L1CL, nil, nil, &ids.L2CL, []stack.L2ELNodeID{
		ids.L2EL,
	}))

	opt.Add(WithSyncTester(ids.SyncTester, []stack.L2ELNodeID{ids.L2EL}))

	opt.Add(WithL2MetricsDashboard())

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = ids
	}))

	return opt
}

// Mantle variant of DefaultSimpleSystemWithSyncTester.
func DefaultMantleSimpleSystemWithSyncTester(dest *DefaultSimpleSystemWithSyncTesterIDs) stack.Option[*Orchestrator] {
	l1ID := DefaultL1ID
	l2ID := DefaultL2AID
	ids := NewDefaultSimpleSystemWithSyncTesterIDs(l1ID, l2ID)

	opt := stack.Combine[*Orchestrator]()
	opt.Add(stack.BeforeDeploy(func(o *Orchestrator) {
		o.P().Logger().Info("Setting up")
	}))

	opt.Add(WithMnemonicKeys(devkeys.TestMnemonic))

	opt.Add(WithMantleDeployer(),
		WithDeployerOptions(
			WithLocalContractSources(),
			WithCommons(ids.L1.ChainID()),
			WithDefaultBPOBlobSchedule,
			WithForkAtL1Genesis(forks.BPO2), // Both ethereum mainnet and sepolia have activated BPO2
			WithPrefundedL2(ids.L1.ChainID(), ids.L2.ChainID()),
		),
		WithDeployerPipelineOption(WithL1MNT(DefaultL1MNT)),
		WithDeployerPipelineOption(WithOperatorFeeVaultRecipient(DefaultOperatorFeeVaultRecipient)),
		WithDeployerPipelineOption(WithMantlePortalPaused(false)),
	)

	opt.Add(WithL1Nodes(ids.L1EL, ids.L1CL))

	opt.Add(WithL2ELNode(ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL, ids.L1CL, ids.L1EL, ids.L2EL, L2CLSequencer()))

	opt.Add(WithBatcher(ids.L2Batcher, ids.L1EL, ids.L2CL, ids.L2EL))
	opt.Add(WithLegacyProposer(ids.L2Proposer, ids.L1EL, &ids.L2CL, nil))

	opt.Add(WithFaucets([]stack.L1ELNodeID{ids.L1EL}, []stack.L2ELNodeID{ids.L2EL}))

	opt.Add(WithTestSequencer(ids.TestSequencer, ids.L1CL, ids.L2CL, ids.L1EL, ids.L2EL))

	opt.Add(WithSyncTester(ids.SyncTester, []stack.L2ELNodeID{ids.L2EL}))

	// Create a SyncTesterEL with the same chain ID as the EL node
	opt.Add(WithSyncTesterL2ELNode(ids.SyncTesterL2EL, ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL2, ids.L1CL, ids.L1EL, ids.SyncTesterL2EL))

	// P2P Connect CLs to signal unsafe heads
	opt.Add(WithL2CLP2PConnection(ids.L2CL, ids.L2CL2))

	opt.Add(WithL2MetricsDashboard())

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = ids
	}))

	return opt
}

// ExternalELSystemWithEndpoint creates a minimal external EL system
// using network config provided directly in the ExtNetworkConfig.
// This supports networks not in the superchain registry (e.g. Mantle) by
// accepting the rollup config and chain config via the ExtNetworkConfig struct.
func ExternalELSystemWithEndpoint(dest *DefaultMinimalExternalELSystemIDs, networkPreset stack.ExtNetworkConfig) stack.Option[*Orchestrator] {
	l2ChainID := networkPreset.L2ChainID

	ids := NewExternalELSystemIDs(networkPreset.L1ChainID, l2ChainID)

	opt := stack.Combine[*Orchestrator]()
	opt.Add(stack.BeforeDeploy(func(o *Orchestrator) {
		o.P().Logger().Info("Setting up external EL system", "network", networkPreset.L2NetworkName)
	}))

	opt.Add(WithMnemonicKeys(devkeys.TestMnemonic))

	// Resolve L1 chain config: prefer externally-provided config, fall back to well-known chain lookup
	l1ChainConfig := networkPreset.L1ChainConfig
	if l1ChainConfig == nil {
		chainID := ids.L1.ChainID()
		l1ChainConfig = eth.L1ChainConfigByChainID(chainID)
		if l1ChainConfig == nil {
			panic(fmt.Sprintf("unsupported L1 chain ID: %s (provide L1ChainConfig in ExtNetworkConfig for non-standard chains)", chainID))
		}
	}

	// Skip deployer since we're using external L1 and externally-provided L2 config
	// Create L1 network record for external L1
	opt.Add(stack.BeforeDeploy(func(o *Orchestrator) {
		l1Net := &L1Network{
			id: ids.L1,
			genesis: &core.Genesis{
				Config: l1ChainConfig,
			},
			blockTime: 12,
		}
		o.l1Nets.Set(ids.L1.ChainID(), l1Net)
	}))

	opt.Add(WithExtL1Nodes(ids.L1EL, ids.L1CL, networkPreset.L1ELEndpoint, networkPreset.L1CLBeaconEndpoint))

	// Use configs from ExtNetworkConfig
	opt.Add(WithEmptyDepSetFromExtConfig(
		stack.L2NetworkID(l2ChainID),
		networkPreset.L2NetworkName,
		networkPreset.RollupConfig,
		networkPreset.L2ChainConfig,
	))

	// Add SyncTester service with external endpoint
	opt.Add(WithSyncTesterWithExternalEndpoint(ids.SyncTester, networkPreset.L2ELEndpoint, l2ChainID))

	// Add SyncTesterL2ELNode as the L2EL replacement for real-world EL endpoint
	opt.Add(WithSyncTesterL2ELNode(ids.L2EL, ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL, ids.L1CL, ids.L1EL, ids.L2EL))

	opt.Add(WithExtL2Node(ids.L2ELReadOnly, networkPreset.L2ELEndpoint))

	opt.Add(WithL2MetricsDashboard())

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = ids
	}))

	return opt
}
