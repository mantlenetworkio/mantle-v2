package pipeline

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-chain-ops/addresses"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum/go-ethereum/common"
)

func DeployMantleProxies(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "deploy-mantle-proxies")

	if !shouldDeployMantleProxies(st, chainID) {
		lgr.Info("mantle proxies deployment not needed")
		return nil
	}

	lgr.Info("deploying Mantle proxies")

	dpo, err := opcm.RunMantleDeployProxiesScript(env.MantleScripts.DeployProxies)
	if err != nil {
		return fmt.Errorf("error deploying Mantle proxies: %w", err)
	}

	// Store the proxy addresses in state
	st.Chains = append(st.Chains, &state.ChainState{
		ID: chainID,
		OpChainContracts: addresses.OpChainContracts{
			OpChainCoreContracts: addresses.OpChainCoreContracts{
				OpChainProxyAdminImpl:             dpo.ProxyAdmin,
				AddressManagerImpl:                dpo.AddressManager,
				L1StandardBridgeProxy:             dpo.L1StandardBridgeProxy,
				L1CrossDomainMessengerProxy:       dpo.L1CrossDomainMessengerProxy,
				OptimismPortalProxy:               dpo.OptimismPortalProxy,
				OptimismMintableErc20FactoryProxy: dpo.OptimismMintableERC20FactoryProxy,
				L1Erc721BridgeProxy:               dpo.L1ERC721BridgeProxy,
				SystemConfigProxy:                 dpo.SystemConfigProxy,
			},
			OpChainLegacyContracts: addresses.OpChainLegacyContracts{
				L2OutputOracleProxy: dpo.L2OutputOracleProxy,
			},
		},
	})

	return nil
}

func shouldDeployMantleProxies(st *state.State, chainID common.Hash) bool {
	for _, chain := range st.Chains {
		if chain.ID == chainID {
			// deploy proxies will create chain state
			return false
		}
	}
	return true
}
