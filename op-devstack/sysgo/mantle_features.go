package sysgo

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/inspect"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/intentbuilder"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	ps "github.com/ethereum-optimism/optimism/op-proposer/proposer"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testutils/tcpproxy"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	gn "github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/preconf"
)

var DefaultL1MNT = common.HexToAddress("0x8000000000000000000000000000000000000000")
var DefaultOperatorFeeVaultRecipient = common.HexToAddress("0x976EA74026E726554dB657fA54763abd0C3a0aa9")

func DefaultMantleMinimalSystem(dest *DefaultMinimalSystemIDs) stack.Option[*Orchestrator] {
	ids := NewDefaultMinimalSystemIDs(DefaultL1ID, DefaultL2AID)
	return defaultMantleMinimalSystemOpts(&ids, dest)
}

func defaultMantleMinimalSystemOpts(ids *DefaultMinimalSystemIDs, dest *DefaultMinimalSystemIDs) stack.CombinedOption[*Orchestrator] {
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
		WithDeployerPipelineOption(WithL1MNT(DefaultL1MNT)),
		WithDeployerPipelineOption(WithOperatorFeeVaultRecipient(DefaultOperatorFeeVaultRecipient)),
	)

	opt.Add(WithL1Nodes(ids.L1EL, ids.L1CL))

	opt.Add(WithMantleL2ELNode(ids.L2EL))
	opt.Add(WithL2CLNode(ids.L2CL, ids.L1CL, ids.L1EL, ids.L2EL, L2CLSequencer()))

	opt.Add(WithBatcher(ids.L2Batcher, ids.L1EL, ids.L2CL, ids.L2EL))
	opt.Add(WithLegacyProposer(ids.L2Proposer, ids.L1EL, &ids.L2CL, nil))

	opt.Add(WithFaucets([]stack.L1ELNodeID{ids.L1EL}, []stack.L2ELNodeID{ids.L2EL}))

	opt.Add(WithTestSequencer(ids.TestSequencer, ids.L1CL, ids.L2CL, ids.L1EL, ids.L2EL))

	// opt.Add(WithL2Challenger(ids.L2Challenger, ids.L1EL, ids.L1CL, nil, nil, &ids.L2CL, []stack.L2ELNodeID{
	// 	ids.L2EL,
	// }))

	opt.Add(WithL2MetricsDashboard())

	opt.Add(stack.Finally(func(orch *Orchestrator) {
		*dest = *ids
	}))

	return opt
}

// WithMantleL2ELNode adds the default type of L2 CL node.
// The default can be configured with DEVSTACK_L2EL_KIND.
// Tests that depend on specific types can use options like WithKonaNode and WithOpNode directly.
func WithMantleL2ELNode(id stack.L2ELNodeID, opts ...L2ELOption) stack.Option[*Orchestrator] {
	switch os.Getenv("DEVSTACK_L2EL_KIND") {
	case "op-reth":
		return WithOpReth(id, opts...)
	default:
		return WithMantleOpGeth(id, opts...)
	}
}

// An alternative way to set the L1MNT and OperatorFeeVaultRecipient is to use the WithDeployerOption.
// It requires we extend the L2Configurator interface to include WithL1MNT and WithOperatorFeeVaultRecipient.
// Since MNT token address is a Mantle-only feature, directly modifying deployer pipeline seems cleaner.
// But if one day we integrate mantle rde-v3 as an orchestrator choice, we should think about which way is better.
func WithL1MNT(l1MNT common.Address) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting L1MNT", "l1MNT", l1MNT.Hex())
		for _, l2 := range intent.Chains {
			l2.L1MNT = l1MNT
		}
	}
}

func WithOperatorFeeVaultRecipient(recipient common.Address) DeployerPipelineOption {
	return func(_ *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting OperatorFeeVaultRecipient", "recipient", recipient.Hex())
		for _, l2 := range intent.Chains {
			l2.OperatorFeeVaultRecipient = recipient
		}
	}
}

func WithLegacyProposer(proposerID stack.L2ProposerID, l1ELID stack.L1ELNodeID,
	l2CLID *stack.L2CLNodeID, supervisorID *stack.SupervisorID) stack.Option[*Orchestrator] {
	return stack.AfterDeploy(func(orch *Orchestrator) {
		WithLegacyProposerOption(orch, proposerID)
		WithProposerPostDeploy(orch, proposerID, l1ELID, l2CLID, supervisorID)
	})
}

func WithLegacyProposerOption(orch *Orchestrator, proposerID stack.L2ProposerID) {
	orch.proposerOptions = append(orch.proposerOptions, func(id stack.L2ProposerID, cfg *ps.CLIConfig) {
		ctx := orch.P().Ctx()
		ctx = stack.ContextWithID(ctx, proposerID)
		p := orch.P().WithCtx(ctx)

		require := p.Require()
		l2Net, ok := orch.l2Nets.Get(proposerID.ChainID())
		require.True(ok)
		cfg.L2OOAddress = l2Net.deployment.L2OOAddress().Hex()
		// turn off DGF by setting relating contract address to empty string
		cfg.DGFAddress = ""
	})
}

func (d *L2Deployment) L2OOAddress() common.Address {
	return d.l2OOAddress
}

func WithMantleForkAtGenesis(fork forks.MantleForkName) DeployerPipelineOption {
	return func(wb *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		wb.require.True(forks.IsValidMantleFork(fork), "invalid mantle fork: %s", string(fork))
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting MantleFork at genesis", "fork", string(fork))
		var opts []DeployerPipelineOption
		var future bool
		for _, refFork := range forks.AllMantleForks {
			if future {
				opts = append(opts, WithMantleForkAtOffset(refFork, nil))
			} else {
				opts = append(opts, WithMantleForkAtOffset(refFork, new(uint64)))
			}

			if refFork == fork {
				future = true
			}
		}

		for _, opt := range opts {
			opt(wb, intent, cfg)
		}
	}
}

func WithMantleForkAtOffset(fork forks.MantleForkName, offset *uint64) DeployerPipelineOption {
	return func(wb *worldBuilder, intent *state.Intent, cfg *deployer.ApplyPipelineOpts) {
		wb.require.True(forks.IsValidMantleFork(fork), "invalid mantle fork: %s", string(fork))
		cfg.Logger.New("stage", "set-up-mantle-env").Info("Setting MantleFork at offset", "fork", string(fork), "offset", offset)
		for _, l2 := range intent.Chains {
			key := fmt.Sprintf("l2Genesis%sTimeOffset", string(fork))
			if offset == nil {
				delete(l2.DeployOverrides, key)
			} else {
				// The typing is important, or op-deployer merge-JSON tricks will fail
				l2.DeployOverrides[key] = (*hexutil.Uint64)(offset)
			}
		}

	}
}

/////////////////////////////////////////////////////////////
// Deployer
/////////////////////////////////////////////////////////////

func WithMantleDeployer() stack.Option[*Orchestrator] {
	return stack.FnOption[*Orchestrator]{
		BeforeDeployFn: func(o *Orchestrator) {
			o.P().Require().Nil(o.wb, "must not already have a world builder")
			o.wb = &worldBuilder{
				p:       o.P(),
				logger:  o.P().Logger(),
				require: o.P().Require(),
				keys:    o.keys,
				builder: intentbuilder.New(),
			}
		},
		DeployFn: func(o *Orchestrator) {
			o.P().Require().NotNil(o.wb, "must have a world builder")
			o.wb.deployerPipelineOptions = o.deployerPipelineOptions
			o.wb.BuildMantle()
		},
		AfterDeployFn: func(o *Orchestrator) {
			wb := o.wb
			require := o.P().Require()
			require.NotNil(o.wb, "must have a world builder")

			l1ID := stack.L1NetworkID(eth.ChainIDFromUInt64(wb.output.AppliedIntent.L1ChainID))
			superchainID := stack.SuperchainID("main")
			clusterID := stack.ClusterID("main")

			l1Net := &L1Network{
				id:        l1ID,
				genesis:   wb.outL1Genesis,
				blockTime: 6,
			}
			o.l1Nets.Set(l1ID.ChainID(), l1Net)

			o.superchains.Set(superchainID, &Superchain{
				id:         superchainID,
				deployment: wb.outSuperchainDeployment,
			})
			o.clusters.Set(clusterID, &Cluster{
				id:     clusterID,
				cfgset: wb.outFullCfgSet,
			})

			for _, chainID := range wb.l2Chains {
				l2Genesis, ok := wb.outL2Genesis[chainID]
				require.True(ok, "L2 genesis must exist")
				l2RollupCfg, ok := wb.outL2RollupCfg[chainID]
				require.True(ok, "L2 rollup config must exist")
				l2Dep, ok := wb.outL2Deployment[chainID]
				require.True(ok, "L2 deployment must exist")

				l2ID := stack.L2NetworkID(chainID)
				l2Net := &L2Network{
					id:         l2ID,
					l1ChainID:  l1ID.ChainID(),
					genesis:    l2Genesis,
					rollupCfg:  l2RollupCfg,
					deployment: l2Dep,
					keys:       o.keys,
				}
				o.l2Nets.Set(l2ID.ChainID(), l2Net)
			}
		},
	}
}

/////////////////////////////////////////////////////////////
// world builder
/////////////////////////////////////////////////////////////

func (wb *worldBuilder) BuildMantle() {
	st := &state.State{
		Version: 1,
	}

	// Work-around of op-deployer design issue.
	// We use the same deployer key for all L1 and L2 chains we deploy here.
	deployerKey, err := wb.keys.Secret(devkeys.DeployerRole.Key(big.NewInt(0)))
	wb.require.NoError(err, "need deployer key")

	intent, err := wb.builder.Build()
	wb.require.NoError(err)

	pipelineOpts := deployer.ApplyPipelineOpts{
		DeploymentTarget:   deployer.DeploymentTargetGenesis,
		L1RPCUrl:           "",
		DeployerPrivateKey: deployerKey,
		Intent:             intent,
		State:              st,
		Logger:             wb.logger,
		StateWriter:        wb, // direct output back here
	}
	for _, opt := range wb.deployerPipelineOptions {
		opt(wb, intent, &pipelineOpts)
	}

	err = deployer.MantleApplyPipeline(wb.p.Ctx(), pipelineOpts)
	wb.require.NoError(err)

	wb.require.NotNil(wb.output, "expected state-write to output")

	for _, id := range wb.output.Chains {
		chainID := eth.ChainIDFromBytes32(id.ID)
		wb.l2Chains = append(wb.l2Chains, chainID)
	}

	wb.buildL1Genesis()
	wb.buildMantleL2Genesis()
	wb.buildL2DeploymentOutputs()
	wb.buildFullConfigSet()
}

func (wb *worldBuilder) buildMantleL2Genesis() {
	wb.outL2Genesis = make(map[eth.ChainID]*core.Genesis)
	wb.outL2RollupCfg = make(map[eth.ChainID]*rollup.Config)
	for _, ch := range wb.output.Chains {
		l2Genesis, l2RollupCfg, err := inspect.MantleGenesisAndRollup(wb.output, ch.ID)
		wb.require.NoError(err, "need L2 genesis and rollup")
		id := eth.ChainIDFromBytes32(ch.ID)
		wb.outL2Genesis[id] = l2Genesis
		wb.outL2RollupCfg[id] = l2RollupCfg
	}
}

/////////////////////////////////////////////////////////////
// mantle op geth
/////////////////////////////////////////////////////////////

func WithMantleOpGeth(id stack.L2ELNodeID, opts ...L2ELOption) stack.Option[*Orchestrator] {
	return stack.AfterDeploy(func(orch *Orchestrator) {
		p := orch.P().WithCtx(stack.ContextWithID(orch.P().Ctx(), id))
		require := p.Require()

		l2Net, ok := orch.l2Nets.Get(id.ChainID())
		require.True(ok, "L2 network required")

		cfg := DefaultL2ELConfig()
		orch.l2ELOptions.Apply(p, id, cfg)       // apply global options
		L2ELOptionBundle(opts).Apply(p, id, cfg) // apply specific options

		jwtPath, jwtSecret := orch.writeDefaultJWT()

		logger := p.Logger()

		l2EL := &OpGeth{
			id:        id,
			p:         orch.P(),
			logger:    logger,
			l2Net:     l2Net,
			jwtPath:   jwtPath,
			jwtSecret: jwtSecret,
		}
		l2EL.StartMantle()
		p.Cleanup(func() {
			l2EL.Stop()
		})
		require.True(orch.l2ELs.SetIfMissing(id, l2EL), "must be unique L2 EL node")
	})
}

func (n *OpGeth) StartMantle() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.l2Geth != nil {
		n.logger.Warn("op-geth already started")
		return
	}

	if n.authProxy == nil {
		n.authProxy = tcpproxy.New(n.logger.New("proxy", "l2el-auth"))
		n.p.Require().NoError(n.authProxy.Start())
		n.p.Cleanup(func() {
			n.authProxy.Close()
		})
		n.authRPC = "ws://" + n.authProxy.Addr()
	}
	if n.userProxy == nil {
		n.userProxy = tcpproxy.New(n.logger.New("proxy", "l2el-user"))
		n.p.Require().NoError(n.userProxy.Start())
		n.p.Cleanup(func() {
			n.userProxy.Close()
		})
		n.userRPC = "ws://" + n.userProxy.Addr()
	}

	require := n.p.Require()
	l2Geth, err := geth.InitL2(n.id.String(), n.l2Net.genesis, n.jwtPath,
		func(ethCfg *ethconfig.Config, nodeCfg *gn.Config) error {
			ethCfg.Miner.PreconfConfig = &preconf.MinerConfig{}
			// disable mantle upgrades so that we can configure mantle forks as we need
			ethCfg.ApplyMantleUpgrades = false
			nodeCfg.P2P = p2p.Config{
				NoDiscovery: true,
				ListenAddr:  "127.0.0.1:0",
				MaxPeers:    10,
			}
			return nil
		})
	require.NoError(err)
	require.NoError(l2Geth.Node.Start())
	n.l2Geth = l2Geth
	n.authProxy.SetUpstream(ProxyAddr(require, l2Geth.AuthRPC().RPC()))
	n.userProxy.SetUpstream(ProxyAddr(require, l2Geth.UserRPC().RPC()))
}
